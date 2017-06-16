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

# setup_single_remote_branch creates a repository as follows:
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

# remove_and_create_local_repo removes, creates, and checks out a local
# repository given by a particular name:
#
#   remove_and_create_local_repo "$reponame"
remove_and_create_local_repo() {
  local reponame="$1"

  rm -rf "$reponame" || true

  git init "$reponame"
  cd "$reponame"
}

# remove_and_create_remote_repo removes, creates, and checks out a remote
# repository both locally and on the gitserver, given by a particular name:
#
#   remove_and_create_remote_repo "$reponame"
remove_and_create_remote_repo() {
  local reponame="$1"

  rm -rf "$reponame" "$REMOTEDIR/$reponame.git" || true

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"
}
