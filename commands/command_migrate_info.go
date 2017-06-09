package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var MigrateInfoCommand = NewCommand("info", migrateInfoCommand)

var (
	// migrateInfoTopN is a flag given to the git-lfs-migrate(1) subcommand
	// 'info' which specifies how many info entries to show by default.
	migrateInfoTopN int
)

func migrateInfoCommand(cmd *cobra.Command, args []string) {
	exts := make(map[string]int64)

	migrate(cmd, args, func(path string, b *odb.Blob) (*odb.Blob, error) {
		ext := fmt.Sprintf("*%s", filepath.Ext(path))
		if len(ext) > 1 {
			exts[ext] = exts[ext] + b.Size
		}

		return b, nil
	})

	entries := EntriesBySize(MapToEntries(exts))
	sort.Sort(sort.Reverse(entries))

	migrateInfoTopN = tools.ClampInt(migrateInfoTopN, 0, len(entries))

	entries = entries[:tools.MaxInt(0, migrateInfoTopN)]

	entries.Print(os.Stderr)
}

// MigrateInfoEntry represents a tuple of filetype to total size taken by that
// file type.
type MigrateInfoEntry struct {
	// Qualifier is the filepath's extension.
	Qualifier string
	// Size is the total size in bytes taken by that filepath's extension.
	Size int64
}

// MapToEntries creates a set of `*MigrateInfoEntry`'s for a given map of
// filepath extensions to file size in bytes.
func MapToEntries(exts map[string]int64) []*MigrateInfoEntry {
	entries := make([]*MigrateInfoEntry, 0, len(exts))
	for qualifier, size := range exts {
		entries = append(entries, &MigrateInfoEntry{
			Qualifier: qualifier,
			Size:      size,
		})
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
func (e EntriesBySize) Less(i, j int) bool { return e[i].Size < e[j].Size }

// Swap swaps the entries given at i, j.
func (e EntriesBySize) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

// LongestQualifier returns the string length of the longest qualifier in the
// set.
func (e EntriesBySize) LongestQualifier() int {
	var longest int
	for _, entry := range e {
		longest = tools.MaxInt(longest, len(entry.Qualifier))
	}

	return longest
}

// Print formats the `*MigrateInfoEntry`'s in the set and prints them to the
// given io.Writer, "to", returning "n" the number of bytes written, and any
// error, if one occurred.
func (e EntriesBySize) Print(to io.Writer) (int, error) {
	output := make([]string, 0, len(e))

	longest := e.LongestQualifier()

	for _, entry := range e {
		offset := longest - len(entry.Qualifier)
		padding := strings.Repeat(" ", tools.MaxInt(0, offset))
		bytes := humanizeBytes(entry.Size)

		line := fmt.Sprintf("%s%s %s", entry.Qualifier, padding, bytes)

		output = append(output, line)
	}

	return fmt.Fprintln(to, strings.Join(output, "\n"))
}

func init() {
	MigrateInfoCommand.Flags().IntVar(&migrateInfoTopN, "top", 5, "--top=<n>")
}
