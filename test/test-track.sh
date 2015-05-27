#!/bin/sh

. "test/testlib.sh"

begin_test "track"
(
  set -e

  # no need to setup a remote repo, since this test doesn't need to push or pull

  mkdir track
  cd track
  git init

  # track *.jpg once
  git lfs track "*.jpg" | grep "Tracking \*.jpg"
  numjpg=$(grep "\*.jpg" .gitattributes | wc -l)
  if [ "$(printf "%d" "$numjpg")" != "1" ]; then
    echo "wrong number of jpgs"
    cat .gitattributes
    exit 1
  fi

  # track *.jpg again
  git lfs track "*.jpg" | grep "*.jpg already supported"
  numjpg=$(grep "\*.jpg" .gitattributes | wc -l)
  if [ "$(printf "%d" "$numjpg")" != "1" ]; then
    echo "wrong number of jpgs"
    cat .gitattributes
    exit 1
  fi

  mkdir -p a/b

  echo "*.mov filter=lfs -crlf" > .git/info/attributes
  echo "*.gif filter=lfs -crlf" > a/.gitattributes
  echo "*.png filter=lfs -crlf" > a/b/.gitattributes

  out=$(git lfs track)
  echo "$out" | grep "Listing tracked paths"
  echo "$out" | grep "*.mov (.git/info/attributes)"
  echo "$out" | grep "*.jpg (.gitattributes)"
  echo "$out" | grep "*.gif (a/.gitattributes)"
  echo "$out" | grep "*.png (a/b/.gitattributes)"
)
end_test

begin_test "track directory"
(
  set -e
  mkdir dir
  cd dir
  git init

  git lfs track "foo bar/*"

  mkdir "foo bar"
  echo "a" > "foo bar/a"
  echo "b" > "foo bar/b"
  git add foo\ bar
  git commit -am "add foo bar"

  assert_pointer "master" "foo bar/a" "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7" 2
  assert_pointer "master" "foo bar/b" "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f" 2
)

begin_test "track without trailing linebreak"
(
  set -e

  mkdir no-linebreak
  cd no-linebreak
  git init
  printf "*.mov filter=lfs -crlf" > .gitattributes

  git lfs track "*.gif"

  expected="*.mov filter=lfs -crlf
*.gif filter=lfs diff=lfs merge=lfs -crlf"

  if [ "$expected" != "$(cat .gitattributes)" ]; then
    exit 1
  fi
)
end_test

begin_test "track outside git repo"
(
  set -e

  git lfs track "*.foo" || {
    # this fails if it's run outside of a git repo using GIT_LFS_TEST_DIR

    # git itself returns an exit status of 128
    # $ git show
    # fatal: Not a git repository (or any of the parent directories): .git
    # $ echo "$?"
    # 128

    [ "$?" = "128" ]
    exit 0
  }

  if [ -n "$GIT_LFS_TEST_DIR" ]; then
    echo "GIT_LFS_TEST_DIR should be set outside of any Git repository"
    exit 1
  fi
)
end_test
