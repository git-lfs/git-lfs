#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "access config with injected custom adapter"
(
  set -e

  # This repository name announces to the server that it should select the
  # "testcustom" transfer adapter when it is advertised by the client.
  reponame="custom-transfer-config-access-injection"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  git push origin main

  git config --file .lfsconfig \
    lfs.url "$(repo_endpoint "$GITSERVER" "$reponame")"
  git config --file .lfsconfig \
    lfs.customtransfer.testcustom.path.access "./adapter$X"

  # We add a malicious custom transfer adapter which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "adapter$X"

  git add .lfsconfig "adapter$X"
  git commit -m "exploit"

  git push origin main

  cd ..
  GIT_TRACE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  rc="${PIPESTATUS[0]}"

  pushd "${reponame}-assert"
    [ ! -f exploit ]

    assert_local_object "$contents_oid" "${#contents}"
  popd

  [ 0 -eq "$rc" ]

  grep "tq: starting transfer adapter \"basic\"" clone.log

  [ 0 -eq "$(grep -c "exec: .*adapter${X//./\\.}" clone.log)" ]
  [ 0 -eq "$(grep -c -F "xfer[adapter$X]: exploit" clone.log)" ]
)
end_test

begin_test "access config with injected standalone custom adapter"
(
  set -e

  reponame="standalone-file-config-access-injection"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  git config --file .lfsconfig \
    lfs.customtransfer.lfs-standalone-file.path.access "./adapter$X"

  # We add a malicious custom transfer adapter which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "adapter$X"

  git add .lfsconfig "adapter$X"
  git commit -m "exploit"

  cd ..
  GIT_TRACE=1 git clone "$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  rc="${PIPESTATUS[0]}"

  pushd "${reponame}-assert"
    [ ! -f exploit ]

    assert_local_object "$contents_oid" "${#contents}"
  popd

  [ 0 -eq "$rc" ]

  [ 0 -eq "$(grep -c "exec: .*adapter${X//./\\.}" clone.log)" ]
  [ 0 -eq "$(grep -c -F "xfer[adapter$X]: exploit" clone.log)" ]
)
end_test

begin_test "remote config with injected custom adapter"
(
  set -e

  # This repository name announces to the server that it should select the
  # "testcustom" transfer adapter when it is advertised by the client.
  reponame="custom-transfer-config-remote-injection"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  git push origin main

  git config --file .lfsconfig \
    lfs.url "$(repo_endpoint "$GITSERVER" "$reponame")"
  git config --file .lfsconfig \
    remote.lfs.customtransfer.testcustom.path.lfsurl "./adapter$X"

  # We add a malicious custom transfer adapter which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "adapter$X"

  git add .lfsconfig "adapter$X"
  git commit -m "exploit"

  git push origin main

  cd ..
  GIT_TRACE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  rc="${PIPESTATUS[0]}"

  pushd "${reponame}-assert"
    [ ! -f exploit ]

    assert_local_object "$contents_oid" "${#contents}"
  popd

  [ 0 -eq "$rc" ]

  grep "tq: starting transfer adapter \"basic\"" clone.log

  [ 0 -eq "$(grep -c "exec: .*adapter${X//./\\.}" clone.log)" ]
  [ 0 -eq "$(grep -c -F "xfer[adapter$X]: exploit" clone.log)" ]
)
end_test

begin_test "remote config with injected standalone custom adapter"
(
  set -e

  reponame="standalone-file-config-remote-injection"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  git config --file .lfsconfig \
    remote.lfs.customtransfer.lfs-standalone-file.path.lfsurl "./adapter$X"

  # We add a malicious custom transfer adapter which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "adapter$X"

  git add .lfsconfig "adapter$X"
  git commit -m "exploit"

  cd ..
  GIT_TRACE=1 git clone "$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  rc="${PIPESTATUS[0]}"

  pushd "${reponame}-assert"
    [ ! -f exploit ]

    assert_local_object "$contents_oid" "${#contents}"
  popd

  [ 0 -eq "$rc" ]

  [ 0 -eq "$(grep -c "exec: .*adapter${X//./\\.}" clone.log)" ]
  [ 0 -eq "$(grep -c -F "xfer[adapter$X]: exploit" clone.log)" ]
)
end_test
