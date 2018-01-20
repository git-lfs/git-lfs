#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "tus-upload-uninterrupted"
(
  set -e

  # this repo name is the indicator to the server to use tus
  reponame="test-tus-upload"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame
  git config lfs.tustransfers true

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="send-verify-action"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin master 2>&1 | tee pushtus.log
  grep "xfer: tus.io uploading" pushtus.log

  assert_server_object "$reponame" "$contents_oid"

)
end_test

begin_test "tus-upload-interrupted-resume"
(
  set -e

  # this repo name is the indicator to the server to use tus, AND to
  # interrupt the upload part way
  reponame="test-tus-upload-interrupt"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame
  git config lfs.tustransfers true

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents_verify="send-verify-action"
  contents_verify_oid="$(calc_oid "$contents_verify")"

  # this string announces to server that we want it to abort the download part
  # way, but reject the Range: header and fall back on re-downloading instead
  contents="234587134187634598o634857619384765b747qcvtuedvoaicwtvseudtvcoqi7280r7qvow4i7r8c46pr9q6v9pri6ioq2r8"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  printf "$contents_verify" > verify.dat
  git add a.dat verify.dat
  git add .gitattributes
  git commit -m "add a.dat, verify.dat" 2>&1 | tee commit.log
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin master 2>&1 | tee pushtus_resume.log
  # first attempt will start from the beginning
  grep "xfer: tus.io uploading" pushtus_resume.log
  grep "HTTP: 500" pushtus_resume.log
  # that will have failed but retry on 500 will resume it
  grep "xfer: tus.io resuming" pushtus_resume.log
  grep "HTTP: 204" pushtus_resume.log

  # should have completed in the end
  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents_verify_oid"

)
end_test
