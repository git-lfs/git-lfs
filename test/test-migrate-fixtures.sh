#!/usr/bin/env bash

# assert_ref_unmoved ensures that the previous and current SHA1 of a given ref
# is equal by string comparison:
#
#   assert_ref_unmoved "HEAD" "$previous_sha" "$current_sha"
#
# If the two are unequal (the ref has moved), a message is printed to stderr and
# the program exits.
assert_ref_unmoved() {
  local name="$1"
  local prev_sha="$2"
  local current_sha="$3"

  if [ "$prev_sha" != "$current_sha" ]; then
    echo >&2 "$name should not have moved (from: $prev_sha, to: $current_sha)"
    exit 1
  fi
}

# setup_local_branch_with_gitattrs creates a repository as follows:
#
#   A---B
#        \
#         refs/heads/master
#
# - Commit 'A' has 120, in a.txt, and a corresponding entry in .gitattributes.
setup_local_branch_with_gitattrs() {
  set -e

  reponame="migrate-single-remote-branch-with-attrs"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 120 > a.txt

  git add a.txt
  git commit -m "initial commit"

  git lfs track "*.txt"
  git lfs track "*.other"

  git add .gitattributes
  git commit -m "add .gitattributes"
}

# setup_local_branch_with_nested_gitattrs creates a repository as follows:
#
#   A---B
#        \
#         refs/heads/master
#
# - Commit 'A' has 120, in a.txt, and a corresponding entry in .gitattributes. There is also
#   140 in a.md, with no corresponding entry in .gitattributes.
#   It also has 140 in subtree/a.md, and a corresponding entry in subtree/.gitattributes
setup_local_branch_with_nested_gitattrs() {
  set -e

  reponame="nested-attrs"

  remove_and_create_local_repo "$reponame"

  mkdir b

  base64 < /dev/urandom | head -c 120 > a.txt
  base64 < /dev/urandom | head -c 140 > a.md
  base64 < /dev/urandom | head -c 140 > b/a.md

  git add a.txt a.md b/a.md
  git commit -m "initial commit"

  git lfs track "*.txt"

  git add .gitattributes
  git commit -m "add .gitattributes"

  cd b

  git lfs track "*.md"

  cd ..

  git add b/.gitattributes
  git commit -m "add nested .gitattributes"
}

# setup_multiple_local_branches creates a repository as follows:
#
#     B
#    / \
#   A   refs/heads/my-feature
#    \
#     refs/heads/master
#
# - Commit 'A' has 120, 140 bytes of data in a.txt, and a.md, respectively.
#
# - Commit 'B' has 30 bytes of data in a.txt, and includes commit 'A' as a
#   parent.
setup_multiple_local_branches() {
  set -e

  reponame="migrate-info-multiple-local-branches"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 120 > a.txt
  base64 < /dev/urandom | head -c 140 > a.md

  git add a.txt a.md
  git commit -m "initial commit"

  git checkout -b my-feature

  base64 < /dev/urandom | head -c 30 > a.md

  git add a.md
  git commit -m "add an additional 30 bytes to a.md"

  git checkout master
}

# setup_multiple_local_branches_with_gitattrs creates a repository in the same way
# as setup_multiple_local_branches, but also adds relevant lfs filters to the
# .gitattributes file in the master branch
setup_multiple_local_branches_with_gitattrs() {
  set -e

  setup_multiple_local_branches

  git lfs track *.txt
  git lfs track *.md

  git add .gitattributes
  git commit -m "add .gitattributes"
}

# setup_local_branch_with_space creates a repository as follows:
#
#   A
#    \
#     refs/heads/master
#
# - Commit 'A' has 50 bytes in a file named "a file.txt".
setup_local_branch_with_space() {
  set -e

  reponame="migrate-local-branch-with-space"
  filename="a file.txt"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 50 > "$filename"

  git add "$filename"
  git commit -m "initial commit"
}

# setup_single_remote_branch creates a repository as follows:
#
#   A---B
#    \   \
#     \   refs/heads/master
#      \
#       refs/remotes/origin/master
#
# - Commit 'A' has 120, 140 bytes of data in a.txt, and a.md, respectively. It
#   is the latest commit pushed to the remote 'origin'.
#
# - Commit 'B' has 30, 50 bytes of data in a.txt, and a.md, respectively.
setup_single_remote_branch() {
  set -e

  reponame="migrate-info-single-remote-branch"

  remove_and_create_remote_repo "$reponame"

  base64 < /dev/urandom | head -c 120 > a.txt
  base64 < /dev/urandom | head -c 140 > a.md

  git add a.txt a.md
  git commit -m "initial commit"

  git push origin master

  base64 < /dev/urandom | head -c 30 > a.txt
  base64 < /dev/urandom | head -c 50 > a.md

  git add a.md a.txt
  git commit -m "add an additional 30, 50 bytes to a.{txt,md}"
}

setup_single_remote_branch_with_gitattrs() {
  set -e

  setup_single_remote_branch

  git lfs track *.txt
  git lfs track *.md

  git add .gitattributes
  git commit -m "add .gitattributes"
}

# Creates a repo identical to setup_single_remote_branch, except with *.md and
# *.txt files tracked by Git LFS, and all commits pushed to master
setup_single_remote_branch_tracked() {
  set -e

  reponame="migrate-info-single-remote-branch"

  remove_and_create_remote_repo "$reponame"

  git lfs track "*.md" "*.txt"
  git add .gitattributes
  git commit -m "initial commit"

  base64 < /dev/urandom | head -c 120 > a.txt
  base64 < /dev/urandom | head -c 140 > a.md

  git add a.txt a.md
  git commit -m "add a.{txt,md}"

  git push origin master

  base64 < /dev/urandom | head -c 30 > a.txt
  base64 < /dev/urandom | head -c 50 > a.md

  git add a.md a.txt
  git commit -m "add an additional 30, 50 bytes to a.{txt,md}"

  git push origin master
}

# setup_multiple_remote_branches creates a repository as follows:
#
#         C
#        / \
#   A---B   refs/heads/my-feature
#    \   \
#     \   refs/heads/master
#      \
#       refs/remotes/origin/master
#
# - Commit 'A' has 10, 11 bytes of data in a.txt, and a.md, respectively. It is
#   the latest commit pushed to the remote 'origin'.
#
# - Commit 'B' has 20, 21 bytes of data in a.txt, and a.md, respectively.
#
# - Commit 'C' has 30, 31 bytes of data in a.txt, and a.md, respectively. It is
#   the latest commit on refs/heads/my-feature.
setup_multiple_remote_branches() {
  set -e

  reponame="migrate-info-exclude-remote-refs-given-branch"

  remove_and_create_remote_repo "$reponame"

  base64 < /dev/urandom | head -c 10 > a.txt
  base64 < /dev/urandom | head -c 11 > a.md
  git add a.txt a.md
  git commit -m "add 10, 11 bytes, a.{txt,md}"

  git push origin master

  base64 < /dev/urandom | head -c 20 > a.txt
  base64 < /dev/urandom | head -c 21 > a.md
  git add a.txt a.md
  git commit -m "add 20, 21 bytes, a.{txt,md}"

  git checkout -b my-feature

  base64 < /dev/urandom | head -c 30 > a.txt
  base64 < /dev/urandom | head -c 31 > a.md
  git add a.txt a.md
  git commit -m "add 30, 31 bytes, a.{txt,md}"

  git checkout master
}

# setup_single_local_branch_with_tags creates a repository as follows:
#
#   A---B
#       |\
#       | refs/heads/master
#       |
#        \
#         refs/tags/v1.0.0
#
# - Commit 'A' has 1 byte of data in 'a.txt'
# - Commit 'B' has 2 bytes of data in 'a.txt', and is tagged at 'v1.0.0'.
setup_single_local_branch_with_tags() {
  set -e

  reponame="migrate-single-local-branch-tags"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 1 > a.txt

  git add a.txt
  git commit -m "initial commit"

  base64 < /dev/urandom | head -c 2 > a.txt

  git add a.txt
  git commit -m "secondary commit"
  git tag "v1.0.0"
}

# setup_single_local_branch_with_annotated_tags creates a repository as follows:
#
#   A---B
#       |\
#       | refs/heads/master
#       |
#        \
#         refs/tags/v1.0.0 (annotated)
#
# - Commit 'A' has 1 byte of data in 'a.txt'
# - Commit 'B' has 2 bytes of data in 'a.txt', and is tagged (with annotation)
#     at 'v1.0.0'.
setup_single_local_branch_with_annotated_tags() {
  set -e

  reponame="migrate-single-local-branch-annotated-tags"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 1 > a.txt

  git add a.txt
  git commit -m "initial commit"

  base64 < /dev/urandom | head -c 2 > a.txt

  git add a.txt
  git commit -m "secondary commit"
  git tag "v1.0.0" -m "v1.0.0"
}

setup_multiple_remotes() {
  set -e

  reponame="migrate-multiple-remotes"
  remove_and_create_remote_repo "$reponame"

  forkname="$(git remote -v \
    | head -n1 \
    | cut -d ' ' -f 1 \
    | sed -e 's/^.*\///g')-fork"
  ( setup_remote_repo "$forkname" )

  git remote add fork "$GITSERVER/$forkname"

  base64 < /dev/urandom | head -c 16 > a.txt
  git add a.txt
  git commit -m "initial commit"
  git push origin master

  base64 < /dev/urandom | head -c 16 > a.txt
  git add a.txt
  git commit -m "another commit"
  git push fork master
}

# setup_single_local_branch_deep_trees creates a repository as follows:
#
#   A
#    \
#     refs/heads/master
#
# - Commit 'A' has 120 bytes of data in 'foo/bar/baz/a.txt'.
setup_single_local_branch_deep_trees() {
  set -e

  reponame="migrate-single-local-branch-with-deep-trees"
  remove_and_create_local_repo "$reponame"

  mkdir -p foo/bar/baz
  base64 < /dev/urandom | head -c 120 > foo/bar/baz/a.txt

  git add foo/bar/baz/a.txt
  git commit -m "initial commit"
}

# setup_local_branch_with_symlink creates a repository as follows:
#
#   A
#    \
#     refs/heads/master
#
# - Commit 'A' has 120, in a.txt, and a symbolic link link.txt to a.txt.
setup_local_branch_with_symlink() {
  set -e

  reponame="migrate-single-local-branch-with-symlink"

  remove_and_create_local_repo "$reponame"

  base64 < /dev/urandom | head -c 120 > a.txt

  git add a.txt
  git commit -m "initial commit"

  add_symlink "a.txt" "link.txt"
  git commit -m "add symlink"
}

# make_bare converts the existing full checkout of a repository into a bare one,
# and then `cd`'s into it.
make_bare() {
  reponame=$(basename "$(pwd)")
  mv .git "../$reponame.git"

  cd ..

  rm -rf "$reponame"
  cd "$reponame.git"

  git config --bool core.bare true
}

# remove_and_create_local_repo removes, creates, and checks out a local
# repository given by a particular name:
#
#   remove_and_create_local_repo "$reponame"
remove_and_create_local_repo() {
  local reponame="$(base64 < /dev/urandom | head -c 8 | $SHASUM | cut -f 1 -d ' ')-$1"

  git init "$reponame"
  cd "$reponame"
}

# remove_and_create_remote_repo removes, creates, and checks out a remote
# repository both locally and on the gitserver, given by a particular name:
#
#   remove_and_create_remote_repo "$reponame"
remove_and_create_remote_repo() {
  local reponame="$(base64 < /dev/urandom | head -c 8 | $SHASUM | cut -f 1 -d ' ')-$1"

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"
}
