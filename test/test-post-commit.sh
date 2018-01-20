#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "post-commit"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"

  git lfs track --lockable "*.dat"
  git lfs track "*.big" # not lockable
  git add .gitattributes
  git commit -m "add git attributes"

  echo "Come with me" > pcfile1.dat
  echo "and you'll be" > pcfile2.dat
  echo "in a world" > pcfile3.big
  echo "of pure imagination" > pcfile4.big

  git add *.dat
  git commit -m "Committed large files"

  # New lockable files should have been made read-only now since not locked
  refute_file_writeable pcfile1.dat
  refute_file_writeable pcfile2.dat
  assert_file_writeable pcfile3.big
  assert_file_writeable pcfile4.big

  git push -u origin master

  # now lock files, then edit
  git lfs lock pcfile1.dat
  git lfs lock pcfile2.dat

  echo "Take a look" > pcfile1.dat
  echo "and you'll see" > pcfile2.dat

  git add pcfile1.dat pcfile2.dat
  git commit -m "Updated"

  # files should remain writeable since locked
  assert_file_writeable pcfile1.dat
  assert_file_writeable pcfile2.dat 

)
end_test

begin_test "post-commit (locked file outside of LFS)"
(
  set -e

  reponame="post-commit-external"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs install

  # This step is intentionally done in two commits, due to a known bug bug in
  # the post-checkout process LFS performs. It compares changed files from HEAD,
  # which is an invalid previous state for the initial commit of a repository.
  echo "*.dat lockable" > .gitattributes
  git add .gitattributes
  git commit -m "initial commit"

  echo "hello" > a.dat
  git add a.dat
  assert_file_writeable a.dat
  git commit -m "add a.dat"

  refute_file_writeable a.dat
)
end_test
