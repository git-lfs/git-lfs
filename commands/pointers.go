package commands

import "github.com/git-lfs/git-lfs/lfs"

func collectPointers(pointerCh *lfs.PointerChannelWrapper) ([]*lfs.WrappedPointer, error) {
	var pointers []*lfs.WrappedPointer
	for p := range pointerCh.Results {
		pointers = append(pointers, p)
	}
	return pointers, pointerCh.Wait()
}
