package realip_zoning

import (
	"fmt"
	"os"

	"io/fs"
	"net/netip"

	"strings"
)

// Fetches and aggregates CIDR lists from specified local directory.
//
// Returns all errors aggregated alongside all successfully parsed prefixes
func fetchCIDRsFromDir(dir string) ([]netip.Prefix, error) {

	fso := os.DirFS(dir)
	entries, err := fs.ReadDir(fso, ".")
	if err != nil {
		return nil, fmt.Errorf("unable to list root folder %q: %w", dir, err)
	}

	syncer := newCollector(len(entries))
	syncer.wg.Add(len(entries))
	for _, ent := range entries {
		go worker(&syncer, fso, ent)
	}

	return syncer.collect()
}

func worker(syncer *fetchCollector, fsTree fs.FS, ent fs.DirEntry) {
	defer syncer.wg.Done()

	filename := ent.Name()

	// Fail fast first
	if ent.IsDir() {
		syncer.chErr <- fmt.Errorf("subdirectories found, flat directory expected: %q", filename)
		return
	}
	if strings.HasPrefix(filename, ".") {
		syncer.chErr <- fmt.Errorf("skipping hidden file: %q", filename)
		return
	}
	if !ent.Type().IsRegular() {
		syncer.chErr <- fmt.Errorf("backing off from non-regular file: %q", filename)
		return
	}

	inputStream, err := fsTree.Open(filename)
	if err != nil {
		syncer.chErr <- fmt.Errorf("unable to open file %q: %w", filename, err)
		return
	}
	defer inputStream.Close()

	if err := parseCIDRList(inputStream, syncer.chRtv); err != nil {
		syncer.chErr <- fmt.Errorf("while parsing CIDR list from %s: %w", filename, err)
	}
}
