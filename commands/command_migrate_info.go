package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git/githistory"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/git-lfs/gitobj/v2"
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

	// migrateInfoUnitFmt is a flag given to the git-lfs-migrate(1)
	// subcommand 'info' specifying a human-readable string of units with
	// which to display the number of bytes.
	migrateInfoUnitFmt string
	// migrateInfoUnit is the number of bytes in the unit given as
	// migrateInfoUnitFmt.
	migrateInfoUnit uint64
)

func migrateInfoCommand(cmd *cobra.Command, args []string) {
	l := tasklog.NewLogger(os.Stderr,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)

	db, err := getObjectDatabase()
	if err != nil {
		ExitWithError(err)
	}
	defer db.Close()

	rewriter := getHistoryRewriter(cmd, db, l)

	exts := make(map[string]*MigrateInfoEntry)

	above, err := humanize.ParseBytes(migrateInfoAboveFmt)
	if err != nil {
		ExitWithError(errors.Wrap(err, "cannot parse --above=<n>"))
	}

	if u := cmd.Flag("unit"); u.Changed {
		unit, err := humanize.ParseByteUnit(u.Value.String())
		if err != nil {
			ExitWithError(errors.Wrap(err, "cannot parse --unit=<unit>"))
		}

		migrateInfoUnit = unit
	}

	migrateInfoAbove = above

	migrate(args, rewriter, l, &githistory.RewriteOptions{
		BlobFn: func(path string, b *gitobj.Blob) (*gitobj.Blob, error) {
			ext := fmt.Sprintf("*%s", filepath.Ext(path))

			// If extension exists, group all items under extension,
			// else just use the file name.
			var groupName string
			if len(ext) > 1 {
				groupName = ext
			} else {
				groupName = filepath.Base(path)
			}

			entry := exts[groupName]
			if entry == nil {
				entry = &MigrateInfoEntry{Qualifier: groupName}
			}

			entry.Total++
			entry.BytesTotal += b.Size

			if b.Size > int64(migrateInfoAbove) {
				entry.TotalAbove++
				entry.BytesAbove += b.Size
			}

			exts[groupName] = entry

			return b, nil
		},
	})
	l.Close()

	entries := EntriesBySize(MapToEntries(exts))
	entries = removeEmptyEntries(entries)
	sort.Sort(sort.Reverse(entries))

	migrateInfoTopN = tools.ClampInt(migrateInfoTopN, len(entries), 0)

	entries = entries[:tools.MaxInt(0, migrateInfoTopN)]

	entries.Print(os.Stdout)
}

// MigrateInfoEntry represents a tuple of filetype to bytes and entry count
// above and below a threshold.
type MigrateInfoEntry struct {
	// Qualifier is the filepath's extension.
	Qualifier string

	// BytesAbove is total size of all files above a given threshold.
	BytesAbove int64
	// TotalAbove is the count of all files above a given size threshold.
	TotalAbove int64
	// BytesTotal is the number of bytes of all files
	BytesTotal int64
	// Total is the count of all files.
	Total int64
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

// removeEmptyEntries removes `*MigrateInfoEntry`'s for which no matching file
// is above the given threshold "--above".
func removeEmptyEntries(entries []*MigrateInfoEntry) []*MigrateInfoEntry {
	nz := make([]*MigrateInfoEntry, 0, len(entries))
	for _, e := range entries {
		if e.TotalAbove > 0 {
			nz = append(nz, e)
		}
	}

	return nz
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
	if len(e) == 0 {
		return 0, nil
	}

	extensions := make([]string, 0, len(e))
	sizes := make([]string, 0, len(e))
	stats := make([]string, 0, len(e))
	percentages := make([]string, 0, len(e))

	for _, entry := range e {
		bytesAbove := uint64(entry.BytesAbove)
		above := entry.TotalAbove
		total := entry.Total
		percentAbove := 100 * (float64(above) / float64(total))

		var size string
		if migrateInfoUnit > 0 {
			size = humanize.FormatBytesUnit(bytesAbove, migrateInfoUnit)
		} else {
			size = humanize.FormatBytes(bytesAbove)
		}

		stat := fmt.Sprintf("%d/%d files(s)",
			above, total)

		percentage := fmt.Sprintf("%.0f%%", percentAbove)

		extensions = append(extensions, entry.Qualifier)
		sizes = append(sizes, size)
		stats = append(stats, stat)
		percentages = append(percentages, percentage)
	}

	extensions = tools.Ljust(extensions)
	sizes = tools.Ljust(sizes)
	stats = tools.Rjust(stats)
	percentages = tools.Rjust(percentages)

	output := make([]string, 0, len(e))
	for i := 0; i < len(e); i++ {
		extension := extensions[i]
		size := sizes[i]
		stat := stats[i]
		percentage := percentages[i]

		line := strings.Join([]string{extension, size, stat, percentage}, "\t")

		output = append(output, line)
	}

	return fmt.Fprintln(to, strings.Join(output, "\n"))
}
