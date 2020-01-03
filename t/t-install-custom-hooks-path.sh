#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# These tests rely on behavior found in 2.9.0 to perform themselves,
# specifically:
#   - core.hooksPath support
ensure_git_version_isnt $VERSION_LOWER "2.9.0"

assert_hooks() {
  local hooks_dir="$1"
  [ -e "$hooks_dir/pre-push" ]
  [ ! -e ".git/pre-push" ]
  [ -e "$hooks_dir/post-checkout" ]
  [ ! -e ".git/post-checkout" ]
  [ -e "$hooks_dir/post-commit" ]
  [ ! -e ".git/post-commit" ]
  [ -e "$hooks_dir/post-merge" ]
  [ ! -e ".git/post-merge" ]
}

refute_hooks() {
  local hooks_dir="$1"
  [ ! -e "$hooks_dir/pre-push" ]
  [ ! -e "$hooks_dir/post-checkout" ]
  [ ! -e "$hooks_dir/post-commit" ]
  [ ! -e "$hooks_dir/post-merge" ]
}

begin_test "install with supported core.hooksPath"
(
  set -e

  repo_name="supported-custom-hooks-path"
  git init "$repo_name"
  cd "$repo_name"

  hooks_dir="custom_hooks_dir"
  mkdir -p "$hooks_dir"

  git config --local core.hooksPath "$hooks_dir"

  git lfs install 2>&1 | tee install.log
  grep "Updated git hooks" install.log

  assert_hooks "$hooks_dir"
)
end_test

begin_test "install with supported core.hooksPath in subdirectory"
(
  set -e

  repo_name="supported-custom-hooks-path-subdir"
  git init "$repo_name"
  cd "$repo_name"

  hooks_dir="custom_hooks_dir"

  mkdir subdir

  git config --local core.hooksPath "$hooks_dir"

  (cd subdir && git lfs install 2>&1 | tee install.log)
  grep "Updated git hooks" subdir/install.log

  assert_hooks "$hooks_dir"
  refute_hooks "subdir/$hooks_dir"
)
end_test

begin_test "install with supported expandable core.hooksPath"
(
  set -e

  repo_name="supported-custom-hooks-expandable-path"
  git init "$repo_name"
  cd "$repo_name"

  hooks_dir="~/custom_hooks_dir"
  mkdir -p "$hooks_dir"

  git config --local core.hooksPath "$hooks_dir"

  git lfs install 2>&1 | tee install.log
  grep "Updated git hooks" install.log

  assert_hooks "$HOME/custom_hooks_dir"
)
end_test
