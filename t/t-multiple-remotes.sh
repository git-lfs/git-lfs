#!/usr/bin/env bash
# Test lfs capability to download data when blobs are stored in different
# endpoints

. "$(dirname "$0")/testlib.sh"

# This feature depends on the treeish parameter that is provided as metadata
# in git versions higher or equal than 2.27
ensure_git_version_isnt $VERSION_LOWER "2.27.0"

reponame="$(basename "$0" ".sh")"

prepare_consumer() {
  local consumer="$1"
  mkdir "$consumer"
  cd "$consumer"
  git init
  git remote add mr "file://$(urlify "$REMOTEDIR/$smain.git")"
  git remote add fr "file://$(urlify "$REMOTEDIR/$sfork.git")"
  git fetch mr
  git fetch fr
}

prepare_forks () {
  local testcase="$1"
  smain="$reponame"-"$testcase"-main-remote
  sfork="$reponame"-"$testcase"-fork-remote
  cmain="$HOME"/"$reponame"-"$testcase"-main-repo
  cfork="$HOME"/"$reponame"-"$testcase"-fork-repo
  setup_remote_repo "$smain"
  setup_remote_repo "$sfork"
  prepare_consumer "$cmain"
  git checkout -b main
  git lfs track '*.bin'
  git add --all
  git commit -m "Initial commit"
  git push -u mr main
  git push -u fr main
  #Add a .bin in main repo
  touch a.bin
  printf "1234" > a.bin
  git add --all
  git commit -m "Add Bin file"
  git push mr main
  prepare_consumer "$cfork"
}

exec_fail_git(){
  set +e
  git "$@"
  res=$?
  set -e
  if [ "$res" = "0" ]; then
    exit 1
  fi
}

begin_test "accept reset to different remote"
(
  set -e
  prepare_forks "a-reset"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect true
  git reset --hard mr/main
)
end_test

begin_test "accept pull from different remote"
(
  set -e
  prepare_forks "a-pull"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect true
  git pull mr main
)
end_test

begin_test "accept checkout different remote"
(
  set -e
  prepare_forks "a-checkout"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect true
  git checkout mr/main
)
end_test

begin_test "accept rebase different remote"
(
  set -e
  prepare_forks "a-rebase"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect true
  git rebase mr/main
)
end_test

begin_test "accept add bin file with sparsecheckout"
(
  set -e
  prepare_forks "a-sparsecheckout"
  git sparse-checkout init --no-cone
  git sparse-checkout set /.gitignore
  git checkout mr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect true
  git sparse-checkout add a.bin
)
end_test

begin_test "accept cherry-pick head different remote"
(
  set -e
  prepare_forks "a-cherrypick"
  git checkout -b main --track fr/main
  git config lfs.remote.searchall true
  git config lfs.remote.autodetect false
  git cherry-pick mr/main
)
end_test

begin_test "reject reset to different remote"
(
  set -e
  prepare_forks "r-reset"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git reset --hard mr/main
)
end_test

begin_test "reject pull from different remote"
(
  set -e
  prepare_forks "r-pull"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git pull mr main
)
end_test

begin_test "reject checkout different remote"
(
  set -e
  prepare_forks "r-checkout"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git checkout mr/main
)
end_test

begin_test "reject rebase different remote"
(
  set -e
  prepare_forks "r-rebase"
  git checkout fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git rebase mr/main
)
end_test

begin_test "reject add bin file with sparsecheckout"
(
  set -e
  prepare_forks "r-sparsecheckout"
  git sparse-checkout init --no-cone
  git sparse-checkout set /.gitignore
  git checkout mr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git sparse-checkout add a.bin
)
end_test

begin_test "reject cherry-pick head different remote"
(
  set -e
  prepare_forks "r-cherrypick"
  git checkout -b main --track fr/main
  git config lfs.remote.searchall false
  git config lfs.remote.autodetect false
  exec_fail_git cherry-pick mr/main
)
end_test
