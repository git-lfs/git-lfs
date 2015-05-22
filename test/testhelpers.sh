#!/bin/sh

# assert_pointer confirms that the pointer in the repository for $path in the
# given $ref matches the given $oid and $size.
#
#   $ assert_pointer "master" "path/to/file" "some-oid" 123
assert_pointer() {
  local ref=$1
  local path=$2
  local oid=$3
  local size=$4

  tree=$(git ls-tree -lr "$ref")
  gitblob=$(echo "$tree" | grep "$path" | cut -f 3 -d " ")
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

# pointer returns a string Git LFS pointer file.
#
#   $ pointer abc-some-oid 123
#   > version ...
pointer() {
  local oid=$1
  local size=$2
  printf "version https://git-lfs.github.com/spec/v1
oid sha256:%s
size %s
" "$oid" "$size"
}

# wait_for_file simply sleeps until a file exists.
#
#   $ wait_for_file "path/to/upcoming/file"
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

# setup_remote_repo intializes a bare Git repository that is accessible through
# the test Git server. The `pwd` is set to the repository's directory, in case
# further commands need to be run. This server is running for every test in a
# script/integration run, so every test file should setup its own remote
# repository to avoid conflicts.
#
#   $ setup_remote_repo "some-name"
#
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
  # and the custom credential helper. This overrides any Git config that is
  # already setup on the system.
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

# clone_repo clones a repository from the test Git server to the subdirectory
# $dir under $TRASHDIR. setup_remote_repo() needs to be run first.
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

# setup initializes the clean, isolated environment for integration tests.
setup() {
  cd "$ROOTDIR"

  rm -rf "$REMOTEDIR"
  mkdir "$REMOTEDIR"

  if [ -z "$SKIPCOMPILE" ]; then
    echo "compile git-lfs for $0"
    script/bootstrap
  fi

  $GITLFS version

  if [ -z "$SKIPCOMPILE" ]; then
    for go in test/cmd/*.go; do
      go build -o "$BINPATH/$(basename $go .go)" "$go"
    done
  fi

  echo "tmp dir: $TMPDIR"
  echo "remote git dir: $REMOTEDIR"
  echo "LFSTEST_URL=$LFS_URL_FILE LFSTEST_DIR=$REMOTEDIR lfstest-gitserver"
  LFSTEST_URL="$LFS_URL_FILE" LFSTEST_DIR="$REMOTEDIR" lfstest-gitserver > "$REMOTEDIR/gitserver.log" 2>&1 &
  wait_for_file "$LFS_URL_FILE"
}

# shutdown cleans the $TRASHDIR and shuts the test Git server down.
shutdown() {
  # every test/test-*.sh file should cleanup its trashdir
  [ -z "$KEEPTRASH" ] && rm -rf "$TRASHDIR"

  if [ "$SHUTDOWN_LFS" != "no" ]; then
    # only cleanup test/remote after script/integration done OR a single
    # test/test-*.sh file is run manually.
    [ -z "$KEEPTRASH" ] && rm -rf "$REMOTEDIR"
    if [ -s "$LFS_URL_FILE" ]; then
      curl "$(cat "$LFS_URL_FILE")/shutdown"
    fi
  fi
}
