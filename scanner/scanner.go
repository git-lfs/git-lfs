package scanner

import (
	"bufio"
	"bytes"
	"github.com/github/git-lfs/pointer"
	"github.com/rubyist/tracerx"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// blobSizeCutoff is used to determine which files to scan for Git LFS
	// pointers.  Any file with a size below this cutoff will be scanned.
	blobSizeCutoff = 140

	// stdoutBufSize is the size of the buffers given to a sub-process stdout
	stdoutBufSize = 16384

	// chanBufSize is the size of the channels used to pass data from one
	// sub-process to another.
	chanBufSize = 100
)

// wrappedPointer wraps a pointer.Pointer and provides the git sha1
// and the file name associated with the object, taken from the
// rev-list output.
type wrappedPointer struct {
	Sha1    string
	Name    string
	SrcName string
	Size    int64
	Status  string
	*pointer.Pointer
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

// Scan takes a ref and returns a slice of wrappedPointer objects
// for all Git LFS pointers it finds for that ref.
func Scan(refLeft, refRight string) ([]*wrappedPointer, error) {
	nameMap := make(map[string]string, 0)
	start := time.Now()

	revs, err := revListShas(refLeft, refRight, refLeft == "", nameMap)
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

	pointers := make([]*wrappedPointer, 0)
	for p := range pointerc {
		if name, ok := nameMap[p.Sha1]; ok {
			p.Name = name
		}
		pointers = append(pointers, p)
	}

	tracerx.PerformanceSince("scan", start)

	return pointers, nil
}

// ScanIndex returns a slice of wrappedPointer objects for all
// Git LFS pointers it finds in the index.
func ScanIndex() ([]*wrappedPointer, error) {
	nameMap := make(map[string]*indexFile, 0)
	start := time.Now()

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

	pointers := make([]*wrappedPointer, 0)
	for p := range pointerc {
		if e, ok := nameMap[p.Sha1]; ok {
			p.Name = e.Name
			p.Status = e.Status
			p.SrcName = e.SrcName
		}
		pointers = append(pointers, p)
	}

	tracerx.PerformanceSince("scan-staging", start)

	return pointers, nil

}

// revListShas uses git rev-list to return the list of object sha1s
// for the given ref. If all is true, ref is ignored. It returns a
// channel from which sha1 strings can be read.
func revListShas(refLeft, refRight string, all bool, nameMap map[string]string) (chan string, error) {
	refArgs := []string{"rev-list", "--objects"}
	if all {
		refArgs = append(refArgs, "--all")
	} else {
		refArgs = append(refArgs, "--no-walk")
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
				nameMap[sha1] = line[41:len(line)]
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
func catFileBatch(revs chan string) (chan *wrappedPointer, error) {
	cmd, err := startCommand("git", "cat-file", "--batch")
	if err != nil {
		return nil, err
	}

	pointers := make(chan *wrappedPointer, chanBufSize)

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

			p, err := pointer.Decode(bytes.NewBuffer(nbuf))
			if err == nil {
				pointers <- &wrappedPointer{
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
