#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "list a single lock with bad ref"
(
  set -e

  reponame="locks-list-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "f" > f.dat
  git add .gitattributes f.dat
  git commit -m "add f.dat"
  git push origin master:other

  git checkout -b other
  git lfs lock --json "f.dat" | tee lock.log

  git checkout master
  git lfs locks --path "f.dat" 2>&1 | tee locks.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs lock \'a.dat\'' to fail"
    exit 1
  fi

  grep 'Expected ref "refs/heads/other", got "refs/heads/master"' locks.log
)
end_test

begin_test "list a single lock"
(
  set -e

  reponame="locks-list-master-branch-required"
  setup_remote_repo_with_file "$reponame" "f.dat"
  clone_repo "$reponame" "$reponame"

  git lfs lock --json "f.dat" | tee lock.log

  id=$(assert_lock lock.log f.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/master"

  git lfs locks --path "f.dat" | tee locks.log
  [ $(wc -l < locks.log) -eq 1 ]
  grep "f.dat" locks.log
  grep "Git LFS Tests" locks.log
)
end_test

begin_test "list a single lock (--json)"
(
  set -e

  reponame="locks_list_single_json"
  setup_remote_repo_with_file "$reponame" "f_json.dat"

  git lfs lock --json "f_json.dat" | tee lock.log

  id=$(assert_lock lock.log f_json.dat)
  assert_server_lock "$reponame" "$id"

  git lfs locks --json --path "f_json.dat" | tee locks.log
  grep "\"path\":\"f_json.dat\"" locks.log
  grep "\"owner\":{\"name\":\"Git LFS Tests\"}" locks.log
)
end_test

begin_test "list locks with a limit"
(
  set -e

  reponame="locks_list_limit"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "clone_$reponame"

  git lfs track "*.dat"
  echo "foo" > "g_1.dat"
  echo "bar" > "g_2.dat"

  git add "g_1.dat" "g_2.dat" ".gitattributes"
  git commit -m "add files" | tee commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 g_1.dat" commit.log
  grep "create mode 100644 g_2.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log


  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log

  git lfs lock --json "g_1.dat" | tee lock.log
  assert_server_lock "$reponame" "$(assert_log "lock.log" g_1.dat)"

  git lfs lock --json "g_2.dat" | tee lock.log
  assert_server_lock "$reponame" "$(assert_lock "lock.log" g_2.dat)"

  git lfs locks --limit 1 | tee locks.log
  [ $(wc -l < locks.log) -eq 1 ]
)
end_test

begin_test "list locks with pagination"
(
  set -e

  reponame="locks_list_paginate"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "clone_$reponame"

  git lfs track "*.dat"
  for i in $(seq 1 5); do
    echo "$i" > "h_$i.dat"
  done

  git add "h_1.dat" "h_2.dat" "h_3.dat" "h_4.dat" "h_5.dat" ".gitattributes"

  git commit -m "add files" | tee commit.log
  grep "6 files changed" commit.log
  for i in $(seq 1 5); do
    grep "create mode 100644 h_$i.dat" commit.log
  done
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log

  for i in $(seq 1 5); do
    git lfs lock --json "h_$i.dat" | tee lock.log
    assert_server_lock "$reponame" "$(assert_lock "lock.log" "h_$1.dat")"
  done

  # The server will return, at most, three locks at a time
  git lfs locks --limit 4 | tee locks.log
  [ $(wc -l < locks.log) -eq 4 ]
)
end_test

begin_test "cached locks"
(
  set -e

  reponame="cached_locks"
  setup_remote_repo "remote_$reponame"
  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat"
  echo "foo" > "cached1.dat"
  echo "bar" > "cached2.dat"

  git add "cached1.dat" "cached2.dat" ".gitattributes"
  git commit -m "add files" | tee commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 cached1.dat" commit.log
  grep "create mode 100644 cached2.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log

  git lfs lock --json "cached1.dat" | tee lock.log
  assert_server_lock "$(assert_lock "lock.log" cached1.dat)"

  git lfs lock --json "cached2.dat" | tee lock.log
  assert_server_lock "$(assert_lock "lock.log" cached2.dat)"

  git lfs locks --local | tee locks.log
  [ $(wc -l < locks.log) -eq 2 ]

  # delete the remote to prove we're using the local records
  git remote remove origin

  git lfs locks --local --path "cached1.dat" | tee locks.log
  [ $(wc -l < locks.log) -eq 1 ]
  grep "cached1.dat" locks.log

  git lfs locks --local --limit 1 | tee locks.log
  [ $(wc -l < locks.log) -eq 1 ]
)
end_test
