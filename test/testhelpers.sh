#!/bin/sh

# assert_pointer confirms that the pointer in the repository for $path in the
# given $ref matches the given $oid and $size.
#
#   $ assert_pointer "master" "path/to/file" "some-oid" 123
assert_pointer() {
  local ref="$1"
  local path="$2"
  local oid="$3"
  local size="$4"

  tree=$(git ls-tree -lr "$ref")
  gitblob=$(echo "$tree" | grep "$path" | cut -f 3 -d " ")
  actual=$(git cat-file -p $gitblob)
  expected=$(pointer $oid $size)

  if [ "$expected" != "$actual" ]; then
    exit 1
  fi
}

# check that the object does not exist in the git lfs server. HTTP log is
# written to http.log. JSON output is written to http.json.
#
#   $ refute_server_object "reponame" "oid"
refute_server_object() {
  local reponame="$1"
  local oid="$2"
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/$oid" \
    -u "user:pass" \
    -o http.json \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log

  grep "404 Not Found" http.log
}

# check that the object does exist in the git lfs server. HTTP log is written
# to http.log. JSON output is written to http.json.
assert_server_object() {
  local reponame="$1"
  local oid="$2"
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/$oid" \
    -u "user:pass" \
    -o http.json \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log
  grep "200 OK" http.log

  grep "download" http.json || {
    cat http.json
    exit 1
  }
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
  local filename="$1"
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
  local reponame="$1"
  echo "set up remote git repository: $reponame"
  repodir="$REMOTEDIR/$reponame.git"
  mkdir -p "$repodir"
  cd "$repodir"
  git init --bare
  git config http.receivepack true
  git config receive.denyCurrentBranch ignore
}

# clone_repo clones a repository from the test Git server to the subdirectory
# $dir under $TRASHDIR. setup_remote_repo() needs to be run first.
clone_repo() {
  cd "$TRASHDIR"

  local reponame="$1"
  local dir="$2"
  echo "clone local git repository $reponame to $dir"
  out=$(git clone "$GITSERVER/$reponame" "$dir" 2>&1)
  cd "$dir"

  git config credential.helper lfstest
  echo "$out"
}

# setup initializes the clean, isolated environment for integration tests.
setup() {
  cd "$ROOTDIR"

  rm -rf "$REMOTEDIR"
  mkdir "$REMOTEDIR"

  if [ -z "$SKIPCOMPILE" ] && [ -z "$LFS_BIN" ]; then
    echo "compile git-lfs for $0"
    script/bootstrap || {
      return $?
    }
  fi

  echo "Git LFS: ${LFS_BIN:-$(which git-lfs)}"
  git lfs version
  git version

  if [ -z "$SKIPCOMPILE" ]; then
    for go in test/cmd/*.go; do
      go build -o "$BINPATH/$(basename $go .go)" "$go"
    done
  fi

  LFSTEST_URL="$LFS_URL_FILE" LFSTEST_DIR="$REMOTEDIR" lfstest-gitserver > "$REMOTEDIR/gitserver.log" 2>&1 &

  mkdir $HOME
  git lfs init
  git config --global credential.helper lfstest
  git config --global user.name "Git LFS Tests"
  git config --global user.email "git-lfs@example.com"
  grep "git-lfs clean" "$REMOTEDIR/home/.gitconfig" > /dev/null || {
    echo "global git config should be set in $REMOTEDIR/home"
    ls -al "$REMOTEDIR/home"
    exit 1
  }
  cp "$HOME/.gitconfig" "$HOME/.gitconfig-backup"

  echo "HOME: $HOME"
  echo "TMP: $TMPDIR"
  echo "lfstest-gitserver:"
  echo "  LFSTEST_URL=$LFS_URL_FILE"
  echo "  LFSTEST_DIR=$REMOTEDIR"
  echo "GIT:"
  git config --global --get-regexp "lfs|credential|user"

  if [[ `git config --system credential.helper | grep osxkeychain` == "osxkeychain" ]]
  then
    # Only OS X will encounter this
    # We can't disable osxkeychain and it gets called on store as well as ours,
    # reporting "A keychain cannot be found to store.." errors because the test
    # user env has no keychain; so create one
    mkdir -p $TMPDIR
    mkdir -p $HOME/Library/Preferences # required to store keychain lists
    security create-keychain -p pass $TMPDIR/temp.keychain
    security list-keychains -s $TMPDIR/temp.keychain
    security unlock-keychain -p pass $TMPDIR/temp.keychain
    security set-keychain-settings -lut 7200 $TMPDIR/temp.keychain
    security default-keychain -s $TMPDIR/temp.keychain
  fi

  wait_for_file "$LFS_URL_FILE"
}

# shutdown cleans the $TRASHDIR and shuts the test Git server down.
shutdown() {
  # every test/test-*.sh file should cleanup its trashdir
  [ -z "$KEEPTRASH" ] && rm -rf "$TRASHDIR"

  if [ "$SHUTDOWN_LFS" != "no" ]; then
    # only cleanup test/remote after script/integration done OR a single
    # test/test-*.sh file is run manually.
    if [ -s "$LFS_URL_FILE" ]; then
      curl "$(cat "$LFS_URL_FILE")/shutdown"
    fi

    if [[ `git config --system credential.helper | grep osxkeychain` == "osxkeychain" ]]
    then
      # explicitly clean up keychain to make sure search list doesn't look for it
      # shouldn't matter because $HOME is separate & keychain prefs are there but still
      security delete-keychain $TMPDIR/temp.keychain
    fi

    [ -z "$KEEPTRASH" ] && rm -rf "$REMOTEDIR"

  fi

}
