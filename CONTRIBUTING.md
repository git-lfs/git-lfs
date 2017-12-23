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
