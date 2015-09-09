#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --dry-run origin master 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  [ $(wc -l < push.log) -eq 1 ]

  git lfs push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  grep "push 82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 => b.dat" push.log
  [ $(wc -l < push.log) -eq 2 ]

  git lfs push origin push-b 2>&1 | tee push.log
  grep "(1 of 2 files, 1 skipped)" push.log
)
end_test

# sets up the tests for the next few push --all tests
push_all_setup() {
  local suffix="$1"
  reponame="$(basename "$0" ".sh")-all"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "repo-all-$suffix"

  git lfs track "*.dat"
  content1="initial"
  content2="update"
  content3="branch"
  content4="tagged"
  content5="master"
  extracontent="extra"
  oid1=$(printf "$content1" | shasum -a 256 | cut -f 1 -d " ")
  oid2=$(printf "$content2" | shasum -a 256 | cut -f 1 -d " ")
  oid3=$(printf "$content3" | shasum -a 256 | cut -f 1 -d " ")
  oid4=$(printf "$content4" | shasum -a 256 | cut -f 1 -d " ")
  oid5=$(printf "$content5" | shasum -a 256 | cut -f 1 -d " ")
  extraoid=$(printf "$extracontent" | shasum -a 256 | cut -f 1 -d " ")

  echo "[
  {
    \"CommitDate\":\"$(get_date -6m)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content1},\"Data\":\"$content1\"}
    ]
  },
  {
    \"CommitDate\":\"$(get_date -5m)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content2},\"Data\":\"$content2\"}
    ]
  },
  {
    \"CommitDate\":\"$(get_date -4m)\",
    \"NewBranch\":\"branch\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content3},\"Data\":\"$content3\"}
    ]
  },
  {
    \"CommitDate\":\"$(get_date -4m)\",
    \"ParentBranches\":[\"master\"],
    \"Tags\":[\"tag\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content4},\"Data\":\"$content4\"}
    ]
  },
  {
    \"CommitDate\":\"$(get_date -2m)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content5},\"Data\":\"$content5\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#extracontent},\"Data\":\"$extracontent\"}
    ]
  }
  ]" | lfstest-testutils addcommits

  git rm file2.dat
  git commit -m "remove file2.dat"

  setup_alternate_remote "$reponame-$suffix"
}

begin_test "push --all (no ref args)"
(
  set -e

  push_all_setup "everything"

  git lfs push --dry-run --all origin 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  grep "push $oid5 => file1.dat" push.log
  grep "push $extraoid => file2.dat" push.log
  [ $(wc -l < push.log) -eq 6 ]

  git push --all origin 2>&1 | tee push.log
  grep "5 files" push.log # should be 6?
  assert_server_object "$reponame-everything" "$oid1"
  assert_server_object "$reponame-everything" "$oid2"
  assert_server_object "$reponame-everything" "$oid3"
  assert_server_object "$reponame-everything" "$oid4"
  assert_server_object "$reponame-everything" "$oid5"
  assert_server_object "$reponame-everything" "$extraoid"
)
end_test

begin_test "push --all (1 ref arg)"
(
  set -e

  push_all_setup "ref"

  git lfs push --dry-run --all origin branch 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  [ $(wc -l < push.log) -eq 3 ]

  git lfs push --all origin branch 2>&1 | tee push.log
  grep "3 files" push.log
  assert_server_object "$reponame-ref" "$oid1"
  assert_server_object "$reponame-ref" "$oid2"
  assert_server_object "$reponame-ref" "$oid3"
  refute_server_object "$reponame-ref" "$oid4"     # in master and the tag
  refute_server_object "$reponame-ref" "$oid5"
  refute_server_object "$reponame-ref" "$extraoid"
)
end_test

begin_test "push --all (multiple ref args)"
(
  set -e

  push_all_setup "multiple-refs"

  git lfs push --dry-run --all origin branch tag 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  [ $(wc -l < push.log) -eq 4 ]

  git lfs push --all origin branch tag 2>&1 | tee push.log
  grep "4 files" push.log
  assert_server_object "$reponame-multiple-refs" "$oid1"
  assert_server_object "$reponame-multiple-refs" "$oid2"
  assert_server_object "$reponame-multiple-refs" "$oid3"
  assert_server_object "$reponame-multiple-refs" "$oid4"
  refute_server_object "$reponame-multiple-refs" "$oid5"     # only in master
  refute_server_object "$reponame-multiple-refs" "$extraoid"
)
end_test

begin_test "push --all (ref with deleted files)"
(
  set -e

  push_all_setup "ref-with-deleted"

  git lfs push --dry-run --all origin master 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  grep "push $oid5 => file1.dat" push.log
  grep "push $extraoid => file2.dat" push.log
  [ $(wc -l < push.log) -eq 5 ]

  git lfs push --all origin master 2>&1 | tee push.log
  grep "5 files" push.log
  assert_server_object "$reponame-ref-with-deleted" "$oid1"
  assert_server_object "$reponame-ref-with-deleted" "$oid2"
  refute_server_object "$reponame-ref-with-deleted" "$oid3" # only in the branch
  assert_server_object "$reponame-ref-with-deleted" "$oid4"
  assert_server_object "$reponame-ref-with-deleted" "$oid5"
  assert_server_object "$reponame-ref-with-deleted" "$extraoid"
)
end_test

begin_test "push dry-run"
(
  set -e

  reponame="$(basename "$0" ".sh")-dry-run"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo-dry-run

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --dry-run origin master 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  [ $(wc -l < push.log) -eq 1 ]

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  grep "push 82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 => b.dat" push.log
  [ $(wc -l < push.log) -eq 2 ]

  # simulate remote ref
  mkdir -p .git/refs/remotes/origin
  git rev-parse HEAD > .git/refs/remotes/origin/HEAD

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  [ $(wc -l < push.log) -eq 0 ]

  git lfs push --dry-run --all origin push-b 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  grep "push 82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 => b.dat" push.log
  [ $(wc -l < push.log) -eq 2 ]
)
end_test

begin_test "push object id(s)"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo2

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    2>&1 | tee push.log
  grep "(0 of 1 files, 1 skipped)" push.log

  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 \
    2>&1 | tee push.log
  grep "(0 of 2 files, 2 skipped)" push.log
)
end_test

begin_test "push modified files"
(
  set -e

  reponame="$(basename "$0" ".sh")-modified"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  # generate content we'll use
  content1="filecontent1"
  content2="filecontent2"
  content3="filecontent3"
  content4="filecontent4"
  content5="filecontent5"
  oid1=$(printf "$content1" | shasum -a 256 | cut -f 1 -d " ")
  oid2=$(printf "$content2" | shasum -a 256 | cut -f 1 -d " ")
  oid3=$(printf "$content3" | shasum -a 256 | cut -f 1 -d " ")
  oid4=$(printf "$content4" | shasum -a 256 | cut -f 1 -d " ")
  oid5=$(printf "$content5" | shasum -a 256 | cut -f 1 -d " ")

  echo "[
  {
    \"CommitDate\":\"$(get_date -6m)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content1}, \"Data\":\"$content1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -3m)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content2}, \"Data\":\"$content2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1m)\",
    \"NewBranch\":\"other_branch\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content5}, \"Data\":\"$content5\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1m)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content3}, \"Data\":\"$content3\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content4}, \"Data\":\"$content4\"}]
  }
  ]" | lfstest-testutils addcommits

  git lfs push origin master
  git lfs push origin other_branch
  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"
  assert_server_object "$reponame" "$oid3"
  assert_server_object "$reponame" "$oid4"
  assert_server_object "$reponame" "$oid5"
)
end_test
