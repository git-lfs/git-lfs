package commands

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	porcelain  = false
	statusJson = false
)

func statusCommand(cmd *cobra.Command, args []string) {
	requireInRepo()
	requireWorkingCopy()

	// tolerate errors getting ref so this works before first commit
	ref, _ := git.CurrentRef()

	scanIndexAt := "HEAD"
	if ref == nil {
		scanIndexAt = git.RefBeforeFirstCommit
	}

	scanner, err := lfs.NewPointerScanner(cfg.OSEnv())
	if err != nil {
		ExitWithError(err)
	}

	if porcelain {
		porcelainStagedPointers(scanIndexAt)
		return
	} else if statusJson {
		jsonStagedPointers(scanner, scanIndexAt)
		return
	}

	statusScanRefRange(ref)

	staged, unstaged, err := scanIndex(scanIndexAt)
	if err != nil {
		ExitWithError(err)
	}

	wd, _ := os.Getwd()
	repo := cfg.LocalWorkingDir()

	wd = tools.ResolveSymlinks(wd)

	Print("\nGit LFS objects to be committed:\n")
	for _, entry := range staged {
		// Find a path from the current working directory to the
		// absolute path of each side of the entry.
		src := relativize(wd, filepath.Join(repo, entry.SrcName))
		dst := relativize(wd, filepath.Join(repo, entry.DstName))

		switch entry.Status {
		case lfs.StatusRename, lfs.StatusCopy:
			Print("\t%s -> %s (%s)", src, dst, formatBlobInfo(scanner, entry))
		default:
			Print("\t%s (%s)", src, formatBlobInfo(scanner, entry))
		}
	}

	Print("\nGit LFS objects not staged for commit:\n")
	for _, entry := range unstaged {
		src := relativize(wd, filepath.Join(repo, entry.SrcName))

		Print("\t%s (%s)", src, formatBlobInfo(scanner, entry))
	}

	Print("")

	if err = scanner.Close(); err != nil {
		ExitWithError(err)
	}
}

var z40 = regexp.MustCompile(`\^?0{40}`)

func formatBlobInfo(s *lfs.PointerScanner, entry *lfs.DiffIndexEntry) string {
	fromSha, fromSrc, err := blobInfoFrom(s, entry)
	if err != nil {
		ExitWithError(err)
	}

	from := fmt.Sprintf("%s: %s", fromSrc, fromSha)
	if entry.Status == lfs.StatusAddition {
		return from
	}

	toSha, toSrc, err := blobInfoTo(s, entry)
	if err != nil {
		ExitWithError(err)
	}
	to := fmt.Sprintf("%s: %s", toSrc, toSha)

	return fmt.Sprintf("%s -> %s", from, to)
}

func blobInfoFrom(s *lfs.PointerScanner, entry *lfs.DiffIndexEntry) (sha, from string, err error) {
	var blobSha string = entry.SrcSha
	if z40.MatchString(blobSha) {
		blobSha = entry.DstSha
	}

	return blobInfo(s, blobSha, entry.SrcName)
}

func blobInfoTo(s *lfs.PointerScanner, entry *lfs.DiffIndexEntry) (sha, from string, err error) {
	var name string = entry.DstName
	if len(name) == 0 {
		name = entry.SrcName
	}

	return blobInfo(s, entry.DstSha, name)
}

func blobInfo(s *lfs.PointerScanner, blobSha, name string) (sha, from string, err error) {
	if !z40.MatchString(blobSha) {
		s.Scan(blobSha)
		if err := s.Err(); err != nil {
			if git.IsMissingObject(err) {
				return "<missing>", "?", nil
			}
			return "", "", err
		}

		var from string
		if s.Pointer() != nil {
			from = "LFS"
		} else {
			from = "Git"
		}

		return s.ContentsSha()[:7], from, nil
	}

	f, err := os.Open(filepath.Join(cfg.LocalWorkingDir(), name))
	if os.IsNotExist(err) {
		return "deleted", "File", nil
	}
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	// We've replaced a file with a directory.
	if fi, err := f.Stat(); err == nil && fi.Mode().IsDir() {
		return "deleted", "File", nil
	}

	shasum := sha256.New()
	if _, err = io.Copy(shasum, f); err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%x", shasum.Sum(nil))[:7], "File", nil
}

func scanIndex(ref string) (staged, unstaged []*lfs.DiffIndexEntry, err error) {
	uncached, err := lfs.NewDiffIndexScanner(ref, false)
	if err != nil {
		return nil, nil, err
	}

	cached, err := lfs.NewDiffIndexScanner(ref, true)
	if err != nil {
		return nil, nil, err
	}

	seenNames := make(map[string]struct{}, 0)

	staged, err = drainScanner(seenNames, cached)
	if err != nil {
		return nil, nil, err
	}

	unstaged, err = drainScanner(seenNames, uncached)
	if err != nil {
		return nil, nil, err
	}

	return
}

func drainScanner(cache map[string]struct{}, scanner *lfs.DiffIndexScanner) ([]*lfs.DiffIndexEntry, error) {
	var to []*lfs.DiffIndexEntry

	for scanner.Scan() {
		entry := scanner.Entry()

		key := keyFromEntry(entry)
		if _, seen := cache[key]; !seen {
			to = append(to, entry)

			cache[key] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return to, nil
}

func keyFromEntry(e *lfs.DiffIndexEntry) string {
	var name string = e.DstName
	if len(name) == 0 {
		name = e.SrcName
	}

	return strings.Join([]string{e.SrcSha, e.DstSha, name}, ":")
}

func statusScanRefRange(ref *git.Ref) {
	if ref == nil {
		return
	}

	Print("On branch %s", ref.Name)

	remoteRef, err := cfg.GitConfig().CurrentRemoteRef()
	if err != nil {
		return
	}

	gitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, "Could not scan for Git LFS objects")
			return
		}

		Print("\t%s (%s)", p.Name, p.Oid)
	})
	defer gitscanner.Close()

	Print("Git LFS objects to be pushed to %s:\n", remoteRef.Name)
	if err := gitscanner.ScanRefRange(ref.Sha, remoteRef.Sha, nil); err != nil {
		Panic(err, "Could not scan for Git LFS objects")
	}

}

type JSONStatusEntry struct {
	Status string `json:"status"`
	From   string `json:"from,omitempty"`
}

type JSONStatus struct {
	Files map[string]JSONStatusEntry `json:"files"`
}

func jsonStagedPointers(scanner *lfs.PointerScanner, ref string) {
	staged, unstaged, err := scanIndex(ref)
	if err != nil {
		ExitWithError(err)
	}

	status := JSONStatus{Files: make(map[string]JSONStatusEntry)}

	for _, entry := range append(unstaged, staged...) {
		_, fromSrc, err := blobInfoFrom(scanner, entry)
		if err != nil {
			ExitWithError(err)
		}

		if fromSrc != "LFS" {
			continue
		}

		switch entry.Status {
		case lfs.StatusRename, lfs.StatusCopy:
			status.Files[entry.DstName] = JSONStatusEntry{
				Status: string(entry.Status), From: entry.SrcName,
			}
		default:
			status.Files[entry.SrcName] = JSONStatusEntry{
				Status: string(entry.Status),
			}
		}
	}

	ret, err := json.Marshal(status)
	if err != nil {
		ExitWithError(err)
	}
	Print(string(ret))
}

func porcelainStagedPointers(ref string) {
	staged, unstaged, err := scanIndex(ref)
	if err != nil {
		ExitWithError(err)
	}

	seenNames := make(map[string]struct{})

	for _, entry := range append(unstaged, staged...) {
		name := entry.DstName
		if len(name) == 0 {
			name = entry.SrcName
		}

		if _, seen := seenNames[name]; !seen {
			Print(porcelainStatusLine(entry))

			seenNames[name] = struct{}{}
		}
	}
}

func porcelainStatusLine(entry *lfs.DiffIndexEntry) string {
	switch entry.Status {
	case lfs.StatusRename, lfs.StatusCopy:
		return fmt.Sprintf("%s  %s -> %s", entry.Status, entry.SrcName, entry.DstName)
	case lfs.StatusModification:
		return fmt.Sprintf(" %s %s", entry.Status, entry.SrcName)
	}

	return fmt.Sprintf("%s  %s", entry.Status, entry.SrcName)
}

// relativize relatives a path from "from" to "to". For instance, note that, for
// any paths "from" and "to", that:
//
//   to == filepath.Clean(filepath.Join(from, relativize(from, to)))
func relativize(from, to string) string {
	if len(from) == 0 {
		return to
	}

	flist := strings.Split(filepath.ToSlash(from), "/")
	tlist := strings.Split(filepath.ToSlash(to), "/")

	var (
		divergence int
		min        int
	)

	if lf, lt := len(flist), len(tlist); lf < lt {
		min = lf
	} else {
		min = lt
	}

	for ; divergence < min; divergence++ {
		if flist[divergence] != tlist[divergence] {
			break
		}
	}

	return strings.Repeat("../", len(flist)-divergence) +
		strings.Join(tlist[divergence:], "/")
}

func init() {
	RegisterCommand("status", statusCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
		cmd.Flags().BoolVarP(&statusJson, "json", "j", false, "Give the output in a stable json format for scripts.")
	})
}
