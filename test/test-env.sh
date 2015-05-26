#!/bin/sh

. "test/testlib.sh"

begin_test "env with no remote"
(
  set -e
  reponame="env-no-remote"
  mkdir $reponame
  cd $reponame
  git init

  expected=$(printf "LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]
)
end_test

begin_test "env with origin remote"
(
  set -e
  reponame="env-origin-remote"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"

  expected=$(printf "Endpoint=$GITSERVER/$reponame.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]
)
end_test

begin_test "env with multiple remotes"
(
  set -e
  reponame="env-multiple-remotes"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"

  expected=$(printf "Endpoint=$GITSERVER/env-origin-remote.git/info/lfs
Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]
)
end_test

begin_test "env with other remote"
(
  set -e
  reponame="env-other-remote"
  mkdir $reponame
  cd $reponame
  git init
  git remote add other "$GITSERVER/env-other-remote"

  expected=$(printf "Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]
)
end_test

begin_test "env with multiple remotes and lfs.url config"
(
  set -e
  reponame="env-multiple-remotes-with-lfs-url"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"

  expected=$(printf "Endpoint=http://foo/bar
Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]
)
end_test

begin_test "env with multiple remotes and lfs configs"
(
  set -e
  reponame="env-multiple-remotes-lfs-configs"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"
  git config remote.origin.lfsurl "http://custom/origin"
  git config remote.other.lfsurl "http://custom/other"

  expected=$(printf "Endpoint=http://foo/bar
Endpoint (other)=http://custom/other
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
GIT_LFS_TEST_DIR=$TMPDIR
GIT_LFS_TEST_MAXPROCS=4
")
  actual=$(git lfs env)
  [ "$expected" == "$actual" ]

  cd .git

  [ "$expected" == "$actual" ]
)
end_test
