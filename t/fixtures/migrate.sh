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
#         refs/heads/main
#
# - Commit 'A' has 120, in a.txt, and a corresponding entry in .gitattributes.
#
#   If "0755" is passed as an argument, the .gitattributes file is created
#   with that permissions mode.
#   If "link" is passed as an argument, the .gitattributes file is created
#   as a symlink to a gitattrs file.
setup_local_branch_with_gitattrs() {
  set -e

  reponame="migrate-single-local-branch-with-attrs"

  remove_and_create_local_repo "$reponame"

  lfstest-genrandom --base64 120 >a.txt

  git add a.txt
  git commit -m "initial commit"

  git lfs track "*.txt"
  git lfs track "*.other"

  if [[ $1 == "0755" ]]; then
    chmod +x .gitattributes
  elif [[ $1 == "link" ]]; then
    mv .gitattributes gitattrs

    add_symlink gitattrs .gitattributes

    git add gitattrs
  fi

  git add .gitattributes
  git commit -m "add .gitattributes"
}

# setup_local_branch_with_nested_gitattrs creates a repository as follows:
#
#   A---B
#        \
#         refs/heads/main
#
# - Commit 'A' has 120, in a.txt, and a corresponding entry in .gitattributes. There is also
#   140 in a.md, with no corresponding entry in .gitattributes.
#   It also has 140 in subtree/a.md, and a corresponding entry in subtree/.gitattributes
setup_local_branch_with_nested_gitattrs() {
  set -e

  reponame="migrate-single-local-branch-nested-attrs"

  remove_and_create_local_repo "$reponame"

  mkdir b

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md
  lfstest-genrandom --base64 140 >b/a.md

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

# setup_single_local_branch_untracked creates a repository as follows:
#
#   A---B
#        \
#         refs/heads/main
#
# - Commit 'A' has 120, in a.txt and 140 in a.md, with neither files tracked as
#   pointers in Git LFS
setup_single_local_branch_untracked() {
  set -e

  local name="${1:-a.md}"

  reponame="migrate-single-local-branch-untracked"

  remove_and_create_local_repo "$reponame"

  git commit --allow-empty -m "initial commit"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >"$name"

  git add a.txt "$name"
  git commit -m "add a.txt and $name"
}

# setup_single_local_branch_tracked creates a repository as follows:
#
#   A---B
#        \
#         refs/heads/main
#
# - Commit 'A' has 120, in a.txt and 140 in a.md, with both files tracked as
#   pointers in Git LFS
#
#   If "0755" is passed as an argument, the .gitattributes file is created
#   with that permissions mode.
#   If "link" is passed as an argument, the .gitattributes file is created
#   as a symlink to a gitattrs file.
setup_single_local_branch_tracked() {
  set -e

  reponame="migrate-single-local-branch-tracked"

  remove_and_create_local_repo "$reponame"

  echo "*.txt filter=lfs diff=lfs merge=lfs -text" > .gitattributes
  echo "*.md filter=lfs diff=lfs merge=lfs -text" >> .gitattributes

  if [[ $1 == "0755" ]]; then
    chmod +x .gitattributes
  fi

  git add .gitattributes
  git commit -m "initial commit"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md

  git add a.txt a.md
  git commit -m "add a.{txt,md}"

  if [[ $1 == "link" ]]; then
    git mv .gitattributes gitattrs

    add_symlink gitattrs .gitattributes

    git commit -m "link .gitattributes"
  fi
}

# setup_single_local_branch_complex_tracked creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 1 byte of text in a.txt and dir/b.txt. According to the
#   .gitattributes files, a.txt should be tracked using Git LFS, but b.txt should
#   not be.
setup_single_local_branch_complex_tracked() {
  set -e

  reponame="migrate-single-local-branch-complex-tracked"

  remove_and_create_local_repo "$reponame"

  mkdir -p dir
  echo "*.txt filter=lfs diff=lfs merge=lfs -text" > .gitattributes
  echo "*.txt !filter !diff !merge" > dir/.gitattributes

  printf "a" > a.txt
  printf "b" > dir/b.txt

  git lfs uninstall

  git add .gitattributes dir/.gitattributes a.txt dir/b.txt
  git commit -m "initial commit"

  git lfs install
}

# setup_single_local_branch_tracked_corrupt creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 120 bytes of random data in a.txt, and tracks *.txt under Git
#   LFS, but a.txt is not stored as an LFS object.
#
#   If "lfsmacro" is passed as an argument, a macro attribute definition
#   which sets the LFS filter attribute is added to the .gitattributes file,
#   and then referenced by the test file pattern attribute.
#   If "macro" is passed as an argument, a macro attribute definition is
#   added to the .gitattributes file.
#   If "link" is passed as an argument, the .gitattributes file is created
#   as a symlink to a gitattrs file.
setup_single_local_branch_tracked_corrupt() {
  set -e

  reponame="migrate-single-local-branch-with-attrs-corrupt"

  remove_and_create_local_repo "$reponame"

  git lfs uninstall

  lfstest-genrandom --base64 120 >a.txt

  if [[ $1 == "lfsmacro" ]]; then
    printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n*.txt lfs\n' \
      >.gitattributes
  else
    echo "*.txt filter=lfs diff=lfs merge=lfs -text" > .gitattributes

    if [[ $1 == "macro" ]]; then
      echo "[attr]foo foo" >>.gitattributes
    elif [[ $1 == "link" ]]; then
      mv .gitattributes gitattrs

      add_symlink gitattrs .gitattributes
    fi
  fi

  git add .gitattributes a.txt
  git commit -m "initial commit"

  git lfs install
}

# setup_multiple_local_branches creates a repository as follows:
#
#     B
#    / \
#   A   refs/heads/my-feature
#    \
#     refs/heads/main
#
# - Commit 'A' has 120, 140 bytes of data in a.txt, and a.md, respectively.
#
# - Commit 'B' has 30 bytes of data in a.md, and includes commit 'A' as a
#   parent.
setup_multiple_local_branches() {
  set -e

  reponame="migrate-info-multiple-local-branches"

  remove_and_create_local_repo "$reponame"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md

  git add a.txt a.md
  git commit -m "initial commit"

  git checkout -b my-feature

  lfstest-genrandom --base64 30 >a.md

  git add a.md
  git commit -m "add an additional 30 bytes to a.md"

  git checkout main
}

# setup_multiple_local_branches_with_alternate_names performs the same task
# as setup_multiple_local_branches, but creates a file with no extension.
setup_multiple_local_branches_with_alternate_names() {
  set -e

  reponame="migrate-info-multiple-local-branches"

  remove_and_create_local_repo "$reponame"

  lfstest-genrandom --base64 120 >no_extension
  lfstest-genrandom --base64 140 >a.txt

  git add no_extension a.txt
  git commit -m "initial commit"

  git checkout -b my-feature

  lfstest-genrandom --base64 30 >a.txt
  lfstest-genrandom --base64 100 >no_extension

  git add no_extension a.txt
  git commit -m "add an additional 30 bytes to a.txt"

  git checkout main
}

# setup_multiple_local_branches_with_gitattrs creates a repository in the same way
# as setup_multiple_local_branches, but also adds relevant lfs filters to the
# .gitattributes file in the main branch
setup_multiple_local_branches_with_gitattrs() {
  set -e

  setup_multiple_local_branches

  git lfs track *.txt
  git lfs track *.md

  git add .gitattributes
  git commit -m "add .gitattributes"
}

# setup_multiple_local_branches_non_standard creates a repository as follows:
#
#      refs/pull/1/head
#     /
#     |
#     B
#    / \
#   A   refs/heads/my-feature
#   |\
#   | refs/heads/main
#    \
#     refs/pull/1/base
#
# With the same contents in 'A' and 'B' as setup_multiple_local_branches.
setup_multiple_local_branches_non_standard() {
  set -e

  setup_multiple_local_branches

  git update-ref refs/pull/1/head "$(git rev-parse my-feature)"
  git update-ref refs/pull/1/base "$(git rev-parse main)"
}

# setup_multiple_local_branches_tracked creates a repo with exactly the same
# structure as in setup_multiple_local_branches, but with all files tracked by
# Git LFS
setup_multiple_local_branches_tracked() {
  set -e

  reponame="migrate-info-multiple-local-branches"

  remove_and_create_local_repo "$reponame"

  echo "*.txt filter=lfs diff=lfs merge=lfs -text" > .gitattributes
  echo "*.md filter=lfs diff=lfs merge=lfs -text" >> .gitattributes
  git add .gitattributes
  git commit -m "initial commit"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md

  git add a.txt a.md
  git commit -m "add a.{txt,md}"

  git checkout -b my-feature

  lfstest-genrandom --base64 30 >a.md

  git add a.md
  git commit -m "add an additional 30 bytes to a.md"

  git checkout main
}

# setup_local_branch_with_space creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 50 bytes in a file named "a file.txt".
setup_local_branch_with_space() {
  set -e

  reponame="migrate-local-branch-with-space"
  filename="a file.txt"

  remove_and_create_local_repo "$reponame"

  lfstest-genrandom --base64 50 >"$filename"

  git add "$filename"
  git commit -m "initial commit"
}

# setup_single_remote_branch creates a repository as follows:
#
#   A---B
#    \   \
#     \   refs/heads/main
#      \
#       refs/remotes/origin/main
#
# - Commit 'A' has 120, 140 bytes of data in a.txt, and a.md, respectively. It
#   is the latest commit pushed to the remote 'origin'.
#
# - Commit 'B' has 30, 50 bytes of data in a.txt, and a.md, respectively.
setup_single_remote_branch() {
  set -e

  reponame="migrate-info-single-remote-branch"

  remove_and_create_remote_repo "$reponame"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md

  git add a.txt a.md
  git commit -m "initial commit"

  git push origin main

  lfstest-genrandom --base64 30 >a.txt
  lfstest-genrandom --base64 50 >a.md

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
# *.txt files tracked by Git LFS
setup_single_remote_branch_tracked() {
  set -e

  reponame="migrate-info-single-remote-branch"

  remove_and_create_remote_repo "$reponame"

  git lfs track "*.md" "*.txt"
  git add .gitattributes
  git commit -m "initial commit"

  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 140 >a.md

  git add a.txt a.md
  git commit -m "add a.{txt,md}"

  git push origin main

  lfstest-genrandom --base64 30 >a.txt
  lfstest-genrandom --base64 50 >a.md

  git add a.md a.txt
  git commit -m "add an additional 30, 50 bytes to a.{txt,md}"
}

# setup_multiple_remote_branches creates a repository as follows:
#
#         C
#        / \
#   A---B   refs/heads/my-feature
#    \   \
#     \   refs/heads/main
#      \
#       refs/remotes/origin/main
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

  lfstest-genrandom --base64 10 >a.txt
  lfstest-genrandom --base64 11 >a.md
  git add a.txt a.md
  git commit -m "add 10, 11 bytes, a.{txt,md}"

  git push origin main

  lfstest-genrandom --base64 20 >a.txt
  lfstest-genrandom --base64 21 >a.md
  git add a.txt a.md
  git commit -m "add 20, 21 bytes, a.{txt,md}"

  git checkout -b my-feature

  lfstest-genrandom --base64 30 >a.txt
  lfstest-genrandom --base64 31 >a.md
  git add a.txt a.md
  git commit -m "add 30, 31 bytes, a.{txt,md}"

  git checkout main
}

# Creates a repo identical to that in setup_multiple_remote_branches(), but
# with all files tracked by Git LFS
setup_multiple_remote_branches_gitattrs() {
  set -e

  reponame="migrate-info-exclude-remote-refs-given-branch"

  remove_and_create_remote_repo "$reponame"

  git lfs track "*.txt" "*.md"
  git add .gitattributes
  git commit -m "initial commit"

  lfstest-genrandom --base64 10 >a.txt
  lfstest-genrandom --base64 11 >a.md
  git add a.txt a.md
  git commit -m "add 10, 11 bytes, a.{txt,md}"

  git push origin main

  lfstest-genrandom --base64 20 >a.txt
  lfstest-genrandom --base64 21 >a.md
  git add a.txt a.md
  git commit -m "add 20, 21 bytes, a.{txt,md}"

  git checkout -b my-feature

  lfstest-genrandom --base64 30 >a.txt
  lfstest-genrandom --base64 31 >a.md
  git add a.txt a.md
  git commit -m "add 30, 31 bytes, a.{txt,md}"

  git checkout main
}

# setup_single_local_branch_with_tags creates a repository as follows:
#
#   A---B
#       |\
#       | refs/heads/main
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

  lfstest-genrandom --base64 1 >a.txt

  git add a.txt
  git commit -m "initial commit"

  lfstest-genrandom --base64 2 >a.txt

  git add a.txt
  git commit -m "secondary commit"
  git tag "v1.0.0"
}

# setup_single_local_branch_with_annotated_tags creates a repository as follows:
#
#   A---B
#       |\
#       | refs/heads/main
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

  lfstest-genrandom --base64 1 >a.txt

  git add a.txt
  git commit -m "initial commit"

  lfstest-genrandom --base64 2 >a.txt

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

  lfstest-genrandom --base64 16 >a.txt
  git add a.txt
  git commit -m "initial commit"
  git push origin main

  lfstest-genrandom --base64 16 >a.txt
  git add a.txt
  git commit -m "another commit"
  git push fork main
}

# setup_single_local_branch_deep_trees creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 120 bytes of data in 'foo/bar/baz/a.txt'.
setup_single_local_branch_deep_trees() {
  set -e

  reponame="migrate-single-local-branch-with-deep-trees"
  remove_and_create_local_repo "$reponame"

  mkdir -p foo/bar/baz
  lfstest-genrandom --base64 120 >foo/bar/baz/a.txt

  git add foo/bar/baz/a.txt
  git commit -m "initial commit"
}

# setup_single_local_branch_same_file_tree_ext creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 120 bytes of data in each of 'a.txt`, `foo/a.txt',
#   `bar.txt/b.md`, and `bar.txt/b.txt`.
setup_single_local_branch_same_file_tree_ext() {
  set -e

  reponame="migrate-single-local-branch-with-same-file-tree-ext"
  remove_and_create_local_repo "$reponame"

  mkdir -p foo bar.txt
  lfstest-genrandom --base64 120 >a.txt
  lfstest-genrandom --base64 120 >foo/a.txt
  lfstest-genrandom --base64 120 >bar.txt/b.md
  lfstest-genrandom --base64 120 >bar.txt/b.txt

  git add a.txt foo bar.txt
  git commit -m "initial commit"
}

# setup_local_branch_with_symlink creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has 120, in a.txt, and a symbolic link link.txt to a.txt.
setup_local_branch_with_symlink() {
  set -e

  reponame="migrate-single-local-branch-with-symlink"

  remove_and_create_local_repo "$reponame"

  lfstest-genrandom --base64 120 >a.txt

  git add a.txt
  git commit -m "initial commit"

  add_symlink "a.txt" "link.txt"
  git commit -m "add symlink"
}

# setup_local_branch_with_dirty_copy creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has the contents "a.txt in a.txt, and marks a.txt as unclean
#   in the working copy.
setup_local_branch_with_dirty_copy() {
  set -e

  reponame="migrate-single-local-branch-with-dirty-copy"
  remove_and_create_local_repo "$reponame"

  printf "a.txt" > a.txt

  git add a.txt
  git commit -m "initial commit"

  printf "2" >> a.txt
}

# setup_local_branch_with_copied_file creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has the contents "a.txt" in a.txt, and another identical file
# (same name and content) in another directory.
setup_local_branch_with_copied_file() {
  set -e

  reponame="migrate-single-local-branch-with-copied-file"
  remove_and_create_local_repo "$reponame"

  printf "a.txt" > a.txt
  mkdir dir
  cp a.txt dir/

  git add a.txt dir/a.txt
  git commit -m "initial commit"
}

# setup_local_branch_with_special_character_files creates a repository as follows:
#
#   A
#    \
#     refs/heads/main
#
# - Commit 'A' has binary files with special characters
setup_local_branch_with_special_character_files() {
  set -e

  reponame="migrate-single-local-branch-with-special-filenames"
  remove_and_create_local_repo "$reponame"

  lfstest-genrandom 80 >'./test - special.bin'
  lfstest-genrandom 100 >'./test (test2) special.bin'
  # Windows does not allow creation of files with '*'
  [ "$IS_WINDOWS" -eq '1' ] || lfstest-genrandom 120 >'./test * ** special.bin'

  git add *.bin
  git commit -m "initial commit"
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
  local reponame="$1-$(lfstest-genrandom --base64url 32)"

  git init "$reponame"
  cd "$reponame"
}

# remove_and_create_remote_repo removes, creates, and checks out a remote
# repository both locally and on the gitserver, given by a particular name:
#
#   remove_and_create_remote_repo "$reponame"
remove_and_create_remote_repo() {
  local reponame="$1-$(lfstest-genrandom --base64url 32)"

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  rm clone.log
}
