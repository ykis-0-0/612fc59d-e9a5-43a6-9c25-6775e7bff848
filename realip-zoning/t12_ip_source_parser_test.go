package realip_zoning

import (
	"errors"
	"io"
	"net/netip"
	"strings"
	"testing"
)

type errReader struct{}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestParseCIDR_ReadError(t *testing.T) {
	ch := make(chan netip.Prefix, 1)

	err := parseCIDRList(io.Reader(errReader{}), ch)
	if err == nil {
		t.Fatal("expected read error, got nil")
	}

	errText := err.Error()
	t.Log(errText)

	if !strings.Contains(errText, "while reading CIDR list") {
		t.Fatalf("error should include scanner context, got: %v", err)
	}
	if !strings.Contains(errText, "boom") {
		t.Fatalf("error should include source read error, got: %v", err)
	}
}

var cidrListTestCases = map[string]struct {
	have  string
	want  []netip.Prefix
	erred bool
}{
	"full-featured": {
		have: strings.Join([]string{
			"// Comment 1",
			" # Comment 2",
			"   192.0.2.0/24",
			"   ",
			"198.51.100.0/25  ",
			"",
			"2001:db8::/32",
		}, "\n"),
		want: []netip.Prefix{
			netip.MustParsePrefix("192.0.2.0/24"),
			netip.MustParsePrefix("198.51.100.0/25"),
			netip.MustParsePrefix("2001:db8::/32"),
		},
	},
	"invalid-line": {
		have: strings.Join([]string{
			"203.0.113.0/24",
			"not-a-cidr",
			"198.51.100.0/24",
		}, "\n"),
		want:  []netip.Prefix{netip.MustParsePrefix("203.0.113.0/24")},
		erred: true,
	},
	"v4-only": {
		have: strings.Join([]string{
			"192.0.2.0/24",
			"198.51.100.0/25",
		}, "\n"),
		want: []netip.Prefix{
			netip.MustParsePrefix("192.0.2.0/24"),
			netip.MustParsePrefix("198.51.100.0/25"),
		},
	},
	"v6-only": {
		have: strings.Join([]string{
			"2001:db8::/32",
			"fc00::/7",
		}, "\n"),
		want: []netip.Prefix{
			netip.MustParsePrefix("2001:db8::/32"),
			netip.MustParsePrefix("fc00::/7"),
		},
	},
}

func TestParseCIDR_Success(t *testing.T) {
	assertParseCase(t, cidrListTestCases["full-featured"])
}

func TestParseCIDR_InvalidLine(t *testing.T) {
	assertParseCase(t, cidrListTestCases["invalid-line"])
}

func assertParseCase(t *testing.T, tc struct {
	have  string
	want  []netip.Prefix
	erred bool
}) {
	t.Helper()

	ch := make(chan netip.Prefix, 8)
	err := parseCIDRList(strings.NewReader(tc.have), ch)

	if !tc.erred && err != nil {
		t.Fatalf("did not expect parse error, got: %v", err)
	}
	if tc.erred && err == nil {
		t.Fatal("expected parse error, got nil")
	}

	for i, want := range tc.want {
		got := <-ch
		if got != want {
			t.Fatalf("prefix[%d]: got %s, want %s", i, got, want)
		}
	}

	select {
	case extra := <-ch:
		t.Fatalf("unexpected extra prefix parsed: %s", extra)
	default:
	}
}
