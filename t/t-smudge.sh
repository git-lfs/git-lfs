#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "smudge"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  output="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9 | git lfs smudge)"
  [ "smudge a" = "$output" ]

  git push origin main

  # download it from the git lfs server
  rm -rf .git/lfs/objects
  output="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9 | git lfs smudge)"
  [ "smudge a" = "$output" ]
)
end_test

begin_test "smudge with temp file"
(
  set -e

  cd repo

  rm -rf .git/lfs/objects
  mkdir -p .git/lfs/tmp/objects
  touch .git/lfs/tmp/objects/fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254-1
  pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9 | GIT_TRACE=5 git lfs smudge | tee smudge.log
  [ "smudge a" = "$(cat smudge.log)" ] || {
    rm -rf .git/lfs/tmp
    git lfs logs last
    exit 1
  }
)
end_test

begin_test "smudge with invalid pointer"
(
  set -e

  cd repo
  [ "wat" = "$(echo "wat" | git lfs smudge)" ]
  [ "not a git-lfs file" = "$(echo "not a git-lfs file" | git lfs smudge)" ]
  [ "version " = "$(echo "version " | git lfs smudge)" ]
)
end_test

begin_test "smudge include/exclude"
(
  set -e

  reponame="$(basename "$0" ".sh")-includeexclude"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" includeexclude

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  pointer="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9)"

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge)" ]

  git push origin main

  # this WOULD download except we're going to prevent it with include/exclude
  rm -rf .git/lfs/objects
  git config "lfs.fetchexclude" "a*"

  [ "$pointer" = "$(echo "$pointer" | git lfs smudge a.dat)" ]
)
end_test

begin_test "smudge with skip"
(
  set -e

  reponame="$(basename "$0" ".sh")-skip"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "skip"

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  pointer="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9)"
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge)" ]

  git push origin main

  # Must clear the cache because smudge will use
  # cached objects even with --skip/GIT_LFS_SKIP_SMUDGE
  # (--skip applies to whether or not it downloads).
  rm -rf .git/lfs/objects

  [ "$pointer" = "$(echo "$pointer" | GIT_LFS_SKIP_SMUDGE=1 git lfs smudge)" ]

  echo "test clone with env"
  export GIT_LFS_SKIP_SMUDGE=1
  env | grep LFS_SKIP
  clone_repo "$reponame" "skip-clone-env"
  [ "$pointer" = "$(cat a.dat)" ]

  git lfs pull
  [ "smudge a" = "$(cat a.dat)" ]

  echo "test clone without env"
  unset GIT_LFS_SKIP_SMUDGE
  clone_repo "$reponame" "no-skip"
  [ "smudge a" = "$(cat a.dat)" ]

  echo "test clone with init --skip-smudge"
  git lfs install --skip-smudge
  clone_repo "$reponame" "skip-clone-init"
  [ "$pointer" = "$(cat a.dat)" ]

  git lfs install --force
)
end_test

begin_test "smudge clone with include/exclude"
(
  set -e

  reponame="smudge_include_exclude"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "repo_$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  assert_local_object "$contents_oid" 1

  git push origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  grep "main -> main" push.log

  assert_server_object "$reponame" "$contents_oid"

  clone="$TRASHDIR/clone_$reponame"
  git -c lfs.fetchexclude="a*" clone "$GITSERVER/$reponame" "$clone"
  cd "$clone"

  # Should have succeeded but not downloaded
  refute_local_object "$contents_oid"

)
end_test

begin_test "smudge skip download failure"
(
  set -e

  reponame="$(basename "$0" ".sh")-skipdownloadfail"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" skipdownloadfail

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  pointer="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9)"

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge)" ]

  git push origin main

  # make it try to download but we're going to make it fail
  rm -rf .git/lfs/objects
  git remote set-url origin httpnope://nope.com/nope

  # this should fail
  set +e
  echo "$pointer" | git lfs smudge a.dat; test ${PIPESTATUS[1]} -ne 0
  set -e

  git config lfs.skipdownloaderrors true
  echo "$pointer" | git lfs smudge a.dat

  # check content too
  [ "$pointer" = "$(echo "$pointer" | git lfs smudge a.dat)" ]

  # now try env var
  git config --unset lfs.skipdownloaderrors

  echo "$pointer" | GIT_LFS_SKIP_DOWNLOAD_ERRORS=1 git lfs smudge a.dat

)
end_test

begin_test "smudge no ref, non-origin"
(
  set -e

  reponame="$(basename "$0" ".sh")-no-ref-non-origin"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame-1"

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main
  main=$(git rev-parse main)

  cd ..
  git init "$reponame"
  cd "$reponame"

  # We intentionally pick a name that is not origin to exercise the remote
  # selection code path. Since there is only one remote, we should use it
  # regardless of its name
  git config remote.random.url "$GITSERVER/$reponame"
  git fetch "$GITSERVER/$reponame"

  git checkout "$main"
  [ "smudge a" = "$(cat a.dat)" ]
)
end_test
