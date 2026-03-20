package realip_zoning

import (
	"net/url"
	"testing"
)

// Integration test using Cloudflare's published list
func TestCidrUrlFetcher_Cloudflare(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	urlV4, err := url.Parse("https://www.cloudflare.com/ips-v4")
	if err != nil {
		t.Fatalf("failed to parse Cloudflare v4 URL: %v", err)
	}

	urlV6, err := url.Parse("https://www.cloudflare.com/ips-v6")
	if err != nil {
		t.Fatalf("failed to parse Cloudflare v6 URL: %v", err)
	}

	arg01 := []*url.URL{urlV4, urlV6}

	cidrs, err := fetchCIDRsFromURLs(arg01)
	if err != nil {
		t.Fatalf("fetchCIDRsFromURLs failed: %v", err)
	}

	if len(cidrs) == 0 {
		t.Fatal("expected at least one CIDR from Cloudflare URLs")
	}

	hasV4, hasV6 := false, false
	for _, cidr := range cidrs {
		hasV4 = hasV4 || cidr.Addr().Is4()
		hasV6 = hasV6 || cidr.Addr().Is6()
	}

	if !hasV4 || !hasV6 {
		t.Fatalf("expected both IPv4 and IPv6 CIDRs, got hasV4=%t hasV6=%t", hasV4, hasV6)
	}
}

// Integration test with an invalid URL to verify error handling
func TestCidrUrlFetcher_InvalidURL(t *testing.T) {
	invalidURL, _ := url.Parse("http://example.invalid.tld:65535/file-not-found.txt")
	urls := []*url.URL{invalidURL}

	cidrs, err := fetchCIDRsFromURLs(urls)

	// Should have an error (connection refused or similar)
	if err == nil {
		t.Error("Expected an error for unreachable URL, got nil")
	}

	// Should have no CIDRs
	if len(cidrs) > 0 {
		t.Errorf("Expected 0 CIDRs from failed fetch, got %d", len(cidrs))
	}
}
