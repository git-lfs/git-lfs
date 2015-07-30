#!/usr/bin/env bash

. "test/testlib.sh"


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
  output=$(git lfs pointer --stdin 2>&1)
  status=$?

  set -e

  expected="Cannot read from STDIN. The --stdin flag expects a pointer file from STDIN."

  [ "$expected" = "$output" ]

  [ "1" = "$status" ]
)

begin_test "pointer --stdin with bad pointer"
(
  output=$(echo "not a pointer" | git lfs pointer --stdin 2>&1)
  status=$?

  set -e

  expected="Pointer from STDIN

Not a valid Git LFS pointer file."

  [ "$expected" = "$output" ]

  [ "1" = "$status" ]
)

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
  [ "open missing-pointer: no such file or directory" = "$output" ]
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

Not a valid Git LFS pointer file."

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
