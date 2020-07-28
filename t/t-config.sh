#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "default config"
(
  set -e
  reponame="default-config"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/$reponame"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)" env.log

  git config --file=.gitconfig lfs.url http://gitconfig-file-ignored
  git config --file=.lfsconfig lfs.url http://lfsconfig-file
  git config --file=.lfsconfig lfs.http://lfsconfig-file.access lfsconfig
  git lfs env | tee env.log
  grep "Endpoint=http://lfsconfig-file (auth=lfsconfig)" env.log

  git config --file=.lfsconfig --unset lfs.url
  git config --file=.lfsconfig --unset lfs.http://lfsconfig-file.access

  # new endpoint url from local git config
  # access setting no longer applied
  git config lfs.url http://local-lfsconfig
  git lfs env | tee env.log
  grep "Endpoint=http://local-lfsconfig (auth=none)" env.log

  # add the access setting to lfsconfig
  git config --file=.lfsconfig lfs.http://local-lfsconfig.access lfsconfig
  git lfs env | tee env.log
  grep "Endpoint=http://local-lfsconfig (auth=lfsconfig)" env.log

  git config --file=.lfsconfig --unset lfs.http://local-lfsconfig.access

  # add the access setting to git config
  git config lfs.http://local-lfsconfig.access gitconfig
  git lfs env | tee env.log
  grep "Endpoint=http://local-lfsconfig (auth=gitconfig)" env.log
)
end_test

begin_test "config reads from repository"
(
  set -e
  reponame="repository-config"
  setup_remote_repo "$reponame"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/$reponame"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)" env.log

  git config --file=.lfsconfig lfs.url http://lfsconfig-file
  git config --file=.lfsconfig lfs.http://lfsconfig-file.access lfsconfig
  git add .lfsconfig
  git commit -m 'Add file'
  git push origin HEAD

  git checkout -b side
  git config --file=.lfsconfig lfs.url http://lfsconfig-file-side
  git config --file=.lfsconfig lfs.http://lfsconfig-file-side.access lfsconfig
  git add .lfsconfig
  git commit -m 'Add file for side'
  git push origin HEAD

  mkdir "../$reponame-2"
  cd "../$reponame-2"
  git init
  git remote add origin "$GITSERVER/$reponame"

  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)" env.log

  git fetch origin
  git symbolic-ref HEAD refs/remotes/origin/side
  git show "HEAD:.lfsconfig"
  git lfs env | tee env.log
  grep "Endpoint=http://lfsconfig-file-side (auth=lfsconfig)" env.log

  git read-tree refs/remotes/origin/main
  git lfs env | tee env.log
  grep "Endpoint=http://lfsconfig-file (auth=lfsconfig)" env.log
)
end_test

begin_test "can read LFS file with name before .lfsconfig"
(
  set -e
  reponame="early-file-config"
  setup_remote_repo "$reponame"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/$reponame"

  git lfs track "*.bin"
  git config --file=.lfsconfig lfs.url "$GITSERVER/$reponame.git/info/lfs"

  echo "abc" > .bin
  echo "def" > a.bin

  git add .
  git commit -m "Add files"
  git push origin HEAD
  rm -fr .git/lfs/objects

  cd ..
  git clone "$reponame" "$reponame-2"
  cd "$reponame-2"
  grep abc .bin
  grep def a.bin
)
end_test

begin_test "extension config"
(
  set -e

  git config --global lfs.extension.env-test.clean "env-test-clean"
  git config --global lfs.extension.env-test.smudge "env-test-smudge"
  git config --global lfs.extension.env-test.priority 0

  reponame="extension-config"
  mkdir $reponame
  cd $reponame
  git init

  expected0="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 0"

  [ "$expected0" = "$(git lfs ext)" ]

  # any git config takes precedence over .lfsconfig
  git config --global --unset lfs.extension.env-test.priority

  git config --file=.lfsconfig lfs.extension.env-test.clean "file-env-test-clean"
  git config --file=.lfsconfig lfs.extension.env-test.smudge "file-env-test-smudge"
  git config --file=.lfsconfig lfs.extension.env-test.priority 1
  cat .lfsconfig
  expected1="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 1"

  [ "$expected1" = "$(GIT_TRACE=5 git lfs ext)" ]

  git config lfs.extension.env-test.clean "local-env-test-clean"
  git config lfs.extension.env-test.smudge "local-env-test-smudge"
  git config lfs.extension.env-test.priority 2
  expected2="Extension: env-test
    clean = local-env-test-clean
    smudge = local-env-test-smudge
    priority = 2"

  [ "$expected2" = "$(git lfs ext)" ]
)
end_test

begin_test "url alias config"
(
  set -e

  mkdir url-alias
  cd url-alias

  git init

  # When more than one insteadOf strings match a given URL, the longest match is used.
  git config url."http://wrong-url/".insteadOf alias
  git config url."http://actual-url/".insteadOf alias:
  git config lfs.url alias:rest
  git lfs env | tee env.log
  grep "Endpoint=http://actual-url/rest (auth=none)" env.log
)
end_test

begin_test "ambiguous url alias"
(
  set -e

  mkdir url-alias-ambiguous
  cd url-alias-ambiguous

  git init

  git config url."http://actual-url/".insteadOf alias:
  git config url."http://dupe-url".insteadOf alias:
  git config lfs.url alias:rest
  git config -l | grep url

  git lfs env 2>&1 | tee env2.log
  grep "WARNING: Multiple 'url.*.insteadof'" env2.log
)
end_test

begin_test "multiple config"
(
  set -e

  mkdir url-alias-multiple
  cd url-alias-multiple

  git init

  # When more than one insteadOf strings match a given URL, the longest match is used.
  git config url."http://wrong-url/".insteadOf alias
  git config url."http://actual-url/".insteadOf alias:
  git config --add url."http://actual-url/".insteadOf alias2:
  git config lfs.url alias:rest
  git lfs env | tee env.log
  grep "Endpoint=http://actual-url/rest (auth=none)" env.log

  git config lfs.url alias2:rest
  git lfs env | tee env.log
  grep "Endpoint=http://actual-url/rest (auth=none)" env.log
)
end_test


begin_test "url alias must be prefix"
(
  set -e

  mkdir url-alias-bad
  cd url-alias-bad

  git init

  git config url."http://actual-url/".insteadOf alias:
  git config lfs.url badalias:rest
  git lfs env | tee env.log
  grep "SSH=badalias:rest" env.log
)
end_test

begin_test "config: ignoring unsafe lfsconfig keys"
(
  set -e

  reponame="config-unsafe-lfsconfig-keys"
  git init "$reponame"
  cd "$reponame"

  # Insert an 'unsafe' key into this repository's '.lfsconfig'.
  git config --file=.lfsconfig core.askpass unsafe

  git lfs env 2>&1 | tee status.log

  grep "WARNING: These unsafe lfsconfig keys were ignored:" status.log
  grep "  core.askpass" status.log
)
end_test
