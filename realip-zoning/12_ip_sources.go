package realip_zoning

import (
	"errors"
	"fmt"

	"bufio"
	"io"
	"strings"

	"net/url"

	"net/netip"
	"slices"
)

type IPSources struct {
	URLSources []urlMarshal `json:"fromURLs,omitempty"`

	// Deprecatable, we can symlink to the files in the dir of DirSource.
	FileSources []string `json:"fromFiles,omitempty"`
	DirSource   string   `json:"fromDir,omitempty"`

	DirectSource []netip.Prefix `json:"fromList,omitempty"`
}

func (srcset IPSources) getCIDRs() ([]netip.Prefix, error) {
	// 1a. Fetch from URLs
	urls := make([]*url.URL, len(srcset.URLSources))
	for i, u := range srcset.URLSources {
		urls[i] = u.URL
	}

	cidrU, errU := fetchCIDRsFromURLs(urls)

	// 1b. Fetch from Files & Dir
	dir := srcset.DirSource
	cidrD, errD := fetchCIDRsFromDir(dir)

	checker := func(p netip.Prefix) error {
		if !p.IsValid() {
			// Mostly redundant, probably already caught in deserialization
			return fmt.Errorf("when parsing YAML zonedef; invalid CIDR: %s", p.String())
		}
		if p.Masked().Addr() != p.Addr() {
			return fmt.Errorf("when parsing YAML zonedef; non-canonical CIDR: %s", p.String())
		}
		return nil
	}
	errL := slices.DeleteFunc(
		slices.Collect(
			mkMapper(checker, slices.Values(srcset.DirectSource)),
		),
		func(err error) bool { return err == nil },
	)

	if len(errL) > 0 || len(errD) > 0 || len(errU) > 0 {
		return nil, errors.Join(slices.Concat(errL, errU, errD)...)
	}

	cidrSet := slices.Concat(srcset.DirectSource, cidrU, cidrD)
	cidrSet = cidr_Consolidate(cidrSet)

	return cidrSet, nil
}

// Parse a machine-readable plain-text list of CIDRs.
// One CIDR per line.
// Empty lines are ignored.
// Spaces are trimmed, Comment constructs not available.
func parseCIDRList(inputStream io.Reader, chRtv chan<- netip.Prefix) error {
	scanner := bufio.NewScanner(inputStream)
	lineNo := 0

	for scanner.Scan() {
		lineNo++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return fmt.Errorf("on line %d with content [%s]: %w", lineNo, line, err)
		}
		if !prefix.IsValid() {
			return fmt.Errorf("on line %d with content [%s]: invalid CIDR", lineNo, line)
		}
		if prefix.Masked().Addr() != prefix.Addr() {
			return fmt.Errorf("on line %d with content [%s]: non-canonical CIDR", lineNo, line)
		}
		chRtv <- prefix
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("while reading CIDR list: %w", err)
	}

	return nil
}
