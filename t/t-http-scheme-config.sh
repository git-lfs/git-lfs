#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "http URL does not use https-scoped sslcert config"
(
  set -e

  reponame="scheme-scoped-cert-http"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "content" > a.dat
  git add .gitattributes a.dat
  git commit -m "add lfs file"

  # Extract host:port from $GITSERVER (e.g. "127.0.0.1:PORT") so we can
  # build a config key that is scoped to the https:// scheme for that host.
  gitserver_hostport="${GITSERVER#*://}"

  # Set client cert config scoped to https:// — should NOT apply to http://
  git config "http.https://$gitserver_hostport/.sslCert" "/nonexistent/cert.pem"
  git config "http.https://$gitserver_hostport/.sslKey" "/nonexistent/key.pem"

  # Push to the http:// server. If the bug is present (https-scoped cert config
  # is applied to the http:// connection), this fails with "Error reading client
  # cert file". Post-fix, the cert config is correctly ignored and push succeeds.
  git lfs push origin main 2>&1 | tee push.log
  if [ "${PIPESTATUS[0]}" -ne 0 ]; then
    echo >&2 "FAIL: git-lfs tried to load client cert for http:// URL from https-scoped config"
    cat push.log >&2
    exit 1
  fi

  if grep -q "Error reading client cert file" push.log; then
    echo >&2 "FAIL: git-lfs tried to load client cert for http:// URL from https-scoped config"
    exit 1
  fi
)
end_test

begin_test "https URL uses https-scoped sslcert config"
(
  set -e

  reponame="scheme-scoped-cert-https"
  setup_remote_repo "$reponame"
  clone_repo_ssl "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "content" > a.dat
  git add .gitattributes a.dat
  git commit -m "add lfs file"

  # Extract host:port from $SSLGITSERVER (e.g. "127.0.0.1:PORT").
  sslserver_hostport="${SSLGITSERVER#*://}"

  # Set client cert config scoped to https:// — should apply to https:// connection.
  git config "http.https://$sslserver_hostport/.sslCert" "/nonexistent/cert.pem"
  git config "http.https://$sslserver_hostport/.sslKey" "/nonexistent/key.pem"

  # Push should fail trying to load the client cert (files don't exist), confirming
  # that the https-scoped config is applied to the https:// connection.
  git lfs push origin main 2>&1 | tee push.log
  if [ "${PIPESTATUS[0]}" -eq 0 ]; then
    echo >&2 "FAIL: expected 'git lfs push' to fail with a cert error"
    exit 1
  fi

  grep -q "Error reading client cert file" push.log
)
end_test
