#!/bin/sh

. "test/testlib.sh"

LFSTEST_SSHDIR="$REMOTEDIR/ssh"
export LFSTEST_SSHDIR

begin_test "ssh"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  # setup SSH, host doesn't matter because we use mock SSH exe
  export GIT_SSH=lfstest-mockssh
  git config -f ".git/config" remote.origin.lfsurl "git@mock.com:$reponame"

  git lfs track "*.dat"
  echo "push a" > a.dat
  echo "push b" > b.dat
  echo "push c" > c.dat
  echo "push d" > d.dat
  echo "push e" > e.dat
  echo "push f" > f.dat
  echo "push g" > g.dat
  echo "push h" > h.dat
  echo "push i" > i.dat
  echo "push j" > j.dat
  git add .gitattributes *.dat
  git commit -m "add files"

  git lfs push origin master 2>&1 | tee ssh-push.log
  grep "(10 of 10 files)" ssh-push.log

  echo "push k" > k.dat
  echo "push l" > l.dat
  git add k.dat l.dat
  git commit -m "2 extra files"

  git lfs push origin master 2>&1 | tee ssh-push.log
  # because we haven't 'git push'ed it will try to push the previous 10 too
  # but be rejected as already existing; only new 2 will be pushed
  grep "(2 of 12 files)" ssh-push.log

  # now fetch them back
  rm *.dat
  rm -rf .git/lfs 
  git lfs fetch 2>&1 | tee ssh-fetch.log
  grep "(12 of 12 files)" ssh-fetch.log

)
end_test