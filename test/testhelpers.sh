#!/usr/bin/env bash

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

# assert_local_object confirms that an object file is stored for the given oid &
# has the correct size
# $ assert_local_object "some-oid" size
assert_local_object() {
  local oid="$1"
  local size="$2"
  local cfg=`git lfs env | grep LocalMediaDir`
  local f="${cfg:14}/${oid:0:2}/${oid:2:2}/$oid"
  actualsize=$(wc -c <"$f" | tr -d '[[:space:]]')
  if [ "$size" != "$actualsize" ]; then
    exit 1
  fi
}

# refute_local_object confirms that an object file is NOT stored for an oid
# $ refute_local_object "some-oid"
refute_local_object() {
  local oid="$1"
  local cfg=`git lfs env | grep LocalMediaDir`
  local regex="LocalMediaDir=(\S+)"
  local f="${cfg:14}/${oid:0:2}/${oid:2:2}/$oid"
  if [ -e $f ]; then
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

# Delete an object on the lfs server. HTTP log is
# written to http.log. JSON output is written to http.json.
#
#   $ delete_server_object "reponame" "oid"
delete_server_object() {
  local reponame="$1"
  local oid="$2"
  curl -v "$GITSERVER/$reponame.git/info/lfs/delete/$oid" \
    -u "user:pass" \
    -o http.json \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log

  grep "200 OK" http.log
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

# creates a bare remote repository for a local clone. Useful to test pushing to
# a fresh remote server.
#
#   $ setup_alternate_remote "$reponame-whatever"
#   $ setup_alternate_remote "$reponame-whatever" "other-remote-name"
#
setup_alternate_remote() {
  local newRemoteName=$1
  local remote=${2:-origin}

  wd=`pwd`

  setup_remote_repo "$newRemoteName"
  cd $wd
  git remote rm "$remote"
  git remote add "$remote" "$GITSERVER/$newRemoteName"
}

# clone_repo clones a repository from the test Git server to the subdirectory
# $dir under $TRASHDIR. setup_remote_repo() needs to be run first. Output is
# written to clone.log.
clone_repo() {
  cd "$TRASHDIR"

  local reponame="$1"
  local dir="$2"
  echo "clone local git repository $reponame to $dir"
  out=$(git clone "$GITSERVER/$reponame" "$dir" 2>&1)
  cd "$dir"

  git config credential.helper lfstest
  echo "$out" > clone.log
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

  # Set up the initial git config and osx keychain if applicable
  HOME="$TESTHOME"
  mkdir "$HOME"
  git lfs init
  git config --global credential.helper lfstest
  git config --global user.name "Git LFS Tests"
  git config --global user.email "git-lfs@example.com"
  grep "git-lfs clean" "$REMOTEDIR/home/.gitconfig" > /dev/null || {
    echo "global git config should be set in $REMOTEDIR/home"
    ls -al "$REMOTEDIR/home"
    exit 1
  }

  # setup the git credential password storage
  mkdir -p "$CREDSDIR"
  printf "user:pass" > "$CREDSDIR/127.0.0.1"

  echo
  echo "HOME: $HOME"
  echo "TMP: $TMPDIR"
  echo "CREDS: $CREDSDIR"
  echo "lfstest-gitserver:"
  echo "  LFSTEST_URL=$LFS_URL_FILE"
  echo "  LFSTEST_DIR=$REMOTEDIR"
  echo "GIT:"
  git config --global --get-regexp "lfs|credential|user"

  if [ "$OSXKEYFILE" ]; then
    # Only OS X will encounter this
    # We can't disable osxkeychain and it gets called on store as well as ours,
    # reporting "A keychain cannot be found to store.." errors because the test
    # user env has no keychain; so create one
    mkdir -p $HOME/Library/Preferences # required to store keychain lists
    security create-keychain -p pass "$OSXKEYFILE"
    security list-keychains -s "$OSXKEYFILE"
    security unlock-keychain -p pass "$OSXKEYFILE"
    security set-keychain-settings -lut 7200 "$OSXKEYFILE"
    security default-keychain -s "$OSXKEYFILE"

    echo "OSX Keychain: $OSXKEYFILE"
  fi

  wait_for_file "$LFS_URL_FILE"

  echo
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

    if [ "$OSXKEYFILE" ]; then
      # explicitly clean up keychain to make sure search list doesn't look for it
      # shouldn't matter because $HOME is separate & keychain prefs are there but still
      security delete-keychain "$OSXKEYFILE"
    fi

    [ -z "$KEEPTRASH" ] && rm -rf "$REMOTEDIR"
  fi
}

ensure_git_version_isnt() {
  local expectedComparison=$1
  local version=$2

  local gitVersion=$(git version | cut -d" " -f3)

  set +e
  compare_version $gitVersion $version
  result=$?
  set -e

  if [[ $result == $expectedComparison ]]; then
    echo "skip: $0 (git version $(comparison_to_operator $expectedComparison) $version)"
    exit
  fi
}

VERSION_EQUAL=0
VERSION_HIGHER=1
VERSION_LOWER=2

# Compare $1 and $2 and return VERSION_EQUAL / VERSION_LOWER / VERSION_HIGHER
compare_version() {
    if [[ $1 == $2 ]]
    then
        return $VERSION_EQUAL
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    # fill empty fields in ver1 with zeros
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++))
    do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++))
    do
        if [[ -z ${ver2[i]} ]]
        then
            # fill empty fields in ver2 with zeros
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]}))
        then
            return $VERSION_HIGHER
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]}))
        then
            return $VERSION_LOWER
        fi
    done
    return $VERSION_EQUAL
}

comparison_to_operator() {
  local comparison=$1
  if [[ $1 == $VERSION_EQUAL ]]; then
    echo "=="
  elif [[ $1 == $VERSION_HIGHER ]]; then
    echo ">"
  elif [[ $1 == $VERSION_LOWER ]]; then
    echo "<"
  else
    echo "???"
  fi
}

calc_oid() {
  printf "$1" | shasum -a 256 | cut -f 1 -d " "
}

# Get a date string with an offset
# Args: One or more date offsets of the form (regex) "[+-]\d+[dmyHM]"
# e.g. +1d = 1 day forward from today
#      -5y = 5 years before today
# Example call:
#   D=$(get_date +1y +1m -5H)
# returns date as string in RFC3339 format ccyy-mm-ddThh:MM:ssZ
# note returns in UTC time not local time hence Z and not +/-
get_date() {
  # Wrapped because BSD (inc OSX) & GNU 'date' functions are different
  # on Windows under Git Bash it's GNU
  if date --version >/dev/null 2>&1 ; then # GNU
    ARGS=""
    for var in "$@"
    do
        # GNU offsets are more verbose
        unit=${var: -1}
        val=${var:0:${#var}-1}
        case "$unit" in
          d) unit="days" ;;
          m) unit="months" ;;
          y) unit="years"  ;;
          H) unit="hours"  ;;
          M) unit="minutes" ;;
        esac
        ARGS="$ARGS $val $unit"
    done
    date -d "$ARGS" -u +%Y-%m-%dT%TZ
  else # BSD
    ARGS=""
    for var in "$@"
    do
        ARGS="$ARGS -v$var"
    done
    date $ARGS -u +%Y-%m-%dT%TZ
  fi
}

# Convert potentially MinGW bash paths to native Windows paths
# Needed to match generic built paths in test scripts to native paths generated from Go
native_path() {
  local arg=$1
  if [ $IS_WINDOWS == "1" ]; then
    # Use params form to avoid interpreting any '\' characters
    printf '%s' "$(cygpath -w $arg)"
  else
    printf '%s' "$arg"
  fi
}

# escape any instance of '\' with '\\' on Windows
escape_path() {
  local unescaped="$1"
  if [ $IS_WINDOWS == "1" ]; then
    printf '%s' "${unescaped//\\/\\\\}"
  else
    printf '%s' "$unescaped"
  fi
}

# As native_path but escape all backslash characters to "\\"
native_path_escaped() {
  local unescaped=$(native_path "$1")
  escape_path "$unescaped"
}

# Compare 2 lists which are newline-delimited in a string, ignoring ordering and blank lines
contains_same_elements() {
  # Remove blank lines then sort
  printf '%s' "$1" | grep -v '^$' | sort > a.txt
  printf '%s' "$2" | grep -v '^$' | sort > b.txt

  set +e
  diff -u a.txt b.txt 1>&2
  res=$?
  set -e
  rm a.txt b.txt
  exit $res
}

is_stdin_attached() {
  test -t0
  echo $?
}

has_test_dir() {
  if [ -z "$GIT_LFS_TEST_DIR" ]; then
    echo "No GIT_LFS_TEST_DIR. Skipping..."
    exit 0
  fi
}
