#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "post-merge"
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

  git push -u origin master branch2

  # re-clone the repo so we start fresh
  cd ..
  rm -rf "$reponame"
  clone_repo "$reponame" "$reponame"

  # this will be master

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

  # merge branch, with readonly option disabled to demonstrate what would happen
  GIT_LFS_SET_LOCKABLE_READONLY=0 git merge origin/branch2
  # branch2 had hanges to file2.dat and file5.dat which were lockable
  # but because we disabled the readonly feature they will be writeable now
  assert_file_writeable file2.dat
  assert_file_writeable file5.dat

  # now let's do it again with the readonly option enabled
  git reset --hard HEAD^
  git merge origin/branch2

  # This time they should be read-only
  refute_file_writeable file2.dat
  refute_file_writeable file5.dat

  # Confirm that contents of existing files were updated even though were read-only
  [ "$(cat file2.dat)" == "file 2 updated in branch2" ]
  [ "$(cat file5.dat)" == "file 5 creation in branch2" ]
)
end_test
