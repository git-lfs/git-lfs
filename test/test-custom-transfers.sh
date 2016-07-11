#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "custom-transfer-wrong-path"
(
  set -e

  # this repo name is the indicator to the server to support custom transfer
  reponame="test-custom-transfer-1"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  # deliberately incorrect path
  git config lfs.customtransfer.testcustom.path path-to-nothing

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="jksgdfljkgsdlkjafg lsjdgf alkjgsd lkfjag sldjkgf alkjsgdflkjagsd kljfg asdjgf kalsd"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  GIT_TRACE=1 git push origin master 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  res=${PIPESTATUS[0]}
  grep "xfer: Custom transfer adapter" pushcustom.log
  grep "Failed to start custom transfer command" pushcustom.log
  if [ "$res" = "0" ]; then
    echo "Push should have failed because of an incorrect custom transfer path."
    exit 1
  fi

)
end_test

