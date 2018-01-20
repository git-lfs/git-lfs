#!/usr/bin/env bash

. "test/test-migrate-fixtures.sh"
. "test/testlib.sh"

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

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info my-feature 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
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

  assert_ref_unmoved "refs/heads/master" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (given branch with filter)"
(
  set -e

  setup_multiple_local_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --include "*.md" my-feature 2>&1 | tail -n 1) <(cat <<-EOF
	*.md	170 B	2/2 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (default branch, exclude remote refs)"
(
  set -e

  setup_single_remote_branch

  git show-ref

  original_remote="$(git rev-parse refs/remotes/origin/master)"
  original_master="$(git rev-parse refs/heads/master)"

  diff -u <(git lfs migrate info 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	50 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse refs/remotes/origin/master)"
  migrated_master="$(git rev-parse refs/heads/master)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/remotes/origin/master" "$original_remote" "$migrated_remote"
)
end_test

begin_test "migrate info (given branch, exclude remote refs)"
(
  set -e

  setup_multiple_remote_branches

  original_remote="$(git rev-parse refs/remotes/origin/master)"
  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info my-feature 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	52 B	2/2 files(s)	100%
	*.txt	50 B	2/2 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse refs/remotes/origin/master)"
  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/remotes/origin/master" "$original_remote" "$migrated_remote"
  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (given ref, --skip-fetch)"
(
  set -e

  setup_single_remote_branch

  original_remote="$(git rev-parse refs/remotes/origin/master)"
  original_master="$(git rev-parse refs/heads/master)"

  git tag pseudo-remote "$original_remote"
  # Remove the refs/remotes/origin/master ref, and instruct 'git lfs migrate' to
  # not fetch it.
  git update-ref -d refs/remotes/origin/master

  diff -u <(git lfs migrate info --skip-fetch 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	190 B	2/2 files(s)	100%
	*.txt	150 B	2/2 files(s)	100%
	EOF)

  migrated_remote="$(git rev-parse pseudo-remote)"
  migrated_master="$(git rev-parse refs/heads/master)"

  assert_ref_unmoved "refs/remotes/origin/master" "$original_remote" "$migrated_remote"
  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
)
end_test

begin_test "migrate info (include/exclude ref)"
(
  set -e

  setup_multiple_remote_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	31 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (include/exclude ref args)"
(
  set -e

  setup_multiple_remote_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    my-feature ^master 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	31 B	1/1 files(s)	100%
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (include/exclude ref with filter)"
(
  set -e

  setup_multiple_remote_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    --include="*.txt" \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master 2>&1 | tail -n 1) <(cat <<-EOF
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (nested sub-trees, no filter)"
(
  set -e

  setup_single_local_branch_deep_trees

  original_master="$(git rev-parse refs/heads/master)"

  diff -u <(git lfs migrate info 2>/dev/null) <(cat <<-EOF
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
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

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --above=130B --top=1 2>&1 | tail -n 1) <(cat <<-EOF
	*.md	140 B	1/1 files(s)	100%
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
    --include-ref=refs/heads/master \
    --exclude-ref=refs/heads/master 2>/dev/null
  )"

  [ "0" -eq "$(echo -n "$migrate" | wc -l | awk '{ print $1 }')" ]
)
end_test

begin_test "migrate info (--everything)"
(
  set -e

  setup_multiple_local_branches
  git checkout master

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --everything 2>&1 | tail -n 2) <(cat <<-EOF
	*.md 	170 B	2/2 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
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

  [ "$(git lfs migrate info --everything master 2>&1)" = \
    "fatal: cannot use --everything with explicit reference arguments" ]
)
end_test

begin_test "migrate info (--everything with --include-ref)"
(
  set -e

  setup_multiple_local_branches

  [ "$(git lfs migrate info --everything --include-ref=refs/heads/master 2>&1)" = \
    "fatal: cannot use --everything with --include-ref or --exclude-ref" ]
)
end_test

exit 0

begin_test "migrate info (--everything with --exclude-ref)"
(
  set -e

  setup_multiple_local_branches

  [ "$(git lfs migrate info --everything --exclude-ref=refs/heads/master 2>&1)" = \
    "fatal: cannot use --everything with --include-ref or --exclude-ref" ]
)
end_test
