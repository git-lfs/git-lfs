#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "pull"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" clone
  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")
  contents2="A"
  contents2_oid=$(calc_oid "$contents2")
  contents3="dir"
  contents3_oid=$(calc_oid "$contents3")

  mkdir dir
  echo "*.log" > .gitignore
  printf "%s" "$contents" > a.dat
  printf "%s" "$contents2" > á.dat
  printf "%s" "$contents3" > dir/dir.dat
  git add .
  git commit -m "add files" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "5 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]

  assert_pointer "main" "a.dat" "$contents_oid" 1
  assert_pointer "main" "á.dat" "$contents2_oid" 1
  assert_pointer "main" "dir/dir.dat" "$contents3_oid" 3

  refute_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents2_oid"
  refute_server_object "$reponame" "$contents3_oid"

  echo "initial push"
  git push origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (3/3), 5 B" push.log
  grep "main -> main" push.log

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # change to the clone's working directory
  cd ../clone

  echo "normal pull"
  git config branch.main.remote origin
  git config branch.main.merge refs/heads/main
  git pull 2>&1

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]

  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull"
  rm -rf a.dat á.dat dir # removing files makes the status dirty
  rm -rf .git/lfs/objects
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status
  git lfs fsck

  echo "lfs pull with remote"
  rm -rf a.dat á.dat dir
  rm -rf .git/lfs/objects
  git lfs pull origin
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status
  git lfs fsck

  echo "lfs pull with local storage"
  rm -rf a.dat á.dat dir
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_clean_status

  echo "lfs pull with include/exclude filters in gitconfig"
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs pull
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs pull
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "lfs pull with include/exclude filters in command line"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs pull --include="a*"
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git lfs pull --exclude="a*"
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "resetting to test status"
  git reset --hard
  assert_clean_status

  echo "lfs pull clean status"
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull with -I"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir/dir.dat"
  git lfs pull -I "a.*,dir/dir.*"
  [ "a" = "$(cat a.dat)" ]
  [ ! -e "á.dat" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  refute_local_object "$contents2_oid"
  assert_local_object "$contents3_oid" 3
  assert_clean_worktree_with_exceptions '\\303\\241\.dat'

  git lfs pull -I "*.dat"
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents2_oid" 1
  assert_clean_status

  echo "lfs pull with empty file"
  touch empty.dat
  git add empty.dat
  git commit -m 'empty'
  git lfs pull
  [ -z "$(cat empty.dat)" ]
  assert_clean_status

  echo "resetting to test status"
  git reset --hard HEAD^
  assert_clean_status

  echo "lfs pull in subdir"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir/dir.dat"
  pushd dir
    git lfs pull
  popd
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_status

  echo "lfs pull in subdir with -I"
  rm -rf .git/lfs/objects
  rm -rf a.dat "á.dat" "dir/dir.dat"
  pushd dir
    git lfs pull -I "á.*,dir/dir.dat"
  popd
  [ ! -e a.dat ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]
  refute_local_object "$contents_oid"
  assert_local_object "$contents2_oid" 1
  assert_local_object "$contents3_oid" 3
  assert_clean_worktree_with_exceptions "a\.dat"

  pushd dir
    git lfs pull -I "*.dat"
  popd
  [ "a" = "$(cat a.dat)" ]
  assert_local_object "$contents_oid" 1
  assert_clean_status
)
end_test

begin_test "pull without clean filter"
(
  set -e

  GIT_LFS_SKIP_SMUDGE=1 git clone $GITSERVER/t-pull no-clean
  cd no-clean
  git lfs uninstall
  git config --list > config.txt
  grep "filter.lfs.clean" config.txt && {
    echo "clean filter still configured:"
    cat config.txt
    exit 1
  }

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid"

  git lfs pull | tee pull.txt
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep "Git LFS is not installed" pull.txt
  echo "pulled!"

  # LFS object downloaded, pointer unchanged
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid" 1
)
end_test

begin_test "pull with raw remote url"
(
  set -e
  mkdir raw
  cd raw
  git init
  git lfs install --local --skip-smudge

  git remote add origin $GITSERVER/t-pull
  git pull origin main

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  git lfs pull "$GITSERVER/t-pull"
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ "0" = "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with multiple remotes"
(
  set -e
  mkdir multiple
  cd multiple
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git remote add bad-remote "invalid-url"
  git pull origin main

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  # pull should default to origin instead of bad-remote
  git lfs pull
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ "0" = "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with invalid insteadof"
(
  set -e
  mkdir insteadof
  cd insteadof
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git pull origin main

  # set insteadOf to rewrite the href of downloading LFS object.
  git config url."$GITSERVER/storage/invalid".insteadOf "$GITSERVER/storage/"
  # Enable href rewriting explicitly.
  git config lfs.transfer.enablehrefrewrite true

  set +e
  git lfs pull > pull.log 2>&1
  res=$?

  set -e
  [ "$res" = "2" ]

  # check rewritten href is used to download LFS object.
  grep "LFS: Repository or object not found: $GITSERVER/storage/invalid" pull.log

  # lfs-pull succeed after unsetting enableHrefRewrite config
  git config --unset lfs.transfer.enablehrefrewrite
  git lfs pull
)
end_test

begin_test "pull with merge conflict"
(
  set -e
  git init pull-merge-conflict
  cd pull-merge-conflict

  git lfs track "*.bin"
  git add .
  git commit -m 'gitattributes'
  printf abc > abc.bin
  git add .
  git commit -m 'abc'

  git checkout -b def
  printf def > abc.bin
  git add .
  git commit -m 'def'

  git checkout main
  printf ghi > abc.bin
  git add .
  git commit -m 'ghi'

  # This will exit nonzero because of the merge conflict.
  GIT_LFS_SKIP_SMUDGE=1 git merge def || true
  git lfs pull > pull.log 2>&1
  [ ! -s pull.log ]
)
end_test

begin_test "pull: with missing object"
(
  set -e

  # this clone is setup in the first test in this file
  cd clone
  rm -rf .git/lfs/objects
  rm a.dat

  contents_oid=$(calc_oid "a")
  reponame="$(basename "$0" ".sh")"
  delete_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents_oid"

  # should return non-zero, but should also download all the other valid files too
  git config branch.main.remote origin
  git config branch.main.merge refs/heads/main
  git lfs pull 2>&1 | tee pull.log
  pull_exit="${PIPESTATUS[0]}"
  [ "$pull_exit" != "0" ]

  grep "$contents_oid does not exist" pull.log

  contents2_oid=$(calc_oid "A")
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents_oid"
)
end_test

begin_test "pull: outside git repository"
(
  set +e
  git lfs pull 2>&1 > pull.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a Git repository" pull.log
)
end_test

begin_test "pull: read-only directory"
(
  set -e

  skip_if_root_or_admin "$test_description"

  reponame="pull-read-only"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="a"
  contents_oid=$(calc_oid "$contents")
  mkdir dir
  printf "%s" "$contents" > dir/a.bin

  git add .gitattributes dir/a.bin
  git commit -m "add dir/a.bin"

  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  rm dir/a.bin
  delete_local_object "$contents_oid"

  if [ "$IS_WINDOWS" -eq 1 ]; then
    icacls dir /inheritance:r
    icacls dir /grant:r Everyone:R
  else
    chmod a-w dir
  fi
  git lfs pull 2>&1 | tee pull.log
  # Note that although the pull command should log an error, at present
  # we still expect a zero exit code.
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  assert_local_object "$contents_oid" 1

  [ ! -e dir/a.bin ]

  grep 'could not check out "dir/a.bin"' pull.log
  grep 'could not create working directory file' pull.log
  grep 'permission denied' pull.log
)
end_test

begin_test "pull with empty file doesn't modify mtime"
(
  set -e
  git init pull-empty-file
  cd pull-empty-file

  git lfs track "*.bin"
  git add .
  git commit -m 'gitattributes'
  printf abc > abc.bin
  git add .
  git commit -m 'abc'

  touch foo.bin
  lfstest-nanomtime foo.bin >foo.mtime

  # This isn't necessary, but it takes a few cycles to make sure that our
  # timestamp changes.
  git add foo.bin
  git commit -m 'foo'

  git lfs pull
  lfstest-nanomtime foo.bin >foo.mtime2
  diff -u foo.mtime foo.mtime2
)
end_test

begin_test "pull with partial clone and sparse checkout and index"
(
  set -e

  # Only test with Git version 2.42.0 as it introduced support for the
  # "objecttype" format option to the "git ls-files" command, which our
  # code requires.
  ensure_git_version_isnt "$VERSION_LOWER" "2.42.0"

  reponame="pull-sparse"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents1="a"
  contents1_oid=$(calc_oid "$contents1")
  contents2="b"
  contents2_oid=$(calc_oid "$contents2")
  contents3="c"
  contents3_oid=$(calc_oid "$contents3")

  mkdir in-dir out-dir
  printf "%s" "$contents1" >a.dat
  printf "%s" "$contents2" >in-dir/b.dat
  printf "%s" "$contents3" >out-dir/c.dat
  git add .
  git commit -m "add files"

  git push origin main

  assert_server_object "$reponame" "$contents1_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # Create a partial clone with a cone-mode sparse checkout of one directory
  # and a sparse index, which is important because otherwise the "git ls-files"
  # command ignores the --sparse option and lists all Git LFS files.
  cd ..
  git clone --filter=tree:0 --depth=1 --no-checkout \
    "$GITSERVER/$reponame" "${reponame}-partial"

  cd "${reponame}-partial"
  git sparse-checkout init --cone --sparse-index
  git sparse-checkout set "in-dir"
  git checkout main

  [ -d "in-dir" ]
  [ ! -e "out-dir" ]

  assert_local_object "$contents1_oid" 1
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents3_oid"

  # Git LFS objects associated with files outside of the sparse cone
  # should not be pulled.
  git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep -q "Downloading LFS objects" pull.log && exit 1

  refute_local_object "$contents3_oid"
)
end_test

begin_test "pull: pointer extension"
(
  set -e

  reponame="pull-pointer-extension"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  setup_case_inverter_extension

  contents="abc"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  mkdir dir1
  printf "%s" "$contents" >dir1/abc.dat

  git add .gitattributes dir1
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$inverted_contents_oid"

  cd ..
  GIT_LFS_SKIP_SMUDGE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert"

  cd "${reponame}-assert"
  refute_local_object "$inverted_contents_oid"

  setup_case_inverter_extension

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  git lfs pull

  assert_local_object "$inverted_contents_oid" 3

  [ "$contents" = "$(cat "dir1/abc.dat")" ]
  grep "smudge: dir1/abc.dat" "$LFSTEST_EXT_LOG"

  rm -rf .git/lfs/objects

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  mkdir dir2

  pushd dir2
    git lfs pull
  popd

  [ "$contents" = "$(cat "dir1/abc.dat")" ]

  # Note that at present we expect "git lfs pull" to run the extension
  # program in the current working directory rather than the repository root,
  # as would occur if it was run within a smudge filter operation started
  # by Git.
  grep "smudge: ../dir1/abc.dat" "$LFSTEST_EXT_LOG"

  assert_local_object "$inverted_contents_oid" 3
)
end_test
