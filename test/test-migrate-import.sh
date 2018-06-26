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

begin_test "migrate import (bare repository)"
(
  set -e

  setup_multiple_remote_branches

  git lfs migrate import --everything
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

begin_test "migrate import (given ref, --skip-fetch)"
(
  set -e

  setup_single_remote_branch

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"

  git tag pseudo-remote "$(git rev-parse refs/remotes/origin/master)"
  # Remove the refs/remotes/origin/master ref, and instruct 'git lfs migrate' to
  # not fetch it.
  git update-ref -d refs/remotes/origin/master

  git lfs migrate import --skip-fetch

  assert_pointer "refs/heads/master" "a.md" "$md_master_oid" "50"
  assert_pointer "pseudo-remote" "a.md" "$md_remote_oid" "140"
  assert_pointer "refs/heads/master" "a.txt" "$txt_master_oid" "30"
  assert_pointer "pseudo-remote" "a.txt" "$txt_remote_oid" "120"

  assert_local_object "$md_master_oid" "50"
  assert_local_object "$txt_master_oid" "30"
  assert_local_object "$md_remote_oid" "140"
  assert_local_object "$txt_remote_oid" "120"

  master="$(git rev-parse refs/heads/master)"
  remote="$(git rev-parse pseudo-remote)"

  master_attrs="$(git cat-file -p "$master:.gitattributes")"
  remote_attrs="$(git cat-file -p "$remote:.gitattributes")"

  echo "$master_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$master_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
  echo "$remote_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"
  echo "$remote_attrs" | grep -q "*.txt filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate import (un-annotated tags)"
(
  set -e

  setup_single_local_branch_with_tags

  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"

  git lfs migrate import --everything

  assert_pointer "refs/heads/master" "a.txt" "$txt_master_oid" "2"
  assert_local_object "$txt_master_oid" "2"

  git tag --points-at "$(git rev-parse HEAD)" | grep -q "v1.0.0"
)
end_test

begin_test "migrate import (annotated tags)"
(
  set -e

  setup_single_local_branch_with_annotated_tags

  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"

  git lfs migrate import --everything

  assert_pointer "refs/heads/master" "a.txt" "$txt_master_oid" "2"
  assert_local_object "$txt_master_oid" "2"

  git tag --points-at "$(git rev-parse HEAD)" | grep -q "v1.0.0"
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

begin_test "migrate import (include/exclude ref args)"
(
  set -e

  setup_multiple_remote_branches

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  md_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.md")")"
  md_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  txt_remote_oid="$(calc_oid "$(git cat-file -p "refs/remotes/origin/master:a.txt")")"
  txt_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.txt")")"

  git lfs migrate import my-feature ^master

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

begin_test "migrate import (bare repository)"
(
  set -e

  setup_multiple_local_branches
  make_bare

  git lfs migrate import \
    --include-ref=master
)
end_test

begin_test "migrate import (nested sub-trees, no filter)"
(
  set -e

  setup_single_local_branch_deep_trees

  oid="$(calc_oid "$(git cat-file -p :foo/bar/baz/a.txt)")"
  size="$(git cat-file -p :foo/bar/baz/a.txt | wc -c | awk '{ print $1 }')"

  git lfs migrate import --everything

  assert_local_object "$oid" "$size"
)
end_test

begin_test "migrate import (prefix include(s))"
(
  set -e

  includes="foo/bar/baz foo/**/baz/a.txt *.txt"
  for include in $includes; do
    setup_single_local_branch_deep_trees

    oid="$(calc_oid "$(git cat-file -p :foo/bar/baz/a.txt)")"

    git lfs migrate import --include="$include"

    assert_local_object "$oid" 120

    cd ..
  done
)
end_test

begin_test "migrate import (--everything)"
(
  set -e

  setup_multiple_local_branches
  git checkout master

  master_txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  master_md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  feature_md_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"
  master_txt_size="$(git cat-file -p :a.txt | wc -c | awk '{ print $1 }')"
  master_md_size="$(git cat-file -p :a.md | wc -c | awk '{ print $1 }')"
  feature_md_size="$(git cat-file -p my-feature:a.md | wc -c | awk '{ print $1 }')"

  git lfs migrate import --everything

  assert_pointer "master" "a.txt" "$master_txt_oid" "$master_txt_size"
  assert_pointer "master" "a.md" "$master_md_oid" "$master_md_size"
  assert_pointer "my-feature" "a.md" "$feature_md_oid" "$feature_md_size"
)
end_test

begin_test "migrate import (ambiguous reference)"
(
  set -e

  setup_multiple_local_branches

  # Create an ambiguously named reference sharing the name as the SHA-1 of
  # "HEAD".
  sha="$(git rev-parse HEAD)"
  git tag "$sha"

  git lfs migrate import --everything
)
end_test

begin_test "migrate import (--everything with args)"
(
  set -e

  setup_multiple_local_branches

  [ "$(git lfs migrate import --everything master 2>&1)" = \
    "fatal: cannot use --everything with explicit reference arguments" ]
)
end_test

begin_test "migrate import (--everything with --include-ref)"
(
  set -e

  setup_multiple_local_branches

  [ "$(git lfs migrate import --everything --include-ref=refs/heads/master 2>&1)" = \
    "fatal: cannot use --everything with --include-ref or --exclude-ref" ]
)
end_test

begin_test "migrate import (--everything with --exclude-ref)"
(
  set -e

  setup_multiple_local_branches

  [ "$(git lfs migrate import --everything --exclude-ref=refs/heads/master 2>&1)" = \
    "fatal: cannot use --everything with --include-ref or --exclude-ref" ]
)
end_test

begin_test "migrate import (--everything and --include with glob pattern)"
(
  set -e

  setup_multiple_local_branches

  md_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.md")")"
  txt_master_oid="$(calc_oid "$(git cat-file -p "refs/heads/master:a.txt")")"
  md_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.md")")"
  txt_feature_oid="$(calc_oid "$(git cat-file -p "refs/heads/my-feature:a.txt")")"

  git lfs migrate import --verbose --everything --include='*.[mM][dD]'

  assert_pointer "refs/heads/master" "a.md" "$md_master_oid" "140"
  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"

  assert_local_object "$md_master_oid" "140"
  assert_local_object "$md_feature_oid" "30"
  refute_local_object "$txt_master_oid"
  refute_local_object "$txt_feature_oid"
)
end_test

begin_test "migrate import (nested sub-trees and --include with wildcard)"
(
  set -e

  setup_single_local_branch_deep_trees

  oid="$(calc_oid "$(git cat-file -p :foo/bar/baz/a.txt)")"
  size="$(git cat-file -p :foo/bar/baz/a.txt | wc -c | awk '{ print $1 }')"

  git lfs migrate import --include="**/*ar/**"

  assert_pointer "refs/heads/master" "foo/bar/baz/a.txt" "$oid" "$size"
  assert_local_object "$oid" "$size"
)
end_test

begin_test "migrate import (handle copies of files)"
(
  set -e

  setup_single_local_branch_deep_trees

  # add the object from the sub-tree to the root directory
  cp foo/bar/baz/a.txt a.txt
  git add a.txt
  git commit -m "duplicated file"

  oid_root="$(calc_oid "$(git cat-file -p :a.txt)")"
  oid_tree="$(calc_oid "$(git cat-file -p :foo/bar/baz/a.txt)")"
  size="$(git cat-file -p :foo/bar/baz/a.txt | wc -c | awk '{ print $1 }')"

  # only import objects under "foo"
  git lfs migrate import --include="foo/**"

  assert_pointer "refs/heads/master" "foo/bar/baz/a.txt" "$oid_tree" "$size"
  assert_local_object "$oid_tree" "$size"

  # "a.txt" is not under "foo" and therefore should not be in LFS
  oid_root_after_migration="$(calc_oid "$(git cat-file -p :a.txt)")"
  [ "$oid_root" = "$oid_root_after_migration" ]
)
end_test

begin_test "migrate import (--object-map)"
(
  set -e

  setup_multiple_local_branches

  output_dir=$(mktemp -d)

  git log --all --pretty='format:%H' > "${output_dir}/old_sha.txt"
  git lfs migrate import --everything --object-map "${output_dir}/object-map.txt"
  git log --all --pretty='format:%H' > "${output_dir}/new_sha.txt"
  paste -d',' "${output_dir}/old_sha.txt" "${output_dir}/new_sha.txt" > "${output_dir}/expected-map.txt"

  diff -u <(sort "${output_dir}/expected-map.txt") <(sort "${output_dir}/object-map.txt")
)
end_test

begin_test "migrate import (--include with space)"
(
  set -e

  setup_local_branch_with_space

  oid="$(calc_oid "$(git cat-file -p :"a file.txt")")"

  git lfs migrate import --include "a file.txt"

  assert_pointer "refs/heads/master" "a file.txt" "$oid" 50
  cat .gitattributes
  if [ 1 -ne "$(grep -c "a\[\[:space:\]\]file.txt" .gitattributes)" ]; then
    echo >&2 "fatal: expected \"a[[:space:]]file.txt\" to appear in .gitattributes"
    echo >&2 "fatal: got"
    sed -e 's/^/  /g' < .gitattributes >&2
    exit 1
  fi
)
end_test

begin_test "migrate import (handle symbolic link)"
(
  set -e

  setup_local_branch_with_symlink

  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  link_oid="$(calc_oid "$(git cat-file -p :link.txt)")"

  git lfs migrate import --include="*.txt"

  assert_pointer "refs/heads/master" "a.txt" "$txt_oid" "120"

  assert_local_object "$txt_oid" "120"
  # "link.txt" is a symbolic link so it should be not in LFS
  refute_local_object "$link_oid" "5"
)
end_test

begin_test "migrate import (commit --allow-empty)"
(
  set -e

  reponame="migrate---allow-empty"
  git init "$reponame"
  cd "$reponame"

  git commit --allow-empty -m "initial commit"

  original_head="$(git rev-parse HEAD)"
  git lfs migrate import --everything
  migrated_head="$(git rev-parse HEAD)"

  assert_ref_unmoved "HEAD" "$original_head" "$migrated_head"
)
end_test

begin_test "migrate import (multiple remotes)"
(
  set -e

  setup_multiple_remotes

  original_master="$(git rev-parse master)"

  git lfs migrate import

  migrated_master="$(git rev-parse master)"

  assert_ref_unmoved "master" "$original_master" "$migrated_master"
)
end_test
