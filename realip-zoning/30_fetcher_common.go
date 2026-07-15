package realip_zoning

import (
	"sync"
)

type fetchCollector[T any] struct {
	wg    sync.WaitGroup
	chRtv chan T
	chErr chan error
}

func newCollector[T any](threads int) fetchCollector[T] {
	// Eliminate edge cases first
	switch {
	case threads == 0:
		threads = 1
	case threads < 0:
		panic("negative thread count")
	}

	return fetchCollector[T]{
		chRtv: make(chan T, 5*threads),
		chErr: make(chan error, threads), // Couldn't be this bad right?
	}
}

func (syncer *fetchCollector[T]) collect() ([]T, []error) {
	// Capacity is guessed
	rtvs := make([]T, 0, 20)
	errs := make([]error, 0, 10)

	// Join childrens and close channels
	go func(syncer *fetchCollector[T]) {
		syncer.wg.Wait()
		close(syncer.chRtv)
		close(syncer.chErr)
	}(syncer)

	// Collection
	for !(syncer.chRtv == nil && syncer.chErr == nil) {
		select {
		case pkt, ok := <-syncer.chRtv:
			if !ok {
				syncer.chRtv = nil
				continue
			}
			rtvs = append(rtvs, pkt)

		case err, ok := <-syncer.chErr:
			if !ok {
				syncer.chErr = nil
				continue
			}
			errs = append(errs, err)
		}
	}

	// Use nil instead of empty slice
	if len(errs) == 0 {
		errs = nil
	}

	// errors.Join() is smart enough to handle (and even return) nils
	return rtvs, errs
}
