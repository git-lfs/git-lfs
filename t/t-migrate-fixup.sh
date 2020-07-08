#!/usr/bin/env bash

. "$(dirname "$0")/fixtures/migrate.sh"
. "$(dirname "$0")/testlib.sh"

begin_test "migrate import (--fixup)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"

  git lfs migrate import --everything --fixup --yes

  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
  assert_local_object "$txt_oid" "120"

  main="$(git rev-parse refs/heads/main)"
  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  echo "$main_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (--fixup, complex nested)"
(
  set -e

  setup_single_local_branch_complex_tracked

  a_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  b_oid="$(calc_oid "$(git cat-file -p :dir/b.txt)")"

  git lfs migrate import --everything --fixup --yes

  assert_pointer "refs/heads/main" "a.txt" "$a_oid" "1"
  refute_pointer "refs/heads/main" "b.txt"

  assert_local_object "$a_oid" "1"
  refute_local_object "$b_oid" "1"

  main="$(git rev-parse refs/heads/main)"
  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  main_dir_attrs="$(git cat-file -p "$main:dir/.gitattributes")"
  echo "$main_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
  echo "$main_dir_attrs" | grep -q "*.txt !filter !diff !merge"
)
end_test

begin_test "migrate import (--fixup, --include)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate import --everything --fixup --yes --include="*.txt" 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --include, --exclude" migrate.log
)
end_test

begin_test "migrate import (--fixup, --exclude)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate import --everything --fixup --yes --exclude="*.txt" 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --include, --exclude" migrate.log
)
end_test

begin_test "migrate import (--fixup, --no-rewrite)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate import --everything --fixup --yes --no-rewrite 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: --no-rewrite and --fixup cannot be combined" migrate.log
)
end_test

begin_test "migrate import (--fixup with remote tags)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs uninstall

  base64 < /dev/urandom | head -c 120 > b.txt
  git add b.txt
  git commit -m "b.txt"

  git tag -m tag1 -a tag1
  git reset --hard HEAD^

  git lfs install

  cwd=$(pwd)
  cd "$TRASHDIR"

  git clone "$cwd" "$reponame-2"
  cd "$reponame-2"

  # We're checking here that this succeeds even though it does nothing in this
  # case.
  git lfs migrate import --fixup --yes main
)
end_test
