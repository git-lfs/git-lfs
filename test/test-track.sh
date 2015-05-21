#!/bin/sh

. "test/testlib.sh"

begin_test "track"
(
  set -e

  # no need to setup a remote repo, since this test doesn't need to push or pull

  mkdir track
  cd track
  git init

  git lfs track "*.jpg"
  cat .gitattributes | grep "*.jpg"
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
