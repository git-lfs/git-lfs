package lfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
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
	c, err := ScanRefsToChan(refLeft, refRight, opt)
	if err != nil {
		return nil, err
	}
	pointers := make([]*WrappedPointer, 0)
	for p := range c {
		pointers = append(pointers, p)
	}

	return pointers, nil

}

// ScanRefsToChan takes a ref and returns a channel of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanRefsToChan(refLeft, refRight string, opt *ScanRefsOptions) (<-chan *WrappedPointer, error) {
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

	pointerc, err := catFileBatch(smallShas)
	if err != nil {
		return nil, err
	}

	retchan := make(chan *WrappedPointer, chanBufSize)
	go func() {
		for p := range pointerc {
			if name, ok := opt.GetName(p.Sha1); ok {
				p.Name = name
			}
			retchan <- p
		}
		close(retchan)
	}()

	return retchan, nil
}

type indexFileMap struct {
	nameMap map[string]*indexFile
	mutex   *sync.Mutex
}

func (m *indexFileMap) Get(sha string) (*indexFile, bool) {
	m.mutex.Lock()
	index, ok := m.nameMap[sha]
	m.mutex.Unlock()
	return index, ok
}

func (m *indexFileMap) Set(sha string, index *indexFile) {
	m.mutex.Lock()
	m.nameMap[sha] = index
	m.mutex.Unlock()
}

// ScanIndex returns a slice of WrappedPointer objects for all
// Git LFS pointers it finds in the index.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanIndex() ([]*WrappedPointer, error) {
	indexMap := &indexFileMap{
		nameMap: make(map[string]*indexFile, 0),
		mutex:   &sync.Mutex{},
	}

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan-staging", start)
	}()

	revs, err := revListIndex(false, indexMap)
	if err != nil {
		return nil, err
	}

	cachedRevs, err := revListIndex(true, indexMap)
	if err != nil {
		return nil, err
	}

	allRevs := make(chan string)
	go func() {
		seenRevs := make(map[string]bool, 0)

		for rev := range revs {
			seenRevs[rev] = true
			allRevs <- rev
		}

		for rev := range cachedRevs {
			if _, ok := seenRevs[rev]; !ok {
				allRevs <- rev
			}
		}
		close(allRevs)
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
	for p := range pointerc {
		if e, ok := indexMap.Get(p.Sha1); ok {
			p.Name = e.Name
			p.Status = e.Status
			p.SrcName = e.SrcName
		}
		pointers = append(pointers, p)
	}

	return pointers, nil

}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(refLeft, refRight string, opt *ScanRefsOptions) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
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
		refArgs = append(refArgs, refLeft, "--not", "--remotes="+opt.RemoteName)
	default:
		return nil, errors.New("scanner: unknown scan type: " + strconv.Itoa(int(opt.ScanMode)))
	}

	cmd, err := startCommand("git", refArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)

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

		cmd.Wait()
		close(revs)
	}()

	return revs, nil
}

// revListIndex uses git diff-index to return the list of object sha1s
// for in the indexf. It returns a channel from which sha1 strings can be read.
// The namMap will be filled indexFile pointers mapping sha1s to indexFiles.
func revListIndex(cache bool, indexMap *indexFileMap) (chan string, error) {
	cmdArgs := []string{"diff-index", "-M"}
	if cache {
		cmdArgs = append(cmdArgs, "--cached")
	}
	cmdArgs = append(cmdArgs, "HEAD")

	cmd, err := startCommand("git", cmdArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	revs := make(chan string, chanBufSize)

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
				indexMap.Set(sha1, &indexFile{files[len(files)-1], files[0], status})
				revs <- sha1
			}
		}

		cmd.Wait()
		close(revs)
	}()

	return revs, nil
}

// catFileBatchCheck uses git cat-file --batch-check to get the type
// and size of a git object. Any object that isn't of type blob and
// under the blobSizeCutoff will be ignored. revs is a channel over
// which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func catFileBatchCheck(revs chan string) (chan string, error) {
	cmd, err := startCommand("git", "cat-file", "--batch-check")
	if err != nil {
		return nil, err
	}

	smallRevs := make(chan string, chanBufSize)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Format is:
			// <sha1> <type> <size>
			// type is at a fixed spot, if we see that it's "blob", we can avoid
			// splitting the line just to get the size.
			if line[41:45] == "blob" {
				size, err := strconv.Atoi(line[46:len(line)])
				if err != nil {
					continue
				}
				if size < blobSizeCutoff {
					smallRevs <- line[0:40]
				}
			}
		}

		cmd.Wait()
		close(smallRevs)
	}()

	go func() {
		for r := range revs {
			cmd.Stdin.Write([]byte(r + "\n"))
		}
		cmd.Stdin.Close()
	}()

	return smallRevs, nil
}

// catFileBatch uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. revs is a channel over which strings containing Git SHA1s
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatch(revs chan string) (chan *WrappedPointer, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *WrappedPointer, chanBufSize)

	go func() {
		for {
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
				}
			}

			_, err = cmd.Stdout.ReadBytes('\n') // Extra \n inserted by cat-file
			if err != nil {
				break
			}
		}

		cmd.Wait()
		close(pointers)
	}()

	go func() {
		for r := range revs {
			cmd.Stdin.Write([]byte(r + "\n"))
		}
		cmd.Stdin.Close()
	}()

	return pointers, nil
}

type wrappedCmd struct {
	Stdin  io.WriteCloser
	Stdout *bufio.Reader
	*exec.Cmd
}

// startCommand starts up a command and creates a stdin pipe and a buffered
// stdout pipe, wrapped in a wrappedCmd. The stdout buffer will be of stdoutBufSize
// bytes.
func startCommand(command string, args ...string) (*wrappedCmd, error) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
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

	return &wrappedCmd{stdin, bufio.NewReaderSize(stdout, stdoutBufSize), cmd}, nil
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
	for p := range pointerc {
		pointers = append(pointers, p)
	}

	return pointers, nil
}

// catFileBatchTree uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. treeblobs is a channel over which blob entries
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatchTree(treeblobs chan TreeBlob) (chan *WrappedPointer, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *WrappedPointer, chanBufSize)

	go func() {
		for t := range treeblobs {
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

		cmd.Stdin.Close()
		cmd.Wait()
		close(pointers)
	}()

	return pointers, nil
}

// Use ls-tree at ref to find a list of candidate tree blobs which might be lfs files
// The returned channel will be sent these blobs which should be sent to catFileBatchTree
// for final check & conversion to Pointer
func lsTreeBlobs(ref string) (chan TreeBlob, error) {
	// Snapshot using ls-tree
	lsArgs := []string{"ls-tree",
		"-r",          // recurse
		"-l",          // report object size (we'll need this)
		"--full-tree", // start at the root regardless of where we are in it
		ref}

	cmd, err := startCommand("git", lsArgs...)
	if err != nil {
		return nil, err
	}

	cmd.Stdin.Close()

	blobs := make(chan TreeBlob, chanBufSize)

	go func() {
		scanner := bufio.NewScanner(cmd.Stdout)
		regex := regexp.MustCompile(`^\d+\s+blob\s+([0-9a-zA-Z]{40})\s+(\d+)\s+(.*)$`)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if match := regex.FindStringSubmatch(line); match != nil {
				sz, err := strconv.ParseInt(match[2], 10, 64)
				if err != nil {
					continue
				}
				sha1 := match[1]
				filename := match[3]
				if sz < blobSizeCutoff {
					blobs <- TreeBlob{sha1, filename}
				}

			}
		}
		close(blobs)
	}()

	return blobs, nil
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
	for p := range pointerchan {
		pointers = append(pointers, p)
	}
	return pointers, nil
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
	for p := range pointerchan {
		pointers = append(pointers, p)
	}
	return pointers, nil

}

// ScanPreviousVersionsToChan scans changes reachable from ref (commit) back to since.
// Returns channel of pointers for *previous* versions that overlap that time. Does not
// include pointers which were still in use at ref (use ScanRefsToChan for that)
func ScanPreviousVersionsToChan(ref string, since time.Time) (chan *WrappedPointer, error) {
	return logPreviousSHAs(ref, since)
}

// ScanUnpushedToChan scans history for all LFS pointers which have been added but
// not pushed to the named remote. remoteName can be left blank to mean 'any remote'
// return progressively in a channel
func ScanUnpushedToChan(remoteName string) (chan *WrappedPointer, error) {
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

	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffAdditions, nil, nil, pchan)
		cmd.Wait()
	}()

	return pchan, nil

}

// logPreviousVersions scans history for all previous versions of LFS pointers
// from 'since' up to (but not including) the final state at ref
func logPreviousSHAs(ref string, since time.Time) (chan *WrappedPointer, error) {
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

	// we pull out deletions, since we want the previous SHAs at commits in the range
	// this means we pick up all previous versions that could have been checked
	// out in the date range, not just if the commit which *introduced* them is in the range
	go func() {
		parseLogOutputToPointers(cmd.Stdout, LogDiffDeletions, nil, nil, pchan)
		cmd.Wait()
	}()

	return pchan, nil

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
// results: a channel which will receive the pointers
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
			currentFileIncluded = FilenamePassesIncludeExcludeFilter(currentFilename, includePaths, excludePaths)
		} else if match := fileMergeHeaderRegex.FindStringSubmatch(line); match != nil {
			// Git merge file header is a little different, only one file
			finishLastPointer()
			currentFilename = match[1]
			currentFileIncluded = FilenamePassesIncludeExcludeFilter(currentFilename, includePaths, excludePaths)
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

	close(results)

}
