package lfs_test // avoid import cycle

import (
	"fmt"
	"sort"
	"testing"

	"github.com/git-lfs/git-lfs/fs"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/test"
	"github.com/stretchr/testify/assert"
)

func TestAllCurrentObjectsNone(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	empty := true
	repo.Filesystem().EachObject(func(obj fs.Object) error {
		empty = false
		t.Logf("Found: %+v", obj)
		return nil
	})
	if !empty {
		t.Error("Should be no objects")
	}
}

func TestAllCurrentObjectsSome(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	// We're not testing commits here, just storage, so just create a single
	// commit input with lots of files to generate many oids
	numFiles := 20
	files := make([]*test.FileInput, 0, numFiles)
	for i := 0; i < numFiles; i++ {
		// Must be >=16 bytes for each file to be unique
		files = append(files, &test.FileInput{Filename: fmt.Sprintf("file%d.txt", i), Size: 30})
	}

	inputs := []*test.CommitInput{
		{Files: files},
	}

	outputs := repo.AddCommits(inputs)

	expected := make([]*lfs.Pointer, 0, numFiles)
	for _, f := range outputs[0].Files {
		expected = append(expected, f)
	}

	actual := make([]*lfs.Pointer, 0)
	repo.Filesystem().EachObject(func(obj fs.Object) error {
		actual = append(actual, lfs.NewPointer(obj.Oid, obj.Size, nil))
		return nil
	})

	// sort to ensure comparison is equal
	sort.Sort(test.PointersByOid(expected))
	sort.Sort(test.PointersByOid(actual))
	assert.Equal(t, expected, actual, "Oids from disk should be the same as in commits")
}
