#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"
lfsname="submodule-config-test-lfs"
reponame="submodule-config-test-repo"
submodname="submodule-config-test-submodule"

begin_test "submodule env with .lfsconfig"
(
  set -e

  # setup dummy repo with lfs store
  # no git data will be pushed, just lfs objects
  setup_remote_repo "$lfsname"
  echo $GITSERVER/$lfsname.git/info/lfs

  # setup submodule
  setup_remote_repo "$submodname"
  clone_repo "$submodname" submod
  mkdir dir
  git config -f .lfsconfig lfs.url "$GITSERVER/$lfsname.git/info/lfs"
  git lfs track "*.dat"
  submodcontent="submodule lfs file"
  submodoid=$(calc_oid "$submodcontent")
  printf "%s" "$submodcontent" > dir/test.dat
  git add .lfsconfig .gitattributes dir
  git commit -m "create submodule"
  git push origin main

  assert_server_object "$lfsname" "$submodoid"

  # setup repo with submodule
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo
  git config -f .lfsconfig lfs.url "$GITSERVER/$lfsname.git/info/lfs"
  git submodule add -b main "$GITSERVER/$submodname" sub
  git submodule update
  git lfs track "*.dat"
  mkdir dir
  repocontent="repository lfs file"
  repooid=$(calc_oid "$repocontent")
  printf "%s" "$repocontent" > dir/test.dat
  git add .gitattributes .lfsconfig .gitmodules dir sub
  git commit -m "create repo"
  git push origin main

  assert_server_object "$lfsname" "$repooid"

  echo "repo"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log

  cd sub
  echo "./sub"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log

  cd dir
  echo "./sub/dir"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log
)
end_test

begin_test "submodule update --init --remote with .lfsconfig"
(
  set -e
  clone_repo "$reponame" clone
  grep "$repocontent" dir/test.dat

  git submodule update --init --remote

  grep "$submodcontent" sub/dir/test.dat
)
end_test
