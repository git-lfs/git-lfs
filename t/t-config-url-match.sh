#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "SSL certificate matches root path"
(
  set -e

  if [ "$IS_WINDOWS" -eq 1 ]; then
    git config --global "http.sslBackend" "openssl"
  fi

  git config --global "http.$LFS_CLIENT_CERT_URL.sslCert" "$LFS_CLIENT_CERT_FILE"
  git config --global "http.$LFS_CLIENT_CERT_URL.sslKey" "$LFS_CLIENT_KEY_FILE"

  reponame="config-ssl-cert-match-root-path"
  setup_remote_repo "$reponame"
  clone_repo_clientcert "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  printf "%s" "$contents" >a.bin

  git add .gitattributes a.bin
  git commit -m "initial commit"

  git push origin main

  # Test with the global "http.<url>.sslCert" option reset to an invalid
  # certificate file and a local "http.<url>/.sslCert" option set to the
  # valid file.  The global option should be ignored in favour of the local
  # one, even though the local option's URL has a trailing slash character.
  # See https://github.com/git-lfs/git-lfs/issues/6112.
  git config --global "http.$LFS_CLIENT_CERT_URL.sslCert" "$TRASHDIR/nonexistent/cert.pem"
  git config "http.$LFS_CLIENT_CERT_URL/.sslCert" "$LFS_CLIENT_CERT_FILE"

  # The reported issue occurred only half the time, on average, due
  # to Go's arbitrary map key ordering.  To increase the likelihood of
  # reproducing the problem, repeat the check up to eight times.
  for ((i=0; i<8; i++)); do
    rm -rf .git/lfs/objects

    git lfs pull

    assert_local_object "$(calc_oid "$contents")" "${#contents}"
  done
)
end_test

begin_test "SSL certificate matches canonical path"
(
  set -e

  if [ "$IS_WINDOWS" -eq 1 ]; then
    git config --global "http.sslBackend" "openssl"
  fi

  git config --global "http.$LFS_CLIENT_CERT_URL.sslCert" "$LFS_CLIENT_CERT_FILE"
  git config --global "http.$LFS_CLIENT_CERT_URL.sslKey" "$LFS_CLIENT_KEY_FILE"

  reponame="config-ssl-cert-match-canonical-path"
  setup_remote_repo "$reponame"
  clone_repo_clientcert "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  printf "%s" "$contents" >a.bin

  git add .gitattributes a.bin
  git commit -m "initial commit"

  git push origin main

  # Test with the global "http.<url>.sslCert" option reset to an invalid
  # certificate file and a local "http.<url>/./.sslCert" option set to the
  # valid file.  The global option should be ignored in favour of the local
  # one, even though the local option's URL has a spurious path segment.
  git config --global "http.$LFS_CLIENT_CERT_URL.sslCert" "$TRASHDIR/nonexistent/cert.pem"
  git config "http.$LFS_CLIENT_CERT_URL/./.sslCert" "$LFS_CLIENT_CERT_FILE"

  # First push a new commit that does not include any Git LFS objects to
  # confirm Git matches the local option's URL to the remote, regardless
  # of the spurious path segment.
  echo "test" >a.txt
  git add a.txt
  git commit -m "second commit"

  git push origin main

  # Next confirm Git LFS also matches the local option's URL to the remote,
  # regardless of the spurious path segment.
  rm -rf .git/lfs/objects

  git lfs pull

  assert_local_object "$(calc_oid "$contents")" "${#contents}"
)
end_test
