#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

contents="a"
contents_oid=$(calc_oid "$contents")
b="b"
b_oid=$(calc_oid "$b")
reponame="$(basename "$0" ".sh")"

begin_test "init for fetch tests"
(
  set -e

  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log


  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  assert_local_object "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  git push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"

  # Add a file in a different branch
  git checkout -b newbranch
  printf "%s" "$b" > b.dat
  git add b.dat
  git commit -m "add b.dat"
  assert_local_object "$b_oid" 1

  git push origin newbranch
  assert_server_object "$reponame" "$b_oid"

  # These clones are used for subsequent tests
  clone_repo "$reponame" clone
  git clone --shared "$TRASHDIR/clone" "$TRASHDIR/shared"
)
end_test

begin_test "fetch"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch 2>&1 | grep "Downloading LFS objects: 100% (1/1), 1 B"
  assert_local_object "$contents_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch (shared repository)"
(
  set -e
  cd shared
  rm -rf .git/lfs/objects

  git lfs fetch 2>&1 | tee fetch.log
  ! grep "Could not scan" fetch.log
  assert_local_object "$contents_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with remote"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch origin 2>&1 | grep "Downloading LFS objects: 100% (1/1), 1 B"
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with remote and branches"
(
  set -e
  cd clone

  git checkout newbranch
  git checkout master

  rm -rf .git/lfs/objects

  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with master commit sha1"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  master_sha1=$(git rev-parse master)
  git lfs fetch origin "$master_sha1"
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with newbranch commit sha1"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  newbranch_sha1=$(git rev-parse newbranch)
  git lfs fetch origin "$newbranch_sha1"
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with include filters in gitconfig"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git config "lfs.fetchinclude" "a*"
  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid"

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with exclude filters in gitconfig"
(
  set -e

  cd clone
  git config --unset "lfs.fetchinclude"
  rm -rf .git/lfs/objects

  git config "lfs.fetchexclude" "a*"
  git lfs fetch origin master newbranch
  refute_local_object "$contents_oid"
  assert_local_object "$b_oid" 1

  git lfs fsck 2>&1 | tee fsck.log
  grep "Git LFS fsck OK" fsck.log
)
end_test

begin_test "fetch with include/exclude filters in gitconfig"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchexclude"

  git config "lfs.fetchinclude" "a*,b*"
  git config "lfs.fetchexclude" "c*,d*"
  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "c*,d*"
  git config "lfs.fetchexclude" "a*,b*"
  git lfs fetch origin master newbranch
  refute_local_object "$contents_oid"
  refute_local_object "$b_oid"
)
end_test

begin_test "fetch with include filter in cli"
(
  set -e
  cd clone
  git config --unset "lfs.fetchinclude"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects

  git lfs fetch --include="a*" origin master newbranch
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid"
)
end_test

begin_test "fetch with exclude filter in cli"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  git lfs fetch --exclude="a*" origin master newbranch
  refute_local_object "$contents_oid"
  assert_local_object "$b_oid" 1
)
end_test

begin_test "fetch with include/exclude filters in cli"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  git lfs fetch -I "a*,b*" -X "c*,d*" origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

  rm -rf .git/lfs/objects
  git lfs fetch --include="c*,d*" --exclude="a*,b*" origin master newbranch
  refute_local_object "$contents_oid"
  refute_local_object "$b_oid"
)
end_test

begin_test "fetch with include filter overriding exclude filter"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  git config lfs.fetchexclude "b*"
  git lfs fetch -I "b.dat" -X "" origin master newbranch
  assert_local_object "$b_oid" "1"
)
end_test

begin_test "fetch with missing object"
(
  set -e
  cd clone
  git config --unset lfs.fetchexclude
  rm -rf .git/lfs/objects

  delete_server_object "$reponame" "$b_oid"
  refute_server_object "$reponame" "$b_oid"

  # should return non-zero, but should also download all the other valid files too
  set +e
  git lfs fetch origin master newbranch
  fetch_exit=$?
  set -e
  [ "$fetch_exit" != "0" ]
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid"
)
end_test

begin_test "fetch-all"
(
  set -e

  reponame="fetch-all"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  NUMFILES=12
  # generate content we'll use
  for ((a=0; a < NUMFILES ; a++))
  do
    content[$a]="filecontent$a"
    oid[$a]=$(calc_oid "${content[$a]}")
  done

  echo "[
  {
    \"CommitDate\":\"$(get_date -180d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[0]}, \"Data\":\"${content[0]}\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content[1]}, \"Data\":\"${content[1]}\"}]
  },
  {
    \"NewBranch\":\"branch1\",
    \"CommitDate\":\"$(get_date -140d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[2]}, \"Data\":\"${content[2]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -100d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  },
  {
    \"NewBranch\":\"remote_branch_only\",
    \"CommitDate\":\"$(get_date -80d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Size\":${#content[4]}, \"Data\":\"${content[4]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -75d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[5]}, \"Data\":\"${content[5]}\"}]
  },
  {
    \"NewBranch\":\"tag_only\",
    \"Tags\":[\"tag1\"],
    \"CommitDate\":\"$(get_date -70d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[6]}, \"Data\":\"${content[6]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -60d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[7]}, \"Data\":\"${content[7]}\"}]
  },
  {
    \"NewBranch\":\"branch3\",
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[8]}, \"Data\":\"${content[8]}\"}]
  },
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[9]}, \"Data\":\"${content[9]}\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content[10]}, \"Data\":\"${content[10]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -30d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[11]}, \"Data\":\"${content[11]}\"}]
  }
  ]" | lfstest-testutils addcommits

  git push origin master
  git push origin branch1
  git push origin branch3
  git push origin remote_branch_only
  git push origin tag_only
  git push origin tag1
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done

  # delete remote_branch_only and make sure that objects are downloaded even
  # though not checked out to a local branch (full backup always)
  git branch -D remote_branch_only

  # delete tag_only to make sure objects are downloaded when only reachable from tag
  git branch -D tag_only

  rm -rf .git/lfs/objects

  git lfs fetch --all origin
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done

  rm -rf .git/lfs/objects

  # fetch all objects reachable from the master branch only
  git lfs fetch --all origin master
  for a in 0 1 3 5 7 9 10 11
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done
  for a in 2 4 6 8
  do
    refute_local_object "${oid[$a]}"
  done

  rm -rf .git/lfs/objects

  # fetch all objects reachable from branch1 and tag1 only
  git lfs fetch --all origin branch1 tag1
  for a in 0 1 2 3 5 6
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done
  for a in 4 7 8 9 10 11
  do
    refute_local_object "${oid[$a]}"
  done

  # Make a bare clone of the repository
  cd ..
  git clone --bare "$GITSERVER/$reponame" "$reponame-bare"
  cd "$reponame-bare"

  # Preform the same assertion as above, on the same data
  git lfs fetch --all origin
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done

  rm -rf lfs/objects

  # fetch all objects reachable from the master branch only
  git lfs fetch --all origin master
  for a in 0 1 3 5 7 9 10 11
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done
  for a in 2 4 6 8
  do
    refute_local_object "${oid[$a]}"
  done

  rm -rf lfs/objects

  # fetch all objects reachable from branch1 and tag1 only
  git lfs fetch --all origin branch1 tag1
  for a in 0 1 2 3 5 6
  do
    assert_local_object "${oid[$a]}" "${#content[$a]}"
  done
  for a in 4 7 8 9 10 11
  do
    refute_local_object "${oid[$a]}"
  done
)
end_test

begin_test "fetch: outside git repository"
(
  set +e
  git lfs fetch 2>&1 > fetch.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" fetch.log
)
end_test

begin_test "fetch with no origin remote"
(
  set -e

  reponame="fetch-no-remote"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" no-remote-clone

  clone_repo "$reponame" no-remote-repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  assert_local_object "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  git push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  grep "master -> master" push.log


  # change to the clone's working directory
  cd ../no-remote-clone

  # pull commits & lfs
  git pull 2>&1
  assert_local_object "$contents_oid" 1

  # now checkout detached HEAD so we're not tracking anything on remote
  git checkout --detach

  # delete lfs
  rm -rf .git/lfs

  # rename remote from 'origin' to 'something'
  git remote rename origin something

  # fetch should still pick this remote as in the case of no tracked remote,
  # and no origin, but only 1 remote, should pick the only one as default
  git lfs fetch
  assert_local_object "$contents_oid" 1
)
end_test

begin_test "fetch --prune"
(
  set -e

  reponame="fetch_prune"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  content_head="HEAD content"
  content_commit2="Content for commit 2 (prune)"
  content_commit1="Content for commit 1 (prune)"
  oid_head=$(calc_oid "$content_head")
  oid_commit2=$(calc_oid "$content_commit2")
  oid_commit1=$(calc_oid "$content_commit1")

  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit1}, \"Data\":\"$content_commit1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -35d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit2}, \"Data\":\"$content_commit2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -25d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all so no unpushed reason to not prune
  git push origin master

  # set no recents so max ability to prune
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentcommitsdays 0

  # delete HEAD object to prove that we still download something
  # also prune at the same time which will remove anything other than HEAD
  delete_local_object "$oid_head"
  git lfs fetch --prune
  assert_local_object "$oid_head" "${#content_head}"
  refute_local_object "$oid_commit1"
  refute_local_object "$oid_commit2"
)
end_test

begin_test "fetch raw remote url"
(
  set -e
  mkdir raw
  cd raw
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/$reponame"
  git pull origin master

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$content_oid" a.dat

  git lfs fetch "$GITSERVER/$reponame"

  # LFS object downloaded, pointer still in working directory
  assert_local_object "$contents_oid" 1
  grep "$content_oid" a.dat
)
end_test

begin_test "fetch with invalid remote"
(
  set -e
  cd repo
  git lfs fetch not-a-remote 2>&1 | tee fetch.log
  grep "Invalid remote name" fetch.log
)
end_test
