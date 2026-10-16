#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

sensitive_headers=(Authorization Cookie)
source_header_marker="lfs-source-marker"
source_cookie_name="lfs-source-cookie"

# The SSL test server also accepts HTTP connections on the same port.
sslgitserver_http="${SSLGITSERVER/#https:/http:}"

setup_header_redirect_repo() {
  local reponame="$1"
  local contents="$2"

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  # Disable lock verification to reduce the number of requests and
  # simplify the checks of HTTP request headers.
  git config lfs.locksVerify false

  # Create or truncate the file into which the server will log HTTP headers.
  : > "$REMOTEDIR/$reponame.headers"
}

add_sensitive_headers() {
  local url="$1"

  local header value
  for header in "${sensitive_headers[@]}"; do
    value="$source_header_marker"
    if [ "$header" = "Cookie" ]; then
      value="$source_cookie_name=$source_header_marker"
    fi
    git config --add "http.$url.extraHeader" "$header: $value"
  done
  git config --add "http.$url.extraHeader" "X-Test-Benign: $source_header_marker"
}

run_header_redirect_push() {
  local url="$1"

  echo "refs/heads/main main refs/heads/main 0000000000000000000000000000000000000000" | \
    git lfs pre-push origin "$url"
}

# assert_header_seen() checks that the test server recorded the given header,
# with the source marker value, in a request at the given redirect stage.
assert_header_seen() {
  local stage="$1"
  local header="$2"

  grep -Eq "^$stage ([^,]+,)*$header(,|$)" "$REMOTEDIR/$reponame.headers"
}

# refute_header_seen() checks that the test server did not record the
# given header, with the source marker value, in a request at the given
# redirect stage.
refute_header_seen() {
  local stage="$1"
  local header="$2"

  if grep -Eq "^$stage ([^,]+,)*$header(,|$)" "$REMOTEDIR/$reponame.headers"; then
    exit 1
  fi
}

assert_redirect_headers() {
  local source_stage="${1:-batch-source}"
  local target_stage="${2:-batch-target}"
  local header

  for header in "${sensitive_headers[@]}" X-Test-Benign; do
    assert_header_seen "$source_stage" "$header"
  done

  assert_header_seen "$target_stage" X-Test-Benign
  for header in "${sensitive_headers[@]}"; do
    refute_header_seen "$target_stage" "$header"
  done
}

begin_test "pre-push redirect to a different host drops sensitive headers"
(
  set -e

  # This repository name announces to the server that it should record
  # which headers it receives carrying the "lfs-source-marker" value,
  # and that it should not require an Authorization header.
  reponame="header-redirect-host"

  contents="host-test"
  contents_oid="$(calc_oid "$contents")"

  setup_header_redirect_repo "$reponame" "$contents"

  # The "/redirect307/abs/" path makes the server redirect to "127.0.0.1".
  source_url="${GITSERVER/127.0.0.1/localhost}/redirect307/abs/$reponame.git/info/lfs"
  git config lfs.url "$source_url"

  add_sensitive_headers "$source_url"

  run_header_redirect_push "$source_url"

  assert_redirect_headers

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push redirect to a different port drops sensitive headers"
(
  set -e

  # This repository name announces to the server that it should record
  # which headers it receives carrying the "lfs-source-marker" value,
  # and that it should not require an Authorization header.
  reponame="header-redirect-port"

  contents="port-test"
  contents_oid="$(calc_oid "$contents")"

  setup_header_redirect_repo "$reponame" "$contents"

  # The "/redirect307/abs/" path makes the server redirect to its main port.
  source_url="$sslgitserver_http/redirect307/abs/$reponame.git/info/lfs"
  git config lfs.url "$source_url"

  add_sensitive_headers "$source_url"

  run_header_redirect_push "$source_url"

  assert_redirect_headers

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push HTTP to HTTPS redirect drops sensitive headers"
(
  set -e

  # This repository name announces to the server that it should record
  # which headers it receives carrying the "lfs-source-marker" value,
  # and that it should not require an Authorization header.
  reponame="header-redirect-scheme"

  contents="scheme-test"
  contents_oid="$(calc_oid "$contents")"

  setup_header_redirect_repo "$reponame" "$contents"

  # The "/redirect307/ssl/" path makes the server redirect to HTTPS on the
  # same port.
  source_url="$sslgitserver_http/redirect307/ssl/$reponame.git/info/lfs"
  git config lfs.url "$source_url"

  add_sensitive_headers "$source_url"

  run_header_redirect_push "$source_url"

  assert_redirect_headers

  assert_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "pre-push same-origin redirect retains headers"
(
  set -e

  # This repository name announces to the server that it should record
  # which headers it receives carrying the "lfs-source-marker" value,
  # and that it should not require an Authorization header.
  reponame="header-redirect-same"

  contents="same-test"
  contents_oid="$(calc_oid "$contents")"

  setup_header_redirect_repo "$reponame" "$contents"

  # The "/redirect307/rel/" path makes the server redirect to a relative URL.
  source_url="$GITSERVER/redirect307/rel/$reponame.git/info/lfs"
  git config lfs.url "$source_url"

  add_sensitive_headers "$source_url"

  run_header_redirect_push "$source_url"

  for header in "${sensitive_headers[@]}" X-Test-Benign; do
    assert_header_seen batch-source "$header"
    assert_header_seen batch-target "$header"
  done

  assert_server_object "$reponame" "$contents_oid"
)
end_test

for mode in auth direct; do
  begin_test "download redirect does not forward sensitive headers ($mode path)"
  (
    set -e

    # These repository names announce to the server that it should record
    # which headers it receives carrying the "lfs-source-marker" value, and
    # use the "localhost" hostname in its Batch API responses' object URLs.
    # The "direct" name also announces that the server should mark downloads
    # as "authenticated", so the client uses its unauthenticated request path.
    reponame="header-redirect-download-$mode"

    # This content announces to the server that its object URLs should be
    # "/redirect307/objects/" URLs, which it will redirect to "127.0.0.1".
    contents="redirect-storage-upload"
    contents_oid="$(calc_oid "$contents")"

    setup_header_redirect_repo "$reponame" "$contents"

    git push origin main

    object_url="${GITSERVER/127.0.0.1/localhost}/redirect307/objects"
    add_sensitive_headers "$object_url"

    rm -rf .git/lfs/objects

    git lfs fetch origin main

    assert_redirect_headers object-source object-target

    assert_local_object "$contents_oid" "${#contents}"
  )
  end_test
done

begin_test "cookie from an earlier response does not follow a cross-origin redirect"
(
  set -e

  # This repository name announces to the server that it should record
  # which headers it receives carrying the "lfs-source-marker" value,
  # include a Set-Cookie header in its Batch API responses, and use the
  # "localhost" hostname in their object URLs.
  reponame="header-redirect-cookie"

  # This content announces to the server that its object URLs should be
  # "/redirect307/objects/" URLs, which it will redirect to "127.0.0.1".
  contents="redirect-storage-upload"
  contents_oid="$(calc_oid "$contents")"

  setup_header_redirect_repo "$reponame" "$contents"

  source_url="${GITSERVER/127.0.0.1/localhost}/$reponame.git/info/lfs"
  git config lfs.url "$source_url"

  # Enable a cookie jar so the client keeps the Batch API response's cookie.
  : > "$REMOTEDIR/$reponame-cookies.txt"
  git config http.cookieFile "$REMOTEDIR/$reponame-cookies.txt"

  run_header_redirect_push "$source_url"

  # The cookie set by the Batch API response should accompany the initial
  # object request to "localhost", but not the redirected one to "127.0.0.1".
  assert_header_seen object-source Cookie
  refute_header_seen object-target Cookie

  assert_server_object "$reponame" "$contents_oid"
)
end_test
