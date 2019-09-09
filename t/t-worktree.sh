#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.5.0"
envInitConfig='git config filter.lfs.process = "git-lfs filter-process"
git config filter.lfs.smudge = "git-lfs smudge -- %f"
git config filter.lfs.clean = "git-lfs clean -- %f"'

unset_vars () {
    # If set, these will cause the test to fail.
    unset GIT_LFS_NO_TEST_COUNT GIT_LFS_LOCK_ACQUIRE_DISABLED
}

begin_test "git worktree"
(
    set -e
    reponame="worktree-main"
    unset_vars
    mkdir $reponame
    cd $reponame
    git init

    # can't create a worktree until there's 1 commit at least
    echo "a" > tmp.txt
    git add tmp.txt
    git commit -m "Initial commit"

    expected=$(printf "%s\n%s\n
LocalWorkingDir=$(canonical_path_escaped "$TRASHDIR/$reponame")
LocalGitDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git")
LocalGitStorageDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git")
LocalMediaDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs/objects")
LocalReferenceDirs=
TempDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs/tmp")
ConcurrentTransfers=3
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs")
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
$(escape_path "$(env | grep "^GIT")")
%s
" "$(git lfs version)" "$(git version)" "$envInitConfig")
    actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
    contains_same_elements "$expected" "$actual"

    worktreename="worktree-2"
    git worktree add "$TRASHDIR/$worktreename"
    cd "$TRASHDIR/$worktreename"

    # git dir in worktree is like submodules (except path is worktrees) but this
    # is only for index, temp etc
    # storage of git objects and lfs objects is in the original .git
    expected=$(printf "%s\n%s\n
LocalWorkingDir=$(canonical_path_escaped "$TRASHDIR/$worktreename")
LocalGitDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/worktrees/$worktreename")
LocalGitStorageDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git")
LocalMediaDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs/objects")
LocalReferenceDirs=
TempDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs/tmp")
ConcurrentTransfers=3
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=$(canonical_path_escaped "$TRASHDIR/$reponame/.git/lfs")
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
$(escape_path "$(env | grep "^GIT")")
%s
" "$(git lfs version)" "$(git version)" "$envInitConfig")
    actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
    contains_same_elements "$expected" "$actual"
)
end_test

begin_test "git worktree with hooks"
(
    set -e
    reponame="worktree-hooks"
    unset_vars
    mkdir $reponame
    cd $reponame
    git init

    # can't create a worktree until there's 1 commit at least
    echo "a" > tmp.txt
    git add tmp.txt
    git commit -m "Initial commit"

    worktreename="worktree-2-hook"
    git worktree add "$TRASHDIR/$worktreename"
    cd "$TRASHDIR/$worktreename"

    # No hooks so far.
    [ ! -e "$TRASHDIR/$reponame/.git/worktrees/$worktreename/hooks" ]
    [ ! -e "$TRASHDIR/$reponame/.git/hooks/pre-push" ]

    git lfs install

    # Make sure we installed the hooks in the main repo, not the worktree dir.
    [ ! -e "$TRASHDIR/$reponame/.git/worktrees/$worktreename/hooks" ]
    [ -x "$TRASHDIR/$reponame/.git/hooks/pre-push" ]
)
end_test
