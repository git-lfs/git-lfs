#!/usr/bin/env bash

. "test/testlib.sh"

clean_setup () {
  mkdir "$1"
  cd "$1"
  git init
}

begin_test "clean simple file"
(
  set -e
  clean_setup "simple"

  echo "whatever" | git lfs clean | tee clean.log
  [ "$(pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9)" = "$(cat clean.log)" ]
)
end_test

begin_test "clean a pointer"
(
  set -e
  clean_setup "pointer"

  pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9 | git lfs clean | tee clean.log
  [ "$(pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9)" = "$(cat clean.log)" ]
)
end_test

begin_test "clean pseudo pointer"
(
  set -e
  clean_setup "pseudo"

  echo "version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 1024

This is my test pointer.  There are many like it, but this one is mine." | git lfs clean | tee clean.log
  [ "$(pointer f492acbebb5faa22da4c1501c022af035469f624f426631f31936575873fefe1 202)" = "$(cat clean.log)" ]
)
end_test

begin_test "clean pseudo pointer with extra data"
(
  set -e
  clean_setup "extra-data"

  # pointer includes enough extra data to fill the 'git lfs clean' buffer
  printf "version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 1024
\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n
This is my test pointer.  There are many like it, but this one is mine.\n" | git lfs clean | tee clean.log
  [ "$(pointer c2f909f6961bf85a92e2942ef3ed80c938a3d0ebaee6e72940692581052333be 586)" = "$(cat clean.log)" ]
)
end_test

begin_test "clean stdin"
(
  set -e

  # git-lfs-clean(1) writes to .git/lfs/objects, and therefore must be executed
  # within a repository.
  reponame="clean-over-stdin"
  git init "$reponame"
  cd "$reponame"

  base64 /dev/urandom | head -c 1024 > small.dat
  base64 /dev/urandom | head -c 2048 > large.dat

  expected_small="$(calc_oid_file "small.dat")"
  expected_large="$(calc_oid_file "large.dat")"

  actual_small="$(git lfs clean < "small.dat" | grep "oid" | cut -d ':' -f 2)"
  actual_large="$(git lfs clean < "large.dat" | grep "oid" | cut -d ':' -f 2)"

  if [ "$expected_small" != "$actual_small" ]; then
    echo >&2 "fatal: expected small OID of: $expected_small, got: $actual_small"
    exit 1
  fi

  if [ "$expected_large" != "$actual_large" ]; then
    echo >&2 "fatal: expected large OID of: $expected_large, got: $actual_large"
    exit 1
  fi
)
end_test
