#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "resume-http-range"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  # this string announces to server that we want http-range support and to
  # interrupt the transfer when started from 0 to cause resume
  contents="status-batch-resume-206"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  git push origin master 

  assert_server_object "$reponame" "$contents_oid"

  # delete local copy then fetch it back
  # server will abort the transfer mid way (so will error) when not resuming
  # then we can restart it
  rm -rf .git/lfs/objects
  git lfs fetch 2>&1 | tee fetchinterrupted.log
  refute_local_object "$contents_oid"

  # now fetch again, this should try to resume and server should send remainder
  # this time (it does not cut short when Range is requested)
  GIT_TRACE=1 git lfs fetch 2>&1 | tee fetchresume.log
  grep "http-range: server accepted resume" fetchresume.log
  assert_local_object "$contents_oid" "${#contents}"

)
end_test

