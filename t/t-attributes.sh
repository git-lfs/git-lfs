#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "macros"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  clone_repo "$reponame" repo

  mkdir dir
  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n*.dat lfs\n' \
    > .gitattributes
  printf '[attr]lfs2 filter=lfs diff=lfs merge=lfs -text\n*.bin lfs2\n' \
    > dir/.gitattributes
  git add .gitattributes dir
  git commit -m 'initial import'

  contents="some data"
  printf "$contents" > foo.dat
  git add *.dat
  git commit -m 'foo.dat'
  assert_local_object "$(calc_oid "$contents")" 9

  contents2="other data"
  printf "$contents2" > dir/foo.bin
  git add dir
  git commit -m 'foo.bin'
  refute_local_object "$(calc_oid "$contents2")"

  git lfs track '*.dat' 2>&1 | tee track.log
  grep '"*.dat" already supported' track.log

  cd dir
  git lfs track '*.bin' 2>&1 | tee track.log
  grep '"*.bin" already supported' track.log && exit 1
  true

  # NOTE: At present we do not test that "git lfs track" reports
  #       "already supported" when it finds a pattern in a subdirectory's
  #       .gitattributes file which references a macro attribute in
  #       the top-level .gitattributes file that sets "filter=lfs".
  #       This is because, while "git check-attr" resolves macro references
  #       from a file such as dir/.gitattributes to .gitattributess,
  #       "git lfs track" only resolves macro references as it reads these
  #       files in depth-first order, so unlike Git it does not expand an
  #       "lfs" reference to "filter=lfs" if it appears in dir/.gitattributes.
)
end_test

begin_test "macros with HOME"
(
  set -e

  reponame="$(basename "$0" ".sh")-home"
  clone_repo "$reponame" repo-home

  mkdir -p "$HOME/.config/git"
  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n*.dat lfs\n' \
    > "$HOME/.config/git/attributes"

  contents="some data"
  printf "$contents" > foo.dat
  git add *.dat
  git commit -m 'foo.dat'
  assert_local_object "$(calc_oid "$contents")" 9

  git lfs track 2>&1 | tee track.log
  grep '*.dat' track.log
)
end_test

begin_test "macros with HOME split"
(
  set -e

  reponame="$(basename "$0" ".sh")-home-split"
  clone_repo "$reponame" repo-home-split

  mkdir -p "$HOME/.config/git"
  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n' \
    > "$HOME/.config/git/attributes"

  printf '*.dat lfs\n' > .gitattributes
  git add .gitattributes
  git commit -m 'initial import'

  contents="some data"
  printf "$contents" > foo.dat
  git add *.dat
  git commit -m 'foo.dat'
  assert_local_object "$(calc_oid "$contents")" 9

  git lfs track '*.dat' 2>&1 | tee track.log
  grep '"*.dat" already supported' track.log
)
end_test

begin_test "macros with unspecified flag"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  clone_repo "$reponame" repo-unspecified

  mkdir dir
  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n**/*.dat lfs\n' \
    > .gitattributes
  printf '*.dat !lfs\n' \
    > dir/.gitattributes
  git add .gitattributes dir
  git commit -m 'initial import'

  contents="some data"
  printf "$contents" > foo.dat
  git add *.dat
  git commit -m 'foo.dat'
  assert_local_object "$(calc_oid "$contents")" 9

  contents2="other data"
  printf "$contents2" > dir/foo.dat
  git add dir
  git commit -m 'dir/foo.dat'
  refute_local_object "$(calc_oid "$contents2")"

  git lfs track '**/*.dat' 2>&1 | tee track.log
  grep '"*\*/\*.dat" already supported' track.log

  # NOTE: The intent of this test is to confirm that running the
  #       "git lfs track '*.dat'" command in the dir/ directory returns
  #       "already supported", because it finds the "*.dat" pattern and
  #       resolves its reference to the "lfs" macro attribute in
  #       top-level .gitattributes file such that a "filter" attribute
  #       is recognized, albeit with the unspecified state set.
  #
  #       However, as noted in the "macros" test above, because the
  #       "git lfs track" command parses the dir/.gitattributes file
  #       before the top-level .gitattributes file, it does not resolve
  #       the macro attribute reference, and our test would fail despite
  #       our ability to parse macro attribute references with a "!"
  #       unspecified flag character prefix.

  #cd dir
  #git lfs track '*.dat' 2>&1 | tee track.log
  #grep '"*.dat" already supported' track.log
)
end_test
