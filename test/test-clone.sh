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
