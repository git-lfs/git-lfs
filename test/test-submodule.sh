#!/usr/bin/env bash

. "test/testlib.sh"
reponame="submodule-test-repo"
submodname="submodule-test-submodule"

begin_test "submodule local git dir"
(
  set -e

  setup_remote_repo "$reponame"
  setup_remote_repo "$submodname"

  clone_repo "$submodname" submod
  mkdir dir
  echo "sub module" > dir/README
  git add dir/README
  git commit -a -m "submodule readme"
  git push origin master

  clone_repo "$reponame" repo
  git submodule add "$GITSERVER/$submodname" sub
  git submodule update
  git add .gitmodules sub
  git commit -m "add submodule"
  git push origin master

  grep "sub module" sub/dir/README || {
    echo "submodule not setup correctly?"
    cat sub/dir/README
    exit 1
  }
)
end_test

begin_test "submodule env"
(
  set -e

  # using the local clone from the above test
  cd repo

  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)$" env.log
  grep "LocalWorkingDir=$(native_path_escaped "$TRASHDIR/repo$")" env.log
  grep "LocalGitDir=$(native_path_escaped "$TRASHDIR/repo/.git$")" env.log
  grep "LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/repo/.git$")" env.log
  grep "LocalMediaDir=$(native_path_escaped "$TRASHDIR/repo/.git/lfs/objects$")" env.log
  grep "TempDir=$(native_path_escaped "$TRASHDIR/repo/.git/lfs/tmp$")" env.log

  cd .git

  echo "./.git"
  git lfs env | tee env.log
  cat env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)$" env.log
  grep "LocalWorkingDir=$" env.log
  grep "LocalGitDir=$(native_path_escaped "$TRASHDIR/repo/.git$")" env.log
  grep "LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/repo/.git$")" env.log
  grep "LocalMediaDir=$(native_path_escaped "$TRASHDIR/repo/.git/lfs/objects$")" env.log
  grep "TempDir=$(native_path_escaped "$TRASHDIR/repo/.git/lfs/tmp$")" env.log

  cd ../sub

  echo "./sub"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$submodname.git/info/lfs (auth=none)$" env.log
  grep "LocalWorkingDir=$(native_path_escaped "$TRASHDIR/repo/sub$")" env.log
  grep "LocalGitDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub$")" env.log
  grep "LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub$")" env.log
  grep "LocalMediaDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub/lfs/objects$")" env.log
  grep "TempDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub/lfs/tmp$")" env.log

  cd dir

  echo "./sub/dir"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$submodname.git/info/lfs (auth=none)$" env.log
  grep "LocalWorkingDir=$(native_path_escaped "$TRASHDIR/repo/sub$")" env.log
  grep "LocalGitDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub$")" env.log
  grep "LocalGitStorageDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub$")" env.log
  grep "LocalMediaDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub/lfs/objects$")" env.log
  grep "TempDir=$(native_path_escaped "$TRASHDIR/repo/.git/modules/sub/lfs/tmp$")" env.log
)
end_test
