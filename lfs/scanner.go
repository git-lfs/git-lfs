package lfs

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/tools"
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
	Status  string
	*Pointer
}

// catFileBatchCheck uses git cat-file --batch-check to get the type
// and size of a git object. Any object that isn't of type blob and
// under the blobSizeCutoff will be ignored. revs is a channel over
// which strings containing git sha1s will be sent. It returns a channel
// from which sha1 strings can be read.
func catFileBatchCheck(revs *StringChannelWrapper, lockableSet *lockableNameSet) (*StringChannelWrapper, chan string, error) {
	smallRevCh := make(chan string, chanBufSize)
	lockableCh := make(chan string, chanBufSize)
	errCh := make(chan error, 2) // up to 2 errors, one from each goroutine
	if err := runCatFileBatchCheck(smallRevCh, lockableCh, lockableSet, revs, errCh); err != nil {
		return nil, nil, err
	}
	return NewStringChannelWrapper(smallRevCh, errCh), lockableCh, nil
}

// catFileBatch uses git cat-file --batch to get the object contents
// of a git object, given its sha1. The contents will be decoded into
// a Git LFS pointer. revs is a channel over which strings containing Git SHA1s
// will be sent. It returns a channel from which point.Pointers can be read.
func catFileBatch(revs *StringChannelWrapper, lockableSet *lockableNameSet, osEnv config.Environment) (*PointerChannelWrapper, chan string, error) {
	pointerCh := make(chan *WrappedPointer, chanBufSize)
	lockableCh := make(chan string, chanBufSize)
	errCh := make(chan error, 5) // shared by 2 goroutines & may add more detail errors?
	if err := runCatFileBatch(pointerCh, lockableCh, lockableSet, revs, errCh, osEnv); err != nil {
		return nil, nil, err
	}
	return NewPointerChannelWrapper(pointerCh, errCh), lockableCh, nil
}

// ChannelWrapper for pointer Scan* functions to more easily return async error data via Wait()
// See NewPointerChannelWrapper for construction / use
type PointerChannelWrapper struct {
	*tools.BaseChannelWrapper
	Results <-chan *WrappedPointer
}

// Construct a new channel wrapper for WrappedPointer
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
// Scan function is required to create error channel large enough not to block (usually 1 is ok)
func NewPointerChannelWrapper(pointerChan <-chan *WrappedPointer, errorChan <-chan error) *PointerChannelWrapper {
	return &PointerChannelWrapper{tools.NewBaseChannelWrapper(errorChan), pointerChan}
}

// ChannelWrapper for string channel functions to more easily return async error data via Wait()
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
// See NewStringChannelWrapper for construction / use
type StringChannelWrapper struct {
	*tools.BaseChannelWrapper
	Results <-chan string
}

// Construct a new channel wrapper for string
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
func NewStringChannelWrapper(stringChan <-chan string, errorChan <-chan error) *StringChannelWrapper {
	return &StringChannelWrapper{tools.NewBaseChannelWrapper(errorChan), stringChan}
}

// ChannelWrapper for TreeBlob channel functions to more easily return async error data via Wait()
// See NewTreeBlobChannelWrapper for construction / use
type TreeBlobChannelWrapper struct {
	*tools.BaseChannelWrapper
	Results <-chan TreeBlob
}

// Construct a new channel wrapper for TreeBlob
// Caller can use s.Results directly for normal processing then call Wait() to finish & check for errors
func NewTreeBlobChannelWrapper(treeBlobChan <-chan TreeBlob, errorChan <-chan error) *TreeBlobChannelWrapper {
	return &TreeBlobChannelWrapper{tools.NewBaseChannelWrapper(errorChan), treeBlobChan}
}
