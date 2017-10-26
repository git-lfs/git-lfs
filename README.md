# Git Large File Storage

| Linux | macOS | Windows |
| :---- | :------ | :---- |
[ ![Linux build status][1]][2] | [![macOS build status][3]][4] | [![Windows build status][5]][6] |

[1]: https://travis-ci.org/git-lfs/git-lfs.svg?branch=master
[2]: https://travis-ci.org/git-lfs/git-lfs
[3]: https://circleci.com/gh/git-lfs/git-lfs.svg?style=shield&circle-token=856152c2b02bfd236f54d21e1f581f3e4ebf47ad
[4]: https://circleci.com/gh/git-lfs/git-lfs
[5]: https://ci.appveyor.com/api/projects/status/46a5yoqc3hk59bl5/branch/master?svg=true
[6]: https://ci.appveyor.com/project/git-lfs/git-lfs/branch/master

Git LFS is a command line extension and [specification](docs/spec.md) for
managing large files with Git. The client is written in Go, with pre-compiled
binaries available for Mac, Windows, Linux, and FreeBSD. Check out the
[Git LFS website][page] for an overview of features.

[page]: https://git-lfs.github.com/

## Getting Started

By default, the Git LFS client needs a Git LFS server to sync the large files
it manages. This works out of the box when using popular git repository
hosting providers like GitHub, Atlassian, etc. When you host your own
vanilla git server, for example, you need to either use a separate
[Git LFS server instance](https://github.com/git-lfs/git-lfs/wiki/Implementations),
or use the [custom transfer adapter](docs/custom-transfers.md) with
a transfer agent in blind mode, without having to use a Git LFS server instance.

You can install the Git LFS client in several different ways, depending on
your setup and preferences.

* Linux users can install Debian or RPM packages from [PackageCloud](https://packagecloud.io/github/git-lfs/install).  See the [Installation Guide](./INSTALLING.md) for details.
* Mac users can install from [Homebrew](https://github.com/Homebrew/homebrew) with `brew install git-lfs`, or from [MacPorts](https://www.macports.org) with `port install git-lfs`.
* Windows users can install from [Chocolatey](https://chocolatey.org/) with `choco install git-lfs`.
* [Binary packages are available][rel] for Windows, Mac, Linux, and FreeBSD.
* You can build it with Go 1.8.1+. See the [Contributing Guide](./CONTRIBUTING.md) for instructions.

[rel]: https://github.com/git-lfs/git-lfs/releases

Note: Git LFS requires Git v1.8.5 or higher.

Once installed, you need to setup the global Git hooks for Git LFS. This only
needs to be done once per machine.

```bash
$ git lfs install
```

Now, it's time to add some large files to a repository. The first step is to
specify file patterns to store with Git LFS. These file patterns are stored in
`.gitattributes`.

```bash
$ mkdir large-repo
$ cd large-repo
$ git init

# Add all zip files through Git LFS
$ git lfs track "*.zip"
```

Now you're ready to push some commits:

```bash
$ git add .gitattributes
$ git add my.zip
$ git commit -m "add zip"
```

You can confirm that Git LFS is managing your zip file:

```bash
$ git lfs ls-files
my.zip
```

Once you've made your commits, push your files to the Git remote:

```bash
$ git push origin master
Sending my.zip
LFS: 12.58 MB / 12.58 MB  100.00 %
Counting objects: 2, done.
Delta compression using up to 8 threads.
Compressing objects: 100% (5/5), done.
Writing objects: 100% (5/5), 548 bytes | 0 bytes/s, done.
Total 5 (delta 1), reused 0 (delta 0)
To https://github.com/git-lfs/git-lfs-test
   67fcf6a..47b2002  master -> master
```

## Need Help?

You can get help on specific commands directly:

```bash
$ git lfs help <subcommand>
```

The [official documentation](docs) has command references and specifications for
the tool. You can ask questions in the [Git LFS chat room][chat], or [file a new
issue][ish]. Be sure to include details about the problem so we can
troubleshoot it.

1. Include the output of `git lfs env`, which shows how your Git environment
is setup.
2. Include `GIT_TRACE=1` in any bad Git commands to enable debug messages.
3. If the output includes a message like `Errors logged to /path/to/.git/lfs/objects/logs/*.log`,
throw the contents in the issue, or as a link to a Gist or paste site.

[chat]: https://gitter.im/git-lfs/git-lfs
[ish]: https://github.com/git-lfs/git-lfs/issues

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for info on working on Git LFS and
sending patches. Related projects are listed on the [Implementations wiki
page][impl]. You can also join [the project's chat room][chat].

[impl]: https://github.com/git-lfs/git-lfs/wiki/Implementations

### Using LFS from other Go code

At the moment git-lfs is only focussed on the stability of its command line
interface, and the [server APIs](docs/api/README.md). The contents of the
source packages is subject to change. We therefore currently discourage other
Go code from depending on the git-lfs packages directly; an API to be used by
external Go code may be provided in future.

## Core Team

These are the humans that form the Git LFS core team, which runs the project.

In alphabetical order:

| [@andyneff](https://github.com/andyneff) | [@rubyist](https://github.com/rubyist) | [@sinbad](https://github.com/sinbad) | [@technoweenie](https://github.com/technoweenie) | [@ttaylorr](https://github.com/ttaylorr) |
|---|---|---|---|---|
| [![](https://avatars1.githubusercontent.com/u/7596961?v=3&s=100)](https://github.com/andyneff) | [![](https://avatars1.githubusercontent.com/u/143?v=3&s=100)](https://github.com/rubyist) | [![](https://avatars1.githubusercontent.com/u/142735?v=3&s=100)](https://github.com/sinbad) | [![](https://avatars3.githubusercontent.com/u/21?v=3&s=100)](https://github.com/technoweenie) | [![](https://avatars3.githubusercontent.com/u/443245?v=3&s=100)](https://github.com/ttaylorr) |
