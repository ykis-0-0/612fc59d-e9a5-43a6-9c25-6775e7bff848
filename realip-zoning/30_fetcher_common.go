package realip_zoning

import (
	"errors"
	"net/netip"
	"sync"
)

type fetchCollector struct {
	wg    sync.WaitGroup
	chRtv chan netip.Prefix
	chErr chan error
}

func newCollector(threads int) fetchCollector {
	// Eliminate edge cases first
	switch {
	case threads == 0:
		threads = 1
	case threads < 0:
		panic("negative thread count")
	}

	return fetchCollector{
		chRtv: make(chan netip.Prefix, 5*threads),
		chErr: make(chan error, threads), // Couldn't be this bad right?
	}
}

func (syncer *fetchCollector) collect() ([]netip.Prefix, error) {
	// Capacity is guessed
	cidrs := make([]netip.Prefix, 0, 20)
	errs := make([]error, 0, 10)

	// Join childrens and close channels
	go func(syncer *fetchCollector) {
		syncer.wg.Wait()
		close(syncer.chRtv)
		close(syncer.chErr)
	}(syncer)

	// Collection
	for !(syncer.chRtv == nil && syncer.chErr == nil) {
		select {
		case cidr, ok := <-syncer.chRtv:
			if !ok {
				syncer.chRtv = nil
				continue
			}
			cidrs = append(cidrs, cidr)

		case err, ok := <-syncer.chErr:
			if !ok {
				syncer.chErr = nil
				continue
			}
			errs = append(errs, err)
		}
	}

	// errors.Join() is smart enough to handle (and even return) nils
	return cidrs, errors.Join(errs...)
}
