package realip_zoning

import (
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

func (srcset IPSources) getCIDRs() []netip.Prefix {
	// 1a. Fetch from URLs
	urls := make([]*url.URL, len(srcset.URLSources))
	for i, u := range srcset.URLSources {
		urls[i] = u.URL
	}

	cidrU, errU := fetchCIDRsFromURLs(urls)
	if errU != nil {
		// TODO
	}

	// 1b. Fetch from Files & Dir
	dir := srcset.DirSource
	cidrD, errD := fetchCIDRsFromDir(dir)
	if errD != nil {
		// TODO
	}

	cidrSet := slices.Concat(srcset.DirectSource, cidrU, cidrD)

	// TODO CIDR merging logic

	return cidrSet
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
		if prefix.IsValid() {
			chRtv <- prefix
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("while reading CIDR list: %w", err)
	}

	return nil
}
