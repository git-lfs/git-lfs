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

begin_test "progress log: unknown initial size (smudge)"
(
  set -e

  reponame="progress-log-unknown-total-smudge"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: gzip" header and send a gzip-compressed response.
  contents="storage-download-encoding-gzip"
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main

  # We force the use of "git lfs smudge" rather than "git lfs filter-process"
  # because the latter downloads object files asynchronously and writes them
  # to local storage, and then only invokes the progress logging callback
  # when spooling those files back to Git in response to its second phase
  # of "smudge" filter commands where it does not specify the "can-delay"
  # flag.  As a result, the total size of each file is known during these
  # callbacks.
  #
  # By contrast, when using "git lfs smudge" the progress logging callback
  # is called as object file data is downloaded by the transfer adapter,
  # and the lack of a Content-Length header means the total size is
  # unknown until EOF is reached.
  git config --global --unset filter.lfs.process

  cd ..
  progress_log="$TRASHDIR/${reponame}-logs/progress.log"
  GIT_LFS_PROGRESS="$progress_log" GIT_TRACE=1 \
    git clone "$GITSERVER/$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  [ 0 -eq "${PIPESTATUS[0]}" ]
  cat "$progress_log"

  # Confirm we saw a gzip-encoded response, which should not include a
  # Content-Length header, so the size of the download will be initially
  # unknown.  For a more comprehensive test of gzip-encoding, see the
  # t/t-batch-storage-encoding.sh script.
  grep -c "decompressed gzipped response" clone.log

  # Confirm that the callback created by the CopyCallbackFile() method
  # of lfs.GitFilter reports the final object size in the progress log.
  #
  # Note that we will probably not see a preceding entry with a "-1"
  # field indicating an unknown total size because an EOF is usually provided
  # at the same time the final bytes are read from the HTTP download stream.
  # Once EOF is reached, the Read() method of tools.CallbackReader should pass
  # the final byte count to the callback as the total size, so the callback
  # will likely never be invoked with a negative (i.e., unknown) total size.
  grep "download 1/1 ${#contents}/${#contents} a.dat" "$progress_log"
)
end_test

begin_test "progress log: unknown initial size (fetch/pull)"
(
  set -e

  reponame="progress-log-unknown-total-pull"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: gzip" header and send a gzip-compressed response.
  contents="storage-download-encoding-gzip"
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main

  progress_log_dir="$TRASHDIR/${reponame}-logs"

  rm -rf .git/lfs/objects a.dat

  progress_log="$progress_log_dir/progress-fetch.log"
  GIT_LFS_PROGRESS="$progress_log" GIT_TRACE=1 git lfs fetch 2>&1 | tee fetch.log
  [ 0 -eq "${PIPESTATUS[0]}" ]
  cat "$progress_log"

  # Confirm we saw a gzip-encoded response, which should not include a
  # Content-Length header, so the size of the download will be initially
  # unknown.  For a more comprehensive test of gzip-encoding, see the
  # t/t-batch-storage-encoding.sh script.
  grep -c "decompressed gzipped response" fetch.log

  # Confirm that the callback created by the ensureAdapterBegun() method
  # of tq.TransferQueue reports the final object size in the progress log.
  #
  # Note that we will probably not see a preceding entry with a "-1"
  # field indicating an unknown total size because an EOF is usually provided
  # at the same time the final bytes are read from the HTTP download stream.
  # Once EOF is reached, the Read() method of tools.CallbackReader should pass
  # the final byte count to the callback as the total size, so the callback
  # will likely never be invoked with a negative (i.e., unknown) total size.
  grep "download 1/1 ${#contents}/${#contents} a.dat" "$progress_log"

  rm -rf .git/lfs/objects

  progress_log="$progress_log_dir/progress-pull.log"
  GIT_LFS_PROGRESS="$progress_log" GIT_TRACE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -eq "${PIPESTATUS[0]}" ]
  cat "$progress_log"

  # Confirm we saw a gzip-encoded response, which should not include a
  # Content-Length header, so the size of the download will be initially
  # unknown.
  grep -c "decompressed gzipped response" pull.log

  # Confirm that the callback created by the ensureAdapterBegun() method
  # of tq.TransferQueue reports the final object size in the progress log.
  #
  # Note that we will probably not see a preceding entry with a "-1"
  # field indicating an unknown total size because an EOF is usually provided
  # at the same time the final bytes are read from the HTTP download stream.
  # Once EOF is reached, the Read() method of tools.CallbackReader should pass
  # the final byte count to the callback as the total size, so the callback
  # will likely never be invoked with a negative (i.e., unknown) total size.
  grep "download 1/1 ${#contents}/${#contents} a.dat" "$progress_log"
)
end_test
