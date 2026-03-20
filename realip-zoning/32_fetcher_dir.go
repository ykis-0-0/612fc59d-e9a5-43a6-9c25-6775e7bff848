package realip_zoning

import (
	"net/netip"
)

func fetchCIDRsFromDir(dir string) ([]netip.Prefix, error) {
	syncer := newCollector()
	// TODO

	return syncer.collect()
}
