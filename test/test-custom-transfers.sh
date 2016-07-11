#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "custom-transfer"
(
  set -e

  # this repo name is the indicator to the server to support custom transfer
  reponame="test-custom-transfer-1"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git config lfs.customtransfer.testcustom.path lfs-custom-transfer-agent

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="jksgdfljkgsdlkjafg lsjdgf alkjgsd lkfjag sldjkgf alkjsgdflkjagsd kljfg asdjgf kalsd"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  GIT_TRACE=1 git push origin master 2>&1 | tee pushcustom.log
  grep "xfer: Custom transfer adapter" pushcustom.log


)
end_test

