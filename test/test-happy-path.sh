#!/bin/sh
# this should run from the git-lfs project root.

. "test/testlib.sh"

begin_test "happy path"
(
  set -e

  reponame="$(basename "$0")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  out=$($GITLFS track "*.dat" 2>&1)
  echo "$out" | grep "Tracking \*.dat"

  contents=$(printf "a")
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  out=$(git commit -m "add a.dat" 2>&1)
  echo "$out" | grep "master (root-commit)"
  echo "$out" | grep "2 files changed"
  echo "$out" | grep "create mode 100644 a.dat"
  echo "$out" | grep "create mode 100644 .gitattributes"

  out=$(cat a.dat)
  if [ "$out" != "a" ]; then
    exit 1
  fi

  assert_pointer "master" "a.dat" "$contents_oid" 1

  refute_server_object "$contents_oid"

  out=$(git push origin master 2>&1)
  echo "$out" | grep "(1 of 1 files) 1 B / 1 B  100.00 %"
  echo "$out" | grep "master -> master"

  assert_server_object "$contents_oid" "$contents"

  out=$(clone_repo "$reponame" clone)
  echo "$out" | grep "Cloning into 'clone'"
  echo "$out" | grep "Downloading a.dat (1 B)"

  out=$(cat a.dat)
  if [ "$out" != "a" ]; then
    exit 1
  fi

  assert_pointer "master" "a.dat" "$contents_oid" 1
)
end_test
