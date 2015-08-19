package git_test // to avoid import cycles

import (
	"testing"
	"time"

	. "github.com/github/git-lfs/git"
	"github.com/github/git-lfs/test"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestRecentBranches(t *testing.T) {
	repo := test.NewTestRepo(t)
	repo.Pushd(t)
	defer func() {
		repo.Popd(t)
		repo.Cleanup(t)
	}()

	now := time.Now()
	// test commits; we'll just modify the same file each time since we're
	// only interested in
	inputs := []*test.TestCommitSetupInput{
		&test.TestCommitSetupInput{ // 0
			CommitDate: now.AddDate(0, 0, -20),
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 20, nil},
			},
		},
		&test.TestCommitSetupInput{ // 1
			CommitDate: now.AddDate(0, 0, -15),
			NewBranch:  "excluded_branch", // new branch & tag but too old
			Tags:       []string{"excluded_tag"},
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 25, nil},
			},
		},
		&test.TestCommitSetupInput{ // 2
			CommitDate:     now.AddDate(0, 0, -12),
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 30, nil},
			},
		},
		&test.TestCommitSetupInput{ // 3
			CommitDate: now.AddDate(0, 0, -6),
			NewBranch:  "included_branch", // new branch within 7 day limit
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 32, nil},
			},
		},
		&test.TestCommitSetupInput{ // 4
			CommitDate: now.AddDate(0, 0, -3),
			NewBranch:  "included_branch_2", // new branch within 7 day limit
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 36, nil},
			},
		},
		&test.TestCommitSetupInput{ // 5
			// Final commit, current date/time
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.TestFileInput{
				&test.TestFileInput{"file1.txt", 21, nil},
			},
		},
	}
	outputs := repo.AddCommits(t, inputs)

	refs, err := RecentBranches(now.AddDate(0, 0, -7), false, "")
	assert.Equal(t, nil, err)
	expectedRefs := []*Ref{
		&Ref{"master", RefTypeLocalBranch, outputs[5].Sha},
		&Ref{"included_branch_2", RefTypeLocalBranch, outputs[4].Sha},
		&Ref{"included_branch", RefTypeLocalBranch, outputs[3].Sha},
	}
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")

}
