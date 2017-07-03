package commands

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	porcelain  = false
	statusJson = false
)

func statusCommand(cmd *cobra.Command, args []string) {
	requireInRepo()

	// tolerate errors getting ref so this works before first commit
	ref, _ := git.CurrentRef()

	scanIndexAt := "HEAD"
	if ref == nil {
		scanIndexAt = git.RefBeforeFirstCommit
	}

	if porcelain {
		porcelainStagedPointers(scanIndexAt)
		return
	} else if statusJson {
		jsonStagedPointers(scanIndexAt)
		return
	}

	statusScanRefRange(ref)

	staged, unstaged, err := scanIndex(scanIndexAt)
	if err != nil {
		ExitWithError(err)
	}

	scanner, err := lfs.NewPointerScanner()
	if err != nil {
		scanner.Close()

		ExitWithError(err)
	}

	Print("\nGit LFS objects to be committed:\n")
	for _, entry := range staged {
		switch entry.Status {
		case lfs.StatusRename, lfs.StatusCopy:
			Print("\t%s -> %s (%s)", entry.SrcName, entry.DstName, formatBlobInfo(scanner, entry))
		default:
			Print("\t%s (%s)", entry.SrcName, formatBlobInfo(scanner, entry))
		}
	}

	Print("\nGit LFS objects not staged for commit:\n")
	for _, entry := range unstaged {
		Print("\t%s (%s)", entry.SrcName, formatBlobInfo(scanner, entry))
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

	from := fmt.Sprintf("%s: %s", fromSrc, fromSha[:7])
	if entry.Status == lfs.StatusAddition {
		return from
	}

	toSha, toSrc, err := blobInfoTo(s, entry)
	if err != nil {
		ExitWithError(err)
	}
	to := fmt.Sprintf("%s: %s", toSrc, toSha[:7])

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
			return "", "", err
		}

		var from string
		if s.Pointer() != nil {
			from = "LFS"
		} else {
			from = "Git"
		}

		return s.ContentsSha(), from, nil
	}

	f, err := os.Open(name)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	shasum := sha256.New()
	if _, err = io.Copy(shasum, f); err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%x", shasum.Sum(nil)), "File", nil
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

	remoteRef, err := git.CurrentRemoteRef()
	if err != nil {
		return
	}

	gitscanner := lfs.NewGitScanner(func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			Panic(err, "Could not scan for Git LFS objects")
			return
		}

		Print("\t%s (%s)", p.Name)
	})
	defer gitscanner.Close()

	Print("Git LFS objects to be pushed to %s:\n", remoteRef.Name)
	if err := gitscanner.ScanRefRange(ref.Sha, "^"+remoteRef.Sha, nil); err != nil {
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

func jsonStagedPointers(ref string) {
	staged, unstaged, err := scanIndex(ref)
	if err != nil {
		ExitWithError(err)
	}

	status := JSONStatus{Files: make(map[string]JSONStatusEntry)}

	for _, entry := range append(unstaged, staged...) {
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

func init() {
	RegisterCommand("status", statusCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&porcelain, "porcelain", "p", false, "Give the output in an easy-to-parse format for scripts.")
		cmd.Flags().BoolVarP(&statusJson, "json", "j", false, "Give the output in a stable json format for scripts.")
	})
}
