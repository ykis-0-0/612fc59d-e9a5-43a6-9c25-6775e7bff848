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
