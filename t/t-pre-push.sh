#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "pre-push with good ref"
(
  set -e
  reponame="pre-push-main-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"

  # for some reason, using 'tee' and $PIPESTATUS does not work here
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 > push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"
)
end_test

begin_test "pre-push with tracked ref"
(
  set -e
  reponame="pre-push-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/tracked"

  # for some reason, using 'tee' and $PIPESTATUS does not work here
  echo "refs/heads/main main refs/heads/tracked 0000000000000000000000000000000000000000" |
    git lfs pre-push origin main 2>&1 > push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/tracked"
)
end_test

begin_test "pre-push with bad ref"
(
  set -e
  reponame="pre-push-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/other"

  # for some reason, using 'tee' and $PIPESTATUS does not work here
  set +e
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2> push.log
  pushcode=$?
  set -e

  if [ "0" -eq "$pushcode" ]; then
    echo "expected command to fail"
    exit 1
  fi

  grep 'Expected ref "refs/heads/other", got "refs/heads/main"' push.log

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/other"
)
end_test

begin_test "pre-push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  git config "lfs.$(repo_endpoint $GITSERVER $reponame).locksverify" true

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" |
    tee push.log
  # no output if nothing to do
  [ "$(du -k push.log | cut -f 1)" == "0" ]

  git lfs track "*.dat"
  echo "hi" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 3 B" push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4
)
end_test

begin_test "pre-push dry-run"
(
  set -e

  reponame="$(basename "$0" ".sh")-dry-run"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-dry-run
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  git config "lfs.$(repo_endpoint $GITSERVER $reponame).locksverify" true

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" |
    tee push.log

  [ "" = "$(cat push.log)" ]

  git lfs track "*.dat"
  echo "dry" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show

  refute_server_object "$reponame" 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" |
    tee push.log
  grep "push 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b => hi.dat" push.log
  cat push.log
  [ `wc -l < push.log` = 1 ]

  refute_server_object "$reponame" 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b
)
end_test

begin_test "pre-push skip-push"
(
  set -e

  reponame="$(basename "$0" ".sh")-skip-push"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-skip-push
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  git config "lfs.$(repo_endpoint $GITSERVER $reponame).locksverify" true

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    GIT_LFS_SKIP_PUSH=true git lfs pre-push origin "$GITSERVER/$reponame" |
    tee push.log

  [ "" = "$(cat push.log)" ]

  git lfs track "*.dat"
  echo "dry" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show

  refute_server_object "$reponame" 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    GIT_LFS_SKIP_PUSH=true git lfs pre-push origin "$GITSERVER/$reponame" |
    tee push.log

  [ "" = "$(cat push.log)" ]

  refute_server_object "$reponame" 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b
)
end_test

begin_test "pre-push 307 redirects"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-307
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  # relative redirect
  git config remote.origin.lfsurl "$GITSERVER/redirect307/rel/$reponame.git/info/lfs"

  git lfs track "*.dat"
  echo "hi" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/redirect307/rel/$reponame.git/info/lfs" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 3 B" push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4

  # absolute redirect
  git config remote.origin.lfsurl "$GITSERVER/redirect307/abs/$reponame.git/info/lfs"

  echo "hi" > hi2.dat
  git add hi2.dat
  git commit -m "add hi2.dat"
  git show

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/redirect307/abs/$reponame.git/info/lfs" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 3 B" push.log
)
end_test

begin_test "pre-push with existing object"
(
  set -e

  reponame="pre-push-existing-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  echo "existing" > existing.dat
  git add existing.dat
  git commit -m "add existing dat"

  git lfs track "*.dat"
  echo "new" > new.dat
  git add new.dat
  git add .gitattributes
  git commit -m "add new file through git lfs"

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 4 B" push.log

  # now the file exists
  assert_server_object "$reponame" 7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c
)
end_test

begin_test "pre-push with existing object (untracked)"
(
  set -e

  reponame="pre-push-existing-object-untracked"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  echo "$(pointer "7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c" 4)" > new.dat
  git add new.dat
  git commit -m "add new pointer"
  mkdir -p .git/lfs/objects/7a/a7
  echo "new" > .git/lfs/objects/7a/a7/7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 4 B" push.log

  # now the file exists
  assert_server_object "$reponame" 7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c
)
end_test

begin_test "pre-push reject missing object"
(
  set -e

  reponame="pre-push-reject-missing-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  oid="7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c"

  echo "$(pointer "$oid" 4)" > new.dat
  git add new.dat
  git commit -m "add new pointer"

  # assert that push fails
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  if [ "2" -ne "${PIPESTATUS[1]}" ]; then
    echo >&2 "fatal: expected 'git lfs pre-push origin $GITSERVER/$reponame' to fail ..."
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (missing) new.dat ($oid)" push.log

  refute_server_object "$reponame" "$oid"
)
end_test

begin_test "pre-push allow missing object (found on server)"
(
  # should permit push if files missing locally but are on server, shouldn't
  # require client to have every file (prune)
  set -e

  reponame="pre-push-allow-missing-object-found-on-server"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="common data"
  contents_oid=$(calc_oid "$contents")
  git lfs track "*.dat"
  printf "%s" "$contents" > common1.dat
  git add common1.dat
  git add .gitattributes
  git commit -m "add first file"

  # push file to the git lfs server
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "Uploading LFS objects: 100% (1/1), 11 B" push.log

  # now the file exists
  assert_server_object "$reponame" "$contents_oid"

  # create another commit referencing same oid, then delete local data & push
  printf "%s" "$contents" > common2.dat
  git add common2.dat
  git commit -m "add second file, same content"
  rm -rf .git/lfs/objects
  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  # make sure there were no errors reported
  [ -z "$(grep -i 'Error' push.log)" ]
)
end_test

begin_test "pre-push allow missing object (lfs.allowincompletepush true)"
(
  set -e

  reponame="pre-push-allow-missing-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush true

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  if [ "0" -ne "${PIPESTATUS[1]}" ]; then
    echo >&2 "fatal: expected 'git lfs pre-push origin $GITSERVER/$reponame' to succeed ..."
    exit 1
  fi

  grep "LFS upload missing objects" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "pre-push reject missing object (lfs.allowincompletepush default)"
(
  set -e

  reponame="pre-push-reject-missing-object-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    GIT_TRACE=1 git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  if [ "2" -ne "${PIPESTATUS[1]}" ]; then
    echo >&2 "fatal: expected 'git lfs pre-push origin $GITSERVER/$reponame' to fail ..."
    exit 1
  fi

  grep "tq: stopping batched queue, object \"$missing_oid\" missing locally and on remote" push.log
  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "pre-push multiple branches"
(
  set -e

  reponame="$(basename "$0" ".sh")-multiple-branches"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"


  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  NUMFILES=6
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
    \"NewBranch\":\"branch1\",
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file2.dat\",\"Size\":${#content[2]}, \"Data\":\"${content[2]}\"}]
  },
  {
    \"ParentBranches\":[\"main\"],
    \"NewBranch\":\"branch2\",
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  },
  {
    \"ParentBranches\":[\"main\"],
    \"NewBranch\":\"branch3\",
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[4]}, \"Data\":\"${content[4]}\"}]
  },
  {
    \"ParentBranches\":[\"main\"],
    \"NewBranch\":\"branch4\",
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[5]}, \"Data\":\"${content[5]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # make sure when called via git push all branches are updated
  git push origin main branch1 branch2 branch3 branch4
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done

)
end_test

begin_test "pre-push with bad remote"
(
  set -e

  cd repo

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" |
    git lfs pre-push not-a-remote "$GITSERVER/$reponame" 2>&1 |
    tee pre-push.log
  grep "Invalid remote name" pre-push.log
)
end_test

begin_test "pre-push unfetched deleted remote branch & server GC"
(
  # point of this is to simulate the case where the local cache of the remote
  # branch state contains a branch which has actually been deleted on the remote,
  # the client just doesn't know yet (hasn't done 'git fetch origin --prune')
  # If the server GC'd the objects that deleted branch contained, but they were
  # referenced by a branch being pushed (earlier commit), push might assume it
  # doesn't have to push it, but it does. Tests that we check the real remote refs
  # before making an assumption about the diff we need to push
  set -e

  reponame="$(basename "$0" ".sh")-server-deleted-branch-gc"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"


  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  NUMFILES=4
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
    \"NewBranch\":\"branch-to-delete\",
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[2]}, \"Data\":\"${content[2]}\"}]
  },
  {
    \"NewBranch\":\"branch-to-push-after\",
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # push only the first 2 branches
  git push origin main branch-to-delete
  for ((a=0; a < 3 ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done
  # confirm we haven't pushed the last one yet
  refute_server_object "$reponame" "${oid[3]}"
  # copy the cached remote ref for the branch we're going to delete remotely
  cp .git/refs/remotes/origin/branch-to-delete branch-to-delete.ref
  # now delete the branch on the server
  git push origin --delete branch-to-delete
  # remove the OID in it, as if GC'd
  delete_server_object "$reponame" "${oid[2]}"
  refute_server_object "$reponame" "${oid[2]}"
  # Now put the cached remote ref back, as if someone else had deleted it but
  # we hadn't done git fetch --prune yet
  mv branch-to-delete.ref .git/refs/remotes/origin/branch-to-delete
  # Confirm that local cache of remote branch is back
  git branch -r 2>&1 | tee branch-r.log
  grep "origin/branch-to-delete" branch-r.log
  # Now push later branch which should now need to re-push previous commits LFS too
  git push origin branch-to-push-after
  # all objects should now be there even though cached remote branch claimed it already had file3.dat
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done

)
end_test

begin_test "pre-push delete branch"
(
  set -e

  reponame="$(basename "$0" ".sh")-delete-branch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"


  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  NUMFILES=4
  # generate content we'll use
  for ((a=0; a < NUMFILES ; a++))
  do
    content[$a]="filecontent$a"
    oid[$a]=$(calc_oid "${content[$a]}")
  done

  echo "[
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[0]}, \"Data\":\"${content[0]}\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content[1]}, \"Data\":\"${content[1]}\"}]
  },
  {
    \"NewBranch\":\"branch-to-delete\",
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[2]}, \"Data\":\"${content[2]}\"}]
  },
  {
    \"ParentBranches\":[\"main\"],
    \"CommitDate\":\"$(get_date -0d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all branches
  git push origin main branch-to-delete
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done

  # deleting a branch with git push should not fail
  # (requires correct special casing of "(delete) 0000000000.." in hook)
  git push origin --delete branch-to-delete
)
end_test

begin_test "pre-push with our lock"
(
  set -e

  reponame="pre_push_owned_locks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "%s" "$contents" > locked.dat
  git add locked.dat
  git commit -m "add locked.dat"

  git push origin main

  git lfs lock --json "locked.dat" | tee lock.log

  id=$(assert_lock lock.log locked.dat)
  assert_server_lock $id

  printf "authorized changes" >> locked.dat
  git add locked.dat
  git commit -m "add unauthorized changes"

  GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log
  grep "Consider unlocking your own locked files" push.log
  grep "* locked.dat" push.log

  assert_server_lock "$id"
)
end_test

begin_test "pre-push with their lock on lfs file"
(
  set -e

  reponame="pre_push_unowned_lock"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"

  # any lock path with "theirs" is returned as "their" lock by /locks/verify
  printf "%s" "$contents" > locked_theirs.dat
  git add locked_theirs.dat
  git commit -m "add locked_theirs.dat"

  git push origin main

  git lfs lock --json "locked_theirs.dat" | tee lock.log
  id=$(assert_lock lock.log locked_theirs.dat)
  assert_server_lock $id

  pushd "$TRASHDIR" >/dev/null
    clone_repo "$reponame" "$reponame-assert"
    git config lfs.locksverify true

    printf "unauthorized changes" >> locked_theirs.dat
    git add locked_theirs.dat
    # --no-verify is used to avoid the pre-commit hook which is not under test
    git commit --no-verify -m "add unauthorized changes"

    git push origin main 2>&1 | tee push.log
    res="${PIPESTATUS[0]}"
    if [ "0" -eq "$res" ]; then
      echo "push should fail"
      exit 1
    fi

    grep "Unable to push locked files" push.log
    grep "* locked_theirs.dat - Git LFS Tests" push.log

    grep "Cannot update locked files." push.log
    refute_server_object "$reponame" "$(calc_oid_file locked_theirs.dat)"
  popd >/dev/null
)
end_test

begin_test "pre-push with their lock on non-lfs lockable file"
(
  set -e

  reponame="pre_push_unowned_lock_not_lfs"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  echo "*.dat lockable" > .gitattributes
  git add .gitattributes
  git commit -m "initial commit"

  # any lock path with "theirs" is returned as "their" lock by /locks/verify
  echo "hi" > readme.txt
  echo "tiny" > tiny_locked_theirs.dat
  git help > large_locked_theirs.dat
  git add readme.txt tiny_locked_theirs.dat large_locked_theirs.dat
  git commit -m "add initial files"

  git push origin main

  git lfs lock --json "tiny_locked_theirs.dat" | tee lock.log
  id=$(assert_lock lock.log tiny_locked_theirs.dat)
  assert_server_lock $id

  git lfs lock --json "large_locked_theirs.dat" | tee lock.log
  id=$(assert_lock lock.log large_locked_theirs.dat)
  assert_server_lock $id

  pushd "$TRASHDIR" >/dev/null
    clone_repo "$reponame" "$reponame-assert"
    git config lfs.locksverify true

    git lfs update # manually add pre-push hook, since lfs clean hook is not used
    echo "other changes" >> readme.txt
    echo "unauthorized changes" >> large_locked_theirs.dat
    echo "unauthorized changes" >> tiny_locked_theirs.dat
    # --no-verify is used to avoid the pre-commit hook which is not under test
    git commit --no-verify -am "add unauthorized changes"

    git push origin main 2>&1 | tee push.log
    res="${PIPESTATUS[0]}"
    if [ "0" -eq "$res" ]; then
      echo "push should fail"
      exit 1
    fi

    grep "Unable to push locked files" push.log
    grep "* large_locked_theirs.dat - Git LFS Tests" push.log
    grep "* tiny_locked_theirs.dat - Git LFS Tests" push.log
    grep "Cannot update locked files." push.log

    refute_server_object "$reponame" "$(calc_oid_file large_locked_theirs.dat)"
    refute_server_object "$reponame" "$(calc_oid_file tiny_locked_theirs.dat)"
  popd >/dev/null
)
end_test

begin_test "pre-push locks verify 5xx with verification enabled"
(
  set -e

  reponame="lock-enabled-verify-5xx"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" true

  git push origin main 2>&1 | tee push.log
  grep "\"origin\" does not support the Git LFS locking API" push.log
  grep "git config lfs.$endpoint.locksverify false" push.log

  refute_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push disable locks verify on exact url"
(
  set -e

  reponame="lock-disabled-verify-5xx"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" false

  git push origin main 2>&1 | tee push.log
  [ "0" -eq "$(grep -c "\"origin\" does not support the Git LFS locking API" push.log)" ]

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push disable locks verify on partial url"
(
  set -e

  reponame="lock-disabled-verify-5xx-partial"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$server/$repo"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" false

  git push origin main 2>&1 | tee push.log
  [ "0" -eq "$(grep -c "\"origin\" does not support the Git LFS locking API" push.log)" ]

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push locks verify 403 with good ref"
(
  set -e

  reponame="lock-verify-main-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$GITSERVER/$reponame.git.locksverify" true
  git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid" "refs/heads/main"
)
end_test

begin_test "pre-push locks verify 403 with good tracked ref"
(
  set -e

  reponame="lock-verify-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config push.default upstream
  git config branch.main.merge refs/heads/tracked
  git config branch.main.remote origin
  git config "lfs.$GITSERVER/$reponame.git.locksverify" true
  git push 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid" "refs/heads/tracked"
)
end_test

begin_test "pre-push locks verify 403 with explicit ref"
(
  set -e

  reponame="lock-verify-explicit-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$GITSERVER/$reponame.git.locksverify" true
  git push origin main:explicit 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid" "refs/heads/explicit"
)
end_test

begin_test "pre-push locks verify 403 with bad ref"
(
  set -e

  reponame="lock-verify-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$GITSERVER/$reponame.git.locksverify" true
  git push origin main 2>&1 | tee push.log
  grep "failed to push some refs" push.log
  refute_server_object "$reponame" "$contents_oid" "refs/heads/other"
)
end_test

begin_test "pre-push locks verify 5xx with verification unset"
(
  set -e

  reponame="lock-unset-verify-5xx"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  [ -z "$(git config "lfs.$endpoint.locksverify")" ]

  git push origin main 2>&1 | tee push.log
  grep "\"origin\" does not support the Git LFS locking API" push.log

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push locks verify 501 with verification enabled"
(
  set -e

  reponame="lock-enabled-verify-501"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" true

  git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid"
  [ "false" = "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test


begin_test "pre-push locks verify 501 with verification disabled"
(
  set -e

  reponame="lock-disabled-verify-501"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" false

  git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid"
  [ "false" = "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test

begin_test "pre-push locks verify 501 with verification unset"
(
  set -e

  reponame="lock-unset-verify-501"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  [ -z "$(git config "lfs.$endpoint.locksverify")" ]

  git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid"
  [ "false" = "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test

begin_test "pre-push locks verify 200"
(
  set -e

  reponame="lock-verify-200"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"
  [ -z "$(git config "lfs.$endpoint.locksverify")" ]

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git push origin main 2>&1 | tee push.log

  grep "Locking support detected on remote \"origin\"." push.log
  grep "git config lfs.$endpoint.locksverify true" push.log
  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push locks verify 403 with verification enabled"
(
  set -e

  reponame="lock-enabled-verify-403"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" true

  git push origin main 2>&1 | tee push.log
  grep "error: Authentication error" push.log

  refute_server_object "$reponame" "$contents_oid"
  [ "true" = "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test

begin_test "pre-push locks verify 403 with verification disabled"
(
  set -e

  reponame="lock-disabled-verify-403"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  git config "lfs.$endpoint.locksverify" false

  git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" "$contents_oid"
  [ "false" = "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test

begin_test "pre-push locks verify 403 with verification unset"
(
  set -e

  reponame="lock-unset-verify-403"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint="$(repo_endpoint $GITSERVER $reponame)"

  contents="example"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat
  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit --message "initial commit"

  [ -z "$(git config "lfs.$endpoint.locksverify")" ]

  git push origin main 2>&1 | tee push.log
  grep "warning: Authentication error" push.log

  assert_server_object "$reponame" "$contents_oid"
  [ -z "$(git config "lfs.$endpoint.locksverify")" ]
)
end_test

begin_test "pre-push with pushDefault and explicit remote"
(
  set -e
  reponame="pre-push-pushdefault-explicit"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git remote add wrong "$(repo_endpoint "$GITSERVER" "wrong-url")"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git config remote.pushDefault wrong
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push origin main 2>&1 | tee push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"
  grep wrong-url push.log && exit 1
  true
)
end_test

begin_test "pre-push uses optimization if remote URL matches"
(
  set -e
  reponame="pre-push-remote-url-optimization"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint=$(git config remote.origin.url)
  contents_oid=$(calc_oid 'hi\n')
  git config "lfs.$endpoint.locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" $contents_oid "refs/heads/main"

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push "$endpoint" main 2>&1 | tee push.log
  grep 'rev-list.*--not --remotes=origin' push.log
)
end_test

begin_test "pre-push does not traverse Git objects server has"
(
  set -e
  reponame="pre-push-traverse-server-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  endpoint=$(git config remote.origin.url)
  contents_oid=$(calc_oid 'hi\n')
  git config "lfs.$endpoint.locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  refute_server_object "$reponame" $contents_oid "refs/heads/main"

  # We use a different URL instead of a named remote or the remote URL so that
  # we can't make use of the optimization that ignores objects we already have
  # in remote tracking branches.
  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push "$endpoint.git" main 2>&1 | tee push.log

  assert_server_object "$reponame" $contents_oid "refs/heads/main"

  contents2_oid=$(calc_oid 'hello\n')
  echo "hello" > b.dat
  git add .gitattributes b.dat
  git commit -m "add b.dat"

  refute_server_object "$reponame" $contents2_oid "refs/heads/main"

  GIT_TRACE=1 GIT_TRANSFER_TRACE=1 git push "$endpoint.git" main 2>&1 | tee push.log

  assert_server_object "$reponame" $contents2_oid "refs/heads/main"

  # Verify that we haven't tried to push or query for the object we already
  # pushed before; i.e., we didn't see it because we ignored its Git object
  # during traversal.
  grep $contents_oid push.log && exit 1
  true
)
end_test

begin_test "pre-push with force-pushed ref"
(
  set -e
  reponame="pre-push-force-pushed-ref"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config "lfs.$(repo_endpoint "$GITSERVER" "$reponame").locksverify" false
  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"
  git tag -a -m tagname tagname

  refute_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"

  git push origin main tagname

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4 "refs/heads/main"

  # We pick a different message so that we get different object IDs even if both
  # commands run in the same second.
  git tag -f -a -m tagname2 tagname
  # Prune the old tag object.
  git reflog expire --all --expire=now
  git gc --prune=now
  # Make sure we deal with us missing the object for the old value of the tag ref.
  git push origin +tagname
)
end_test

begin_test "pre-push with local path"
(
  set -e
  reponame="pre-push-local-path"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame-2"
  cd ..
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hi" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # Push to the other repo.
  git push "../$reponame-2" main:foo

  # Push to . to make sure that works.
  git push "." main:foo

  git lfs fsck
  cd "../$reponame-2"
  git checkout foo
  git lfs fsck
)
end_test
