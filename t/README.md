# `t`

This directory contains one of the two types of tests that the Git LFS project
uses to protect against regression. The first, scattered in `*_test.go` files
throughout the repository are _unit tests_, and written in Go, designed to
uncover failures at the unit level.

The second kind--and the one contained in this directory--are _integration
tests_, which are designed to exercise Git LFS in an end-to-end fashion,
running the `git`, and `git-lfs` binaries, along with a mock Git server.

You can run all tests in this directory with any of the following:

```ShellSession
$ make
$ make test
$ make PROVE_EXTRA_ARGS=-j9 test
```

Or run a single test (for example, `t-checkout.sh`) by any of the following:

```ShellSession
$ make ./t-checkout.sh
$ make PROVE_EXTRA_ARGS=-v ./t-checkout.sh
$ ./t-checkout.sh
```

Alternatively, one can run a selection of tests (via explicitly listing them or
making use of the built-in shell globbing) by any of the following:

```ShellSession
$ make ./t-*.sh
$ make PROVE_EXTRA_ARGS=-j9 ./t-*.sh
$ ./t-*.sh
```

## Test File(s)

There are a few important kinds of files to know about in the `t` directory:

- `cmd/`: contains the source code of binaries that are useful during test
  time, like the mocked Git server, or the test counting binary. For more about
  the contents of this directory, see [test lifecycle](#test-lifecycle) below.

  The file `t/cmd/testutils.go` is automatically linked and included during the
  build process of each file in `cmd`.

- `fixtures/`: contains shell scripts that load fixture repositories useful for
  testing against.

- `t-*.sh`: file(s) containing zero or more tests, typically related to
  a similar topic (c.f,. `t/t-push.sh`, `t/t-pull.sh`, etc.)

- `testenv.sh`: loads environment variables useful during tests. This file is
  sourced by `testlib.sh`.

- `testhelpers.sh`: loads shell functions useful during tests, like
  `setup_remote_repo`, and `clone_repo`.

- `testlib.sh`: loads the `begin_test`, `end_test`, and similar functions
  useful for instantiating a particular test.

## Test Lifecycle

When a test is run, the following occurs, in order:

1. Missing test binaries are compiled into the `bin` directory in the
   repository root. Note: this does _not_ include the `git-lfs` binary, which
   is re-compiled via `script/boostrap`.

2. An integration server is started by either (1) the `Makefile` or (2) the
   `cmd/lfstest-count-test.go` program, which keeps track of the number of
   running tests and starts an integration server any time the number of active
   tests goes from `0` to `1`, and stops the server when it goes from `n` to
   `0`.

3. After sourcing `t/testlib.sh` (& loading `t/testenv.sh`), each test is run
   in sequence per file. (In other words, multiple test files can be run in
   parallel, but the tests in a single file are run in sequence.)

4. An individual test will finish, and (if running under `prove`) another will
   be started in its place. Once all tests are done, `t/test_count` will go to
   `0`, and the test server will be torn down.

## Test Environment

There are a few environment variables that you can set to change the test suite
behavior:

* `GIT_LFS_TEST_DIR=path` - This sets the directory that is used as the current
working directory of the tests. By default, this will be in your temp dir. It's
recommended that this is set to a directory outside of any Git repository.

* `KEEPTRASH=1` - This will leave the local repository data in a `tmp` directory
and the remote repository data in `test/remote`.

Also ensure that your `noproxy` environment variable contains `127.0.0.1` host,
to allow git commands to reach the local Git server `lfstest-gitserver`.

## Writing new tests

A new test file should be named `t/t-*.sh`, where `*` is the topic of Git LFS
being tested. It should look as follows:

```bash
#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "my test"
(
  set -e

  # ...
)
end_test
```
