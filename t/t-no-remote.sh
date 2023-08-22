#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

reponame_src="$(basename "$0" ".sh")-src"
reponame_dst="$(basename "$0" ".sh")-dst"

begin_test "fetch lfs-tracked file despite no remote"
(
  set -e

  # First, a repo with an lfs-tracked file we can fetch from
  setup_remote_repo_with_file "$reponame_src" "test_file.dat"

  # Grab the rev for `git archive` later
  echo $(pwd)
  rev=$(git rev-parse HEAD) 
  cd ..
  
  # Initialize a bare repo we can fetch into
  mkdir $reponame_dst
  cd $reponame_dst
  git init . --bare
  echo $(pwd)
  git fetch "$GITSERVER/$reponame_src" refs/heads/main:refs/heads/main
  git archive $rev -o archive.out

  # Verify archive contains our file
  grep "test_file.dat" archive.out
)
end_test
