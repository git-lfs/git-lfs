#!/usr/bin/env bash

. "test/testlib.sh"

push_fail_test() {
  local contents="$1"

  set -e

  reponame="$(basename "$0" ".sh")-error"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame-$contents"

  git lfs track "*.dat"
  printf "hi" > good.dat
  printf "$contents" > bad.dat
  git add .gitattributes good.dat bad.dat
  git commit -m "welp"

  set +e
  git push origin master
  if [ "$?" = "0" ]; then
    echo "push successful?"
    exit 1
  fi

  refute_server_object "$reponame" "$(calc_oid "$contents")"

  exit 0
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
