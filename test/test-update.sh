#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "update"
(
  set -e

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }
git lfs pre-push \"\$@\""

  mkdir without-pre-push
  cd without-pre-push
  git init

  [ "Updated pre-push hook" = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # run it again
  [ "Updated pre-push hook" = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 1
  echo "#!/bin/sh
git lfs push --stdin \$*" > .git/hooks/pre-push
  [ "Updated pre-push hook" = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 2
  echo "#!/bin/sh
git lfs push --stdin \"\$@\"" > .git/hooks/pre-push
  [ "Updated pre-push hook" = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 3
  echo "#!/bin/sh
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated pre-push hook" = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # don't replace unexpected hook
  echo "test" > .git/hooks/pre-push
  expected="Hook already exists: pre-push

test

Run \`git lfs update --force\` to overwrite this hook."

  [ "$expected" = "$(git lfs update 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]

  # force replace unexpected hook
  [ "Updated pre-push hook" = "$(git lfs update --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
)
end_test
