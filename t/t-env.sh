#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

envInitConfig='git config filter.lfs.process = "git-lfs filter-process"
git config filter.lfs.smudge = "git-lfs smudge -- %f"
git config filter.lfs.clean = "git-lfs clean -- %f"'

unset_vars() {
    # If set, these will cause the test to fail.
    unset GIT_LFS_NO_TEST_COUNT GIT_LFS_LOCK_ACQUIRE_DISABLED
}

begin_test "env with no remote"
(
  set -e
  reponame="env-no-remote"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")

  contains_same_elements "$expected" "$actual"
)
end_test

begin_test "env with origin remote"
(
  set -e
  reponame="env-origin-remote"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"

  endpoint="$GITSERVER/$reponame.git/info/lfs (auth=none)"
  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$endpoint" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with multiple remotes"
(
  set -e
  reponame="env-multiple-remotes"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"

  endpoint="$GITSERVER/env-origin-remote.git/info/lfs (auth=none)"
  endpoint2="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$endpoint" "$endpoint2" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with other remote"
(
  set -e
  reponame="env-other-remote"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add other "$GITSERVER/env-other-remote"

  endpoint="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf '%s
%s

Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$endpoint" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with multiple remotes and lfs.url config"
(
  set -e
  reponame="env-multiple-remotes-with-lfs-url"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=http://foo/bar (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with multiple remotes and lfs configs"
(
  set -e
  reponame="env-multiple-remotes-lfs-configs"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"
  git config remote.origin.lfsurl "http://custom/origin"
  git config remote.other.lfsurl "http://custom/other"

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=http://foo/bar (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with multiple remotes and batch configs"
(
  set -e
  reponame="env-multiple-remotes-lfs-batch-configs"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.concurrenttransfers 5
  git config remote.origin.lfsurl "http://foo/bar"
  git config remote.other.lfsurl "http://custom/other"

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=http://custom/other (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=5
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd .git
  expected2=$(echo "$expected" | sed -e 's/LocalWorkingDir=.*/LocalWorkingDir=/')
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected2" "$actual2"
)
end_test

begin_test "env with .lfsconfig"
(
  set -e
  reponame="env-with-lfsconfig"
  unset_vars

  git init $reponame
  cd $reponame

  git remote add origin "$GITSERVER/env-origin-remote"
  echo '[remote "origin"]
	lfsurl = http://foobar:8080/
[lfs]
     batch = false
	concurrenttransfers = 5
' > .lfsconfig
echo '[remote "origin"]
lfsurl = http://foobar:5050/
[lfs]
   batch = true
concurrenttransfers = 50
' > .gitconfig

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foobar:8080/ (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  mkdir a
  cd a
  actual2=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual2"
)
end_test

begin_test "env with environment variables"
(
  set -e
  reponame="env-with-envvars"
  unset_vars
  git init $reponame
  mkdir -p $reponame/a/b/c

  gitDir=$(canonical_path "$TRASHDIR/$reponame/.git")
  workTree=$(canonical_path "$TRASHDIR/$reponame/a/b")

  localwd=$(canonical_path "$TRASHDIR/$reponame/a/b")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars="$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree env | grep "^GIT" | sort)"
  expected=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")

  actual=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env \
            | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

  cd $TRASHDIR/$reponame
  actual2=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env \
            | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual2"

  cd $TRASHDIR/$reponame/.git
  actual3=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env \
            | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual3"

  cd $TRASHDIR/$reponame/a/b/c
  actual4=$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree git lfs env \
            | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual4"

  envVars="$(GIT_DIR=$gitDir GIT_WORK_TREE=a/b env | grep "^GIT" | sort)"

  # `a/b` is an invalid relative path from where we are now and results in an
  # error, so resulting output will have many fields blank or invalid
  mediaDir5=$(native_path "lfs/objects")
  tempDir5=$(native_path "lfs/tmp")
  expected5=$(printf '%s
%s

LocalWorkingDir=
LocalGitDir=
LocalGitStorageDir=
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=lfs
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
git config filter.lfs.process = ""
git config filter.lfs.smudge = ""
git config filter.lfs.clean = ""
' "$(git lfs version)" "$(git version)" "$mediaDir5" "$tempDir5" "$envVars")
  actual5=$(GIT_DIR=$gitDir GIT_WORK_TREE=a/b git lfs env \
            | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected5" "$actual5"

  cd $TRASHDIR/$reponame/a/b
  envVars="$(GIT_DIR=$gitDir env | grep "^GIT" | sort)"
  expected7=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual7=$(GIT_DIR=$gitDir git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected7" "$actual7"

  cd $TRASHDIR/$reponame/a
  envVars="$(GIT_WORK_TREE=$workTree env | grep "^GIT" | sort)"
  expected8=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual8=$(GIT_WORK_TREE=$workTree git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected8" "$actual8"
)
end_test

begin_test "env with bare repo"
(
  set -e
  reponame="env-with-bare-repo"
  unset_vars
  git init --bare $reponame
  cd $reponame

  localgit=$(canonical_path "$TRASHDIR/$reponame")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf "%s\n%s\n
LocalWorkingDir=
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
" "$(git lfs version)" "$(git version)" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"

)
end_test

begin_test "env with multiple ssh remotes"
(
  set -e
  reponame="env-with-ssh"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin git@git-server.com:user/repo.git
  git remote add other git@other-git-server.com:user/repo.git

  expected='Endpoint=https://git-server.com/user/repo.git/info/lfs (auth=none)
  SSH=git@git-server.com:user/repo.git
Endpoint (other)=https://other-git-server.com/user/repo.git/info/lfs (auth=none)
  SSH=git@other-git-server.com:user/repo.git
GIT_SSH=lfs-ssh-echo'

  contains_same_elements "$expected" "$(git lfs env \
    | grep -v "^GIT_EXEC_PATH=" | grep -e "Endpoint" -e "SSH=")"
)
end_test

begin_test "env with skip download errors"
(
  set -e
  reponame="env-with-skip-dl"
  git init $reponame
  cd $reponame

  git config lfs.skipdownloaderrors 1

  localgit=$(canonical_path "$TRASHDIR/$reponame")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expectedenabled=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=true
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expectedenabled" "$actual"

  git config --unset lfs.skipdownloaderrors
  # prove it's usually off
  expecteddisabled=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expecteddisabled" "$actual"

  # now enable via env var
  envVarsEnabled=$(printf "%s" "$(GIT_LFS_SKIP_DOWNLOAD_ERRORS=1 env | grep "^GIT")")
  expectedenabled2=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=true
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVarsEnabled" "$envInitConfig")
  actual=$(GIT_LFS_SKIP_DOWNLOAD_ERRORS=1 git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expectedenabled2" "$actual"




)
end_test

begin_test "env with extra transfer methods"
(
  set -e
  reponame="env-with-transfers"
  unset_vars
  git init $reponame
  cd $reponame

  git config lfs.tustransfers true
  git config lfs.customtransfer.supertransfer.path /path/to/something

  localgit=$(canonical_path "$TRASHDIR/$reponame")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expectedenabled=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=true
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file,supertransfer
UploadTransfers=basic,lfs-standalone-file,supertransfer,tus
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expectedenabled" "$actual"

)
end_test

begin_test "env with multiple remotes and ref"
(
  set -e
  reponame="env-multiple-remotes-ref"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  endpoint="$GITSERVER/env-origin-remote.git/info/lfs (auth=none)"
  endpoint2="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$endpoint" "$endpoint2" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"
)
end_test


begin_test "env with unicode"
(
  set -e
  # This contains a Unicode apostrophe, an E with grave accent, and a Euro sign.
  # Only the middle one is representable in ISO-8859-1.
  reponame="env-d’autre-nom-très-bizarr€"
  unset_vars
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  # Set by the testsuite.
  unset LC_ALL

  endpoint="$GITSERVER/env-origin-remote.git/info/lfs (auth=none)"
  endpoint2="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(canonical_path "$TRASHDIR/$reponame")
  localgit=$(canonical_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(canonical_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(canonical_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDirs=
TempDir=%s
ConcurrentTransfers=8
TusTransfers=false
BasicTransfersOnly=false
SkipDownloadErrors=false
FetchRecentAlways=false
FetchRecentRefsDays=7
FetchRecentCommitsDays=0
FetchRecentRefsIncludeRemotes=true
PruneOffsetDays=3
PruneVerifyRemoteAlways=false
PruneRemoteName=origin
LfsStorageDir=%s
AccessDownload=none
AccessUpload=none
DownloadTransfers=basic,lfs-standalone-file
UploadTransfers=basic,lfs-standalone-file
%s
%s
' "$(git lfs version)" "$(git version)" "$endpoint" "$endpoint2" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expected" "$actual"
)
end_test
