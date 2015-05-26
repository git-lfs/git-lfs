#!/bin/sh

. "test/testlib.sh"

begin_test "pre-push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log |
    grep "(0 of 0 files) 0 B 0" || {
      cat push.log
      exit 1
    }

  git lfs track "*.dat"
  echo "hi" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show
  git lfs env

  # file isn't on the git lfs server yet
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4" \
    -u "user:pass" \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log |
    grep "404 Not Found" || {
      cat http.log
      exit 1
    }

  # push file to the git lfs server
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log |
    grep "(1 of 1 files) 3 B / 3 B  100.00 %" || {
      cat push.log
      exit 1
    }

  # now the file exists
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4" \
    -u "user:pass" \
    -o lfs.json \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log |
    grep "200 OK" || {
      cat http.log
      exit 1
    }

  grep "download" lfs.json || {
    cat lfs.json
    exit 1
  }
)
end_test

begin_test "pre-push dry-run"
(
  set -e

  reponame="$(basename "$0" ".sh")-dry-run"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-dry-run
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  [ "" == "$(cat push.log)" ]

  git lfs track "*.dat"
  echo "dry" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show
  git lfs env

  # file doesn't exist yet
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b" \
    -u "user:pass" \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log |
    grep "404 Not Found" || {
      cat http.log
      exit 1
    }

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log |
    grep "push hi.dat" || {
      cat push.log
      exit 1
    }

  # file still doesn't exist
  curl -v "$GITSERVER/$reponame.git/info/lfs/objects/2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b" \
    -u "user:pass" \
    -H "Accept: application/vnd.git-lfs+json" 2>&1 |
    tee http.log |
    grep "404 Not Found" || {
      cat http.log
      exit 1
    }
)
end_test
