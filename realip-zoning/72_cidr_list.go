package realip_zoning

import (
	"fmt"
	"math"

	"iter"
	bigMath "math/big"
	"slices"

	"net"
	"net/netip"
)

// Sort and merge list of CIDR prefixes,
// canonical forms are assumed and WILL be panicked on.
// Empty list get empty result.
func cidr_Consolidate(cidrs []netip.Prefix) []netip.Prefix {
	// guard
	if len(cidrs) == 0 {
		return []netip.Prefix{}
	}

	// unpack
	it_UnpackedPairs := mkMapper(cidr_Unpack, slices.Values(cidrs))

	// sort
	ar_UnpackedPairs := slices.SortedFunc(it_UnpackedPairs, cidr_SortFunc)

	// merge
	it_mergedPairs := cidr_TryMerge(ar_UnpackedPairs)

	// repack
	ar_mergedCidrs := slices.Collect(cidr_Repack(it_mergedPairs))

	return ar_mergedCidrs
}

type cidr_AddrPair_t = Zipped2[netip.Addr, netip.Addr]
type cidr_it_AddrPair_t = iter.Seq[cidr_AddrPair_t]
type cidr_ar_AddrPair_t = []cidr_AddrPair_t
type cidr_it_Prefix_t = iter.Seq[netip.Prefix]

func cidr_Unpack(cidr netip.Prefix) cidr_AddrPair_t {
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

	return cidr_AddrPair_t{el01: addr1st, el02: rebuilt}
}

func cidr_SortFunc(a, b cidr_AddrPair_t) int {
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

func cidr_TryMerge(unpackedCidrs cidr_ar_AddrPair_t) cidr_it_AddrPair_t {
	return func(yield func(cidr_AddrPair_t) bool) {
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

func cidr_Repack(unpackedCidrs cidr_it_AddrPair_t) cidr_it_Prefix_t {
	return func(yield func(netip.Prefix) bool) {
		for thisPair := range unpackedCidrs {
			for cidr := range cidr_AddrPairToPrefixes(thisPair) {
				if !yield(cidr) {
					return
				}
			}
		}
	}
}

// Use https://github.com/python/cpython/blob/3.14/Lib/ipaddress.py#L200
func cidr_AddrPairToPrefixes(pair cidr_AddrPair_t) cidr_it_Prefix_t {
	return func(yield func(netip.Prefix) bool) {
		type bigInt_t = bigMath.Int
		type p_bigInt_t = *bigInt_t

		if pair.el01.BitLen() != pair.el02.BitLen() {
			panic("addr pair length mismatch")
		}
		bitLen := pair.el01.BitLen()

		packedL, packedR := cidr_mkBigInt(pair.el01), cidr_mkBigInt(pair.el02)

		var packedRplus1 bigInt_t
		if addrRplus1 := pair.el02.Next(); addrRplus1.IsValid() {
			packedRplus1 = cidr_mkBigInt(addrRplus1)
		} else {
			p_bigInt_t.Lsh(&packedRplus1, bigMath.NewInt(1), uint(bitLen))
		}

		for p_bigInt_t.Cmp(&packedL, &packedR) != +1 {
			rangeBits := (bigMath.NewInt(0).Sub(&packedRplus1, &packedL)).BitLen() - 1

			pLengthL_ := packedL.TrailingZeroBits()

			var pLengthL int
			switch {
			case math.MaxInt < pLengthL_:
				panic("unreachable code, sanity guard")
			case packedL.BitLen() == 0: // Fastest way to check zero
				pLengthL = int(bitLen)
			default:
				pLengthL = int(pLengthL_)
			}

			pushAddr, ok := netip.AddrFromSlice(packedL.FillBytes(make([]byte, bitLen/8)))
			if !ok {
				panic("failed to rebuild address from slice")
			}

			pushPLen := min(pLengthL, rangeBits)
			pushCidr := netip.PrefixFrom(pushAddr, bitLen-pushPLen)
			if !yield(pushCidr) {
				return
			}

			var consumed bigInt_t
			p_bigInt_t.Lsh(&consumed, bigMath.NewInt(1), uint(pushPLen))
			p_bigInt_t.Add(&packedL, &packedL, &consumed)
		}
	}
}

func cidr_mkBigInt(addr netip.Addr) bigMath.Int {
	var bigAddr bigMath.Int
	bigAddr.SetBytes(addr.AsSlice())

	return bigAddr
}
