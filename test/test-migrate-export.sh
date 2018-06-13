#!/usr/bin/env bash

. "test/test-migrate-fixtures.sh"
. "test/testlib.sh"

begin_test "migrate export (default branch)"
(
  set -e

  setup_multiple_local_branches

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"

  git lfs migrate import

  assert_pointer "refs/heads/master" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "120"

  git lfs migrate export --include="*.md, *.txt"

  [ ! $(assert_pointer "refs/heads/master" "a.md" "$md_oid" "140") ]
  [ ! $(assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "120") ]

  master="$(git rev-parse refs/heads/master)"
  master_attrs="$(git cat-file -p "$master:.gitattributes")"

  echo "$master_attrs" | grep -q "*.md text -filter -merge -diff"
  echo "$master_attrs" | grep -q "*.txt text -filter -merge -diff"
)
end_test
