#!/bin/sh

. "test/testlib.sh"

begin_test "ext"
(
  set -e

  # no need to setup a remote repo, since this test does not need to push or pull

  mkdir ext
  cd ext
  git init

  git config lfs-ext.foo.clean "foo-clean %f"
  git config lfs-ext.foo.smudge "foo-smudge %f"
  git config lfs-ext.foo.priority 0

  git config lfs-ext.bar.clean "bar-clean %f"
  git config lfs-ext.bar.smudge "bar-smudge %f"
  git config lfs-ext.bar.priority 1

  git config lfs-ext.baz.clean "baz-clean %f"
  git config lfs-ext.baz.smudge "baz-smudge %f"
  git config lfs-ext.baz.priority 2

  fooExpected="Extension: foo
    clean = foo-clean %f
    smudge = foo-smudge %f
    priority = 0"

  barExpected="Extension: bar
    clean = bar-clean %f
    smudge = bar-smudge %f
    priority = 1"

  bazExpected="Extension: baz
    clean = baz-clean %f
    smudge = baz-smudge %f
    priority = 2"

  actual=$(git lfs ext list foo)
  [ "$actual" = "$fooExpected" ]

  actual=$(git lfs ext list bar)
  [ "$actual" = "$barExpected" ]

  actual=$(git lfs ext list baz)
  [ "$actual" = "$bazExpected" ]

  actual=$(git lfs ext list foo bar)
  expected=$(printf "%s\n%s" "$fooExpected" "$barExpected")
  [ "$actual" = "$expected" ]

  actual=$(git lfs ext list)
  expected=$(printf "%s\n%s\n%s" "$fooExpected" "$barExpected" "$bazExpected")
  [ "$actual" = "$expected" ]

  actual=$(git lfs ext)
  [ "$actual" = "$expected" ]
)
end_test
