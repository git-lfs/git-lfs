#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

urlify () {
  if [ "$IS_WINDOWS" -eq 1 ]
  then
    echo "$1" | sed -e 's,\\,/,g' -e 's,:,%3a,g' -e 's, ,%20,g'
  else
    echo "$1"
  fi
}

do_upload_download_test () {
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  git checkout -b test

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

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin test 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  [ ${PIPESTATUS[0]} = "0" ]

  ourobjects=$(cd .git && find lfs/objects -type f | sort)
  theirobjects=$(cd $gitdir && find lfs/objects -type f | sort)

  # Make sure the lock verification is not attempted.
  grep "locks/verify$" pushcustom.log && false

  grep "xfer: started custom adapter process" pushcustom.log
  grep "Uploading LFS objects: 100% (12/12)" pushcustom.log
  [ "$ourobjects" = "$theirobjects" ]

  rm -rf .git/lfs/objects
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs fetch --all  2>&1 | tee fetchcustom.log
  [ ${PIPESTATUS[0]} = "0" ]

  objectlist=$(find .git/lfs/objects -type f)
  [ "$(echo "$objectlist" | wc -l)" -eq 12 ]
}

begin_test "standalone-file-upload-download-bare"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="standalone-file-upload-download-bare"
  setup_remote_repo "$reponame"

  git init --bare "$reponame-2.git"
  gitdir="$(pwd)/$reponame-2.git"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  git remote set-url origin "file://$(urlify "$gitdir")"

  do_upload_download_test
)
end_test

begin_test "standalone-file-upload-download-non-bare"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="standalone-file-upload-download-non-bare"
  setup_remote_repo "$reponame"

  git init "$reponame-2.git"
  repo2="$(pwd)/$reponame-2.git"
  gitdir="$(pwd)/$reponame-2.git/.git"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  git remote set-url origin "file://$(urlify "$repo2")"

  do_upload_download_test
)
end_test

begin_test "standalone-file-download-missing-file"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="standalone-file-download-missing-file"
  setup_remote_repo "$reponame"

  otherrepo="$(pwd)/$reponame-2.git"
  git init --bare "$otherrepo"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  git remote set-url origin "file://$(urlify "$otherrepo")"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  git checkout -b test

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

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin test 2>&1 | tee pushcustom.log
  # use PIPESTATUS otherwise we get exit code from tee
  [ ${PIPESTATUS[0]} = "0" ]

  # Make sure the lock verification is not attempted.
  grep "locks/verify$" pushcustom.log && false

  grep "xfer: started custom adapter process" pushcustom.log
  grep "Uploading LFS objects: 100% (12/12)" pushcustom.log

  # Delete an object from the remote side. Any object will do.
  rm -f $(find "$otherrepo/lfs/objects" -type f | head -n1)

  rm -rf .git/lfs/objects
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs fetch --all  2>&1 | tee fetchcustom.log
  # Make sure we failed.
  [ ${PIPESTATUS[0]} != "0" ]

  # Make sure we downloaded the rest of the objects.
  objectlist=$(find .git/lfs/objects -type f)
  [ "$(echo "$objectlist" | wc -l)" -eq 11 ]
)
end_test

begin_test "standalone-file-clone"
(
  set -e

  # setup a git repo to be used as a local repo, not remote
  reponame="standalone-file-clone"
  setup_remote_repo "$reponame"

  # clone directly, not through lfstest-gitserver
  clone_repo_url "$REMOTEDIR/$reponame.git" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  git checkout -b test

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

  testdir="$(pwd)"

  cd "$TRASHDIR"

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git clone "file://$(urlify "$testdir")" repo3  2>&1 | tee clonecustom.log

  grep "xfer: started custom adapter process" clonecustom.log
)
end_test
