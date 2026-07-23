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
	URLSources   []urlMarshal   `json:"fromURLs,omitempty"`
	DirSource    string         `json:"fromDir,omitempty"`
	DirectSource []netip.Prefix `json:"fromList,omitempty"`
}

func (srcset IPSources) intoIpRanges() ([]ipRange, error) {
	// 1a. Fetch from URLs
	urls := make([]*url.URL, len(srcset.URLSources))
	for i, u := range srcset.URLSources {
		urls[i] = u.URL
	}

	cidrUrl, errU := fetchCIDRsFromURLs(urls)

	// 1b. Fetch from Files & Dir
	cidrDir, errD := fetchCIDRsFromDir(srcset.DirSource)

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
	errL := slices.Collect(
		mkMapper(checker, slices.Values(srcset.DirectSource)),
	)

	if err := errors.Join(slices.Concat(errL, errU, errD)...); err != nil {
		return nil, err
	}

	cidrSet := slices.Concat(srcset.DirectSource, cidrUrl, cidrDir)
	allIps := cidr_Consolidate(cidrSet)

	return allIps, nil
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
