#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "multiple revs with same OID get pushed once"
(
  set -e

  reponame="mutliple-revs-one-oid"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="contents"
  contents_oid="$(calc_oid "$contents")"

  # Stash the contents of the file that we want to commit in .git/lfs/objects.
  object_dir="$(echo $contents_oid \
    | awk '{ print substr($0, 0, 2) "/" substr($0, 3, 2) }')"
  mkdir -p ".git/lfs/objects/$object_dir"
  printf "$contents" > ".git/lfs/objects/$object_dir/$contents_oid"

  # Create a pointer with the old "http://git-media.io" spec
  legacy_pointer="$(pointer $contents_oid 8 http://git-media.io/v/2)"
  # Create a pointer with the latest spec to create a modification, but leave
  # the OID untouched.
  latest_pointer="$(pointer $contents_oid 8)"

  # Commit the legacy pointer
  printf "$legacy_pointer" > a.dat
  git add a.dat
  git commit -m "commit legacy"

  # Commit the new pointer, causing a diff on a.dat, but leaving the OID
  # unchanged.
  printf "$latest_pointer" > a.dat
  git add a.dat
  git commit -m "commit latest"

  # Delay the push until here, so the server doesn't have a copy of the OID that
  # we're trying to push.
  git push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 8 B" push.log

  assert_server_object "$reponame" "$contents_oid"
)
end_test
