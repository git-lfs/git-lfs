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

  # force use of a spool file with non-pointer input longer than max buffer
  spool="$(lfstest-genrandom --base64 2048)"
  [ "$spool" = "$(echo "$spool" | git lfs smudge)" ]
)
end_test

begin_test "smudge with pointer extension"
(
  set -e

  reponame="smudge-pointer-extension"
  git init "$reponame"
  cd "$reponame"

  setup_case_inverter_extension

  git lfs track "*.dat"

  contents="$(printf "%s\n%s" "abc" "def")"
  contents_oid="$(calc_oid "$contents")"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  mkdir dir1
  printf "%s" "$contents" >dir1/abc.dat
  git add .gitattributes dir1

  pointer="$(case_inverter_extension_pointer "$contents_oid" "$inverted_contents_oid" 7)"

  assert_local_object "$inverted_contents_oid" 7

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  [ "$contents" = "$(echo "$pointer" | git lfs smudge -- "dir1/abc.dat")" ]
  grep "smudge: dir1/abc.dat" "$LFSTEST_EXT_LOG"
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

  mkdir -p foo/bar
  echo "smudge a" > foo/a.dat
  echo "smudge a" > foo/bar/a.dat
  git add foo
  git commit -m 'add foo'

  git push origin main

  # The Git LFS objects for a.dat and foo/bar/a.dat would both download except
  # we're going to prevent them from doing so with include/exclude.
  rm -rf .git/lfs/objects
  # We also need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config "lfs.fetchinclude" "/foo"
  MSYS_NO_PATHCONV=1 git config "lfs.fetchexclude" "/foo/bar"

  [ "$pointer" = "$(echo "$pointer" | git lfs smudge a.dat)" ]
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge foo/a.dat)" ]
  [ "$pointer" = "$(echo "$pointer" | git lfs smudge foo/bar/a.dat)" ]
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
  pushd "$clone"
    # Should have succeeded but not downloaded
    refute_local_object "$contents_oid"
  popd
  rm -rf "$clone"

  contents2="b"
  contents2_oid=$(calc_oid "$contents2")
  contents3="c"
  contents3_oid=$(calc_oid "$contents3")

  mkdir -p foo/bar
  printf "%s" "$contents2" > foo/b.dat
  printf "%s" "$contents3" > foo/bar/c.dat
  git add foo
  git commit -m 'add foo'

  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 1

  git push origin main

  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # The Git LFS objects for a.dat and foo/bar/a.dat would both download except
  # we're going to prevent them from doing so with include/exclude.
  # We also need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config --global "lfs.fetchinclude" "/foo"
  MSYS_NO_PATHCONV=1 git config --global "lfs.fetchexclude" "/foo/bar"
  git clone "$GITSERVER/$reponame" "$clone"
  pushd "$clone"
    refute_local_object "$contents_oid"
    assert_local_object "$contents2_oid" 1
    refute_local_object "$contents3_oid"
  popd
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
