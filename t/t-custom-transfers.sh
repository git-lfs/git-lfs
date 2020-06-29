#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "custom-transfer-wrong-path"
(
  set -e

  # this repo name is the indicator to the server to support custom transfer
  reponame="test-custom-transfer-fail"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  # deliberately incorrect path
  git config lfs.customtransfer.testcustom.path path-to-nothing

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="jksgdfljkgsdlkjafg lsjdgf alkjgsd lkfjag sldjkgf alkjsgdflkjagsd kljfg asdjgf kalsd"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  res=${PIPESTATUS[0]}
  grep "xfer: adapter \"testcustom\" Begin()" pushcustom.log
  grep "xfer: Aborting worker process" pushcustom.log
  if [ "$res" = "0" ]; then
    echo "Push should have failed because of an incorrect custom transfer path."
    exit 1
  fi

)
end_test

begin_test "custom-transfer-upload-download"
(
  set -e

  # this repo name is the indicator to the server to support custom transfer
  reponame="test-custom-transfer-1"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  # set up custom transfer adapter
  git config lfs.customtransfer.testcustom.path lfstest-customadapter

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  # set up a decent amount of data so that there's work for multiple concurrent adapters
  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"verify.dat\",\"Size\":18,\"Data\":\"send-verify-action\"},
      {\"Filename\":\"file1.dat\",\"Size\":1024},
      {\"Filename\":\"file2.dat\",\"Size\":750}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":1050},
      {\"Filename\":\"file3.dat\",\"Size\":660},
      {\"Filename\":\"file4.dat\",\"Size\":230}]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":1200},
      {\"Filename\":\"file6.dat\",\"Size\":300}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":120},
      {\"Filename\":\"file5.dat\",\"Size\":450},
      {\"Filename\":\"file7.dat\",\"Size\":520},
      {\"Filename\":\"file8.dat\",\"Size\":2048}]
  }
  ]" | lfstest-testutils addcommits

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  [ ${PIPESTATUS[0]} = "0" ]

  grep "xfer: started custom adapter process" pushcustom.log
  grep "xfer\[lfstest-customadapter\]:" pushcustom.log
  grep "Uploading LFS objects: 100% (12/12)" pushcustom.log

  rm -rf .git/lfs/objects
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs fetch --all  2>&1 | tee fetchcustom.log
  [ ${PIPESTATUS[0]} = "0" ]

  grep "xfer: started custom adapter process" fetchcustom.log
  grep "xfer\[lfstest-customadapter\]:" fetchcustom.log

  grep "Terminating test custom adapter gracefully" fetchcustom.log

  objectlist=`find .git/lfs/objects -type f`
  [ "$(echo "$objectlist" | wc -l)" -eq 12 ]
)
end_test

begin_test "custom-transfer-standalone"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="test-custom-transfer-standalone"
  setup_remote_repo "$reponame"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  # set up custom transfer adapter to use a specific transfer agent
  git config lfs.customtransfer.testcustom.path lfstest-standalonecustomadapter
  git config lfs.customtransfer.testcustom.args "--arg1 '--arg2 --arg3' --arg4"
  git config lfs.customtransfer.testcustom.concurrent false
  git config lfs.standalonetransferagent testcustom
  export TEST_STANDALONE_BACKUP_PATH="$(pwd)/test-custom-transfer-standalone-backup"
  mkdir -p $TEST_STANDALONE_BACKUP_PATH
  rm -rf $TEST_STANDALONE_BACKUP_PATH/*

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  # set up a decent amount of data so that there's work for multiple concurrent adapters
  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"verify.dat\",\"Size\":18,\"Data\":\"send-verify-action\"},
      {\"Filename\":\"file1.dat\",\"Size\":1024},
      {\"Filename\":\"file2.dat\",\"Size\":750}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":1050},
      {\"Filename\":\"file3.dat\",\"Size\":660},
      {\"Filename\":\"file4.dat\",\"Size\":230}]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":1200},
      {\"Filename\":\"file6.dat\",\"Size\":300}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":120},
      {\"Filename\":\"file5.dat\",\"Size\":450},
      {\"Filename\":\"file7.dat\",\"Size\":520},
      {\"Filename\":\"file8.dat\",\"Size\":2048}]
  }
  ]" | lfstest-testutils addcommits

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  [ ${PIPESTATUS[0]} = "0" ]

  # Make sure the lock verification is not attempted.
  grep "locks/verify$" pushcustom.log && false

  grep "xfer: started custom adapter process" pushcustom.log
  grep "xfer\[lfstest-standalonecustomadapter\]:" pushcustom.log
  grep "Uploading LFS objects: 100% (12/12)" pushcustom.log

  rm -rf .git/lfs/objects
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs fetch --all  2>&1 | tee fetchcustom.log
  [ ${PIPESTATUS[0]} = "0" ]

  grep "xfer: started custom adapter process" fetchcustom.log
  grep "xfer\[lfstest-standalonecustomadapter\]:" fetchcustom.log

  grep "Terminating test custom adapter gracefully" fetchcustom.log

  # Test argument parsing.
  grep 'Saw argument "--arg1"' fetchcustom.log
  grep 'Saw argument "--arg2 --arg3"' fetchcustom.log
  grep 'Saw argument "--arg4"' fetchcustom.log

  objectlist=`find .git/lfs/objects -type f`
  [ "$(echo "$objectlist" | wc -l)" -eq 12 ]

  git lfs fsck
)
end_test

begin_test "custom-transfer-standalone-urlmatch"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="test-custom-transfer-standalone-urlmatch"
  setup_remote_repo "$reponame"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  # set up custom transfer adapter to use a specific transfer agent, using a URL prefix match
  git config lfs.customtransfer.testcustom.path lfstest-standalonecustomadapter
  git config lfs.customtransfer.testcustom.concurrent false
  git config remote.origin.lfsurl https://git.example.com/example/path/to/repo
  git config lfs.https://git.example.com/example/path/.standalonetransferagent testcustom
  git config lfs.standalonetransferagent invalid-agent

  # git config lfs.standalonetransferagent testcustom
  export TEST_STANDALONE_BACKUP_PATH="$(pwd)/test-custom-transfer-standalone-urlmatch-backup"
  mkdir -p $TEST_STANDALONE_BACKUP_PATH
  rm -rf $TEST_STANDALONE_BACKUP_PATH/*

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  # set up a decent amount of data so that there's work for multiple concurrent adapters
  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"verify.dat\",\"Size\":18,\"Data\":\"send-verify-action\"},
      {\"Filename\":\"file1.dat\",\"Size\":1024},
      {\"Filename\":\"file2.dat\",\"Size\":750}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":1050},
      {\"Filename\":\"file3.dat\",\"Size\":660},
      {\"Filename\":\"file4.dat\",\"Size\":230}]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":1200},
      {\"Filename\":\"file6.dat\",\"Size\":300}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":120},
      {\"Filename\":\"file5.dat\",\"Size\":450},
      {\"Filename\":\"file7.dat\",\"Size\":520},
      {\"Filename\":\"file8.dat\",\"Size\":2048}]
  }
  ]" | lfstest-testutils addcommits

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  [ ${PIPESTATUS[0]} = "0" ]

  # Make sure the lock verification is not attempted.
  grep "locks/verify$" pushcustom.log && false

  grep "xfer: started custom adapter process" pushcustom.log
  grep "xfer\[lfstest-standalonecustomadapter\]:" pushcustom.log
  grep "Uploading LFS objects: 100% (12/12)" pushcustom.log

  rm -rf .git/lfs/objects
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs fetch --all  2>&1 | tee fetchcustom.log
  [ ${PIPESTATUS[0]} = "0" ]

  grep "xfer: started custom adapter process" fetchcustom.log
  grep "xfer\[lfstest-standalonecustomadapter\]:" fetchcustom.log

  grep "Terminating test custom adapter gracefully" fetchcustom.log

  objectlist=`find .git/lfs/objects -type f`
  [ "$(echo "$objectlist" | wc -l)" -eq 12 ]

  git lfs fsck
)
end_test
