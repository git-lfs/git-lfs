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

  # replace blank hook
  rm .git/hooks/pre-push
  touch .git/hooks/pre-push
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

To resolve this, either:
  1: run \`git lfs update --manual\` for instructions on how to merge hooks.
  2: run \`git lfs update --force\` to overwrite your hook."

  [ "$expected" = "$(git lfs update 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]

  # Make sure returns non-zero
  set +e
  git lfs update
  if [ $? -eq 0 ]
  then
    exit 1
  fi
  set -e

  # test manual steps
  expected="Add the following to .git/hooks/pre-push :

#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\n\"; exit 2; }
git lfs pre-push \"\$@\""

  [ "$expected" = "$(git lfs update --manual 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]

  # force replace unexpected hook
  [ "Updated pre-push hook." = "$(git lfs update --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  has_test_dir || exit 0

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
  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  set +e
  git lfs update 2>&1 > check.log
  res=$?
  set -e

  if [ "$res" = "0" ]; then
    if [ -z "$GIT_LFS_TEST_DIR" ]; then
      echo "Passes because $GIT_LFS_TEST_DIR is unset."
      exit 0
    fi
  fi

  [ "$res" = "128" ]

  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  cat check.log
  grep "Not in a git repository" check.log
)
end_test
