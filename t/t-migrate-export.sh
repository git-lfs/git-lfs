#!/usr/bin/env bash

. "$(dirname "$0")/fixtures/migrate.sh"
. "$(dirname "$0")/testlib.sh"

begin_test "migrate export (default branch)"
(
  set -e

  setup_multiple_local_branches_tracked

  # Add b.md, a pointer existing only on main
  base64 < /dev/urandom | head -c 160 > b.md
  git add b.md
  git commit -m "add b.md"

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"
  b_md_oid="$(calc_oid "$(cat b.md)")"

  git checkout my-feature
  md_feature_oid="$(calc_oid "$(cat a.md)")"
  git checkout main

  assert_pointer "refs/heads/main" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
  assert_pointer "refs/heads/main" "b.md" "$b_md_oid" "160"
  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"

  git lfs migrate export --include="*.md, *.txt"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"
  refute_pointer "refs/heads/main" "b.md"
  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"

  # b.md should be pruned as no pointer exists to reference it
  refute_local_object "$b_md_oid" "160"

  # Other objects should not be pruned as they're still referenced in `feature`
  # by pointers
  assert_local_object "$md_oid" "140"
  assert_local_object "$txt_oid" "120"
  assert_local_object "$md_feature_oid" "30"

  main="$(git rev-parse refs/heads/main)"
  feature="$(git rev-parse refs/heads/my-feature)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$main_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff"

  [ ! $(echo "$feature_attrs" | grep -q "*.md !text !filter !merge !diff") ]
  [ ! $(echo "$feature_attrs" | grep -q "*.txt !text !filter !merge !diff") ]
)
end_test

begin_test "migrate export (with remote)"
(
  set -e

  setup_single_remote_branch_tracked

  git push origin main

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  assert_pointer "refs/heads/main" "a.md" "$md_oid" "50"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "30"

  assert_pointer "refs/remotes/origin/main" "a.md" "$md_oid" "50"
  assert_pointer "refs/remotes/origin/main" "a.txt" "$txt_oid" "30"

  # Flush the cache to ensure all objects have to be downloaded
  rm -rf .git/lfs/objects

  git lfs migrate export --everything --include="*.md, *.txt"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"

  # All pointers have been exported, so all objects should be pruned
  refute_local_object "$md_oid" "50"
  refute_local_object "$txt_oid" "30"

  main="$(git rev-parse refs/heads/main)"
  main_attrs="$(git cat-file -p "$main:.gitattributes")"

  echo "$main_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff"
)
end_test

begin_test "migrate export (include/exclude args)"
(
  set -e

  setup_single_local_branch_tracked

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "140"

  git lfs migrate export --include="*" --exclude="a.md"

  refute_pointer "refs/heads/main" "a.txt"
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "140"

  refute_local_object "$txt_oid" "120"
  assert_local_object "$md_oid" "140"

  main="$(git rev-parse refs/heads/main)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"

  echo "$main_attrs" | grep -q "* !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "a.md filter=lfs diff=lfs merge=lfs"
)
end_test

begin_test "migrate export (bare repository)"
(
  set -e

  setup_single_remote_branch_tracked
  git push origin main

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  make_bare

  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "30"
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "50"

  git lfs migrate export --everything --include="*"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"

  # All pointers have been exported, so all objects should be pruned
  refute_local_object "$md_oid" "50"
  refute_local_object "$txt_oid" "30"
)
end_test

begin_test "migrate export (given branch)"
(
  set -e

  setup_multiple_local_branches_tracked

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  git checkout my-feature
  md_feature_oid="$(calc_oid "$(cat a.md)")"
  git checkout main

  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "30"
  assert_pointer "refs/heads/my-feature" "a.txt" "$txt_oid" "120"
  assert_pointer "refs/heads/main" "a.md" "$md_oid" "140"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "120"

  git lfs migrate export --include="*.md,*.txt" my-feature

  refute_pointer "refs/heads/my-feature" "a.md"
  refute_pointer "refs/heads/my-feature" "a.txt"
  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"

  # No pointers left, so all objects should be pruned
  refute_local_object "$md_feature_oid" "30"
  refute_local_object "$txt_oid" "120"
  refute_local_object "$md_oid" "140"

  main="$(git rev-parse refs/heads/main)"
  feature="$(git rev-parse refs/heads/my-feature)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  echo "$main_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff"
  echo "$feature_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$feature_attrs" | grep -q "*.txt !text !filter !merge !diff"
)
end_test

begin_test "migrate export (no filter)"
(
  set -e

  setup_multiple_local_branches_tracked

  git lfs migrate export --yes 2>&1 | tee migrate.log
  if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo >&2 "fatal: expected git lfs migrate export to fail, didn't"
    exit 1
  fi

  grep "fatal: one or more files must be specified with --include" migrate.log
)
end_test

begin_test "migrate export (exclude remote refs)"
(
  set -e

  setup_single_remote_branch_tracked

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  git checkout refs/remotes/origin/main
  md_remote_oid="$(calc_oid "$(cat a.md)")"
  txt_remote_oid="$(calc_oid "$(cat a.txt)")"
  git checkout main

  assert_pointer "refs/heads/main" "a.md" "$md_oid" "50"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "30"

  assert_pointer "refs/remotes/origin/main" "a.md" "$md_remote_oid" "140"
  assert_pointer "refs/remotes/origin/main" "a.txt" "$txt_remote_oid" "120"

  git lfs migrate export --include="*.md,*.txt"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"

  refute_local_object "$md_oid" "50"
  refute_local_object "$txt_oid" "30"

  assert_pointer "refs/remotes/origin/main" "a.md" "$md_remote_oid" "140"
  assert_pointer "refs/remotes/origin/main" "a.txt" "$txt_remote_oid" "120"

  # Since these two objects exist on the remote, they should be removed with
  # our prune operation
  refute_local_object "$md_remote_oid" "140"
  refute_local_object "$txt_remote_oid" "120"

  main="$(git rev-parse refs/heads/main)"
  remote="$(git rev-parse refs/remotes/origin/main)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  remote_attrs="$(git cat-file -p "$remote:.gitattributes")"

  echo "$main_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff"

  [ ! $(echo "$remote_attrs" | grep -q "*.md !text !filter !merge !diff") ]
  [ ! $(echo "$remote_attrs" | grep -q "*.txt !text !filter !merge !diff") ]
)
end_test

begin_test "migrate export (--skip-fetch)"
(
  set -e

  setup_single_remote_branch_tracked

  md_main_oid="$(calc_oid "$(cat a.md)")"
  txt_main_oid="$(calc_oid "$(cat a.txt)")"

  git checkout refs/remotes/origin/main
  md_remote_oid="$(calc_oid "$(cat a.md)")"
  txt_remote_oid="$(calc_oid "$(cat a.txt)")"
  git checkout main

  git tag pseudo-remote "$(git rev-parse refs/remotes/origin/main)"
  # Remove the refs/remotes/origin/main ref, and instruct 'git lfs migrate' to
  # not fetch it.
  git update-ref -d refs/remotes/origin/main

  assert_pointer "refs/heads/main" "a.md" "$md_main_oid" "50"
  assert_pointer "pseudo-remote" "a.md" "$md_remote_oid" "140"
  assert_pointer "refs/heads/main" "a.txt" "$txt_main_oid" "30"
  assert_pointer "pseudo-remote" "a.txt" "$txt_remote_oid" "120"

  git lfs migrate export --skip-fetch --include="*.md,*.txt"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "pseudo-remote" "a.md"
  refute_pointer "refs/heads/main" "a.txt"
  refute_pointer "pseudo-remote" "a.txt"

  refute_local_object "$md_main_oid" "50"
  refute_local_object "$md_remote_oid" "140"
  refute_local_object "$txt_main_oid" "30"
  refute_local_object "$txt_remote_oid" "120"

  main="$(git rev-parse refs/heads/main)"
  remote="$(git rev-parse pseudo-remote)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  remote_attrs="$(git cat-file -p "$remote:.gitattributes")"

  echo "$main_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff"
  echo "$remote_attrs" | grep -q "*.md !text !filter !merge !diff"
  echo "$remote_attrs" | grep -q "*.txt !text !filter !merge !diff"
)
end_test

begin_test "migrate export (include/exclude ref)"
(
  set -e

  setup_multiple_remote_branches_gitattrs

  md_main_oid="$(calc_oid "$(cat a.md)")"
  txt_main_oid="$(calc_oid "$(cat a.txt)")"

  git checkout refs/remotes/origin/main
  md_remote_oid="$(calc_oid "$(cat a.md)")"
  txt_remote_oid="$(calc_oid "$(cat a.txt)")"

  git checkout my-feature
  md_feature_oid="$(calc_oid "$(cat a.md)")"
  txt_feature_oid="$(calc_oid "$(cat a.txt)")"

  git checkout main

  git lfs migrate export \
    --include="*.txt" \
    --include-ref=refs/heads/my-feature \
    --exclude-ref=refs/heads/main

  assert_pointer "refs/heads/main" "a.md" "$md_main_oid" "21"
  assert_pointer "refs/heads/main" "a.txt" "$txt_main_oid" "20"

  assert_pointer "refs/remotes/origin/main" "a.md" "$md_remote_oid" "11"
  assert_pointer "refs/remotes/origin/main" "a.txt" "$txt_remote_oid" "10"

  assert_pointer "refs/heads/my-feature" "a.md" "$md_feature_oid" "31"
  refute_pointer "refs/heads/my-feature" "a.txt"

  # Master objects should not be pruned as they exist in unpushed commits
  assert_local_object "$md_main_oid" "21"
  assert_local_object "$txt_main_oid" "20"

  # Remote main objects should be pruned as they exist in the remote
  refute_local_object "$md_remote_oid" "11"
  refute_local_object "$txt_remote_oid" "10"

  # txt_feature_oid should be pruned as it's no longer a pointer, but
  # md_feature_oid should remain as it's still a pointer in unpushed commits
  assert_local_object "$md_feature_oid" "31"
  refute_local_object "$txt_feature_oid" "30"

  main="$(git rev-parse refs/heads/main)"
  feature="$(git rev-parse refs/heads/my-feature)"
  remote="$(git rev-parse refs/remotes/origin/main)"

  main_attrs="$(git cat-file -p "$main:.gitattributes")"
  remote_attrs="$(git cat-file -p "$remote:.gitattributes")"
  feature_attrs="$(git cat-file -p "$feature:.gitattributes")"

  [ ! $(echo "$main_attrs" | grep -q "*.txt !text !filter !merge !diff") ]
  [ ! $(echo "$remote_attrs" | grep -q "*.txt !text !filter !merge !diff") ]
  echo "$feature_attrs" | grep -q "*.txt !text !filter !merge !diff"
)
end_test

begin_test "migrate export (--object-map)"
(
  set -e

  setup_multiple_local_branches_tracked

  output_dir=$(mktemp -d)

  git log --all --pretty='format:%H' > "${output_dir}/old_sha.txt"
  git lfs migrate export --everything --include="*" --object-map "${output_dir}/object-map.txt"
  git log --all --pretty='format:%H' > "${output_dir}/new_sha.txt"
  paste -d',' "${output_dir}/old_sha.txt" "${output_dir}/new_sha.txt" > "${output_dir}/expected-map.txt"

  diff -u <(sort "${output_dir}/expected-map.txt") <(sort "${output_dir}/object-map.txt")
)
end_test

begin_test "migrate export (--verbose)"
(
  set -e

  setup_multiple_local_branches_tracked

  git lfs migrate export --everything --include="*" --verbose 2>&1 | grep -q "migrate: commit "
)
end_test

begin_test "migrate export (--remote)"
(
  set -e

  setup_single_remote_branch_tracked

  git push origin main

  md_oid="$(calc_oid "$(cat a.md)")"
  txt_oid="$(calc_oid "$(cat a.txt)")"

  assert_pointer "refs/heads/main" "a.md" "$md_oid" "50"
  assert_pointer "refs/heads/main" "a.txt" "$txt_oid" "30"

  # Flush the cache to ensure all objects have to be downloaded
  rm -rf .git/lfs/objects

  # Setup a new remote and invalidate the default
  remote_url="$(git config --get remote.origin.url)"
  git remote add zeta "$remote_url"
  git remote set-url origin ""

  git lfs migrate export --everything --remote="zeta" --include="*.md, *.txt"

  refute_pointer "refs/heads/main" "a.md"
  refute_pointer "refs/heads/main" "a.txt"

  refute_local_object "$md_oid" "50"
  refute_local_object "$txt_oid" "30"
)
end_test

begin_test "migrate export (invalid --remote)"
(
  set -e

  setup_single_remote_branch_tracked

  git lfs migrate export --include="*" --remote="zz" --yes 2>&1 \
    | tee migrate.log
  if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo >&2 "fatal: expected git lfs migrate export to fail, didn't"
    exit 1
  fi

  grep "fatal: invalid remote zz provided" migrate.log
)
end_test

begin_test "migrate export (invalid pointer)"
(
  set -e

  git init repo1
  git init repo2

  cd repo1
  echo "git-lfs" > problematic_file
  git add .
  git commit -m "create repo"

  git lfs migrate export --include="*" --everything --yes

  cd ../repo2
  echo "not git-lfs" > problematic_file
  git add .
  git commit -m "create repo"

  git lfs migrate export --include="*" --everything --yes
)
end_test
