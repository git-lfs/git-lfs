#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# push_fail_test preforms a test expecting a `git lfs push` to fail given the
# contents of a particular file contained within that push. The Git server used
# during tests has certain special cases that are triggered by finding specific
# keywords within a file (as given by the first argument).
#
# An optional second argument can be included, "msg", that assert that the
# contents "msg" was included in the output of a `git lfs push`.
push_fail_test() {
  local contents="$1"
  local msg="$2"

  set -e

  local reponame="$(basename "$0" ".sh")-$contents"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  printf "%s" "$contents" > bad.dat
  git add .gitattributes good.dat bad.dat
  git commit -m "welp"

  set +e
  git push origin main 2>&1 | tee push.log
  res="${PIPESTATUS[0]}"
  set -e

  if [ ! -z "$msg" ]; then
    grep "$msg" push.log
  fi

  refute_server_object "$reponame" "$(calc_oid "$contents")"
  if [ "$res" = "0" ]; then
    echo "push successful?"
    exit 1
  fi
}

begin_test "push: upload file with storage 403"
(
  set -e

  push_fail_test "status-storage-403"
)
end_test

begin_test "push: upload file with storage 404"
(
  set -e

  push_fail_test "status-storage-404"
)
end_test

begin_test "push: upload file with storage 410"
(
  set -e

  push_fail_test "status-storage-410"
)
end_test

begin_test "push: upload file with storage 500"
(
  set -e

  push_fail_test "status-storage-500"
)
end_test

begin_test "push: upload file with storage 503"
(
  set -e

  push_fail_test "status-storage-503" "LFS is temporarily unavailable"
)
end_test

begin_test "push: upload file with api 403"
(
  set -e

  push_fail_test "status-batch-403"
)
end_test

begin_test "push: upload file with api 404"
(
  set -e

  push_fail_test "status-batch-404"
)
end_test

begin_test "push: upload file with api 410"
(
  set -e

  push_fail_test "status-batch-410"
)
end_test

begin_test "push: upload file with api 422"
(
  set -e

  push_fail_test "status-batch-422"
)
end_test

begin_test "push: upload file with api 500"
(
  set -e

  push_fail_test "status-batch-500"
)
end_test
