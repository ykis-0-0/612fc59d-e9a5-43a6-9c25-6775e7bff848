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

func newCollector() fetchCollector {
	return fetchCollector{
		chRtv: make(chan netip.Prefix, 10),
		chErr: make(chan error),
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
