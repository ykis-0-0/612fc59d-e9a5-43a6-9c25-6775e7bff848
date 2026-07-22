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

var cidrTestCases_Consolidate = testCases_t[[]netip.Prefix, []ipRange]{
	"empty": {
		have: []netip.Prefix{},
		want: []ipRange{},
	},
	"ipv4-sanity": {
		have: []netip.Prefix{mkCidr("127.0.0.0/8")},
		want: []ipRange{mkRange("127.0.0.0", "127.255.255.255")},
	},
	"ipv6-sanity": {
		have: []netip.Prefix{mkCidr("2001:db8::/32")},
		want: []ipRange{mkRange("2001:db8::", "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff")},
	},
	"ipv4-whole": {
		have: []netip.Prefix{mkCidr("0.0.0.0/1"), mkCidr("128.0.0.0/1")},
		want: []ipRange{mkRange("0.0.0.0", "255.255.255.255")},
	},
	"ipv4-adjacent-merge": {
		have: []netip.Prefix{mkCidr("10.0.0.0/24"), mkCidr("10.0.1.0/24")},
		want: []ipRange{mkRange("10.0.0.0", "10.0.1.255")},
	},
	"ipv4-overlap-contained": {
		have: []netip.Prefix{mkCidr("10.0.0.0/22"), mkCidr("10.0.2.0/24")},
		want: []ipRange{mkRange("10.0.0.0", "10.0.3.255")},
	},
	"ipv4-overlap-partial": {
		have: []netip.Prefix{mkCidr("10.0.0.0/24"), mkCidr("10.0.0.128/25")},
		want: []ipRange{mkRange("10.0.0.0", "10.0.0.255")},
	},
	"ipv4-unsorted-disjoint": {
		have: []netip.Prefix{mkCidr("10.0.2.0/24"), mkCidr("10.0.0.0/24")},
		want: []ipRange{
			mkRange("10.0.0.0", "10.0.0.255"),
			mkRange("10.0.2.0", "10.0.2.255"),
		},
	},
	"ipv6-adjacent-merge": {
		have: []netip.Prefix{mkCidr("2001:db8::/127"), mkCidr("2001:db8::2/127")},
		want: []ipRange{mkRange("2001:db8::", "2001:db8::3")},
	},
}

func mkCidr(s string) netip.Prefix {
	return netip.MustParsePrefix(s)
}

func mkAddr(s string) netip.Addr {
	return netip.MustParseAddr(s)
}

func mkRange(start, end string) ipRange {
	return ipRange{el01: mkAddr(start), el02: mkAddr(end)}
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

func TestRepack_Consolidate(t *testing.T) {
	type tCase_t = testCase_t[[]netip.Prefix, []ipRange]

	subRunner := func(tCase tCase_t) func(t *testing.T) {
		return func(t *testing.T) {
			got := cidr_Consolidate(tCase.have)

			mkBody := func(seq []ipRange) string {
				indentMapper := func(rng ipRange) string {
					return fmt.Sprintf("\t%s - %s", rng.el01, rng.el02)
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
