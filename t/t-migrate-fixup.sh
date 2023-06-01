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
  refute_pointer "refs/heads/main" "dir/b.txt"

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
    echo >&2 "Expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "Cannot use --fixup with --include, --exclude" migrate.log
)
end_test

begin_test "migrate import (--fixup, --exclude)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate import --everything --fixup --yes --exclude="*.txt" 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "Expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "Cannot use --fixup with --include, --exclude" migrate.log
)
end_test

begin_test "migrate import (--fixup, --no-rewrite)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate import --everything --fixup --yes --no-rewrite 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "Expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -qe "--no-rewrite and --fixup cannot be combined" migrate.log
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

begin_test "migrate import (--fixup, .gitattributes symlink)"
(
  set -e

  setup_single_local_branch_tracked_corrupt link

  git lfs migrate import --everything --fixup --yes 2>&1 | tee migrate.log
  if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo >&2 "fatal: expected git lfs migrate import to fail, didn't"
    exit 1
  fi

  grep "migrate: expected '.gitattributes' to be a file, got a symbolic link" migrate.log

  main="$(git rev-parse refs/heads/main)"

  attrs_main_sha="$(git show $main:.gitattributes | git hash-object --stdin)"

  diff -u <(git ls-tree $main -- .gitattributes) <(cat <<-EOF
120000 blob $attrs_main_sha	.gitattributes
EOF
  )
)
end_test

begin_test "migrate import (--fixup, .gitattributes with macro)"
(
  set -e

  setup_single_local_branch_tracked_corrupt macro

  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"

  git lfs migrate import --everything --fixup --yes

  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
  assert_local_object "$txt_oid" "120"

  main="$(git rev-parse refs/heads/main)"
  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  echo "$main_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

# NOTE: We skip this test for now as the "git lfs migrate" commands do not
#       fully process macro attribute definitions yet.
#begin_test "migrate info (--fixup, .gitattributes with LFS macro)"
#(
#  set -e
#
#  setup_single_local_branch_tracked_corrupt lfsmacro
#
#  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
#
#  git lfs migrate import --everything --fixup --yes
#
#  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
#  assert_local_object "$txt_oid" "120"
#
#  main="$(git rev-parse refs/heads/main)"
#  main_attrs="$(git cat-file -p "$main:.gitattributes")"
#  echo "$main_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
#)
#end_test

begin_test "migrate import (no potential fixup, --fixup, no .gitattributes)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports nothing if no files are tracked by LFS.
  git lfs migrate import --everything --fixup --yes >migrate.log
  [ "0" -eq "$(cat migrate.log | wc -l)" ]

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate import (no potential fixup, --fixup, .gitattributes with macro)"
(
  set -e

  setup_multiple_local_branches

  echo "[attr]foo foo" >.gitattributes
  base64 < /dev/urandom | head -c 30 > a.md
  git add .gitattributes a.md
  git commit -m macro

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports nothing if no files are tracked by LFS.
  git lfs migrate import --everything --fixup --yes >migrate.log
  [ "0" -eq "$(cat migrate.log | wc -l)" ]

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test
