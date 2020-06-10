package git_test // to avoid import cycles

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	. "github.com/git-lfs/git-lfs/git"
	test "github.com/git-lfs/git-lfs/t/cmd/util"
	"github.com/stretchr/testify/assert"
)

func TestRefString(t *testing.T) {
	const sha = "0000000000000000000000000000000000000000"
	for s, r := range map[string]*Ref{
		"refs/heads/master": {
			Name: "master",
			Type: RefTypeLocalBranch,
			Sha:  sha,
		},
		"refs/remotes/origin/master": {
			Name: "origin/master",
			Type: RefTypeRemoteBranch,
			Sha:  sha,
		},
		"refs/tags/v1.0.0": {
			Name: "v1.0.0",
			Type: RefTypeLocalTag,
			Sha:  sha,
		},
		"HEAD": {
			Name: "HEAD",
			Type: RefTypeHEAD,
			Sha:  sha,
		},
		"other": {
			Name: "other",
			Type: RefTypeOther,
			Sha:  sha,
		},
	} {
		assert.Equal(t, s, r.Refspec())
	}
}

func TestParseRefs(t *testing.T) {
	tests := map[string]RefType{
		"refs/heads":   RefTypeLocalBranch,
		"refs/tags":    RefTypeLocalTag,
		"refs/remotes": RefTypeRemoteBranch,
	}

	for prefix, expectedType := range tests {
		r := ParseRef(prefix+"/branch", "abc123")
		assert.Equal(t, "abc123", r.Sha, "prefix: "+prefix)
		assert.Equal(t, "branch", r.Name, "prefix: "+prefix)
		assert.Equal(t, expectedType, r.Type, "prefix: "+prefix)
	}

	r := ParseRef("refs/foo/branch", "abc123")
	assert.Equal(t, "abc123", r.Sha, "prefix: refs/foo")
	assert.Equal(t, "refs/foo/branch", r.Name, "prefix: refs/foo")
	assert.Equal(t, RefTypeOther, r.Type, "prefix: refs/foo")

	r = ParseRef("HEAD", "abc123")
	assert.Equal(t, "abc123", r.Sha, "prefix: HEAD")
	assert.Equal(t, "HEAD", r.Name, "prefix: HEAD")
	assert.Equal(t, RefTypeHEAD, r.Type, "prefix: HEAD")
}

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
	gitConf := repo.GitConfig()
	ref, err := CurrentRef()
	assert.Nil(t, err)
	assert.Equal(t, &Ref{
		Name: "branch3",
		Type: RefTypeLocalBranch,
		Sha:  outputs[3].Sha,
	}, ref)
	test.RunGitCommand(t, true, "checkout", "master")
	ref, err = CurrentRef()
	assert.Nil(t, err)
	assert.Equal(t, &Ref{
		Name: "master",
		Type: RefTypeLocalBranch,
		Sha:  outputs[2].Sha,
	}, ref)
	// Check remote
	repo.AddRemote("origin")
	test.RunGitCommand(t, true, "push", "-u", "origin", "master:someremotebranch")
	ref, err = gitConf.CurrentRemoteRef()
	assert.Nil(t, err)
	assert.Equal(t, &Ref{
		Name: "origin/someremotebranch",
		Type: RefTypeRemoteBranch,
		Sha:  outputs[2].Sha,
	}, ref)

	refname, err := gitConf.RemoteRefNameForCurrentBranch()
	assert.Nil(t, err)
	assert.Equal(t, "refs/remotes/origin/someremotebranch", refname)

	ref, err = ResolveRef(outputs[2].Sha)
	assert.Nil(t, err)
	assert.Equal(t, &Ref{
		Name: outputs[2].Sha,
		Type: RefTypeOther,
		Sha:  outputs[2].Sha,
	}, ref)
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
		{
			Name: "master",
			Type: RefTypeLocalBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "included_branch_2",
			Type: RefTypeLocalBranch,
			Sha:  outputs[4].Sha,
		},
		{
			Name: "included_branch",
			Type: RefTypeLocalBranch,
			Sha:  outputs[3].Sha,
		},
	}
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")

	// Recent, remotes too (all of them)
	refs, err = RecentBranches(now.AddDate(0, 0, -7), true, "")
	assert.Equal(t, nil, err)
	expectedRefs = []*Ref{
		{
			Name: "master",
			Type: RefTypeLocalBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "included_branch_2",
			Type: RefTypeLocalBranch,
			Sha:  outputs[4].Sha,
		},
		{
			Name: "included_branch",
			Type: RefTypeLocalBranch,
			Sha:  outputs[3].Sha,
		},
		{
			Name: "upstream/master",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "upstream/included_branch_2",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[4].Sha,
		},
		{
			Name: "origin/master",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "origin/included_branch",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[3].Sha,
		},
	}
	// Need to sort for consistent comparison
	sort.Sort(test.RefsByName(expectedRefs))
	sort.Sort(test.RefsByName(refs))
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")

	// Recent, only single remote
	refs, err = RecentBranches(now.AddDate(0, 0, -7), true, "origin")
	assert.Equal(t, nil, err)
	expectedRefs = []*Ref{
		{
			Name: "master",
			Type: RefTypeLocalBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "origin/master",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[5].Sha,
		},
		{
			Name: "included_branch_2",
			Type: RefTypeLocalBranch,
			Sha:  outputs[4].Sha,
		},
		{
			Name: "included_branch",
			Type: RefTypeLocalBranch,
			Sha:  outputs[3].Sha,
		},
		{
			Name: "origin/included_branch",
			Type: RefTypeRemoteBranch,
			Sha:  outputs[3].Sha,
		},
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
	if !IsGitVersionAtLeast("2.5.0") {
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
		{
			Name: "master",
			Type: RefTypeLocalBranch,
			Sha:  outputs[0].Sha,
		},
		{
			Name: "branch2",
			Type: RefTypeLocalBranch,
			Sha:  outputs[1].Sha,
		},
		{
			Name: "branch4",
			Type: RefTypeLocalBranch,
			Sha:  outputs[3].Sha,
		},
	}
	// Need to sort for consistent comparison
	sort.Sort(test.RefsByName(expectedRefs))
	sort.Sort(test.RefsByName(refs))
	assert.Equal(t, expectedRefs, refs, "Refs should be correct")
}

func TestVersionCompare(t *testing.T) {
	assert.True(t, IsVersionAtLeast("2.6.0", "2.6.0"))
	assert.True(t, IsVersionAtLeast("2.6.0", "2.6"))
	assert.True(t, IsVersionAtLeast("2.6.0", "2"))
	assert.True(t, IsVersionAtLeast("2.6.10", "2.6.5"))
	assert.True(t, IsVersionAtLeast("2.8.1", "2.7.2"))

	assert.False(t, IsVersionAtLeast("1.6.0", "2"))
	assert.False(t, IsVersionAtLeast("2.5.0", "2.6"))
	assert.False(t, IsVersionAtLeast("2.5.0", "2.5.1"))
	assert.False(t, IsVersionAtLeast("2.5.2", "2.5.10"))
}

func TestGitAndRootDirs(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	git, root, err := GitAndRootDirs()
	if err != nil {
		t.Fatal(err)
	}

	expected, err := os.Stat(git)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := os.Stat(filepath.Join(root, ".git"))
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, os.SameFile(expected, actual))
}

func TestGetTrackedFiles(t *testing.T) {
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
				{Filename: "file2.txt", Size: 20},
				{Filename: "folder1/file10.txt", Size: 20},
				{Filename: "folder1/anotherfile.txt", Size: 20},
			},
		},
		{ // 1
			Files: []*test.FileInput{
				{Filename: "file3.txt", Size: 20},
				{Filename: "file4.txt", Size: 20},
				{Filename: "folder2/something.txt", Size: 20},
				{Filename: "folder2/folder3/deep.txt", Size: 20},
			},
		},
	}
	repo.AddCommits(inputs)

	tracked, err := GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked) // for direct comparison
	fulllist := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "folder1/anotherfile.txt", "folder1/file10.txt", "folder2/folder3/deep.txt", "folder2/something.txt"}
	assert.Equal(t, fulllist, tracked)

	tracked, err = GetTrackedFiles("*file*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	sublist := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "folder1/anotherfile.txt", "folder1/file10.txt"}
	assert.Equal(t, sublist, tracked)

	tracked, err = GetTrackedFiles("folder1/*")
	assert.Nil(t, err)
	sort.Strings(tracked)
	sublist = []string{"folder1/anotherfile.txt", "folder1/file10.txt"}
	assert.Equal(t, sublist, tracked)

	tracked, err = GetTrackedFiles("folder2/*")
	assert.Nil(t, err)
	sort.Strings(tracked)
	sublist = []string{"folder2/folder3/deep.txt", "folder2/something.txt"}
	assert.Equal(t, sublist, tracked)

	// relative dir
	os.Chdir("folder1")
	tracked, err = GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	sublist = []string{"anotherfile.txt", "file10.txt"}
	assert.Equal(t, sublist, tracked)
	os.Chdir("..")

	// absolute paths only includes matches in repo root
	tracked, err = GetTrackedFiles("/*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	assert.Equal(t, []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"}, tracked)

	// Test includes staged but uncommitted files
	ioutil.WriteFile("z_newfile.txt", []byte("Hello world"), 0644)
	test.RunGitCommand(t, true, "add", "z_newfile.txt")
	tracked, err = GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	fulllist = append(fulllist, "z_newfile.txt")
	assert.Equal(t, fulllist, tracked)

	// Test includes modified files (not staged)
	ioutil.WriteFile("file1.txt", []byte("Modifications"), 0644)
	tracked, err = GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	assert.Equal(t, fulllist, tracked)

	// Test includes modified files (staged)
	test.RunGitCommand(t, true, "add", "file1.txt")
	tracked, err = GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	assert.Equal(t, fulllist, tracked)

	// Test excludes deleted files (not committed)
	test.RunGitCommand(t, true, "rm", "file2.txt")
	tracked, err = GetTrackedFiles("*.txt")
	assert.Nil(t, err)
	sort.Strings(tracked)
	deletedlist := []string{"file1.txt", "file3.txt", "file4.txt", "folder1/anotherfile.txt", "folder1/file10.txt", "folder2/folder3/deep.txt", "folder2/something.txt", "z_newfile.txt"}
	assert.Equal(t, deletedlist, tracked)

}

func TestLocalRefs(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	repo.AddCommits([]*test.CommitInput{
		{
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
			},
		},
		{
			NewBranch:      "branch",
			ParentBranches: []string{"master"},
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
			},
		},
	})

	test.RunGitCommand(t, true, "tag", "v1")

	refs, err := LocalRefs()
	if err != nil {
		t.Fatal(err)
	}

	actual := make(map[string]bool)
	for _, r := range refs {
		t.Logf("REF: %s", r.Name)
		switch r.Type {
		case RefTypeHEAD:
			t.Errorf("Local HEAD ref: %v", r)
		case RefTypeOther:
			t.Errorf("Stash or unknown ref: %v", r)
		default:
			actual[r.Name] = true
		}
	}

	expected := []string{"master", "branch", "v1"}
	found := 0
	for _, refname := range expected {
		if actual[refname] {
			found += 1
		} else {
			t.Errorf("could not find ref %q", refname)
		}
	}

	if found != len(expected) {
		t.Errorf("Unexpected local refs: %v", actual)
	}
}

func TestGetFilesChanges(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()

	commits := repo.AddCommits([]*test.CommitInput{
		{
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 20},
			},
		},
		{
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 25},
				{Filename: "file2.txt", Size: 20},
				{Filename: "folder/file3.txt", Size: 10},
			},
			Tags: []string{"tag1"},
		},
		{
			NewBranch:      "abranch",
			ParentBranches: []string{"master"},
			Files: []*test.FileInput{
				{Filename: "file1.txt", Size: 30},
				{Filename: "file4.txt", Size: 40},
			},
		},
	})

	expected0to1 := []string{"file1.txt", "file2.txt", "folder/file3.txt"}
	expected1to2 := []string{"file1.txt", "file4.txt"}
	expected0to2 := []string{"file1.txt", "file2.txt", "file4.txt", "folder/file3.txt"}
	// Test 2 SHAs
	changes, err := GetFilesChanged(commits[0].Sha, commits[1].Sha)
	assert.Nil(t, err)
	assert.Equal(t, expected0to1, changes)
	// Test SHA & tag
	changes, err = GetFilesChanged(commits[0].Sha, "tag1")
	assert.Nil(t, err)
	assert.Equal(t, expected0to1, changes)
	// Test SHA & branch
	changes, err = GetFilesChanged(commits[0].Sha, "abranch")
	assert.Nil(t, err)
	assert.Equal(t, expected0to2, changes)
	// Test tag & branch
	changes, err = GetFilesChanged("tag1", "abranch")
	assert.Nil(t, err)
	assert.Equal(t, expected1to2, changes)
	// Test fail
	_, err = GetFilesChanged("tag1", "nonexisting")
	assert.NotNil(t, err)
	_, err = GetFilesChanged("nonexisting", "tag1")
	assert.NotNil(t, err)
	// Test Single arg version
	changes, err = GetFilesChanged(commits[1].Sha, "")
	assert.Nil(t, err)
	assert.Equal(t, expected0to1, changes)
	changes, err = GetFilesChanged("abranch", "")
	assert.Nil(t, err)
	assert.Equal(t, expected1to2, changes)

}

func TestValidateRemoteURL(t *testing.T) {
	assert.Nil(t, ValidateRemoteURL("https://github.com/git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("http://github.com/git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("git://github.com/git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("ssh://git@github.com/git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("ssh://git@github.com:22/git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("git@github.com:git-lfs/git-lfs"))
	assert.Nil(t, ValidateRemoteURL("git@server:/absolute/path.git"))
	assert.NotNil(t, ValidateRemoteURL("ftp://git@github.com/git-lfs/git-lfs"))
}

func TestRefTypeKnownPrefixes(t *testing.T) {
	for typ, expected := range map[RefType]struct {
		Prefix string
		Ok     bool
	}{
		RefTypeLocalBranch:  {"refs/heads", true},
		RefTypeRemoteBranch: {"refs/remotes", true},
		RefTypeLocalTag:     {"refs/tags", true},
		RefTypeHEAD:         {"", false},
		RefTypeOther:        {"", false},
	} {
		prefix, ok := typ.Prefix()

		assert.Equal(t, expected.Prefix, prefix)
		assert.Equal(t, expected.Ok, ok)
	}
}

func TestRemoteURLs(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()
	cfg := repo.GitConfig()
	cfg.SetLocal("remote.foo.url", "https://github.com/git-lfs/git-lfs.git")
	cfg.SetLocal("remote.bar.url", "https://github.com/git-lfs/wildmatch.git")
	cfg.SetLocal("remote.bar.pushurl", "https://github.com/git-lfs/pktline.git")

	expected := make(map[string][]string)
	expected["foo"] = []string{"https://github.com/git-lfs/git-lfs.git"}
	expected["bar"] = []string{"https://github.com/git-lfs/wildmatch.git"}
	actual, err := RemoteURLs(false)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

	expected["bar"] = []string{"https://github.com/git-lfs/pktline.git"}
	actual, err = RemoteURLs(true)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestMapRemoteURL(t *testing.T) {
	repo := test.NewRepo(t)
	repo.Pushd()
	defer func() {
		repo.Popd()
		repo.Cleanup()
	}()
	cfg := repo.GitConfig()
	cfg.SetLocal("remote.foo.url", "https://github.com/git-lfs/git-lfs.git")
	cfg.SetLocal("remote.bar.url", "https://github.com/git-lfs/wildmatch.git")
	cfg.SetLocal("remote.bar.pushurl", "https://github.com/git-lfs/pktline.git")

	tests := []struct {
		url   string
		push  bool
		match bool
		val   string
	}{
		{
			"https://github.com/git-lfs/git-lfs.git",
			false,
			true,
			"foo",
		},
		{
			"https://github.com/git-lfs/git-lfs.git",
			true,
			true,
			"foo",
		},
		{
			"https://github.com/git-lfs/wildmatch.git",
			false,
			true,
			"bar",
		},
		{
			"https://github.com/git-lfs/pktline.git",
			true,
			true,
			"bar",
		},
		{
			"https://github.com/git-lfs/pktline.git",
			false,
			false,
			"https://github.com/git-lfs/pktline.git",
		},
		{
			"https://github.com/git/git.git",
			true,
			false,
			"https://github.com/git/git.git",
		},
	}
	for _, test := range tests {
		val, ok := MapRemoteURL(test.url, test.push)
		assert.Equal(t, ok, test.match)
		assert.Equal(t, val, test.val)
	}
}

func TestIsValidObjectIDLength(t *testing.T) {
	// Lengths are 40, 64, 39, and 12.
	assert.Equal(t, HasValidObjectIDLength("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), true)
	assert.Equal(t, HasValidObjectIDLength("2222222222222222222222222222222222222222222222222222222222222222"), true)
	assert.Equal(t, HasValidObjectIDLength("555555555555555555555555555555555555555"), false)
	assert.Equal(t, HasValidObjectIDLength("0123456789ab"), false)
}

func TestIsZeroObjectID(t *testing.T) {
	assert.Equal(t, IsZeroObjectID("0000000000000000000000000000000000000000"), true)
	assert.Equal(t, IsZeroObjectID("0000000000000000000000000000000000000000000000000000000000000000"), true)
	assert.Equal(t, IsZeroObjectID("000000000000000000000000000000000000000"), false)
	assert.Equal(t, IsZeroObjectID("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"), false)
	assert.Equal(t, IsZeroObjectID("473a0f4c3be8a93681a267e3b1e9a7dcda1185436fe141f7749120a303721813"), false)
}
