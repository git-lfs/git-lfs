package git_test // to avoid import cycles

import (
	"testing"
	"time"

	. "github.com/github/git-lfs/git"
	"github.com/github/git-lfs/test"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestRecentBranches(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	now := time.Now()
	// test commits; we'll just modify the same file each time since we're
	// only interested in branches & dates
	inputs := []*test.CommitInput{
		&test.CommitInput{ // 0
			CommitDate: now.AddDate(0, 0, -20),
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 20, nil},
			},
		},
		&test.CommitInput{ // 1
			CommitDate: now.AddDate(0, 0, -15),
			NewBranch:  "excluded_branch", // new branch & tag but too old
			Tags:       []string{"excluded_tag"},
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 25, nil},
			},
		},
		&test.CommitInput{ // 2
			CommitDate:     now.AddDate(0, 0, -12),
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 30, nil},
			},
		},
		&test.CommitInput{ // 3
			CommitDate: now.AddDate(0, 0, -6),
			NewBranch:  "included_branch", // new branch within 7 day limit
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 32, nil},
			},
		},
		&test.CommitInput{ // 4
			CommitDate: now.AddDate(0, 0, -3),
			NewBranch:  "included_branch_2", // new branch within 7 day limit
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 36, nil},
			},
		},
		&test.CommitInput{ // 5
			// Final commit, current date/time
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				&test.FileInput{"file1.txt", 21, nil},
			},
		},
	}
	outputs := repo.AddCommits(inputs)

	// Add a couple of remotes and push some branches
	repo.AddRemote("origin")
	repo.AddRemote("upstream")

	test.RunGitCommand(t, true, "push", "origin", "master")
	test.RunGitCommand(t, true, "push", "origin", "excluded_branch")
	test.RunGitCommand(t, true, "push", "origin", "included_branch")
	test.RunGitCommand(t, true, "push", "upstream", "master")
	test.RunGitCommand(t, true, "push", "upstream", "included_branch_2")

	// Recent, local only
	refs, err := RecentBranches(now.AddDate(0, 0, -7), false, "")
	assert.Equal(t, nil, err)
	expectedRefs := []*Ref{
		&Ref{"master", RefTypeLocalBranch, outputs[5].Sha},
		&Ref{"included_branch_2", RefTypeLocalBranch, outputs[4].Sha},
		&Ref{"included_branch", RefTypeLocalBranch, outputs[3].Sha},
	}
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")

	// Recent, remotes too (all of them)
	refs, err = RecentBranches(now.AddDate(0, 0, -7), true, "")
	assert.Equal(t, nil, err)
	expectedRefs = []*Ref{
		&Ref{"master", RefTypeLocalBranch, outputs[5].Sha},
		&Ref{"upstream/master", RefTypeRemoteBranch, outputs[5].Sha},
		&Ref{"origin/master", RefTypeRemoteBranch, outputs[5].Sha},
		&Ref{"upstream/included_branch_2", RefTypeRemoteBranch, outputs[4].Sha},
		&Ref{"included_branch_2", RefTypeLocalBranch, outputs[4].Sha},
		&Ref{"included_branch", RefTypeLocalBranch, outputs[3].Sha},
		&Ref{"origin/included_branch", RefTypeRemoteBranch, outputs[3].Sha},
	}
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")

	// Recent, only single remote
	refs, err = RecentBranches(now.AddDate(0, 0, -7), true, "origin")
	assert.Equal(t, nil, err)
	expectedRefs = []*Ref{
		&Ref{"master", RefTypeLocalBranch, outputs[5].Sha},
		&Ref{"origin/master", RefTypeRemoteBranch, outputs[5].Sha},
		&Ref{"included_branch_2", RefTypeLocalBranch, outputs[4].Sha},
		&Ref{"included_branch", RefTypeLocalBranch, outputs[3].Sha},
		&Ref{"origin/included_branch", RefTypeRemoteBranch, outputs[3].Sha},
	}
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")
}
