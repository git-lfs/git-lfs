#!/usr/bin/env bash

. "test/testlib.sh"

assert_same_inode() {
  local repo1=$1
  local repo2=$2
  local oid=$3

  if ! uname -s | grep -qE 'CYGWIN|MSYS|MINGW'; then
    cfg1=$(cd "$repo1"; git lfs env | grep LocalMediaDir)
    f1="${cfg1:14}/${oid:0:2}/${oid:2:2}/$oid"
    inode1=$(ls -i $f1 | cut -f1 -d\ )

    cfg2=$(cd "$repo2"; git lfs env | grep LocalMediaDir)
    f2="${cfg2:14}/${oid:0:2}/${oid:2:2}/$oid"
    inode2=$(ls -i $f2 | cut -f1 -d\ )

    [ "$inode1" == "$inode2" ]
  fi
}

begin_test "clone with reference"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  ref_repo=clone_reference_repo
  ref_repo_dir=$TRASHDIR/$ref_repo
  clone_repo "$reponame" "$ref_repo"
  git lfs track "*.dat"
  contents="a"
  oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1
  git push origin master

  delete_server_object "$reponame" "$oid"

  repo=test_repo
  repo_dir=$TRASHDIR/$repo
  git clone --reference "$ref_repo_dir/.git" \
      "$GITSERVER/$reponame" "$repo_dir"

  cd "$TRASHDIR/$repo"

  assert_pointer "master" "a.dat" "$oid" 1
  assert_same_inode "$repo_dir" "$ref_repo_dir" "$oid"
)
end_test

begin_test "fetch from clone reference"
(
  set -e

  reponame="$(basename "$0" ".sh")2"
  setup_remote_repo "$reponame"

  ref_repo=clone_reference_repo2
  ref_repo_dir=$TRASHDIR/$ref_repo
  clone_repo "$reponame" "$ref_repo"

  repo=test_repo2
  repo_dir=$TRASHDIR/$repo
  git clone --reference "$ref_repo_dir/.git" \
      "$GITSERVER/$reponame" "$repo_dir" 2> clone.log

  cd "$ref_repo_dir"
  git lfs track "*.dat"
  contents="a"
  oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1
  git push origin master

  delete_server_object "$reponame" "$oid"

  cd "$repo_dir"
  GIT_LFS_SKIP_SMUDGE=1 git pull
  git lfs pull

  assert_pointer "master" "a.dat" "$oid" 1
  assert_same_inode "$TRASHDIR/$repo" "$TRASHDIR/$ref_repo" "$oid"
)
end_test
