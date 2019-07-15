#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# sets up the repos for the first few push tests. The passed argument is the
# name of the repo to setup. The resuling repo will have a local file tracked
# with LFS and committed, but not yet pushed to the remote
push_repo_setup() {
  reponame="$1"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"
}

begin_test "push with good ref"
(
  set -e
  push_repo_setup "push-master-branch-required"

  git lfs push origin master
)
end_test

begin_test "push with tracked ref"
(
  set -e

  push_repo_setup "push-tracked-branch-required"

  git config push.default upstream
  git config branch.master.merge refs/heads/tracked
  git lfs push origin master
)
end_test

begin_test "push with bad ref"
(
  set -e
  push_repo_setup "push-other-branch-required"

  git lfs push origin master 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo "expected command to fail"
    exit 1
  fi

  grep 'batch response: Expected ref "refs/heads/other", got "refs/heads/master"' push.log
)
end_test

begin_test "push with given remote, configured pushRemote"
(
  set -e
  push_repo_setup "push-given-and-config"

  git remote add bad-remote "invalid-url"

  git config branch.master.pushRemote bad-remote

  git lfs push --all origin
)
end_test

begin_test "push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" true

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --dry-run origin master 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  [ $(grep -c "^push " push.log) -eq 1 ]

  git lfs push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  grep "push 4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 => a.dat" push.log
  grep "push 82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 => b.dat" push.log
  [ $(grep -c "^push " < push.log) -eq 2 ]

  # simulate remote ref
  mkdir -p .git/refs/remotes/origin
  git rev-parse HEAD > .git/refs/remotes/origin/HEAD

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  [ $(grep -c "^push " push.log) -eq 0 ]

  rm -rf .git/refs/remotes

  git lfs push origin push-b 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (2/2), 14 B" push.log
)
end_test

# sets up the tests for the next few push --all tests
push_all_setup() {
  suffix="$1"
  reponame="$(basename "$0" ".sh")-all"
  content1="initial"
  content2="update"
  content3="branch"
  content4="tagged"
  content5="master"
  extracontent="extra"
  oid1=$(calc_oid "$content1")
  oid2=$(calc_oid "$content2")
  oid3=$(calc_oid "$content3")
  oid4=$(calc_oid "$content4")
  oid5=$(calc_oid "$content5")
  extraoid=$(calc_oid "$extracontent")

  # if the local repo exists, it has already been bootstrapped
  [ -d "push-all" ] && exit 0

  clone_repo "$reponame" "push-all"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" true
  git lfs track "*.dat"

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

  # simulate remote ref
  mkdir -p .git/refs/remotes/origin
  git rev-parse HEAD > .git/refs/remotes/origin/HEAD

  setup_alternate_remote "$reponame-$suffix"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame-$suffix").locksverify" true
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
  [ $(grep -c "^push " < push.log) -eq 6 ]

  git push --all origin 2>&1 | tee push.log
  [ $(grep -c "Uploading LFS objects: 100% (6/6), 36 B" push.log) -eq 1 ]
  assert_server_object "$reponame-$suffix" "$oid1"
  assert_server_object "$reponame-$suffix" "$oid2"
  assert_server_object "$reponame-$suffix" "$oid3"
  assert_server_object "$reponame-$suffix" "$oid4"
  assert_server_object "$reponame-$suffix" "$oid5"
  assert_server_object "$reponame-$suffix" "$extraoid"

  echo "push while missing old objects locally"
  setup_alternate_remote "$reponame-$suffix-2"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame-$suffix-2").locksverify" true

  git lfs push --object-id origin $oid1
  assert_server_object "$reponame-$suffix-2" "$oid1"
  refute_server_object "$reponame-$suffix-2" "$oid2"
  refute_server_object "$reponame-$suffix-2" "$oid3"
  refute_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
  rm ".git/lfs/objects/${oid1:0:2}/${oid1:2:2}/$oid1"

  echo "dry run missing local object that exists on server"
  git lfs push --dry-run --all origin 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  grep "push $oid5 => file1.dat" push.log
  grep "push $extraoid => file2.dat" push.log
  [ $(grep -c "^push " push.log) -eq 6 ]

  git push --all origin 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (6/6), 36 B" push.log
  assert_server_object "$reponame-$suffix-2" "$oid2"
  assert_server_object "$reponame-$suffix-2" "$oid3"
  assert_server_object "$reponame-$suffix-2" "$oid4"
  assert_server_object "$reponame-$suffix-2" "$oid5"
  assert_server_object "$reponame-$suffix-2" "$extraoid"
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
  [ $(grep -c "^push " < push.log) -eq 3 ]

  git lfs push --all origin branch 2>&1 | tee push.log
  grep "3 files" push.log
  assert_server_object "$reponame-$suffix" "$oid1"
  assert_server_object "$reponame-$suffix" "$oid2"
  assert_server_object "$reponame-$suffix" "$oid3"
  refute_server_object "$reponame-$suffix" "$oid4"     # in master and the tag
  refute_server_object "$reponame-$suffix" "$oid5"
  refute_server_object "$reponame-$suffix" "$extraoid"

  echo "push while missing old objects locally"
  setup_alternate_remote "$reponame-$suffix-2"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame-$suffix-2").locksverify" true
  git lfs push --object-id origin $oid1
  assert_server_object "$reponame-$suffix-2" "$oid1"
  refute_server_object "$reponame-$suffix-2" "$oid2"
  refute_server_object "$reponame-$suffix-2" "$oid3"
  refute_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
  rm ".git/lfs/objects/${oid1:0:2}/${oid1:2:2}/$oid1"

  # dry run doesn't change
  git lfs push --dry-run --all origin branch 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  [ $(grep -c "^push " push.log) -eq 3 ]

  git push --all origin branch 2>&1 | tee push.log
  grep "5 files, 1 skipped" push.log # should be 5?
  assert_server_object "$reponame-$suffix-2" "$oid2"
  assert_server_object "$reponame-$suffix-2" "$oid3"
  refute_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
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
  [ $(grep -c "^push " push.log) -eq 4 ]

  git lfs push --all origin branch tag 2>&1 | tee push.log
  grep "4 files" push.log
  assert_server_object "$reponame-$suffix" "$oid1"
  assert_server_object "$reponame-$suffix" "$oid2"
  assert_server_object "$reponame-$suffix" "$oid3"
  assert_server_object "$reponame-$suffix" "$oid4"
  refute_server_object "$reponame-$suffix" "$oid5"     # only in master
  refute_server_object "$reponame-$suffix" "$extraoid"

  echo "push while missing old objects locally"
  setup_alternate_remote "$reponame-$suffix-2"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame-$suffix-2").locksverify" true
  git lfs push --object-id origin $oid1
  assert_server_object "$reponame-$suffix-2" "$oid1"
  refute_server_object "$reponame-$suffix-2" "$oid2"
  refute_server_object "$reponame-$suffix-2" "$oid3"
  refute_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
  rm ".git/lfs/objects/${oid1:0:2}/${oid1:2:2}/$oid1"

  # dry run doesn't change
  git lfs push --dry-run --all origin branch tag 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid3 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  [ $(grep -c "^push " push.log) -eq 3 ]

  git push --all origin branch tag 2>&1 | tee push.log
  grep "5 files, 1 skipped" push.log # should be 5?
  assert_server_object "$reponame-$suffix-2" "$oid2"
  assert_server_object "$reponame-$suffix-2" "$oid3"
  assert_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
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
  [ $(grep -c "^push " push.log) -eq 5 ]

  git lfs push --all origin master 2>&1 | tee push.log
  grep "5 files" push.log
  assert_server_object "$reponame-$suffix" "$oid1"
  assert_server_object "$reponame-$suffix" "$oid2"
  refute_server_object "$reponame-$suffix" "$oid3" # only in the branch
  assert_server_object "$reponame-$suffix" "$oid4"
  assert_server_object "$reponame-$suffix" "$oid5"
  assert_server_object "$reponame-$suffix" "$extraoid"

  echo "push while missing old objects locally"
  setup_alternate_remote "$reponame-$suffix-2"
  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame-$suffix-2").locksverify" true
  git lfs push --object-id origin $oid1
  assert_server_object "$reponame-$suffix-2" "$oid1"
  refute_server_object "$reponame-$suffix-2" "$oid2"
  refute_server_object "$reponame-$suffix-2" "$oid3"
  refute_server_object "$reponame-$suffix-2" "$oid4"
  refute_server_object "$reponame-$suffix-2" "$oid5"
  refute_server_object "$reponame-$suffix-2" "$extraoid"
  rm ".git/lfs/objects/${oid1:0:2}/${oid1:2:2}/$oid1"

  # dry run doesn't change
  git lfs push --dry-run --all origin master 2>&1 | tee push.log
  grep "push $oid1 => file1.dat" push.log
  grep "push $oid2 => file1.dat" push.log
  grep "push $oid4 => file1.dat" push.log
  grep "push $oid5 => file1.dat" push.log
  grep "push $extraoid => file2.dat" push.log
  [ $(grep -c "^push " push.log) -eq 5 ]

  git push --all origin master 2>&1 | tee push.log
  grep "5 files, 1 skipped" push.log # should be 5?
  assert_server_object "$reponame-$suffix-2" "$oid2"
  refute_server_object "$reponame-$suffix-2" "$oid3"
  assert_server_object "$reponame-$suffix-2" "$oid4"
  assert_server_object "$reponame-$suffix-2" "$oid5"
  assert_server_object "$reponame-$suffix-2" "$extraoid"
)
end_test

begin_test "push object id(s)"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo2

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" true

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log

  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 \
    2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (2/2), 14 B" push.log
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
  oid1=$(calc_oid "$content1")
  oid2=$(calc_oid "$content2")
  oid3=$(calc_oid "$content3")
  oid4=$(calc_oid "$content4")
  oid5=$(calc_oid "$content5")

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

begin_test "push with invalid remote"
(
  set -e
  cd repo
  git lfs push not-a-remote 2>&1 | tee push.log
  grep "Invalid remote name" push.log
)
end_test

begin_test "push ambiguous branch name"
(
  set -e

  reponame="$(basename "$0" ".sh")-ambiguous-branch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"


  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  NUMFILES=5
  # generate content we'll use
  for ((a=0; a < NUMFILES ; a++))
  do
    content[$a]="filecontent$a"
    oid[$a]=$(calc_oid "${content[$a]}")
  done

  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[0]}, \"Data\":\"${content[0]}\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content[1]}, \"Data\":\"${content[1]}\"}]
  },
  {
    \"NewBranch\":\"ambiguous\",
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[2]}, \"Data\":\"${content[2]}\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[4]}, \"Data\":\"${content[4]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # create tag with same name as branch
  git tag ambiguous

  # lfs push master, should work
  git lfs push origin master

  # push ambiguous, does not fail since lfs scans git with sha, not ref name
  git lfs push origin ambiguous
)
end_test

begin_test "push (retry with expired actions)"
(
  set -e

  reponame="push_retry_expired_action"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  contents="return-expired-action"
  contents_oid="$(calc_oid "$contents")"
  contents_size="$(printf "%s" "$contents" | wc -c | awk '{ print $1 }')"
  printf "%s" "$contents" > a.dat
  git add .gitattributes a.dat

  git commit -m "add a.dat, .gitattributes" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  GIT_TRACE=1 git push origin master 2>&1 | tee push.log

  expected="enqueue retry #1 for \"$contents_oid\" (size: $contents_size): LFS: tq: action \"upload\" expires at"

  grep "$expected" push.log
  grep "Uploading LFS objects: 100% (1/1), 21 B" push.log
)
end_test

begin_test "push to raw remote url"
(
  set -e

  setup_remote_repo "push-raw"
  mkdir push-raw
  cd push-raw
  git init

  git lfs track "*.dat"

  contents="raw"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > raw.dat
  git add raw.dat .gitattributes
  git commit -m "add" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 raw.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  refute_server_object push-raw "$contents_oid"

  git lfs push $GITSERVER/push-raw master

  assert_server_object push-raw "$contents_oid"
)
end_test

begin_test "push (with invalid object size)"
(
  set -e

  reponame="push-invalid-object-size"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  contents="return-invalid-size"
  printf "%s" "$contents" > a.dat

  git add a.dat .gitattributes
  git commit -m "add a.dat, .gitattributes" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  set +e
  git push origin master 2>&1 2> push.log
  res="$?"
  set -e

  grep "invalid size (got: -1)" push.log
  [ "0" -eq "$(grep -c "panic" push.log)" ]
  [ "0" -ne "$res" ]

  refute_server_object "$reponame" "$(calc_oid "$contents")"
)
end_test

begin_test "push with deprecated _links"
(
  set -e

  reponame="$(basename "$0" ".sh")-deprecated"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="send-deprecated-links"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git add a.dat
  git commit -m "add a.dat"

  git push origin master

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "push with invalid pushInsteadof"
(
  set -e

  push_repo_setup "push-invalid-pushinsteadof"

  # set pushInsteadOf to rewrite the href of uploading LFS object.
  git config url."$GITSERVER/storage/invalid".pushInsteadOf "$GITSERVER/storage/"
  # Enable href rewriting explicitly.
  git config lfs.transfer.enablehrefrewrite true

  set +e
  git lfs push origin master > push.log 2>&1
  res=$?

  set -e
  [ "$res" = "2" ]

  # check rewritten href is used to upload LFS object.
  grep "LFS: Authorization error: $GITSERVER/storage/invalid" push.log

  # lfs-push succeed after unsetting enableHrefRewrite config
  git config --unset lfs.transfer.enablehrefrewrite
  git lfs push origin master
)
end_test

begin_test 'push with data the server already has'
(
  set -e

  reponame="push-server-data"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="abc123"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git add a.dat
  git commit -m "add a.dat"

  git push origin master

  assert_server_object "$reponame" "$contents_oid"

  git checkout -b side

  # Use a different file name for the second file; otherwise, this test
  # unexpectedly passes with the old code, since we fail to notice that the
  # object we run through the clean filter is not the object we wanted.
  contents2="def456"
  contents2_oid="$(calc_oid "$contents2")"
  printf "%s" "$contents2" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  # We remove the original object. The server already has this.
  delete_local_object "$contents_oid"

  # We use the URL so that we cannot take advantage of the existing "origin/*"
  # refs that we know the server must have. We will traverse the entire history
  # for this push, and we should not fail because the server already has the
  # object we deleted above.
  git push "$(git config remote.origin.url)" side

  assert_server_object "$reponame" "$contents2_oid"
)
end_test
