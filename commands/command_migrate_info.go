package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/spf13/cobra"
)

var (
	// migrateInfoTopN is a flag given to the git-lfs-migrate(1) subcommand
	// 'info' which specifies how many info entries to show by default.
	migrateInfoTopN int

	// migrateInfoAboveFmt is a flag given to the git-lfs-migrate(1)
	// subcommand 'info' specifying a human-readable string threshold of
	// filesize before entries are counted.
	migrateInfoAboveFmt string
	// migrateInfoAbove is the number of bytes parsed from the above
	// migrateInfoAboveFmt flag.
	migrateInfoAbove uint64
)

func migrateInfoCommand(cmd *cobra.Command, args []string) {
	exts := make(map[string]*MigrateInfoEntry)

	above, err := humanize.ParseBytes(migrateInfoAboveFmt)
	if err != nil {
		ExitWithError(errors.Wrap(err, "cannot parse --above=<n>"))
	}

	migrateInfoAbove = above

	migrate(cmd, args, func(path string, b *odb.Blob) (*odb.Blob, error) {
		ext := fmt.Sprintf("*%s", filepath.Ext(path))

		if len(ext) > 1 {
			entry := exts[ext]
			if entry == nil {
				entry = &MigrateInfoEntry{Qualifier: ext}
			}

			entry.Total++
			entry.BytesTotal += b.Size

			if b.Size > int64(migrateInfoAbove) {
				entry.TotalAbove++
				entry.BytesAbove += b.Size
			}

			// TODO(@ttaylorr): needed?
			exts[ext] = entry
		}

		return b, nil
	})

	entries := EntriesBySize(MapToEntries(exts))
	sort.Sort(sort.Reverse(entries))

	migrateInfoTopN = tools.ClampInt(migrateInfoTopN, len(entries), 0)

	entries = entries[:tools.MaxInt(0, migrateInfoTopN)]

	entries.Print(os.Stderr)
}

// MigrateInfoEntry represents a tuple of filetype to total size taken by that
// file type.
type MigrateInfoEntry struct {
	// Qualifier is the filepath's extension.
	Qualifier string

	BytesAbove int64
	TotalAbove int64
	BytesTotal int64
	Total      int64
}

// MapToEntries creates a set of `*MigrateInfoEntry`'s for a given map of
// filepath extensions to file size in bytes.
func MapToEntries(exts map[string]*MigrateInfoEntry) []*MigrateInfoEntry {
	entries := make([]*MigrateInfoEntry, 0, len(exts))
	for _, entry := range exts {
		entries = append(entries, entry)
	}

	return entries
}

// EntriesBySize is an implementation of sort.Interface that sorts a set of
// `*MigrateInfoEntry`'s
type EntriesBySize []*MigrateInfoEntry

// Len returns the total length of the set of `*MigrateInfoEntry`'s.
func (e EntriesBySize) Len() int { return len(e) }

// Less returns the whether or not the MigrateInfoEntry given at `i` takes up
// less total size than the MigrateInfoEntry given at `j`.
func (e EntriesBySize) Less(i, j int) bool { return e[i].BytesAbove < e[j].BytesAbove }

// Swap swaps the entries given at i, j.
func (e EntriesBySize) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// Print formats the `*MigrateInfoEntry`'s in the set and prints them to the
// given io.Writer, "to", returning "n" the number of bytes written, and any
// error, if one occurred.
func (e EntriesBySize) Print(to io.Writer) (int, error) {
	extensions := make([]string, 0, len(e))
	files := make([]string, 0, len(e))
	percentages := make([]string, 0, len(e))

	// build columns for each entry
	for _, entry := range e {
		extensions = append(extensions, entry.Qualifier)
		files = append(files, fmt.Sprintf("%s, %d/%d files(s)",
			humanize.FormatBytes(uint64(entry.BytesAbove)),
			entry.TotalAbove,
			entry.Total,
		))
		percentages = append(percentages, fmt.Sprintf("%.0f%%",
			100*(float64(entry.TotalAbove)/float64(entry.Total)),
		))
	}

	// pad columns so they align
	extensions = tools.Ljust(extensions)
	files = tools.Rjust(files)
	percentages = tools.Rjust(percentages)

	// write the header
	n, err := fmt.Fprintf(to, "Files above %s:\n", humanize.FormatBytes(migrateInfoAbove))
	if err != nil {
		return n, err
	}

	// write each line with aligned columns
	for i := 0; i < len(e); i++ {
		extension := extensions[i]
		fileCount := files[i]
		percentage := percentages[i]
		ln, err := fmt.Fprintln(to, strings.Join([]string{extension, fileCount, percentage}, "\t"))
		n += ln
		if err != nil {
			return n, err
		}
	}

	return n, nil
}
