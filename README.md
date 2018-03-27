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

[Git LFS](https://git-lfs.github.com) is a command line extension and
[specification](docs/spec.md) for managing large files with Git.

The client is written in Go, with pre-compiled binaries available for Mac,
Windows, Linux, and FreeBSD. Check out the [website](http://git-lfs.github.com)
for an overview of features.

## Getting Started

### Installation

You can install the Git LFS client in several different ways, depending on your
setup and preferences.

* **Linux users**. Debian and RPM packages are available from
  [PackageCloud](https://packagecloud.io/git-lfs/install).
* **macOS users**. [Homebrew](https://brew.sh) bottles are distributed, and can
  be installed via `brew install git-lfs`.
* **Windows users**. Chocolatey packages are distributed, and can be installed
  via `choco install git-lfs`.

In addition, [binary packages](https://github.com/git-lfs/git-lfs/releases) are
available for Linux, macOS, Windows, and FreeBSD. This repository can also be
built-from-source using the latest version of [Go](https://golang.org).

### Usage

Git LFS requires a global installation once per-machine. This can be done by
running:

```bash
$ git lfs install
```

To begin using Git LFS within your Git repository, you can indicate which files
you would like Git LFS to manage. This can be done by running the following
_from within Git repository_:

```bash
$ git lfs track "*.psd"
```

(Where `*.psd` is the pattern of filenames that you wish to track. You can read
more about this pattern syntax
[here](https://git-scm.org/docs/git-attribute://git-scm.com/docs/gitattributes)).

After any invocation of `git-lfs-track(1)` or `git-lfs-untrack(1)`, you _must
commit changes to your `.gitattributes` file_. This can be done by running:

```bash
$ git add .gitattributes
$ git commit -m "track *.psd files using Git LFS"
```

You can now interact with your Git repository as usual, and Git LFS will take
care of managing your large files. For example, changing a file named `my.psd`
(tracked above via `*.psd`):

```bash
$ git add my.psd
$ git commit -m "add psd"
```

You can confirm that Git LFS is managing your PSD file:

```bash
$ git lfs ls-files
3c2f7aedfb * my.psd
```

Once you've made your commits, push your files to the Git remote:

```bash
$ git push origin master
Uploading LFS objects: 100% (1/1), 810 B, 1.2 KB/s
# ...
To https://github.com/git-lfs/git-lfs-test
   67fcf6a..47b2002  master -> master
```

Note: Git LFS requires Git v1.8.5 or higher.

## Limitations

Git LFS maintains a list of currently known limitations, which you can find and
edit [here](https://github.com/git-lfs/git-lfs/wiki/Limitations).

## Need Help?

You can get help on specific commands directly:

```bash
$ git lfs help <subcommand>
```

The [official documentation](docs) has command references and specifications for
the tool.

You can always [open an issue](https://github.com/git-lfs/git-lfs/issues), and
one of the Core Team members will respond to you. Please be sure to include:

1. The output of `git lfs env`, which displays helpful information about your
   Git repository useful in debugging.
2. Any failed commands re-run with `GIT_TRACE=1` in the environment, which
   displays additional information pertaining to why a command crashed.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for info on working on Git LFS and
sending patches. Related projects are listed on the [Implementations wiki
page](https://github.com/git-lfs/git-lfs/wiki/Implementations).

## Core Team

These are the humans that form the Git LFS core team, which runs the project.

In alphabetical order:

| [@andyneff](https://github.com/andyneff) | [@rubyist](https://github.com/rubyist) | [@sinbad](https://github.com/sinbad) | [@technoweenie](https://github.com/technoweenie) | [@ttaylorr](https://github.com/ttaylorr) |
|---|---|---|---|---|
| [![](https://avatars1.githubusercontent.com/u/7596961?v=3&s=100)](https://github.com/andyneff) | [![](https://avatars1.githubusercontent.com/u/143?v=3&s=100)](https://github.com/rubyist) | [![](https://avatars1.githubusercontent.com/u/142735?v=3&s=100)](https://github.com/sinbad) | [![](https://avatars3.githubusercontent.com/u/21?v=3&s=100)](https://github.com/technoweenie) | [![](https://avatars3.githubusercontent.com/u/443245?v=3&s=100)](https://github.com/ttaylorr) |
