#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "malformed pointers"
(
  set -e

  reponame="malformed-pointers"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  base64 /dev/urandom | head -c 1023 > malformed_small.dat
  base64 /dev/urandom | head -c 1024 > malformed_exact.dat
  base64 /dev/urandom | head -c 1025 > malformed_large.dat

  git \
    -c "filter.lfs.process=" \
    -c "filter.lfs.clean=cat" \
    -c "filter.lfs.required=false" \
    add *.dat
  git commit -m "add malformed pointer"

  git push origin master

  pushd .. >/dev/null
    clone_repo "$reponame" "$reponame-assert"

    grep "malformed_small.dat" clone.log
    grep "malformed_exact.dat" clone.log
    grep "malformed_large.dat" clone.log

    expected_small="$(cat ../$reponame/malformed_small.dat)"
    expected_exact="$(cat ../$reponame/malformed_exact.dat)"
    expected_large="$(cat ../$reponame/malformed_large.dat)"

    actual_small="$(cat malformed_small.dat)"
    actual_exact="$(cat malformed_exact.dat)"
    actual_large="$(cat malformed_large.dat)"

    [ "$expected_small" = "$actual_small" ]
    [ "$expected_exact" = "$actual_exact" ]
    [ "$expected_large" = "$actual_large" ]
  popd >/dev/null
)
end_test
