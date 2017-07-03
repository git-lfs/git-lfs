#!/usr/bin/env bash

. "test/testlib.sh"

# These tests rely on behavior found in Git versions less than 2.9.0 to perform
# themselves, specifically:
#   - lack of core.hooksPath support
ensure_git_version_isnt $VERSION_HIGHER "2.9.0"

begin_test "install with unsupported core.hooksPath"
(
  set -e

  repo_name="unsupported-custom-hooks-path"
  git init "$repo_name"
  cd "$repo_name"

  hooks_dir="custom_hooks_dir"
  mkdir -p "$hooks_dir"

  git config --local core.hooksPath "$hooks_dir"

  git lfs install 2>&1 | tee install.log
  grep "Updated git hooks" install.log

  [ ! -e "$hooks_dir/pre-push" ]
  [ -e ".git/hooks/pre-push" ]
)
end_test
