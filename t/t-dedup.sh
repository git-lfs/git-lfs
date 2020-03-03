#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "dedup"
(
  set -e

  reponame="dedup"
  git init $reponame
  cd $reponame

  # Confirm Git LFS extensions prevent de-duplication
  git config lfs.extension.foo.clean "foo-clean %f"
  git config lfs.extension.foo.smudge "foo-smudge %f"
  git config lfs.extension.foo.priority 0

  result=$(git lfs dedup 2>&1) && true
  if ( echo $result | grep "This system does not support deduplication." ); then
    exit
  fi
  echo "$result" | grep 'This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used.'

  git config --unset lfs.extension.foo.clean
  git config --unset lfs.extension.foo.smudge
  git config --unset lfs.extension.foo.priority

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  # Delete file b and lock directory
  bOid=$(git log --patch b.dat | grep "^+oid" | cut -d ":" -f 2)
  bOid12=$(echo $bOid | cut -b 1-2)
  bOid34=$(echo $bOid | cut -b 3-4)
  rm ".git/lfs/objects/$bOid12/$bOid34/$bOid"

  # DO
  result=$(git lfs dedup 2>&1) && true

  # VERIFY: Expected
  #  Success: a.dat
  #  Success: b.dat
  echo "$result" | grep 'Success: a.dat'
  echo "$result" | grep -E 'Success:\s+b.dat|Skipped:\s+b.dat'
  # Sometimes mediafile of b.bat is restored by timing issue?
)
end_test

begin_test "dedup test"
(
  set -e

  reponame="dedup_test"
  git init $reponame
  cd $reponame

  # Confirm Git LFS extensions prevent de-duplication
  git config lfs.extension.foo.clean "foo-clean %f"
  git config lfs.extension.foo.smudge "foo-smudge %f"
  git config lfs.extension.foo.priority 0

  result=$(git lfs dedup --test 2>&1) && true
  if ( echo $result | grep "This system does not support deduplication." ); then
    exit
  fi
  echo "$result" | grep 'This platform supports file de-duplication, however, Git LFS extensions are configured and therefore de-duplication can not be used.'

  git config --unset lfs.extension.foo.clean
  git config --unset lfs.extension.foo.smudge
  git config --unset lfs.extension.foo.priority

  # DO
  result=$(git lfs dedup --test 2>&1) && true

  # Verify: This platform and repository support file de-duplication.
  echo "$result" | grep 'This platform and repository support file de-duplication.'
)
end_test

begin_test "dedup dirty workdir"
(
  set -e

  reponame="dedup_dirty_workdir"
  git init $reponame
  cd $reponame

  # Make working tree dirty.
  echo "test data" > a.dat
  git add a.dat
  git commit -m "first commit"
  echo "modify" >> a.dat

  # DO
  result=$(git lfs dedup 2>&1) && true
  if ( echo $result | grep "This system does not support deduplication." ); then
    exit
  fi

  # Verify: Working tree is dirty. Please commit or reset your change.
  echo "$result" | grep 'Working tree is dirty. Please commit or reset your change.'
)
end_test
