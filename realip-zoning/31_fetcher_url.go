package realip_zoning

import (
	"fmt"

	"net/http"
	"net/url"

	"net/netip"
)

func mkUrlFetcher(syncer *fetchCollector, href *url.URL) {
	defer syncer.wg.Done()

	resp, err := http.Get(href.String())
	if err != nil {
		syncer.chErr <- err
		return
	}
	defer resp.Body.Close()

	if err := parseCIDRList(resp.Body, syncer.chRtv); err != nil {
		syncer.chErr <- fmt.Errorf("while parsing CIDR list from %s: %w", href.String(), err)
	}
}

// Fetches and aggregates CIDR lists from list of URLs.
//
// Returns all errors aggregated alongside all successfully parsed prefixes
func fetchCIDRsFromURLs(urls []*url.URL) ([]netip.Prefix, error) {
	syncer := newCollector(len(urls))

	syncer.wg.Add(len(urls))
	for _, href := range urls {
		go mkUrlFetcher(&syncer, href)
	}

	return syncer.collect()
}
