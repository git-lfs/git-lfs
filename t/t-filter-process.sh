#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# HACK(taylor): git uses ".g<hash>" in the version name to signal that it is
# from the "next" branch, which is the only (current) version of Git that has
# support for the filter protocol.
#
ensure_git_version_isnt $VERSION_LOWER "2.11.0"

begin_test "filter process: checking out a branch"
(
  set -e

  reponame="filter_process_checkout"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents_a="contents_a"
  contents_a_oid="$(calc_oid $contents_a)"
  printf "%s" "$contents_a" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git checkout -b b

  contents_b="contents_b"
  contents_b_oid="$(calc_oid $contents_b)"
  printf "%s" "$contents_b" > b.dat

  git add b.dat
  git commit -m "add b.dat"

  git push origin --all

  pushd ..
    # Git will choose filter.lfs.process over `filter.lfs.clean` and
    # `filter.lfs.smudge`
    GIT_TRACE_PACKET=1 git \
      -c "filter.lfs.process=git-lfs filter-process" \
      -c "filter.lfs.clean=false"\
      -c "filter.lfs.smudge=false" \
      -c "filter.lfs.required=true" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    # Assert that we are on the "main" branch, and have a.dat
    [ "main" = "$(git rev-parse --abbrev-ref HEAD)" ]
    [ "$contents_a" = "$(cat a.dat)" ]
    assert_pointer "main" "a.dat" "$contents_a_oid" 10

    git checkout b

    # Assert that we are on the "b" branch, and have b.dat
    [ "b" = "$(git rev-parse --abbrev-ref HEAD)" ]
    [ "$contents_b" = "$(cat b.dat)" ]
    assert_pointer "b" "b.dat" "$contents_b_oid" 10
  popd
)
end_test

begin_test "filter process: include/exclude"
(
  set -e

  reponame="$(basename "$0" ".sh")-includeexclude"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  mkdir -p foo/bar

  contents_a="contents_a"
  contents_a_oid="$(calc_oid $contents_a)"
  printf "%s" "$contents_a" > a.dat
  cp a.dat foo
  cp a.dat foo/bar

  git add .gitattributes a.dat foo
  git commit -m "initial commit"

  git push origin main

  # The Git LFS objects for a.dat and foo/bar/a.dat would both download except
  # we're going to prevent them from doing so with include/exclude.
  # We also need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config --global "lfs.fetchinclude" "/foo"
  MSYS_NO_PATHCONV=1 git config --global "lfs.fetchexclude" "/foo/bar"

  pushd ..
    # Git will choose filter.lfs.process over `filter.lfs.clean` and
    # `filter.lfs.smudge`
    GIT_TRACE_PACKET=1 git \
      -c "filter.lfs.process=git-lfs filter-process" \
      -c "filter.lfs.clean=false"\
      -c "filter.lfs.smudge=false" \
      -c "filter.lfs.required=true" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    pointer="$(pointer "$contents_a_oid" 10)"

    [ "$pointer" = "$(cat a.dat)" ]
    assert_pointer "main" "a.dat" "$contents_a_oid" 10

    [ "$contents_a" = "$(cat foo/a.dat)" ]
    assert_pointer "main" "foo/a.dat" "$contents_a_oid" 10

    [ "$pointer" = "$(cat foo/bar/a.dat)" ]
    assert_pointer "main" "foo/bar/a.dat" "$contents_a_oid" 10
  popd
)
end_test

begin_test "filter process: adding a file"
(
  set -e

  reponame="filter_process_add"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="contents"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git add a.dat

  expected="$(pointer "$contents_oid" "${#contents}")"
  got="$(git cat-file -p :a.dat)"

  diff -u <(echo "$expected") <(echo "$got")
)
end_test

# https://github.com/git-lfs/git-lfs/issues/1697
begin_test "filter process: add a file with 1024 bytes"
(
  set -e

  mkdir repo-issue-1697
  cd repo-issue-1697
  git init
  git lfs track "*.dat"
  dd if=/dev/zero of=first.dat bs=1024 count=1
  printf "any contents" > second.dat
  git add .
)
end_test

begin_test "filter process: hash-object --stdin --path does not hang"
(
  set -e

  mkdir repo-hash-object
  cd repo-hash-object
  git init
  git lfs track "*.dat"
  contents="test"
  contents_oid="$(calc_oid "$contents")"
  expected=$(pointer "$contents_oid" 4 | git hash-object --stdin)

  dd if=/dev/zero of=first.dat bs=1000 count=1
  echo a > second.dat
  # Works for existing file longer than this one.
  output=$(printf test | git hash-object --path first.dat --stdin)
  [ "$expected" = "$output" ]
  # Works for existing file shorter than this one.
  output=$(printf test | git hash-object --path second.dat --stdin)
  [ "$expected" = "$output" ]
  # Works for absent file.
  output=$(printf test | git hash-object --path third.dat --stdin)
  [ "$expected" = "$output" ]

  dd if=/dev/zero of=large.dat bs=65537 count=1
  oid=$(calc_oid_file large.dat)
  expected=$(pointer "$oid" 65537 | git hash-object --stdin)
  output=$(git hash-object --path third.dat --stdin <large.dat)
  [ "$expected" = "$output" ]
  git add .
)
end_test

begin_test "filter process: checking out a branch with --skip-smudge and checkout-index"
(
  set -e

  reponame="filter-process-skip-smudge-checkout-index"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents_a="contents_a"
  contents_a_oid="$(calc_oid $contents_a)"
  printf "%s" "$contents_a" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git checkout -b b

  contents_b="contents_b"
  contents_b_oid="$(calc_oid $contents_b)"
  printf "%s" "$contents_b" > b.dat

  git add b.dat
  git commit -m "add b.dat"

  git lfs install --local --skip-smudge

  git checkout main

  rm a.dat
  git checkout-index -af
  git lfs pointer --check --file a.dat

  assert_pointer "main" "a.dat" "$contents_a_oid" 10

  git checkout b

  rm *.dat
  git checkout-index -af
  git lfs pointer --check --file a.dat
  git lfs pointer --check --file b.dat

  # Assert that we are on the "b" branch, and have b.dat
  assert_pointer "b" "b.dat" "$contents_b_oid" 10
)
end_test

begin_test "filter process: git archive does not invoke SSH"
(
  set -e

  setup_pure_ssh

  reponame="filter-process-archive"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  contents="test"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  git push origin main 2>&1
  cd ..
  GIT_TRACE=1 git clone "$sshurl" "$reponame-2" 2>&1 | tee trace.log
  grep "lfs-ssh-echo.*git-lfs-transfer .*$reponame.git download" trace.log
  cd "$reponame-2"
  GIT_TRACE=1 GIT_TRACE_PACKET=1 git archive -o foo.tar HEAD 2>&1 | tee archive.log
  grep 'pure SSH' archive.log && exit 1
  true
)
end_test
