#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "prune"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  # generate content we'll use
  content_unreferenced="To delete: unreferenced"
  content_oldandpushed="To delete: pushed and too old"
  content_oldandunchanged="Keep: pushed and created a while ago, but still current"
  content_keepunpushed="Keep: unpushed"
  content_keepunpushedtagged="Keep: unpushed tagged only"
  content_keephead="Keep: HEAD"
  content_keeprecentbranch="Keep: Recent branch"
  content_keeprecentcommithead="Keep: Recent commit on HEAD"
  content_keeprecentcommitbranch="Keep: Recent commit on recent branch"
  content_keepmergedrecent="Keep: merged secondary branch that's recent"
  content_keepmergedunpushed="Keep: merged secondary branch that's not pushed"
  oid_unreferenced=$(calc_oid "$content_unreferenced")
  oid_oldandpushed=$(calc_oid "$content_oldandpushed")
  oid_oldandunchanged=$(calc_oid "$content_oldandunchanged")
  oid_keepunpushed=$(calc_oid "$content_keepunpushed")
  oid_keepunpushedtagged=$(calc_oid "$content_keepunpushedtagged")
  oid_keephead=$(calc_oid "$content_keephead")
  oid_keeprecentbranch=$(calc_oid "$content_keeprecentbranch")
  oid_keeprecentcommithead=$(calc_oid "$content_keeprecentcommithead")
  oid_keeprecentcommitbranch=$(calc_oid "$content_keeprecentcommitbranch")
  oid_keepmergedrecent=$(calc_oid "$content_keepmergedrecent")
  oid_keepmergedunpushed=$(calc_oid "$content_keepmergedunpushed")

  echo "[
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_oldandpushed}, \"Data\":\"$content_oldandpushed\"},
      {\"Filename\":\"file2.dat\",\"Size\":${#content_oldandunchanged}, \"Data\":\"$content_oldandunchanged\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"NewBranch\":\"branch_to_delete\",    
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_unreferenced}, \"Data\":\"$content_unreferenced\"}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch_to_delete_tagged\",    
    \"Tags\":[\"retain_tag\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_keepunpushedtagged}, \"Data\":\"$content_keepunpushedtagged\"}]
  },
  {
    \"CommitDate\":\"$(get_date -6d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"merge_branch_unpushed\",    
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_keepmergedunpushed}, \"Data\":\"$content_keepmergedunpushed\"}]
  },
  {
    \"CommitDate\":\"$(get_date -3d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"merge_branch_recent\",    
    \"Files\":[
      {\"Filename\":\"file4.dat\",\"Size\":${#content_keepmergedrecent}, \"Data\":\"$content_keepmergedrecent\"}]
  },
  {
    \"CommitDate\":\"$(get_date -6d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch_unpushed\",    
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":${#content_keepunpushedother}, \"Data\":\"$content_keepunpushedother\"}]
  },
  {
    \"CommitDate\":\"$(get_date -6d)\",
    \"ParentBranches\":[\"branch_unpushed\", \"merge_branch_unpushed\"]
  },
  {
    \"CommitDate\":\"$(get_date -4d)\",
    \"ParentBranches\":[\"master\"],
    \"NewBranch\":\"branch_recent\",    
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":${#content_keeprecentcommitbranch}, \"Data\":\"$content_keeprecentcommitbranch\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":${#content_keeprecentbranch}, \"Data\":\"$content_keeprecentbranch\"}]
  },
  {
    \"CommitDate\":\"$(get_date -4d)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_keeprecentcommithead}, \"Data\":\"$content_keeprecentcommithead\"}]
  },
  {
    \"CommitDate\":\"$(get_date -2d)\",
    \"ParentBranches\":[\"master\", \"merge_branch_recent\"]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_keephead}, \"Data\":\"$content_keephead\"}]
  }
  ]" | lfstest-testutils addcommits

  # delete temporary branches
  git branch -D branch_to_delete
  git branch -D branch_to_delete_tagged
  git branch -D merge_branch_unpushed
  git branch -D merge_branch_recent

  # only push master
  git push origin master

  # now create another commit on master that is unpushed
  echo "[
  {
    \"CommitDate\":\"$(get_date -0d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":${#content_keephead}, \"Data\":\"$content_keephead\"}]
  }
  ]" | lfstest-testutils addcommits

  git config lfs.fetchrecentrefsdays 5
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 3
  git config lfs.pruneoffset 2



)
end_test


begin_test "prune unpushed HEAD"
(
    # old commits on HEAD but latest few are not pushed so keep those
    # even changes pre-HEAD
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
