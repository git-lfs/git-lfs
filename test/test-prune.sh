#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "prune unreferenced and old"
(
  set -e

  reponame="prune_unref_old"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  # generate content we'll use
  content_unreferenced="To delete: unreferenced"
  content_oldandpushed="To delete: pushed and too old"
  content_oldandunchanged="Keep: pushed and created a while ago, but still current"
  oid_unreferenced=$(calc_oid "$content_unreferenced")
  oid_oldandpushed=$(calc_oid "$content_oldandpushed")
  oid_oldandunchanged=$(calc_oid "$content_oldandunchanged")
  content_retain1="Retained content 1"
  content_retain2="Retained content 2"
  oid_retain1=$(calc_oid "$content_retain1")
  oid_retain2=$(calc_oid "$content_retain2")


  # Remember for something to be 'too old' it has to appear on the MINUS side
  # of the diff outside the prune window, i.e. it's not when it was introduced
  # but when it disappeared from relevance. That's why changes to file1.dat on master
  # from 7d ago are included even though the commit itself is outside of the window,
  # that content of file1.dat was relevant until it was removed with a commit, inside the window
  # think of it as windows of relevance that overlap until the content is replaced

  # we also make sure we commit today on master so that the recent commits measured
  # from latest commit on master tracks back from there
  echo "[
  {
    \"CommitDate\":\"$(get_date -20d)\",
    \"Files\":[
      {\"Filename\":\"old.dat\",\"Size\":${#content_oldandpushed}, \"Data\":\"$content_oldandpushed\"},
      {\"Filename\":\"stillcurrent.dat\",\"Size\":${#content_oldandunchanged}, \"Data\":\"$content_oldandunchanged\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"old.dat\",\"Size\":${#content_retain1}, \"Data\":\"$content_retain1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -4d)\",
    \"NewBranch\":\"branch_to_delete\",
    \"Files\":[
      {\"Filename\":\"unreferenced.dat\",\"Size\":${#content_unreferenced}, \"Data\":\"$content_unreferenced\"}]
  },
  {
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"old.dat\",\"Size\":${#content_retain2}, \"Data\":\"$content_retain2\"}]
  }
  ]" | lfstest-testutils addcommits

  git push origin master
  git branch -D branch_to_delete

  git config lfs.fetchrecentrefsdays 5
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 3
  git config lfs.pruneoffsetdays 2

  git lfs prune --dry-run --verbose 2>&1 | tee prune.log

  grep "prune: 5 local object(s), 3 retained" prune.log
  grep "prune: 2 file(s) would be pruned" prune.log
  grep "$oid_oldandpushed" prune.log
  grep "$oid_unreferenced" prune.log

  assert_local_object "$oid_oldandpushed" "${#content_oldandpushed}"
  assert_local_object "$oid_unreferenced" "${#content_unreferenced}"
  git lfs prune
  refute_local_object "$oid_oldandpushed" "${#content_oldandpushed}"
  refute_local_object "$oid_unreferenced" "${#content_unreferenced}"
  assert_local_object "$oid_retain1" "${#content_retain1}"
  assert_local_object "$oid_retain2" "${#content_retain2}"

  # now only keep AT refs, no recents
  git config lfs.fetchrecentcommitsdays 0

  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 3 local object(s), 2 retained" prune.log
  grep "prune: Deleting objects: 100% (1/1), done" prune.log
  grep "$oid_retain1" prune.log
  refute_local_object "$oid_retain1"
  assert_local_object "$oid_retain2" "${#content_retain2}"
)
end_test

begin_test "prune keep unpushed"
(
  set -e

  # need to set up many commits on each branch with old data so that would
  # get deleted if it were not for unpushed status (heads would never be pruned but old changes would)
  reponame="prune_keep_unpushed"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log


  content_keepunpushedhead1="Keep: unpushed HEAD 1"
  content_keepunpushedhead2="Keep: unpushed HEAD 2"
  content_keepunpushedhead3="Keep: unpushed HEAD 3"
  content_keepunpushedbranch1="Keep: unpushed second branch 1"
  content_keepunpushedbranch2="Keep: unpushed second branch 2"
  content_keepunpushedbranch3="Keep: unpushed second branch 3"
  oid_keepunpushedhead1=$(calc_oid "$content_keepunpushedhead1")
  oid_keepunpushedhead2=$(calc_oid "$content_keepunpushedhead2")
  oid_keepunpushedhead3=$(calc_oid "$content_keepunpushedhead3")
  oid_keepunpushedbranch1=$(calc_oid "$content_keepunpushedbranch1")
  oid_keepunpushedbranch2=$(calc_oid "$content_keepunpushedbranch2")
  oid_keepunpushedbranch3=$(calc_oid "$content_keepunpushedbranch3")
  oid_keepunpushedtagged1=$(calc_oid "$content_keepunpushedtagged1")
  oid_keepunpushedtagged2=$(calc_oid "$content_keepunpushedtagged1")

  echo "[
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedhead1}, \"Data\":\"$content_keepunpushedhead1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -31d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch_unpushed\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedbranch1}, \"Data\":\"$content_keepunpushedbranch1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -16d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedbranch2}, \"Data\":\"$content_keepunpushedbranch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedbranch3}, \"Data\":\"$content_keepunpushedbranch3\"}]
  },
  {
    \"CommitDate\":\"$(get_date -21d)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedhead2}, \"Data\":\"$content_keepunpushedhead2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -0d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keepunpushedhead3}, \"Data\":\"$content_keepunpushedhead3\"}]
  }
  ]" | lfstest-testutils addcommits

  git config lfs.fetchrecentrefsdays 5
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0 # only keep AT refs, no recents
  git config lfs.pruneoffsetdays 2

  git lfs prune

  # Now push master and show that older versions on master will be removed
  git push origin master

  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 6 local object(s), 4 retained" prune.log
  grep "prune: Deleting objects: 100% (2/2), done" prune.log
  grep "$oid_keepunpushedhead1" prune.log
  grep "$oid_keepunpushedhead2" prune.log
  refute_local_object "$oid_keepunpushedhead1"
  refute_local_object "$oid_keepunpushedhead2"

  # MERGE the secondary branch, delete the branch then push master, then make sure
  # we delete the intermediate commits but also make sure they're on server
  # resolve conflicts by taking other branch
  git merge -Xtheirs branch_unpushed
  git branch -D branch_unpushed
  git lfs prune --dry-run
  git push origin master

  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained" prune.log
  grep "prune: Deleting objects: 100% (3/3), done" prune.log
  grep "$oid_keepunpushedbranch1" prune.log
  grep "$oid_keepunpushedbranch2" prune.log
  grep "$oid_keepunpushedhead3" prune.log
  refute_local_object "$oid_keepunpushedbranch1"
  refute_local_object "$oid_keepunpushedbranch2"
  # we used -Xtheirs so old head state is now obsolete, is the last state on branch
  refute_local_object "$oid_keepunpushedhead3"
  assert_server_object "remote_$reponame" "$oid_keepunpushedbranch1"
  assert_server_object "remote_$reponame" "$oid_keepunpushedbranch2"
  assert_server_object "remote_$reponame" "$oid_keepunpushedhead3"

)
end_test

begin_test "prune keep recent"
(
  set -e

  reponame="prune_recent"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  content_keephead="Keep: HEAD"
  content_keeprecentbranch1tip="Keep: Recent branch 1 tip"
  content_keeprecentbranch2tip="Keep: Recent branch 2 tip"
  content_keeprecentcommithead="Keep: Recent commit on HEAD"
  content_keeprecentcommitbranch1="Keep: Recent commit on recent branch 1"
  content_keeprecentcommitbranch2="Keep: Recent commit on recent branch 2"
  content_prunecommitoldbranch1="Prune: old commit on old branch"
  content_prunecommitoldbranch2="Prune: old branch tip"
  content_prunecommitbranch1="Prune: old commit on recent branch 1"
  content_prunecommitbranch2="Prune: old commit on recent branch 2"
  content_prunecommithead="Prune: old commit on HEAD"
  oid_keephead=$(calc_oid "$content_keephead")
  oid_keeprecentbranch1tip=$(calc_oid "$content_keeprecentbranch1tip")
  oid_keeprecentbranch2tip=$(calc_oid "$content_keeprecentbranch2tip")
  oid_keeprecentcommithead=$(calc_oid "$content_keeprecentcommithead")
  oid_keeprecentcommitbranch1=$(calc_oid "$content_keeprecentcommitbranch1")
  oid_keeprecentcommitbranch2=$(calc_oid "$content_keeprecentcommitbranch2")
  oid_prunecommitoldbranch=$(calc_oid "$content_prunecommitoldbranch1")
  oid_prunecommitoldbranch2=$(calc_oid "$content_prunecommitoldbranch2")
  oid_prunecommitbranch1=$(calc_oid "$content_prunecommitbranch1")
  oid_prunecommitbranch2=$(calc_oid "$content_prunecommitbranch2")
  oid_prunecommithead=$(calc_oid "$content_prunecommithead")


  # use a single file so each commit supercedes the last, if different files
  # then history becomes harder to track
  # Also note that when considering 'recent' when editing a single file, it means
  # that the snapshot state overlapped; so the latest commit *before* the day
  # that you're looking at, not just the commits on/after.
  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_prunecommithead}, \"Data\":\"$content_prunecommithead\"}]
  },
  {
    \"CommitDate\":\"$(get_date -30d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keeprecentcommithead}, \"Data\":\"$content_keeprecentcommithead\"}]
  },
  {
    \"CommitDate\":\"$(get_date -8d)\",
    \"NewBranch\":\"branch_old\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_prunecommitoldbranch1}, \"Data\":\"$content_prunecommitoldbranch1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_prunecommitoldbranch2}, \"Data\":\"$content_prunecommitoldbranch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -9d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch1\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_prunecommitbranch1}, \"Data\":\"$content_prunecommitbranch1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -8d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keeprecentcommitbranch1}, \"Data\":\"$content_keeprecentcommitbranch1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keeprecentbranch1tip}, \"Data\":\"$content_keeprecentbranch1tip\"}]
  },
  {
    \"CommitDate\":\"$(get_date -17d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch2\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_prunecommitbranch2}, \"Data\":\"$content_prunecommitbranch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keeprecentcommitbranch2}, \"Data\":\"$content_keeprecentcommitbranch2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keeprecentbranch2tip}, \"Data\":\"$content_keeprecentbranch2tip\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_keephead}, \"Data\":\"$content_keephead\"}]
  }
  ]" | lfstest-testutils addcommits

  # keep refs for 6 days & any prev commit that overlaps 2 days before tip (recent + offset)
  git config lfs.fetchrecentrefsdays 5
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 1
  git config lfs.pruneoffsetdays 1

  # push everything so that's not a reason to retain
  git push origin master:master branch_old:branch_old branch1:branch1 branch2:branch2


  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 11 local object(s), 6 retained, done" prune.log
  grep "prune: Deleting objects: 100% (5/5), done" prune.log
  grep "$oid_prunecommitoldbranch" prune.log
  grep "$oid_prunecommitoldbranch2" prune.log
  grep "$oid_prunecommitbranch1" prune.log
  grep "$oid_prunecommitbranch2" prune.log
  grep "$oid_prunecommithead" prune.log

  refute_local_object "$oid_prunecommitoldbranch"
  refute_local_object "$oid_prunecommitoldbranch2"
  refute_local_object "$oid_prunecommitbranch1"
  refute_local_object "$oid_prunecommitbranch2"
  refute_local_object "$oid_prunecommithead"
  assert_local_object "$oid_keephead" "${#content_keephead}"
  assert_local_object "$oid_keeprecentbranch1tip" "${#content_keeprecentbranch1tip}"
  assert_local_object "$oid_keeprecentbranch2tip" "${#content_keeprecentbranch2tip}"
  assert_local_object "$oid_keeprecentcommithead" "${#content_keeprecentcommithead}"
  assert_local_object "$oid_keeprecentcommitbranch1" "${#content_keeprecentcommitbranch1}"
  assert_local_object "$oid_keeprecentcommitbranch2" "${#content_keeprecentcommitbranch2}"

  # now don't include any recent commits in fetch & hence don't retain
  # still retain tips of branches
  git config lfs.fetchrecentcommitsdays 0
  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 6 local object(s), 3 retained, done" prune.log
  grep "prune: Deleting objects: 100% (3/3), done" prune.log
  assert_local_object "$oid_keephead" "${#content_keephead}"
  assert_local_object "$oid_keeprecentbranch1tip" "${#content_keeprecentbranch1tip}"
  assert_local_object "$oid_keeprecentbranch2tip" "${#content_keeprecentbranch2tip}"
  refute_local_object "$oid_keeprecentcommithead"
  refute_local_object "$oid_keeprecentcommitbranch1"
  refute_local_object "$oid_keeprecentcommitbranch2"

  # now don't include any recent refs at all, only keep HEAD
  git config lfs.fetchrecentrefsdays 0
  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 3 local object(s), 1 retained, done" prune.log
  grep "prune: Deleting objects: 100% (2/2), done" prune.log
  assert_local_object "$oid_keephead" "${#content_keephead}"
  refute_local_object "$oid_keeprecentbranch1tip"
  refute_local_object "$oid_keeprecentbranch2tip"

)
end_test

begin_test "prune remote tests"
(
  set -e

  reponame="prune_no_or_nonorigin_remote"
  git init "$reponame"
  cd "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":30}]
  },
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":28}]
  },
  {
    \"CommitDate\":\"$(get_date -35d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":37}]
  },
  {
    \"CommitDate\":\"$(get_date -25d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":42}]
  }
  ]" | lfstest-testutils addcommits

  # set no recents so max ability to prune normally
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0
  git config lfs.pruneoffsetdays 1

  # can never prune with no remote
  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 4 retained, done" prune.log


  # also make sure nothing is pruned when remote is not origin
  # create 2 remotes, neither of which is called origin & push to both
  setup_remote_repo "remote1_$reponame"
  setup_remote_repo "remote2_$reponame"
  cd "$TRASHDIR/$reponame"
  git remote add not_origin "$GITSERVER/remote1_$reponame"
  git push not_origin master

  git lfs prune --verbose 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 4 retained, done" prune.log

  # now set the prune remote to be not_origin, should now prune
  # do a dry run so we can also verify
  git config lfs.pruneremotetocheck not_origin

  git lfs prune --verbose --dry-run 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained, done" prune.log
  grep "prune: 3 file(s) would be pruned" prune.log



)
end_test

begin_test "prune verify"
(
  set -e

  reponame="prune_verify"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  content_head="HEAD content"
  content_commit3="Content for commit 3 (prune)"
  content_commit2_failverify="Content for commit 2 (prune - fail verify)"
  content_commit1="Content for commit 1 (prune)"
  oid_head=$(calc_oid "$content_head")
  oid_commit3=$(calc_oid "$content_commit3")
  oid_commit2_failverify=$(calc_oid "$content_commit2_failverify")
  oid_commit1=$(calc_oid "$content_commit1")

  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit1}, \"Data\":\"$content_commit1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit2_failverify}, \"Data\":\"$content_commit2_failverify\"}]
  },
  {
    \"CommitDate\":\"$(get_date -35d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit3}, \"Data\":\"$content_commit3\"}]
  },
  {
    \"CommitDate\":\"$(get_date -25d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all so no unpushed reason to not prune
  git push origin master

  # set no recents so max ability to prune normally
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0
  git config lfs.pruneoffsetdays 1

  # confirm that it would prune with verify when no issues
  git lfs prune --dry-run --verify-remote --verbose 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained, 3 verified with remote, done" prune.log
  grep "prune: 3 file(s) would be pruned" prune.log
  grep "$oid_commit3" prune.log
  grep "$oid_commit2_failverify" prune.log
  grep "$oid_commit1" prune.log

  # delete one file on the server to make the verify fail
  delete_server_object "remote_$reponame" "$oid_commit2_failverify"
  # this should now fail
  git lfs prune --verify-remote 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained, 2 verified with remote, done" prune.log
  grep "missing on remote:" prune.log
  grep "$oid_commit2_failverify" prune.log
  # Nothing should have been deleted
  assert_local_object "$oid_commit1" "${#content_commit1}"
  assert_local_object "$oid_commit2_failverify" "${#content_commit2_failverify}"
  assert_local_object "$oid_commit3" "${#content_commit3}"

  # Now test with the global option
  git config lfs.pruneverifyremotealways true
  # no verify arg but should be pulled from global
  git lfs prune 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained, 2 verified with remote, done" prune.log
  grep "missing on remote:" prune.log
  grep "$oid_commit2_failverify" prune.log
  # Nothing should have been deleted
  assert_local_object "$oid_commit1" "${#content_commit1}"
  assert_local_object "$oid_commit2_failverify" "${#content_commit2_failverify}"
  assert_local_object "$oid_commit3" "${#content_commit3}"

  # now try overriding the global option
  git lfs prune --no-verify-remote 2>&1 | tee prune.log
  grep "prune: 4 local object(s), 1 retained, done" prune.log
  grep "prune: Deleting objects: 100% (3/3), done" prune.log
  # should now have been deleted
  refute_local_object "$oid_commit1"
  refute_local_object "$oid_commit2_failverify"
  refute_local_object "$oid_commit3"

)
end_test

begin_test "prune verify large numbers of refs"
(
  set -e

  reponame="prune_verify_large"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  content_head="HEAD content"
  content_commit1="Recent commit"
  content_oldcommit="Old content"
  oid_head=$(calc_oid "$content_head")

  # Add two recent commits that should not be pruned
  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_oldcommit}, \"Data\":\"$(uuidgen)\"}]
  },
  {
    \"CommitDate\":\"$(get_date -45d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_oldcommit}, \"Data\":\"$(uuidgen)\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_commit1}, \"Data\":\"$content_commit1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # Generate a large number of refs to old commits make sure prune has a lot of data to read
  git checkout $(git log --pretty=oneline  master | tail -2 | awk '{print $1}' | head -1)
  for i in $(seq 0 1000); do
    git tag v$i
  done
  git checkout master

  # push all so no unpushed reason to not prune
  # git push origin master

  # set no recents so max ability to prune normally
  git config lfs.fetchrecentrefsdays 3
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 3
  git config lfs.pruneoffsetdays 3

  # confirm that prune does not hang
  git lfs prune --dry-run --verify-remote --verbose 2>&1 | tee prune.log

)
end_test
