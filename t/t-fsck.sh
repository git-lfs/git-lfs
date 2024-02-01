#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "fsck default"
(
  set -e

  reponame="fsck-default"
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  [ "Git LFS fsck OK" = "$(git lfs fsck)" ]

  aOid=$(git log --patch a.dat | grep "^+oid" | cut -d ":" -f 2)
  aOid12=$(echo $aOid | cut -b 1-2)
  aOid34=$(echo $aOid | cut -b 3-4)
  if [ "$aOid" != "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat does not match"
    exit 1
  fi

  bOid=$(git log --patch b.dat | grep "^+oid" | cut -d ":" -f 2)
  bOid12=$(echo $bOid | cut -b 1-2)
  bOid34=$(echo $bOid | cut -b 3-4)
  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi

  echo "CORRUPTION" >> .git/lfs/objects/$aOid12/$aOid34/$aOid

  moved=$(canonical_path "$TRASHDIR/$reponame/.git/lfs/bad")
  expected="$(printf 'objects: corruptObject: a.dat (%s) is corrupt
objects: repair: moving corrupt objects to %s' "$aOid" "$moved")"
  [ "$expected" = "$(git lfs fsck)" ]

  [ -e ".git/lfs/bad/$aOid" ]
  [ ! -e ".git/lfs/objects/$aOid12/$aOid34/$aOid" ]
  [ "$bOid" = "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]
)
end_test

begin_test "fsck dry run"
(
  set -e

  reponame="fsck-dry-run"
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  [ "Git LFS fsck OK" = "$(git lfs fsck --dry-run)" ]

  aOid=$(git log --patch a.dat | grep "^+oid" | cut -d ":" -f 2)
  aOid12=$(echo $aOid | cut -b 1-2)
  aOid34=$(echo $aOid | cut -b 3-4)
  if [ "$aOid" != "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat does not match"
    exit 1
  fi

  bOid=$(git log --patch b.dat | grep "^+oid" | cut -d ":" -f 2)
  bOid12=$(echo $bOid | cut -b 1-2)
  bOid34=$(echo $bOid | cut -b 3-4)
  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi

  echo "CORRUPTION" >> .git/lfs/objects/$aOid12/$aOid34/$aOid

  [ "objects: corruptObject: a.dat ($aOid) is corrupt" = "$(git lfs fsck --dry-run)" ]

  if [ "$aOid" = "$(calc_oid_file .git/lfs/objects/$aOid12/$aOid34/$aOid)" ]; then
    echo "oid for a.dat still matches match"
    exit 1
  fi

  if [ "$bOid" != "$(calc_oid_file .git/lfs/objects/$bOid12/$bOid34/$bOid)" ]; then
    echo "oid for b.dat does not match"
    exit 1
  fi
)
end_test

begin_test "fsck does not fail with shell characters in paths"
(
  set -e

  mkdir '[[path]]'
  cd '[[path]]'
  reponame="fsck-shell-paths"
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  # Verify that the pack code handles glob patterns properly.
  git gc --aggressive --prune=now

  [ "Git LFS fsck OK" = "$(git lfs fsck)" ]
)
end_test

begin_test "fsck: outside git repository"
(
  set +e
  git lfs fsck 2>&1 > fsck.log
  res=$?
  set -e

  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a Git repository" fsck.log
)
end_test

create_invalid_pointers() {
  valid="$1"
  ext="${2:-dat}"

  git cat-file blob ":$valid" | awk '{ sub(/$/, "\r"); print }' >"crlf.$ext"
  base64 < /dev/urandom | head -c 1025 >"large.$ext"
  git \
    -c "filter.lfs.process=" \
    -c "filter.lfs.clean=cat" \
    -c "filter.lfs.required=false" \
    add "crlf.$ext" "large.$ext"
  git commit -m "invalid pointers"
}

setup_invalid_pointers () {
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  git add .gitattributes *.dat
  git commit -m "first commit"

  create_invalid_pointers "a.dat"
}

begin_test "fsck detects invalid pointers"
(
  set -e

  reponame="fsck-pointers"
  setup_invalid_pointers

  set +e
  git lfs fsck >test.log 2>&1
  RET=$?
  git lfs fsck --pointers >>test.log 2>&1
  RET2=$?
  set -e

  [ "$RET" -eq 1 ]
  [ "$RET2" -eq 1 ]
  [ $(grep -c 'pointer: nonCanonicalPointer: Pointer.*was not canonical' test.log) -eq 2 ]
  [ $(grep -c 'pointer: unexpectedGitObject: "large.dat".*should have been a pointer but was not' test.log) -eq 2 ]
)
end_test

begin_test "fsck detects invalid pointers with macro patterns"
(
  set -e

  reponame="fsck-pointers-macros"
  git init $reponame
  cd $reponame

  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n*.dat lfs\n' \
    >.gitattributes
  echo "test data" >a.dat
  mkdir dir
  printf '*.bin lfs\n' >dir/.gitattributes
  git add .gitattributes a.dat dir
  git commit -m "first commit"

  create_invalid_pointers "a.dat"

  cd dir
  create_invalid_pointers "a.dat" "bin"
  cd ..

  # NOTE: We should also create a .dir directory with the same files as
  #       as in the dir/ directory, and confirm those .dir/*.bin files are
  #       reported by "git lfs fsck" as well.  However, at the moment
  #       "git lfs fsck" will not resolve a macro attribute reference
  #       in .dir/.gitattributes because it sorts that file before
  #       .gitattributes and then processes them in that order.

  set +e
  git lfs fsck >test.log 2>&1
  RET=$?
  git lfs fsck --pointers >>test.log 2>&1
  RET2=$?
  set -e

  [ "$RET" -eq 1 ]
  [ "$RET2" -eq 1 ]
  [ $(grep -c 'pointer: nonCanonicalPointer: Pointer.*was not canonical' test.log) -eq 4 ]
  [ $(grep -c 'pointer: unexpectedGitObject: "large.dat".*should have been a pointer but was not' test.log) -eq 2 ]
  [ $(grep -c 'pointer: unexpectedGitObject: "dir/large.bin".*should have been a pointer but was not' test.log) -eq 2 ]
)
end_test

begin_test "fsck detects invalid pointers with GIT_OBJECT_DIRECTORY"
(
  set -e

  reponame="fsck-pointers-object-directory"
  setup_invalid_pointers

  head=$(git rev-parse HEAD)
  objdir="$(lfstest-realpath .git/objects)"
  cd ..
  git init "$reponame-2"
  gitdir="$(lfstest-realpath "$reponame-2/.git")"
  GIT_WORK_TREE="$reponame-2" GIT_DIR="$gitdir" GIT_OBJECT_DIRECTORY="$objdir" git update-ref refs/heads/main "$head"

  set +e
  GIT_WORK_TREE="$reponame-2" GIT_DIR="$gitdir" GIT_OBJECT_DIRECTORY="$objdir" git lfs fsck --pointers >test.log 2>&1
  RET=$?
  set -e

  [ "$RET" -eq 1 ]
  grep 'pointer: nonCanonicalPointer: Pointer.*was not canonical' test.log
  grep 'pointer: unexpectedGitObject: "large.dat".*should have been a pointer but was not' test.log
)
end_test

begin_test "fsck does not detect invalid pointers with no LFS objects"
(
  set -e

  reponame="fsck-pointers-none"
  git init "$reponame"
  cd "$reponame"

  echo "# README" > README.md
  git add README.md
  git commit -m "Add README"

  git lfs fsck
  git lfs fsck --pointers
)
end_test

begin_test "fsck does not detect invalid pointers with symlinks"
(
  set -e

  reponame="fsck-pointers-symlinks"
  git init "$reponame"
  cd "$reponame"

  git lfs track '*.dat'

  echo "# Test" > a.dat
  ln -s a.dat b.dat
  git add .gitattributes *.dat
  git commit -m "Add files"

  git lfs fsck
  git lfs fsck --pointers
)
end_test

begin_test "fsck does not detect invalid pointers with negated patterns"
(
  set -e

  reponame="fsck-pointers-none"
  git init "$reponame"
  cd "$reponame"

  cat > .gitattributes <<EOF
*.dat filter=lfs diff=lfs merge=lfs -text
b.dat !filter !diff !merge text
EOF

  echo "# Test" > a.dat
  cp a.dat b.dat
  git add .gitattributes *.dat
  git commit -m "Add files"

  git lfs fsck
  git lfs fsck --pointers
)
end_test

begin_test "fsck does not detect invalid pointers with negated macro patterns"
(
  set -e

  reponame="fsck-pointers-macros-none"
  git init "$reponame"
  cd "$reponame"

  printf '[attr]lfs filter=lfs diff=lfs merge=lfs -text\n*.dat lfs\nb.dat !lfs\n' \
    >.gitattributes
  echo "test data" >a.dat
  cp a.dat b.dat
  mkdir dir .dir
  printf '*.dat !lfs\n' >dir/.gitattributes
  cp b.dat dir
  printf '*.dat !lfs\n' >.dir/.gitattributes
  cp b.dat .dir
  git add .gitattributes *.dat dir .dir
  git commit -m "first commit"

  # NOTE: The "git lfs fsck" command exempts the .dir/b.dat file from the
  #       *.dat pattern from the top-level .gitattributes and so permits
  #       it as a valid non-pointer file; however, it permits it for a
  #       different reason than the dir/b.dat file, because it processes
  #       the .dir/.gitattributes file before the .gitattributes one
  #       and does not recognize the "!lfs" macro attribute reference until
  #       after it has processed .gitattributes.  Ideally both the dir/
  #       and .dir/ directories should be processed identically.

  git lfs fsck
  git lfs fsck --pointers
)
end_test

setup_invalid_objects () {
  git init $reponame
  cd $reponame

  # Create a commit with some files tracked by git-lfs
  git lfs track *.dat
  echo "test data" > a.dat
  echo "test data 2" > b.dat
  mkdir foo
  echo "test test 3" > foo/a.dat
  echo "test data 4" > foo/b.dat
  git add .gitattributes *.dat foo
  git commit -m "first commit"

  oid1=$(calc_oid_file a.dat)
  oid2=$(calc_oid_file b.dat)
  oid3=$(calc_oid_file foo/a.dat)
  oid4=$(calc_oid_file foo/b.dat)
  echo "CORRUPTION" >>".git/lfs/objects/${oid1:0:2}/${oid1:2:2}/$oid1"
  rm ".git/lfs/objects/${oid2:0:2}/${oid2:2:2}/$oid2"
  echo "CORRUPTION" >>".git/lfs/objects/${oid3:0:2}/${oid3:2:2}/$oid3"
  rm ".git/lfs/objects/${oid4:0:2}/${oid4:2:2}/$oid4"
}

begin_test "fsck detects invalid objects"
(
  set -e

  reponame="fsck-objects"
  setup_invalid_objects

  set +e
  git lfs fsck >test.log 2>&1
  RET=$?
  set -e

  [ "$RET" -eq 1 ]
  [ $(grep -c 'objects: corruptObject: a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: corruptObject: foo/a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: foo/b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: repair: moving corrupt objects to .*' test.log) -eq 1 ]

  cd ..
  rm -rf $reponame
  setup_invalid_objects

  set +e
  git lfs fsck --objects >test.log 2>&1
  RET=$?
  set -e

  [ "$RET" -eq 1 ]
  [ $(grep -c 'objects: corruptObject: a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: corruptObject: foo/a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: foo/b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: repair: moving corrupt objects to .*' test.log) -eq 1 ]
)
end_test

begin_test "fsck detects invalid objects except in excluded paths"
(
  set -e

  reponame="fsck-objects-exclude"
  setup_invalid_objects

  # We need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config "lfs.fetchexclude" "/foo"

  set +e
  git lfs fsck >test.log 2>&1
  RET=$?
  set -e

  [ "$RET" -eq 1 ]
  [ $(grep -c 'objects: corruptObject: a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: corruptObject: foo/a.dat (.*) is corrupt' test.log) -eq 0 ]
  [ $(grep -c 'objects: openError: foo/b.dat (.*) could not be checked: .*' test.log) -eq 0 ]
  [ $(grep -c 'objects: repair: moving corrupt objects to .*' test.log) -eq 1 ]

  cd ..
  rm -rf $reponame
  setup_invalid_objects

  # We need to prevent MSYS from rewriting /foo into a Windows path.
  MSYS_NO_PATHCONV=1 git config "lfs.fetchexclude" "/foo"

  set +e
  git lfs fsck --objects >test.log 2>&1
  RET=$?
  set -e

  [ "$RET" -eq 1 ]
  [ $(grep -c 'objects: corruptObject: a.dat (.*) is corrupt' test.log) -eq 1 ]
  [ $(grep -c 'objects: openError: b.dat (.*) could not be checked: .*' test.log) -eq 1 ]
  [ $(grep -c 'objects: corruptObject: foo/a.dat (.*) is corrupt' test.log) -eq 0 ]
  [ $(grep -c 'objects: openError: foo/b.dat (.*) could not be checked: .*' test.log) -eq 0 ]
  [ $(grep -c 'objects: repair: moving corrupt objects to .*' test.log) -eq 1 ]
)
end_test

begin_test "fsck does not detect invalid objects with no LFS objects"
(
  set -e

  reponame="fsck-objects-none"
  git init "$reponame"
  cd "$reponame"

  echo "# README" > README.md
  git add README.md
  git commit -m "Add README"

  git lfs fsck
  git lfs fsck --objects
)
end_test

begin_test "fsck operates on specified refs"
(
  set -e

  reponame="fsck-refs"
  setup_invalid_pointers

  git rm -f crlf.dat large.dat
  echo "# Test" > new.dat
  git add new.dat
  git commit -m 'third commit'

  git commit --allow-empty -m 'fourth commit'

  # Should succeed.  (HEAD and index).

  git lfs fsck
  git lfs fsck HEAD
  git lfs fsck HEAD^^ && exit 1
  git lfs fsck HEAD^
  git lfs fsck HEAD^..HEAD
  git lfs fsck HEAD^^^..HEAD && exit 1
  git lfs fsck HEAD^^^..HEAD^ && exit 1

  git lfs fsck --pointers HEAD^^^..HEAD^^ >test.log 2>&1 && exit 1

  grep 'pointer: nonCanonicalPointer: Pointer.*was not canonical' test.log
  grep 'pointer: unexpectedGitObject: "large.dat".*should have been a pointer but was not' test.log

  oid=$(calc_oid_file new.dat)
  echo "CORRUPTION" >>".git/lfs/objects/${oid:0:2}/${oid:2:2}/$oid"

  git lfs fsck --objects HEAD^^..HEAD^ >test.log 2>&1 && exit 1

  grep 'objects: corruptObject: new.dat (.*) is corrupt' test.log
  grep 'objects: repair: moving corrupt objects to .*' test.log

  # Make the result of the subshell a success.
  true
)
end_test

begin_test "fsck detects invalid ref"
(
  set -e
  reponame="fsck-default"
  git init $reponame
  cd $reponame

  git lfs fsck jibberish >fsck.log 2>&1 && exit 1
  grep "can't resolve ref" fsck.log
)
end_test
