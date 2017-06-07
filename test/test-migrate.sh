#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "migrate"
(
  set -e

  reponame="migrate"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="contents"
  contents_len="$(printf "$contents" | wc -c | awk '{ print $1 }')"
  contents_oid="$(calc_oid "$contents")"

  printf "$contents" > a.dat
  git add a.dat
  git commit -m "add a.dat"

  git rev-list --all | git lfs migrate -I "*.dat"

  [ "$contents" = "$(cat a.dat)" ]
  assert_local_object "$contents_oid" "$contents_len"

  git push origin master

  assert_server_object "$reponame" "$contents_oid"
)
end_test
