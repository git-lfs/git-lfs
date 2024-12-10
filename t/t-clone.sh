#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.2.0"

# Check for a libnss3 dependency in Git until we drop support for CentOS 7.
GIT_LIBNSS=0
if [ "$IS_WINDOWS" -eq 0 -a "$IS_MAC" -eq 0 ]; then
  GIT_LIBNSS="$(ldd "$(git --exec-path)"/git-remote-https | grep -c '^\s*libnss3\.' || true)"
fi
export GIT_LIBNSS

export CREDSDIR="$REMOTEDIR/creds-clone"
setup_creds

begin_test "clone"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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

  git push origin main

  # Now clone again, test specific clone dir
  cd "$TRASHDIR"

  newclonedir="testclone1"
  git lfs clone "$GITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Downloading LFS objects:" lfsclone.log
  # should be no filter errors
  grep "filter" lfsclone.log && exit 1
  grep "error" lfsclone.log && exit 1
  # should be cloned into location as per arg
  [ -d "$newclonedir" ]

  # check a few file sizes to make sure pulled
  pushd "$newclonedir"
    [ $(wc -c < "file1.dat") -eq 110 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 66 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd

  # Now check clone with implied dir
  rm -rf "$reponame"
  git lfs clone "$GITSERVER/$reponame" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Downloading LFS objects:" lfsclone.log
  # should be no filter errors
  grep "filter" lfsclone.log && exit 1
  grep "error" lfsclone.log && exit 1
  # clone location should be implied
  [ -d "$reponame" ]

  pushd "$reponame"
    [ $(wc -c < "file1.dat") -eq 110 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 66 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd

  # Now check clone with standard 'git clone' and smudge download
  rm -rf "$reponame"
  git clone "$GITSERVER/$reponame" 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
  grep "Cloning into" clone.log
  [ -d "$reponame" ]

  pushd "$reponame"
    [ $(wc -c < "file1.dat") -eq 110 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 66 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    [ "6" -eq "$(find "$(dot_git_dir)/lfs/objects" -type f | wc -l)" ]
    assert_clean_status
  popd
)
end_test

begin_test "cloneSSL"
(
  set -e

  if [ "$GIT_LIBNSS" -eq 1 ]; then
    echo "skip: libnss does not support the Go httptest server certificate"
    exit 0
  fi

  if [ "$IS_WINDOWS" -eq 1 ]; then
    git config --global "http.sslBackend" "openssl"
  fi

  reponame="test-cloneSSL"
  setup_remote_repo "$reponame"
  clone_repo_ssl "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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

  git push origin main

  # Now SSL clone again with 'git lfs clone', test specific clone dir
  cd "$TRASHDIR"

  newclonedir="testcloneSSL1"
  git lfs clone "$SSLGITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Downloading LFS objects:" lfsclone.log
  # should be no filter errors
  grep "filter" lfsclone.log && exit 1
  grep "error" lfsclone.log && exit 1
  # should be cloned into location as per arg
  [ -d "$newclonedir" ]

  # check a few file sizes to make sure pulled
  pushd "$newclonedir"
    [ $(wc -c < "file1.dat") -eq 100 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 30 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd

  # Now check SSL clone with standard 'git clone' and smudge download
  rm -rf "$reponame"
  git clone "$SSLGITSERVER/$reponame" 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
  grep "Cloning into" clone.log
  [ -d "$reponame" ]

  pushd "$reponame"
    [ $(wc -c < "file1.dat") -eq 100 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 30 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    [ "3" -eq "$(find "$(dot_git_dir)/lfs/objects" -type f | wc -l)" ]
    assert_clean_status
  popd
)
end_test

begin_test "clone ClientCert"
(
  set -e

  if [ "$GIT_LIBNSS" -eq 1 ]; then
    echo "skip: libnss does not support the Go httptest server certificate"
    exit 0
  fi

  if [ "$IS_WINDOWS" -eq 1 ]; then
    git config --global "http.sslBackend" "openssl"
  fi

  # Note that the record files we create in the $CREDSDIR directory are not
  # used until we set the "http.sslCertPasswordProtected" option to "true"
  # and the "http.<url>.sslKey" option with the path to our TLS/SSL client
  # certificate's encrypted private key file.  (The PEM certificate file
  # itself is not encrypted and does not contain the private key.)
  #
  # When these options are set, however, Git and Git LFS will independently
  # invoke "git credential fill" to retrieve the passphrase for the
  # encrypted private key.  Because the "http.sslCertPasswordProtected"
  # option is set, Git will query the credential helper, passing a
  # "protocol=cert" line and a "path=<certfile>" line with the path
  # from the "http.<url>.sslCert" option.  Note that this path refers
  # to our unencrypted certificate file; Git does not use the path to
  # the encrypted private key file from the "http.<url>.sslKey" option
  # in its query to the credential helper.
  #
  # Separately, the Git LFS client will detect that the private key file
  # specified by the "http.<url>.sslKey" option is encrypted, and so will
  # invoke "git credential fill" to retrieve its passphrase, passing a
  # "protocol=cert" line and a "path=<keyfile>" line with the path
  # from the "http.<url>.sslKey" option.
  #
  # In order to satisfy both requests, our git-credential-lfstest helper
  # therefore needs two record files, both with the passphrase for the
  # encrypted private key file.  For Git, one is associated with the path
  # to the certificate file, and for Git LFS, one is associated with the
  # path to the key file.
  write_creds_file "::pass" "$CREDSDIR/--$(echo "$LFS_CLIENT_CERT_FILE" | tr / -)"
  write_creds_file "::pass" "$CREDSDIR/--$(echo "$LFS_CLIENT_KEY_FILE_ENCRYPTED" | tr / -)"

  git config --global "http.$LFS_CLIENT_CERT_URL/.sslCert" "$LFS_CLIENT_CERT_FILE"
  git config --global "http.$LFS_CLIENT_CERT_URL/.sslKey" "$LFS_CLIENT_KEY_FILE"

  reponame="test-cloneClientCert"
  setup_remote_repo "$reponame"
  clone_repo_clientcert "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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

  git push origin main

  # Now clone again with 'git lfs clone', test specific clone dir
  # Test with both unencrypted and encrypted client certificate keys
  cd "$TRASHDIR"

  for enc in "false" "true"; do
    if [ "$enc" = "true" ]; then
      git config --global "http.$LFS_CLIENT_CERT_URL/.sslKey" "$LFS_CLIENT_KEY_FILE_ENCRYPTED"
      git config --global "http.sslCertPasswordProtected" "$enc"
    fi

    newclonedir="${reponame}-${enc}"
    git lfs clone "$CLIENTCERTGITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
    grep "Cloning into" lfsclone.log
    grep "Downloading LFS objects:" lfsclone.log
    # should be no filter errors
    grep "filter" lfsclone.log && exit 1
    grep "error" lfsclone.log && exit 1
    # should be cloned into location as per arg
    [ -d "$newclonedir" ]

    # check a few file sizes to make sure pulled
    pushd "$newclonedir"
      [ $(wc -c < "file1.dat") -eq 100 ]
      [ $(wc -c < "file2.dat") -eq 75 ]
      [ $(wc -c < "file3.dat") -eq 30 ]
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

    # Now check clone with standard 'git clone' and smudge download
    rm -rf "$reponame"
    git clone "$CLIENTCERTGITSERVER/$reponame" 2>&1 | tee clone.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected clone to succeed ..."
      exit 1
    fi
    grep "Cloning into" clone.log
    [ -d "$reponame" ]

    pushd "$reponame"
      [ $(wc -c < "file1.dat") -eq 100 ]
      [ $(wc -c < "file2.dat") -eq 75 ]
      [ $(wc -c < "file3.dat") -eq 30 ]
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      [ "3" -eq "$(find "$(dot_git_dir)/lfs/objects" -type f | wc -l)" ]
      assert_clean_status
    popd
  done
)
end_test

begin_test "clone ClientCert with homedir certs"
(
  set -e

  if [ "$GIT_LIBNSS" -eq 1 ]; then
    echo "skip: libnss does not support the Go httptest server certificate"
    exit 0
  fi

  if [ "$IS_WINDOWS" -eq 1 ]; then
    git config --global "http.sslBackend" "openssl"
  fi

  cp "$LFS_CLIENT_CERT_FILE" "$HOME/lfs-client-cert-file"
  cp "$LFS_CLIENT_KEY_FILE" "$HOME/lfs-client-key-file"
  cp "$LFS_CLIENT_KEY_FILE_ENCRYPTED" "$HOME/lfs-client-key-file-encrypted"

  # Note that the record files we create in the $CREDSDIR directory are not
  # used until we set the "http.sslCertPasswordProtected" option to "true"
  # and the "http.<url>.sslKey" option with the path to our TLS/SSL client
  # certificate's encrypted private key file.  (The PEM certificate file
  # itself is not encrypted and does not contain the private key.)
  #
  # When these options are set, however, Git and Git LFS will independently
  # invoke "git credential fill" to retrieve the passphrase for the
  # encrypted private key.  Because the "http.sslCertPasswordProtected"
  # option is set, Git will query the credential helper, passing a
  # "protocol=cert" line and a "path=<certfile>" line with the path
  # from the "http.<url>.sslCert" option.  Note that this path refers
  # to our unencrypted certificate file; Git does not use the path to
  # the encrypted private key file from the "http.<url>.sslKey" option
  # in its query to the credential helper.
  #
  # Separately, the Git LFS client will detect that the private key file
  # specified by the "http.<url>.sslKey" option is encrypted, and so will
  # invoke "git credential fill" to retrieve its passphrase, passing a
  # "protocol=cert" line and a "path=<keyfile>" line with the path
  # from the "http.<url>.sslKey" option.
  #
  # In order to satisfy both requests, our git-credential-lfstest helper
  # therefore needs two record files, both with the passphrase for the
  # encrypted private key file.  For Git, one is associated with the path
  # to the certificate file, and for Git LFS, one is associated with the
  # path to the key file.
  if [ "$IS_WINDOWS" -eq 1 ]; then
    # In our MSYS2 CI environment we have to convert the Unix-style path
    # in $HOME, which starts with /tmp/, into a path of the form /a/...
    # so that the credential record filename we create from it matches
    # the one our git-credential-lfstest helper will construct from the
    # "path" values it receives from Git and Git LFS.
    homedir="$(cygpath -m "$HOME" | sed 's,^\([A-Z]\):,/\L\1,')"
  else
    homedir="$HOME"
  fi
  write_creds_file "::pass" "$CREDSDIR/--$(echo "$homedir/lfs-client-cert-file" | tr / -)"
  write_creds_file "::pass" "$CREDSDIR/--$(echo "$homedir/lfs-client-key-file-encrypted" | tr / -)"

  git config --global "http.$LFS_CLIENT_CERT_URL/.sslCert" "~/lfs-client-cert-file"
  git config --global "http.$LFS_CLIENT_CERT_URL/.sslKey" "~/lfs-client-key-file"

  reponame="test-cloneClientCert-homedir"
  setup_remote_repo "$reponame"
  clone_repo_clientcert "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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

  git push origin main

  # Now clone again with 'git lfs clone', test specific clone dir
  # Test with both unencrypted and encrypted client certificate keys
  cd "$TRASHDIR"

  for enc in "false" "true"; do
    if [ "$enc" = "true" ]; then
      git config --global "http.$LFS_CLIENT_CERT_URL/.sslKey" "~/lfs-client-key-file-encrypted"
      git config --global "http.sslCertPasswordProtected" "$enc"
    fi

    newclonedir="${reponame}-${enc}"
    git lfs clone "$CLIENTCERTGITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
    grep "Cloning into" lfsclone.log
    grep "Downloading LFS objects:" lfsclone.log
    # should be no filter errors
    grep "filter" lfsclone.log && exit 1
    grep "error" lfsclone.log && exit 1
    # should be cloned into location as per arg
    [ -d "$newclonedir" ]

    # check a few file sizes to make sure pulled
    pushd "$newclonedir"
      [ $(wc -c < "file1.dat") -eq 100 ]
      [ $(wc -c < "file2.dat") -eq 75 ]
      [ $(wc -c < "file3.dat") -eq 30 ]
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

    # Now check clone with standard 'git clone' and smudge download
    rm -rf "$reponame"
    git clone "$CLIENTCERTGITSERVER/$reponame" 2>&1 | tee clone.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected clone to succeed ..."
      exit 1
    fi
    grep "Cloning into" clone.log
    [ -d "$reponame" ]

    pushd "$reponame"
      [ $(wc -c < "file1.dat") -eq 100 ]
      [ $(wc -c < "file2.dat") -eq 75 ]
      [ $(wc -c < "file3.dat") -eq 30 ]
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      [ "3" -eq "$(find "$(dot_git_dir)/lfs/objects" -type f | wc -l)" ]
      assert_clean_status
    popd
  done
)
end_test

begin_test "clone with flags"
(
  set -e

  reponame="$(basename "$0" ".sh")-flags"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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
    \"ParentBranches\":[\"main\"],
    \"Files\":[
      {\"Filename\":\"file3.dat\",\"Size\":120},
      {\"Filename\":\"file4.dat\",\"Size\":30}]
  }
  ]" | lfstest-testutils addcommits

  git push origin main branch2

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
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
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
  assert_hooks "$gitdir"
  pushd "$newclonedir"
    [ ! -e "lfs" ]
    assert_clean_status
  popd
  rm -rf "$newclonedir"
  rm -rf "$gitdir"

  # specific test for --bare
  git lfs clone --bare "$GITSERVER/$reponame" "$newclonedir"
  [ -d "$newclonedir/objects" ]
  rm -rf "$newclonedir"

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
  grep "Tracking \"\*.dat\"" track.log

  contents_a="a"
  contents_a_oid=$(calc_oid "$contents_a")
  printf "%s" "$contents_a" > "a.dat"
  printf "%s" "$contents_a" > "a-dupe.dat"
  printf "%s" "$contents_a" > "dupe-a.dat"

  contents_b="b"
  contents_b_oid=$(calc_oid "$contents_b")
  printf "%s" "$contents_b" > "b.dat"

  git add *.dat .gitattributes
  git commit -m "add a.dat, b.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "5 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 a-dupe.dat" commit.log
  grep "create mode 100644 dupe-a.dat" commit.log
  grep "create mode 100644 b.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin main 2>&1 | tee push.log
  grep "main -> main" push.log
  grep "Uploading LFS objects: 100% (2/2), 2 B" push.log

  cd "$TRASHDIR"

  local_reponame="clone_with_includes"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "a*.dat"
  pushd "$local_reponame"
    assert_local_object "$contents_a_oid" 1
    refute_local_object "$contents_b_oid"
    [ "a" = "$(cat a.dat)" ]
    [ "a" = "$(cat a-dupe.dat)" ]
    [ "$(pointer $contents_a_oid 1)" = "$(cat dupe-a.dat)" ]
    [ "$(pointer $contents_b_oid 1)" = "$(cat b.dat)" ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd

  local_reponame="clone_with_excludes"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "b.dat" -X "a.dat"
  pushd "$local_reponame"
    assert_local_object "$contents_b_oid" 1
    refute_local_object "$contents_a_oid"
    [ "$(pointer $contents_a_oid 1)" = "$(cat a.dat)" ]
    [ "b" = "$(cat b.dat)" ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
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
  grep "Tracking \"\*.dat\"" track.log

  contents_a="a"
  contents_a_oid=$(calc_oid "$contents_a")
  printf "%s" "$contents_a" > "a.dat"

  contents_b="b"
  contents_b_oid=$(calc_oid "$contents_b")
  printf "%s" "$contents_b" > "b.dat"

  git add a.dat b.dat .gitattributes
  git commit -m "add a.dat, b.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 b.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git config -f ".lfsconfig" "lfs.fetchinclude" "a*"
  git add ".lfsconfig"
  git commit -m "config lfs.fetchinclude a*" 2>&1 | tee commit.log
  grep "main" commit.log
  grep "1 file changed" commit.log
  grep "create mode 100644 .lfsconfig" commit.log

  git push origin main 2>&1 | tee push.log
  grep "main -> main" push.log
  grep "Uploading LFS objects: 100% (2/2), 2 B" push.log

  pushd "$TRASHDIR"

    echo "test: clone with lfs.fetchinclude in .lfsconfig"
    local_reponame="clone_with_config_include"
    set +x
    git lfs clone "$GITSERVER/$reponame" "$local_reponame"
    ok="$?"
    set -x
    if [ "0" -ne "$ok" ]; then
      # TEMP: used to catch transient failure from above `clone` command, as in:
      # https://github.com/git-lfs/git-lfs/pull/1782#issuecomment-267678319
      echo >&2 "[!] \`git lfs clone $GITSERVER/$reponame $local_reponame\` failed"
      git lfs logs last

      exit 1
    fi
    pushd "$local_reponame"
      assert_local_object "$contents_a_oid" 1
      refute_local_object "$contents_b_oid"
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

    echo "test: clone with lfs.fetchinclude in .lfsconfig, and args"
    local_reponame="clone_with_config_include_and_args"
    git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "b.dat"
    pushd "$local_reponame"
      refute_local_object "$contents_a_oid"
      assert_local_object "$contents_b_oid" 1
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

  popd

  git config -f ".lfsconfig" "lfs.fetchinclude" "b*"
  git config -f ".lfsconfig" "lfs.fetchexclude" "a*"
  git add .lfsconfig
  git commit -m "config lfs.fetchinclude a*" 2>&1 | tee commit.log
  grep "main" commit.log
  grep "1 file changed" commit.log
  git push origin main 2>&1 | tee push.log
  grep "main -> main" push.log

  pushd "$TRASHDIR"

    echo "test: clone with lfs.fetchexclude in .lfsconfig"
    local_reponame="clone_with_config_exclude"
    git lfs clone "$GITSERVER/$reponame" "$local_reponame"
    pushd "$local_reponame"
      cat ".lfsconfig"
      assert_local_object "$contents_b_oid" 1
      refute_local_object "$contents_a_oid"
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

    echo "test: clone with lfs.fetchexclude in .lfsconfig, and args"
    local_reponame="clone_with_config_exclude_and_args"
    git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "a.dat" -X "b.dat"
    pushd "$local_reponame"
      assert_local_object "$contents_a_oid" 1
      refute_local_object "$contents_b_oid"
      assert_hooks "$(dot_git_dir)"
      [ ! -e "lfs" ]
      assert_clean_status
    popd

  popd
)
end_test

begin_test "clone (without clean filter)"
(
  set -e

  reponame="clone_with_clean"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents_a="a"
  contents_a_oid=$(calc_oid "$contents_a")
  printf "%s" "$contents_a" > "a.dat"

  git add *.dat .gitattributes
  git commit -m "add a.dat, b.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log

  git push origin main 2>&1 | tee push.log
  grep "main -> main" push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log

  cd "$TRASHDIR"

  git lfs uninstall
  git config --list > config.txt
  grep "filter.lfs.clean" config.txt && {
    echo "clean filter still configured:"
    cat config.txt
    exit 1
  }

  local_reponame="clone_without_clean"
  git lfs clone "$GITSERVER/$reponame" "$local_reponame" -I "a*.dat" | tee clone.txt
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
  grep "Git LFS is not installed" clone.txt

  cd "$local_reponame"
  assert_local_object "$contents_a_oid" 1
  [ "$(pointer $contents_a_oid 1)" = "$(cat a.dat)" ]
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
  grep "Tracking \"\*.dat\"" track.log

  contents_sub2="Inception. Now, before you bother telling me it's impossible..."
  contents_sub2_oid=$(calc_oid "$contents_sub2")
  printf "%s" "$contents_sub2" > "sub2.dat"
  git add sub2.dat .gitattributes
  git commit -m "Nested submodule level 2"
  git push origin main

  clone_repo "$submodname1" submod1
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents_sub1="We're dreaming?"
  contents_sub1_oid=$(calc_oid "$contents_sub1")
  printf "%s" "$contents_sub1" > "sub1.dat"
  # add submodule2 as submodule of submodule1
  git submodule add "$GITSERVER/$submodname2" sub2
  git submodule update
  git add sub2 sub1.dat .gitattributes
  git commit -m "Nested submodule level 1"
  git push origin main

  clone_repo "$reponame" rootrepo
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents_root="Downwards is the only way forwards."
  contents_root_oid=$(calc_oid "$contents_root")
  printf "%s" "$contents_root" > "root.dat"
  # add submodule1 as submodule of root
  git submodule add "$GITSERVER/$submodname1" sub1
  git submodule update
  git add sub1 root.dat .gitattributes
  git commit -m "Root repo"
  git push origin main

  pushd "$TRASHDIR"

    local_reponame="submod-clone"
    git lfs clone --recursive "$GITSERVER/$reponame" "$local_reponame"

    # check everything is where it should be
    cd $local_reponame
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
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
  grep "Tracking \"\*.dat\"" track.log

  contents="contents"
  contents_oid="$(calc_oid "$contents")"

  printf "%s" "$contents" > a.dat

  git add .gitattributes a.dat

  git commit -m "initial commit" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin main 2>&1 | tee push.log

  pushd $TRASHDIR
    mkdir "$reponame-clone"
    cd "$reponame-clone"

    git lfs clone $GITSERVER/$reponame "."
    git lfs fsck

    assert_local_object "$contents_oid" 8
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd
)
end_test

begin_test "clone empty repository"
(
  set -e

  reponame="clone_empty"
  setup_remote_repo "$reponame"

  cd "$TRASHDIR"
  git lfs clone "$GITSERVER/$reponame" "$reponame" 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
)
end_test

begin_test "clone bare empty repository"
(
  set -e

  reponame="clone_bare_empty"
  setup_remote_repo "$reponame"

  cd "$TRASHDIR"
  git lfs clone "$GITSERVER/$reponame" "$reponame" --bare 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected clone to succeed ..."
    exit 1
  fi
)
end_test

begin_test "clone (HTTP server/proxy require cookies)"
(
  set -e

  # golang net.http.Cookie ignores cookies with IP instead of domain/hostname
  GITSERVER=$(echo "$GITSERVER" | sed 's/127\.0\.0\.1/localhost/')
  cp "$CREDSDIR/127.0.0.1" "$CREDSDIR/localhost"
  printf "localhost\tTRUE\t/\tFALSE\t2145916800\tCOOKIE_GITLFS\tsecret\n" >> "$REMOTEDIR/cookies.txt"
  git config --global http.cookieFile "$REMOTEDIR/cookies.txt"

  reponame="require-cookie-test"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  git add .gitattributes
  git commit -m "Track *.dat"

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

  git push origin main

  # Now clone again, test specific clone dir
  cd "$TRASHDIR"

  newclonedir="require-cookie-test1"
  git lfs clone "$GITSERVER/$reponame" "$newclonedir" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Downloading LFS objects:" lfsclone.log
  # should be no filter errors
  grep "filter" lfsclone.log && exit 1
  grep "error" lfsclone.log && exit 1
  # should be cloned into location as per arg
  [ -d "$newclonedir" ]

  # check a few file sizes to make sure pulled
  pushd "$newclonedir"
    [ $(wc -c < "file1.dat") -eq 110 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 66 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd

  # Now check clone with implied dir
  rm -rf "$reponame"
  git lfs clone "$GITSERVER/$reponame" 2>&1 | tee lfsclone.log
  grep "Cloning into" lfsclone.log
  grep "Downloading LFS objects:" lfsclone.log
  # should be no filter errors
  grep "filter" lfsclone.log && exit 1
  grep "error" lfsclone.log && exit 1
  # clone location should be implied
  [ -d "$reponame" ]

  pushd "$reponame"
    [ $(wc -c < "file1.dat") -eq 110 ]
    [ $(wc -c < "file2.dat") -eq 75 ]
    [ $(wc -c < "file3.dat") -eq 66 ]
    assert_hooks "$(dot_git_dir)"
    [ ! -e "lfs" ]
    assert_clean_status
  popd
)
end_test
