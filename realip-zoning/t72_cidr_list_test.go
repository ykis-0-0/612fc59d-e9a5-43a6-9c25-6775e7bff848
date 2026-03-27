package realip_zoning

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"net/netip"
)

type testCase_t[T, R any] struct {
	have T
	want R
}
type testCases_t[T, R any] = map[string]testCase_t[T, R]

var cidrTestCases_AddrPairToPrefixes = testCases_t[[2]netip.Addr, []netip.Prefix]{
	"ipv4-single-prefix": {
		have: [2]netip.Addr{mkAddr("192.168.1.0"), mkAddr("192.168.1.255")},
		want: []netip.Prefix{mkCidr("192.168.1.0/24")},
	},
	"ipv6-single-prefix": {
		have: [2]netip.Addr{mkAddr("2001:db8::"), mkAddr("2001:db8::ffff")},
		want: []netip.Prefix{mkCidr("2001:db8::/112")},
	},
	"ipv4-split-range": {
		have: [2]netip.Addr{mkAddr("192.168.1.1"), mkAddr("192.168.1.130")},
		want: []netip.Prefix{
			mkCidr("192.168.1.1/32"),
			mkCidr("192.168.1.2/31"),
			mkCidr("192.168.1.4/30"),
			mkCidr("192.168.1.8/29"),
			mkCidr("192.168.1.16/28"),
			mkCidr("192.168.1.32/27"),
			mkCidr("192.168.1.64/26"),
			mkCidr("192.168.1.128/31"),
			mkCidr("192.168.1.130/32"),
		},
	},
}

var cidrTestCases_Consolidate = testCases_t[[]netip.Prefix, []netip.Prefix]{
	"ipv4-sanity": {
		have: []netip.Prefix{mkCidr("127.0.0.0/8")},
		want: []netip.Prefix{mkCidr("127.0.0.0/8")},
	},
	"ipv6-sanity": {
		have: []netip.Prefix{mkCidr("2001:db8::/32")},
		want: []netip.Prefix{mkCidr("2001:db8::/32")},
	},
	"ipv4-whole": {
		have: []netip.Prefix{mkCidr("0.0.0.0/1"), mkCidr("128.0.0.0/1")},
		want: []netip.Prefix{mkCidr("0.0.0.0/0")},
	},
	"ipv4-adjacent-merge": {
		have: []netip.Prefix{mkCidr("10.0.0.0/24"), mkCidr("10.0.1.0/24")},
		want: []netip.Prefix{mkCidr("10.0.0.0/23")},
	},
	"ipv4-overlap-contained": {
		have: []netip.Prefix{mkCidr("10.0.0.0/22"), mkCidr("10.0.2.0/24")},
		want: []netip.Prefix{mkCidr("10.0.0.0/22")},
	},
	"ipv4-overlap-partial": {
		have: []netip.Prefix{mkCidr("10.0.0.0/24"), mkCidr("10.0.0.128/25")},
		want: []netip.Prefix{mkCidr("10.0.0.0/24")},
	},
	"ipv4-unsorted-disjoint": {
		have: []netip.Prefix{mkCidr("10.0.2.0/24"), mkCidr("10.0.0.0/24")},
		want: []netip.Prefix{mkCidr("10.0.0.0/24"), mkCidr("10.0.2.0/24")},
	},
	"ipv6-adjacent-merge": {
		have: []netip.Prefix{mkCidr("2001:db8::/127"), mkCidr("2001:db8::2/127")},
		want: []netip.Prefix{mkCidr("2001:db8::/126")},
	},
}

func mkCidr(s string) netip.Prefix {
	return netip.MustParsePrefix(s)
}

func mkAddr(s string) netip.Addr {
	return netip.MustParseAddr(s)
}

func TestInvalidAddr(t *testing.T) {
	mty := netip.Prefix{}
	masked := mty.Masked()
	addr := masked.Addr()

	t.Logf("Masked: %s, Addr: %s", masked, addr)
}

func TestSorter_sanity_01(t *testing.T) {
	left, right := netip.MustParseAddr("0.0.0.2"), netip.MustParseAddr("::1")
	result := netip.Addr.Compare(left, right)

	switch result {
	case -1:
		t.Logf("netip.Addr.Compare(%s, %s) = %d", left, right, result)
	default:
		t.Errorf("netip.Addr.Compare(%s, %s) == %d (expected -1)", left, right, result)
	}
}

func TestSorter_sanity_02(t *testing.T) {
	left, right := netip.MustParseAddr("255.255.255.255"), netip.MustParseAddr("::")
	result := netip.Addr.Compare(left, right)

	switch result {
	case -1:
		t.Logf("netip.Addr.Compare(%s, %s) = %d", left, right, result)
	default:
		t.Errorf("netip.Addr.Compare(%s, %s) == %d (expected -1)", left, right, result)
	}
}

func TestAddrUnpack_sanity_IPv4(t *testing.T) {
	cidr := mkCidr("192.168.0.0/16")
	mustStart := netip.MustParseAddr("192.168.0.0")
	mustEnd := netip.MustParseAddr("192.168.255.255")

	testHelper_addrUnpack(t, cidr, mustStart, mustEnd)
}

func TestAddrUnpack_sanity_IPv6(t *testing.T) {
	cidr := mkCidr("2001:db8::/32")
	mustStart := netip.MustParseAddr("2001:db8::")
	mustEnd := netip.MustParseAddr("2001:db8:ffff:ffff:ffff:ffff:ffff:ffff")

	testHelper_addrUnpack(t, cidr, mustStart, mustEnd)
}

func testHelper_addrUnpack(t *testing.T, cidr netip.Prefix, mustStart netip.Addr, mustEnd netip.Addr) {
	t.Helper()

	unpacked := cidr_Unpack(cidr)
	t.Logf("cidr_Unpack(%s) = (%s, %s)", cidr, unpacked.el01, unpacked.el02)

	if unpacked.el01 != mustStart {
		t.Errorf("Expected start %s, got %s", mustStart, unpacked.el01)
	}
	if unpacked.el02 != mustEnd {
		t.Errorf("Expected end %s, got %s", mustEnd, unpacked.el02)
	}
}

func TestRepack_AddrPairToPrefixes(t *testing.T) {
	type tCase_t = testCase_t[[2]netip.Addr, []netip.Prefix]

	subRunner := func(tCase tCase_t) func(t *testing.T) {
		return func(t *testing.T) {
			got := slices.Collect(cidr_AddrPairToPrefixes(
				cidr_AddrPair_t{tCase.have[0], tCase.have[1]},
			))

			mkBody := func(seq []netip.Prefix) string {
				indentMapper := func(cidr netip.Prefix) string {
					return fmt.Sprintf("\t%s", cidr)
				}
				lines := slices.Collect(mkMapper(indentMapper, slices.Values(seq)))
				return strings.Join(lines, "\n")
			}
			goodBody := mkBody(tCase.want)
			badBody := mkBody(got)

			switch slices.Equal(tCase.want, got) {
			case true:
				t.Logf("cidr_Repack(%v) =>\n%s\n", tCase.have, goodBody)
			case false:
				t.Errorf("cidr_Repack(%v) =>\n%s\nbut expect\n%s\n", tCase.have, badBody, goodBody)
			}
		}
	}

	for name, tCase := range cidrTestCases_AddrPairToPrefixes {
		t.Run(name, subRunner(tCase))
	}
}

func TestRepack_Consolidate(t *testing.T) {
	type tCase_t = testCase_t[[]netip.Prefix, []netip.Prefix]

	subRunner := func(tCase tCase_t) func(t *testing.T) {
		return func(t *testing.T) {
			got := cidr_Consolidate(tCase.have)

			mkBody := func(seq []netip.Prefix) string {
				indentMapper := func(cidr netip.Prefix) string {
					return fmt.Sprintf("\t%s", cidr)
				}
				lines := slices.Collect(mkMapper(indentMapper, slices.Values(seq)))
				return strings.Join(lines, "\n")
			}
			goodBody := mkBody(tCase.want)
			badBody := mkBody(got)

			switch slices.Equal(tCase.want, got) {
			case true:
				t.Logf("cidr_Consolidate(%v) =>\n%s\n", tCase.have, goodBody)
			case false:
				t.Errorf("cidr_Consolidate(%v) =>\n%s\nbut expect\n%s\n", tCase.have, badBody, goodBody)
			}
		}
	}

	for name, tCase := range cidrTestCases_Consolidate {
		t.Run(name, subRunner(tCase))
	}
}
