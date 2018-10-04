#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

if [ $IS_WINDOWS -eq 1 ]; then
  echo "skip $0: Windows lacks POSIX permissions"
  exit
fi

clean_setup () {
  mkdir "$1"
  cd "$1"
  git init
}

perms_for () {
  local file=$(echo "$1" | sed "s!^\(..\)\(..\)!.git/lfs/objects/\1/\2/\1\2!")
  ls -l "$file" | awk '{print $1}'
}

begin_test "honors umask"
(
  set -e
  clean_setup "simple"

  umask 027
  echo "whatever" | git lfs clean | tee clean.log
  [ "$(perms_for cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411)" = "-rw-r-----" ]

  umask 007
  echo "random" | git lfs clean | tee clean.log
  [ "$(perms_for 87c1b129fbadd7b6e9abc0a9ef7695436d767aece042bec198a97e949fcbe14c)" = "-rw-rw----" ]
)
end_test
