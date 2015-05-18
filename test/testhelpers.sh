#!/bin/sh

shutdown() {
  if [ "$SHUTDOWN_LFS" != "no" ]; then
    curl "$GITSERVER/shutdown"
    rm -rf "$LFS_URL_FILE"
  fi
}

wait_for_file() {
  local filename=$1
  for ((n=30; n>0; n--)); do
    if [ -s $filename ]; then
      return 0
    fi

    sleep 0.5
  done

  return 1
}

setup_remote_repo() {
  local reponame=$1
  echo "set up 'remote' git repository: $reponame"
  repodir="$REMOTEDIR/$reponame.git"
  mkdir -p $repodir
  cd $repodir
  git init --bare
  git config http.receivepack true
  git config receive.denyCurrentBranch ignore

  cd $TRASHDIR
}

setup() {
  cd $ROOTDIR
  echo "compile git-lfs"
  script/bootstrap

  rm -rf "test/remote"
  mkdir -p "test/remote"

  LFSTEST_URL=$LFS_URL_FILE LFSTEST_DIR=$REMOTEDIR go run "$ROOTDIR/test/cmd/lfstest-gitserver.go" &
  wait_for_file $LFS_URL_FILE
}
