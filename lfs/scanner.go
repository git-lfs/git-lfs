package lfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

const (
	// blobSizeCutoff is used to determine which files to scan for Git LFS
	// pointers.  Any file with a size below this cutoff will be scanned.
	blobSizeCutoff = 1024

	// stdoutBufSize is the size of the buffers given to a sub-process stdout
	stdoutBufSize = 16384

	// chanBufSize is the size of the channels used to pass data from one
	// sub-process to another.
	chanBufSize = 100
)

var (
	// Arguments to append to a git log call which will limit the output to
	// lfs changes and format the output suitable for parseLogOutput.. method(s)
	logLfsSearchArgs = []string{
		"-G", "oid sha256:", // only diffs which include an lfs file SHA change
		"-p",   // include diff so we can read the SHA
		"-U12", // Make sure diff context is always big enough to support 10 extension lines to get whole pointer
		`--format=lfs-commit-sha: %H %P`, // just a predictable commit header we can detect
	}
)

// WrappedPointer wraps a pointer.Pointer and provides the git sha1
// and the file name associated with the object, taken from the
// rev-list output.
type WrappedPointer struct {
	Sha1    string
	Name    string
	SrcName string
	Size    int64
	Status  string
	*Pointer
}

// indexFile is used when scanning the index. It stores the name of
// the file, the status of the file in the index, and, in the case of
// a moved or copied file, the original name of the file.
type indexFile struct {
	Name    string
	SrcName string
	Status  string
}

var z40 = regexp.MustCompile(`\^?0{40}`)

type ScanningMode int

const (
	ScanRefsMode         = ScanningMode(iota) // 0 - or default scan mode
	ScanAllMode          = ScanningMode(iota)
	ScanLeftToRemoteMode = ScanningMode(iota)
)

type ScanRefsOptions struct {
	ScanMode         ScanningMode
	RemoteName       string
	SkipDeletedBlobs bool
	nameMap          map[string]string
	mutex            *sync.Mutex
}

func (o *ScanRefsOptions) GetName(sha string) (string, bool) {
	o.mutex.Lock()
	name, ok := o.nameMap[sha]
	o.mutex.Unlock()
	return name, ok
}

func (o *ScanRefsOptions) SetName(sha, name string) {
	o.mutex.Lock()
	o.nameMap[sha] = name
	o.mutex.Unlock()
}

func NewScanRefsOptions() *ScanRefsOptions {
	return &ScanRefsOptions{
		nameMap: make(map[string]string, 0),
		mutex:   &sync.Mutex{},
	}
}

// ScanRefs takes a ref and returns a slice of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanRefs(refLeft, refRight string, opt *ScanRefsOptions) ([]*WrappedPointer, error) {
	s, err := ScanRefsToChan(refLeft, refRight, opt)
	if err != nil {
		return nil, err
	}
	pointers := make([]*WrappedPointer, 0)
	for p := range s.Results {
		pointers = append(pointers, p)
	}
	err = s.Wait()

	return pointers, err

}

// ScanRefsToChan takes a ref and returns a channel of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanRefsToChan(refLeft, refRight string, opt *ScanRefsOptions) (*PointerChannelWrapper, error) {
	if opt == nil {
		opt = NewScanRefsOptions()
	}
	if refLeft == "" {
		opt.ScanMode = ScanAllMode
	}

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan", start)
	}()

	revs, err := revListShas(refLeft, refRight, opt)
	if err != nil {
		return nil, err
	}

	smallShas, err := catFileBatchCheck(revs)
	if err != nil {
		return nil, err
	}

	pointers, err := catFileBatch(smallShas)
	if err != nil {
		return nil, err
	}

	retchan := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 1)
	go func() {
		for p := range pointers.Results {
			if name, ok := opt.GetName(p.Sha1); ok {
				p.Name = name
			}
			retchan <- p
		}
		err := pointers.Wait()
		if err != nil {
			errchan <- err
		}
		close(retchan)
		close(errchan)
	}()

	return NewPointerChannelWrapper(retchan, errchan), nil
}

type indexFileMap struct {
	// mutex guards nameMap and nameShaPairs
	mutex *sync.Mutex
	// nameMap maps SHA1s to a slice of `*indexFile`s
	nameMap map[string][]*indexFile
	// nameShaPairs maps "sha1:name" -> bool
	nameShaPairs map[string]bool
}

// FilesFor returns all `*indexFile`s that match the given `sha`.
func (m *indexFileMap) FilesFor(sha string) []*indexFile {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.nameMap[sha]
}

// Add appends unique index files to the given SHA, "sha". A file is considered
// unique if its combination of SHA and current filename have not yet been seen
// by this instance "m" of *indexFileMap.
func (m *indexFileMap) Add(sha string, index *indexFile) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pairKey := strings.Join([]string{sha, index.Name}, ":")
	if m.nameShaPairs[pairKey] {
		return
	}

	m.nameMap[sha] = append(m.nameMap[sha], index)
	m.nameShaPairs[pairKey] = true
}

// ScanIndex returns a slice of WrappedPointer objects for all Git LFS pointers
// it finds in the index.
//
// Ref is the ref at which to scan, which may be "HEAD" if there is at least one
// commit.
func ScanIndex(ref string) ([]*WrappedPointer, error) {
	indexMap := &indexFileMap{
		nameMap:      make(map[string][]*indexFile),
		nameShaPairs: make(map[string]bool),
		mutex:        &sync.Mutex{},
	}

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan-staging", start)
	}()

	revs, err := revListIndex(ref, false, indexMap)
	if err != nil {
		return nil, err
	}

	cachedRevs, err := revListIndex(ref, true, indexMap)
	if err != nil {
		return nil, err
	}

	allRevsErr := make(chan error, 5) // can be multiple errors below
	allRevsChan := make(chan string, 1)
	allRevs := NewStringChannelWrapper(allRevsChan, allRevsErr)
	go func() {
		seenRevs := make(map[string]bool, 0)

		for rev := range revs.Results {
			if !seenRevs[rev] {
				allRevsChan <- rev
				seenRevs[rev] = true
			}
		}
		err := revs.Wait()
		if err != nil {
			allRevsErr <- err
		}

		for rev := range cachedRevs.Results {
			if !seenRevs[rev] {
				allRevsChan <- rev
				seenRevs[rev] = true
			}
		}
		err = cachedRevs.Wait()
		if err != nil {
			allRevsErr <- err
		}
		close(allRevsChan)
		close(allRevsErr)
	}()

	smallShas, err := catFileBatchCheck(allRevs)
	if err != nil {
		return nil, err
	}

	pointerc, err := catFileBatch(smallShas)
	if err != nil {
		return nil, err
	}

	pointers := make([]*WrappedPointer, 0)
	for p := range pointerc.Results {
		for _, file := range indexMap.FilesFor(p.Sha1) {
			// Append a new *WrappedPointer that combines the data
			// from the index file, and the pointer "p".
			pointers = append(pointers, &WrappedPointer{
				Sha1:    p.Sha1,
				Name:    file.Name,
				SrcName: file.SrcName,
				Status:  file.Status,
				Size:    p.Size,
				Pointer: p.Pointer,
			})
		}
	}
	err = pointerc.Wait()

	return pointers, err

}

// Get additional arguments needed to limit 'git rev-list' to just the changes
// in revTo that are also not on remoteName.
//
// Returns a slice of string command arguments, and a slice of string git
// commits to pass to `git rev-list` via STDIN.
func revListArgsRefVsRemote(refTo, remoteName string) ([]string, []string) {
	// We need to check that the locally cached versions of remote refs are still
	// present on the remote before we use them as a 'from' point. If the
	// server implements garbage collection and a remote branch had been deleted
	// since we last did 'git fetch --prune', then the objects in that branch may
	// have also been deleted on the server if unreferenced.
	// If some refs are missing on the remote, use a more explicit diff

	cachedRemoteRefs, _ := git.CachedRemoteRefs(remoteName)
	actualRemoteRefs, _ := git.RemoteRefs(remoteName)

	// Only check for missing refs on remote; if the ref is different it has moved
	// forward probably, and if not and the ref has changed to a non-descendant
	// (force push) then that will cause a re-evaluation in a subsequent command anyway
	missingRefs := tools.NewStringSet()
	for _, cachedRef := range cachedRemoteRefs {
		found := false
		for _, realRemoteRef := range actualRemoteRefs {
			if cachedRef.Type == realRemoteRef.Type && cachedRef.Name == realRemoteRef.Name {
				found = true
				break
			}
		}
		if !found {
			missingRefs.Add(cachedRef.Name)
		}
	}

	if len(missingRefs) > 0 {
		// Use only the non-missing refs as 'from' points
		commits := make([]string, 1, len(cachedRemoteRefs)+1)
		commits[0] = refTo
		for _, cachedRef := range cachedRemoteRefs {
			if !missingRefs.Contains(cachedRef.Name) {
				commits = append(commits, "^"+cachedRef.Sha)
			}
		}
		return []string{"--stdin"}, commits
	} else {
		// Safe to use cached
		return []string{refTo, "--not", "--remotes=" + remoteName}, nil
	}
}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(refLeft, refRight string, opt *ScanRefsOptions) (*StringChannelWrapper, error) {
	refArgs := []string{"rev-list", "--objects"}
	var stdin []string
	switch opt.ScanMode {
	case ScanRefsMode:
		if opt.SkipDeletedBlobs {
			refArgs = append(refArgs, "--no-walk")
		} else {
			refArgs = append(refArgs, "--do-walk")
		}

		refArgs = append(refArgs, refLeft)
		if refRight != "" && !z40.MatchString(refRight) {
			refArgs = append(refArgs, refRight)
		}
	case ScanAllMode:
		refArgs = append(refArgs, "--all")
	case ScanLeftToRemoteMode:
		args, commits := revListArgsRefVsRemote(refLeft, opt.RemoteName)
		refArgs = append(refArgs, args...)
		if len(commits) > 0 {
			stdin = commits
		}
	default:
		return nil, errors.New("scanner: unknown scan type: " + strconv.Itoa(int(opt.ScanMode)))
	}

	// Use "--" at the end of the command to disambiguate arguments as refs,
	// so Git doesn't complain about ambiguity if you happen to also have a
	// file named "master".
	refArgs = append(refArgs, "--")

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, err
	}

	if len(stdin) > 0 {
		cmd.Stdin.Write([]byte(strings.Join(stdin, "\n")))
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)
	errchan := make(chan error, 5) // may be multiple errors

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) < 40 {
				continue
			}

			sha1 := line[0:40]
			if len(line) > 40 {
				opt.SetName(sha1, line[41:len(line)])
			}
			revs <- sha1
		}

		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git rev-list --objects: %v %v", err, string(stderr))
		} else {
			// Special case detection of ambiguous refs; lower level commands like
			// git rev-list do not return non-zero exit codes in this case, just warn
			ambiguousRegex := regexp.MustCompile(`warning: refname (.*) is ambiguous`)
			if match := ambiguousRegex.FindStringSubmatch(string(stderr)); match != nil {
				// Promote to fatal & exit
				errchan <- fmt.Errorf("Error: ref %s is ambiguous", match[1])
			}
		}
		close(revs)
		close(errchan)
	}()

	return NewStringChannelWrapper(revs, errchan), nil
}

// revListIndex uses git diff-index to return the list of object sha1s
// for in the indexf. It returns a channel from which sha1 strings can be read.
// The namMap will be filled indexFile pointers mapping sha1s to indexFiles.
func revListIndex(atRef string, cache bool, indexMap *indexFileMap) (*StringChannelWrapper, error) {
	cmdArgs := []string{"diff-index", "-M"}
	if cache {
		cmdArgs = append(cmdArgs, "--cached")
	}
	cmdArgs = append(cmdArgs, atRef)

	cmd, err := startCommand("git", cmdArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			// Format is:
			// :100644 100644 c5b3d83a7542255ec7856487baa5e83d65b1624c 9e82ac1b514be060945392291b5b3108c22f6fe3 M foo.gif
			// :<old mode> <new mode> <old sha1> <new sha1> <status>\t<file name>[\t<file name>]
			line := scanner.Text()
			parts := strings.Split(line, "\t")
			if len(parts) < 2 {
				continue
			}

			description := strings.Split(parts[0], " ")
			files := parts[1:len(parts)]

			if len(description) >= 5 {
				status := description[4][0:1]
				sha1 := description[3]
				if status == "M" {
					sha1 = description[2] // This one is modified but not added
				}

				indexMap.Add(sha1, &indexFile{
					Name:    files[len(files)-1],
					SrcName: files[0],
					Status:  status,
				})
				revs <- sha1
			}
		}

		// Note: deliberately not checking result code here, because doing that
		// can fail fsck process too early since clean filter will detect errors
		// and set this to non-zero. How to cope with this better?
		// stderr, _ := ioutil.ReadAll(cmd.Stderr)
		// err := cmd.Wait()
		// if err != nil {
		// 	errchan <- fmt.Errorf("Error in git diff-index: %v %v", err, string(stderr))
		// }
		cmd.Wait()
		close(revs)
		close(errchan)
	}()

	return NewStringChannelWrapper(revs, errchan), nil
}

// catFileBatchCheck uses git cat-file --batch-check to get the type
// and size of a git object. Any object that isn't of type blob and
// under the blobSizeCutoff will be ignored. revs is a channel over
// which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func catFileBatchCheck(revs *StringChannelWrapper) (*StringChannelWrapper, error) {
	smallRevCh := make(chan string, chanBufSize)
	errCh := make(chan error, 2) // up to 2 errors, one from each goroutine
	if err := runCatFileBatchCheck(smallRevCh, revs, errCh); err != nil {
		return nil, err
	}
	return NewStringChannelWrapper(smallRevCh, errCh), nil
}

// catFileBatch uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. revs is a channel over which strings containing Git SHA1s
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatch(revs *StringChannelWrapper) (*PointerChannelWrapper, error) {
	pointerCh := make(chan *WrappedPointer, chanBufSize)
	errCh := make(chan error, 5) // shared by 2 goroutines & may add more detail errors?
	if err := runCatFileBatch(pointerCh, revs, errCh); err != nil {
		return nil, err
	}
	return NewPointerChannelWrapper(pointerCh, errCh), nil
}

type wrappedCmd struct {
	Stdin  io.WriteCloser
	Stdout *bufio.Reader
	Stderr *bufio.Reader
	*exec.Cmd
}

// startCommand starts up a command and creates a stdin pipe and a buffered
// stdout & stderr pipes, wrapped in a wrappedCmd. The stdout buffer will be of stdoutBufSize
// bytes.
func startCommand(command string, args ...string) (*wrappedCmd, error) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	tracerx.Printf("run_command: %s %s", command, strings.Join(args, " "))
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &wrappedCmd{
		stdin,
		bufio.NewReaderSize(stdout, stdoutBufSize),
		bufio.NewReaderSize(stderr, stdoutBufSize),
		cmd,
	}, nil
}

// An entry from ls-tree or rev-list including a blob sha and tree path
type TreeBlob struct {
	Sha1     string
	Filename string
}

// ScanTree takes a ref and returns a slice of WrappedPointer objects in the tree at that ref
// Differs from ScanRefs in that multiple files in the tree with the same content are all reported
func ScanTree(ref string) ([]*WrappedPointer, error) {
	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan", start)
	}()

	// We don't use the nameMap approach here since that's imprecise when >1 file
	// can be using the same content
	treeShas, err := lsTreeBlobs(ref)
	if err != nil {
		return nil, err
	}

	pointerc, err := catFileBatchTree(treeShas)
	if err != nil {
		return nil, err
	}

	pointers := make([]*WrappedPointer, 0)
	for p := range pointerc.Results {
		pointers = append(pointers, p)
	}
	err = pointerc.Wait()

	return pointers, err
}

// catFileBatchTree uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. treeblobs is a channel over which blob entries
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatchTree(treeblobs *TreeBlobChannelWrapper) (*PointerChannelWrapper, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 10) // Multiple errors possible

	go func() {
		for t := range treeblobs.Results {
			cmd.Stdin.Write([]byte(t.Sha1 + "\n"))
			l, err := cmd.Stdout.ReadBytes('\n')
			if err != nil {
				break
			}

			// Line is formatted:
			// <sha1> <type> <size>
			fields := bytes.Fields(l)
			s, _ := strconv.Atoi(string(fields[2]))

			nbuf := make([]byte, s)
			_, err = io.ReadFull(cmd.Stdout, nbuf)
			if err != nil {
				break // Legit errors
			}

			p, err := DecodePointer(bytes.NewBuffer(nbuf))
			if err == nil {
				pointers <- &WrappedPointer{
					Sha1:    string(fields[0]),
					Size:    p.Size,
					Pointer: p,
					Name:    t.Filename,
				}
			}

			_, err = cmd.Stdout.ReadBytes('\n') // Extra \n inserted by cat-file
			if err != nil {
				break
			}
		}
		// Deal with nested error from incoming treeblobs
		err := treeblobs.Wait()
		if err != nil {
			errchan <- err
		}

		cmd.Stdin.Close()

		// also errors from our command
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err = cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git cat-file: %v %v", err, string(stderr))
		}
		close(pointers)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pointers, errchan), nil
}

// Use ls-tree at ref to find a list of candidate tree blobs which might be lfs files
// The returned channel will be sent these blobs which should be sent to catFileBatchTree
// for final check & conversion to Pointer
func lsTreeBlobs(ref string) (*TreeBlobChannelWrapper, error) {
	// Snapshot using ls-tree
	lsArgs := []string{"ls-tree",
		"-r",          // recurse
		"-l",          // report object size (we'll need this)
		"-z",          // null line termination
		"--full-tree", // start at the root regardless of where we are in it
		ref}

	cmd, err := startCommand("git", lsArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	blobs := make(chan TreeBlob, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		parseLsTree(cmd.Stdout, blobs)
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git ls-tree: %v %v", err, string(stderr))
		}
		close(blobs)
		close(errchan)
	}()

	return NewTreeBlobChannelWrapper(blobs, errchan), nil
}

func parseLsTree(reader io.Reader, output chan TreeBlob) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(scanNullLines)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		attrs := strings.SplitN(parts[0], " ", 4)
		if len(attrs) < 4 {
			continue
		}

		if attrs[1] != "blob" {
			continue
		}

		sz, err := strconv.ParseInt(strings.TrimSpace(attrs[3]), 10, 64)
		if err != nil {
			continue
		}

		if sz < blobSizeCutoff {
			sha1 := attrs[2]
			filename := parts[1]
			output <- TreeBlob{sha1, filename}
		}
	}
}

func scanNullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\000'); i >= 0 {
		// We have a full null-terminated line.
		return i + 1, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

// ScanUnpushed scans history for all LFS pointers which have been added but not
// pushed to the named remote. remoteName can be left blank to mean 'any remote'
func ScanUnpushed(remoteName string) ([]*WrappedPointer, error) {

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan", start)
	}()

	pointerchan, err := ScanUnpushedToChan(remoteName)
	if err != nil {
		return nil, err
	}
	pointers := make([]*WrappedPointer, 0, 10)
	for p := range pointerchan.Results {
		pointers = append(pointers, p)
	}
	err = pointerchan.Wait()
	return pointers, err
}

// ScanPreviousVersions scans changes reachable from ref (commit) back to since.
// Returns pointers for *previous* versions that overlap that time. Does not
// return pointers which were still in use at ref (use ScanRef for that)
func ScanPreviousVersions(ref string, since time.Time) ([]*WrappedPointer, error) {
	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan", start)
	}()

	pointerchan, err := ScanPreviousVersionsToChan(ref, since)
	if err != nil {
		return nil, err
	}
	pointers := make([]*WrappedPointer, 0, 10)
	for p := range pointerchan.Results {
		pointers = append(pointers, p)
	}
	err = pointerchan.Wait()
	return pointers, err

}

// ScanPreviousVersionsToChan scans changes reachable from ref (commit) back to since.
// Returns channel of pointers for *previous* versions that overlap that time. Does not
// include pointers which were still in use at ref (use ScanRefsToChan for that)
func ScanPreviousVersionsToChan(ref string, since time.Time) (*PointerChannelWrapper, error) {
	return logPreviousSHAs(ref, since)
}

// ScanUnpushedToChan scans history for all LFS pointers which have been added but
// not pushed to the named remote. remoteName can be left blank to mean 'any remote'
// return progressively in a channel
func ScanUnpushedToChan(remoteName string) (*PointerChannelWrapper, error) {
	logArgs := []string{"log",
		"--branches", "--tags", // include all locally referenced commits
		"--not"} // but exclude everything that comes after

	if len(remoteName) == 0 {
		logArgs = append(logArgs, "--remotes")
	} else {
		logArgs = append(logArgs, fmt.Sprintf("--remotes=%v", remoteName))
	}
	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)

	cmd, err := startCommand("git", logArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	pchan := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 1)

	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffAdditions, nil, nil, pchan)
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git log: %v %v", err, string(stderr))
		}
		close(pchan)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pchan, errchan), nil

}

// logPreviousVersions scans history for all previous versions of LFS pointers
// from 'since' up to (but not including) the final state at ref
func logPreviousSHAs(ref string, since time.Time) (*PointerChannelWrapper, error) {
	logArgs := []string{"log",
		fmt.Sprintf("--since=%v", git.FormatGitDate(since)),
	}
	// Add standard search args to find lfs references
	logArgs = append(logArgs, logLfsSearchArgs...)
	// ending at ref
	logArgs = append(logArgs, ref)

	cmd, err := startCommand("git", logArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	pchan := make(chan *WrappedPointer, chanBufSize)
	errchan := make(chan error, 1)

	// we pull out deletions, since we want the previous SHAs at commits in the range
	// this means we pick up all previous versions that could have been checked
	// out in the date range, not just if the commit which *introduced* them is in the range
	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffDeletions, nil, nil, pchan)
		stderr, _ := ioutil.ReadAll(cmd.Stderr)
		err := cmd.Wait()
		if err != nil {
			errchan <- fmt.Errorf("Error in git log: %v %v", err, string(stderr))
		}
		close(pchan)
		close(errchan)
	}()

	return NewPointerChannelWrapper(pchan, errchan), nil

}

// When scanning diffs e.g. parseLogOutputToPointers, which direction of diff to include
// data from, i.e. '+' or '-'. Depending on what you're scanning for either might be useful
type LogDiffDirection byte

const (
	LogDiffAdditions = LogDiffDirection('+') // include '+' diffs
	LogDiffDeletions = LogDiffDirection('-') // include '-' diffs
)

// parseLogOutputToPointers parses log output formatted as per logLfsSearchArgs & return pointers
// log: a stream of output from git log with at least logLfsSearchArgs specified
// dir: whether to include results from + or - diffs
// includePaths, excludePaths: filter the results by filename
// results: a channel which will receive the pointers (caller must close)
func parseLogOutputToPointers(log io.Reader, dir LogDiffDirection,
	includePaths, excludePaths []string, results chan *WrappedPointer) {

	// For each commit we'll get something like this:
	/*
		lfs-commit-sha: 60fde3d23553e10a55e2a32ed18c20f65edd91e7 e2eaf1c10b57da7b98eb5d722ec5912ddeb53ea1

		diff --git a/1D_Noise.png b/1D_Noise.png
		new file mode 100644
		index 0000000..2622b4a
		--- /dev/null
		+++ b/1D_Noise.png
		@@ -0,0 +1,3 @@
		+version https://git-lfs.github.com/spec/v1
		+oid sha256:f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6
		+size 1289
	*/
	// There can be multiple diffs per commit (multiple binaries)
	// Also when a binary is changed the diff will include a '-' line for the old SHA

	// Define regexes to capture commit & diff headers
	filter := filepathfilter.New(includePaths, excludePaths)
	commitHeaderRegex := regexp.MustCompile(`^lfs-commit-sha: ([A-Fa-f0-9]{40})(?: ([A-Fa-f0-9]{40}))*`)
	fileHeaderRegex := regexp.MustCompile(`diff --git a\/(.+?)\s+b\/(.+)`)
	fileMergeHeaderRegex := regexp.MustCompile(`diff --cc (.+)`)
	pointerDataRegex := regexp.MustCompile(`^([\+\- ])(version https://git-lfs|oid sha256|size|ext-).*$`)
	var pointerData bytes.Buffer
	var currentFilename string
	currentFileIncluded := true

	// Utility func used at several points below (keep in narrow scope)
	finishLastPointer := func() {
		if pointerData.Len() > 0 {
			if currentFileIncluded {
				p, err := DecodePointer(&pointerData)
				if err == nil {
					results <- &WrappedPointer{Name: currentFilename, Size: p.Size, Pointer: p}
				} else {
					tracerx.Printf("Unable to parse pointer from log: %v", err)
				}
			}
			pointerData.Reset()
		}
	}

	scanner := bufio.NewScanner(log)
	for scanner.Scan() {
		line := scanner.Text()
		if match := commitHeaderRegex.FindStringSubmatch(line); match != nil {
			// Currently we're not pulling out commit groupings, but could if we wanted
			// This just acts as a delimiter for finishing a multiline pointer
			finishLastPointer()

		} else if match := fileHeaderRegex.FindStringSubmatch(line); match != nil {
			// Finding a regular file header
			finishLastPointer()
			// Pertinent file name depends on whether we're listening to additions or removals
			if dir == LogDiffAdditions {
				currentFilename = match[2]
			} else {
				currentFilename = match[1]
			}
			currentFileIncluded = filter.Allows(currentFilename)
		} else if match := fileMergeHeaderRegex.FindStringSubmatch(line); match != nil {
			// Git merge file header is a little different, only one file
			finishLastPointer()
			currentFilename = match[1]
			currentFileIncluded = filter.Allows(currentFilename)
		} else if currentFileIncluded {
			if match := pointerDataRegex.FindStringSubmatch(line); match != nil {
				// An LFS pointer data line
				// Include only the entirety of one side of the diff
				// -U3 will ensure we always get all of it, even if only
				// the SHA changed (version & size the same)
				changeType := match[1][0]
				// Always include unchanged context lines (normally just the version line)
				if LogDiffDirection(changeType) == dir || changeType == ' ' {
					// Must skip diff +/- marker
					pointerData.WriteString(line[1:])
					pointerData.WriteString("\n") // newline was stripped off by scanner
				}
			}
		}
	}
	// Final pointer if in progress
	finishLastPointer()
}

// Interface for all types of wrapper around a channel of results and an error channel
// Implementors will expose a type-specific channel for results
// Call the Wait() function after processing the results channel to catch any errors
// that occurred during the async processing
type ChannelWrapper interface {
	// Call this after processing results channel to check for async errors
	Wait() error
}

// Base implementation of channel wrapper to just deal with errors
type BaseChannelWrapper struct {
	errorChan <-chan error
}

func (w *BaseChannelWrapper) Wait() error {
	var err error
	for e := range w.errorChan {
		if err != nil {
			// Combine in case multiple errors
			err = fmt.Errorf("%v\n%v", err, e)

		} else {
			err = e
		}
	}

	return err
}

// ChannelWrapper for pointer Scan* functions to more easily return async error data via Wait()
// See NewPointerChannelWrapper for construction / use
type PointerChannelWrapper struct {
	*BaseChannelWrapper
	Results <-chan *WrappedPointer
}

// Construct a new channel wrapper for WrappedPointer
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
// Scan function is required to create error channel large enough not to block (usually 1 is ok)
func NewPointerChannelWrapper(pointerChan <-chan *WrappedPointer, errorChan <-chan error) *PointerChannelWrapper {
	return &PointerChannelWrapper{&BaseChannelWrapper{errorChan}, pointerChan}
}

// ChannelWrapper for string channel functions to more easily return async error data via Wait()
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
// See NewStringChannelWrapper for construction / use
type StringChannelWrapper struct {
	*BaseChannelWrapper
	Results <-chan string
}

// Construct a new channel wrapper for string
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
func NewStringChannelWrapper(stringChan <-chan string, errorChan <-chan error) *StringChannelWrapper {
	return &StringChannelWrapper{&BaseChannelWrapper{errorChan}, stringChan}
}

// ChannelWrapper for TreeBlob channel functions to more easily return async error data via Wait()
// See NewTreeBlobChannelWrapper for construction / use
type TreeBlobChannelWrapper struct {
	*BaseChannelWrapper
	Results <-chan TreeBlob
}

// Construct a new channel wrapper for TreeBlob
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
func NewTreeBlobChannelWrapper(treeBlobChan <-chan TreeBlob, errorChan <-chan error) *TreeBlobChannelWrapper {
	return &TreeBlobChannelWrapper{&BaseChannelWrapper{errorChan}, treeBlobChan}
}
