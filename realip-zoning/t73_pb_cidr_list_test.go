package realip_zoning

import (
	"encoding/binary"
	"slices"
	"testing"

	"net/netip"

	"pgregory.net/rapid"
)

// LLM-DID: generators for canonical same-family CIDR lists
func genCanonicalPrefix(isV6 bool) *rapid.Generator[netip.Prefix] {
	if isV6 {
		return rapid.Custom(func(t *rapid.T) netip.Prefix {
			hi := rapid.Uint64().Draw(t, "v6.hi")
			lo := rapid.Uint64().Draw(t, "v6.lo")
			bits := rapid.IntRange(0, 128).Draw(t, "v6.bits")

			var raw [16]byte
			binary.BigEndian.PutUint64(raw[0:8], hi)
			binary.BigEndian.PutUint64(raw[8:16], lo)

			addr := netip.AddrFrom16(raw)
			return netip.PrefixFrom(addr, bits).Masked()
		})
	}

	return rapid.Custom(func(t *rapid.T) netip.Prefix {
		raw := rapid.Uint32().Draw(t, "v4.raw")
		bits := rapid.IntRange(0, 32).Draw(t, "v4.bits")

		addr := netip.AddrFrom4([4]byte{
			byte(raw >> 24),
			byte(raw >> 16),
			byte(raw >> 8),
			byte(raw),
		})

		return netip.PrefixFrom(addr, bits).Masked()
	})
}

func genPrefixList(t *rapid.T, minLen int, maxLen int) []netip.Prefix {
	isV6 := rapid.Bool().Draw(t, "family.v6")
	return rapid.SliceOfN(genCanonicalPrefix(isV6), minLen, maxLen).Draw(t, "cidrs")
}

func containsAddr(ranges []ipRange, addr netip.Addr) bool {
	for _, rng := range ranges {
		if rng.el01.BitLen() != addr.BitLen() {
			continue
		}
		if netip.Addr.Compare(rng.el01, addr) <= 0 && netip.Addr.Compare(addr, rng.el02) <= 0 {
			return true
		}
	}

	return false
}

// LLM-END

// LLM-DID: property that consolidation is invariant to input ordering
func TestPB_ConsolidatePermutationInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cidrs := genPrefixList(t, 1, 64)
		gotBase := cidr_Consolidate(cidrs)

		perm := rapid.Permutation(slices.Clone(cidrs)).Draw(t, "perm")
		gotPerm := cidr_Consolidate(perm)

		if !slices.Equal(gotBase, gotPerm) {
			t.Fatalf("consolidation changed after permutation: base=%v perm=%v", gotBase, gotPerm)
		}
	})
}

// LLM-END

// LLM-DID: properties for output shape and boundary coverage
func TestPB_ConsolidateInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cidrs := genPrefixList(t, 1, 64)
		got := cidr_Consolidate(cidrs)

		for i, rng := range got {
			if netip.Addr.Compare(rng.el01, rng.el02) > 0 {
				t.Fatalf("invalid range order at %d: %v", i, rng)
			}
			if i == 0 {
				continue
			}

			prev := got[i-1]
			if netip.Addr.Compare(prev.el01, rng.el01) > 0 {
				t.Fatalf("ranges out of order: prev=%v curr=%v", prev, rng)
			}
			if prev.el01.BitLen() == rng.el01.BitLen() {
				if netip.Addr.Compare(prev.el02, rng.el01) >= 0 {
					t.Fatalf("ranges overlap or touch: prev=%v curr=%v", prev, rng)
				}
				if prev.el02.Next() == rng.el01 {
					t.Fatalf("ranges are still adjacent after consolidate: prev=%v curr=%v", prev, rng)
				}
			}
		}

		for _, cidr := range cidrs {
			unpacked := cidr_Unpack(cidr)
			if !containsAddr(got, unpacked.el01) {
				t.Fatalf("missing start boundary %s from %s", unpacked.el01, cidr)
			}
			if !containsAddr(got, unpacked.el02) {
				t.Fatalf("missing end boundary %s from %s", unpacked.el02, cidr)
			}
		}
	})
}

// LLM-END
