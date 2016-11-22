#!/usr/bin/env bash
# Tests issues with paths over 248 characters. See
# https://github.com/golang/go/blob/2925427a47f41622f28f84889ad7aade27581144/src/os/path_windows.go#L131-L138

. "test/testlib.sh"

begin_test "(long path) fetch"
(
  set -e

  reponame="long-paths"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo
  git lfs track "*.bin"
  mkdir -p foo/bar/objects
  echo "hi" > foo/bar/objects/hi.bin
  echo "long filename" > "foo/bar/objects/$(printf '%*s' "64" | tr ' ' "0").bin"
  git add foo .gitattributes
  git commit -m "add files"
  git show

  assert_local_object "7da4fb02e8b23ed26d05e4a955565a7b1ff77846d0bf0d9ce5651bf20ed2e5d3" "14"
  assert_local_object "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4" "3"
  refute_server_object "$reponame" "7da4fb02e8b23ed26d05e4a955565a7b1ff77846d0bf0d9ce5651bf20ed2e5d3"
  refute_server_object "$reponame" "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4"
  git push origin master
  assert_server_object "$reponame" "7da4fb02e8b23ed26d05e4a955565a7b1ff77846d0bf0d9ce5651bf20ed2e5d3"
  assert_server_object "$reponame" "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4"

  cd ..
  root=`pwd`
  rootsize=${#root}
  clonedirsize=$((200-$rootsize))
  clonedir=$(printf '%*s' "$clonedirsize" | tr ' ' "0")
  clone_repo "$reponame" "$clonedir"
  assert_local_object "7da4fb02e8b23ed26d05e4a955565a7b1ff77846d0bf0d9ce5651bf20ed2e5d3" "14"
  assert_local_object "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4" "3"
)
end_test
