package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	fsckDryRun   bool
	fsckObjects  bool
	fsckPointers bool
)

type corruptPointer struct {
	blobOid string
	treeOid string
	lfsOid  string
	path    string
	message string
	kind    string
}

func (p corruptPointer) String() string {
	return fmt.Sprintf("%s: %s", p.kind, p.message)
}

// TODO(zeroshirts): 'git fsck' reports status (percentage, current#/total) as
// it checks... we should do the same, as we are rehashing potentially gigs and
// gigs of content.
//
// NOTE(zeroshirts): Ideally git would have hooks for fsck such that we could
// chain a lfs-fsck, but I don't think it does.
func fsckCommand(cmd *cobra.Command, args []string) {
	installHooks(false)
	setupRepository()

	useIndex := false
	exclude := ""
	include := "HEAD"

	switch len(args) {
	case 0:
		useIndex = true
		ref, err := git.CurrentRef()
		if err != nil {
			ExitWithError(err)
		}
		include = ref.Sha
	case 1:
		pieces := strings.SplitN(args[0], "..", 2)
		refs, err := git.ResolveRefs(pieces)
		if err != nil {
			ExitWithError(err)
		}
		if len(refs) == 2 {
			exclude = refs[0].Sha
			include = refs[1].Sha
		} else {
			include = refs[0].Sha
		}
	}

	if !fsckPointers && !fsckObjects {
		fsckPointers = true
		fsckObjects = true
	}

	ok := true
	var corruptOids []string
	var corruptPointers []corruptPointer
	if fsckObjects {
		corruptOids = doFsckObjects(include, exclude, useIndex)
		ok = ok && len(corruptOids) == 0
	}
	if fsckPointers {
		corruptPointers = doFsckPointers(include, exclude)
		ok = ok && len(corruptPointers) == 0
	}

	if ok {
		Print(tr.Tr.Get("Git LFS fsck OK"))
		return
	}

	if fsckDryRun || len(corruptOids) == 0 {
		os.Exit(1)
	}

	badDir := filepath.Join(cfg.LFSStorageDir(), "bad")
	Print("objects: repair: %s", tr.Tr.Get("moving corrupt objects to %s", badDir))

	if err := tools.MkdirAll(badDir, cfg); err != nil {
		ExitWithError(err)
	}

	for _, oid := range corruptOids {
		badFile := filepath.Join(badDir, oid)
		srcFile := cfg.Filesystem().ObjectPathname(oid)
		if srcFile == os.DevNull {
			continue
		}
		if err := os.Rename(srcFile, badFile); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			ExitWithError(err)
		}
	}
	os.Exit(1)
}

// doFsckObjects checks that the objects in the given ref are correct and exist.
func doFsckObjects(include, exclude string, useIndex bool) []string {
	var corruptOids []string
	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err == nil {
			var pointerOk bool
			pointerOk, err = fsckPointer(p.Name, p.Oid, p.Size)
			if !pointerOk {
				corruptOids = append(corruptOids, p.Oid)
			}
		}

		if err != nil {
			Panic(err, tr.Tr.Get("Error checking Git LFS files"))
		}
	})

	// If 'lfs.fetchexclude' is set and 'git lfs fsck' is run after the
	// initial fetch (i.e., has elected to fetch a subset of Git LFS
	// objects), the "missing" ones will fail the fsck.
	//
	// Attach a filepathfilter to avoid _only_ the excluded paths.
	gitscanner.Filter = filepathfilter.New(nil, cfg.FetchExcludePaths(), filepathfilter.GitIgnore)

	if exclude == "" {
		if err := gitscanner.ScanRef(include, nil); err != nil {
			ExitWithError(err)
		}
	} else {
		if err := gitscanner.ScanRefRange(include, exclude, nil); err != nil {
			ExitWithError(err)
		}
	}

	if useIndex {
		if err := gitscanner.ScanIndex("HEAD", "", nil); err != nil {
			ExitWithError(err)
		}
	}

	return corruptOids
}

// doFsckPointers checks that the pointers in the given ref are correct and canonical.
func doFsckPointers(include, exclude string) []corruptPointer {
	var corruptPointers []corruptPointer
	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if p != nil {
			Debug(tr.Tr.Get("Examining %v (%v)", p.Oid, p.Name))
			if !p.Canonical {
				cp := corruptPointer{
					blobOid: p.Sha1,
					lfsOid:  p.Oid,
					message: tr.Tr.Get("Pointer for %s (blob %s) was not canonical", p.Oid, p.Sha1),
					kind:    "nonCanonicalPointer",
				}
				Print("pointer: %s", cp.String())
				corruptPointers = append(corruptPointers, cp)
			}
		} else if errors.IsPointerScanError(err) {
			psErr, ok := err.(errors.PointerScanError)
			if ok {
				cp := corruptPointer{
					treeOid: psErr.OID(),
					path:    psErr.Path(),
					message: tr.Tr.Get("%q (treeish %s) should have been a pointer but was not", psErr.Path(), psErr.OID()),
					kind:    "unexpectedGitObject",
				}
				Print("pointer: %s", cp.String())
				corruptPointers = append(corruptPointers, cp)
			}
		} else {
			Panic(err, tr.Tr.Get("Error checking Git LFS files"))
		}
	})

	if exclude == "" {
		if err := gitscanner.ScanRefByTree(include, nil); err != nil {
			ExitWithError(err)
		}
	} else {
		if err := gitscanner.ScanRefRangeByTree(include, exclude, nil); err != nil {
			ExitWithError(err)
		}
	}

	return corruptPointers
}

func fsckPointer(name, oid string, size int64) (bool, error) {
	path := cfg.Filesystem().ObjectPathname(oid)

	Debug(tr.Tr.Get("Examining %v (%v)", name, path))

	f, err := os.Open(path)
	if pErr, pOk := err.(*os.PathError); pOk {
		// This is an empty file.  No problem here.
		if size == 0 {
			return true, nil
		}
		Print("objects: openError: %s", tr.Tr.Get("%s (%s) could not be checked: %s", name, oid, pErr.Err))
		return false, nil
	}

	if err != nil {
		return false, err
	}

	oidHash := sha256.New()
	_, err = io.Copy(oidHash, f)
	f.Close()
	if err != nil {
		return false, err
	}

	recalculatedOid := hex.EncodeToString(oidHash.Sum(nil))
	if recalculatedOid == oid {
		return true, nil
	}

	Print(fmt.Sprintf("objects: corruptObject: %s", tr.Tr.Get("%s (%s) is corrupt", name, oid)))
	return false, nil
}

func init() {
	RegisterCommand("fsck", fsckCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fsckDryRun, "dry-run", "d", false, "List corrupt objects without deleting them.")
		cmd.Flags().BoolVarP(&fsckObjects, "objects", "", false, "Fsck objects.")
		cmd.Flags().BoolVarP(&fsckPointers, "pointers", "", false, "Fsck pointers.")
	})
}
