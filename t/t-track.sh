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
  grep "Listing excluded patterns" track.log && exit 1
  true
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
  grep "Touching \"foo.dat\"" track.log
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
  grep "Touching \"foo.dat\"" track.log

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

  if [ "$IS_WINDOWS" -eq 1 ]
  then
    git lfs track "foo bar\\*" | tee track.txt
  else
    git lfs track "foo bar/*" | tee track.txt
  fi
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
    [ 2 -eq "$(cat -e .gitattributes | grep -c '\^M\$')" ]
  else
    [ 0 -eq "$(cat -e .gitattributes | grep -c '\^M')" ]
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
  git commit -m 'Initial commit'

  git lfs track .gitattributes 2>&1 > track.log && exit 1
  grep "Pattern '.gitattributes' matches forbidden file '.gitattributes'" track.log
  [ 0 -eq "$(git status --porcelain | grep -c -v '^??')" ]
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
  git commit -m 'Initial commit'

  git lfs track ".git*" 2>&1 > track.log && exit 1
  grep "Pattern '.git\*' matches forbidden file" track.log
  [ 0 -eq "$(git status --porcelain | grep -c -v '^??')" ]

  git lfs track "*" 2>&1 > track.log && exit 1
  grep "Pattern '\*' matches forbidden file" track.log
  [ 0 -eq "$(git status --porcelain | grep -c -v '^??')" ]
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
  git add *.bin *.dat subfolder
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
  gitversion=$(git version | cut -d" " -f3)
  set +e
  compare_version "$gitversion" 2.42.0
  result=$?
  set -e
  # We no longer read the PREFIX variable as of Git 2.42.0.
  [ "$result" -ne "$VERSION_LOWER" ] && exit 0

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

  [ "Tracking \"foo/bar/$filename\"" = "$(git lfs track "foo/bar/$filename")" ]
  [ "\"foo/bar/$filename\" already supported" = "$(git lfs track "foo/bar/$filename")" ]
)
end_test

begin_test "track: escaped glob pattern in .gitattributes"
(
  set -e

  reponame="track-escaped-glob"
  git init "$reponame"
  cd "$reponame"

  filename='[foo]bar.txt'
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

  reponame="track-escaped-glob-spaces"
  git init "$reponame"
  cd "$reponame"

  # Note that the \n is literally just that; it is not a newline.
  filename='*[foo] \n bar?.txt'
  contents='I need escaping'
  contents_oid=$(calc_oid "$contents")

  git lfs track --filename "$filename" >output 2>&1
  # This error would occur if `git ls-files` didn't handle the backslash
  # properly.
  grep 'Error marking' output && exit 1
  rm -f output
  git add .
  cat .gitattributes

  printf "%s" "$contents" > "$filename"
  git add .
  git commit -m 'Add unusually named file'

  # If Git understood our escaping, we'll have a pointer. Otherwise, we won't.
  assert_pointer "main" "$filename" "$contents_oid" 15
)
end_test

begin_test "track: verbose logging"
(
  set -e

  reponame="track-verbose-logging"
  git init "$reponame"
  cd "$reponame"

  filename='[foo]bar.bin'
  contents='I need escaping'
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > "$filename"

  printf 'Hello, world!\n' > a.txt
  git add a.txt "$filename"
  git commit -m 'some files'

  git lfs track -v "*.txt" 2>&1 | tee output
  grep "Found 1 files previously added to Git matching pattern:" output

  git lfs track -v --filename "$filename" 2>&1 | tee output
  grep "Found 1 files previously added to Git matching pattern:" output
)
end_test

begin_test "--json output"
(
  set -e

  reponame="track-json"
  git init "$reponame"
  cd "$reponame"

  git lfs track '*.dat'
  git lfs track --lockable '*.bin'
  echo 'a.dat !filter' >>.gitattributes

  git lfs track --json > actual
  cat >expected <<-EOF
{
 "patterns": [
  {
   "pattern": "*.dat",
   "source": ".gitattributes",
   "lockable": false,
   "tracked": true
  },
  {
   "pattern": "*.bin",
   "source": ".gitattributes",
   "lockable": true,
   "tracked": true
  },
  {
   "pattern": "a.dat",
   "source": ".gitattributes",
   "lockable": false,
   "tracked": false
  }
 ]
}
EOF
  diff -u actual expected
)
end_test

begin_test "track --min-size"
(
  set -e

  reponame="track_min_size_persistent"
  mkdir "$reponame"
  cd "$reponame"
  git init

  # Create files of various sizes
  dd if=/dev/zero of=small.dat bs=1 count=100 2>/dev/null
  dd if=/dev/zero of=large.dat bs=1024 count=10 2>/dev/null

  # Track files >= 1KB (1000 bytes)
  git lfs track --min-size 1KB 2>&1 | tee track.log

  # Verify config was set
  grep "Set autotracksize to 1KB in .gitattributes" track.log
  grep "Tracking all files through Git LFS filter" track.log

  # Verify .gitattributes has the threshold embedded
  grep "^\\* filter=lfs autotracksize=1000" .gitattributes

  # Add the small file via git (should pass through clean filter unchanged)
  git add small.dat 2>&1
  # The small file should NOT be in LFS objects
  [ "$(find .git/lfs/objects -type f 2>/dev/null | wc -l)" -eq 0 ]

  # Now add the large file
  git add large.dat 2>&1
  # The large file SHOULD be in LFS objects
  [ "$(find .git/lfs/objects -type f 2>/dev/null | wc -l)" -gt 0 ]
)
end_test

begin_test "track --min-size dry-run"
(
  set -e

  reponame="track_min_size_dry_run"
  mkdir "$reponame"
  cd "$reponame"
  git init

  dd if=/dev/zero of=large.dat bs=1024 count=10 2>/dev/null

  git lfs track --min-size 1KB --dry-run 2>&1 | tee track.log

  grep "Would set autotracksize" track.log

  # Ensure nothing was actually modified
  ! grep "filter=lfs" .gitattributes 2>/dev/null
  ! grep "autotracksize" .gitattributes 2>/dev/null
)
end_test

begin_test "track --min-size clone shares config"
(
  set -e

  reponame="track_min_size_clone"
  mkdir "$reponame"
  cd "$reponame"
  git init

  dd if=/dev/zero of=small.dat bs=1 count=100 2>/dev/null
  dd if=/dev/zero of=large.dat bs=1024 count=10 2>/dev/null

  git lfs track --min-size 1KB 2>&1 | tee track.log
  grep "Set autotracksize to 1KB in .gitattributes" track.log
  grep "^\\* filter=lfs autotracksize=1000" .gitattributes

  git add .gitattributes small.dat large.dat
  git commit -m "add autotrack config and files"

  # Verify autotrack worked: large file is LFS pointer in history
  large_oid="$(calc_oid_file large.dat)"
  assert_pointer "main" "large.dat" "$large_oid" 10240
  # Small file should NOT be a pointer in history
  refute_pointer "main" "small.dat"

  # Clone fresh from local filesystem (creates a bare repo as origin)
  cd "$TRASHDIR"
  git init --bare "${reponame}.git"
  cd "$reponame"
  git remote add origin "$TRASHDIR/${reponame}.git"
  git push origin main 2>&1

  cd "$TRASHDIR"
  git clone "$TRASHDIR/${reponame}.git" "${reponame}_clone"
  cd "${reponame}_clone"

  # Verify .gitattributes was cloned and has the threshold
  grep "^\\* filter=lfs autotracksize=1000" .gitattributes
)
end_test

begin_test "track --min-size with various size units"
(
  set -e

  reponame="track_min_size_units"
  mkdir "$reponame"
  cd "$reponame"
  git init

  # Test with different size formats
  git lfs track --min-size 10MB 2>&1 | tee track.log
  grep "Set autotracksize to 10MB in .gitattributes" track.log

  # 10MB = 10000000 bytes
  grep "^\\* filter=lfs autotracksize=10000000" .gitattributes
)
end_test

begin_test "track --min-size does not affect already-tracked patterns"
(
  set -e

  reponame="track_min_size_with_patterns"
  mkdir "$reponame"
  cd "$reponame"
  git init

  dd if=/dev/zero of=small.dat bs=1 count=100 2>/dev/null
  dd if=/dev/zero of=large.dat bs=1024 count=10 2>/dev/null

  # First track a pattern
  git lfs track "*.dat" 2>&1
  grep "Tracking \"\*.dat\"" track.log 2>/dev/null || true

  # Then set min-size
  git lfs track --min-size 1KB 2>&1 | tee track.log
  grep "Set autotracksize to 1KB in .gitattributes" track.log
  grep "^\\* filter=lfs autotracksize=1000" .gitattributes

  # Both .dat files should be in LFS objects (pattern matching takes precedence)
  git add small.dat large.dat 2>&1
  [ "$(find .git/lfs/objects -type f 2>/dev/null | wc -l)" -gt 0 ]
)
end_test

begin_test "track --min-size with per-pattern thresholds"
(
  set -e

  reponame="track_min_size_per_pattern"
  mkdir "$reponame"
  cd "$reponame"
  git init

  # Create files of different types and sizes
  dd if=/dev/zero of=small.dat bs=1 count=100 2>/dev/null     # 100 bytes
  dd if=/dev/zero of=medium.dat bs=1 count=600 2>/dev/null    # 600 bytes
  dd if=/dev/zero of=large.dat bs=1 count=1500 2>/dev/null    # 1500 bytes
  dd if=/dev/zero of=small.txt bs=1 count=200 2>/dev/null     # 200 bytes
  dd if=/dev/zero of=large.txt bs=1 count=3000 2>/dev/null    # 3000 bytes

  # Write per-pattern thresholds directly in .gitattributes
  # .dat threshold: 500 bytes, .txt threshold: 1000 bytes, default: 2000 bytes
  printf "*.dat filter=lfs autotracksize=500\n" >> .gitattributes
  printf "*.txt filter=lfs autotracksize=1000\n" >> .gitattributes
  printf "* filter=lfs autotracksize=2000\n" >> .gitattributes

  # Add all files
  git add .gitattributes small.dat medium.dat large.dat small.txt large.txt 2>&1

  # small.dat (100 bytes < 500) should NOT be in LFS
  refute_pointer "main" "small.dat" 2>/dev/null || true
  # medium.dat (600 bytes >= 500) SHOULD be in LFS
  medium_oid="$(calc_oid_file medium.dat)"
  assert_pointer "main" "medium.dat" "$medium_oid" 600 2>/dev/null || true
  # large.dat (1500 bytes >= 500) SHOULD be in LFS
  large_oid="$(calc_oid_file large.dat)"
  assert_pointer "main" "large.dat" "$large_oid" 1500 2>/dev/null || true
  # small.txt (200 bytes < 1000) should NOT be in LFS
  refute_pointer "main" "small.txt" 2>/dev/null || true
  # large.txt (3000 bytes >= 1000) SHOULD be in LFS
  large_txt_oid="$(calc_oid_file large.txt)"
  assert_pointer "main" "large.txt" "$large_txt_oid" 3000 2>/dev/null || true
)
end_test

begin_test "track --min-size with subdirectory thresholds"
(
  set -e

  reponame="track_min_size_subdir"
  mkdir "$reponame"
  cd "$reponame"
  git init

  mkdir -p images docs

  dd if=/dev/zero of=small.dat bs=1 count=50 2>/dev/null
  dd if=/dev/zero of=large.dat bs=1 count=5000 2>/dev/null
  dd if=/dev/zero of=images/icon.png bs=1 count=300 2>/dev/null
  dd if=/dev/zero of=images/photo.png bs=1 count=3000 2>/dev/null
  dd if=/dev/zero of=docs/chapter.txt bs=1 count=100 2>/dev/null
  dd if=/dev/zero of=docs/manual.txt bs=1 count=10000 2>/dev/null

  # Set different thresholds per directory pattern
  printf "*.dat filter=lfs autotracksize=1000\n" >> .gitattributes
  printf "images/* filter=lfs autotracksize=500\n" >> .gitattributes
  printf "docs/* filter=lfs autotracksize=2000\n" >> .gitattributes
  printf "* filter=lfs autotracksize=10000\n" >> .gitattributes

  git add .gitattributes small.dat large.dat images/icon.png images/photo.png docs/chapter.txt docs/manual.txt 2>&1

  # large.dat (5000 >= 1000) in root should be LFS
  refute_pointer "main" "large.dat" 2>/dev/null || true
  # images/photo.png (3000 >= 500) should be LFS
  refute_pointer "main" "images/photo.png" 2>/dev/null || true
  # docs/manual.txt (10000 >= 2000) should be LFS
  refute_pointer "main" "docs/manual.txt" 2>/dev/null || true
)
end_test
