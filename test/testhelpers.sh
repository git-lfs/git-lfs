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

  gitblob=$(git ls-tree -lrz "$ref" |
    while read -r -d $'\0' x; do
      echo $x
    done |
    grep "$path" | cut -f 3 -d " ")

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

# refute_local_object confirms that an object file is NOT stored for an oid.
# If "$size" is given as the second argument, assert that the file exists _and_
# that it does _not_ the expected size
#
# $ refute_local_object "some-oid"
# $ refute_local_object "some-oid" "123"
refute_local_object() {
  local oid="$1"
  local size="$2"
  local cfg=`git lfs env | grep LocalMediaDir`
  local regex="LocalMediaDir=(\S+)"
  local f="${cfg:14}/${oid:0:2}/${oid:2:2}/$oid"
  if [ -e $f ]; then
    if [ -z "$size" ]; then
      exit 1
    fi

    actual_size="$(wc -c < "$f" | awk '{ print $1 }')"
    if [ "$size" -eq "$actual_size" ]; then
      echo >&2 "fatal: expected object $oid not to have size: $size"
      exit 1
    fi
  fi
}

# delete_local_object deletes the local storage for an oid
# $ delete_local_object "some-oid"
delete_local_object() {
  local oid="$1"
  local cfg=`git lfs env | grep LocalMediaDir`
  local f="${cfg:14}/${oid:0:2}/${oid:2:2}/$oid"
  rm "$f"
}

# corrupt_local_object corrupts the local storage for an oid
# $ corrupt_local_object "some-oid"
corrupt_local_object() {
  local oid="$1"
  local cfg=`git lfs env | grep LocalMediaDir`
  local f="${cfg:14}/${oid:0:2}/${oid:2:2}/$oid"
  cp /dev/null "$f"
}


# check that the object does not exist in the git lfs server. HTTP log is
# written to http.log. JSON output is written to http.json.
#
#   $ refute_server_object "reponame" "oid"
refute_server_object() {
  local reponame="$1"
  local oid="$2"
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/batch" \
    -u "user:pass" \
    -o http.json \
    -d "{\"operation\":\"download\",\"objects\":[{\"oid\":\"$oid\"}]}" \
    -H "Accept: application/vnd.git-lfs+json" \
    -H "X-Check-Object: 1" \
    -H "X-Ignore-Retries: true" 2>&1 |
    tee http.log

  [ "0" = "$(grep -c "download" http.json)" ] || {
    cat http.json
    exit 1
  }
}

# Delete an object on the lfs server. HTTP log is
# written to http.log. JSON output is written to http.json.
#
#   $ delete_server_object "reponame" "oid"
delete_server_object() {
  local reponame="$1"
  local oid="$2"
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/$oid" \
    -X DELETE \
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
  local refspec="$3"
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/batch" \
    -u "user:pass" \
    -o http.json \
    -d "{\"operation\":\"download\",\"objects\":[{\"oid\":\"$oid\"}],\"ref\":{\"name\":\"$refspec\"}}" \
    -H "Accept: application/vnd.git-lfs+json" \
    -H "X-Check-Object: 1" \
    -H "X-Ignore-Retries: true" 2>&1 |
    tee http.log
  grep "200 OK" http.log

  grep "download" http.json || {
    cat http.json
    exit 1
  }
}

# This asserts the lock path and returns the lock ID by parsing the response of
#
#   git lfs lock --json <path>
assert_lock() {
  local log="$1"
  local path="$2"

  if [ $(grep -c "\"path\":\"$path\"" "$log") -eq 0 ]; then
    echo "path '$path' not found in:"
    cat "$log"
    exit 1
  fi

  local jsonid=$(grep -oh "\"id\":\"\w\+\"" "$log")
  echo "${jsonid:3}" | tr -d \"\:
}

# assert that a lock with the given ID exists on the test server
assert_server_lock() {
  local reponame="$1"
  local id="$2"
  local refspec="$3"

  curl -v "$GITSERVER/$reponame.git/info/lfs/locks?refspec=$refspec" \
    -u "user:pass" \
    -o http.json \
    -H "Accept:application/vnd.git-lfs+json" 2>&1 |
    tee http.log

  grep "200 OK" http.log
  grep "$id" http.json || {
    cat http.json
    exit 1
  }
}

# refute that a lock with the given ID exists on the test server
refute_server_lock() {
  local reponame="$1"
  local id="$2"
  local refspec="$3"

  curl -v "$GITSERVER/$reponame.git/info/lfs/locks?refspec=$refspec" \
    -u "user:pass" \
    -o http.json \
    -H "Accept:application/vnd.git-lfs+json" 2>&1 | tee http.log

  grep "200 OK" http.log

  [ $(grep -c "$id" http.json) -eq 0 ]
}

# Assert that .gitattributes contains a given attribute N times
assert_attributes_count() {
  local fileext="$1"
  local attrib="$2"
  local count="$3"

  pattern="\(*.\)\?$fileext\(.*\)$attrib"
  actual=$(grep -e "$pattern" .gitattributes | wc -l)
  if [ "$(printf "%d" "$actual")" != "$count" ]; then
    echo "wrong number of $attrib entries for $fileext"
    echo "expected: $count actual: $actual"
    cat .gitattributes
    exit 1
  fi
}

assert_file_writeable() {
  ls -l "$1" | grep -e "^-rw"
}

refute_file_writeable() {
  ls -l "$1" | grep -e "^-r-"
}

git_root() {
  git rev-parse --show-toplevel 2>/dev/null
}

dot_git_dir() {
  echo "$(git_root)/.git"
}

assert_hooks() {
  local git_root="$1"

  if [ -z "$git_root" ]; then
    echo >&2 "fatal: (assert_hooks) not in git repository"
    exit 1
  fi

  [ -x "$git_root/hooks/post-checkout" ]
  [ -x "$git_root/hooks/post-commit" ]
  [ -x "$git_root/hooks/post-merge" ]
  [ -x "$git_root/hooks/pre-push" ]
}

assert_clean_status() {
  status="$(git status)"
  echo "$status" | grep "working tree clean" || {
    echo $status
    git lfs status
  }
}

# pointer returns a string Git LFS pointer file.
#
#   $ pointer abc-some-oid 123 <version>
#   > version ...
pointer() {
  local oid=$1
  local size=$2
  local version=${3:-https://git-lfs.github.com/spec/v1}
  printf "version %s
oid sha256:%s
size %s
" "$version" "$oid" "$size"
}

# wait_for_file simply sleeps until a file exists.
#
#   $ wait_for_file "path/to/upcoming/file"
wait_for_file() {
  local filename="$1"
  n=0
  wait_time=1
  while [ $n -lt 17 ]; do
    if [ -s $filename ]; then
      return 0
    fi

    sleep $wait_time
    n=`expr $n + 1`
    if [ $wait_time -lt 4 ]; then
      wait_time=`expr $wait_time \* 2`
    fi
  done

  echo "$filename did not appear after 60 seconds."
  return 1
}

# setup_remote_repo initializes a bare Git repository that is accessible through
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

# clone_repo_url clones a Git repository to the subdirectory $dir under $TRASHDIR.
# setup_remote_repo() needs to be run first. Output is written to clone.log.
clone_repo_url() {
  cd "$TRASHDIR"

  local repo="$1"
  local dir="$2"
  echo "clone git repository $repo to $dir"
  out=$(git clone "$repo" "$dir" 2>&1)
  cd "$dir"

  git config credential.helper lfstest
  echo "$out" > clone.log
  echo "$out"
}

# clone_repo_ssl clones a repository from the test Git server to the subdirectory
# $dir under $TRASHDIR, using the SSL endpoint.
# setup_remote_repo() needs to be run first. Output is written to clone_ssl.log.
clone_repo_ssl() {
  cd "$TRASHDIR"

  local reponame="$1"
  local dir="$2"
  echo "clone local git repository $reponame to $dir"
  out=$(git clone "$SSLGITSERVER/$reponame" "$dir" 2>&1)
  cd "$dir"

  git config credential.helper lfstest

  echo "$out" > clone_ssl.log
  echo "$out"
}

# clone_repo_clientcert clones a repository from the test Git server to the subdirectory
# $dir under $TRASHDIR, using the client cert endpoint.
# setup_remote_repo() needs to be run first. Output is written to clone_client_cert.log.
clone_repo_clientcert() {
  cd "$TRASHDIR"

  local reponame="$1"
  local dir="$2"
  echo "clone $CLIENTCERTGITSERVER/$reponame to $dir"
  set +e
  out=$(git clone "$CLIENTCERTGITSERVER/$reponame" "$dir" 2>&1)
  res="${PIPESTATUS[0]}"
  set -e

  if [ "0" -eq "$res" ]; then
    cd "$dir"
    echo "$out" > clone_client_cert.log

    git config credential.helper lfstest
    exit 0
  fi

  echo "$out" > clone_client_cert.log
  if [ $(grep -c "NSInvalidArgumentException" clone_client_cert.log) -gt 0 ]; then
    echo "client-cert-mac-openssl" > clone_client_cert.log
    exit 0
  fi

  exit 1
}

# setup_remote_repo_with_file creates a remote repo, clones it locally, commits
# a file tracked by LFS, and pushes it to the remote:
#
#     setup_remote_repo_with_file "reponame" "filename"
setup_remote_repo_with_file() {
  local reponame="$1"
  local filename="$2"
  local dirname="$(dirname "$filename")"

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "clone_$reponame"

  mkdir -p "$dirname"

  git lfs track "$filename"
  echo "$filename" > "$filename"
  git add .gitattributes $filename
  git commit -m "add $filename" | tee commit.log

  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 $filename" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log
}

# substring_position returns the position of a substring in a 1-indexed search
# space.
#
#     [ "$(substring_position "foo bar baz" "baz")" -eq "9" ]
substring_position() {
  local str="$1"
  local substr="$2"

  # 1) Print the string...
  # 2) Remove the substring and everything after it
  # 3) Count the number of characters (bytes) left, i.e., the offset of the
  #    string we were looking for.

  echo "$str" \
    | sed "s/$substr.*$//" \
    | wc -c
}

# repo_endpoint returns the LFS endpoint for a given server and repository.
#
#     [ "$GITSERVER/example/repo.git/info/lfs" = "$(repo_endpoint $GITSERVER example-repo)" ]
repo_endpoint() {
  local server="$1"
  local repo="$2"

  echo "$server/$repo.git/info/lfs"
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
    [ $IS_WINDOWS -eq 1 ] && EXT=".exe"
    for go in test/cmd/*.go; do
      GO15VENDOREXPERIMENT=1 go build -o "$BINPATH/$(basename $go .go)$EXT" "$go"
    done
    if [ -z "$SKIPAPITESTCOMPILE" ]; then
      # Ensure API test util is built during tests to ensure it stays in sync
      GO15VENDOREXPERIMENT=1 go build -o "$BINPATH/git-lfs-test-server-api$EXT" "test/git-lfs-test-server-api/main.go" "test/git-lfs-test-server-api/testdownload.go" "test/git-lfs-test-server-api/testupload.go"
    fi
  fi

  LFSTEST_URL="$LFS_URL_FILE" LFSTEST_SSL_URL="$LFS_SSL_URL_FILE" LFSTEST_CLIENT_CERT_URL="$LFS_CLIENT_CERT_URL_FILE" LFSTEST_DIR="$REMOTEDIR" LFSTEST_CERT="$LFS_CERT_FILE" LFSTEST_CLIENT_CERT="$LFS_CLIENT_CERT_FILE" LFSTEST_CLIENT_KEY="$LFS_CLIENT_KEY_FILE" lfstest-gitserver > "$REMOTEDIR/gitserver.log" 2>&1 &

  wait_for_file "$LFS_URL_FILE"
  wait_for_file "$LFS_SSL_URL_FILE"
  wait_for_file "$LFS_CLIENT_CERT_URL_FILE"
  wait_for_file "$LFS_CERT_FILE"
  wait_for_file "$LFS_CLIENT_CERT_FILE"
  wait_for_file "$LFS_CLIENT_KEY_FILE"

  LFS_CLIENT_CERT_URL=`cat $LFS_CLIENT_CERT_URL_FILE`

  # Set up the initial git config and osx keychain if applicable
  HOME="$TESTHOME"
  mkdir "$HOME"
  git lfs install --skip-repo
  git config --global credential.usehttppath true
  git config --global credential.helper lfstest
  git config --global user.name "Git LFS Tests"
  git config --global user.email "git-lfs@example.com"
  git config --global http.sslcainfo "$LFS_CERT_FILE"
  git config --global http.$LFS_CLIENT_CERT_URL/.sslKey "$LFS_CLIENT_KEY_FILE"
  git config --global http.$LFS_CLIENT_CERT_URL/.sslCert "$LFS_CLIENT_CERT_FILE"
  git config --global http.$LFS_CLIENT_CERT_URL/.sslVerify "false"

  ( grep "git-lfs clean" "$REMOTEDIR/home/.gitconfig" > /dev/null && grep "git-lfs filter-process" "$REMOTEDIR/home/.gitconfig" > /dev/null ) || {
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
  echo "  LFSTEST_SSL_URL=$LFS_SSL_URL_FILE"
  echo "  LFSTEST_CLIENT_CERT_URL=$LFS_CLIENT_CERT_URL_FILE ($LFS_CLIENT_CERT_URL)"
  echo "  LFSTEST_CERT=$LFS_CERT_FILE"
  echo "  LFSTEST_CLIENT_CERT=$LFS_CLIENT_CERT_FILE"
  echo "  LFSTEST_CLIENT_KEY=$LFS_CLIENT_KEY_FILE"
  echo "  LFSTEST_DIR=$REMOTEDIR"
  echo "GIT:"
  git config --global --get-regexp "lfs|credential|user"

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
      curl -s "$(cat "$LFS_URL_FILE")/shutdown"
    fi

    [ -z "$KEEPTRASH" ] && rm -rf "$REMOTEDIR"

    # delete entire lfs test root if we created it (double check pattern)
    if [ -z "$KEEPTRASH" ] && [ "$RM_GIT_LFS_TEST_DIR" = "yes" ] && [[ $GIT_LFS_TEST_DIR == *"$TEMPDIR_PREFIX"* ]]; then
      rm -rf "$GIT_LFS_TEST_DIR"
    fi
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

# Calculate the object ID from the string passed as the argument
calc_oid() {
  printf "$1" | $SHASUM | cut -f 1 -d " "
}

# Calculate the object ID from the file passed as the argument
calc_oid_file() {
  $SHASUM "$1" | cut -f 1 -d " "
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
  if [ $IS_WINDOWS -eq 1 ]; then
    # Use params form to avoid interpreting any '\' characters
    printf '%s' "$(cygpath -w $arg)"
  else
    printf '%s' "$arg"
  fi
}

# escape any instance of '\' with '\\' on Windows
escape_path() {
  local unescaped="$1"
  if [ $IS_WINDOWS -eq 1 ]; then
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

cat_end() {
  if [ $IS_WINDOWS -eq 1 ]; then
    printf '^M$'
  else
    printf '$'
  fi
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

add_symlink() {
  local src=$1
  local dest=$2

  prefix=`git rev-parse --show-prefix`
  hashsrc=`printf "$src" | git hash-object -w --stdin`

  git update-index --add --cacheinfo 120000 "$hashsrc" "$prefix$dest"
  git checkout -- "$dest"
}
