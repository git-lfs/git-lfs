#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

reponame="$(basename "$0" ".sh")"

begin_test "push zero len file"
(
  set -e

  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat"
  touch empty.dat

  contents="full"
  contents_oid=$(calc_oid "$contents")
  printf "%s" "$contents" > full.dat
  git add .gitattributes *.dat
  git commit -m "add files" | tee commit.log

  # cut from commit output
  #   $ git cat-file -p main
  #   tree 2d67d025fb1f9df9fa349412b4b130e982314e92
  tree="$(git cat-file -p main | cut -f 2 -d " " | head -n 1)"

  # cut from tree output
  #   $ git cat-file -p "$tree"
  #   100644 blob 1e9f8f7cafb6af3a6f6ddf211fa39c45fccea7ab	.gitattributes
  #   100644 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391	empty.dat
  #   100644 blob c5de5ac7dec1c40bafe60d24da9b498937640332	full.dat
  emptyblob="$(git cat-file -p "$tree" | cut -f 3 -d " " | grep "empty.dat" | cut -f 1 -d$'\t')"

  # look for lfs pointer in git blob
  [ "0" = "$(git cat-file -p "$emptyblob" | grep "lfs" -c)" ]

  assert_pointer "main" "full.dat" "$contents_oid" 4

  git push origin main | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 4 B" push.log
)
end_test

begin_test "pull zero len file"
(
  set -e

  clone_repo "$reponame" clone
  rm clone.log

  git status | grep -E "working (directory|tree) clean"
  ls -al

  if [ -s "empty.dat" ]; then
    echo "empty.dat has content:"
    cat empty.dat
    exit 1
  fi

  [ "full" = "$(cat full.dat)" ]
)
end_test
