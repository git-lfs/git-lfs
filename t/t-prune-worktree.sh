#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.5.0"

begin_test "prune worktree"
(
  set -e

  reponame="prune_worktree"
  setup_remote_repo "remote_$reponame"

  clone_repo "remote_$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  content_head="First checkout HEAD"
  content_worktree1head="Worktree 1 head"
  content_worktree1excluded="Worktree 1 excluded by filter"
  content_worktree1indexed="Worktree 1 indexed"
  content_worktree2head="Worktree 2 head"
  content_worktree2excluded="Worktree 2 excluded by filter"
  content_worktree2indexed="Worktree 2 indexed"
  content_oldcommit1="Always pruned 1"
  content_oldcommit2="Always pruned 2"
  content_oldcommit3="Always pruned 3"

  oid_head=$(calc_oid "$content_head")
  oid_worktree1head=$(calc_oid "$content_worktree1head")
  oid_worktree1excluded=$(calc_oid "$content_worktree1excluded")
  oid_worktree1indexed=$(calc_oid "$content_worktree1indexed")
  oid_worktree2head=$(calc_oid "$content_worktree2head")
  oid_worktree2excluded=$(calc_oid "$content_worktree2excluded")
  oid_worktree2indexed=$(calc_oid "$content_worktree2indexed")
  oid_oldcommit1=$(calc_oid "$content_oldcommit1"])
  oid_oldcommit2=$(calc_oid "$content_oldcommit2")
  oid_oldcommit3=$(calc_oid "$content_oldcommit3")

  echo "[
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_oldcommit1}, \"Data\":\"$content_oldcommit1\"}]
  },
  {
    \"CommitDate\":\"$(get_date -35d)\",
    \"NewBranch\":\"branch1\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_oldcommit2}, \"Data\":\"$content_oldcommit2\"}]
  },
  {
    \"CommitDate\":\"$(get_date -20d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_worktree1head}, \"Data\":\"$content_worktree1head\"},
      {\"Filename\":\"foo/file.dat\",\"Size\":${#content_worktree1excluded}, \"Data\":\"$content_worktree1excluded\"}]
  },
  {
    \"CommitDate\":\"$(get_date -30d)\",
    \"ParentBranches\":[\"main\"],
    \"NewBranch\":\"branch2\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_oldcommit3}, \"Data\":\"$content_oldcommit3\"}]
  },
  {
    \"CommitDate\":\"$(get_date -15d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_worktree2head}, \"Data\":\"$content_worktree2head\"},
      {\"Filename\":\"foo/file.dat\",\"Size\":${#content_worktree2excluded}, \"Data\":\"$content_worktree2excluded\"}]
  },
  {
    \"CommitDate\":\"$(get_date -30d)\",
    \"ParentBranches\":[\"main\"],
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # push everything so that's not a retention issue
  git push origin main:main branch1:branch1 branch2:branch2

  # don't keep any recent, just checkouts
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0

  # We need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config "lfs.fetchexclude" "/foo"

  # before worktree, everything except current checkout would be pruned
  git lfs prune --dry-run 2>&1 | tee prune.log
  grep "prune: 8 local objects, 1 retained, done." prune.log
  grep "prune: 7 files would be pruned" prune.log

  # now add worktrees on the other branches
  git worktree add "../w1_$reponame" "branch1"
  git worktree add "../w2_$reponame" "branch2"

  # stage files in worktrees
  cd "../w1_$reponame"
  echo "$content_worktree1indexed" > indexed.dat
  git lfs track "*.dat"
  git add indexed.dat

  cd "../w2_$reponame"
  echo "$content_worktree2indexed" > indexed.dat
  git lfs track "*.dat"
  git add indexed.dat

  cd "../$reponame"

  # now should retain all 3 heads except for paths excluded by filter plus the indexed files
  git lfs prune --dry-run 2>&1 | tee prune.log
  grep "prune: 10 local objects, 5 retained, done." prune.log
  grep "prune: 5 files would be pruned" prune.log

  # also check that the same result is obtained when inside worktree rather than main
  cd "../w1_$reponame"
  git lfs prune --dry-run 2>&1 | tee prune.log
  grep "prune: 10 local objects, 5 retained, done." prune.log
  grep "prune: 5 files would be pruned" prune.log

  # now remove a worktree & prove that frees up 1 head while keeping the other
  cd "../$reponame"
  rm -rf "../w1_$reponame"
  git worktree prune # required to get git to tidy worktree metadata
  git lfs prune --dry-run 2>&1 | tee prune.log
  grep "prune: 10 local objects, 3 retained, done." prune.log
  grep "prune: 7 files would be pruned" prune.log
)
end_test
