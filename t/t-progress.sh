#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "progress log: file counts"
(
  set -e

  reponame="progress-log-file-counts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  echo "a" > a.dat
  echo "b" > b.dat
  echo "c" > c.dat
  echo "d" > d.dat
  echo "e" > e.dat

  progress_log_dir="$TRASHDIR/${reponame}-logs"

  progress_log="$progress_log_dir/progress-clean.log"
  GIT_LFS_PROGRESS="$progress_log" git add .gitattributes *.dat
  cat "$progress_log"

  [ 5 -eq "$(grep -c "clean 1/1" "$progress_log")" ]

  git commit -m "add files"

  progress_log="$progress_log_dir/progress-push.log"
  GIT_LFS_PROGRESS="$progress_log" git push origin main 2>&1 | tee push.log
  [ 0 -eq "${PIPESTATUS[0]}" ]
  cat "$progress_log"

  grep "Uploading LFS objects: 100% (5/5), 10 B" push.log

  grep "upload 1/5" "$progress_log"
  grep "upload 2/5" "$progress_log"
  grep "upload 3/5" "$progress_log"
  grep "upload 4/5" "$progress_log"
  grep "upload 5/5" "$progress_log"

  cd ..

  progress_log="$progress_log_dir/progress-clone.log"
  GIT_LFS_PROGRESS="$progress_log" git lfs clone "$GITSERVER/$reponame" "${reponame}-assert"
  cat "$progress_log"

  grep "download 1/5" "$progress_log"
  grep "download 2/5" "$progress_log"
  grep "download 3/5" "$progress_log"
  grep "download 4/5" "$progress_log"
  grep "download 5/5" "$progress_log"

  rm -rf "${reponame}-assert"

  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"
  cd "${reponame}-assert"

  rm -rf .git/lfs/objects

  progress_log="$progress_log_dir/progress-fetch.log"
  GIT_LFS_PROGRESS="$progress_log" git lfs fetch --all
  cat "$progress_log"

  grep "download 1/5" "$progress_log"
  grep "download 2/5" "$progress_log"
  grep "download 3/5" "$progress_log"
  grep "download 4/5" "$progress_log"
  grep "download 5/5" "$progress_log"

  progress_log="$progress_log_dir/progress-checkout.log"
  GIT_LFS_PROGRESS="$progress_log" git lfs checkout
  cat "$progress_log"

  grep "checkout 1/5" "$progress_log"
  grep "checkout 2/5" "$progress_log"
  grep "checkout 3/5" "$progress_log"
  grep "checkout 4/5" "$progress_log"
  grep "checkout 5/5" "$progress_log"
)
end_test
