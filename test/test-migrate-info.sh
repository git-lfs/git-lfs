#!/usr/bin/env bash

. "test/test-migrate-fixtures.sh"
. "test/testlib.sh"

begin_test "migrate info (default branch)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --above=0b 2>&1) <(cat <<-EOF
	*.md 	140 B	1/1 files(s)	100%
	*.txt	120 B	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate info (given branch)"
(
  set -e

  setup_multiple_local_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info --above=0b my-feature 2>&1) <(cat <<-EOF
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

  diff -u <(git lfs migrate info --above=0b --include "*.md" 2>&1) <(cat <<-EOF
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

  diff -u <(git lfs migrate info --above=0b --include "*.md" my-feature 2>&1) <(cat <<-EOF
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

  diff -u <(git lfs migrate info --above=0b 2>&1) <(cat <<-EOF
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

  diff -u <(git lfs migrate info --above=0b my-feature 2>&1) <(cat <<-EOF
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

begin_test "migrate info (include/exclude ref)"
(
  set -e

  setup_multiple_remote_branches

  original_master="$(git rev-parse refs/heads/master)"
  original_feature="$(git rev-parse refs/heads/my-feature)"

  diff -u <(git lfs migrate info \
    --above=0b \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master 2>&1) <(cat <<-EOF
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
    --above=0b \
    --include="*.txt" \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master 2>&1) <(cat <<-EOF
	*.txt	30 B	1/1 files(s)	100%
	EOF)

  migrated_master="$(git rev-parse refs/heads/master)"
  migrated_feature="$(git rev-parse refs/heads/my-feature)"

  assert_ref_unmoved "refs/heads/master" "$original_master" "$migrated_master"
  assert_ref_unmoved "refs/heads/my-feature" "$original_feature" "$migrated_feature"
)
end_test

begin_test "migrate info (above threshold)"
(
  set -e

  setup_multiple_local_branches

  original_head="$(git rev-parse HEAD)"

  diff -u <(git lfs migrate info --above=130B 2>&1) <(cat <<-EOF
	*.md 	140 B	1/1 files(s)	100%
	*.txt	0 B  	0/1 files(s)	  0%
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

  diff -u <(git lfs migrate info --above=130B --top=1 2>&1) <(cat <<-EOF
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

  diff -u <(git lfs migrate info --above=0b --unit=kb 2>&1) <(cat <<-EOF
	*.md 	0.1	1/1 files(s)	100%
	*.txt	0.1	1/1 files(s)	100%
	EOF)

  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test
