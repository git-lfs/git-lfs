#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "update"
(
  set -e

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\\n\"; exit 2; }
git lfs pre-push \"\$@\""

  mkdir without-pre-push
  cd without-pre-push
  git init

  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # run it again
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 1
  echo "#!/bin/sh
git lfs push --stdin \$*" > .git/hooks/pre-push
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 2
  echo "#!/bin/sh
git lfs push --stdin \"\$@\"" > .git/hooks/pre-push
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 3
  echo "#!/bin/sh
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 4
  echo "#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }
git lfs pre-push \"$@\""
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 5
  echo "#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }
git lfs pre-push \"$@\""
  [ "Updated pre-push hook." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # don't replace unexpected hook
  echo "test" > .git/hooks/pre-push
  expected="Hook already exists: pre-push

test

Run \`git lfs update --force\` to overwrite this hook."

  [ "$expected" = "$(git lfs update 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]

  # force replace unexpected hook
  [ "Updated pre-push hook." = "$(git lfs update --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  echo "test with bare repository"
  cd ..
  git clone --mirror without-pre-push bare
  cd bare
  git lfs env
  git lfs update
  ls -al hooks
  [ "$pre_push_hook" = "$(cat hooks/pre-push)" ]
)
end_test

begin_test "update lfs.{url}.access"
(
  set -e

  mkdir update-access
  cd update-access
  git init
  git config lfs.http://example.com.access private
  git config lfs.https://example.com.access private
  git config lfs.https://example2.com.access basic
  git config lfs.https://example3.com.access other

  [ "private" = "$(git config lfs.http://example.com.access)" ]
  [ "private" = "$(git config lfs.https://example.com.access)" ]
  [ "basic" = "$(git config lfs.https://example2.com.access)" ]
  [ "other" = "$(git config lfs.https://example3.com.access)" ]

  expected="Updated pre-push hook.
Updated http://example.com access from private to basic.
Updated https://example.com access from private to basic.
Removed invalid https://example3.com access of other."
)
end_test

begin_test "update: outside git repository"
(
  set +e
  git lfs update 2>&1 > check.log
  res=$?
  overwrite="$(grep "overwrite" check.log)"

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  [ -z "$overwrite" ]
  grep "Not in a git repository" check.log
)
end_test
