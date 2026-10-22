#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"
lfsname="submodule-config-test-lfs"
reponame="submodule-config-test-repo"
submodname="submodule-config-test-submodule"

begin_test "submodule env with .lfsconfig"
(
  set -e

  # setup dummy repo with lfs store
  # no git data will be pushed, just lfs objects
  setup_remote_repo "$lfsname"
  echo $GITSERVER/$lfsname.git/info/lfs

  # setup submodule
  setup_remote_repo "$submodname"
  clone_repo "$submodname" submod
  mkdir dir
  git config -f .lfsconfig lfs.url "$GITSERVER/$lfsname.git/info/lfs"
  git lfs track "*.dat"
  submodcontent="submodule lfs file"
  submodoid=$(calc_oid "$submodcontent")
  printf "%s" "$submodcontent" > dir/test.dat
  git add .lfsconfig .gitattributes dir
  git commit -m "create submodule"
  git push origin main

  assert_server_object "$lfsname" "$submodoid"

  # setup repo with submodule
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo
  git config -f .lfsconfig lfs.url "$GITSERVER/$lfsname.git/info/lfs"
  git submodule add -b main "$GITSERVER/$submodname" sub
  git submodule update
  git lfs track "*.dat"
  mkdir dir
  repocontent="repository lfs file"
  repooid=$(calc_oid "$repocontent")
  printf "%s" "$repocontent" > dir/test.dat
  git add .gitattributes .lfsconfig .gitmodules dir sub
  git commit -m "create repo"
  git push origin main

  assert_server_object "$lfsname" "$repooid"

  echo "repo"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log

  cd sub
  echo "./sub"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log

  cd dir
  echo "./sub/dir"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$lfsname.git/info/lfs (auth=basic)$" env.log
)
end_test

begin_test "submodule update --init --remote with .lfsconfig"
(
  set -e
  clone_repo "$reponame" clone
  grep "$repocontent" dir/test.dat

  git submodule update --init --remote

  grep "$submodcontent" sub/dir/test.dat
)
end_test

begin_test "submodule URL with injected custom adapter"
(
  set -e

  reponame="submodule-adapter-injection"
  git init "$reponame"
  cd "$reponame"

  # We generate a placeholder commit OID to use in the submodule entry and
  # let Git choose the hash algorithm based on the value of GIT_DEFAULT_HASH.
  git commit --allow-empty -m "exploit submodule"
  exploit_submodule_oid="$(git rev-parse HEAD)"
  git update-ref -d refs/heads/main

  git lfs track "*.bin"

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  # We add a malicious custom transfer adapter which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "adapter$X"

  exploit_submodule_url="$(printf \
    '%s/%s-submodule\nlfs.customtransfer.lfs-standalone-file.path=./adapter%s' \
    "$GITSERVER" "$reponame" "$X" \
  )"

  git config --file .gitmodules submodule.exploit.path exploit-submodule
  git config --file .gitmodules submodule.exploit.url "$exploit_submodule_url"

  git add .gitattributes .gitmodules "adapter$X" test.bin
  git update-index --add \
    --cacheinfo "160000,$exploit_submodule_oid,exploit-submodule"
  git commit -m "initial commit"

  cd ..

  git clone "$reponame" "${reponame}-assert"
  cd "${reponame}-assert"

  # Git copies the submodule's URL into ".git/config", but does not contact
  # the URL.  The URL attempts to inject an additional configuration entry
  # that would override the default standalone custom transfer adapter used
  # when fetching from a local repository.  We expect that when we re-fetch
  # an object, the default adapter is executed and not the malicious one.
  git submodule init

  rm -rf .git/lfs/objects

  GIT_TRACE=1 git lfs pull 2>&1 | tee pull.log
  rc="${PIPESTATUS[0]}"

  [ ! -f exploit ]

  [ 0 -eq "$rc" ]

  [ 0 -eq "$(grep -c "exec: .*adapter$X" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"
)
end_test

begin_test "submodule URL with injected pointer extension"
(
  set -e

  reponame="submodule-pointer-extension-injection"
  git init "$reponame"
  cd "$reponame"

  git lfs track "*.bin"

  # We generate a placeholder commit OID to use in the submodule entry and
  # let Git choose the hash algorithm based on the value of GIT_DEFAULT_HASH.
  git commit --allow-empty -m "exploit submodule"
  exploit_submodule_oid="$(git rev-parse HEAD)"
  git update-ref -d refs/heads/main

  # We add a malicious pointer extension program which should not be executed
  # when the repository is cloned.
  cp "$BINPATH/lfstest-exploitcheck$X" "ext-clean$X"

  exploit_submodule_url="$(printf \
    '%s/%s-submodule\nlfs.extension.exploit.clean=./ext-clean%s %%f\nlfs.extension.exploit.priority=1' \
    "$GITSERVER" "$reponame" "$X" \
  )"

  git config --file .gitmodules submodule.exploit.path exploit-submodule
  git config --file .gitmodules submodule.exploit.url "$exploit_submodule_url"

  git add .gitattributes .gitmodules "ext-clean$X"
  git update-index --add \
    --cacheinfo "160000,$exploit_submodule_oid,exploit-submodule"

  # Git copies the submodule's URL into ".git/config", but does not contact
  # the URL.  The URL attempts to inject an additional configuration entry
  # for a Git LFS pointer extension program that would be run by our "clean"
  # filter when files are added to the index.  We expect that the malicious
  # pointer extension program is not executed when we add a file to the index.
  git submodule init

  contents="test"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  GIT_TRACE=1 git add test.bin >add.log 2>&1

  [ ! -f exploit ]

  [ 0 -eq "$(grep -c "exec: .*ext-clean$X" add.log)" ]

  [ 0 -eq "$(git cat-file -p :test.bin | grep -c -F "ext-0-exploit")" ]

  assert_local_object "$contents_oid" "${#contents}"
)
end_test
