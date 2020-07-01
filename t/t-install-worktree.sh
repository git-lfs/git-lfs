#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# These tests rely on behavior found in Git versions higher than 2.20.0 to
# perform themselves, specifically:
#   - worktreeConfig extension support
ensure_git_version_isnt $VERSION_LOWER "2.20.0"

begin_test "install --worktree outside repository"
(
  set -e

  # If run inside the git-lfs source dir this will update its .git/config & cause issues
  if [ "$GIT_LFS_TEST_DIR" == "" ]; then
    echo "Skipping install --worktree because GIT_LFS_TEST_DIR is not set"
    exit 0
  fi

  has_test_dir || exit 0

  set +e
  git lfs install --worktree >out.log
  res=$?
  set -e

  [ "Not in a git repository." = "$(cat out.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "install --worktree with single working tree"
(
  set -e

  # old values that should be ignored by `install --worktree`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"

  reponame="$(basename "$0" ".sh")-single-tree"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install --worktree

  # local configs are correct
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --local filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --local filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --worktree filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --local filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --worktree filter.lfs.process)" ]

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]
)
end_test

begin_test "install --worktree with multiple working trees"
(
  set -e

  reponame="$(basename "$0" ".sh")-multi-tree"
  mkdir "$reponame"
  cd "$reponame"
  git init

  # old values that should be ignored by `install --worktree`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"
  git config --local filter.lfs.smudge "local smudge"
  git config --local filter.lfs.clean "local clean"
  git config --local filter.lfs.process "local filter"

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  git config core.repositoryformatversion 1
  git config extensions.worktreeConfig true

  treename="../$reponame-wt"
  git worktree add "$treename"
  cd "$treename"

  git lfs install --worktree

  # worktree configs are correct
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --worktree filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --worktree filter.lfs.process)" ]

  # local configs are correct
  [ "local smudge" = "$(git config --local filter.lfs.smudge)" ]
  [ "local clean" = "$(git config --local filter.lfs.clean)" ]
  [ "local filter" = "$(git config --local filter.lfs.process)" ]

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]
)
end_test

begin_test "install --worktree without worktreeConfig extension"
(
  set -e

  reponame="$(basename "$0" ".sh")-multi-tree-no-config"
  mkdir "$reponame"
  cd "$reponame"
  git init

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  treename="../$reponame-wt"
  git worktree add "$treename"
  cd "$treename"

  set +e
  git lfs install --worktree >out.log
  res=$?
  set -e

  cat out.log
  grep -E "error running.*git.*config" out.log
  [ "$res" -eq 2 ]
)
end_test

begin_test "install --worktree with conflicting scope"
(
  set -e

  reponame="$(basename "$0" ".sh")-scope-conflict"
  mkdir "$reponame"
  cd "$reponame"
  git init

  set +e
  git lfs install --local --worktree 2>err.log
  res=$?
  set -e

  [ "Only one of --local and --worktree options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]

  set +e
  git lfs install --worktree --system 2>err.log
  res=$?
  set -e

  [ "Only one of --worktree and --system options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]
)
end_test
