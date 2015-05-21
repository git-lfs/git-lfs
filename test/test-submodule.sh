#!/bin/sh

. "test/testlib.sh"

begin_test "submodule local git dir"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  submodname="$reponame-submodule"

  setup_remote_repo "$reponame"
  setup_remote_repo "$submodname"

  clone_repo "$submodname" submod
  mkdir dir
  echo "sub module" > dir/README
  git add dir/README
  git commit -a -m "submodule readme"
  git push origin master

  clone_repo "$reponame" repo
  git submodule add "$GITSERVER/$submodname" sub
  cd sub/dir
  cat README | grep "sub module"
  git lfs help
)
end_test
