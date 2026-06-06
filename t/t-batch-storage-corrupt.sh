#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP download of corrupt object"
(
  set -e

  reponame="batch-storage-download-corrupt-http"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  initial_sha="$(git rev-parse HEAD)"

  # This content announces to the server that it should corrupt the
  # contents of the object before returning it.
  contents="storage-download-corrupt"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  rm -f a.dat
  rm -rf .git/lfs/objects

  # We expect the server to corrupt the object's contents by inverting
  # the case of each character.
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"

  GIT_TRACE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "expected OID $contents_oid, got $inverted_contents_oid after ${#contents} bytes written" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  expected="${contents_oid:0:10} - a.dat"
  [ "$expected" = "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 1 -eq "${PIPESTATUS[0]}" ]

  grep "objects: openError: a.dat ($contents_oid) could not be checked: .*" fsck.log

  # Test again using "git pull", which should not advance HEAD after the
  # smudge filter reports an error.
  git reset --hard HEAD^

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git pull origin main 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "expected OID $contents_oid, got $inverted_contents_oid after ${#contents} bytes written" pull.log
  grep "Smudge error: Error downloading a.dat ($contents_oid)" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  [ "$initial_sha" = "$(git rev-parse HEAD)" ]

  [ -z "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 0 -eq "${PIPESTATUS[0]}" ]

  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "batch storage SSH download of corrupt object"
(
  set -e

  setup_pure_ssh

  reponame="batch-storage-download-corrupt-ssh"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  initial_sha="$(git rev-parse HEAD)"

  contents="storage-download-corrupt"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git push origin main

  assert_remote_object "$reponame" "$contents_oid" "${#contents}"

  rm -f a.dat
  rm -rf .git/lfs/objects

  # We corrupt the object's contents in the remote repository by inverting
  # the case of each character.
  inverted_contents="$(invert_case "$contents")"
  inverted_contents_oid="$(calc_oid "$inverted_contents")"

  remote_path="$(canonical_path "$REMOTEDIR/$reponame.git")"
  pushd "$remote_path"
    printf "%s" "$inverted_contents" >"$(local_object_path "$contents_oid")"
  popd

  GIT_TRACE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "expected OID $contents_oid, got $inverted_contents_oid after ${#contents} bytes written" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  expected="${contents_oid:0:10} - a.dat"
  [ "$expected" = "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 1 -eq "${PIPESTATUS[0]}" ]

  grep "objects: openError: a.dat ($contents_oid) could not be checked: .*" fsck.log

  # Test again using "git pull", which should not advance HEAD after the
  # smudge filter reports an error.
  git reset --hard HEAD^

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git pull origin main 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "expected OID $contents_oid, got $inverted_contents_oid after ${#contents} bytes written" pull.log
  grep "Smudge error: Error downloading a.dat ($contents_oid)" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  [ "$initial_sha" = "$(git rev-parse HEAD)" ]

  [ -z "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 0 -eq "${PIPESTATUS[0]}" ]

  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "batch storage custom adapter download of corrupt object"
(
  set -e

  # The "custom-transfer-" prefix indicates to the server to advertise
  # support for custom transfers.
  reponame="custom-transfer-batch-storage-download-corrupt"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  initial_sha="$(git rev-parse HEAD)"

  # This content announces to the server that it should corrupt the
  # contents of the object before returning it.
  contents="storage-download-corrupt"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git config lfs.customTransfer.testcustom.path lfstest-customadapter

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -eq "${PIPESTATUS[0]}" ]

  grep "tq: starting transfer adapter" push.log
  grep "xfer: started custom adapter process" push.log

  assert_server_object "$reponame" "$contents_oid"

  rm -f a.dat
  rm -rf .git/lfs/objects

  # We expect the server to corrupt the object's contents by inverting
  # the case of each character.
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "tq: starting transfer adapter" pull.log
  grep "xfer: started custom adapter process" pull.log

  grep "downloaded file failed checks" pull.log
  grep "has an invalid hash $inverted_contents_oid, expected $contents_oid" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  expected="${contents_oid:0:10} - a.dat"
  [ "$expected" = "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 1 -eq "${PIPESTATUS[0]}" ]

  grep "objects: openError: a.dat ($contents_oid) could not be checked: .*" fsck.log

  # Test again using "git pull", which should not advance HEAD after the
  # smudge filter reports an error.
  git reset --hard HEAD^

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git pull origin main 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "tq: starting transfer adapter" pull.log
  grep "xfer: started custom adapter process" pull.log

  grep "downloaded file failed checks" pull.log
  grep "has an invalid hash $inverted_contents_oid, expected $contents_oid" pull.log
  grep "Smudge error: Error downloading a.dat ($contents_oid)" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$inverted_contents_oid"

  [ "$initial_sha" = "$(git rev-parse HEAD)" ]

  [ -z "$(git lfs ls-files)" ]

  git lfs fsck --objects 2>&1 | tee fsck.log
  [ 0 -eq "${PIPESTATUS[0]}" ]

  grep "Git LFS fsck OK" fsck.log
)
end_test
