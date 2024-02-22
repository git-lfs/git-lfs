package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools/humanize"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	longOIDs            = false
	lsFilesScanAll      = false
	lsFilesScanDeleted  = false
	lsFilesShowSize     = false
	lsFilesShowNameOnly = false
	lsFilesJSON         = false
	debug               = false
)

type lsFilesObject struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Checkout   bool   `json:"checkout"`
	Downloaded bool   `json:"downloaded"`
	OidType    string `json:"oid_type"`
	Oid        string `json:"oid"`
	Version    string `json:"version"`
}

func lsFilesCommand(cmd *cobra.Command, args []string) {
	setupRepository()

	var ref string
	var includeRef string
	var scanRange = false
	if len(args) > 0 {
		if lsFilesScanAll {
			Exit(tr.Tr.Get("Cannot use --all with explicit reference"))
		} else if args[0] == "--all" {
			// Since --all is a valid argument to "git rev-parse",
			// if we try to give it to git.ResolveRef below, we'll
			// get an unexpected result.
			//
			// So, let's check early that the caller invoked the
			// command correctly.
			Exit(tr.Tr.Get("Did you mean `git lfs ls-files --all --` ?"))
		}

		ref = args[0]
		if len(args) > 1 {
			if lsFilesScanDeleted {
				Exit(tr.Tr.Get("Cannot use --deleted with reference range"))
			}
			includeRef = args[1]
			scanRange = true
		}
	} else {
		fullref, err := git.CurrentRef()
		if err != nil {
			ref, err = git.EmptyTree()
			if err != nil {
				ExitWithError(errors.Wrap(
					err, tr.Tr.Get("Could not read empty Git tree object")))
			}
		} else {
			ref = fullref.Sha
		}
	}

	showOidLen := 10
	if longOIDs {
		showOidLen = 64
	}

	seen := make(map[string]struct{})
	var items []lsFilesObject

	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Exit(tr.Tr.Get("Could not scan for Git LFS tree: %s", err))
			return
		}

		if p.Size == 0 {
			return
		}

		if !lsFilesScanAll && !scanRange {
			if _, ok := seen[p.Name]; ok {
				return
			}
		}

		if debug {
			// TRANSLATORS: these strings should have the colons
			// aligned in a column.
			Print(
				tr.Tr.Get("filepath: %s\n    size: %d\ncheckout: %v\ndownload: %v\n     oid: %s %s\n version: %s\n",
					p.Name,
					p.Size,
					fileExistsOfSize(p),
					cfg.LFSObjectExists(p.Oid, p.Size),
					p.OidType,
					p.Oid,
					p.Version))
		} else if lsFilesJSON {
			items = append(items, lsFilesObject{
				Name:       p.Name,
				Size:       p.Size,
				Checkout:   fileExistsOfSize(p),
				Downloaded: cfg.LFSObjectExists(p.Oid, p.Size),
				OidType:    p.OidType,
				Oid:        p.Oid,
				Version:    p.Version,
			})
		} else {
			msg := []string{p.Oid[:showOidLen], lsFilesMarker(p), p.Name}
			if lsFilesShowNameOnly {
				msg = []string{p.Name}
			}
			if lsFilesShowSize {
				size := humanize.FormatBytes(uint64(p.Size))
				msg = append(msg, "("+size+")")
			}

			Print(strings.Join(msg, " "))
		}

		seen[p.Name] = struct{}{}
	})

	includeArg, excludeArg := getIncludeExcludeArgs(cmd)
	gitscanner.Filter = buildFilepathFilter(cfg, includeArg, excludeArg, false)

	if len(args) == 0 {
		// Only scan the index when "git lfs ls-files" was invoked with
		// no arguments.
		//
		// Do so to avoid showing "mixed" results, e.g., ls-files output
		// from a specific historical revision, and the index.
		if err := gitscanner.ScanIndex(ref, "", nil); err != nil {
			Exit(tr.Tr.Get("Could not scan for Git LFS index: %s", err))
		}
	}
	if lsFilesScanAll {
		if err := gitscanner.ScanAll(nil); err != nil {
			Exit(tr.Tr.Get("Could not scan for Git LFS history: %s", err))
		}
	} else {
		var err error
		if lsFilesScanDeleted {
			err = gitscanner.ScanRefWithDeleted(ref, nil)
		} else if scanRange {
			err = gitscanner.ScanRefRange(includeRef, ref, nil)
		} else {
			err = gitscanner.ScanTree(ref, nil)
		}

		if err != nil {
			Exit(tr.Tr.Get("Could not scan for Git LFS tree: %s", err))
		}
	}
	if lsFilesJSON {
		data := struct {
			Files []lsFilesObject `json:"files"`
		}{Files: items}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", " ")
		if err := encoder.Encode(data); err != nil {
			ExitWithError(err)
		}
	}
}

// Returns true if a pointer appears to be properly smudge on checkout
func fileExistsOfSize(p *lfs.WrappedPointer) bool {
	path := cfg.Filesystem().DecodePathname(p.Name)
	info, err := os.Stat(filepath.Join(cfg.LocalWorkingDir(), path))
	return err == nil && info.Size() == p.Size
}

func lsFilesMarker(p *lfs.WrappedPointer) string {
	if fileExistsOfSize(p) {
		return "*"
	}
	return "-"
}

func init() {
	RegisterCommand("ls-files", lsFilesCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&longOIDs, "long", "l", false, "")
		cmd.Flags().BoolVarP(&lsFilesShowSize, "size", "s", false, "")
		cmd.Flags().BoolVarP(&lsFilesShowNameOnly, "name-only", "n", false, "")
		cmd.Flags().BoolVarP(&debug, "debug", "d", false, "")
		cmd.Flags().BoolVarP(&lsFilesScanAll, "all", "a", false, "")
		cmd.Flags().BoolVar(&lsFilesScanDeleted, "deleted", false, "")
		cmd.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")
		cmd.Flags().BoolVarP(&lsFilesJSON, "json", "", false, "print output in JSON")
	})
}
