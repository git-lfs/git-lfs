#!/bin/sh

assert_pointer() {
  local ref=$1
  local path=$2
  local oid=$3
  local size=$4

  gitblob=$(git ls-tree -l $ref | grep $path | cut -f 3 -d " ")
  actual=$(git cat-file -p $gitblob)
  expected=$(pointer $oid $size)

  if [ "$expected" != "$actual" ]; then
    exit 1
  fi
}

# no-op.  check that the object does not exist in the git lfs server
refute_server_object() {
  echo "refute server object: no-op"
}

# no-op.  check that the object does exist in the git lfs server
assert_server_object() {
  echo "assert server object: no-op"
}

pointer() {
  local oid=$1
  local size=$2
  printf "version https://git-lfs.github.com/spec/v1
oid sha256:%s
size %s
" "$oid" "$size"
}

wait_for_file() {
  local filename=$1
  n=0
  while [ $n -lt 10 ]; do
    if [ -s $filename ]; then
      return 0
    fi

    sleep 0.5
    n=`expr $n + 1`
  done

  return 1
}

setup_remote_repo() {
  local reponame=$1
  echo "set up remote git repository: $reponame"
  repodir="$REMOTEDIR/$reponame.git"
  mkdir -p "$repodir"
  cd "$repodir"
  git init --bare
  git config http.receivepack true
  git config receive.denyCurrentBranch ignore

  # dump a simple git config file so clones use this test's Git LFS command
  # and the custom credential helper
  printf "[filter \"lfs\"]
	required = true
	smudge = %s smudge %%f
	clean = %s clean %%f
[credential]
	helper = %s
[remote \"origin\"]
	url = %s/%s
	fetch = +refs/heads/*:refs/remotes/origin/*
" "$GITLFS" "$GITLFS" lfstest "$GITSERVER" "$reponame" > "$LFS_CONFIG-$reponame"
}

clone_repo() {
  cd "$TRASHDIR"

  local reponame=$1
  local dir=$2
  echo "clone local git repository $reponame to $dir"
  out=$(GIT_CONFIG="$LFS_CONFIG-$reponame" git clone "$GITSERVER/$reponame" "$dir" 2>&1)
  cd "$dir"

  git config credential.helper lfstest
  echo "$out"
}

setup() {
  cd "$ROOTDIR"

  rm -rf "test/remote"
  mkdir "test/remote"

  echo "compile git-lfs for $0"
  script/bootstrap

  for go in test/cmd/*.go; do
    go build -o "$BINPATH/$(basename $go .go)" "$go"
  done

  echo "PATH=$BINPATH:\$PATH"
  echo "LFSTEST_URL=$LFS_URL_FILE LFSTEST_DIR=$REMOTEDIR lfstest-gitserver"
  LFSTEST_URL="$LFS_URL_FILE" LFSTEST_DIR="$REMOTEDIR" lfstest-gitserver > "$TRASHDIR/gitserver.log" 2>&1 &
  wait_for_file "$LFS_URL_FILE"
}

shutdown() {
  local failures=$0

  if [ "$SHUTDOWN_LFS" == "no" ]; then
    exit 0
  fi

  curl "$GITSERVER/shutdown"
  rm -rf "$LFS_URL_FILE"

  if [ -s "$TRASHDIR/gitserver.log" ]; then
    echo ""
    echo "gitserver.log:"
    cat "$TRASHDIR/gitserver.log"
  fi

  echo ""
  echo "env"
  env

  rm -rf "$TRASHDIR"
}
