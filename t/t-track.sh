#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "track"
(
  set -e

  # no need to setup a remote repo, since this test doesn't need to push or pull

  mkdir track
  cd track
  git init

  echo "###############################################################################
# Set default behavior to automatically normalize line endings.
###############################################################################
* text=auto

#*.cs     diff=csharp" > .gitattributes

  # track *.jpg once
  git lfs track "*.jpg" | grep "Tracking \"\*.jpg\""
  assert_attributes_count "jpg" "filter=lfs" 1

  # track *.jpg again
  git lfs track "*.jpg" | grep "\"*.jpg\" already supported"
  assert_attributes_count "jpg" "filter=lfs" 1

  mkdir -p a/b .git/info

  echo "*.mov filter=lfs -text" > .git/info/attributes
  echo "*.gif filter=lfs -text" > a/.gitattributes
  echo "*.png filter=lfs -text" > a/b/.gitattributes

  git lfs track | tee track.log
  grep "Listing tracked patterns" track.log
  grep "*.mov ($(native_path_escaped ".git/info/attributes"))" track.log
  grep "*.jpg (.gitattributes)" track.log
  grep "*.gif ($(native_path_escaped "a/.gitattributes"))" track.log
  grep "*.png ($(native_path_escaped "a/b/.gitattributes"))" track.log

  grep "Set default behavior" .gitattributes
  grep "############" .gitattributes
  grep "* text=auto" .gitattributes
  grep "diff=csharp" .gitattributes
  grep "*.jpg" .gitattributes

  echo "*.gif -filter -text" >> a/b/.gitattributes
  echo "*.mov -filter -text" >> a/b/.gitattributes

  git lfs track | tee track.log
  tail -n 3 track.log | head -n 1 | grep "Listing excluded patterns"
  tail -n 3 track.log | grep "*.gif ($(native_path_escaped "a/b/.gitattributes"))"
  tail -n 3 track.log | grep "*.mov ($(native_path_escaped "a/b/.gitattributes"))"
)
end_test

begin_test "track --no-excluded"
(
  set -e

  reponame="track_no_excluded"
  mkdir "$reponame"
  cd "$reponame"
  git init

  mkdir -p a/b .git/info

  echo "*.mov filter=lfs -text" > .git/info/attributes
  echo "*.gif filter=lfs -text" > a/.gitattributes
  echo "*.png filter=lfs -text" > a/b/.gitattributes
  echo "*.gif -filter -text" >> a/b/.gitattributes
  echo "*.mov -filter=lfs -text" >> a/b/.gitattributes

  git lfs track --no-excluded | tee track.log
  ! grep "Listing excluded patterns" track.log
)
end_test

begin_test "track --verbose"
(
  set -e

  reponame="track_verbose_logs"
  mkdir "$reponame"
  cd "$reponame"
  git init

  touch foo.dat
  git add foo.dat

  git lfs track --verbose "foo.dat" 2>&1 > track.log
  grep "touching \"foo.dat\"" track.log
)
end_test

begin_test "track --dry-run"
(
  set -e

  reponame="track_dry_run"
  mkdir "$reponame"
  cd "$reponame"
  git init

  touch foo.dat
  git add foo.dat

  git lfs track --dry-run "foo.dat" 2>&1 > track.log
  grep "Tracking \"foo.dat\"" track.log
  grep "Git LFS: touching \"foo.dat\"" track.log

  git status --porcelain 2>&1 > status.log
  grep "A  foo.dat" status.log
)
end_test

begin_test "track directory"
(
  set -e
  mkdir dir
  cd dir
  git init

  git lfs track "foo bar\\*" | tee track.txt
  [ "foo[[:space:]]bar/* filter=lfs diff=lfs merge=lfs -text" = "$(cat .gitattributes)" ]
  [ "Tracking \"foo bar/*\"" = "$(cat track.txt)" ]

  mkdir "foo bar"
  echo "a" > "foo bar/a"
  echo "b" > "foo bar/b"
  git add foo\ bar
  git commit -am "add foo bar"

  assert_pointer "main" "foo bar/a" "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7" 2
  assert_pointer "main" "foo bar/b" "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f" 2
)
end_test

begin_test "track without trailing linebreak"
(
  set -e

  mkdir no-linebreak
  cd no-linebreak
  git init

  printf "*.mov filter=lfs -text" > .gitattributes
  [ "*.mov filter=lfs -text" = "$(cat .gitattributes)" ]

  git lfs track "*.gif"
  expected="*.mov filter=lfs -text$(cat_end)
*.gif filter=lfs diff=lfs merge=lfs -text$(cat_end)"
  [ "$expected" = "$(cat -e .gitattributes)" ]
)
end_test

begin_test "track with existing crlf"
(
  set -e

  mkdir existing-crlf
  cd existing-crlf
  git init

  git config core.autocrlf true
  git lfs track "*.mov"
  git lfs track "*.gif"
  expected="*.mov filter=lfs diff=lfs merge=lfs -text^M$
*.gif filter=lfs diff=lfs merge=lfs -text^M$"
  [ "$expected" = "$(cat -e .gitattributes)" ]

  git config core.autocrlf false
  git lfs track "*.jpg"
  expected="*.mov filter=lfs diff=lfs merge=lfs -text^M$
*.gif filter=lfs diff=lfs merge=lfs -text^M$
*.jpg filter=lfs diff=lfs merge=lfs -text^M$"
  [ "$expected" = "$(cat -e .gitattributes)" ]
)
end_test

begin_test "track with autocrlf=true"
(
  set -e

  mkdir autocrlf-true
  cd autocrlf-true
  git init
  git config core.autocrlf true

  printf "*.mov filter=lfs -text" > .gitattributes
  [ "*.mov filter=lfs -text" = "$(cat .gitattributes)" ]

  git lfs track "*.gif"
  expected="*.mov filter=lfs -text^M$
*.gif filter=lfs diff=lfs merge=lfs -text^M$"
  [ "$expected" = "$(cat -e .gitattributes)" ]
)
end_test

begin_test "track with autocrlf=input"
(
  set -e

  mkdir autocrlf-input
  cd autocrlf-input
  git init
  git config core.autocrlf input

  printf "*.mov filter=lfs -text" > .gitattributes
  [ "*.mov filter=lfs -text" = "$(cat .gitattributes)" ]

  git lfs track "*.gif"
  if [ $IS_WINDOWS -eq 1 ]
  then
      cat -e .gitattributes | grep '\^M\$'
  else
      cat -e .gitattributes | grep -v '\^M'
  fi
)
end_test

begin_test "track outside git repo"
(
  set -e

  git lfs track "*.foo" || {
    # this fails if it's run outside of a git repo using GIT_LFS_TEST_DIR

    # git itself returns an exit status of 128
    # $ git show
    # fatal: Not a git repository (or any of the parent directories): .git
    # $ echo "$?"
    # 128

    [ "$?" = "128" ]
    exit 0
  }

  if [ -n "$GIT_LFS_TEST_DIR" ]; then
    echo "GIT_LFS_TEST_DIR should be set outside of any Git repository"
    exit 1
  fi

  git init track-outside
  cd track-outside

  git lfs track "*.file"

  git lfs track "../*.foo" || {

    # git itself returns an exit status of 128
    # $ git add ../test.foo
    # fatal: ../test.foo: '../test.foo' is outside repository
    # $ echo "$?"
    # 128

    [ "$?" = "128" ]
    exit 0
  }
  exit 1
)
end_test

begin_test "track representation"
(
  set -e

  git init track-representation
  cd track-representation

  git lfs track "*.jpg"

  mkdir a
  git lfs track "a/test.file"
  cd a
  out3=$(git lfs track "test.file")

  if [ "$out3" != "\"test.file\" already supported" ]; then
    echo "Track didn't recognize duplicate path"
    cat .gitattributes
    exit 1
  fi

  git lfs track "file.bin"
  cd ..
  out4=$(git lfs track "a/file.bin")
  if [ "$out4" != "\"a/file.bin\" already supported" ]; then
    echo "Track didn't recognize duplicate path"
    cat .gitattributes
    exit 1
  fi
)
end_test

begin_test "track absolute"
(
  # MinGW bash intercepts '/images' and passes 'C:/Program Files/Git/images' as arg!
  if [[ $(uname) == *"MINGW"* ]]; then
    echo "Skipping track absolute on Windows"
    exit 0
  fi

  set -e

  git init track-absolute
  cd track-absolute

  git lfs track "/images"
  cat .gitattributes
  grep "^/images" .gitattributes
)
end_test

begin_test "track in gitDir"
(
  set -e

  git init track-in-dot-git
  cd track-in-dot-git

  echo "some content" > test.file

  cd .git
  git lfs track "../test.file" || {
    # this fails if it's run inside a .git directory

    # git itself returns an exit status of 128
    # $ git add ../test.file
    # fatal: This operation must be run in a work tree
    # $ echo "$?"
    # 128

	[ "$?" = "128" ]
	exit 0
  }

  # fail if track passed
  exit 1
)
end_test

begin_test "track in symlinked dir"
(
  set -e

  git init track-symlinkdst
  ln -s track-symlinkdst track-symlinksrc
  cd track-symlinksrc

  git lfs track "*.png"
  grep "^*.png" .gitattributes || {
    echo ".gitattributes doesn't contain the expected relative path *.png:"
    cat .gitattributes
    exit 1
  }
)
end_test

begin_test "track blocklisted files by name"
(
  set -e

  repo="track_blocklisted_by_name"
  mkdir "$repo"
  cd "$repo"
  git init

  touch .gitattributes
  git add .gitattributes

  git lfs track .gitattributes 2>&1 > track.log
  grep "Pattern .gitattributes matches forbidden file .gitattributes" track.log
)
end_test

begin_test "track blocklisted files with glob"
(
  set -e

  repo="track_blocklisted_glob"
  mkdir "$repo"
  cd "$repo"
  git init

  touch .gitattributes
  git add .gitattributes

  git lfs track ".git*" 2>&1 > track.log
  grep "Pattern .git\* matches forbidden file" track.log
)
end_test

begin_test "track lockable"
(
  set -e

  repo="track_lockable"
  mkdir "$repo"
  cd "$repo"
  git init

  # track *.jpg once, lockable
  git lfs track --lockable "*.jpg" | grep "Tracking \"\*.jpg\""
  assert_attributes_count "jpg" "lockable" 1
  # track *.jpg again, don't change anything. Should retain lockable
  git lfs track "*.jpg" | grep "\"*.jpg\" already supported"
  assert_attributes_count "jpg" "lockable" 1


  # track *.png once, not lockable yet
  git lfs track "*.png" | grep "Tracking \"\*.png\""
  assert_attributes_count "png" "filter=lfs" 1
  assert_attributes_count "png" "lockable" 0

  # track png again, enable lockable, should replace
  git lfs track --lockable "*.png" | grep "Tracking \"\*.png\""
  assert_attributes_count "png" "filter=lfs" 1
  assert_attributes_count "png" "lockable" 1

  # track png again, disable lockable, should replace
  git lfs track --not-lockable "*.png" | grep "Tracking \"\*.png\""
  assert_attributes_count "png" "filter=lfs" 1
  assert_attributes_count "png" "lockable" 0

  # check output reflects lockable
  out=$(git lfs track)
  echo "$out" | grep "Listing tracked patterns"
  echo "$out" | grep "*.jpg \[lockable\] (.gitattributes)"
  echo "$out" | grep "*.png (.gitattributes)"
)
end_test

begin_test "track lockable read-only/read-write"
(
  set -e

  repo="track_lockable_ro_rw"
  mkdir "$repo"
  cd "$repo"
  git init

  echo "blah blah" > test.bin
  echo "foo bar" > test.dat
  mkdir subfolder
  echo "sub blah blah" > subfolder/test.bin
  echo "sub foo bar" > subfolder/test.dat
  # should start writeable
  assert_file_writeable test.bin
  assert_file_writeable test.dat
  assert_file_writeable subfolder/test.bin
  assert_file_writeable subfolder/test.dat

  # track *.bin, not lockable yet
  git lfs track "*.bin" | grep "Tracking \"\*.bin\""
  # track *.dat, lockable immediately
  git lfs track --lockable "*.dat" | grep "Tracking \"\*.dat\""

  # bin should remain writeable, dat should have been made read-only

  assert_file_writeable test.bin
  refute_file_writeable test.dat
  assert_file_writeable subfolder/test.bin
  refute_file_writeable subfolder/test.dat

  git add .gitattributes test.bin test.dat
  git commit -m "First commit"

  # bin should still be writeable
  assert_file_writeable test.bin
  assert_file_writeable subfolder/test.bin
  # now make bin lockable
  git lfs track --lockable "*.bin" | grep "Tracking \"\*.bin\""
  # bin should now be read-only
  refute_file_writeable test.bin
  refute_file_writeable subfolder/test.bin

  # remove lockable again
  git lfs track --not-lockable "*.bin" | grep "Tracking \"\*.bin\""
  # bin should now be writeable again
  assert_file_writeable test.bin
  assert_file_writeable subfolder/test.bin
)
end_test

begin_test "track escaped pattern"
(
  set -e

  reponame="track-escaped-pattern"
  git init "$reponame"
  cd "$reponame"

  git lfs track " " | grep "Tracking \" \""
  assert_attributes_count "[[:space:]]" "filter=lfs" 1

  git lfs track "#" | grep "Tracking \"#\""
  assert_attributes_count "\\#" "filter=lfs" 1
)
end_test

begin_test "track (symlinked repository)"
(
  set -e

  reponame="tracked-symlinked-repository"
  git init "$reponame"
  cd "$reponame"

  touch a.dat

  pushd .. > /dev/null
    dir="tracked-symlinked-repository-tmp"

    mkdir -p "$dir"

    ln -s "../$reponame" "./$dir"

    cd "$dir/$reponame"

    [ "Tracking \"a.dat\"" = "$(git lfs track "a.dat")" ]
    [ "\"a.dat\" already supported" = "$(git lfs track "a.dat")" ]
  popd > /dev/null
)
end_test

begin_test "track (\$GIT_LFS_TRACK_NO_INSTALL_HOOKS)"
(
  set -e

  reponame="track-no-setup-hooks"
  git init "$reponame"
  cd "$reponame"

  [ ! -f .git/hooks/pre-push ]
  [ ! -f .git/hooks/post-checkout ]
  [ ! -f .git/hooks/post-commit ]
  [ ! -f .git/hooks/post-merge ]

  GIT_LFS_TRACK_NO_INSTALL_HOOKS=1 git lfs track

  [ ! -f .git/hooks/pre-push ]
  [ ! -f .git/hooks/post-checkout ]
  [ ! -f .git/hooks/post-commit ]
  [ ! -f .git/hooks/post-merge ]
)
end_test

begin_test "track (with comments)"
(
  set -e

  reponame="track-with=comments"
  git init "$reponame"
  cd "$reponame"

  echo "*.jpg filter=lfs diff=lfs merge=lfs -text" >> .gitattributes
  echo "# *.png filter=lfs diff=lfs merge=lfs -text" >> .gitattributes
  echo "*.pdf filter=lfs diff=lfs merge=lfs -text" >> .gitattributes

  git add .gitattributes
  git commit -m "initial commit"

  git lfs track 2>&1 | tee track.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "expected \`git lfs track\` command to exit cleanly, didn't"
    exit 1
  fi

  [ "1" -eq "$(grep -c "\.jpg" track.log)" ]
  [ "1" -eq "$(grep -c "\.pdf" track.log)" ]
  [ "0" -eq "$(grep -c "\.png" track.log)" ]
)
end_test

begin_test "track (with current-directory prefix)"
(
  set -e

  reponame="track-with-current-directory-prefix"
  git init "$reponame"
  cd "$reponame"

  git lfs track "./a.dat"
  printf "a" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  grep -e "^a.dat" .gitattributes
)
end_test

begin_test "track (global gitattributes)"
(
  set -e

  reponame="track-global-gitattributes"
  git init "$reponame"
  cd "$reponame"

  global="$(cd .. && pwd)/gitattributes-global"

  echo "*.dat filter=lfs diff=lfs merge=lfs -text" > "$global"
  git config --local core.attributesfile "$global"

  git lfs track 2>&1 | tee track.log
  grep "*.dat" track.log
)
end_test

begin_test "track (system gitattributes)"
(
  set -e

  reponame="track-system-gitattributes"
  git init "$reponame"
  cd "$reponame"

  pushd "$TRASHDIR" > /dev/null
    mkdir -p "prefix/${reponame}/etc"
    cd "prefix/${reponame}/etc"
    echo "*.dat filter=lfs diff=lfs merge=lfs -text" > gitattributes
  popd > /dev/null

  PREFIX="${TRASHDIR}/prefix/${reponame}" git lfs track 2>&1 | tee track.log
  grep "*.dat" track.log
)
end_test

begin_test "track: escaped pattern in .gitattributes"
(
  set -e

  reponame="track-escaped"
  git init "$reponame"
  cd "$reponame"

  filename="file with spaces.#"

  echo "I need escaping" > "$filename"

  [ "Tracking \"$filename\"" = "$(git lfs track "$filename")" ]
  [ "\"$filename\" already supported" = "$(git lfs track "$filename")" ]

  #changing flags should track the file again
  [ "Tracking \"$filename\"" = "$(git lfs track -l "$filename")" ]

  if [ 1 -ne "$(wc -l .gitattributes | awk '{ print $1 }')" ]; then
    echo >&2 "changing flag for an existing tracked file shouldn't add another line"
    exit 1
  fi
)
end_test

begin_test "track: escaped glob pattern in .gitattributes"
(
  set -e

  # None of these characters are valid in the Win32 subsystem.
  [ "$IS_WINDOWS" -eq 1 ] && exit 0

  reponame="track-escaped-glob"
  git init "$reponame"
  cd "$reponame"

  filename='*[foo]bar?.txt'
  contents='I need escaping'
  contents_oid=$(calc_oid "$contents")

  git lfs track --filename "$filename"
  git lfs track --filename "$filename" | grep 'already supported'
  git add .
  cat .gitattributes

  printf "%s" "$contents" > "$filename"
  git add .
  git commit -m 'Add unusually named file'

  # If Git understood our escaping, we'll have a pointer. Otherwise, we won't.
  assert_pointer "main" "$filename" "$contents_oid" 15
)
end_test

begin_test "track: escaped glob pattern with spaces in .gitattributes"
(
  set -e

  # None of these characters are valid in the Win32 subsystem.
  [ "$IS_WINDOWS" -eq 1 ] && exit 0

  reponame="track-escaped-glob"
  git init "$reponame"
  cd "$reponame"

  filename="*[foo] bar?.txt"
  contents='I need escaping'
  contents_oid=$(calc_oid "$contents")

  git lfs track --filename "$filename"
  git add .
  cat .gitattributes

  printf "%s" "$contents" > "$filename"
  git add .
  git commit -m 'Add unusually named file'

  # If Git understood our escaping, we'll have a pointer. Otherwise, we won't.
  assert_pointer "main" "$filename" "$contents_oid" 15
)
end_test
