# Git LFS Tests

Git LFS uses two form of tests: unit tests for the internals written in Go, and
integration tests that run `git` and `git-lfs` in a real shell environment.
You can run them separately:

```
$ script/test        # Tests the Go packages.
$ script/integration # Tests the commands in shell scripts.
```

CI servers should always run both:

```
$ script/cibuild
```

## Internal Package Tests

The internal tests use Go's builtin [testing][t] package.

You can run individual tests by passing arguments to `script/test`:

```
# test a specific Go package
$ script/test lfs

# pass other `go test` arguments
$ script/test lfs -run TestSuccessStatus -v
github.com/rubyist/tracerx
github.com/git-lfs/git-lfs/git
github.com/technoweenie/assert
=== RUN TestSuccessStatus
--- PASS: TestSuccessStatus (0.00 seconds)
PASS
ok  	_/Users/rick/git-lfs/git-lfs/lfs	0.011s
```

[t]: http://golang.org/pkg/testing/

## Integration Tests

Git LFS integration tests are shell scripts that test the `git-lfs` command from
the shell.  Each test file can be run individually, or in parallel through
`script/integration`. Some tests will change the `pwd`, so it's important that
they run in separate OS processes.

```
$ test/test-happy-path.sh
compile git-lfs for test/test-happy-path.sh
LFSTEST_URL=/Users/rick/git-lfs/git-lfs/test/remote/url LFSTEST_DIR=/Users/rick/git-lfs/git-lfs/test/remote lfstest-gitserver
test: happy path ...                                               OK
```

1. The integration tests should not rely on global system or git config.
2. The tests should be cross platform (Linux, Mac, Windows).
3. Tests should bootstrap an isolated, clean environment.  See the Test Suite
section.
4. Successful test runs should have minimal output.
5. Failing test runs should dump enough information to diagnose the bug.  This
includes stdout, stderr, any log files, and even the OS environment.

There are a few environment variables that you can set to change the test suite
behavior:

* `GIT_LFS_TEST_DIR=path` - This sets the directory that is used as the current
working directory of the tests. By default, this will be in your temp dir. It's
recommended that this is set to a directory outside of any Git repository.
* `GIT_LFS_TEST_MAXPROCS=N` - This tells `script/integration` how many tests to
run in parallel.  Default: 4.
* `KEEPTRASH=1` - This will leave the local repository data in a `tmp` directory
and the remote repository data in `test/remote`.
* `SKIPCOMPILE=1` - This skips the Git LFS compilation step.  Speeds up the
tests when you're running the same test script multiple times without changing
any Go code.

Also ensure that your `noproxy` environment variable contains `127.0.0.1` host,
to allow git commands to reach the local Git server `lfstest-gitserver`.

### Test Suite

The `testenv.sh` script includes some global variables used in tests.  This
should be automatically included in every `test/test-*.sh` script and
`script/integration`.

`testhelpers.sh` defines some shell functions.  Most are only used in the test
scripts themselves.  `script/integration` uses the `setup()` and `shutdown()`
functions.

`testlib.sh` is a [fork of a lightweight shell testing lib][testlib] that is
used internally at GitHub.  Only the `test/test-*.sh` scripts should include
this.

Tests live in this `./test` directory, and must have a unique name like:
`test-{name}.sh`. All tests should start with a basic template.  See
`test/test-happy-path.sh` for an example.

```
#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "template"
(
  set -e

  echo "your tests go here"
)
end_test
```

The `set -e` command will bail on the test at the first command that returns a
non zero exit status. Use simple shell commands like `grep` as assertions.

The test suite has standard `setup` and `shutdown` functions that should be
run only once, before/after running the tests.  The `setup` and `shutdown`
functions are run by `script/integration` and also by individual test scripts
when they are executed directly. Setup does the following:

* Resets temporary test directories.
* Compiles git-lfs with the latest code changes.
* Compiles Go files in `test/cmd` to `bin`, and adds them the PATH.
* Spins up a test Git and Git LFS server so the entire push/pull flow can be
exercised.
* Sets up a git credential helper that always returns a set username and
password.

The test Git server writes a `test/remote/url` file when it's complete.  This
file is how individual test scripts detect if `script/integration` is being
run.  You can fake this by manually spinning up the Git server using the
`lfstest-gitserver` line that is output after Git LFS is compiled.

By default, temporary directories in `tmp` and the `test/remote` directory are
cleared after test runs. Send the "KEEPTRASH" if you want to keep these files
around for debugging failed tests.

[testlib]: https://gist3.github.com/rtomayko/3877539
