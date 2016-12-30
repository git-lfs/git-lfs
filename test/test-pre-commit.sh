#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "pre-commit with own locked files"
(
  set -e

  reponame="pre-commit-owned-locks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "$contents" > locked_pc_auth.dat
  git add locked_pc_auth.dat
  git commit -m "add locked_pc_auth.dat"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_pc_auth.dat" | tee lock.log
  grep "'locked_pc_auth.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  printf "authorized changes" >> locked_pc_auth.dat
  git add locked_pc_auth.dat
  git commit -m "locked changes"
)
end_test

begin_test "pre-commit with unowned locked files"
(
  set -e

  reponame="pre-commit-unowned-locks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config --local user.name "Example Locker"
  git config --local user.email "locker@xample.com"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "$contents" > locked_pc_unauth.dat
  git add locked_pc_unauth.dat
  git commit -m "add locked_pc_unauth.dat"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_pc_unauth.dat" | tee lock.log
  grep "'locked_pc_unauth.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  pushd "$TRASHDIR" >/dev/null
    clone_repo "$reponame" "$reponame-assert"

    printf "authorized changes" >> locked_pc_unauth.dat
    git add locked_pc_unauth.dat

    set +e
    git commit -m "locked_pc_unauth changes" 2>&1 | tee commit.log
    ok="${PIPESTATUS[0]}"
    set -e

    if [ "0" -eq "$ok" ]; then
      echo >&2 "ERR: expected \`git commit\` to fail, didn't..."
      exit 1
    fi

    grep "Some files are locked" commit.log
    grep "locked_pc_unauth.dat" commit.log
  popd >/dev/null
)
end_test
