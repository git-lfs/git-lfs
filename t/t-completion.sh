#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "completion: bash script"
(
  set -e

  git lfs completion bash | cmp - "$COMPLETIONSDIR/git-lfs-completion.bash"
)
end_test

begin_test "completion: fish script"
(
  set -e

  git lfs completion fish | cmp - "$COMPLETIONSDIR/git-lfs-completion.fish"
)
end_test

begin_test "completion: zsh script"
(
  set -e

  git lfs completion zsh | cmp - "$COMPLETIONSDIR/git-lfs-completion.zsh"
)
end_test

begin_test "completion: missing shell argument"
(
  set -e

  git lfs completion 2>&1 | tee completion.log
  grep "accepts 1 arg" completion.log
)
end_test

begin_test "completion: invalid shell argument"
(
  set -e

  git lfs completion ksh 2>&1 | tee completion.log
  grep "invalid argument" completion.log
)
end_test
