#!/usr/bin/env bash

. "test/testlib.sh"

push_fail_test() {
  local contents="$1"

  set -e

  local reponame="$(basename "$0" ".sh")-$contents"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "hi" > good.dat
  printf "$contents" > bad.dat
  git add .gitattributes good.dat bad.dat
  git commit -m "welp"

  set +e
  git push origin master
  res="$?"
  set -e

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

begin_test "push: upload file with storage 422"
(
  set -e

  push_fail_test "status-storage-422"
)
end_test

begin_test "push: upload file with storage 500"
(
  set -e

  push_fail_test "status-storage-500"
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
