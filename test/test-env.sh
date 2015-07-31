#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "env with no remote"
(
  set -e
  reponame="env-no-remote"
  mkdir $reponame
  cd $reponame
  git init

  expected=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]
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

  expected=$(printf "%s\n%s\n
Endpoint=$GITSERVER/$reponame.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
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

  expected=$(printf "%s\n%s\n
Endpoint=$GITSERVER/env-origin-remote.git/info/lfs
Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
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

  expected=$(printf "%s\n%s\n
Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
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

  expected=$(printf "%s\n%s\n
Endpoint=http://foo/bar
Endpoint (other)=$GITSERVER/env-other-remote.git/info/lfs
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
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

  expected=$(printf "%s\n%s\n
Endpoint=http://foo/bar
Endpoint (other)=http://custom/other
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
)
end_test

begin_test "env with multiple remotes and lfs url and batch configs"
(
  set -e
  reponame="env-multiple-remotes-lfs-batch-configs"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"
  git config lfs.batch true
  git config lfs.concurrenttransfers 5
  git config remote.origin.lfsurl "http://custom/origin"
  git config remote.other.lfsurl "http://custom/other"

  expected=$(printf "%s\n%s\n
Endpoint=http://foo/bar
Endpoint (other)=http://custom/other
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=5
BatchTransfer=true
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env)
  [ "$expected2" = "$actual2" ]
)
end_test

begin_test "env with .gitconfig"
(
  set -e
  reponame="env-with-gitconfig"

  git init $reponame
  cd $reponame

  git remote add origin "$GITSERVER/env-origin-remote"
  echo '[remote "origin"]
	lfsurl = http://foobar:8080/
[lfs]
     batch = true
	concurrenttransfers = 5
' > .gitconfig

  expected=$(printf "%s\n%s\n
Endpoint=http://foobar:8080/
LocalWorkingDir=$TRASHDIR/$reponame
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=5
BatchTransfer=true
$(env | grep "^GIT")
" "$(git lfs version)" "$(git version)")

  actual=$(git lfs env)
  [ "$expected" = "$actual" ]

  mkdir a
  cd a
  actual2=$(git lfs env)
  [ "$expected" = "$actual2" ]
)
end_test

begin_test "env with environment variables"
(
  set -e
  reponame="env-with-envvars"
  git init $reponame
  mkdir -p $reponame/a/b/c

  gitDir=$TRASHDIR/$reponame/.git
  workTree=$TRASHDIR/$reponame/a/b

  expected=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame/a/b
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree env | grep "^GIT")
" "$(git lfs version)" "$(git version)")

  actual=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env)
  [ "$expected" = "$actual" ]

  cd $TRASHDIR/$reponame
  actual2=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env)
  [ "$expected" = "$actual2" ]

  cd $TRASHDIR/$reponame/.git
  actual3=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env)
  [ "$expected" = "$actual3" ]

  cd $TRASHDIR/$reponame/a/b/c
  actual4=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env)
  [ "$expected" = "$actual4" ]

  expected5=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame/a/b
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(GIT_DIR=$gitDir GIT_WORK_TREE=a/b env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual5=$(GIT_DIR=$gitDir GIT_WORK_TREE=a/b git lfs env)
  [ "$expected5" = "$actual5" ]

  expected6=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame/a/b
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(GIT_WORK_TREE=a/b env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual6=$(GIT_WORK_TREE=a/b git lfs env)
  [ "$expected6" = "$actual6" ]

  cd $TRASHDIR/$reponame/a/b
  expected7=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame/a/b
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(GIT_DIR=$gitDir env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual7=$(GIT_DIR=$gitDir git lfs env)
  [ "$expected7" = "$actual7" ]

  cd $TRASHDIR/$reponame/a
  expected8=$(printf "%s\n%s\n
LocalWorkingDir=$TRASHDIR/$reponame/a/b
LocalGitDir=$TRASHDIR/$reponame/.git
LocalGitStorageDir=$TRASHDIR/$reponame/.git
LocalMediaDir=$TRASHDIR/$reponame/.git/lfs/objects
TempDir=$TRASHDIR/$reponame/.git/lfs/tmp
ConcurrentTransfers=3
BatchTransfer=false
$(GIT_WORK_TREE=$workTree env | grep "^GIT")
" "$(git lfs version)" "$(git version)")
  actual8=$(GIT_WORK_TREE=$workTree git lfs env)
  [ "$expected8" = "$actual8" ]
)
end_test
