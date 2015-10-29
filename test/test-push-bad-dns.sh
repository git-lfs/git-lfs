#!/usr/bin/env bash

. "test/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.3.0"

begin_test "push: upload to bad dns"
(
  set -e

  reponame="$(basename "$0" ".sh")-bad-dns"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  git add .gitattributes good.dat
  git commit -m "welp"

  port="$(echo "http://127.0.0.1:63378" | cut -f 3 -d ":")"
  git config lfs.url "http://git-lfs-bad-dns:$port"

  set +e
  GIT_TERMINAL_PROMPT=0 git push origin master
  res="$?"
  set -e

  refute_server_object "$reponame" "$(calc_oid "hi")"
  if [ "$res" = "0" ]; then
    echo "push successful?"
    exit 1
  fi
)
end_test

begin_test "push (legacy): upload to bad dns"
(
  set -e

  reponame="$(basename "$0" ".sh")-bad-dns-legacy"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  git add .gitattributes good.dat
  git commit -m "welp"

  port="$(echo "http://127.0.0.1:63378" | cut -f 3 -d ":")"
  git config lfs.batch false
  git config lfs.url "http://git-lfs-bad-dns:$port"

  set +e
  GIT_TERMINAL_PROMPT=0 git push origin master
  res="$?"
  set -e

  refute_server_object "$reponame" "$(calc_oid "hi")"
  if [ "$res" = "0" ]; then
    echo "push successful?"
    exit 1
  fi
)
end_test
