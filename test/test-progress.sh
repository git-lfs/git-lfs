#!/usr/bin/env bash

. "test/testlib.sh"

reponame="$(basename "$0" ".sh")"

begin_test "GIT_LFS_PROGRESS"
(
  set -e
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "a" > a.dat
  echo "b" > b.dat
  echo "c" > c.dat
  echo "d" > d.dat
  echo "e" > e.dat
  git add .gitattributes *.dat
  git commit -m "add files"
  git push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (5/5), 10 B" push.log

  cd ..
  GIT_LFS_PROGRESS="$TRASHDIR/progress.log" git lfs clone "$GITSERVER/$reponame" clone
  cat progress.log
  grep "download 1/5" progress.log
  grep "download 2/5" progress.log
  grep "download 3/5" progress.log
  grep "download 4/5" progress.log
  grep "download 5/5" progress.log

  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" clone2
  cd clone2

  rm -rf "$TRASHDIR/progress.log" .git/lfs/objects
  GIT_LFS_PROGRESS="$TRASHDIR/progress.log" git lfs fetch --all
  cat ../progress.log
  grep "download 1/5" ../progress.log
  grep "download 2/5" ../progress.log
  grep "download 3/5" ../progress.log
  grep "download 4/5" ../progress.log
  grep "download 5/5" ../progress.log

  rm -rf "$TRASHDIR/progress.log"
  GIT_LFS_PROGRESS="$TRASHDIR/progress.log" git lfs checkout
  cat ../progress.log
  grep "checkout 1/5" ../progress.log
  grep "checkout 2/5" ../progress.log
  grep "checkout 3/5" ../progress.log
  grep "checkout 4/5" ../progress.log
  grep "checkout 5/5" ../progress.log
)
end_test
