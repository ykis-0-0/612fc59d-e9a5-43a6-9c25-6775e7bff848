package realip_zoning

import (
	"fmt"
	"iter"
	"slices"

	"net"
	"net/netip"
)

type ipRange Zipped2[netip.Addr, netip.Addr]

// Sort and merge list of CIDR prefixes into IP Ranges,
// canonical forms are assumed and WILL be panicked on.
// Empty list get empty result.
func cidr_Consolidate(cidrs []netip.Prefix) []ipRange {
	// guard
	if len(cidrs) == 0 {
		return []ipRange{}
	}

	// unpack
	it_UnpackedPairs := mkMapper(cidr_Unpack, slices.Values(cidrs))

	// sort
	ar_UnpackedPairs := slices.SortedFunc(it_UnpackedPairs, cidr_SortFunc)

	// merge
	it_mergedPairs := cidr_TryMerge(ar_UnpackedPairs)

	// collect
	ar_mergedPairs := slices.Collect(it_mergedPairs)

	return ar_mergedPairs
}

func cidr_Unpack(cidr netip.Prefix) ipRange {
	mask := net.CIDRMask(cidr.Bits(), cidr.Addr().BitLen())

	addr1st := cidr.Masked().Addr()
	if addr1st != cidr.Addr() {
		panic(fmt.Errorf("non-canonical CIDR: %q", cidr))
	}

	sliceToBaddr := func(tup Zipped2[byte, byte]) byte {
		return tup.el01 | ^tup.el02
	}
	zipped := mkZip_short(slices.Values(addr1st.AsSlice()), slices.Values(mask))
	remapped := slices.Collect(mkMapper(sliceToBaddr, zipped))

	rebuilt, ok := netip.AddrFromSlice(remapped)
	if !ok {
		panic(fmt.Errorf("failed to rebuild address from slice: %v", remapped))
	}

	return ipRange{el01: addr1st, el02: rebuilt}
}

func cidr_SortFunc(a, b ipRange) int {
	lead := netip.Addr.Compare(a.el01, b.el01)
	end := netip.Addr.Compare(a.el02, b.el02)

	switch {
	case lead != 0:
		return lead
	case end != 0:
		return end
	default:
		// Honestly we don't need this clause
		// but normal if-else is like, so fucking ugly
		return 0
	}
}

func cidr_TryMerge(unpackedCidrs []ipRange) iter.Seq[ipRange] {
	return func(yield func(ipRange) bool) {
		keeping := unpackedCidrs[0]
		for _, looking := range unpackedCidrs[1:] {
			cmp := netip.Addr.Compare(keeping.el02, looking.el01)
			isAdjacent := keeping.el02.Next() == looking.el01

			switch {
			case cmp == -1 && !isAdjacent:
				// Complete disjoint, send and reaggregate
				if !yield(keeping) {
					return
				}
				keeping = looking
			default:
				// Touching, Border, and Full Overlap
				if netip.Addr.Compare(keeping.el02, looking.el02) != -1 {
					// strict subset, keep current
					continue
				}
				keeping.el02 = looking.el02
			}
		}

		// Send the last one
		_ = yield(keeping)
	}
}
