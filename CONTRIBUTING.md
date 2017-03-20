## Contributing to Git Large File Storage

Hi there! We're thrilled that you'd like to contribute to this project. Your
help is essential for keeping it great.

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

You can see what the Git LFS team is prioritizing work on in the
[roadmap](./ROADMAP.md).

## Project Management

The Git LFS project is managed completely through this open source project and
its [chat room][chat]. The [roadmap][] shows the high level items that are
prioritized for future work. Suggestions for major features should be submitted
as a pull request that adds a markdown file to `docs/proposals` discussing the
feature. This gives the community time to discuss it before a lot of code has
been written. Roadmap items are linked to one or more Issue task lists ([example][roadmap-items]), with the `roadmap` label, that go into more detail.

[chat]: https://gitter.im/git-lfs/git-lfs
[roadmap]: ./ROADMAP.md
[roadmap-items]: https://github.com/git-lfs/git-lfs/issues/490

The Git LFS teams mark issues and pull requests with the following labels:

* `bug` - An issue describing a bug.
* `core-team` - An issue relating to the governance of the project.
* `enhancement` - An issue for a possible new feature.
* `review` - A pull request ready to be reviewed.
* `release` - A checklist issue showing items marked for an upcoming release.
* `roadmap` - A checklist issue with tasks to fulfill something from the
[roadmap](./ROADMAP.md)

## Branching strategy

In general, contributors should develop on branches based off of `master` and pull requests should be to `master`.

## Submitting a pull request

0. [Fork][] and clone the repository
0. Configure and install the dependencies: `script/bootstrap`
0. Make sure the tests pass on your machine: `script/test`
0. Create a new branch based on `master`: `git checkout -b <my-branch-name> master`
0. Make your change, add tests, and make sure the tests still pass
0. Push to your fork and [submit a pull request][pr] from your branch to `master`
0. Accept the [GitHub CLA][cla]
0. Pat yourself on the back and wait for your pull request to be reviewed

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

* Follow the [style guide][style] where possible.
* Write tests.
* Update documentation as necessary.  Commands have [man pages](./docs/man).
* Keep your change as focused as possible. If there are multiple changes you
would like to make that are not dependent upon each other, consider submitting
them as separate pull requests.
* Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Building

### Prerequisites

Git LFS depends on having a working Go 1.7.3+ environment, with your standard
`$GOROOT` and `$GOPATH` environment variables set.

On RHEL etc. e.g. Red Hat Enterprise Linux Server release 7.2 (Maipo), you will neet the minimum packages installed to build Git LFS:

```
$ sudo yum install gcc
$ sudo yum install perl-Digest-SHA
```

In order to run the RPM build `rpm/build_rpms.bsh` you will also need to:

`$ sudo yum install ruby-devel`

(note on an AWS instance you may first need to `sudo yum-config-manager --enable rhui-REGION-rhel-server-optional`)

### Building Git LFS

The easiest way to download Git LFS for making changes is `go get`:

    $ go get github.com/git-lfs/git-lfs

This clones the Git LFS repository to your `$GOPATH`. If you typically keep
your projects in a specific directory, you can symlink it from `$GOPATH`:

    $ cd ~/path/to/your/projects
    $ ln -s $GOPATH/src/github.com/git-lfs/git-lfs

From here, run `script/bootstrap` to build Git LFS in the `./bin` directory.
Before submitting changes, be sure to run the Go tests and the shell integration
tests:

    $ script/test        # runs just the Go tests
    $ script/integration # runs the shell tests in ./test
    $ script/cibuild     # runs everything, with verbose debug output

## Updating 3rd party packages

0. Update `glide.yaml`.
0. Run `script/vendor` to update the code in the `vendor` directory.
0. Commit the change.  Git LFS vendors the full source code in the repository.
0. Submit a pull request.

## Releasing

If you are the current maintainer:

* Create a [new draft Release](https://github.com/git-lfs/git-lfs/releases/new).
List any changes with links to related PRs.
* Make sure your local dependencies are up to date: `script/bootstrap`
* Ensure that tests are green: `script/cibuild`
* Bump the version in `lfs/lfs.go`, [like this](https://github.com/git-lfs/git-lfs/commit/dd17828e4a6f2394cbba8621037199dc28f046e8).
* Add the new version to the top of CHANGELOG.md
* Build for all platforms with `script/bootstrap -all` (you need Go setup for
cross compiling with Mac, Linux, FreeBSD, and Windows support).
* Test the command locally.  The compiled version will be in `bin/releases/{os}-{arch}/git-lfs-{version}/git-lfs`
* Get the draft Release ID from the GitHub API: `curl -in https://api.github.com/repos/git-lfs/git-lfs/releases`
* Run `script/release -id {id}` to upload all of the compiled binaries to the
release.
* Publish the Release on GitHub.
* Update [Git LFS website](https://github.com/git-lfs/git-lfs.github.com/blob/gh-pages/_config.yml#L4)
(release engineer access rights required).
* Ping external teams on GitHub:
  * @github/desktop
* Build packages:
  * rpm
  * apt
* Bump homebrew version and generate the homebrew hash with `curl --location https://github.com/git-lfs/git-lfs/archive/vx.y.z.tar.gz | shasum -a 256` ([example](https://github.com/Homebrew/homebrew-core/pull/413/commits/dc0eb1f62514f48f3f5a8d01ad3bea06f78bd566))
* Create release branch for bug fixes, such as `release-1.5`.
* Increment version in `config/version.go` to the next expected version. If
v1.5 just shipped, set the version in master to `1.6-pre`, for example.

## Resources

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [Using Pull Requests](https://help.github.com/articles/using-pull-requests/)
- [GitHub Help](https://help.github.com)

[fork]: https://github.com/git-lfs/git-lfs/fork
[pr]: https://github.com/git-lfs/git-lfs/compare
[style]: https://github.com/golang/go/wiki/CodeReviewComments
[cla]: https://cla.github.com/git-lfs/git-lfs/accept
