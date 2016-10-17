#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "post-checkout"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"

  git lfs track --lockable "*.dat"
  git lfs track "*.big" # not lockable
  git add .gitattributes
  git commit -m "add git attributes"


  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":100},
      {\"Filename\":\"file2.dat\",\"Size\":75}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":110},
      {\"Filename\":\"file3.big\",\"Size\":66},
      {\"Filename\":\"file4.big\",\"Size\":23}],
    \"Tags\":[\"atag\"]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Size\":87}]
  },
  {
    \"CommitDate\":\"$(get_date -3d)\",
    \"NewBranch\":\"branch2\",    
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":120},
      {\"Filename\":\"file6.big\",\"Size\":30}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Size\":33},
      {\"Filename\":\"file3.big\",\"Size\":55}]
  }
  ]" | lfstest-testutils addcommits

  git push -u origin master branch2

  # post-commit will make added files read-only in future, but don't rely on
  # that for this test
  chmod -w *.dat

  git checkout master

  [ -e file1.dat ]
  [ -e file2.dat ]
  [ -e file3.big ]
  [ -e file4.big ]
  [ ! -e file5.dat ]
  [ ! -e file6.big ]
  # without the post-checkout hook, any changed files would now be writeable
  [ ! -w file1.dat ]
  [ ! -w file2.dat ]
  [ -w file3.big ]
  [ -w file4.big ]

  # checkout branch
  git checkout branch2
  [ -e file5.dat ]
  [ -e file6.big ]
  [ ! -w file1.dat ]
  [ ! -w file2.dat ]
  [ ! -w file5.dat ]
  [ -w file3.big ]
  [ -w file4.big ]
  [ -w file6.big ]

  # restore files inside a branch (causes full scan since no diff)
  rm -f *.dat
  git checkout file1.dat file2.dat file5.dat
  [ -e file1.dat ]
  [ -e file2.dat ]
  [ -e file5.dat ]
  [ ! -w file1.dat ]
  [ ! -w file2.dat ]
  [ ! -w file5.dat ]

  # now lock files, then remove & restore
  GITLFSLOCKSENABLED=1 git lfs lock file1.dat 
  GITLFSLOCKSENABLED=1 git lfs lock file2.dat
  [ -w file1.dat ]
  [ -w file2.dat ]
  rm -f *.dat
  git checkout file1.dat file2.dat file5.dat
  [ -w file1.dat ]
  [ -w file2.dat ]
  [ ! -w file5.dat ]



)
end_test

