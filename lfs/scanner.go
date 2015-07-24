package lfs

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

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

type ScanRefsOptions struct {
	SkipDeletedBlobs bool
	scanAll          bool
	nameMap          map[string]string
}

// ScanRefs takes a ref and returns a slice of WrappedPointer objects
// for all Git LFS pointers it finds for that ref.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanRefs(refLeft, refRight string, opt *ScanRefsOptions) ([]*WrappedPointer, error) {
	if opt == nil {
		opt = &ScanRefsOptions{}
	}
	opt.scanAll = refLeft == ""
	opt.nameMap = make(map[string]string, 0)

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan", start)
	}()

	revs, err := revListShas(refLeft, refRight, *opt)
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

	pointers := make([]*WrappedPointer, 0)
	for p := range pointerc {
		if name, ok := opt.nameMap[p.Sha1]; ok {
			p.Name = name
		}
		pointers = append(pointers, p)
	}

	return pointers, nil
}

// ScanIndex returns a slice of WrappedPointer objects for all
// Git LFS pointers it finds in the index.
// Reports unique oids once only, not multiple times if >1 file uses the same content
func ScanIndex() ([]*WrappedPointer, error) {
	nameMap := make(map[string]*indexFile, 0)

	start := time.Now()
	defer func() {
		tracerx.PerformanceSince("scan-staging", start)
	}()

	revs, err := revListIndex(false, nameMap)
	if err != nil {
		return nil, err
	}

	cachedRevs, err := revListIndex(true, nameMap)
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
		if e, ok := nameMap[p.Sha1]; ok {
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
func revListShas(refLeft, refRight string, opt ScanRefsOptions) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
	if opt.scanAll {
		refArgs = append(refArgs, "--all")
	} else {
		if opt.SkipDeletedBlobs {
			refArgs = append(refArgs, "--no-walk")
		} else {
			refArgs = append(refArgs, "--do-walk")
		}

		refArgs = append(refArgs, refLeft)
		if refRight != "" && !z40.MatchString(refRight) {
			refArgs = append(refArgs, refRight)
		}
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
				opt.nameMap[sha1] = line[41:len(line)]
			}
			revs <- sha1
		}
		close(revs)
	}()

	return revs, nil
}

// revListIndex uses git diff-index to return the list of object sha1s
// for in the indexf. It returns a channel from which sha1 strings can be read.
// The namMap will be filled indexFile pointers mapping sha1s to indexFiles.
func revListIndex(cache bool, nameMap map[string]*indexFile) (chan string, error) {
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
				nameMap[sha1] = &indexFile{files[len(files)-1], files[0], status}
				revs <- sha1
			}
		}
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
		close(pointers)
		cmd.Stdin.Close()
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
