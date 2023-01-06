#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "update"
(
  set -e

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'pre-push' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\\n\"; exit 2; }
git lfs pre-push \"\$@\""

  post_checkout_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-checkout' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\\n\"; exit 2; }
git lfs post-checkout \"\$@\""

  post_commit_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-commit' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\\n\"; exit 2; }
git lfs post-commit \"\$@\""

  post_merge_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-merge' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\\n\"; exit 2; }
git lfs post-merge \"\$@\""

  mkdir without-pre-push
  cd without-pre-push
  git init

  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  # run it again
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  # replace old hook 1
  echo "#!/bin/sh
git lfs push --stdin \$*" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 2
  echo "#!/bin/sh
git lfs push --stdin \"\$@\"" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 3
  echo "#!/bin/sh
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace blank hook
  rm .git/hooks/pre-push
  touch .git/hooks/pre-push
  touch .git/hooks/post-checkout
  touch .git/hooks/post-merge
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  # replace old hook 4
  echo "#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook 5
  echo "#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]

  # replace old hook 6
  echo "#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting '.git/hooks/pre-push'.\\n\"; exit 2; }
git lfs pre-push \"\$@\"" > .git/hooks/pre-push
  [ "Updated Git hooks." = "$(git lfs update)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # don't replace unexpected hook
  echo "test" > .git/hooks/pre-push
  echo "test" > .git/hooks/post-checkout
  echo "test" > .git/hooks/post-commit
  echo "test" > .git/hooks/post-merge
  expected="Hook already exists: pre-push

	test

To resolve this, either:
  1: run \`git lfs update --manual\` for instructions on how to merge hooks.
  2: run \`git lfs update --force\` to overwrite your hook."

  [ "$expected" = "$(git lfs update 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]
  [ "test" = "$(cat .git/hooks/post-checkout)" ]
  [ "test" = "$(cat .git/hooks/post-commit)" ]
  [ "test" = "$(cat .git/hooks/post-merge)" ]

  # Make sure returns non-zero
  set +e
  git lfs update
  if [ $? -eq 0 ]
  then
    exit 1
  fi
  set -e

  # test manual steps
  expected="Add the following to '.git/hooks/pre-push':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'pre-push' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs pre-push \"\$@\"

Add the following to '.git/hooks/post-checkout':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-checkout' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-checkout \"\$@\"

Add the following to '.git/hooks/post-commit':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-commit' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-commit \"\$@\"

Add the following to '.git/hooks/post-merge':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-merge' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-merge \"\$@\""

  [ "$expected" = "$(git lfs update --manual 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]
  [ "test" = "$(cat .git/hooks/post-checkout)" ]
  [ "test" = "$(cat .git/hooks/post-commit)" ]
  [ "test" = "$(cat .git/hooks/post-merge)" ]

  # force replace unexpected hook
  [ "Updated Git hooks." = "$(git lfs update --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  # test manual steps with core.hookspath
  gitversion=$(git version | cut -d" " -f3)
  set +e
  compare_version "$gitversion" 2.9.0
  result=$?
  set -e
  if [ "$result" -ne "$VERSION_LOWER" ]
  then
    mkdir hooks
    rm -fr .git/hooks
    git config core.hookspath hooks
    echo "test" > hooks/pre-push
    echo "test" > hooks/post-checkout
    echo "test" > hooks/post-commit
    echo "test" > hooks/post-merge

    expected="Add the following to 'hooks/pre-push':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'pre-push' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs pre-push \"\$@\"

Add the following to 'hooks/post-checkout':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-checkout' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-checkout \"\$@\"

Add the following to 'hooks/post-commit':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-commit' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-commit \"\$@\"

Add the following to 'hooks/post-merge':

	#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting the 'post-merge' file in the hooks directory (set by 'core.hookspath'; usually '.git/hooks').\n\"; exit 2; }
	git lfs post-merge \"\$@\""
    [ "$expected" = "$(git lfs update --manual 2>&1)" ]
    [ "test" = "$(cat hooks/pre-push)" ]
    [ "test" = "$(cat hooks/post-checkout)" ]
    [ "test" = "$(cat hooks/post-commit)" ]
    [ "test" = "$(cat hooks/post-merge)" ]

    # force replace unexpected hook
    [ "Updated Git hooks." = "$(git lfs update --force)" ]
    [ "$pre_push_hook" = "$(cat hooks/pre-push)" ]
    [ "$post_checkout_hook" = "$(cat hooks/post-checkout)" ]
    [ "$post_commit_hook" = "$(cat hooks/post-commit)" ]
    [ "$post_merge_hook" = "$(cat hooks/post-merge)" ]

    test -d .git/hooks && exit 1
  fi

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

begin_test "update with leading spaces"
(
  set -e

  reponame="update-leading-spaces"
  git init "$reponame"
  cd "$reponame"

  [ "Updated Git hooks." = "$(git lfs update)" ]

  # $pre_push_hook contains leading TAB '\t' characters
  pre_push_hook="#!/bin/sh
	command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting '.git/hooks/pre-push'.\\n\"; exit 2; }
	git lfs pre-push \"\$@\""

  echo -n "$pre_push_hook" > .git/hooks/pre-push

  [ "Updated Git hooks." = "$(git lfs update)" ]
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

  expected="Updated Git hooks.
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
  grep "Not in a Git repository" check.log
)
end_test
