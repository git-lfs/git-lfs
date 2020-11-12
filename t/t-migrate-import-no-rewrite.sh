#!/usr/bin/env bash

. "$(dirname "$0")/fixtures/migrate.sh"
. "$(dirname "$0")/testlib.sh"

begin_test "migrate import --no-rewrite (default branch)"
(
  set -e

  setup_local_branch_with_gitattrs

  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  prev_commit_oid="$(git rev-parse HEAD)"

  git lfs migrate import --no-rewrite --yes *.txt

  # Ensure our desired files were imported into git-lfs
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
  assert_local_object "$txt_oid" "120"

  # Ensure the git history remained the same
  new_commit_oid="$(git rev-parse HEAD~1)"
  if [ "$prev_commit_oid" != "$new_commit_oid" ]; then
    exit 1
  fi

  # Ensure a new commit was made
  new_head_oid="$(git rev-parse HEAD)"
  if [ "$prev_commit_oid" == "$new_oid" ]; then
    exit 1
  fi

  # Ensure a new commit message was generated based on the list of imported files
  commit_msg="$(git log -1 --pretty=format:%s)"
  echo "$commit_msg" | grep -q "a.txt: convert to Git LFS"

  # Ensure we write a valid commit object.
  git fsck
)
end_test

begin_test "migrate import --no-rewrite (bare repository)"
(
  set -e

  setup_single_remote_branch_with_gitattrs

  prev_commit_oid="$(git rev-parse HEAD)"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"

  git lfs migrate import --no-rewrite --yes a.txt a.md

  # Ensure our desired files were imported
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "30"
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "50"

  # Ensure the git history remained the same
  new_commit_oid="$(git rev-parse HEAD~1)"
  if [ "$prev_commit_oid" != "$new_commit_oid" ]; then
    exit 1
  fi

  # Ensure a new commit was made
  new_head_oid="$(git rev-parse HEAD)"
  if [ "$prev_commit_oid" == "$new_oid" ]; then
    exit 1
  fi
)
end_test

begin_test "migrate import --no-rewrite (multiple branches)"
(
  set -e

  setup_multiple_local_branches_with_gitattrs

  prev_commit_oid="$(git rev-parse HEAD)"

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"
  md_feature_oid="$(calc_oid "$(git cat-file -p my-feature:a.md)")"

  git lfs migrate import --no-rewrite --yes *.txt *.md

  # Ensure our desired files were imported
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"

  assert_local_object "$md_oid" "140"
  assert_local_object "$txt_oid" "120"

  # Ensure our other branch was unmodified
  refute_local_object "$md_feature_oid" "30"

  # Ensure the git history remained the same
  new_commit_oid="$(git rev-parse HEAD~1)"
  if [ "$prev_commit_oid" != "$new_commit_oid" ]; then
    exit 1
  fi

  # Ensure a new commit was made
  new_head_oid="$(git rev-parse HEAD)"
  if [ "$prev_commit_oid" == "$new_oid" ]; then
    exit 1
  fi
)
end_test

begin_test "migrate import --no-rewrite (no .gitattributes)"
(
  set -e

  setup_multiple_local_branches

  # Ensure command fails if no .gitattributes files are present
  git lfs migrate import --no-rewrite --yes *.txt *.md 2>&1 | tee migrate.log
  if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo >&2 "fatal: expected git lfs migrate import --no-rewrite to fail, didn't"
    exit 1
  fi

  grep "no Git LFS filters found in .gitattributes" migrate.log
)
end_test

begin_test "migrate import --no-rewrite (nested .gitattributes)"
(
  set -e

  setup_local_branch_with_nested_gitattrs

  # Ensure a .md filter does not exist in the top-level .gitattributes
  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  [ !"$(echo "$main_attrs" | grep -q ".md")" ]

  # Ensure a .md filter exists in the nested .gitattributes
  nested_attrs="$(git cat-file -p "$main:b/.gitattributes")"
  echo "$nested_attrs" | grep -q "*.md filter=lfs diff=lfs merge=lfs"

  md_oid="$(calc_oid "$(git cat-file -p :a.md)")"
  nested_md_oid="$(calc_oid "$(git cat-file -p :b/a.md)")"
  txt_oid="$(calc_oid "$(git cat-file -p :a.txt)")"

  git lfs migrate import --no-rewrite --yes a.txt b/a.md

  # Ensure a.txt and subtree/a.md were imported, even though *.md only exists in the
  # nested subtree/.gitattributes file
  assert_pointer "refs/heads/main" "b/a.md" "$nested_md_oid" "140"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"

  assert_local_object "$nested_md_oid" 140
  assert_local_object "$txt_oid" 120
  refute_local_object "$md_oid" 140

  # Failure should occur when trying to import a.md as no entry exists in
  # top-level .gitattributes file
  git lfs migrate import --no-rewrite --yes a.md 2>&1 | tee migrate.log
  if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo >&2 "fatal: expected git lfs migrate import --no-rewrite to fail, didn't"
    exit 1
  fi

  grep "a.md did not match any Git LFS filters in .gitattributes" migrate.log
)
end_test

begin_test "migrate import --no-rewrite (with commit message)"
(
  set -e

  setup_local_branch_with_gitattrs

  prev_commit_oid="$(git rev-parse HEAD)"
  expected_commit_msg="run git-lfs migrate import --no-rewrite"

  git lfs migrate import --message "$expected_commit_msg" --no-rewrite --yes *.txt

  # Ensure the git history remained the same
  new_commit_oid="$(git rev-parse HEAD~1)"
  if [ "$prev_commit_oid" != "$new_commit_oid" ]; then
    exit 1
  fi

  # Ensure a new commit was made
  new_head_oid="$(git rev-parse HEAD)"
  if [ "$prev_commit_oid" == "$new_oid" ]; then
    exit 1
  fi

  # Ensure the provided commit message was used
  commit_msg="$(git log -1 --pretty=format:%s)"
  if [ "$commit_msg" != "$expected_commit_msg" ]; then
    exit 1
  fi
)
end_test

begin_test "migrate import --no-rewrite (with empty commit message)"
(
  set -e

  setup_local_branch_with_gitattrs

  prev_commit_oid="$(git rev-parse HEAD)"

  git lfs migrate import -m "" --no-rewrite --yes *.txt

  # Ensure the git history remained the same
  new_commit_oid="$(git rev-parse HEAD~1)"
  if [ "$prev_commit_oid" != "$new_commit_oid" ]; then
    exit 1
  fi

  # Ensure a new commit was made
  new_head_oid="$(git rev-parse HEAD)"
  if [ "$prev_commit_oid" == "$new_oid" ]; then
    exit 1
  fi

  # Ensure the provided commit message was used
  commit_msg="$(git log -1 --pretty=format:%s)"
  if [ "$commit_msg" != "" ]; then
    exit 1
  fi
)
end_test

begin_test "migrate import --no-rewrite (strict .gitattributes)"
(
  set -e

  reponame="$(basename "$0" ".sh")-strict-match"
  clone_repo "$reponame" repo-strict-match

  mkdir -p major-oak/mainst/.yarn-offline-mirror/
  mkdir -p major-oak/major-oak/frontend/.yarn-offline-mirror/
  foo_contents="foo"
  foo_oid=$(calc_oid "$foo_contents")
  bar_contents="bar"
  bar_oid=$(calc_oid "$bar_contents")
  printf "$foo_contents" > major-oak/mainst/.yarn-offline-mirror/typescript-3.4.3.tgz
  printf "$bar_contents" > major-oak/major-oak/frontend/.yarn-offline-mirror/typescript-2.9.2.tgz
  git add .
  git commit -m 'Initial import'

  cat >.gitattributes <<EOF
major-oak/mainst/.yarn-offline-mirror/typescript-3.4.3.tgz filter=lfs diff=lfs merge=lfs -text
major-oak/major-oak/frontend/.yarn-offline-mirror/typescript-2.9.2.tgz filter=lfs diff=lfs merge=lfs -text
EOF

  git add .
  git commit -m '.gitattributes'

  git lfs migrate import --yes --no-rewrite \
    major-oak/mainst/.yarn-offline-mirror/typescript-3.4.3.tgz \
    major-oak/major-oak/frontend/.yarn-offline-mirror/typescript-2.9.2.tgz

  assert_pointer "refs/heads/main" "major-oak/mainst/.yarn-offline-mirror/typescript-3.4.3.tgz" "$foo_oid" "3"
  assert_pointer "refs/heads/main" "major-oak/major-oak/frontend/.yarn-offline-mirror/typescript-2.9.2.tgz" "$bar_oid" "3"

  assert_local_object "$foo_oid" "3"
  assert_local_object "$bar_oid" "3"
)
end_test
