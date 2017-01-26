#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "pre-push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "add git attributes"

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
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
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "(1 of 1 files)" push.log

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

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log

  [ "" = "$(cat push.log)" ]

  git lfs track "*.dat"
  echo "dry" > hi.dat
  git add hi.dat
  git commit -m "add hi.dat"
  git show

  refute_server_object "$reponame" 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push --dry-run origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "push 2840e0eafda1d0760771fe28b91247cf81c76aa888af28a850b5648a338dc15b => hi.dat" push.log
  cat push.log
  [ `wc -l < push.log` = 1 ]

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
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/redirect307/rel/$reponame.git/info/lfs" 2>&1 |
    tee push.log
  grep "(0 of 0 files, 1 skipped)" push.log

  assert_server_object "$reponame" 98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4

  # absolute redirect
  git config remote.origin.lfsurl "$GITSERVER/redirect307/abs/$reponame.git/info/lfs"

  echo "hi" > hi2.dat
  git add hi2.dat
  git commit -m "add hi2.dat"
  git show

  # push file to the git lfs server
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/redirect307/abs/$reponame.git/info/lfs" 2>&1 |
    tee push.log
  grep "(0 of 0 files, 1 skipped)" push.log
)
end_test

begin_test "pre-push with existing file"
(
  set -e

  reponame="$(basename "$0" ".sh")-existing-file"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" existing-file
  echo "existing" > existing.dat
  git add existing.dat
  git commit -m "add existing dat"

  git lfs track "*.dat"
  echo "new" > new.dat
  git add new.dat
  git add .gitattributes
  git commit -m "add new file through git lfs"

  # push file to the git lfs server
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "(1 of 1 files)" push.log

  # now the file exists
  assert_server_object "$reponame" 7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c
)
end_test

begin_test "pre-push with existing pointer"
(
  set -e

  reponame="$(basename "$0" ".sh")-existing-pointer"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" existing-pointer

  echo "$(pointer "7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c" 4)" > new.dat
  git add new.dat
  git commit -m "add new pointer"
  mkdir -p .git/lfs/objects/7a/a7
  echo "new" > .git/lfs/objects/7a/a7/7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c

  # push file to the git lfs server
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "(1 of 1 files)" push.log
)
end_test

begin_test "pre-push with missing pointer not on server"
(
  set -e

  reponame="$(basename "$0" ".sh")-missing-pointer"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" missing-pointer

  echo "$(pointer "7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c" 4)" > new.dat
  git add new.dat
  git commit -m "add new pointer"

  # assert that push fails
  set +e
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  set -e
  grep "Unable to find object (7aa7a5359173d05b63cfd682e3c38487f3cb4f7f1d60659fe59fab1505977d4c) locally." push.log
)
end_test

begin_test "pre-push with missing pointer which is on server"
(
  # should permit push if files missing locally but are on server, shouldn't
  # require client to have every file (prune)
  set -e

  reponame="$(basename "$0" ".sh")-missing-but-on-server"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" missing-but-on-server

  contents="common data"
  contents_oid=$(calc_oid "$contents")
  git lfs track "*.dat"
  printf "$contents" > common1.dat
  git add common1.dat
  git add .gitattributes
  git commit -m "add first file"

  # push file to the git lfs server
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  grep "(1 of 1 files)" push.log

  # now the file exists
  assert_server_object "$reponame" "$contents_oid"

  # create another commit referencing same oid, then delete local data & push
  printf "$contents" > common2.dat
  git add common2.dat
  git commit -m "add second file, same content"
  rm -rf .git/lfs/objects
  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
    git lfs pre-push origin "$GITSERVER/$reponame" 2>&1 |
    tee push.log
  # make sure there were no errors reported
  [ -z "$(grep -i 'Error' push.log)" ]

)
end_test

begin_test "pre-push multiple branches"
(
  set -e

  reponame="$(basename "$0" ".sh")-multiple-branches"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"


  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

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
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch2\",
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch3\",
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content[4]}, \"Data\":\"${content[4]}\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch4\",
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[5]}, \"Data\":\"${content[5]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # make sure when called via git push all branches are updated
  git push origin master branch1 branch2 branch3 branch4
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

  echo "refs/heads/master master refs/heads/master 0000000000000000000000000000000000000000" |
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
  grep "Tracking \*.dat" track.log

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
  git push origin master branch-to-delete
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
  grep "Tracking \*.dat" track.log

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
    \"ParentBranches\":[\"master\"],
    \"CommitDate\":\"$(get_date -0d)\",
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content[3]}, \"Data\":\"${content[3]}\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all branches
  git push origin master branch-to-delete
  for ((a=0; a < NUMFILES ; a++))
  do
    assert_server_object "$reponame" "${oid[$a]}"
  done

  # deleting a branch with git push should not fail
  # (requires correct special casing of "(delete) 0000000000.." in hook)
  git push origin --delete branch-to-delete


)
end_test

begin_test "pre-push with own lock"
(
  set -e

  reponame="pre_push_owned_locks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "$contents" > locked.dat
  git add locked.dat
  git commit -m "add locked.dat"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked.dat" | tee lock.log
  grep "'locked.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  printf "authorized changes" >> locked.dat
  git add locked.dat
  git commit -m "add unauthroized changes"

  git push origin master 2>&1 | tee push.log

  assert_server_lock "$id"

  grep "Consider unlocking your own locked file(s)" push.log
  grep "* locked.dat" push.log
)
end_test

begin_test "pre-push with unowned lock"
(
  set -e

  reponame="pre_push_unowned_lock"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # Use a different Git persona so the locks are owned by a different person
  git config --local user.name "Example Locker"
  git config --local user.email "locker@example.com"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "$contents" > locked_unowned.dat
  git add locked_unowned.dat
  git commit -m "add locked_unowned.dat"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_unowned.dat" | tee lock.log
  grep "'locked_unowned.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  pushd "$TRASHDIR" >/dev/null
    clone_repo "$reponame" "$reponame-assert"

    printf "unauthorized changes" >> locked_unowned.dat
    git add locked_unowned.dat
    # --no-verify is used to avoid the pre-commit hook which is not under test
    git commit --no-verify -m "add unauthroized changes"

    git push origin master 2>&1 | tee push.log

    grep "Unable to push 1 locked file(s)" push.log
    grep "* locked_unowned.dat - Example Locker <locker@example.com>" push.log
  popd >/dev/null
)
end_test
