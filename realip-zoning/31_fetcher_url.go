package realip_zoning

import (
	"net/url"
	"net/netip"
)

// Fetches and aggregates CIDR lists from list of URLs.
//
// Returns all errors aggregated alongside all successfully parsed prefixes
func fetchCIDRsFromURLs(urls []*url.URL) ([]netip.Prefix, error) {
	syncer := newCollector()
	// TODO

	return syncer.collect()
}
