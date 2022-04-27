#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

setup_successful_repo () {
  if [ "$1" != "--track-later" ]
  then
    git lfs track '*.dat'
  else
    touch .gitattributes
  fi

  seq 1 10 > a.dat
  git add .gitattributes *.dat
  git commit -m 'Initial import'

  git checkout -b other

  # sed -i isn't portable.
  sed -e 's/9/B/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'B'

  if [ "$1" = "--track-later" ]
  then
    git lfs track '*.dat'
    git add .gitattributes
    git commit -m 'B2'
  fi

  git checkout --force main
  sed -e 's/2/A/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'A'

  if [ "$1" = "--track-later" ]
  then
    git lfs track '*.dat'
    git add -u
    git commit -m 'A2'
  fi
}

setup_custom_repo () {
  git lfs track '*.dat'

  seq 1 10 > a.dat
  git add .gitattributes *.dat
  git commit -m 'Initial import'

  git checkout -b other

  # sed -i isn't portable.
  sed -e 's/2/B/' -e 's/9/B/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'B'

  git checkout main
  sed -e 's/2/A/' -e 's/9/A/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'A'
}

setup_conflicting_repo () {
  if [ "$1" != "--track-later" ]
  then
    git lfs track '*.dat'
  else
    touch .gitattributes
  fi

  seq 1 10 > a.dat
  git add .gitattributes *.dat
  git commit -m 'Initial import'

  git checkout -b other

  # sed -i isn't portable.
  sed -e 's/3/B/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'B'

  if [ "$1" = "--track-later" ]
  then
    git lfs track '*.dat'
    git add .gitattributes
    git commit -m 'B2'
  fi

  git checkout --force main
  sed -e 's/2/A/' a.dat > b.dat
  mv b.dat a.dat
  git add -u
  git commit -m 'A'

  if [ "$1" = "--track-later" ]
  then
    git lfs track '*.dat'
    git add -u
    git commit -m 'A2'
  fi
}

begin_test "merge-driver uses Git merge by default"
(
  set -e

  reponame="merge-driver-basic"
  git init "$reponame"
  cd "$reponame"

  result="07b26d7b3123467282635a68fdf9b59e81269cf9faf12282cedf30f393a55e5b"
  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A'

  setup_successful_repo

  git merge other

  (
    set -
    echo 1
    echo A
    seq 3 8
    echo B
    echo 10
  ) > expected.dat
  diff -u a.dat expected.dat
  assert_pointer "main" "a.dat" "$result" 21
  assert_local_object "$result" 21
)
end_test

begin_test "merge-driver uses Git merge when explicit"
(
  set -e

  reponame="merge-driver-explicit"
  git init "$reponame"
  cd "$reponame"

  result="07b26d7b3123467282635a68fdf9b59e81269cf9faf12282cedf30f393a55e5b"
  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A --program '\''git merge-file --stdout --marker-size=%%L %%A %%O %%B >%%D'\'''
  git lfs track '*.dat'

  setup_successful_repo

  git merge other

  (
    set -e
    echo 1
    echo A
    seq 3 8
    echo B
    echo 10
  ) > expected.dat
  diff -u a.dat expected.dat
  assert_pointer "main" "a.dat" "$result" 21
  assert_local_object "$result" 21
)
end_test

begin_test "merge-driver uses custom driver when explicit"
(
  set -e

  reponame="merge-driver-custom"
  git init "$reponame"
  cd "$reponame"

  result="07b26d7b3123467282635a68fdf9b59e81269cf9faf12282cedf30f393a55e5b"
  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A --program '\''(sed -n 1,5p %%A; sed -n 6,10p %%B) >%%D'\'''
  git lfs track '*.dat'

  setup_custom_repo

  git merge other

  (
    set -e
    echo 1
    echo A
    seq 3 8
    echo B
    echo 10
  ) > expected.dat
  diff -u a.dat expected.dat
  assert_pointer "main" "a.dat" "$result" 21
  assert_local_object "$result" 21
)
end_test

begin_test "merge-driver reports conflicts"
(
  set -e

  reponame="merge-driver-conflicts"
  git init "$reponame"
  cd "$reponame"

  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A --program '\''git merge-file --stdout --marker-size=%%L %%A %%O %%B >%%D'\'''
  git lfs track '*.dat'

  setup_conflicting_repo

  git merge other && exit 1
  sed -e 's/<<<<<<<.*/<<<<<<</' -e 's/>>>>>>>.*/>>>>>>>/' a.dat > actual.dat
  (
    set -e
    echo 1
    echo "<<<<<<<"
    echo A
    echo 3
    echo "======="
    echo 2
    echo B
    echo ">>>>>>>"
    seq 4 10
  ) > expected.dat
  diff -u actual.dat expected.dat
)
end_test

begin_test "merge-driver gracefully handles non-pointer"
(
  set -e

  reponame="merge-driver-non-pointer"
  git init "$reponame"
  cd "$reponame"

  result="07b26d7b3123467282635a68fdf9b59e81269cf9faf12282cedf30f393a55e5b"
  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A'

  setup_successful_repo --track-later

  git merge other

  (
    set -
    echo 1
    echo A
    seq 3 8
    echo B
    echo 10
  ) > expected.dat
  diff -u a.dat expected.dat
  assert_pointer "main" "a.dat" "$result" 21
  assert_local_object "$result" 21
)
end_test

begin_test "merge-driver reports conflicts with non-pointer"
(
  set -e

  reponame="conflicts-non-pointer"
  git init "$reponame"
  cd "$reponame"

  git config merge.lfs.driver 'git lfs merge-driver --ancestor %O --current %A --other %B --marker-size %L --output %A --program '\''git merge-file --stdout --marker-size=%%L %%A %%O %%B >%%D'\'''

  setup_conflicting_repo --track-later

  git merge other && exit 1
  sed -e 's/<<<<<<<.*/<<<<<<</' -e 's/>>>>>>>.*/>>>>>>>/' a.dat > actual.dat
  (
    set -e
    echo 1
    echo "<<<<<<<"
    echo A
    echo 3
    echo "======="
    echo 2
    echo B
    echo ">>>>>>>"
    seq 4 10
  ) > expected.dat
  diff -u actual.dat expected.dat
)
end_test
