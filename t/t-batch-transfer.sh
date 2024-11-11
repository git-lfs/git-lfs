#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch transfer"
(
  set -e

  # This initializes a new bare git repository in test/remote.
  # These remote repositories are global to every test, so keep the names
  # unique.
  reponame1="$(basename "$0" ".sh")"
  reponame2="CAPITALLETTERS"
  reponame=$reponame1$reponame2
  setup_remote_repo "$reponame"

  # Clone the repository from the test Git server.  This is empty, and will be
  # used to test a "git pull" below. The repo is cloned to $TRASHDIR/clone
  clone_repo "$reponame" clone

  # Clone the repository again to $TRASHDIR/repo. This will be used to commit
  # and push objects.
  clone_repo "$reponame" repo

  # This executes Git LFS from the local repo that was just cloned.
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  # This is a small shell function that runs several git commands together.
  assert_pointer "main" "a.dat" "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  # This pushes to the remote repository set up at the top of the test.
  git push origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  grep "main -> main" push.log

  assert_server_object "$reponame" "$contents_oid"

  # change to the clone's working directory
  cd ../clone

  git pull origin main

  [ "a" = "$(cat a.dat)" ]

  assert_pointer "main" "a.dat" "$contents_oid" 1
)
end_test

begin_test "batch transfers occur in reverse order by size"
(
  set -e

  reponame="batch-order-test"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  small_contents="small"
  small_oid="$(calc_oid "$small_contents")"
  printf "%s" "$small_contents" > small.dat

  bigger_contents="bigger"
  bigger_oid="$(calc_oid "$bigger_contents")"
  printf "%s" "$bigger_contents" > bigger.dat

  git add *.dat
  git commit -m "add small and large objects"

  GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  batch="$(grep "{\"operation\":\"upload\"" push.log | head -1)"

  pos_small="$(substring_position "$batch" "$small_oid")"
  pos_large="$(substring_position "$batch" "$bigger_oid")"

  # Assert that the larger object shows up earlier in the batch than the
  # smaller object
  [ "$pos_large" -lt "$pos_small" ]
)
end_test

begin_test "batch transfers succeed with an empty hash algorithm"
(
  set -e

  reponame="batch-test-empty-algo"
  contents="batch-hash-algo-empty"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  printf "%s" "$contents" > special.dat
  git add .gitattributes good.dat special.dat
  git commit -m "hi"

  git push origin main
  assert_server_object "$reponame" "$(calc_oid "$contents")"
)
end_test

begin_test "batch transfers fail with an unknown hash algorithm"
(
  set -e

  reponame="batch-test-invalid-algo"
  contents="batch-hash-algo-invalid"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  printf "%s" "$contents" > special.dat
  git add .gitattributes good.dat special.dat
  git commit -m "hi"

  git push origin main 2>&1 | tee push.log
  grep 'unsupported hash algorithm' push.log
  refute_server_object "$reponame" "$(calc_oid "$contents")"
)
end_test

begin_test "batch transfers with ssh endpoint (git-lfs-authenticate)"
(
  set -e

  reponame="batch-ssh"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://git@}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main >push.log 2>&1
  [ "1" -eq "$(grep -c "exec: lfs-ssh-echo.*git-lfs-authenticate /$reponame upload" push.log)" ]
  assert_server_object "$reponame" "$(calc_oid "$contents")"
)
end_test

assert_ssh_transfer_session_counts() {
  local log="$1"
  local msg="$2"
  local min="$3"
  local max="$4"

  local count="$(grep -c "$msg" "$log")"

  [ "$max" -ge "$count" ]
  [ "$min" -le "$count" ]
}

assert_ssh_transfer_sessions() {
  local log="$1"
  local direction="$2"
  local num_objs="$3"
  local objs_per_batch="$4"

  local min_expected_start=1
  local max_expected_start=$(( num_objs > objs_per_batch ? objs_per_batch : num_objs ))
  local min_expected_end=1
  local max_expected_end="$max_expected_start"

  local expected_ctrl=1

  # On upload we currently spawn one extra control socket SSH connection
  # to run locking commands and never shut it down cleanly, so our expected
  # start counts are higher than our expected termination counts.
  if [ "upload" = "$direction" ]; then
    (( ++expected_ctrl ))
    (( ++min_expected_start ))
    (( ++max_expected_start ))
  fi

  # Versions of Git prior to 2.11.0 invoke Git LFS via the "smudge" filter
  # rather than the "process" filter, so a separate Git LFS process runs for
  # each downloaded object and spawns its own control socket SSH connection.
  if [ "download" = "$direction" ]; then
    gitversion="$(git version | cut -d" " -f3)"
    set +e
    compare_version "$gitversion" '2.11.0'
    result=$?
    set -e
    if [ "$result" -eq "$VERSION_LOWER" ]; then
      min_expected_start="$num_objs"
      max_expected_start="$num_objs"
      min_expected_end="$num_objs"
      max_expected_end="$num_objs"
      expected_ctrl="$num_objs"
    fi
  fi

  local max_expected_nonctrl=$(( max_expected_start - expected_ctrl ))

  local lines="$(grep "exec: lfs-ssh-echo.*git-lfs-transfer .*${reponame}.git $direction" "$log")"
  local ctrl_count="$(printf '%s' "$lines" | grep -c -- '-oControlMaster=yes')"
  local nonctrl_count="$(printf '%s' "$lines" | grep -c -- '-oControlMaster=no')"

  [ "$expected_ctrl" -eq "$ctrl_count" ]
  [ "$max_expected_nonctrl" -ge "$nonctrl_count" ]

  assert_ssh_transfer_session_counts "$log" 'spawning pure SSH connection' \
    "$min_expected_start" "$max_expected_start"
  assert_ssh_transfer_session_counts "$log" 'pure SSH connection successful' \
    "$min_expected_start" "$max_expected_start"
  assert_ssh_transfer_session_counts "$log" 'terminating pure SSH connection' \
    "$min_expected_end" "$max_expected_end"
}

begin_test "batch transfers with ssh endpoint (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="batch-ssh-transfer"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  contents="test"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  # On Windows we do not multiplex SSH connections by default, so we
  # enforce their use in order to match other platforms' connection counts.
  git config --global lfs.ssh.autoMultiplex true

  GIT_TRACE=1 git push origin main >push.log 2>&1
  assert_ssh_transfer_sessions 'push.log' 'upload' 1 8
  assert_remote_object "$reponame" "$(calc_oid "$contents")" "${#contents}"

  cd ..
  GIT_TRACE=1 git clone "$sshurl" "$reponame-2" 2>&1 | tee clone.log
  assert_ssh_transfer_sessions 'clone.log' 'download' 1 8

  cd "$reponame-2"
  git lfs fsck
)
end_test

begin_test "batch transfers with ssh endpoint and multiple objects (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="batch-ssh-transfer-multiple"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents1="test1"
  contents2="test2"
  contents3="test3"
  git lfs track "*.dat"
  printf "%s" "$contents1" >test1.dat
  printf "%s" "$contents2" >test2.dat
  printf "%s" "$contents3" >test3.dat
  git add .gitattributes test*.dat
  git commit -m "initial commit"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  # On Windows we do not multiplex SSH connections by default, so we
  # enforce their use in order to match other platforms' connection counts.
  git config --global lfs.ssh.autoMultiplex true

  GIT_TRACE=1 git push origin main >push.log 2>&1
  assert_ssh_transfer_sessions 'push.log' 'upload' 3 8
  assert_remote_object "$reponame" "$(calc_oid "$contents1")" "${#contents1}"
  assert_remote_object "$reponame" "$(calc_oid "$contents2")" "${#contents2}"
  assert_remote_object "$reponame" "$(calc_oid "$contents3")" "${#contents3}"

  cd ..
  GIT_TRACE=1 git clone "$sshurl" "$reponame-2" 2>&1 | tee clone.log
  assert_ssh_transfer_sessions 'clone.log' 'download' 3 8

  cd "$reponame-2"
  git lfs fsck
)
end_test

begin_test "batch transfers with ssh endpoint and multiple objects and batches (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="batch-ssh-transfer-multiple-batch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents1="test1"
  contents2="test2"
  contents3="test3"
  git lfs track "*.dat"
  printf "%s" "$contents1" >test1.dat
  printf "%s" "$contents2" >test2.dat
  printf "%s" "$contents3" >test3.dat
  git add .gitattributes test*.dat
  git commit -m "initial commit"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  # On Windows we do not multiplex SSH connections by default, so we
  # enforce their use in order to match other platforms' connection counts.
  git config --global lfs.ssh.autoMultiplex true

  # Allow no more than two objects to be transferred in each batch.
  git config --global lfs.concurrentTransfers 2

  GIT_TRACE=1 git push origin main >push.log 2>&1
  assert_ssh_transfer_sessions 'push.log' 'upload' 3 2
  assert_remote_object "$reponame" "$(calc_oid "$contents1")" "${#contents1}"
  assert_remote_object "$reponame" "$(calc_oid "$contents2")" "${#contents2}"
  assert_remote_object "$reponame" "$(calc_oid "$contents3")" "${#contents3}"

  cd ..
  GIT_TRACE=1 git clone "$sshurl" "$reponame-2" 2>&1 | tee clone.log
  assert_ssh_transfer_sessions 'clone.log' 'download' 3 2

  cd "$reponame-2"
  git lfs fsck
)
end_test
