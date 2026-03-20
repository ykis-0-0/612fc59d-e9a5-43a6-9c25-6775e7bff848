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
