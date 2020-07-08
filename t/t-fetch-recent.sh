#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

reponame="fetch-recent"

# generate content we'll use
content0="filecontent0"
content1="filecontent1"
content2="filecontent2"
content3="filecontent3"
content4="filecontent4"
content5="filecontent5"
oid0=$(calc_oid "$content0")
oid1=$(calc_oid "$content1")
oid2=$(calc_oid "$content2")
oid3=$(calc_oid "$content3")
oid4=$(calc_oid "$content4")
oid5=$(calc_oid "$content5")

begin_test "init fetch-recent"
(
  set -e

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  echo "[
  {
    \"CommitDate\":\"$(get_date -18d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content0}, \"Data\":\"$content0\"},
      {\"Filename\":\"file3.dat\",\"Size\":${#content5}, \"Data\":\"$content5\"}]
  },
  {
    \"CommitDate\":\"$(get_date -14d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content1}, \"Data\":\"$content1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"NewBranch\":\"other_branch\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content4}, \"Data\":\"$content4\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"ParentBranches\":[\"main\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content2}, \"Data\":\"$content2\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content3}, \"Data\":\"$content3\"}]
  }
  ]" | lfstest-testutils addcommits

  git push origin main
  git push origin other_branch
  assert_server_object "$reponame" "$oid0"
  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"
  assert_server_object "$reponame" "$oid3"
  assert_server_object "$reponame" "$oid4"

  # This clone is used for subsequent tests
  clone_repo "$reponame" clone
  git checkout other_branch
  git checkout main
)
end_test

begin_test "fetch-recent normal"
(
  set -e

  cd clone
  rm -rf .git/lfs/objects

  git config lfs.fetchrecentalways false
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs false
  git config lfs.fetchrecentcommitsdays 7

  # fetch normally, should just get the last state for file1/2
  git lfs fetch origin main
  assert_local_object "$oid2" "${#content2}"
  assert_local_object "$oid3" "${#content3}"
  assert_local_object "$oid5" "${#content5}"
  refute_local_object "$oid0"
  refute_local_object "$oid1"
  refute_local_object "$oid4"
)
end_test

begin_test "fetch-recent commits"
(
  set -e

  cd clone
  rm -rf .git/lfs/objects

  # now fetch recent - just commits for now
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs false
  git config lfs.fetchrecentcommitsdays 7

  git lfs fetch --recent origin
  # that should have fetched main plus previous state needed within 7 days
  # current state
  assert_local_object "$oid2" "${#content2}"
  assert_local_object "$oid3" "${#content3}"
  # previous state is the 'before' state of any commits made in last 7 days
  # ie you can check out anything in last 7 days (may have non-LFS commits in between)
  assert_local_object "$oid1" "${#content1}"
  refute_local_object "$oid0"
  refute_local_object "$oid4"
)
end_test

begin_test "fetch-recent days"
(
  set -e

  cd clone
  rm -rf .git/lfs/objects

  # now fetch other_branch as well
  git config lfs.fetchrecentrefsdays 6
  git config lfs.fetchrecentremoterefs false
  git config lfs.fetchrecentcommitsdays 7

  git lfs fetch --recent origin
  # that should have fetched main plus previous state needed within 7 days
  # current state PLUS refs within 6 days (& their commits within 7)
  assert_local_object "$oid2" "${#content2}"
  assert_local_object "$oid3" "${#content3}"
  assert_local_object "$oid1" "${#content1}"
  assert_local_object "$oid4" "${#content4}"
  # still omits oid0 since that's at best 13 days prior to other_branch tip
  refute_local_object "$oid0"
)
end_test

begin_test "fetch-recent older commits"
(
  set -e

  cd clone
  # now test that a 14 day limit picks oid0 up from other_branch
  # because other_branch was itself 5 days ago, 5+14=19 day search limit
  git config lfs.fetchrecentcommitsdays 14

  rm -rf .git/lfs/objects
  git lfs fetch --recent origin
  assert_local_object "$oid0" "${#content0}"
)
end_test

begin_test "fetch-recent remote branch"
(
  set -e

  cd "$reponame"
  # push branch & test remote branch recent
  git push origin other_branch

  cd ../clone
  git branch -D other_branch
  rm -rf .git/lfs/objects
  git config lfs.fetchrecentcommitsdays 0
  git config lfs.fetchrecentremoterefs false
  git config lfs.fetchrecentrefsdays 6

  git lfs fetch --recent origin
  # should miss #4 until we include remote branches (#1 will always be missing commitdays=0)
  assert_local_object "$oid2" "${#content2}"
  assert_local_object "$oid3" "${#content3}"
  refute_local_object "$oid1"
  refute_local_object "$oid0"
  refute_local_object "$oid4"
)
end_test

begin_test "fetch-recent remote refs"
(
  set -e

  cd clone
  rm -rf .git/lfs/objects

  # pick up just snapshot at remote ref, ie #4
  git config lfs.fetchrecentremoterefs true
  git lfs fetch --recent origin
  assert_local_object "$oid4" "${#content4}"
  refute_local_object "$oid0"
  refute_local_object "$oid1"
)
end_test
