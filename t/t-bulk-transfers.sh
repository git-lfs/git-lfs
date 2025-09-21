#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "bulk transfer basic upload download"
(
  set -e

  # this repo name is the indicator to the server to support bulk transfer
  reponame="test-bulk-transfer-upload-download"
  setup_remote_repo "$reponame"

  # Clone the repo
  clone_repo "$reponame" $reponame

  # Set up bulk transfer adapter configuration for this repo
  git config lfs.bulk.transfer.testbulk.path lfstest-bulkadapter
  git config lfs.bulk.transfer.testbulk.concurrent true
  git config lfs.bulk.transfer.testbulk.bulkSize 3
  git config lfs.bulk.transfer.testbulk.direction both

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "Tracking"

  # Create simple test files
  echo "bulk test data 1" > bulk1.dat
  echo "bulk test data 2 with more content" > bulk2.dat  
  echo "bulk test data 3" > bulk3.dat
  
  git add *.dat
  git commit -m "Add bulk test files"

  # Test upload with bulk transfer
  git push origin main

  # Create a fresh clone to test download
  cd ..
  clone_repo "$reponame" "${reponame}-download-test"

  # Test download with bulk transfer
  git lfs fetch --all

  # Verify the files are present
  objectlist=`find .git/lfs/objects -type f`
  [ "$(echo "$objectlist" | wc -l)" -eq 3 ]
)
end_test

begin_test "bulk transfer multiple bulks upload with concurrency"
(
  set -e

  reponame="test-bulk-transfer-sizes"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  # Set up bulk transfer adapter configuration for this repo
  git config lfs.bulk.transfer.testbulk.path lfstest-bulkadapter
  git config lfs.bulk.transfer.testbulk.concurrent true
  git config lfs.bulk.transfer.testbulk.bulkSize 2
  git config lfs.bulk.transfer.testbulk.direction both

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "Tracking"

  # Create exactly 5 files to test bulk grouping (2+2+1)
  echo "size test data 1" > size1.dat
  echo "size test data 2 with more content" > size2.dat  
  echo "size test data 3 even more content here" > size3.dat
  echo "size test data 4 with lots of additional content to make it bigger" > size4.dat
  echo "size test data 5 with the most content of all files to test different sizes" > size5.dat
  
  git add *.dat
  git commit -m "Add size test files"

  git push origin main
)
end_test

begin_test "bulk transfer multiple bulks upload no concurrency"
(
  set -e

  reponame="test-bulk-transfer-errors"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git config lfs.bulk.transfer.testbulk.path lfstest-bulkadapter
  git config lfs.bulk.transfer.testbulk.bulkSize 2
  git config lfs.bulk.transfer.testbulk.concurrent false
  git config lfs.bulk.transfer.testbulk.direction upload

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "Tracking"

  echo "error test data 1" > error1.dat
  echo "error test data 2 with some content" > error2.dat
  echo "error test data 3 with even more content" > error3.dat
  echo "error test data 4 with lots of additional content to make it bigger" > error4.dat
  
  git add *.dat
  git commit -m "Add error test files"

  git push origin main
)
end_test
