#!/usr/bin/env bash

. "$(dirname "$0")/fixtures/migrate.sh"
. "$(dirname "$0")/testlib.sh"

begin_test "migrate info (default branch)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	140 B	1/1 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (bare repository)"
(
  set -e

  setup_multiple_remote_branches

  git lfs migrate info --everything
)
end_test

begin_test "migrate info (given branch)"
(
  set -e

  setup_multiple_local_branches

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info my-feature 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (default branch with filter)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --include "*.md" 2>&1 | tail -n 1) <(cat <<-EOF
	*.md	140 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "refs/heads/main" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (given branch with filter)"
(
  set -e

  setup_multiple_local_branches

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --include "*.md" my-feature 2>&1 | tail -n 1) <(cat <<-EOF
	*.md	170 B	2/2 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (default branch, exclude remote refs)"
(
  set -e

  setup_single_remote_branch

  git show-ref

  original_remote="$(git rev-parse refs/remotes/origin/main)"
  original_main="$(git rev-parse refs/heads/main)"

  diff -u <(git lfs migrate info 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	50 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse refs/remotes/origin/main)"
  migrated_main="$(git rev-parse refs/heads/main)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/remotes/origin/main" "$original_remote" "$migrated_remote"
)
end_test

begin_test "migrate info (given branch, exclude remote refs)"
(
  set -e

  setup_multiple_remote_branches

  original_remote="$(git rev-parse refs/remotes/origin/main)"
  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info my-feature 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	52 B	2/2 files(s)	100%
	*.txt	50 B	2/2 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse refs/remotes/origin/main)"
  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/remotes/origin/main" "$original_remote" "$migrated_remote"
  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (given ref, --skip-fetch)"
(
  set -e

  setup_single_remote_branch

  original_remote="$(git rev-parse refs/remotes/origin/main)"
  original_main="$(git rev-parse refs/heads/main)"

  git tag pseudo-remote "$original_remote"
  # Remove the refs/remotes/origin/main ref, and instruct 'git lfs migrate' to
  # not fetch it.
  git update-ref -d refs/remotes/origin/main

  diff -u <(git lfs migrate info --skip-fetch 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	190 B	2/2 files(s)	100%
	*.txt	150 B	2/2 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse pseudo-remote)"
  migrated_main="$(git rev-parse refs/heads/main)"

  assert_ref_unmoved "refs/remotes/origin/main" "$original_remote" "$migrated_remote"
  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
)
end_test

begin_test "migrate info (include/exclude ref)"
(
  set -e

  setup_multiple_remote_branches

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/main 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	31 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (include/exclude ref args)"
(
  set -e

  setup_multiple_remote_branches

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    my-feature ^main 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	31 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (include/exclude ref with filter)"
(
  set -e

  setup_multiple_remote_branches

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    --include="*.txt" \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/main 2>&1 | tail -n 1) <(cat <<-EOF
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (nested sub-trees, no filter)"
(
  set -e

  setup_single_local_branch_deep_trees

  original_main="$(git rev-parse refs/heads/main)"

  diff -u <(git lfs migrate info 2>/dev/null) <(cat <<-EOF
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
)
end_test

begin_test "migrate info (above threshold)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --above=130B 2>&1 | tail -n 1) <(cat <<-EOF
	*.md	140 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (above threshold, top)"
(
  set -e

  setup_multiple_local_branches

  base64 < /dev/urandom | head -c 160 > b.bin
  git add b.bin
  git commit -m "b.bin"

  original_head="$(git rev-parse HEAD)"

  # Ensure command reports only single highest entry due to --top=1 argument.
  diff -u <(git lfs migrate info --above=130B --top=1 2>&1 | tail -n 1) <(cat <<-EOF
	*.bin	160 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (top)"
(
  set -e

  setup_multiple_local_branches

  base64 < /dev/urandom | head -c 160 > b.bin
  git add b.bin
  git commit -m "b.bin"

  original_head="$(git rev-parse HEAD)"

  # Ensure command reports nothing if --top argument is less than zero.
  [ "0" -eq "$(git lfs migrate info --everything --top=-1 2>/dev/null | wc -l)" ]

  # Ensure command reports nothing if --top argument is zero.
  [ "0" -eq "$(git lfs migrate info --everything --top=0 2>/dev/null | wc -l)" ]

  # Ensure command reports no more entries than specified by --top argument.
  diff -u <(git lfs migrate info --everything --top=2 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.bin	160 B	1/1 files(s)	100%
	EOF)

  # Ensure command succeeds if --top argument is greater than total number of entries.
  diff -u <(git lfs migrate info --everything --top=10 2>&1 | tail -n 3) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.bin	160 B	1/1 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (given unit)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --unit=kb 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	0.1	1/1 files(s)	100%
	*.txt	0.1	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (doesn't show empty info entries)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  [ "0" -eq "$(git lfs migrate info --above=1mb 2>/dev/null | wc -l)" ]

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (empty set)"
(
  set -e

  setup_multiple_local_branches

  migrate="$(git lfs migrate info \
    --include-ref=refs/heads/main \
    --exclude-ref=refs/heads/main 2>/dev/null
  )"

  [ "0" -eq "$(echo -n "$migrate" | wc -c | awk '{ print $1 }')" ]
)
end_test

begin_test "migrate info (no-extension files)"
(
  set -e

  setup_multiple_local_branches_with_alternate_names
  git checkout main

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --everything 2>&1 | tail -n 2) <(cat <<-EOF
	no_extension	220 B	2/2 files(s)	100%
	*.txt       	170 B	2/2 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (--everything)"
(
  set -e

  setup_multiple_local_branches
  git checkout main

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --everything 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (--fixup, no .gitattributes)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports nothing if no files are tracked by LFS.
  [ "0" -eq "$(git lfs migrate info --everything --fixup 2>/dev/null | wc -l)" ]

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked)"
(
  set -e

  setup_single_local_branch_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure default command reports objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info 2>&1 | tail -n 3) <(cat <<-EOF
	*.gitattributes	83 B 	1/1 files(s)	100%

	LFS Objects    	260 B	2/2 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked, --pointers=follow)"
(
  set -e

  setup_single_local_branch_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure "follow" command reports objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info --pointers=follow 2>&1 | tail -n 3) <(cat <<-EOF
	*.gitattributes	83 B 	1/1 files(s)	100%

	LFS Objects    	260 B	2/2 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked, --pointers=no-follow)"
(
  set -e

  setup_single_local_branch_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure "no-follow" command reports pointers if all files are tracked by LFS.
  diff -u <(git lfs migrate info --pointers=no-follow 2>&1 | tail -n 3) <(cat <<-EOF
	*.md           	128 B	1/1 files(s)	100%
	*.txt          	128 B	1/1 files(s)	100%
	*.gitattributes	83 B 	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked, --pointers=ignore)"
(
  set -e

  setup_single_local_branch_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure "ignore" command reports no objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info --pointers=ignore 2>&1 | tail -n 1) <(cat <<-EOF
	*.gitattributes	83 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked, --fixup)"
(
  set -e

  setup_single_local_branch_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports nothing if all files are tracked by LFS.
  [ "0" -eq "$(git lfs migrate info --fixup 2>/dev/null | wc -l)" ]

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (all files tracked, --everything)"
(
  set -e

  setup_multiple_local_branches_tracked

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  # Ensure default command reports objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info --everything 2>&1 | tail -n 3) <(cat <<-EOF
	*.gitattributes	83 B 	1/1 files(s)	100%

	LFS Objects    	290 B	3/3 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (all files tracked, --everything and --pointers=follow)"
(
  set -e

  setup_multiple_local_branches_tracked

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  # Ensure "follow" command reports objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info --everything --pointers=follow 2>&1 | tail -n 3) <(cat <<-EOF
	*.gitattributes	83 B 	1/1 files(s)	100%

	LFS Objects    	290 B	3/3 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (all files tracked, --everything and --pointers=no-follow)"
(
  set -e

  setup_multiple_local_branches_tracked

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  # Ensure "no-follow" command reports pointers if all files are tracked by LFS.
  diff -u <(git lfs migrate info --everything --pointers=no-follow 2>&1 | tail -n 3) <(cat <<-EOF
	*.md           	255 B	2/2 files(s)	100%
	*.txt          	128 B	1/1 files(s)	100%
	*.gitattributes	83 B 	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (all files tracked, --everything and --pointers=ignore)"
(
  set -e

  setup_multiple_local_branches_tracked

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  # Ensure "ignore" command reports no objects if all files are tracked by LFS.
  diff -u <(git lfs migrate info --everything --pointers=ignore 2>&1 | tail -n 1) <(cat <<-EOF
	*.gitattributes	83 B	1/1 files(s)	100%
	EOF)

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (all files tracked, --everything and --fixup)"
(
  set -e

  setup_multiple_local_branches_tracked

  original_main="$(git rev-parse refs/heads/main)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  # Ensure "fixup" command reports nothing if all files are tracked by LFS.
  [ "0" -eq "$(git lfs migrate info --everything --fixup 2>/dev/null | wc -l)" ]

  migrated_main="$(git rev-parse refs/heads/main)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/main" "$original_main" "$migrated_main"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (potential fixup)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  original_head="$(git rev-parse HEAD)"

  # Ensure command reports files which should be tracked but have not been
  # stored properly as LFS pointers.
  diff -u <(git lfs migrate info 2>&1 | tail -n 2) <(cat <<-EOF
	*.txt          	120 B	1/1 files(s)	100%
	*.gitattributes	42 B 	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (potential fixup, --fixup)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports files which should be tracked but have not
  # been stored properly as LFS pointers, and ignores .gitattributes files.
  diff -u <(git lfs migrate info --fixup 2>&1 | tail -n 1) <(cat <<-EOF
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (potential fixup, complex nested)"
(
  set -e

  setup_single_local_branch_complex_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure command reports the file which should be tracked but has not been
  # stored properly (a.txt) and the file which is not tracked (dir/b.txt).
  diff -u <(git lfs migrate info 2>&1 | tail -n 2) <(cat <<-EOF
	*.gitattributes	69 B	2/2 files(s)	100%
	*.txt          	2 B 	2/2 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (potential fixup, complex nested, --fixup)"
(
  set -e

  setup_single_local_branch_complex_tracked

  original_head="$(git rev-parse HEAD)"

  # Ensure "fixup" command reports the file which should be tracked but has not
  # been stored properly (a.txt), and ignores .gitattributes files and
  # the file which is not tracked (dir/b.txt).
  diff -u <(git lfs migrate info --fixup 2>&1 | tail -n 1) <(cat <<-EOF
	*.txt	1 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (ambiguous reference)"
(
  set -e

  setup_multiple_local_branches

  # Create an ambiguously named reference sharing the name as the SHA-1 of
  # "HEAD".
  sha="$(git rev-parse HEAD)"
  git tag "$sha"

  git lfs migrate info --everything
)
end_test

begin_test "migrate info (--everything with args)"
(
  set -e

  setup_multiple_local_branches

  git lfs migrate info --everything main 2>&1 | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 1 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --everything with explicit reference arguments" \
    migrate.log
)
end_test

begin_test "migrate info (--everything with --include-ref)"
(
  set -e

  setup_multiple_local_branches

  git lfs migrate info --everything --include-ref=refs/heads/main 2>&1 | \
    tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 1 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --everything with --include-ref or --exclude-ref" \
    migrate.log
)
end_test

begin_test "migrate info (--everything with --exclude-ref)"
(
  set -e

  setup_multiple_local_branches

  git lfs migrate info --everything --exclude-ref=refs/heads/main 2>&1 | \
    tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 1 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --everything with --include-ref or --exclude-ref" \
    migrate.log
)
end_test

begin_test "migrate info (--pointers invalid)"
(
  set -e

  setup_multiple_local_branches

  git lfs migrate info --everything --pointers=foo 2>&1 | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 1 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: unsupported --pointers option value" migrate.log
)
end_test

begin_test "migrate info (--fixup, --pointers=follow)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate info --everything --fixup --pointers=follow 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 1 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --pointers=follow" migrate.log
)
end_test

begin_test "migrate info (--fixup, --pointers=no-follow)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate info --everything --fixup --pointers=no-follow 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --pointers=no-follow" migrate.log
)
end_test

begin_test "migrate info (--fixup, --include)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate info --everything --fixup --include="*.txt" 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --include, --exclude" migrate.log
)
end_test

begin_test "migrate info (--fixup, --exclude)"
(
  set -e

  setup_single_local_branch_tracked_corrupt

  git lfs migrate info --everything --fixup --exclude="*.txt" 2>&1 \
    | tee migrate.log

  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "fatal: expected 'git lfs migrate ...' to fail, didn't ..."
    exit 1
  fi

  grep -q "fatal: cannot use --fixup with --include, --exclude" migrate.log
)
end_test
