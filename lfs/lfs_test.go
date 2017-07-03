package lfs_test // avoid import cycle

import (
	"fmt"
	"sort"
	"testing"

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

	actual := lfs.AllObjects()
	if len(actual) > 0 {
		for _, file := range actual {
			t.Logf("Found: %v", file)
		}
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

	actualObjects := lfs.AllObjects()
	actual := make([]*lfs.Pointer, len(actualObjects))
	for idx, f := range actualObjects {
		actual[idx] = lfs.NewPointer(f.Oid, f.Size, nil)
	}

	// sort to ensure comparison is equal
	sort.Sort(test.PointersByOid(expected))
	sort.Sort(test.PointersByOid(actual))
	assert.Equal(t, expected, actual, "Oids from disk should be the same as in commits")

}
