#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "dirty index (git-lfs/git-lfs#1726)"
(
  set -e
  reponame="dirty-index-prep"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo_before_all

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  printf "contents" > a.dat
  ln -s a.dat a.symlink.dat

  git add *.dat
  git commit -m "contents"
  git push origin master

  # ~

  git config --global core.symlinks false

  clone_repo "$reponame" repo
  git status --porcelain 1>&1 | tee status.log

  [ 0 -eq "$(grep -c "M a.symlink.dat" status.log)" ]
)
end_test
