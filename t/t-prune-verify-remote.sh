#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "prune --verify-remote with a file:// remote"
(
  set -e

  reponame="prune-verify-remote-standalone"
  setup_remote_repo "remote_$reponame"

  cd "$TRASHDIR"
  git init --bare "$reponame-remote.git"
  gitdir="$(pwd)/$reponame-remote.git"

  clone_repo_url "$REMOTEDIR/remote_$reponame.git" "$reponame"

  git remote set-url origin "file://$(urlify "$gitdir")"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  content_head="newer version content"
  content_old="older version content"
  oid_head=$(calc_oid "$content_head")
  oid_old=$(calc_oid "$content_old")

  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_old}, \"Data\":\"$content_old\"}]
  },
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all so no unpushed reason to not prune
  git push origin main

  # set no recents so max ability to prune normally
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0
  git config lfs.pruneoffsetdays 1

  old_object_subpath="${oid_old:0:2}/${oid_old:2:2}/$oid_old"

  # Simulate remote loss of the older, successfully pushed object.
  rm "$gitdir/lfs/objects/$old_object_subpath"

  # Prune must report the object as missing on the remote, exit nonzero,
  # and keep the local copy.
  git lfs prune --verify-remote 2>&1 | tee prune.log
  [ "${PIPESTATUS[0]}" != "0" ]
  grep "2 local objects, 1 retained, 1 not on remote, done." prune.log
  grep "missing on remote:" prune.log
  grep "$oid_old" prune.log
  assert_local_object "$oid_old" "${#content_old}"
  assert_local_object "$oid_head" "${#content_head}"

  # Restore the object on the remote; it should now verify and be pruned.
  mkdir -p "$gitdir/lfs/objects/${oid_old:0:2}/${oid_old:2:2}"
  cp "$(local_object_path "$oid_old")" "$gitdir/lfs/objects/$old_object_subpath"

  git lfs prune --verify-remote 2>&1 | tee prune.log
  grep "2 local objects, 1 retained, 1 verified with remote, done." prune.log
  grep "Deleting objects: 100% (1/1), done." prune.log
  refute_local_object "$oid_old"
  assert_local_object "$oid_head" "${#content_head}"
)
end_test

begin_test "prune --dry-run --verify-remote does not modify local objects"
(
  set -e

  reponame="prune-verify-remote-dry-run"
  setup_remote_repo "remote_$reponame"

  cd "$TRASHDIR"
  git init --bare "$reponame-remote.git"
  gitdir="$(pwd)/$reponame-remote.git"

  clone_repo_url "$REMOTEDIR/remote_$reponame.git" "$reponame"

  git remote set-url origin "file://$(urlify "$gitdir")"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Tracking"

  content_head="newer version content"
  content_old="older version content"
  oid_head=$(calc_oid "$content_head")
  oid_old=$(calc_oid "$content_old")

  echo "[
  {
    \"CommitDate\":\"$(get_date -50d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_old}, \"Data\":\"$content_old\"}]
  },
  {
    \"CommitDate\":\"$(get_date -40d)\",
    \"Files\":[
      {\"Filename\":\"file.dat\",\"Size\":${#content_head}, \"Data\":\"$content_head\"}]
  }
  ]" | lfstest-testutils addcommits

  # push all so no unpushed reason to not prune
  git push origin main

  # set no recents so max ability to prune normally
  git config lfs.fetchrecentrefsdays 0
  git config lfs.fetchrecentremoterefs true
  git config lfs.fetchrecentcommitsdays 0
  git config lfs.pruneoffsetdays 1

  old_object_subpath="${oid_old:0:2}/${oid_old:2:2}/$oid_old"

  # The standalone file adapter links the remote copy to the same inode as
  # the local object, so the file the download stages would share it too
  # and a rename over the local path could never change its mtime.
  # Re-create the remote copy as a separate file to avoid that.
  rm "$gitdir/lfs/objects/$old_object_subpath"
  cp "$(local_object_path "$oid_old")" "$gitdir/lfs/objects/$old_object_subpath"

  # Pin the local object's mtime so that any rewrite is detectable.
  local_old="$(local_object_path "$oid_old")"
  touch -t 200902132331.30 "$local_old"
  mtime_before="$(lfstest-nanomtime "$local_old")"

  git lfs prune --dry-run --verify-remote --verbose 2>&1 | tee prune.log
  grep "2 local objects, 1 retained, 1 verified with remote, done." prune.log
  grep "1 file would be pruned" prune.log
  grep "$oid_old" prune.log

  mtime_after="$(lfstest-nanomtime "$local_old")"
  if [ "$mtime_before" != "$mtime_after" ]; then
    echo >&2 "fatal: --dry-run changed local object mtime: $mtime_before -> $mtime_after"
    exit 1
  fi

  # A real prune must still verify the object and delete it.
  git lfs prune --verify-remote 2>&1 | tee prune.log
  grep "2 local objects, 1 retained, 1 verified with remote, done." prune.log
  grep "Deleting objects: 100% (1/1), done." prune.log
  refute_local_object "$oid_old"
  assert_local_object "$oid_head" "${#content_head}"
)
end_test
