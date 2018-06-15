#!/usr/bin/env bash

. "test/testlib.sh"

envInitConfig='git config filter.lfs.process = "git-lfs filter-process"
git config filter.lfs.smudge = "git-lfs smudge -- %f"
git config filter.lfs.clean = "git-lfs clean -- %f"'

begin_test "env with no remote"
(
  set -e
  reponame="env-no-remote"
  mkdir $reponame
  cd $reponame
  git init

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"

  endpoint="$GITSERVER/$reponame.git/info/lfs (auth=none)"
  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"

  endpoint="$GITSERVER/env-origin-remote.git/info/lfs (auth=none)"
  endpoint2="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=%s
Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  mkdir $reponame
  cd $reponame
  git init
  git remote add other "$GITSERVER/env-other-remote"

  endpoint="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf '%s
%s

Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/env-origin-remote"
  git remote add other "$GITSERVER/env-other-remote"
  git config lfs.url "http://foo/bar"

  endpoint="$GITSERVER/env-other-remote.git/info/lfs (auth=none)"
  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=%s
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=http://custom/other (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  git config lfs.concurrenttransfers 5
  git config remote.origin.lfsurl "http://custom/origin"
  git config remote.other.lfsurl "http://custom/other"

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foo/bar (auth=none)
Endpoint (other)=http://custom/other (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
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
DownloadTransfers=basic
UploadTransfers=basic
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

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")
  expected=$(printf '%s
%s

Endpoint=http://foobar:8080/ (auth=none)
LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  git init $reponame
  mkdir -p $reponame/a/b/c

  gitDir=$(native_path "$TRASHDIR/$reponame/.git")
  workTree=$(native_path "$TRASHDIR/$reponame/a/b")

  localwd=$(native_path "$TRASHDIR/$reponame/a/b")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars="$(GIT_DIR=$gitDir GIT_WORK_TREE=$workTree env | grep "^GIT" | sort)"
  expected=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  expected5=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars")
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
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
  git init --bare $reponame
  cd $reponame

  localgit=$(native_path "$TRASHDIR/$reponame")
  localgitstore=$(native_path "$TRASHDIR/$reponame")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expected=$(printf "%s\n%s\n
LocalWorkingDir=
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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

  localgit=$(native_path "$TRASHDIR/$reponame")
  localgitstore=$(native_path "$TRASHDIR/$reponame")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expectedenabled=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
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
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic
UploadTransfers=basic
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expecteddisabled" "$actual"

  # now enable via env var
  actual=$(GIT_LFS_SKIP_DOWNLOAD_ERRORS=1 git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expectedenabled" "$actual"




)
end_test

begin_test "env with extra transfer methods"
(
  set -e
  reponame="env-with-transfers"
  git init $reponame
  cd $reponame

  git config lfs.tustransfers true
  git config lfs.customtransfer.supertransfer.path /path/to/something

  localgit=$(native_path "$TRASHDIR/$reponame")
  localgitstore=$(native_path "$TRASHDIR/$reponame")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  localwd=$(native_path "$TRASHDIR/$reponame")
  localgit=$(native_path "$TRASHDIR/$reponame/.git")
  localgitstore=$(native_path "$TRASHDIR/$reponame/.git")
  lfsstorage=$(native_path "$TRASHDIR/$reponame/.git/lfs")
  localmedia=$(native_path "$TRASHDIR/$reponame/.git/lfs/objects")
  tempdir=$(native_path "$TRASHDIR/$reponame/.git/lfs/tmp")
  envVars=$(printf "%s" "$(env | grep "^GIT")")

  expectedenabled=$(printf '%s
%s

LocalWorkingDir=%s
LocalGitDir=%s
LocalGitStorageDir=%s
LocalMediaDir=%s
LocalReferenceDir=
TempDir=%s
ConcurrentTransfers=3
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
DownloadTransfers=basic,supertransfer
UploadTransfers=basic,supertransfer,tus
%s
%s
' "$(git lfs version)" "$(git version)" "$localwd" "$localgit" "$localgitstore" "$localmedia" "$tempdir" "$lfsstorage" "$envVars" "$envInitConfig")
  actual=$(git lfs env | grep -v "^GIT_EXEC_PATH=")
  contains_same_elements "$expectedenabled" "$actual"

)
end_test
