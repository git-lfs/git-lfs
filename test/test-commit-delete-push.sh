#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "commit, delete, then push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"

  deleted_oid=$(echo "deleted" | shasum -a 256 | cut -f 1 -d " ")
  echo "deleted" > deleted.dat
  git add deleted.dat .gitattributes
  git commit -m "add deleted file"

  git lfs push origin master --dry-run | grep "push ee31ef227442936872744b50d3297385c08b40ffc7baeaf34a39e6d81d6cd9ee => deleted.dat"

  assert_pointer "master" "deleted.dat" "$deleted_oid" 8

  added_oid=$(echo "added" | shasum -a 256 | cut -f 1 -d " ")
  echo "added" > added.dat
  git add added.dat
  git commit -m "add file"

  git lfs push origin master --dry-run | tee dryrun.log
  grep "push ee31ef227442936872744b50d3297385c08b40ffc7baeaf34a39e6d81d6cd9ee => deleted.dat" dryrun.log
  grep "push 3428719b7688c78a0cc8ba4b9e80b4e464c815fbccfd4b20695a15ffcefc22af => added.dat" dryrun.log

  git rm deleted.dat
  git commit -m "did not need deleted.dat after all"

  GIT_TRACE=1 git lfs push origin master --dry-run 2>&1 | tee dryrun.log
  grep "push ee31ef227442936872744b50d3297385c08b40ffc7baeaf34a39e6d81d6cd9ee => deleted.dat" dryrun.log
  grep "push 3428719b7688c78a0cc8ba4b9e80b4e464c815fbccfd4b20695a15ffcefc22af => added.dat" dryrun.log

  git log
  GIT_TRACE=1 git push origin master 2>&1 > push.log || {
    cat push.log
    git lfs logs last
    exit 1
  }
  grep "(2 of 2 files)" push.log | cat push.log

  assert_server_object "$reponame" "$deleted_oid"
  assert_server_object "$reponame" "$added_oid"
)
end_test
