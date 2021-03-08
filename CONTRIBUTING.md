## Contributing to Git Large File Storage

Hi there! We're thrilled that you'd like to contribute to this project. Your
help is essential for making it the best it can be.

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) to the public under the [project's open source license](LICENSE.md).

This project adheres to the [Open Code of Conduct](./CODE-OF-CONDUCT.md). By participating, you are expected to uphold this code.

## Feature Requests

Feature requests are welcome, but will have a much better chance of being
accepted if they meet the first principles for the project. Git LFS is intended
for end users, not Git experts. It should fit into the standard workflow as
much as possible, and require little client configuration.

* Large objects are pushed to Git LFS servers during git push.
* Large objects are downloaded during git checkout.
* Git LFS servers are linked to Git remotes by default. Git hosts can support
users without requiring them to set up anything extra. Users can access
different Git LFS servers like they can with different Git remotes.
* Upload and download requests should use the same form of authentication built
into Git: SSH through public keys, and HTTPS through Git credential helpers.
* Git LFS servers use a JSON API designed around progressive enhancement.
Servers can simply host off cloud storage, or implement more efficient methods
of transferring data.

Since the focus for the project is on end users, we're generally hesitant about
introducing new features that make data loss easy or are prone to misuse.
However, we're not necessarily opposed to adding generally applicable
customizability or features for advanced users if they don't conflict with other
project goals.

## Project Management

The Git LFS project is managed completely through this open source project. The
[milestones][] show the high level items that are prioritized for future work.
Suggestions for major features should be submitted as a pull request that adds a
markdown file to `docs/proposals` discussing the feature. This gives the
community time to discuss it before a lot of code has been written.

[milestones]: https://github.com/git-lfs/git-lfs/milestones

The Git LFS teams mark issues and pull requests with the following labels:

* `bug` - An issue describing a bug.
* `enhancement` - An issue for a possible new feature.
* `review` - A pull request ready to be reviewed.
* `release` - A checklist issue showing items marked for an upcoming release.

## Branching strategy

In general, contributors should develop on branches based off of `main` and pull requests should be to `main`.

## Submitting a pull request

1. [Fork][] and clone the repository
1. Configure and install the dependencies: `make`
1. Make sure the tests pass on your machine: `make test`
1. Create a new branch based on `main`: `git checkout -b <my-branch-name> main`
1. Make your change, add tests, and make sure the tests still pass
1. Push to your fork and [submit a pull request][pr] from your branch to `main`
1. Pat yourself on the back and wait for your pull request to be reviewed

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

* Follow the [style guide][style] where possible.
* Write tests.
* Update documentation as necessary.  Commands have [man pages](./docs/man).
* Keep your change as focused as possible. If there are multiple changes you
would like to make that are not dependent upon each other, consider submitting
them as separate pull requests.
* Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).
* Explain the rationale for your change in the pull request. You can often use
  part of a good commit message as a starting point.

## Discussions

[Our discussions](https://github.com/git-lfs/git-lfs/discussions) are the
perfect place to ask a question if you're not sure on something, provide
feedback that isn't a bug report or feature request, or learn about use cases or
best practices with Git LFS.  There's even a search box to help you see if
someone has already answered your question!

You can also check [the FAQ](https://github.com/git-lfs/git-lfs/wiki/FAQ) to see
if your question is well known and already has an easy answer.

## Issues

If you think you've found a bug or have an issue, we'd love to hear about it!
Here are some tips for getting your question answered as quickly as possible:

* It's helpful if your issue includes the output of `git lfs env`, plus any
  relevant information about platform or configuration (e.g., container or CI
  usage, Cygwin, WSL, or non-Basic authentication).
* Take a look at the
  [troubleshooting](https://github.com/git-lfs/git-lfs/wiki/Troubleshooting) and
  [FAQ](https://github.com/git-lfs/git-lfs/wiki/FAQ) pages on the wiki. We
  update them from time to time with information on how to track down problems.
  If it seems relevant, include any information you've learned by following
  those steps.
* If you're having problems with GitHub's server-side LFS support, it's best to
  reach out to [GitHub's support team](https://github.com/contact) to get help.
  We aren't able to address GitHub-specific issues in this project, but the
  GitHub support team will do their best to help you out.
* If you see an old issue that's closed as fixed, but you're still experiencing
  the problem on your system, please open a new issue. The problem you're seeing
  is likely different, at least in the way it works internally, and we can help
  best when we have a new issue with all the information.

## Building

### Prerequisites

Git LFS depends on having a working Go development environment.  We officially
support the latest version of Go, although we try not to break backwards
compatibility with older versions if it's possible to avoid doing so.

On RHEL etc. e.g. Red Hat Enterprise Linux Server release 7.2 (Maipo), you will neet the minimum packages installed to build Git LFS:

```ShellSession
$ sudo yum install gcc
$ sudo yum install perl-Digest-SHA
```

In order to run the RPM build `rpm/build_rpms.bsh` you will also need to:

```ShellSession
$ sudo yum install ruby-devel
```

(note on an AWS instance you may first need to `sudo yum-config-manager --enable rhui-REGION-rhel-server-optional`)

### Building Git LFS

The easiest way to download Git LFS for making changes is `git clone`:

```ShellSession
$ git clone git@github.com:git-lfs/git-lfs.git
$ cd git-lfs
```

From here, run `make` to build Git LFS in the `./bin` directory. Before
submitting changes, be sure to run the Go tests and the shell integration
tests:

```ShellSession
$ make test          # runs just the Go tests
$ cd t && make test  # runs the shell tests in ./test
$ script/cibuild     # runs everything, with verbose debug output
```

## Updating 3rd party packages

1. Update `go.mod`.
1. Run `make vendor` to update the code in the `vendor` directory.
1. Commit the change.  Git LFS vendors the full source code in the repository.
1. Submit a pull request.

## Releasing

If you are the current maintainer, see
[the release howto](./docs/howto/release-git-lfs.md) for how to perform a release.

## Resources

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [Using Pull Requests](https://help.github.com/articles/using-pull-requests/)
- [GitHub Help](https://help.github.com)

[fork]: https://github.com/git-lfs/git-lfs/fork
[pr]: https://github.com/git-lfs/git-lfs/compare
[style]: https://github.com/golang/go/wiki/CodeReviewComments
