#!/usr/bin/env bash

. "test/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.2.0"

begin_test "clone"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  # generate some test data & commits with random LFS data
  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":100},
      {\"Filename\":\"file2.dat\",\"Size\":75}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":110},
      {\"Filename\":\"file3.dat\",\"Size\":66},
      {\"Filename\":\"file4.dat\",\"Size\":23}]
  },
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file5.dat\",\"Size\":120},
      {\"Filename\":\"file6.dat\",\"Size\":30}]
  }
  ]" | lfstest-testutils addcommits

  git push origin master

  # Now clone again, test specific clone dir
  cd "$TRASHDIR"

  newclonedir="testclone1"
  git lfs clone "$GITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Git LFS:" lfsclone.log
  # should be no filter errors
  [ ! $(grep "filter" lfsclone.log) ]
  [ ! $(grep "error" lfsclone.log) ]
  # should be cloned into location as per arg
  [ -d "$newclonedir" ]

  # check a few file sizes to make sure pulled
  pushd "$newclonedir"
  [ $(wc -c < "file1.dat") -eq 110 ] 
  [ $(wc -c < "file2.dat") -eq 75 ] 
  [ $(wc -c < "file3.dat") -eq 66 ] 
  [ ! -e "lfs" ]
  popd
  # Now check clone with implied dir
  rm -rf "$reponame"
  git lfs clone "$GITSERVER/$reponame" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Git LFS:" lfsclone.log
  # should be no filter errors
  [ ! $(grep "filter" lfsclone.log) ]
  [ ! $(grep "error" lfsclone.log) ]
  # clone location should be implied
  [ -d "$reponame" ]
  pushd "$reponame"
  [ $(wc -c < "file1.dat") -eq 110 ] 
  [ $(wc -c < "file2.dat") -eq 75 ] 
  [ $(wc -c < "file3.dat") -eq 66 ] 
  [ ! -e "lfs" ]
  popd

)
end_test

begin_test "cloneSSL"
(
  set -e
  if $TRAVIS; then
    echo "Skipping SSL tests, Travis has weird behaviour in validating custom certs, test locally only"
    exit 0
  fi

  reponame="test-cloneSSL"
  setup_remote_repo "$reponame"
  clone_repo_ssl "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  # generate some test data & commits with random LFS data
  echo "[
  {
    \"CommitDate\":\"$(get_date -5d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":100},
      {\"Filename\":\"file2.dat\",\"Size\":75}]
  },
  {
    \"CommitDate\":\"$(get_date -1d)\",
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":30}]
  }
  ]" | lfstest-testutils addcommits

  git push origin master

  # Now SSL clone again with 'git lfs clone', test specific clone dir
  cd "$TRASHDIR"

  newclonedir="testcloneSSL1"
  git lfs clone "$SSLGITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Git LFS:" lfsclone.log
  # should be no filter errors
  [ ! $(grep "filter" lfsclone.log) ]
  [ ! $(grep "error" lfsclone.log) ]
  # should be cloned into location as per arg
  [ -d "$newclonedir" ]

  # check a few file sizes to make sure pulled
  pushd "$newclonedir"
  [ $(wc -c < "file1.dat") -eq 100 ] 
  [ $(wc -c < "file2.dat") -eq 75 ] 
  [ $(wc -c < "file3.dat") -eq 30 ] 
  popd


  # Now check SSL clone with standard 'git clone' and smudge download
  rm -rf "$reponame"
  git clone "$SSLGITSERVER/$reponame"

)
end_test


begin_test "clone with flags"
(
  set -e

  reponame="$(basename "$0" ".sh")-flags"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  # generate some test data & commits with random LFS data
  echo "[
  {
    \"CommitDate\":\"$(get_date -10d)\",
    \"Files\":[
      {\"Filename\":\"file1.dat\",\"Size\":100},
      {\"Filename\":\"file2.dat\",\"Size\":75}]
  },
  {
    \"CommitDate\":\"$(get_date -7d)\",
    \"NewBranch\":\"branch2\",    
    \"Files\":[
      {\"Filename\":\"fileonbranch2.dat\",\"Size\":66}]
  },
  {
    \"CommitDate\":\"$(get_date -3d)\",
    \"ParentBranches\":[\"master\"],
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":120},
      {\"Filename\":\"file4.dat\",\"Size\":30}]
  }
  ]" | lfstest-testutils addcommits

  git push origin master branch2

  # Now clone again, test specific clone dir
  cd "$TRASHDIR"
  mkdir "$TRASHDIR/templatedir"

  newclonedir="testflagsclone1"
  # many of these flags won't do anything but make sure they're not rejected
  git lfs clone --template "$TRASHDIR/templatedir" --local --no-hardlinks --shared --verbose --progress --recursive "$GITSERVER/$reponame" "$newclonedir"
  rm -rf "$newclonedir"

  # specific test for --no-checkout
  git lfs clone --quiet --no-checkout "$GITSERVER/$reponame" "$newclonedir"
  if [ -e "$newclonedir/file1.dat" ]; then
    exit 1
  fi
  rm -rf "$newclonedir"

  # specific test for --branch and --origin
  git lfs clone --branch branch2 --recurse-submodules --origin differentorigin "$GITSERVER/$reponame" "$newclonedir"
  pushd "$newclonedir"
  # this file is only on branch2
  [ -e "fileonbranch2.dat" ]
  # confirm remote is called differentorigin
  git remote get-url differentorigin
  popd
  rm -rf "$newclonedir"

  # specific test for --separate-git-dir
  gitdir="$TRASHDIR/separategitdir"
  git lfs clone --separate-git-dir "$gitdir" "$GITSERVER/$reponame" "$newclonedir"
  # .git should be a file not dir
  if [ -d "$newclonedir/.git" ]; then
    exit 1
  fi
  [ -e "$newclonedir/.git" ]
  [ -d "$gitdir/objects" ]
  rm -rf "$newclonedir"
  rm -rf "$gitdir"

  # specific test for --bare
  git lfs clone --bare "$GITSERVER/$reponame" "$newclonedir"
  [ -d "$newclonedir/objects" ]  

  # short flags
  git lfs clone -l -v -n -s -b branch2 "$GITSERVER/$reponame" "$newclonedir"
  rm -rf "$newclonedir"

)
end_test

begin_test "clone (with include/exclude args)"
(
  set -e

  reponame="clone_include_exclude"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents_a="a"
  contents_a_oid=$(calc_oid "$contents_a")
  printf "$contents_a" > "a.dat"

  contents_b="b"
  contents_b_oid=$(calc_oid "$contents_b")
  printf "$contents_b" > "b.dat"

  git add a.dat b.dat .gitattributes
  git commit -m "add a.dat, b.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 b.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log
  grep "Git LFS: (2 of 2 files)" push.log

  cd "$TRASHDIR"

  local_reponame="clone_with_includes"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "a.dat"
  pushd "$local_reponame"
  assert_local_object "$contents_a_oid" 1
  refute_local_object "$contents_b_oid"
  popd

  local_reponame="clone_with_excludes"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "b.dat" -X "a.dat"
  pushd "$local_reponame"
  assert_local_object "$contents_b_oid" 1
  refute_local_object "$contents_a_oid"
  popd
)
end_test

begin_test "clone (with .lfsconfig)"
(
  set -e

  reponame="clone_with_lfsconfig"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents_a="a"
  contents_a_oid=$(calc_oid "$contents_a")
  printf "$contents_a" > "a.dat"

  contents_b="b"
  contents_b_oid=$(calc_oid "$contents_b")
  printf "$contents_b" > "b.dat"



  git add a.dat b.dat .gitattributes
  git commit -m "add a.dat, b.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 b.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git config -f ".lfsconfig" "lfs.fetchinclude" "a*"
  git add ".lfsconfig"
  git commit -m "config lfs.fetchinclude a*" 2>&1 | tee commit.log
  grep "master" commit.log
  grep "1 file changed" commit.log
  grep "create mode 100644 .lfsconfig" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log
  grep "Git LFS: (2 of 2 files)" push.log

  pushd "$TRASHDIR"

  echo "test: clone with lfs.fetchinclude in .lfsconfig"
  local_reponame="clone_with_config_include"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame"
  pushd "$local_reponame"
  assert_local_object "$contents_a_oid" 1
  refute_local_object "$contents_b_oid"
  popd

  echo "test: clone with lfs.fetchinclude in .lfsconfig, and args"
  local_reponame="clone_with_config_include_and_args"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "b.dat"
  pushd "$local_reponame"
  refute_local_object "$contents_a_oid"
  assert_local_object "$contents_b_oid" 1
  popd

  popd

  git config -f ".lfsconfig" "lfs.fetchinclude" "b*"
  git config -f ".lfsconfig" "lfs.fetchexclude" "a*"
  git add .lfsconfig
  git commit -m "config lfs.fetchinclude a*" 2>&1 | tee commit.log
  grep "master" commit.log
  grep "1 file changed" commit.log
  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log

  pushd "$TRASHDIR"

  echo "test: clone with lfs.fetchexclude in .lfsconfig"
  local_reponame="clone_with_config_exclude"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame"
  pushd "$local_reponame"
  cat ".lfsconfig"
  assert_local_object "$contents_b_oid" 1
  refute_local_object "$contents_a_oid"
  popd

  echo "test: clone with lfs.fetchexclude in .lfsconfig, and args"
  local_reponame="clone_with_config_exclude_and_args"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "a.dat" -X "b.dat"
  pushd "$local_reponame"
  assert_local_object "$contents_a_oid" 1
  refute_local_object "$contents_b_oid"
  popd

  popd
)
end_test

begin_test "clone with submodules"
(
  set -e

  # set up a doubly nested submodule, each with LFS content
  reponame="submod-root"
  submodname1="submod-level1"
  submodname2="submod-level2"

  setup_remote_repo "$reponame"
  setup_remote_repo "$submodname1"
  setup_remote_repo "$submodname2"

  clone_repo "$submodname2" submod2
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents_sub2="Inception. Now, before you bother telling me it's impossible..."
  contents_sub2_oid=$(calc_oid "$contents_sub2")
  printf "$contents_sub2" > "sub2.dat"
  git add sub2.dat .gitattributes 
  git commit -m "Nested submodule level 2"
  git push origin master


  clone_repo "$submodname1" submod1
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents_sub1="We're dreaming?"
  contents_sub1_oid=$(calc_oid "$contents_sub1")
  printf "$contents_sub1" > "sub1.dat"
  # add submodule2 as submodule of submodule1
  git submodule add "$GITSERVER/$submodname2" sub2
  git submodule update
  git add sub2 sub1.dat .gitattributes 
  git commit -m "Nested submodule level 1"
  git push origin master


  clone_repo "$reponame" rootrepo
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents_root="Downwards is the only way forwards."
  contents_root_oid=$(calc_oid "$contents_root")
  printf "$contents_root" > "root.dat"
  # add submodule1 as submodule of root
  git submodule add "$GITSERVER/$submodname1" sub1
  git submodule update
  git add sub1 root.dat .gitattributes 
  git commit -m "Root repo"
  git push origin master

  pushd "$TRASHDIR"

  local_reponame="submod-clone"
  git lfs clone --recursive "$GITSERVER/$reponame" "$local_reponame"

  # check everything is where it should be
  cd $local_reponame
  # check LFS store and working copy
  assert_local_object "$contents_root_oid" "${#contents_root}"
  [ $(wc -c < "root.dat") -eq ${#contents_root} ] 
  # and so on for nested subs
  cd sub1
  assert_local_object "$contents_sub1_oid" "${#contents_sub1}"
  [ $(wc -c < "sub1.dat") -eq ${#contents_sub1} ] 
  cd sub2
  assert_local_object "$contents_sub2_oid" "${#contents_sub2}"
  [ $(wc -c < "sub2.dat") -eq ${#contents_sub2} ] 


  popd


)
end_test

begin_test "clone in current directory"
(
  set -e

  reponame="clone_in_current_dir"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="contents"
  contents_oid="$(calc_oid "$contents")"

  printf "$contents" > a.dat

  git add .gitattributes a.dat

  git commit -m "initial commit" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log

  pushd $TRASHDIR
    mkdir "$reponame-clone"
    cd "$reponame-clone"

    git lfs clone $GITSERVER/$reponame "." 2>&1 | grep "Git LFS"

    assert_local_object "$contents_oid" 8
    [ ! -f ./lfs ]
  popd
)
end_test
