#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

declare -a expiration_types=("absolute" "relative" "both")

for typ in "${expiration_types[@]}"; do
  begin_test "expired action ($typ time)"
  (
    set -e

    reponame="expired-$typ"
    setup_remote_repo "$reponame"
    clone_repo "$reponame" "$reponame"

    contents="contents"
    contents_oid="$(calc_oid "$contents")"

    git lfs track "*.dat"
    git add .gitattributes
    git commit -m "initial commit"

    printf "%s" "$contents" > a.dat

    git add a.dat
    git commit -m "add a.dat"

    GIT_TRACE=1 git push origin main 2>&1 | tee push.log
    if [ "0" -eq "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected push to fail, didn't"
      exit 1
    fi

    refute_server_object "$reponame" "$contents_oid"
  )
  end_test
done

for typ in "${expiration_types[@]}"; do
  begin_test "ssh expired ($typ time)"
  (
    set -e

    reponame="ssh-expired-$typ"
    setup_remote_repo "$reponame"
    clone_repo "$reponame" "$reponame"

    sshurl="${GITSERVER/http:\/\//ssh://git@}/$reponame"
    git config lfs.url "$sshurl"

    contents="contents"
    contents_oid="$(calc_oid "$contents")"

    git lfs track "*.dat"
    git add .gitattributes
    git commit -m "initial commit"

    printf "%s" "$contents" > a.dat

    git add a.dat
    git commit -m "add a.dat"

    GIT_TRACE=1 git push origin main 2>&1 | tee push.log
    grep "ssh cache expired" push.log
  )
  end_test
done
