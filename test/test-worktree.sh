#!/usr/bin/env bash

. "test/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.5.0"
envInitConfig='git config filter.lfs.smudge = "git-lfs smudge %f"
git config filter.lfs.clean = "git-lfs clean %f"'

begin_test "git worktree"
(
    set -e
    reponame="worktree-main"
    mkdir $reponame
    cd $reponame
    git init

    # can't create a worktree until there's 1 commit at least
    echo "a" > tmp.txt
    git add tmp.txt
    git commit -m "Initial commit"

    expected=$(printf "%s\n%s\n
LocalWorkingDir=$(native_path_escaped "$TRASHDIR/$reponame")
LocalGitDir=$(native_path_escaped "$TRASHDIR/$reponame/.git")
LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/$reponame/.git")
LocalMediaDir=$(native_path_escaped "$TRASHDIR/$reponame/.git/lfs/objects")
TempDir=$(native_path_escaped "$TRASHDIR/$reponame/.git/lfs/tmp")
ConcurrentTransfers=3
BatchTransfer=true
$(escape_path "$(env | grep "^GIT")")
%s
" "$(git lfs version)" "$(git version)" "$envInitConfig")
    actual=$(git lfs env)
    contains_same_elements "$expected" "$actual"

    worktreename="worktree-2"
    git worktree add "$TRASHDIR/$worktreename"
    cd "$TRASHDIR/$worktreename"

    # git dir in worktree is like submodules (except path is worktrees) but this
    # is only for index, temp etc
    # storage of git objects and lfs objects is in the original .git
    expected=$(printf "%s\n%s\n
LocalWorkingDir=$(native_path_escaped "$TRASHDIR/$worktreename")
LocalGitDir=$(native_path_escaped "$TRASHDIR/$reponame/.git/worktrees/$worktreename")
LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/$reponame/.git")
LocalMediaDir=$(native_path_escaped "$TRASHDIR/$reponame/.git/lfs/objects")
TempDir=$(native_path_escaped "$TRASHDIR/$reponame/.git/worktrees/$worktreename/lfs/tmp")
ConcurrentTransfers=3
BatchTransfer=true
$(escape_path "$(env | grep "^GIT")")
%s
" "$(git lfs version)" "$(git version)" "$envInitConfig")
    actual=$(git lfs env)
    contains_same_elements "$expected" "$actual"
)
end_test
