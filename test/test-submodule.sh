#!/bin/sh

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

  cat sub/dir/README | grep "sub module"
)
end_test

begin_test "submodule env"
(
  set -e

  # using the local clone from the above test
  cd repo
  expected=$(printf "Endpoint=$GITSERVER/$reponame.git/info/lfs
LocalWorkingDir=$TRASHDIR/repo
LocalGitDir=$TRASHDIR/repo/.git
LocalMediaDir=$TRASHDIR/repo/.git/lfs/objects
TempDir=$TRASHDIR/repo/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]

  cd ../sub

  [ "$expected" == "$actual" ]
)
end_test
