#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "transfer queue rejects unknown OIDs"
(
  set -e

  reponame="unknown-oids"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="unknown-oid"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add objects"

  set +e
  git push origin master 2>&1 | tee push.log
  res="${PIPESTATUS[0]}"
  set -e

  refute_server_object "$reponame" "$(calc_oid "$contents")"
  if [ "0" -eq "$res" ]; then
    echo "push successful?"
    exit 1
  fi

  grep "\[unknown-oid\] The server returned an unknown OID." push.log
)
end_test
