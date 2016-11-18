package lfs

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

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
	skippedRefs      []string
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

func newScanRefsOptions() *ScanRefsOptions {
	return &ScanRefsOptions{
		nameMap: make(map[string]string, 0),
		mutex:   &sync.Mutex{},
	}
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
