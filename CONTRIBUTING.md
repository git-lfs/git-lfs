## Contributing to Git Large File Storage

Hi there! We're thrilled that you'd like to contribute to this project. Your
help is essential for keeping it great.

This project adheres to the [Open Code of Conduct][code-of-conduct]. By participating, you are expected to uphold this code.
[code-of-conduct]: http://todogroup.org/opencodeofconduct/#Git LFS/opensource@github.com

## Issue Labels

The Git LFS teams mark issues and pull requests with the following labels:

* `bug` - An issue describing a bug.
* `enhancement` - An issue for a possible new feature.
* `review` - An issue ready to be reviewed.
* `release` - A checklist issue showing items marked for an upcoming release.
* `roadmap` - A checklist issue with tasks to fulfill something from the
[roadmap](./ROADMAP.md)
* `storage` - Used internally by the core contributors from GitHub. It just
means we're paying extra attention to it.

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

## Submitting a pull request

0. [Fork][] and clone the repository
0. Configure and install the dependencies: `script/bootstrap`
0. Make sure the tests pass on your machine: `script/test`
0. Create a new branch: `git checkout -b my-branch-name`
0. Make your change, add tests, and make sure the tests still pass
0. Push to your fork and [submit a pull request][pr]
0. Accept the [GitHub CLA][cla]
0. Pat your self on the back and wait for your pull request to be reviewed.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Follow the [style guide][style] where possible.
- Write tests.
- Update documentation as necessary.  Commands have [man pages](./docs/man).
- Keep your change as focused as possible. If there are multiple changes you
would like to make that are not dependent upon each other, consider submitting
them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Updating 3rd party packages

0. Update `Nut.toml`.
0. Run `script/vendor` to update the code in the `.vendor/src` directory.
0. Commit the change.  Git LFS vendors the full source code in the repository.
0. Submit a pull request.

## Releasing

If you are the current maintainer:

* Create a [new draft Release](https://github.com/github/git-lfs/releases/new).
List any changes with links to related PRs.
* Make sure your local dependencies are up to date: `script/bootstrap`
* Ensure that tests are green: `script/cibuild`
* Bump the version in `lfs/lfs.go`, [like this](https://github.com/github/git-lfs/commit/dd17828e4a6f2394cbba8621037199dc28f046e8).
* Add the new version to the top of CHANGELOG.md
* Build for all platforms with `script/bootstrap -all` (you need Go setup for
cross compiling with Mac, Linux, FreeBSD, and Windows support).
* Test the command locally.  The compiled version will be in `bin/releases/{os}-{arch}/git-lfs-{version}/git-lfs`
* Get the draft Release ID from the GitHub API: `curl -in https://api.github.com/repos/github/git-lfs/releases`
* Run `script/release -id {id}` to upload all of the compiled binaries to the
release.
* Publish the Release on GitHub.
* Update [Git LFS website](https://github.com/github/git-lfs.github.com/blob/gh-pages/_config.yml#L4).
* Ping external teams on GitHub:
  * @github/desktop
* Build packages:
  * rpm
  * apt

## Resources

- [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/)
- [Using Pull Requests](https://help.github.com/articles/using-pull-requests/)
- [GitHub Help](https://help.github.com)

[fork]: https://github.com/github/git-lfs/fork
[pr]: https://github.com/github/git-lfs/compare
[style]: https://github.com/golang/go/wiki/CodeReviewComments
[cla]: https://cla.github.com/github/git-lfs/accept
