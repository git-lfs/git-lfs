#!/usr/bin/env bash

. "test/test-migrate-fixtures.sh"
. "test/testlib.sh"

begin_test "migrate import (default branch)"
(
  set -e

  setup_multiple_local_branches

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_feature_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"

  git lfs migrate import

  assert_pointer "refs/heads/master" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "120"

  assert_local_object "$md_oid" "140"
  assert_local_object "$txt_oid" "120"
  refute_local_object "$md_feature_oid" "30"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  [ ! $(git cat-file -p "$feature:.gitattributes") ]

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (given branch)"
(
  set -e

  setup_multiple_local_branches

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_feature_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"

  git lfs migrate import my-feature

  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"
  assert_pointer "refs/heads/my-feature" "a.txt" "$txt_oid" "120"
  assert_pointer "refs/heads/master" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "120"

  assert_local_object "$md_oid" "140"
  assert_local_object "$md_feature_oid" "30"
  assert_local_object "$txt_oid" "120"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (default branch with filter)"
(
  set -e

  setup_multiple_local_branches

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_feature_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"

  git lfs migrate import --include "*.md"

  assert_pointer "refs/heads/master" "a.md" "$md_oid" "140"

  assert_local_object "$md_oid" "140"
  refute_local_object "$txt_oid" "120"
  refute_local_object "$md_feature_oid" "30"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  [ ! $(git cat-file -p "$feature:.gitattributes") ]

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -vq "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (given branch with filter)"
(
  set -e

  setup_multiple_local_branches

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_feature_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"

  git lfs migrate import --include "*.md" my-feature

  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"
  assert_pointer "refs/heads/my-feature~1" "a.md" "$md_oid" "140"

  assert_local_object "$md_oid" "140"
  assert_local_object "$md_feature_oid" "30"
  refute_local_object "$txt_oid" "120"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -vq "*.txt filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -vq "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (default branch, exclude remote refs)"
(
  set -e

  setup_single_remote_branch

  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"
  md_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  txt_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"

  git lfs migrate import

  assert_pointer "refs/heads/master" "a.md" "$md_oid" "50"
  assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "30"

  assert_local_object "$md_oid" "50"
  assert_local_object "$txt_oid" "30"
  refute_local_object "$md_remote_oid" "140"
  refute_local_object "$txt_remote_oid" "120"

  master="$(git rev-parse refs/heads/master)"
  remote="$(git rev-parse refs/remotes/origin/master)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  [ ! $(git cat-file -p "$remote:.gitattributes") ]

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -vq "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (given branch, exclude remote refs)"
(
  set -e

  setup_multiple_remote_branches

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  md_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"
  txt_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.txt")")"

  git lfs migrate import my-feature

  assert_pointer "refs/heads/master" "a.md" "$md_master_oid" "21"
  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "31"
  assert_pointer "refs/heads/master" "a.txt" "$txt_master_oid" "20"
  assert_pointer "refs/heads/my-feature" "a.txt" "$txt_feature_oid" "30"

  assert_local_object "$md_feature_oid" "31"
  assert_local_object "$md_master_oid" "21"
  assert_local_object "$txt_feature_oid" "30"
  assert_local_object "$txt_master_oid" "20"
  refute_local_object "$md_remote_oid" "11"
  refute_local_object "$txt_remote_oid" "10"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"
  remote="$(git rev-parse refs/remotes/origin/master)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  [ ! $(git cat-file -p "$remote:.gitattributes") ]
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -vq "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (include/exclude ref)"
(
  set -e

  setup_multiple_remote_branches

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  md_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"
  txt_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.txt")")"

  git lfs migrate import \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master

  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "31"
  assert_pointer "refs/heads/my-feature" "a.txt" "$txt_feature_oid" "30"

  assert_local_object "$md_feature_oid" "31"
  refute_local_object "$md_master_oid" "21"
  assert_local_object "$txt_feature_oid" "30"
  refute_local_object "$txt_master_oid" "20"
  refute_local_object "$md_remote_oid" "11"
  refute_local_object "$txt_remote_oid" "10"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"
  remote="$(git rev-parse refs/remotes/origin/master)"

  [ ! $(git cat-file -p "$master:.gitattributes") ]
  [ ! $(git cat-file -p "$remote:.gitattributes") ]
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$feature_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (include/exclude ref with filter)"
(
  set -e

  setup_multiple_remote_branches

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  md_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"
  txt_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.txt")")"

  git lfs migrate import \
    --include="*.txt" \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/master

  assert_pointer "refs/heads/my-feature" "a.txt" "$txt_feature_oid" "30"

  refute_local_object "$md_feature_oid" "31"
  refute_local_object "$md_master_oid" "21"
  assert_local_object "$txt_feature_oid" "30"
  refute_local_object "$txt_master_oid" "20"
  refute_local_object "$md_remote_oid" "11"
  refute_local_object "$txt_remote_oid" "10"

  master="$(git rev-parse refs/heads/master)"
  feature="$(git rev-parse refs/heads/my-feature)"
  remote="$(git rev-parse refs/remotes/origin/master)"

  [ ! $(git cat-file -p "$master:.gitattributes") ]
  [ ! $(git cat-file -p "$remote:.gitattributes") ]
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$feature_attrs" | grep -vq "*.md filter=lfs diff=lfs merge=lfs"
  echo "$feature_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (existing .gitattributes)"
(
  set -e

  setup_local_branch_with_gitattrs

  pwd

  master="$(git rev-parse refs/heads/master)"

  txt_master_oid="$(calc_oid "$(git cat-file -p "$master:a.txt")")"

  git lfs migrate import --include-ref=refs/heads/master --include="*.txt"

  assert_local_object "$txt_master_oid" "120"

  master="$(git rev-parse refs/heads/master)"
  prev="$(git rev-parse refs/heads/master^1)"

  diff -u <(git cat-file -p $master:.gitattributes) <(cat <<-EOF
*.txt filter=lfs diff=lfs merge=lfs -text
*.other filter=lfs diff=lfs merge=lfs -text
EOF)

  diff -u <(git cat-file -p $prev:.gitattributes) <(cat <<-EOF
*.txt filter=lfs diff=lfs merge=lfs -text
EOF)
)
end_test
