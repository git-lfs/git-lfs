#!/usr/bin/env bash

. "test/testlib.sh"

if [ "$GIT_LFS_USE_LEGACY_FILTER" == "1" ]; then
  echo "skip: $0 (filter process disabled)"
  exit
fi

# HACK(taylor): git uses ".g<hash>" in the version name to signal that it is
# from the "next" branch, which is the only (current) version of Git that has
# support for the filter protocol.
#
# Once 2.11 is released, replace this with:
#
# ```
# ensure_git_version_isnt $VERSION_LOWER "2.11.0"
# ```
if [ "1" -ne "$(git version | cut -d ' ' -f 3 | grep -c "g")" ]; then
  echo "skip: $0 git version does not include support for filter protocol"
  exit
fi

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
  printf "$contents_a" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git checkout -b b

  contents_b="contents_b"
  contents_b_oid="$(calc_oid $contents_b)"
  printf "$contents_b" > b.dat

  git add b.dat
  git commit -m "add b.dat"

  git push origin --all

  pushd ..
    git \
      -c "filter.lfs.process=git-lfs filter" \
      -c "filter.lfs.clean="\
      -c "filter.lfs.smudge=" \
      -c "filter.lfs.required=true" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    # Assert that we are on the "master" branch, and have a.dat
    [ "master" = "$(git rev-parse --abbrev-ref HEAD)" ]
    [ "$contents_a" = "$(cat a.dat)" ]
    assert_pointer "master" "a.dat" "$contents_a_oid" 10

    git checkout b

    # Assert that we are on the "b" branch, and have b.dat
    [ "b" = "$(git rev-parse --abbrev-ref HEAD)" ]
    [ "$contents_b" = "$(cat b.dat)" ]
    assert_pointer "b" "b.dat" "$contents_b_oid" 10
  popd
)
end_test
