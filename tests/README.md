# Git LFS integration tests

Git LFS integration tests use declarative statements to test the `git-lfs`
command from the outside.  Each test file is bootstrapped with a default
Git LFS environment, runs sequential commands, and checks the output.

The default environment should include the following:

* An empty Git repository in a random test directory, so test files can safely
run in parallel.
* An HTTPS endpoint for the Git repository.
* An SSH endpoint for the Git repository too.
* Git LFS compiled from the current ref.
* Git clean/smudge filters configured for the empty repository and the Git LFS
binary compiled for the test.
* A simple Git LFS server.

## Why not shell?

Using Go lets us run the test suite on multiple platforms without having to
worry about compatibility issues between environments.  We can also define
custom helpers, instead of being stuck with the lowest common denominator of
shell tools.  Though Go code will not be nearly as concise as shell, the explicit
Go syntax should lead to more descriptive assertions and fewer errors from those
less experienced with shell scripting.

The idea of a portable set of shell scripts that can be used to test Git LFS
compatibility between different client and server implementations is appealing,
so it's very possible the tests will end up as shell scripts eventually.
