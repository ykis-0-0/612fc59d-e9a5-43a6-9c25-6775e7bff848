package realip_zoning

import (
	"net/netip"
	"slices"
	"testing"

	"pgregory.net/rapid"
)

// LLM: add a property-based invariant test for CIDR consolidation output
func TestProp_Consolidate_IdempotentAndCanonical(t *testing.T) {
	mkIPv4Prefix := rapid.Custom(func(t *rapid.T) netip.Prefix {
		raw := rapid.Uint32().Draw(t, "raw")
		bits := rapid.IntRange(0, 32).Draw(t, "bits")
		addr := netip.AddrFrom4([4]byte{
			byte(raw >> 24),
			byte(raw >> 16),
			byte(raw >> 8),
			byte(raw),
		})

		return netip.PrefixFrom(addr, bits).Masked()
	})

	rapid.Check(t, func(t *rapid.T) {
		input := rapid.SliceOfN(mkIPv4Prefix, 0, 64).Draw(t, "input")

		consolidated := cidr_Consolidate(input)
		consolidatedAgain := cidr_Consolidate(consolidated)
		if !slices.Equal(consolidated, consolidatedAgain) {
			t.Fatalf("consolidation must be idempotent, got %v then %v", consolidated, consolidatedAgain)
		}

		for idx, prefix := range consolidated {
			if prefix != prefix.Masked() {
				t.Fatalf("consolidated prefix must be canonical: %s", prefix)
			}
			if idx == 0 {
				continue
			}

			prev := cidr_Unpack(consolidated[idx-1])
			curr := cidr_Unpack(prefix)

			if netip.Addr.Compare(prev.el01, curr.el01) != -1 {
				t.Fatalf("consolidated prefixes are not sorted by start address: %s then %s", consolidated[idx-1], prefix)
			}

			if netip.Addr.Compare(prev.el02, curr.el01) != -1 {
				t.Fatalf("consolidated prefixes overlap: %s and %s", consolidated[idx-1], prefix)
			}

			next := prev.el02.Next()
			if next.IsValid() && next == curr.el01 {
				t.Fatalf("consolidated prefixes must not remain adjacent: %s and %s", consolidated[idx-1], prefix)
			}
		}
	})
}

// LLM-END
