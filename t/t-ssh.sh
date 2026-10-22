#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "ssh with proxy command in lfs.url (default variant)"
(
  set -e

  reponame="batch-ssh-proxy-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  unset GIT_SSH_VARIANT
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo -- '-oProxyCommand[^']*'" push.log
)
end_test

begin_test "ssh with proxy command in lfs.url (custom variant)"
(
  set -e

  reponame="batch-ssh-proxy-simple"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  export GIT_SSH_VARIANT=simple
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo 'oProxyCommand[^']*'" push.log
)
end_test

# ssh_authenticate_unavailable_test performs a test which expects "git push"
# to fail, but only after the Git LFS client correctly defaults to a
# fallback HTTPS URL when the "git-lfs-authenticate" command is reported as
# not found on the SSH remote.
#
# The SSH client used by the tests is triggered by the repository name to
# simulate a specific exit code and optional error message indicating that
# the "git-lfs-authenticate" command is not available.
ssh_authenticate_unavailable_test() {
  local reponame="$1"

  set -e

  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  encoded_reponame="${reponame/%-newline/%0A-newline}"

  # We construct an invalid host name (127.0.0.1.invalid) to guarantee that
  # the HTTPS fallback URL, which will default to port 443, cannot contact
  # any local or remote services not controlled by our test suite.
  gitserver_hostport="${GITSERVER#http://}"
  gitserver_host="${gitserver_hostport/%:*/}"
  gitserver_port="${gitserver_hostport/#*:/}"
  invalid_host="$gitserver_host.invalid"
  git config lfs.url "ssh://git@$invalid_host:$gitserver_port/$encoded_reponame"

  git lfs track "*.dat"

  contents="test"
  printf "%s" "$contents" >test.dat

  git add .gitattributes test.dat
  git commit -m "initial commit"

  # Disable SSH object transfer support to reduce the number of requests
  # in the log file and simplify the checks of the "git-lfs-authenticate"
  # command's arguments.
  git config lfs.sshTransfer never

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  # Requests to both the Locking API and the Batch API should fall back to
  # an HTTPS URL, which should in turn fail because the host name is invalid.
  if echo "$encoded_reponame" | grep -qF "%0A"; then
    [ 2 -eq "$(grep -c "exec: lfs-ssh-echo.*git-lfs-authenticate '\\\\''/${encoded_reponame%%%0A-newline}" push.log)" ]
    [ 2 -eq "$(grep -c "exec (cont\.): -newline'\\\\'' upload" push.log)" ]
  else
    [ 2 -eq "$(grep -c "exec: lfs-ssh-echo.*git-lfs-authenticate /$reponame upload" push.log)" ]
  fi
  [ 2 -eq "$(grep -c -F "ssh: git@$invalid_host does not provide git-lfs-authenticate, falling back to guessed LFS endpoint" push.log)" ]
  [ 2 -eq "$(grep -c -F "HTTP: POST https://$invalid_host/$encoded_reponame/" push.log)" ]
}

begin_test "SSH git-lfs-authenticate unavailable (fallback to HTTPS)"
(
  set -e

  # This repository name announces to the SSH test utility that it should
  # exit as if the "git-lfs-authenticate" command was not found.
  reponame="ssh-unavailable"
  ssh_authenticate_unavailable_test "$reponame"
)
end_test

begin_test "SSH git-lfs-authenticate unavailable (fallback to HTTPS, URL-encoded newline)"
(
  set -e

  # This repository name (after a line feed character is interpolated)
  # announces to the SSH test utility that it should exit as if the
  # "git-lfs-authenticate" command was not found.
  reponame="ssh-unavailable-newline"
  ssh_authenticate_unavailable_test "$reponame"
)
end_test

begin_test "SSH git-lfs-authenticate unavailable message (fallback to HTTPS)"
(
  set -e

  # This repository name announces to the SSH test utility that it should
  # exit with a generic error status code and a message indicating that the
  # "git-lfs-authenticate" command was not found, as may occur when using
  # third-party tools such as Gerrit Code Review.
  reponame="ssh-unavailable-message"
  ssh_authenticate_unavailable_test "$reponame"
)
end_test

begin_test "SSH git-lfs-authenticate invalid URL"
(
  set -e

  # This repository name announces to the SSH test utility that it should
  # return a JSON response containing a URL with invalid encoding.
  reponame="ssh-invalid-url"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://git@}/$reponame"
  git config lfs.url "$sshurl"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "batch request: parse \"$GITSERVER/$reponame\\\\n-newline.*\": net/url: invalid control character in URL" push.log
  grep "Uploading LFS objects:   0% (0/1), 0 B" push.log
)
end_test

ssh_injection_test() {
  local remote_cmd="$1"
  local ssh_cmd="$2"
  local url_option="$3"
  local injection="$4"

  set -e

  if [ "$remote_cmd" = "transfer" ]; then
    setup_pure_ssh
  fi

  reponame="ssh-injection-${remote_cmd}-${ssh_cmd}-${url_option}-${injection}"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  # Disable lock verification to reduce the number of requests in the
  # log file and simplify the checks for SSH remote command injection.
  git config lfs.locksVerify false

  newline=$'\n'
  newline_exploit=0
  case "$injection" in
    *-list)
      # Windows does not permit vertical bar characters in file paths.
      [ "$IS_WINDOWS" -eq 1 ] && exit 0

      exploit_delim="||"
      ;;
    *-newline)
      # Windows does not permit ASCII control characters in file paths.
      [ "$IS_WINDOWS" -eq 1 ] && exit 0

      exploit_delim="$newline"

      newline_exploit=1
      ;;
    *)
      exploit_delim=";"
      ;;
  esac

  exploit="${exploit_delim} touch exploit"

  encoded_exploit="${exploit//$newline/%0a}"
  encoded_exploit="${encoded_exploit// /%20}"
  encoded_exploit="${encoded_exploit//;/%3b}"
  encoded_exploit="${encoded_exploit//|/%7c}"

  if [ "$remote_cmd" = "authenticate" ]; then
    sshurl="${GITSERVER/http:\/\//ssh://git@}/$reponame"
    log_match_path="/$reponame$exploit"

    # Disable SSH object transfer support to reduce the number of requests
    # in the log file and simplify the checks for SSH remote command injection.
    git config lfs.sshTransfer never

    if [ 1 -eq "$newline_exploit" ]; then
      # The "git credential fill" command will reject a "path" attribute
      # which contains a line feed (LF) character, so we avoid passing
      # that attribute to the command.
      git config --global --unset credential.useHttpPath
    fi
  else
    sshurl="$(ssh_remote "$reponame")"
    remote_repo_path="$(native_path "$REMOTEDIR/$reponame.git$exploit")"
    log_match_path="$remote_repo_path"
    if [ "$IS_WINDOWS" -eq 1 ]; then
      log_match_path="/${log_match_path//\\//}"
    fi

    # Enforce SSH object transfer support, which also precludes SSH
    # authentication requests for HTTP credentials and therefore simplifies
    # checks for SSH remote command injection during object transfers.
    git config lfs.sshTransfer always

    # The scutiger git-lfs-transfer utility will reject object transfer
    # requests unless a "remote" repository already exists in which to
    # store the objects.
    #
    # On Windows we use the path returned by "cygpath -w" to avoid problems
    # with automatic MSYS2 file path conversion when the final component of
    # the full path includes a semicolon character, as the Windows drive
    # letter is otherwise miscalculated.  For example, "/d/..." is converted
    # to "C:/d/..." instead of "D:/...".
    git init "$remote_repo_path"
  fi

  case "$injection" in
    encoded*)
      sshurl_with_exploit="$sshurl$encoded_exploit"
      ;;
    *)
      sshurl_with_exploit="$sshurl$exploit"
      ;;
  esac

  git config "lfs.$url_option" "$sshurl_with_exploit"

  if [ "$ssh_cmd" = "custom" ]; then
    export GIT_SSH_COMMAND="lfs-ssh-echo -vvv"

    log_match="sh -c '$GIT_SSH_COMMAND.*'\\\\''git-lfs-$remote_cmd '\\\\''\\\\'\\\\'''\\\\''$log_match_path'\\\\''\\\\'\\\\'''\\\\'' upload'\\\\'''"
  else
    log_match="lfs-ssh-echo.*'git-lfs-$remote_cmd '\\\\''$log_match_path'\\\\'' upload'"
  fi

  if [ 1 -eq "$newline_exploit" ]; then
    IFS="$newline" read -d '' -r log_match log_cont_match <<<"$log_match" || true
  fi

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  rc="${PIPESTATUS[0]}"

  [ ! -f exploit ]

  [ 0 -eq "$rc" ]

  grep "exec: $log_match" push.log
  if [ ! -z "$log_cont_match" ]; then
    grep "exec (cont.): $log_cont_match" push.log
  fi

  refute_server_object "$reponame" "$contents_oid"

  if [ "$remote_cmd" = "authenticate" ]; then
    # We expect the lfs-ssh-echo utility to simulate a JSON response from the
    # git-lfs-authenticate command, with an "href" element whose value is a
    # URL constructed from the git-lfs-authenticate command's first argument.
    # This argument contains a shell meta-character, so we expect that
    # character to appear in the URL-encoded response.  We then expect
    # the HTTP object transfer to succeed into a repository whose name
    # contains the meta-character.
    #
    # Note that we need to URL-encode the repository name for the curl(1)
    # command run by the assert_server_object() assertion.
    assert_server_object "$reponame$encoded_exploit" "$contents_oid"
  else
    # We expect the lfs-ssh-echo utility to invoke the git-lfs-transfer
    # utility with a fully quoted repository path argument.  This argument
    # contains a shell meta-character so we expect the object transfer to
    # succeed into a repository whose name contains that character.
    assert_remote_object_reponame_complete "$reponame.git$exploit" "$contents_oid" "${#contents}"
  fi
}

for remote_cmd in "authenticate" "transfer"; do
  begin_test "git-lfs-$remote_cmd with injected command (lfs.url)"
  (
    set -e

    ssh_injection_test "$remote_cmd" "default" "url" "raw"
  )
  end_test

  begin_test "git-lfs-$remote_cmd with injected command (lfs.url, URL-encoded) (GIT_SSH_COMMAND)"
  (
    set -e

    ssh_injection_test "$remote_cmd" "custom" "url" "encoded"
  )
  end_test

  begin_test "git-lfs-$remote_cmd with injected command (lfs.pushurl, command list) (GIT_SSH_COMMAND)"
  (
    set -e

    ssh_injection_test "$remote_cmd" "custom" "pushurl" "raw-list"
  )
  end_test

  begin_test "git-lfs-$remote_cmd with injected command (lfs.pushurl, URL-encoded newline)"
  (
    set -e

    ssh_injection_test "$remote_cmd" "default" "pushurl" "encoded-newline"
  )
  end_test
done
