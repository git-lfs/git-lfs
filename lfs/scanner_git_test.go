package lfs_test // to avoid import cycles

// This is for doing complete git-level tests using test utils
// Needs to be a separate file from scanner_test so that we can use a diff package
// which avoids import cycles with testutils

import (
	"testing"

	. "github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/test"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
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
				{"file1.txt", 20, nil},
			},
		},
		{ // 1
			NewBranch: "branch2",
			Files: []*test.FileInput{
				{"file1.txt", 25, nil},
			},
		},
		{ // 2
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{"file1.txt", 30, nil},
			},
		},
		{ // 3
			NewBranch: "branch3",
			Files: []*test.FileInput{
				{"file1.txt", 32, nil},
			},
		},
	}
	repo.AddCommits(inputs)

	// Add a couple of remotes and test state depending on what's pushed
	repo.AddRemote("origin")
	repo.AddRemote("upstream")

	pointers, err := ScanUnpushed()
	assert.Equal(t, nil, err, "Should be no error calling ScanUnpushed")
	assert.Equal(t, 4, len(pointers), "Should be 4 pointers because none pushed")

	test.RunGitCommand(t, true, "push", "origin", "branch2")
	// Branch2 will have pushed 2 commits
	pointers, err = ScanUnpushed()
	assert.Equal(t, nil, err, "Should be no error calling ScanUnpushed")
	assert.Equal(t, 2, len(pointers), "Should be 2 pointers")

	test.RunGitCommand(t, true, "push", "upstream", "master")
	// Master pushes 1 more commit
	pointers, err = ScanUnpushed()
	assert.Equal(t, nil, err, "Should be no error calling ScanUnpushed")
	assert.Equal(t, 1, len(pointers), "Should be 1 pointer")

	test.RunGitCommand(t, true, "push", "origin", "branch3")
	// All pushed (somewhere)
	pointers, err = ScanUnpushed()
	assert.Equal(t, nil, err, "Should be no error calling ScanUnpushed")
	assert.Equal(t, 0, len(pointers), "Should be 0 pointers unpushed")

}
