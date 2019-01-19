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
  local perms=$(ls -l "$file" | awk '{print $1}')
  # Trim extended attributes:
  echo ${perms:0:10}
}

assert_dir_perms () {
  local perms="$1"
  [ "$(find .git/lfs -type d -ls | grep -vE "$perms")" = "" ]
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

begin_test "honors umask for directories"
(
  set -e
  reponame="simple-directories"
  setup_remote_repo_with_file "$reponame" "a.dat"

  umask 027
  clone_repo "$reponame" "$reponame-a"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxr-[xs]---"

  umask 007
  clone_repo "$reponame" "$reponame-b"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxrw[xs]---"
)
end_test

# This is tested more comprehensively in the unit tests.
begin_test "honors core.sharedrepository"
(
  set -e
  clean_setup "shared-repo"

  umask 027
  git config core.sharedRepository 0660
  echo "whatever" | git lfs clean | tee clean.log
  [ "$(perms_for cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411)" = "-rw-rw----" ]

  git config core.sharedRepository everybody
  echo "random" | git lfs clean | tee clean.log
  [ "$(perms_for 87c1b129fbadd7b6e9abc0a9ef7695436d767aece042bec198a97e949fcbe14c)" = "-rw-rw-r--" ]

  git config core.sharedRepository false
  echo "something else" | git lfs clean | tee clean.log
  [ "$(perms_for a1621be95040239ee14362c16e20510ddc20f527d772d823b2a1679b33f5cd74)" = "-rw-r-----" ]

  umask 007
  echo "who cares" | git lfs clean | tee clean.log
  [ "$(perms_for 261ded5f01a8ca18d9fb1958e8f58c53fa77648cc88a6d67c93d241a91133f3e)" = "-rw-rw----" ]

)
end_test

begin_test "honors core.sharedrepository for directories"
(
  set -e
  reponame="shared-repo-directories"
  setup_remote_repo_with_file "$reponame" "a.dat"

  umask 027
  git config --global core.sharedRepository 0660
  clone_repo "$reponame" "$reponame-a"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxrw[xs]---"

  git config --global core.sharedRepository everybody
  clone_repo "$reponame" "$reponame-b"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxrw[xs]r-x"

  git config --global core.sharedRepository false
  clone_repo "$reponame" "$reponame-c"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxr-[xs]---"

  umask 007
  clone_repo "$reponame" "$reponame-d"
  echo "whatever" | git lfs clean
  assert_dir_perms "drwxrw[xs]---"

  git config --global --unset-all core.sharedRepository
)
end_test
