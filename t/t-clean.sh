#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

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

  contents="version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 9999

This is my test pointer.  There are many like it, but this one is mine."
  contents_oid="$(calc_oid "$contents")"

  printf "%s" "$contents" | git lfs clean | tee clean.log
  [ "$(pointer "$contents_oid" "${#contents}")" = "$(cat clean.log)" ]
)
end_test

begin_test "clean pseudo pointer with extra data"
(
  set -e
  clean_setup "extra-data"

  # Test with an invalid pointer larger than the size of the read buffer used
  # when decoding pointers.  See https://github.com/git-lfs/git-lfs/pull/271.
  max_pointer_size="$(lfstest-getlimit --max-pointer-size)"
  fill="$(head -c "$max_pointer_size" /dev/zero | tr '\0' '\n'; printf "EOF")"
  contents="$(printf "version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 9999
${fill%EOF}
This is my test pointer.  There are many like it, but this one is mine.")"
  contents_oid="$(calc_oid "$contents")"

  printf "%s" "$contents" | git lfs clean | tee clean.log
  [ "$(pointer "$contents_oid" "${#contents}")" = "$(cat clean.log)" ]
)
end_test

begin_test "clean with pointer extension"
(
  set -e
  clean_setup "pointer-extension"

  setup_case_inverter_extension

  contents="$(printf "%s\n%s" "abc" "def")"
  contents_oid="$(calc_oid "$contents")"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  printf "%s" "$contents" | git lfs clean -- "dir1/abc.dat" | tee clean.log

  pointer="$(case_inverter_extension_pointer "$contents_oid" "$inverted_contents_oid" 7)"

  assert_local_object "$inverted_contents_oid" 7

  [ "$pointer" = "$(cat clean.log)" ]
  grep "clean: dir1/abc.dat" "$LFSTEST_EXT_LOG"
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

  # Test with file sizes equal to and larger than the
  # size of the read buffer used when decoding pointers.
  # See https://github.com/git-lfs/git-lfs/issues/2487
  # and https://github.com/git-lfs/git-lfs/pull/2488.
  max_pointer_size="$(lfstest-getlimit --max-pointer-size)"
  lfstest-genrandom --base64 "$max_pointer_size" >small.dat
  lfstest-genrandom --base64 $((max_pointer_size * 2)) >large.dat

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
