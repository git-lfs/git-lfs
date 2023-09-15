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

reponame_src_A="$(basename "$0" ".sh")-src-A"
reponame_src_B="$(basename "$0" ".sh")-src-B"
reponame_dst_2="$(basename "$0" ".sh")-dst-2"
begin_test "fallback ignored when remote present"
(
  set -e

  # Initialize 2 repos with different files
  setup_remote_repo_with_file "$reponame_src_A" "test_file_A.dat"
  rev=$(git rev-parse HEAD)
  cd ..
  setup_remote_repo_with_file "$reponame_src_B" "test_file_B.dat"
  cd ..

  mkdir $reponame_dst_2
  cd $reponame_dst_2
  git init . --bare
  echo $(pwd)
  # This part is subtle
  # Add repo A as a remote and fetch from it
  # But then fetch from repo B. This points FETCH_HEAD to repo B
  # We're testing that git-lfs will ignore FETCH_HEAD, since FETCH_HEAD is
  # a fallback, only used when no remote is set
  git remote add origin "$GITSERVER/$reponame_src_A"
  git fetch
  git fetch "$GITSERVER/$reponame_src_B" refs/heads/main:refs/heads/main
  git archive $rev -o archive.out

  # Verify archive contains file from second repo, but not first repo
  grep "test_file_A.dat" archive.out
  grep -v "test_file_B.dat" archive.out
)
end_test

