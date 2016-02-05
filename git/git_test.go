package git_test // to avoid import cycles

import (
	"path/filepath"
	"sort"
	"testing"
	"time"

	. "github.com/github/git-lfs/git"
	"github.com/github/git-lfs/test"
	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
)

func TestCurrentRefAndCurrentRemoteRef(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	// test commits; we'll just modify the same file each time since we're
	// only interested in branches
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
	outputs := repo.AddCommits(inputs)
	// last commit was on branch3
	ref, err := CurrentRef()
	assert.Equal(t, nil, err)
	assert.Equal(t, &Ref{"branch3", RefTypeLocalBranch, outputs[3].Sha}, ref)
	test.RunGitCommand(t, true, "checkout", "master")
	ref, err = CurrentRef()
	assert.Equal(t, nil, err)
	assert.Equal(t, &Ref{"master", RefTypeLocalBranch, outputs[2].Sha}, ref)
	// Check remote
	repo.AddRemote("origin")
	test.RunGitCommand(t, true, "push", "-u", "origin", "master:someremotebranch")
	ref, err = CurrentRemoteRef()
	assert.Equal(t, nil, err)
	assert.Equal(t, &Ref{"origin/someremotebranch", RefTypeRemoteBranch, outputs[2].Sha}, ref)

	refname, err := RemoteRefNameForCurrentBranch()
	assert.Equal(t, nil, err)
	assert.Equal(t, "origin/someremotebranch", refname)

	remote, err := RemoteForCurrentBranch()
	assert.Equal(t, nil, err)
	assert.Equal(t, "origin", remote)
}

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
		{ // 0
			CommitDate: now.AddDate(0, 0, -20),
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
			},
		},
		{ // 1
			CommitDate: now.AddDate(0, 0, -15),
			NewBranch:  "excluded_branch", // new branch & tag but too old
			Tags:       []string{"excluded_tag"},
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 25},
			},
		},
		{ // 2
			CommitDate:     now.AddDate(0, 0, -12),
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 30},
			},
		},
		{ // 3
			CommitDate: now.AddDate(0, 0, -6),
			NewBranch:  "included_branch", // new branch within 7 day limit
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 32},
			},
		},
		{ // 4
			CommitDate: now.AddDate(0, 0, -3),
			NewBranch:  "included_branch_2", // new branch within 7 day limit
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 36},
			},
		},
		{ // 5
			// Final commit, current date/time
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 21},
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
		&Ref{"included_branch_2", RefTypeLocalBranch, outputs[4].Sha},
		&Ref{"included_branch", RefTypeLocalBranch, outputs[3].Sha},
		&Ref{"upstream/master", RefTypeRemoteBranch, outputs[5].Sha},
		&Ref{"upstream/included_branch_2", RefTypeRemoteBranch, outputs[4].Sha},
		&Ref{"origin/master", RefTypeRemoteBranch, outputs[5].Sha},
		&Ref{"origin/included_branch", RefTypeRemoteBranch, outputs[3].Sha},
	}
	// Need to sort for consistent comparison
	sort.Sort(test.RefsByName(expectedRefs))
	sort.Sort(test.RefsByName(refs))
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
	// Need to sort for consistent comparison
	sort.Sort(test.RefsByName(expectedRefs))
	sort.Sort(test.RefsByName(refs))
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")
}

func TestResolveEmptyCurrentRef(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	_, err := CurrentRef()
	assert.NotEqual(t, nil, err)
}

func TestWorkTrees(t *testing.T) {

	// Only git 2.5+
	if !Config.IsGitVersionAtLeast("2.5.0") {
		return
	}

	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	// test commits; we'll just modify the same file each time since we're
	// only interested in branches & dates
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
			NewBranch:      "branch3",
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 30},
			},
		},
		{ // 3
			NewBranch:      "branch4",
			ParentBranches: []string{"master"}, // back on master
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 40},
			},
		},
	}
	outputs := repo.AddCommits(inputs)
	// Checkout master again otherwise can't create a worktree from branch4 if we're on it here
	test.RunGitCommand(t, true, "checkout", "master")

	// We can create worktrees as subfolders for convenience
	// Each one is checked out to a different branch
	// Note that we *won't* create one for branch3
	test.RunGitCommand(t, true, "worktree", "add", "branch2_wt", "branch2")
	test.RunGitCommand(t, true, "worktree", "add", "branch4_wt", "branch4")

	refs, err := GetAllWorkTreeHEADs(filepath.Join(repo.Path, ".git"))
	assert.Equal(t, nil, err)
	expectedRefs := []*Ref{
		&Ref{"master", RefTypeLocalBranch, outputs[0].Sha},
		&Ref{"branch2", RefTypeLocalBranch, outputs[1].Sha},
		&Ref{"branch4", RefTypeLocalBranch, outputs[3].Sha},
	}
	// Need to sort for consistent comparison
	sort.Sort(test.RefsByName(expectedRefs))
	sort.Sort(test.RefsByName(refs))
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")
}

func TestVersionCompare(t *testing.T) {
	assert.Equal(t, true, IsVersionAtLeast("2.6.0", "2.6.0"))
	assert.Equal(t, true, IsVersionAtLeast("2.6.0", "2.6"))
	assert.Equal(t, true, IsVersionAtLeast("2.6.0", "2"))
	assert.Equal(t, true, IsVersionAtLeast("2.6.10", "2.6.5"))
	assert.Equal(t, true, IsVersionAtLeast("2.8.1", "2.7.2"))

	assert.Equal(t, false, IsVersionAtLeast("1.6.0", "2"))
	assert.Equal(t, false, IsVersionAtLeast("2.5.0", "2.6"))
	assert.Equal(t, false, IsVersionAtLeast("2.5.0", "2.5.1"))
	assert.Equal(t, false, IsVersionAtLeast("2.5.2", "2.5.10"))
}

func TestGitAndRootDirs(t *testing.T) {
	git, root, err := GitAndRootDirs()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, git, filepath.Join(root, ".git"))
}
