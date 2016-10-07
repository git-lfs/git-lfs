#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "track files using wildcard pattern with leading slash"
(
  set -e

  reponame="track-wildcard-leading-slash"
  mkdir -p "$reponame/dir"
  cd $reponame

  git init

  # Adding files before being tracked by LFS
  printf "contents" > a.dat
  printf "contents" > dir/b.dat

  git add a.dat dir/b.dat
  git commit -m "initial commit"

  # Track only in the root
  git lfs track "/*.dat"

  git add .gitattributes a.dat dir/b.dat
  sleep 1
  git commit -m "convert to LFS"

  git lfs ls-files | tee files.log

  grep "a.dat" files.log
  [ ! $(grep "dir/b.dat" files.log) ] # Subdirectories ignored

  # Add files after being tracked by LFS
  printf "contents" > c.dat
  printf "contents" > dir/d.dat

  git add c.dat dir/d.dat
  sleep 1
  git commit -m "more lfs files"

  git lfs ls-files | tee new_files.log

  grep "a.dat" new_files.log
  [ ! $(grep "dir/b.dat" new_files.log) ]
  grep "c.dat" new_files.log
  [ ! $(grep "dir/d.dat" new_files.log) ]
)
end_test

begin_test "track files using filename pattern with leading slash"
(
  set -e

  reponame="track-absolute-leading-slash"
  mkdir -p "$reponame/dir"
  cd $reponame

  git init

  # Adding files before being tracked by LFS
  printf "contents" > a.dat
  printf "contents" > dir/b.dat

  git add a.dat dir/b.dat
  sleep 1
  git commit -m "initial commit"

  # These are added by git.GetTrackedFiles
  git lfs track "/a.dat"
  git lfs track "/dir/b.dat"
  # These are added by Git's `clean` filter
  git lfs track "/c.dat"
  git lfs track "/dir/d.dat"

  git add .gitattributes a.dat dir/b.dat
  sleep 1
  git commit -m "convert to LFS"

  git lfs ls-files | tee files.log

  grep "a.dat" files.log
  grep "dir/b.dat" files.log

  # Add files after being tracked by LFS
  printf "contents" > c.dat
  printf "contents" > dir/d.dat

  git add c.dat dir/d.dat
  git commit -m "more lfs files"

  git lfs ls-files | tee new_files.log

  grep "a.dat" new_files.log
  grep "dir/b.dat" new_files.log
  grep "c.dat" new_files.log
  grep "dir/d.dat" new_files.log
)
end_test

