#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "prune unreferenced and old"
(
  set -e

  reponame="prune_unref_old"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

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

  cat prune.log
  grep "5 local objects, 3 retained" prune.log
  grep "2 files would be pruned" prune.log
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
  grep "3 local objects, 2 retained" prune.log
  grep "Pruning 1 files" prune.log
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
  grep "Tracking \*.dat" track.log


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

  git lfs prune 2>&1 | tee prune.log
  cat prune.log
  grep "Nothing to prune" prune.log

  # Now push master and show that older versions on master will be removed
  git push origin master

  git lfs prune --verbose 2>&1 | tee prune.log
  cat prune.log
  grep "6 local objects, 4 retained" prune.log
  grep "Pruning 2 files" prune.log
  grep "$oid_keepunpushedhead1" prune.log
  grep "$oid_keepunpushedhead2" prune.log
  refute_local_object "$oid_keepunpushedhead1"
  refute_local_object "$oid_keepunpushedhead2"


  # MERGE the secondary branch, delete the branch then push master, then make sure 
  # we delete the intermediate commits but also make sure they're on server
  # resolve conflicts by taking other branch
  git merge -Xtheirs branch_unpushed
  git branch -D branch_unpushed
  git lfs prune --dry-run | grep "Nothing to prune"
  git push origin master

  git lfs prune --verbose 2>&1 | tee prune.log
  cat prune.log
  grep "4 local objects, 1 retained" prune.log
  grep "Pruning 3 files" prune.log
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


begin_test "prune worktree"
(
    # old commits on HEAD but latest few are not pushed so keep those
    # even changes pre-HEAD
)
end_test

begin_test "prune no remote"
(
)
end_test

begin_test "prune specify remote"
(
)
end_test
