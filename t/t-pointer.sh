#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "pointer --file --stdin"
(
  set -e

  echo "simple" > some-file

  input="version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7"

  expected="Git LFS pointer for some-file

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee

Pointer from STDIN

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee"

  [ "$expected" = "$(echo "$input" | git lfs pointer --file=some-file --stdin 2>&1)" ]
)
end_test

begin_test "pointer --file --stdin mismatch"
(
  set -e

  echo "simple" > some-file

  input="version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 123"

  set +e
  output=$(echo "$input" | git lfs pointer --file=some-file --stdin 2>&1)
  status=$?
  set -e

  [ "1" = "$status" ]

  expected="Git LFS pointer for some-file

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee

Pointer from STDIN

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 123

Git blob OID: 905bcc24b5dc074ab870f9944178e398eec3b470

Pointers do not match"

  [ "$expected" = "$output" ]
)
end_test

begin_test "pointer --stdin"
(
  set -e

  echo "version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7" > valid-pointer

  output=$(cat valid-pointer | git lfs pointer --stdin 2>&1)
  expected="Pointer from STDIN

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7"

  [ "$expected" = "$output" ]
)
end_test

begin_test "pointer --stdin without stdin"
(
  # this test doesn't work on Windows, it just operates like 'bad pointer' case
  # stdin isn't detectable as detached, it just times out with no content
  if [[ "$(is_stdin_attached)" == "0" ]]; then
    echo "Skipping pointer without stdin because STDIN attached"
    exit 0
  fi
  output=$(echo "" | git lfs pointer --stdin 2>&1)
  status=$?

  set -e

  expected="Cannot read from STDIN. The --stdin flag expects a pointer file from STDIN."

  [ "$expected" = "$output" ]

  [ "1" = "$status" ]
)
end_test

begin_test "pointer --stdin with bad pointer"
(
  output=$(echo "not a pointer" | git lfs pointer --stdin 2>&1)
  status=$?

  set -e

  expected="Pointer from STDIN

Pointer file error: invalid header"

  diff -u <(printf "%s" "$expected") <(printf "%s" "$output")

  [ "1" = "$status" ]
)
end_test

begin_test "pointer --file --pointer mismatch"
(
  set -e
  echo "simple" > some-file
  echo "version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 123" > invalid-pointer

  expected="Git LFS pointer for some-file

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee

Pointer from invalid-pointer

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 123

Git blob OID: 905bcc24b5dc074ab870f9944178e398eec3b470

Pointers do not match"

  set +e
  output=$(git lfs pointer --file=some-file --pointer=invalid-pointer 2>&1)
  status=$?
  set -e

  [ "1" = "$status" ]

  [ "$expected" = "$output" ]
)
end_test

begin_test "pointer --file --pointer"
(
  set -e
  echo "simple" > some-file
  echo "version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7" > valid-pointer

  expected="Git LFS pointer for some-file

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee

Pointer from valid-pointer

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7

Git blob OID: e18acd45d7e3ce0451d1d637f9697aa508e07dee"


  [ "$expected" = "$(git lfs pointer --file=some-file --pointer=valid-pointer 2>&1)" ]
)
end_test

begin_test "pointer --pointer"
(
  set -e

  echo "version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7" > valid-pointer

  expected="Pointer from valid-pointer

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7"

  [ "$expected" = "$(git lfs pointer --pointer=valid-pointer 2>&1)" ]
)
end_test

begin_test "pointer missing --pointer"
(
  output=$(git lfs pointer --pointer=missing-pointer 2>&1)
  status=$?
  set -e

  [ "1" = "$status" ]

  echo "$output"
  echo "$output" | grep "open missing-pointer:"
)
end_test

begin_test "pointer invalid --pointer"
(
  set -e

  echo "not a pointer" > some-pointer

  set +e
  output=$(git lfs pointer --pointer=some-pointer 2>&1)
  status=$?
  set -e

  [ "1" = "$status" ]

  expected="Pointer from some-pointer

Pointer file error: invalid header"

  diff -u <(printf "%s" "$expected") <(printf "%s" "$output")

  [ "$expected" = "$output" ]
)
end_test

begin_test "pointer --file"
(
  set -e
  echo "simple" > some-file

  expected="Git LFS pointer for some-file

version https://git-lfs.github.com/spec/v1
oid sha256:6c17f2007cbe934aee6e309b28b2dba3c119c5dff2ef813ed124699efe319868
size 7"

  [ "$expected" = "$(git lfs pointer --file=some-file 2>&1)" ]
)
end_test

begin_test "pointer without args"
(
  output=$(git lfs pointer 2>&1)
  status=$?
  set -e

  [ "Nothing to do!" = "$output" ]
  [ "1" = "$status" ]
)
end_test

begin_test "pointer stdout/stderr"
(
  set -e
  echo "pointer-stdout-test" > pointer-stdout-test.txt
  git lfs pointer --file=pointer-stdout-test.txt > stdout.txt 2> stderr.txt
  echo "stdout:"
  cat stdout.txt
  [ $(wc -l stdout.txt | sed -e 's/^[[:space:]]*//' | cut -f1 -d' ') -eq 3 ]
  grep "oid sha256:e96ec1bd71eea8df78b24c64a7ab9d42dd7f821c4e503f0e2288273b9bff6c16" stdout.txt
  [ $(grep -c "Git LFS pointer" stdout.txt) -eq 0 ]

  echo "stderr:"
  cat stderr.txt
  grep "Git LFS pointer" stderr.txt
  [ $(grep -c "oid sha256:" stderr.txt) -eq 0 ]
)
end_test

begin_test "pointer to console"
(
  set -e
  echo "pointer-stdout-test" > pointer-stdout-test.txt
  git lfs pointer --file=pointer-stdout-test.txt 2>&1 | tee pointer.txt
  grep "Git LFS pointer" pointer.txt
  grep "oid sha256:e96ec1bd71eea8df78b24c64a7ab9d42dd7f821c4e503f0e2288273b9bff6c16" pointer.txt
)
end_test

begin_test "pointer --check (with valid pointer)"
(
  set -e

  reponame="pointer---check-valid-pointer"
  git init "$reponame"
  cd "$reponame"

  echo "contents" > good.txt
  git lfs pointer --file good.txt > good.ptr

  cat good.ptr

  git lfs pointer --check --file good.ptr
  git lfs pointer --check --stdin < good.ptr
  git lfs pointer --check --no-strict --file good.ptr
  git lfs pointer --check --no-strict --stdin < good.ptr
  git lfs pointer --check --strict --file good.ptr
  git lfs pointer --check --strict --stdin < good.ptr
)
end_test

begin_test "pointer --check (with invalid pointer)"
(
  set -e

  reponame="pointer---check-invalid-pointer"
  git init "$reponame"
  cd "$reponame"

  echo "not-a-pointer" > bad.ptr

  git lfs pointer --check --file bad.ptr && exit 1
  git lfs pointer --check --stdin < bad.ptr && exit 1
  git lfs pointer --check --no-strict --file bad.ptr && exit 1
  git lfs pointer --check --no-strict --stdin < bad.ptr && exit 1
  git lfs pointer --check --strict --file bad.ptr && exit 1
  git lfs pointer --check --strict --stdin < bad.ptr && exit 1
  # Make the result of the subshell a success.
  true
)
end_test

begin_test "pointer --check (with empty file)"
(
  set -e

  reponame="pointer---check-empty-file"
  git init "$reponame"
  cd "$reponame"

  touch empty.ptr

  git lfs pointer --check --file empty.ptr
  git lfs pointer --check --stdin < empty.ptr
  git lfs pointer --check --no-strict --file empty.ptr
  git lfs pointer --check --no-strict --stdin < empty.ptr
  git lfs pointer --check --strict --file empty.ptr
  git lfs pointer --check --strict --stdin < empty.ptr
)
end_test

begin_test "pointer --check (with size 0 pointer)"
(
  set -e

  reponame="pointer---check-size-0"
  git init "$reponame"
  cd "$reponame"

  printf '%s\n' \
      'version https://git-lfs.github.com/spec/v1' \
      'oid sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855' \
      'size 0' \
   >zero.ptr

  git lfs pointer --check --file zero.ptr
  git lfs pointer --check --stdin < zero.ptr
  git lfs pointer --check --no-strict --file zero.ptr
  git lfs pointer --check --no-strict --stdin < zero.ptr
  git lfs pointer --check --strict --file zero.ptr && exit 1
  git lfs pointer --check --strict --stdin < zero.ptr && exit 1
  # Make the result of the subshell a success.
  true
)
end_test

begin_test "pointer --check (with CRLF endings)"
(
  set -e

  reponame="pointer---check-crlf"
  git init "$reponame"
  cd "$reponame"

  printf '%s\r\n' \
    'version https://git-lfs.github.com/spec/v1' \
    'oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393' \
    'size 12345' \
    >crlf.ptr

  git lfs pointer --check --file crlf.ptr
  git lfs pointer --check --stdin < crlf.ptr
  git lfs pointer --check --no-strict --file crlf.ptr
  git lfs pointer --check --no-strict --stdin < crlf.ptr
  git lfs pointer --check --strict --file crlf.ptr && exit 1
  git lfs pointer --check --strict --stdin < crlf.ptr && exit 1
  # Make the result of the subshell a success.
  true
)
end_test

begin_test "pointer --check (with invalid arguments)"
(
  set -e

  reponame="pointer---check-invalid-pointer"
  git init "$reponame"
  cd "$reponame"

  touch a.txt

  # git-lfs-pointer(1) --check with invalid combination --compare
  git lfs pointer --check --compare && exit 1

  # git-lfs-pointer(1) --check without --file or --stdin
  git lfs pointer --check && exit 1

  # git-lfs-pointer(1) --check with --file and --stdin
  git lfs pointer --check --file a.txt --stdin && exit 1

  # Make the result of the subshell a success.
  true
)
end_test
