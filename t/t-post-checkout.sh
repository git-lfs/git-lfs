#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

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
      {\"Filename\":\"file1.dat\",\"Data\":\"file 1 creation\"},
      {\"Filename\":\"file2.dat\",\"Data\":\"file 2 creation\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Data\":\"file 1 updated commit 2\"},
      {\"Filename\":\"file3.big\",\"Data\":\"file 3 creation\"},
      {\"Filename\":\"file4.big\",\"Data\":\"file 4 creation\"}],
    \"Tags\":[\"atag\"]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Data\":\"file 2 updated commit 3\"}]
  },
  {
    \"CommitDate\":\"$(get_date -3d)\",
    \"NewBranch\":\"branch2\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Data\":\"file 5 creation in branch2\"},
      {\"Filename\":\"file6.big\",\"Data\":\"file 6 creation in branch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Data\":\"file 2 updated in branch2\"},
      {\"Filename\":\"file3.big\",\"Data\":\"file 3 updated in branch2\"}]
  }
  ]" | GIT_LFS_SET_LOCKABLE_READONLY=0 lfstest-testutils addcommits

  # skipped setting read-only above to make bulk load simpler (no read-only issues)

  git push -u origin main branch2

  # re-clone the repo so we start fresh
  cd ..
  rm -rf "$reponame"
  clone_repo "$reponame" "$reponame"

  # this will be main

  [ "$(cat file1.dat)" == "file 1 updated commit 2" ]
  [ "$(cat file2.dat)" == "file 2 updated commit 3" ]
  [ "$(cat file3.big)" == "file 3 creation" ]
  [ "$(cat file4.big)" == "file 4 creation" ]
  [ ! -e file5.dat ]
  [ ! -e file6.big ]
  # without the post-checkout hook, any changed files would now be writeable
  refute_file_writeable file1.dat
  refute_file_writeable file2.dat
  assert_file_writeable file3.big
  assert_file_writeable file4.big

  # checkout branch
  git checkout branch2
  [ -e file5.dat ]
  [ -e file6.big ]
  refute_file_writeable file1.dat
  refute_file_writeable file2.dat
  refute_file_writeable file5.dat
  assert_file_writeable file3.big
  assert_file_writeable file4.big
  assert_file_writeable file6.big

  # Confirm that contents of existing files were updated even though were read-only
  [ "$(cat file2.dat)" == "file 2 updated in branch2" ]
  [ "$(cat file3.big)" == "file 3 updated in branch2" ]


  # restore files inside a branch (causes full scan since no diff)
  rm -f *.dat
  [ ! -e file1.dat ]
  [ ! -e file2.dat ]
  [ ! -e file5.dat ]
  git checkout file1.dat file2.dat file5.dat
  [ "$(cat file1.dat)" == "file 1 updated commit 2" ]
  [ "$(cat file2.dat)" == "file 2 updated in branch2" ]
  [ "$(cat file5.dat)" == "file 5 creation in branch2" ]
  refute_file_writeable file1.dat
  refute_file_writeable file2.dat
  refute_file_writeable file5.dat

  # now lock files, then remove & restore
  git lfs lock file1.dat
  git lfs lock file2.dat
  assert_file_writeable file1.dat
  assert_file_writeable file2.dat
  rm -f *.dat
  git checkout file1.dat file2.dat file5.dat
  assert_file_writeable file1.dat
  assert_file_writeable file2.dat
  refute_file_writeable file5.dat

)
end_test

begin_test "post-checkout with subdirectories"
(
  set -e

  reponame="post-checkout-subdirectories"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"

  git lfs track --lockable "bin/*.dat"
  git lfs track "*.big" # not lockable
  git add .gitattributes
  git commit -m "add git attributes"

  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"bin/file1.dat\",\"Data\":\"file 1 creation\"},
      {\"Filename\":\"bin/file2.dat\",\"Data\":\"file 2 creation\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"bin/file1.dat\",\"Data\":\"file 1 updated commit 2\"},
      {\"Filename\":\"file3.big\",\"Data\":\"file 3 creation\"},
      {\"Filename\":\"file4.big\",\"Data\":\"file 4 creation\"}],
    \"Tags\":[\"atag\"]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"bin/file2.dat\",\"Data\":\"file 2 updated commit 3\"}]
  },
  {
    \"CommitDate\":\"$(get_date -3d)\",
    \"NewBranch\":\"branch2\",
    \"Files\":[
      {\"Filename\":\"bin/file5.dat\",\"Data\":\"file 5 creation in branch2\"},
      {\"Filename\":\"file6.big\",\"Data\":\"file 6 creation in branch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"bin/file2.dat\",\"Data\":\"file 2 updated in branch2\"},
      {\"Filename\":\"file3.big\",\"Data\":\"file 3 updated in branch2\"}]
  }
  ]" | GIT_LFS_SET_LOCKABLE_READONLY=0 lfstest-testutils addcommits

  # skipped setting read-only above to make bulk load simpler (no read-only issues)

  git push -u origin main branch2

  # re-clone the repo so we start fresh
  cd ..
  rm -rf "$reponame"
  clone_repo "$reponame" "$reponame"

  # this will be main

  [ "$(cat bin/file1.dat)" == "file 1 updated commit 2" ]
  [ "$(cat bin/file2.dat)" == "file 2 updated commit 3" ]
  [ "$(cat file3.big)" == "file 3 creation" ]
  [ "$(cat file4.big)" == "file 4 creation" ]
  [ ! -e bin/file5.dat ]
  [ ! -e file6.big ]
  # without the post-checkout hook, any changed files would now be writeable
  refute_file_writeable bin/file1.dat
  refute_file_writeable bin/file2.dat
  assert_file_writeable file3.big
  assert_file_writeable file4.big

  # checkout branch
  git checkout branch2
  [ -e bin/file5.dat ]
  [ -e file6.big ]
  refute_file_writeable bin/file1.dat
  refute_file_writeable bin/file2.dat
  refute_file_writeable bin/file5.dat
  assert_file_writeable file3.big
  assert_file_writeable file4.big
  assert_file_writeable file6.big

  # Confirm that contents of existing files were updated even though were read-only
  [ "$(cat bin/file2.dat)" == "file 2 updated in branch2" ]
  [ "$(cat file3.big)" == "file 3 updated in branch2" ]
)
end_test
