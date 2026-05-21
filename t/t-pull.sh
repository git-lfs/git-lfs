#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "pull"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" clone
  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")
  contents2="A"
  contents2_oid=$(calc_oid "$contents2")
  contents3="dir"
  contents3_oid=$(calc_oid "$contents3")

  mkdir -p dir1 dir2/dir3/dir4
  echo "*.log" > .gitignore
  printf "%s" "$contents" > a.dat
  printf "%s" "$contents2" > á.dat
  printf "%s" "$contents3" > dir1/dir.dat
  printf "%s" "$contents3" > dir2/dir3/dir4/dir.dat
  git add .
  git commit -m "add files" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "6 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]

  assert_pointer "main" "a.dat" "$contents_oid" 1
  assert_pointer "main" "á.dat" "$contents2_oid" 1
  assert_pointer "main" "dir1/dir.dat" "$contents3_oid" 3

  refute_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents2_oid"
  refute_server_object "$reponame" "$contents3_oid"

  echo "initial push"
  git push origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (3/3), 5 B" push.log
  grep "main -> main" push.log

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # change to the clone's working directory
  cd ../clone

  echo "normal pull"
  git config branch.main.remote origin
  git config branch.main.merge refs/heads/main
  git pull 2>&1

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]

  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull"
  rm -rf a.dat á.dat dir1 dir2 # removing files makes the status dirty
  rm -rf .git/lfs/objects
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status
  git lfs fsck

  echo "lfs pull with remote"
  rm -rf a.dat á.dat dir1 dir2
  rm -rf .git/lfs/objects
  git lfs pull origin
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status
  git lfs fsck

  echo "lfs pull with local storage"
  rm -rf a.dat á.dat dir1 dir2
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_clean_status

  echo "test pre-existing directories"
  rm -rf dir1/dir.dat dir2/dir3/dir4
  git lfs pull
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_clean_status

  echo "lfs pull with include/exclude filters in gitconfig"
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs pull
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs pull
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "lfs pull with include/exclude filters in command line"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs pull --include="a*"
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git lfs pull --exclude="a*"
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "resetting to test status"
  git reset --hard
  assert_clean_status

  echo "lfs pull clean status"
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull with -I"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir1/dir.dat" dir2
  git lfs pull -I "a.*,dir1/dir.*,dir2/**"
  [ "a" = "$(cat a.dat)" ]
  [ ! -e "á.dat" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  refute_local_object "$contents2_oid"
  assert_local_object "$contents3_oid" 3
  assert_clean_worktree_with_exceptions '\\303\\241\.dat'

  git lfs pull -I "*.dat"
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents2_oid" 1
  assert_clean_status

  echo "lfs pull with empty file"
  touch empty.dat
  git add empty.dat
  git commit -m 'empty'
  git lfs pull
  [ -z "$(cat empty.dat)" ]
  assert_clean_status

  echo "resetting to test status"
  git reset --hard HEAD^
  assert_clean_status

  echo "lfs pull in subdir"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir1/dir.dat" dir2
  pushd dir1
    git lfs pull
  popd
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull in subdir with -I"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir1/dir.dat" dir2
  pushd dir1
    git lfs pull -I "á.*,dir1/dir.dat,dir2/**"
  popd
  [ ! -e a.dat ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir1/dir.dat")" ]
  [ "dir" = "$(cat "dir2/dir3/dir4/dir.dat")" ]
  refute_local_object "$contents_oid"
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_worktree_with_exceptions "a\.dat"

  pushd dir1
    git lfs pull -I "*.dat"
  popd
  [ "a" = "$(cat a.dat)" ]
  assert_local_object "$contents_oid" 1
  assert_clean_status
)
end_test

begin_test "pull: skip directory file conflicts"
(
  set -e

  reponame="pull-skip-dir-file-conflicts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir -p dir1 dir2/dir3/dir4
  printf "%s" "$contents" >dir1/a.dat
  printf "%s" "$contents" >dir2/dir3/dir4/a.dat

  git add .gitattributes dir1 dir2
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  rm -rf dir1 dir2/dir3
  touch dir1 dir2/dir3

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"dir1/a\.dat": not a directory' pull.log
  grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log

  assert_local_object "$contents_oid" 1

  [ -f "dir1" ]
  [ -f "dir2/dir3" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd dir2
    git lfs pull 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected pull to succeed ..."
      exit 1
    fi
    grep '"dir1/a\.dat": not a directory' pull.log
    grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log
  popd

  assert_local_object "$contents_oid" 1

  [ -f "dir1" ]
  [ -f "dir2/dir3" ]
  assert_clean_index
)
end_test

begin_test "pull: skip directory symlink conflicts"
(
  set -e

  skip_if_symlinks_unsupported

  reponame="pull-skip-dir-symlink-conflicts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir -p dir1 dir2/dir3/dir4
  printf "%s" "$contents" >dir1/a.dat
  printf "%s" "$contents" >dir2/dir3/dir4/a.dat

  git add .gitattributes dir1 dir2
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  # test with symlinks to directories
  rm -rf dir1 dir2/dir3 ../link*
  mkdir ../link1 ../link2
  ln -s ../link1 dir1
  ln -s ../../link2 dir2/dir3

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"dir1/a\.dat": not a directory' pull.log
  grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log
  [ 0 -eq "$(grep -c "is beyond a symbolic link" pull.log)" ]

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ ! -e "../link1/a.dat" ]
  [ ! -e "../link2/dir4" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  rm -rf dir1 dir2/dir3
  mkdir link1 link2
  ln -s link1 dir1
  ln -s ../link2 dir2/dir3

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"dir1/a\.dat": not a directory' pull.log
  grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log
  [ 0 -eq "$(grep -c "is beyond a symbolic link" pull.log)" ]

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ ! -e "link1/a.dat" ]
  [ ! -e "link2/dir4" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd dir2
    git lfs pull 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected pull to succeed ..."
      exit 1
    fi
    grep '"dir1/a\.dat": not a directory' pull.log
    grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log
    [ 0 -eq "$(grep -c "is beyond a symbolic link" pull.log)" ]
  popd

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ ! -e "link1/a.dat" ]
  [ ! -e "link2/dir4" ]
  assert_clean_index

  # test with symlink to file and dangling symlink
  rm -rf .git/lfs/objects

  rm -rf dir1 dir2/dir3 ../link*
  touch ../link1
  ln -s ../link1 dir1
  ln -s ../../link2 dir2/dir3

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"dir1/a\.dat": not a directory' pull.log
  grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ -f "../link1" ]
  [ ! -e "../link2" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  rm -rf dir1 dir2/dir3 link*
  touch link1
  ln -s link1 dir1
  ln -s ../link2 dir2/dir3

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"dir1/a\.dat": not a directory' pull.log
  grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ -f "link1" ]
  [ ! -e "link2" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd dir2
    git lfs pull 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected pull to succeed ..."
      exit 1
    fi
    grep '"dir1/a\.dat": not a directory' pull.log
    grep '"dir2/dir3/dir4/a\.dat": not a directory' pull.log
  popd

  assert_local_object "$contents_oid" 1

  [ -L "dir1" ]
  [ -L "dir2/dir3" ]
  [ -f "link1" ]
  [ ! -e "link2" ]
  assert_clean_index
)
end_test

begin_test "pull: skip file symlink conflicts"
(
  set -e

  skip_if_symlinks_unsupported

  reponame="pull-skip-file-symlink-conflicts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir -p dir1/dir2/dir3
  printf "%s" "$contents" >a.dat
  printf "%s" "$contents" >dir1/dir2/dir3/a.dat

  git add .gitattributes a.dat dir1
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  # test with symlinks to pointer files
  rm -rf a.dat dir1/dir2/dir3/a.dat ../link*
  contents_pointer="$(git cat-file -p ":a.dat")"
  printf "%s" "$contents_pointer" >../link1
  printf "%s" "$contents_pointer" >../link2
  ln -s ../link1 a.dat
  ln -s ../../../../link2 dir1/dir2/dir3/a.dat

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"a\.dat": not a regular file' pull.log
  grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -f "../link1" ]
  [ "$contents_pointer" = "$(cat ../link1)" ]
  [ -f "../link2" ]
  [ "$contents_pointer" = "$(cat ../link2)" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  rm -rf a.dat dir1/dir2/dir3/a.dat link*
  printf "%s" "$contents_pointer" >link1
  printf "%s" "$contents_pointer" >link2
  ln -s link1 a.dat
  ln -s ../../../link2 dir1/dir2/dir3/a.dat

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"a\.dat": not a regular file' pull.log
  grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -f "link1" ]
  [ "$contents_pointer" = "$(cat link1)" ]
  [ -f "link2" ]
  [ "$contents_pointer" = "$(cat link2)" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd dir1/dir2
    git lfs pull 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected pull to succeed ..."
      exit 1
    fi
    grep '"a\.dat": not a regular file' pull.log
    grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log
  popd

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -f "link1" ]
  [ "$contents_pointer" = "$(cat link1)" ]
  [ -f "link2" ]
  [ "$contents_pointer" = "$(cat link2)" ]
  assert_clean_index

  # test with symlink to directory and dangling symlink
  rm -rf .git/lfs/objects

  rm -rf a.dat dir1/dir2/dir3/a.dat ../link*
  mkdir ../link1
  ln -s ../link1 a.dat
  ln -s ../../../../link2 dir1/dir2/dir3/a.dat

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"a\.dat": not a regular file' pull.log
  grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -d "../link1" ]
  [ ! -e "../link2" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  rm -rf a.dat dir1/dir2/dir3/a.dat link*
  mkdir link1
  ln -s link1 a.dat
  ln -s ../../../link2 dir1/dir2/dir3/a.dat

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep '"a\.dat": not a regular file' pull.log
  grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -d "link1" ]
  [ ! -e "link2" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd dir1/dir2
    git lfs pull 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected pull to succeed ..."
      exit 1
    fi
    grep '"a\.dat": not a regular file' pull.log
    grep '"dir1/dir2/dir3/a\.dat": not a regular file' pull.log
  popd

  assert_local_object "$contents_oid" 1

  [ -L "a.dat" ]
  [ -L "dir1/dir2/dir3/a.dat" ]
  [ -d "link1" ]
  [ ! -e "link2" ]
  assert_clean_index
)
end_test

# This test applies to case-preserving but case-insensitive filesystems,
# such as APFS and NTFS when in their default configurations.
# On case-sensitive filesystems this test has no particular value and
# should always pass.
begin_test "pull: skip case-based symlink conflicts"
(
  set -e

  skip_if_symlinks_unsupported

  # Only test with Git version 2.20.0 as it introduced detection of
  # case-insensitive filesystems to the "git clone" command, which the
  # test depends on to determine the filesystem type.
  ensure_git_version_isnt "$VERSION_LOWER" "2.20.0"

  reponame="pull-skip-case-symlink-conflicts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  mkdir dir1
  ln -s ../link1 A.dat
  ln -s ../../link2 dir1/a.dat
  ln -s ../link3 DIR3
  ln -s ../../link4 dir1/dir2

  git add A.dat dir1 DIR3
  git commit -m "initial commit"

  rm A.dat dir1/* DIR3

  echo "*.dat filter=lfs diff=lfs merge=lfs -text" >.gitattributes

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir dir3 dir1/DIR2
  printf "%s" "$contents" >a.dat
  printf "%s" "$contents" >dir1/A.dat
  printf "%s" "$contents" >dir3/a.dat
  printf "%s" "$contents" >dir1/DIR2/a.dat

  git -c core.ignoreCase=false add .gitattributes a.dat dir1/A.dat \
    dir3/a.dat dir1/DIR2/a.dat
  git commit -m "case-conflicting commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
  collision="$(grep -c "collided" clone.log)" || true

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  rm -rf *.dat dir1 *3 ../link*
  mkdir ../link3 ../link4

  git lfs pull

  assert_local_object "$contents_oid" 1

  [ -f "a.dat" ]
  [ "$contents" = "$(cat "a.dat")" ]
  [ -f "dir1/A.dat" ]
  [ "$contents" = "$(cat "dir1/A.dat")" ]
  [ -f "dir3/a.dat" ]
  [ "$contents" = "$(cat "dir3/a.dat")" ]
  [ -f "dir1/DIR2/a.dat" ]
  [ "$contents" = "$(cat "dir1/DIR2/a.dat")" ]
  [ ! -e "../link1" ]
  [ ! -e "../link2" ]
  [ ! -e "../link3/a.dat" ]
  [ ! -e "../link4/a.dat" ]
  assert_clean_index

  rm -rf a.dat dir1/A.dat dir3 dir1/DIR2
  git checkout -- A.dat dir1/a.dat DIR3 dir1/dir2

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  if [ "$collision" -gt "0" ]; then
    # case-insensitive filesystem
    grep '"a\.dat": not a regular file' pull.log
    grep '"dir1/A\.dat": not a regular file' pull.log
    grep '"dir3/a\.dat": not a directory' pull.log
    grep '"dir1/DIR2/a\.dat": not a directory' pull.log
    [ 0 -eq "$(grep -c "is beyond a symbolic link" pull.log)" ]
  fi

  if [ "$collision" -eq "0" ]; then
    # case-sensitive filesystem
    [ -f "a.dat" ]
    [ "$contents" = "$(cat "a.dat")" ]
    [ -f "dir1/A.dat" ]
    [ "$contents" = "$(cat "dir1/A.dat")" ]
    [ -f "dir3/a.dat" ]
    [ "$contents" = "$(cat "dir3/a.dat")" ]
    [ -f "dir1/DIR2/a.dat" ]
    [ "$contents" = "$(cat "dir1/DIR2/a.dat")" ]
  else
    # case-insensitive filesystem
    [ -L "a.dat" ]
    [ -L "dir1/A.dat" ]
    [ -L "dir3" ]
    [ -L "dir1/DIR2" ]
  fi
  [ ! -e "../link1" ]
  [ ! -e "../link2" ]
  [ ! -e "../link3/a.dat" ]
  [ ! -e "../link4/a.dat" ]
  assert_clean_index
)
end_test

begin_test "pull: skip changed files"
(
  set -e

  reponame="pull-skip-changed-files"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  contents_new="$contents +extra"
  printf "%s" "$contents_new" >a.dat

  git lfs pull
  assert_local_object "$contents_oid" 1

  [ "$contents_new" = "$(cat a.dat)" ]
  assert_clean_index

  rm a.dat
  mkdir a.dat

  rm -rf .git/lfs/objects
  git lfs pull
  assert_local_object "$contents_oid" 1

  [ -d "a.dat" ]
  assert_clean_index

  rm -rf .git/lfs/objects

  pushd a.dat
    git lfs pull
  popd

  assert_local_object "$contents_oid" 1

  [ -d "a.dat" ]
  assert_clean_index
)
end_test

begin_test "pull: break hard links to existing files"
(
  set -e

  reponame="pull-break-file-hardlinks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir -p dir1/dir2/dir3
  printf "%s" "$contents" >a.dat
  printf "%s" "$contents" >dir1/dir2/dir3/a.dat

  git add .gitattributes a.dat dir1
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$contents_oid" 1

  rm -f a.dat dir1/dir2/dir3/a.dat ../link
  pointer="$(git cat-file -p ":a.dat")"
  echo "$pointer" >../link
  ln ../link a.dat
  ln ../link dir1/dir2/dir3/a.dat

  git lfs pull
  assert_local_object "$contents_oid" 1

  [ "$contents" = "$(cat a.dat)" ]
  [ "$contents" = "$(cat dir1/dir2/dir3/a.dat)" ]
  [ "$pointer" = "$(cat ../link)" ]
  assert_clean_status

  rm a.dat dir1/dir2/dir3/a.dat
  ln ../link a.dat
  ln ../link dir1/dir2/dir3/a.dat

  rm -rf .git/lfs/objects

  pushd dir1/dir2
    git lfs pull
  popd

  assert_local_object "$contents_oid" 1

  [ "$contents" = "$(cat a.dat)" ]
  [ "$contents" = "$(cat dir1/dir2/dir3/a.dat)" ]
  [ "$pointer" = "$(cat ../link)" ]
  assert_clean_status
)
end_test

begin_test "pull without clean filter"
(
  set -e

  GIT_LFS_SKIP_SMUDGE=1 git clone $GITSERVER/t-pull no-clean
  cd no-clean
  git lfs uninstall
  git config --list > config.txt
  grep "filter.lfs.clean" config.txt && {
    echo "clean filter still configured:"
    cat config.txt
    exit 1
  }

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid"

  git lfs pull | tee pull.txt
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep "Git LFS is not installed" pull.txt
  echo "pulled!"

  # LFS object downloaded, pointer unchanged
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid" 1
)
end_test

begin_test "pull with raw remote url"
(
  set -e
  mkdir raw
  cd raw
  git init
  git lfs install --local --skip-smudge

  git remote add origin $GITSERVER/t-pull
  git pull origin main

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  git lfs pull "$GITSERVER/t-pull"
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ 0 -eq "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with multiple remotes"
(
  set -e
  mkdir multiple
  cd multiple
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git remote add bad-remote "invalid-url"
  git pull origin main

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  # pull should default to origin instead of bad-remote
  git lfs pull
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ 0 -eq "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with invalid insteadof"
(
  set -e
  mkdir insteadof
  cd insteadof
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git pull origin main

  # set insteadOf to rewrite the href of downloading LFS object.
  git config url."$GITSERVER/storage/invalid".insteadOf "$GITSERVER/storage/"
  # Enable href rewriting explicitly.
  git config lfs.transfer.enablehrefrewrite true

  set +e
  git lfs pull > pull.log 2>&1
  res=$?

  set -e
  [ "$res" = "2" ]

  # check rewritten href is used to download LFS object.
  grep "LFS: Repository or object not found: $GITSERVER/storage/invalid" pull.log

  # lfs-pull succeed after unsetting enableHrefRewrite config
  git config --unset lfs.transfer.enablehrefrewrite
  git lfs pull
)
end_test

begin_test "pull with merge conflict"
(
  set -e
  git init pull-merge-conflict
  cd pull-merge-conflict

  git lfs track "*.bin"
  git add .
  git commit -m 'gitattributes'
  printf abc > abc.bin
  git add .
  git commit -m 'abc'

  git checkout -b def
  printf def > abc.bin
  git add .
  git commit -m 'def'

  git checkout main
  printf ghi > abc.bin
  git add .
  git commit -m 'ghi'

  # This will exit nonzero because of the merge conflict.
  GIT_LFS_SKIP_SMUDGE=1 git merge def || true
  git lfs pull > pull.log 2>&1
  [ ! -s pull.log ]
)
end_test

begin_test "pull: with missing object"
(
  set -e

  # this clone is setup in the first test in this file
  cd clone
  rm -rf .git/lfs/objects
  rm a.dat

  contents_oid=$(calc_oid "a")
  reponame="$(basename "$0" ".sh")"
  delete_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents_oid"

  # should return non-zero, but should also download all the other valid files too
  git config branch.main.remote origin
  git config branch.main.merge refs/heads/main
  git lfs pull 2>&1 | tee pull.log
  pull_exit="${PIPESTATUS[0]}"
  [ "$pull_exit" != "0" ]

  grep "$contents_oid does not exist" pull.log

  contents2_oid=$(calc_oid "A")
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents_oid"
)
end_test

begin_test "pull: outside git repository"
(
  set +e
  git lfs pull 2>&1 > pull.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a Git repository" pull.log
)
end_test

begin_test "pull: read-only directory"
(
  set -e

  skip_if_root_or_admin "$test_description"

  reponame="pull-read-only"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="a"
  contents_oid=$(calc_oid "$contents")
  mkdir dir
  printf "%s" "$contents" > dir/a.bin

  git add .gitattributes dir/a.bin
  git commit -m "add dir/a.bin"

  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  rm dir/a.bin
  delete_local_object "$contents_oid"

  if [ "$IS_WINDOWS" -eq 1 ]; then
    icacls dir /inheritance:r
    icacls dir /grant:r Everyone:R
  else
    chmod a-w dir
  fi
  git lfs pull 2>&1 | tee pull.log
  # Note that although the pull command should log an error, at present
  # we still expect a zero exit code.
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  assert_local_object "$contents_oid" 1

  [ ! -e dir/a.bin ]

  grep 'could not check out "dir/a.bin"' pull.log
  grep 'could not create working directory file' pull.log
  grep 'permission denied' pull.log
)
end_test

begin_test "pull: read-only file"
(
  set -e

  reponame="pull-locked"
  filename="a.txt"

  setup_remote_repo_with_file "$reponame" "$filename"

  pushd "$TRASHDIR" > /dev/null
    GIT_LFS_SKIP_SMUDGE=1 clone_repo "$reponame" "${reponame}-assert"

    chmod a-w "$filename"

    refute_file_writeable "$filename"
    assert_pointer "refs/heads/main" "$filename" "$(calc_oid "$filename\n")" 6

    git lfs pull

    refute_file_writeable "$filename"
    [ "$filename" = "$(cat "$filename")" ]
  popd > /dev/null
)
end_test

begin_test "pull with empty file doesn't modify mtime"
(
  set -e
  git init pull-empty-file
  cd pull-empty-file

  git lfs track "*.bin"
  git add .
  git commit -m 'gitattributes'
  printf abc > abc.bin
  git add .
  git commit -m 'abc'

  touch foo.bin
  lfstest-nanomtime foo.bin >foo.mtime

  # This isn't necessary, but it takes a few cycles to make sure that our
  # timestamp changes.
  git add foo.bin
  git commit -m 'foo'

  git lfs pull
  lfstest-nanomtime foo.bin >foo.mtime2
  diff -u foo.mtime foo.mtime2
)
end_test

begin_test "pull: bare repository"
(
  set -e

  reponame="pull-bare"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >a.dat

  # The "git lfs pull" command should never check out files in a bare
  # repository, either into a directory within the repository or one
  # outside it.  To verify this, we add a Git LFS pointer file whose path
  # inside the repository is one which, if it were instead treated as an
  # absolute filesystem path, corresponds to a writable directory.
  # The "git lfs pull" command should not check out files into either
  # this external directory or the bare repository.
  external_dir="$TRASHDIR/${reponame}-external"
  internal_dir="$(printf "%s" "$external_dir" | sed 's/^\/*//')"
  mkdir -p "$internal_dir"
  printf "%s" "$contents" >"$internal_dir/a.dat"

  git add .gitattributes a.dat "$internal_dir/a.dat"
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$contents_oid"

  cd ..
  git clone --bare "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  [ ! -e lfs ]
  refute_local_object "$contents_oid"

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi

  # When Git version 2.42.0 or higher is available, the "git lfs pull"
  # command will use the "git ls-files" command rather than the
  # "git ls-tree" command to list files.  By default a bare repository
  # lacks an index, so we expect no Git LFS objects to be fetched when
  # "git ls-files" is used because Git v2.42.0 or higher is available.
  gitversion="$(git version | cut -d" " -f3)"
  set +e
  compare_version "$gitversion" '2.42.0'
  result=$?
  set -e
  if [ "$result" -eq "$VERSION_LOWER" ]; then
    grep "Downloading LFS objects" pull.log

    assert_local_object "$contents_oid" 1
  else
    grep -q "Downloading LFS objects" pull.log && exit 1

    refute_local_object "$contents_oid"
  fi

  [ ! -e "a.dat" ]
  [ ! -e "$internal_dir/a.dat" ]
  [ ! -e "$external_dir/a.dat" ]

  rm -rf lfs/objects
  refute_local_object "$contents_oid"

  # When Git version 2.42.0 or higher is available, the "git lfs pull"
  # command will use the "git ls-files" command rather than the
  # "git ls-tree" command to list files.  By default a bare repository
  # lacks an index, so we expect no Git LFS objects to be fetched when
  # "git ls-files" is used because Git v2.42.0 or higher is available.
  #
  # Therefore to verify that the "git lfs pull" command never checks out
  # files in a bare repository, we first populate the index with Git LFS
  # pointer files and then retry the command.
  contents_git_oid="$(git ls-tree HEAD a.dat | awk '{ print $3 }')"
  git update-index --add --cacheinfo 100644 "$contents_git_oid" a.dat
  git update-index --add --cacheinfo 100644 "$contents_git_oid" "$internal_dir/a.dat"

  # When Git version 2.42.0 or higher is available, the "git lfs pull"
  # command will use the "git ls-files" command rather than the
  # "git ls-tree" command to list files, and does so by passing an
  # "attr:filter=lfs" pathspec to the "git ls-files" command so it only
  # lists files which match that filter attribute.
  #
  # In a bare repository, however, the "git ls-files" command will not read
  # attributes from ".gitattributes" files in the index, so by default it
  # will not list any Git LFS pointer files even if those files and the
  # corresponding ".gitattributes" files have been added to the index and
  # the pointer files would otherwise match the "attr:filter=lfs" pathspec.
  #
  # Therefore, instead of adding the ".gitattributes" file to the index, we
  # copy it to "info/attributes" so that the pathspec filter will match our
  # pointer file index entries and they will be listed by the "git ls-files"
  # command.  This allows us to verify that with Git v2.42.0 or higher, the
  # "git lfs pull" command will fetch the objects for these pointer files
  # in the index when the command is run in a bare repository.
  #
  # Note that with older versions of Git, the "git lfs pull" command will
  # use the "git ls-tree" command to list the files in the tree referenced
  # by HEAD.  The Git LFS objects for any well-formed pointer files found in
  # that list will then be fetched (unless local copies already exist),
  # regardless of whether the pointer files actually match a "filter=lfs"
  # attribute in any ".gitattributes" file in the index, the tree
  # referenced by HEAD, or the current work tree.
  if [ "$result" -ne "$VERSION_LOWER" ]; then
    mkdir -p info
    git show HEAD:.gitattributes >info/attributes
  fi

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep "Downloading LFS objects" pull.log

  assert_local_object "$contents_oid" 1

  [ ! -e "a.dat" ]
  [ ! -e "$internal_dir/a.dat" ]
  [ ! -e "$external_dir/a.dat" ]
)
end_test

begin_test "pull with partial clone and sparse checkout and index"
(
  set -e

  # Only test with Git version 2.25.0 as it introduced the
  # "git sparse-checkout" command.  (Note that this test also requires
  # that the "git rev-list" command support the "tree:0" filter, which
  # was introduced with Git version 2.20.0.)
  ensure_git_version_isnt "$VERSION_LOWER" "2.25.0"

  reponame="pull-sparse"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents1="a"
  contents1_oid=$(calc_oid "$contents1")
  contents2="b"
  contents2_oid=$(calc_oid "$contents2")
  contents3="c"
  contents3_oid=$(calc_oid "$contents3")

  mkdir in-dir out-dir
  printf "%s" "$contents1" >a.dat
  printf "%s" "$contents2" >in-dir/b.dat
  printf "%s" "$contents3" >out-dir/c.dat
  git add .
  git commit -m "add files"

  git push origin main

  assert_server_object "$reponame" "$contents1_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # Create a partial clone with a cone-mode sparse checkout of one directory
  # and a sparse index, which is important because otherwise the "git ls-files"
  # command ignores the --sparse option and lists all Git LFS files.
  cd ..
  git clone --filter=tree:0 --depth=1 --no-checkout \
    "$GITSERVER/$reponame" "${reponame}-partial"

  cd "${reponame}-partial"
  git sparse-checkout init --cone --sparse-index
  git sparse-checkout set "in-dir"
  git checkout main

  [ -d "in-dir" ]
  [ ! -e "out-dir" ]

  assert_local_object "$contents1_oid" 1
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents3_oid"

  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi

  # When Git version 2.42.0 or higher is available, the "git lfs pull"
  # command will use the "git ls-files" command rather than the
  # "git ls-tree" command to list files.  Git v2.42.0 introduced support
  # in the "git ls-files" command for the "objecttype" format option and
  # so Git LFS can use this command to avoid pulling objects outside
  # the sparse cone.  Otherwise, all Git LFS objects will be pulled.
  gitversion="$(git version | cut -d" " -f3)"
  set +e
  compare_version "$gitversion" '2.42.0'
  result=$?
  set -e
  if [ "$result" -eq "$VERSION_LOWER" ]; then
    grep "Downloading LFS objects" pull.log

    [ -f "out-dir/c.dat" ]
    [ "$contents3" = "$(cat "out-dir/c.dat")" ]

    assert_local_object "$contents3_oid" 1
  else
    grep -q "Downloading LFS objects" pull.log && exit 1

    [ ! -e "out-dir" ]

    refute_local_object "$contents3_oid"
  fi
)
end_test

begin_test "pull: pointer extension"
(
  set -e

  reponame="pull-pointer-extension"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  setup_case_inverter_extension

  contents="abc"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  mkdir dir1
  printf "%s" "$contents" >dir1/abc.dat

  git add .gitattributes dir1
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$inverted_contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$inverted_contents_oid"

  setup_case_inverter_extension

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  git lfs pull

  assert_local_object "$inverted_contents_oid" 3

  [ "$contents" = "$(cat "dir1/abc.dat")" ]
  grep "smudge: dir1/abc.dat" "$LFSTEST_EXT_LOG"

  rm -rf .git/lfs/objects

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  mkdir dir2

  pushd dir2
    git lfs pull
  popd

  [ "$contents" = "$(cat "dir1/abc.dat")" ]
  grep "smudge: dir1/abc.dat" "$LFSTEST_EXT_LOG"

  assert_local_object "$inverted_contents_oid" 3
)
end_test
