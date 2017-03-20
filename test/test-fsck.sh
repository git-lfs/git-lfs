#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "fsck default"
(
  set -e

  reponame="fsck-default"
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  [ "Git LFS fsck OK" = "$(git lfs fsck)" ]

  aOid=$(git log --patch a.dat | grep "^+oid" | cut -d ":" -f 2)
  aOid12=$(echo $aOid | cut -b 1-2)
  aOid34=$(echo $aOid | cut -b 3-4)
  if [ "$aOid" != "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat does not match"
    exit 1
  fi

  bOid=$(git log --patch b.dat | grep "^+oid" | cut -d ":" -f 2)
  bOid12=$(echo $bOid | cut -b 1-2)
  bOid34=$(echo $bOid | cut -b 3-4)
  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi


  echo "CORRUPTION" >> .git/lfs/objects/$aOid12/$aOid34/$aOid

  moved=$(native_path "$TRASHDIR/$reponame/.git/lfs/bad")
  expected="$(printf 'Object a.dat (%s) is corrupt
Moving corrupt objects to %s' "$aOid" "$moved")"
  [ "$expected" = "$(git lfs fsck)" ]

  [ -e ".git/lfs/bad/$aOid" ]
  [ ! -e ".git/lfs/objects/$aOid12/$aOid34/$aOid" ]
  [ "$bOid" = "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]
)
end_test

begin_test "fsck dry run"
(
  set -e

  reponame="fsck-dry-run"
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  [ "Git LFS fsck OK" = "$(git lfs fsck --dry-run)" ]

  aOid=$(git log --patch a.dat | grep "^+oid" | cut -d ":" -f 2)
  aOid12=$(echo $aOid | cut -b 1-2)
  aOid34=$(echo $aOid | cut -b 3-4)
  if [ "$aOid" != "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat does not match"
    exit 1
  fi

  bOid=$(git log --patch b.dat | grep "^+oid" | cut -d ":" -f 2)
  bOid12=$(echo $bOid | cut -b 1-2)
  bOid34=$(echo $bOid | cut -b 3-4)
  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi

  echo "CORRUPTION" >> .git/lfs/objects/$aOid12/$aOid34/$aOid

  [ "Object a.dat ($aOid) is corrupt" = "$(git lfs fsck --dry-run)" ]

  if [ "$aOid" = "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat still matches match"
    exit 1
  fi

  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi
)
end_test

begin_test "fsck: outside git repository"
(
  set +e
  git lfs fsck 2>&1 > fsck.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" fsck.log
)
end_test
