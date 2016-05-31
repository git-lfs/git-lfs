package lfs_test // to avoid import cycles

// This is for doing complete git-level tests using test utils
// Needs to be a separate file from scanner_test so that we can use a diff package
// which avoids import cycles with testutils

import (
	"sort"
	"testing"
	"time"

	. "github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/test"
	"github.com/stretchr/testify/assert"
)

func TestScanUnpushed(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	inputs := []*test.CommitInput{
		{ // 0
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
			},
		},
		{ // 1
			NewBranch: "branch2",
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 25},
			},
		},
		{ // 2
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 30},
			},
		},
		{ // 3
			NewBranch: "branch3",
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 32},
			},
		},
	}
	repo.AddCommits(inputs)

	// Add a couple of remotes and test state depending on what's pushed
	repo.AddRemote("origin")
	repo.AddRemote("upstream")

	pointers, err := ScanUnpushed("")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Len(t, pointers, 4, "Should be 4 pointers because none pushed")

	test.RunGitCommand(t, true, "push", "origin", "branch2")
	// Branch2 will have pushed 2 commits
	pointers, err = ScanUnpushed("")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Len(t, pointers, 2, "Should be 2 pointers")

	test.RunGitCommand(t, true, "push", "upstream", "master")
	// Master pushes 1 more commit
	pointers, err = ScanUnpushed("")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Len(t, pointers, 1, "Should be 1 pointer")

	test.RunGitCommand(t, true, "push", "origin", "branch3")
	// All pushed (somewhere)
	pointers, err = ScanUnpushed("")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Empty(t, pointers, "Should be 0 pointers unpushed")

	// Check origin
	pointers, err = ScanUnpushed("origin")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Empty(t, pointers, "Should be 0 pointers unpushed to origin")

	// Check upstream
	pointers, err = ScanUnpushed("upstream")
	assert.Nil(t, err, "Should be no error calling ScanUnpushed")
	assert.Len(t, pointers, 2, "Should be 2 pointers unpushed to upstream")
}

func TestScanPreviousVersions(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	now := time.Now()

	inputs := []*test.CommitInput{
		{ // 0
			CommitDate: now.AddDate(0, 0, -20),
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
				{Filename: "file2.txt", Size: 30},
				{Filename: "folder/nested.txt", Size: 40},
				{Filename: "folder/nested2.txt", Size: 31},
			},
		},
		{ // 1
			CommitDate: now.AddDate(0, 0, -10),
			Files: []*test.FileInput{
				{Filename: "file2.txt", Size: 22},
			},
		},
		{ // 2
			NewBranch:  "excluded",
			CommitDate: now.AddDate(0, 0, -6),
			Files: []*test.FileInput{
				{Filename: "file2.txt", Size: 12},
				{Filename: "folder/nested2.txt", Size: 16},
			},
		},
		{ // 3
			ParentBranches: []string{"master"},
			CommitDate:     now.AddDate(0, 0, -4),
			Files: []*test.FileInput{
				{Filename: "folder/nested.txt", Size: 42},
				{Filename: "folder/nested2.txt", Size: 6},
			},
		},
		{ // 4
			Files: []*test.FileInput{
				{Filename: "folder/nested.txt", Size: 22},
			},
		},
	}
	outputs := repo.AddCommits(inputs)

	// Previous commits excludes final state of each file, which is:
	// file1.txt            [0] (unchanged since first commit so excluded)
	// file2.txt            [1] (because [2] is on another branch so excluded)
	// folder/nested.txt    [4] (updated at last commit)
	// folder/nested2.txt   [3]

	// The only changes which will be included are changes prior to final state
	// where the '-' side of the diff is inside the date range

	// 7 day limit excludes [0] commit, but includes state from that if there
	// was a subsequent change
	pointers, err := ScanPreviousVersions("master", now.AddDate(0, 0, -7))
	assert.Equal(t, nil, err)

	// Includes the following 'before' state at commits:
	// folder/nested.txt [-diff at 4, ie 3, -diff at 3 ie 0]
	// folder/nested2.txt [-diff at 3 ie 0]
	// others are either on diff branches, before this window, or unchanged
	expected := []*WrappedPointer{
		{Name: "folder/nested.txt", Size: outputs[3].Files[0].Size, Pointer: outputs[3].Files[0]},
		{Name: "folder/nested.txt", Size: outputs[0].Files[2].Size, Pointer: outputs[0].Files[2]},
		{Name: "folder/nested2.txt", Size: outputs[0].Files[3].Size, Pointer: outputs[0].Files[3]},
	}
	// Need to sort to compare equality
	sort.Sort(test.WrappedPointersByOid(expected))
	sort.Sort(test.WrappedPointersByOid(pointers))
	assert.Equal(t, expected, pointers)

}
