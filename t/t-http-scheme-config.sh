#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "http URL does not use https-scoped sslcert config"
(
  set -e

  reponame="scheme-scoped-cert"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install --local

  git lfs track "*.bin"
  git add .gitattributes

  # Create an LFS pointer file directly
  cat > file.bin <<EOF
version https://git-lfs.github.com/spec/v1
oid sha256:45942f670e64f1a015f3ca66ac758d2b2f1eaa42c086f15662e36e6485ea4977
size 49042944
EOF
  git add file.bin
  git commit -m "initial commit"

  # Configure an http:// LFS endpoint
  git remote add origin "http://192.0.2.1:9999/repo.git"
  git config lfs.url "http://192.0.2.1:9999/repo/info/lfs"
  git config lfs.locksverify false
  git config lfs.transfer.maxretries 1

  # Set client cert config scoped to https:// only — should NOT apply to http://
  git config "http.https://192.0.2.1:9999/.sslcert" "/nonexistent/cert.pem"
  git config "http.https://192.0.2.1:9999/.sslkey" "/nonexistent/key.pem"

  # git lfs fetch should fail with a connection error (non-routable IP),
  # NOT with "Error reading client cert file"
  set +e
  git lfs fetch 2>&1 | tee fetch.log
  set -e

  if grep -q "Error reading client cert file" fetch.log; then
    echo "FAIL: git-lfs tried to load client cert for http:// URL from https-scoped config"
    exit 1
  fi
)
end_test

begin_test "https URL uses https-scoped sslcert config"
(
  set -e

  reponame="scheme-scoped-cert-https"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install --local

  git lfs track "*.bin"
  git add .gitattributes

  cat > file.bin <<EOF
version https://git-lfs.github.com/spec/v1
oid sha256:45942f670e64f1a015f3ca66ac758d2b2f1eaa42c086f15662e36e6485ea4977
size 49042944
EOF
  git add file.bin
  git commit -m "initial commit"

  # Configure an https:// LFS endpoint
  git remote add origin "https://192.0.2.1:9999/repo.git"
  git config lfs.url "https://192.0.2.1:9999/repo/info/lfs"
  git config lfs.locksverify false
  git config lfs.transfer.maxretries 1

  # Set client cert config scoped to https:// — should apply to https:// URL
  git config "http.https://192.0.2.1:9999/.sslcert" "/nonexistent/cert.pem"
  git config "http.https://192.0.2.1:9999/.sslkey" "/nonexistent/key.pem"

  # git lfs fetch should fail trying to load the cert (files don't exist)
  set +e
  git lfs fetch 2>&1 | tee fetch.log
  set -e

  grep -q "Error reading client cert file" fetch.log
)
end_test
