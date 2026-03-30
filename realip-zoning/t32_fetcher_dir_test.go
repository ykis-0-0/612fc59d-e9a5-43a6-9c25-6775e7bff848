package realip_zoning

import (
	"errors"
	"testing"
	"time"

	"net/netip"

	"os"
	"path/filepath"
)

func runFetchWithTimeout(t *testing.T, dir string, d time.Duration) ([]netip.Prefix, error) {
	t.Helper()

	type result struct {
		cidrs []netip.Prefix
		err   []error
	}
	ch := make(chan result, 1)

	go func() {
		cidrs, errs := fetchCIDRsFromDir(dir)
		ch <- result{cidrs, errs}
	}()

	select {
	case res := <-ch:
		return res.cidrs, errors.Join(res.err...)
	case <-time.After(d):
		t.Fatal("fetchCIDRsFromDir timed out (possible deadlock)")
		return nil, nil
	}
}

func TestCidrDirFetcher_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	cidrs, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if err != nil {
		t.Errorf("unexpected error for empty dir: %v", err)
	}

	if len(cidrs) != 0 {
		t.Errorf("expected 0 CIDRs, got %d: %v", len(cidrs), cidrs)
	}
}

func TestCidrDirFetcher_SingleFile(t *testing.T) {
	tc := cidrListTestCases["full-featured"]

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cidrs.txt"), []byte(tc.have), 0o644); err != nil {
		t.Fatal(err)
	}

	cidrs, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cidrs) != len(tc.want) {
		t.Errorf("expected %d CIDRs, got %d: %v", len(tc.want), len(cidrs), cidrs)
	}
}

func TestCidrDirFetcher_MultipleFiles(t *testing.T) {
	tcV4, tcV6 := cidrListTestCases["v4-only"], cidrListTestCases["v6-only"]
	files := map[string]string{
		"v4.txt": tcV4.have,
		"v6.txt": tcV6.have,
	}
	expected := len(tcV4.want) + len(tcV6.want)

	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cidrs, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cidrs) != expected {
		t.Errorf("expected %d CIDRs, got %d: %v", expected, len(cidrs), cidrs)
	}
}

func TestCidrDirFetcher_NonExistentDir(t *testing.T) {
	_, err := runFetchWithTimeout(t, "/invalid/non-existent/path/", 3*time.Second)
	if err == nil {
		t.Error("expected an error for nonexistent directory, got nil")
	}
}

func TestCidrDirFetcher_InvalidCIDRContent(t *testing.T) {
	tc := cidrListTestCases["invalid-line"]

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.txt"), []byte(tc.have), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if tc.erred && err == nil {
		t.Error("expected an error for invalid CIDR content, got nil")
	}
}

func TestCidrDirFetcher_WithSubdir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if err == nil {
		t.Error("expected an error when a subdirectory is present, got nil")
	}
}

func TestCidrDirFetcher_HiddenFile(t *testing.T) {
	tc := cidrListTestCases["full-featured"]

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte(tc.have), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := runFetchWithTimeout(t, dir, 3*time.Second)
	if err == nil {
		t.Error("expected an error for hidden file, got nil")
	}
}
