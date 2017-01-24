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
  [ ! -w pcfile1.dat ]
  [ ! -w pcfile2.dat ]
  [ -w pcfile3.big ]
  [ -w pcfile4.big ]

  git push -u origin master

  # now lock files, then edit
  GITLFSLOCKSENABLED=1 git lfs lock pcfile1.dat 
  GITLFSLOCKSENABLED=1 git lfs lock pcfile2.dat

  echo "Take a look" > pcfile1.dat
  echo "and you'll see" > pcfile2.dat

  git add pcfile1.dat pcfile2.dat
  git commit -m "Updated"

  # files should remain writeable since locked
  [ -w pcfile1.dat ]
  [ -w pcfile2.dat ]

)
end_test

