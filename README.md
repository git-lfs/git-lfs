# Git Large File Storage

[![CI status][ci_badge]][ci_url]

[ci_badge]: https://github.com/git-lfs/git-lfs/workflows/CI/badge.svg
[ci_url]: https://github.com/git-lfs/git-lfs/actions?query=workflow%3ACI

[Git LFS](https://git-lfs.github.com) is a command line extension and
[specification](docs/spec.md) for managing large files with Git.

The client is written in Go, with pre-compiled binaries available for Mac,
Windows, Linux, and FreeBSD. Check out the [website](http://git-lfs.github.com)
for an overview of features.

## Getting Started

### Downloading

You can install the Git LFS client in several different ways, depending on your
setup and preferences.

* **Linux users**. Debian and RPM packages are available from
  [PackageCloud](https://packagecloud.io/github/git-lfs/install).
* **macOS users**. [Homebrew](https://brew.sh) bottles are distributed, and can
  be installed via `brew install git-lfs`.
* **Windows users**. Git LFS is included in the distribution of
  [Git for Windows](https://gitforwindows.org/). Alternatively, you can
  install a recent version of Git LFS from the [Chocolatey](https://chocolatey.org/) package manager.
* **Binary packages**. In addition, [binary packages](https://github.com/git-lfs/git-lfs/releases) are
available for Linux, macOS, Windows, and FreeBSD.
* **Building from source**. [This repository](https://github.com/git-lfs/git-lfs.git) can also be
built from source using the latest version of [Go](https://golang.org), and the
available instructions in our
[Wiki](https://github.com/git-lfs/git-lfs/wiki/Installation#source).

### Installing

#### From binary

The [binary packages](https://github.com/git-lfs/git-lfs/releases) include a script which will:

- Install Git LFS binaries onto the system `$PATH`
- Run `git lfs install` to
perform required global configuration changes.

```ShellSession
$ ./install.sh
```

#### From source

- Place the `git-lfs` binary on your system’s executable `$PATH` or equivalent.
- Git LFS requires global configuration changes once per-machine. This can be done by
running:

```ShellSession
$ git lfs install
```

#### Verifying releases

Releases are signed with the OpenPGP key of one of the core team members.  To
get these keys, you can run the following command, which will print them to
standard output:

```ShellSession
$ curl -L https://api.github.com/repos/git-lfs/git-lfs/tarball/core-gpg-keys | tar -Ozxf -
```

Once you have the keys, you can download the `sha256sums.asc` file and verify
the file you want like so:

```ShellSession
$ gpg -d sha256sums.asc | grep git-lfs-linux-amd64-v2.10.0.tar.gz | shasum -a 256 -c
```

## Example Usage

To begin using Git LFS within a Git repository that is not already configured
for Git LFS, you can indicate which files you would like Git LFS to manage.
This can be done by running the following _from within a Git repository_:

```bash
$ git lfs track "*.psd"
```

(Where `*.psd` is the pattern of filenames that you wish to track. You can read
more about this pattern syntax
[here](https://git-scm.com/docs/gitattributes)).

> *Note:* the quotation marks surrounding the pattern are important to
> prevent the glob pattern from being expanded by the shell.

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

> _Tip:_ if you have large files already in your repository's history, `git lfs
> track` will _not_ track them retroactively. To migrate existing large files
> in your history to use Git LFS, use `git lfs migrate`. For example:
>
> ```
> $ git lfs migrate import --include="*.psd" --everything
> ```
>
> For more information, read [`git-lfs-migrate(1)`](https://github.com/git-lfs/git-lfs/blob/master/docs/man/git-lfs-migrate.1.ronn).

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

Note: Git LFS requires at least Git 1.8.2 on Linux or 1.8.5 on macOS.

## Limitations

Git LFS maintains a list of currently known limitations, which you can find and
edit [here](https://github.com/git-lfs/git-lfs/wiki/Limitations).

## Need Help?

You can get help on specific commands directly:

```bash
$ git lfs help <subcommand>
```

The [official documentation](docs) has command references and specifications for
the tool.  There's also a [FAQ](https://github.com/git-lfs/git-lfs/wiki/FAQ) on
the wiki which answers some common questions.

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

See also [SECURITY.md](SECURITY.md) for info on how to submit reports
of security vulnerabilities.

## Core Team

These are the humans that form the Git LFS core team, which runs the project.

In alphabetical order:

| [@bk2204][bk2204-user] | [@chrisd8088][chrisd8088-user] | [@larsxschneider][larsxschneider-user] |
| :---: | :---: | :---: |
| [![][bk2204-img]][bk2204-user] | [![][chrisd8088-img]][chrisd8088-user] | [![][larsxschneider-img]][larsxschneider-user] |
| [PGP 0223B187][bk2204-pgp] | [PGP 088335A9][chrisd8088-pgp] | [PGP A5795889][larsxschneider-pgp] |

[bk2204-img]: https://avatars1.githubusercontent.com/u/497054?s=100&v=4
[chrisd8088-img]: https://avatars1.githubusercontent.com/u/28857117?s=100&v=4
[larsxschneider-img]: https://avatars1.githubusercontent.com/u/477434?s=100&v=4
[bk2204-user]: https://github.com/bk2204
[chrisd8088-user]: https://github.com/chrisd8088
[larsxschneider-user]: https://github.com/larsxschneider
[bk2204-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x88ace9b29196305ba9947552f1ba225c0223b187
[chrisd8088-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x09b4bf756670b76d63515717506c7945088335a9
[larsxschneider-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0xaa3b3450295830d2de6db90caba67be5a5795889

### Alumni

These are the humans that have in the past formed the Git LFS core team, or
have otherwise contributed a significant amount to the project. Git LFS would
not be possible without them.

In alphabetical order:

| [@andyneff][andyneff-user] | [@PastelMobileSuit][PastelMobileSuit-user] | [@rubyist][rubyist-user] | [@sinbad][sinbad-user] | [@technoweenie][technoweenie-user] | [@ttaylorr][ttaylorr-user] |
| :---: | :---: | :---: | :---: | :---: | :---: |
| [![][andyneff-img]][andyneff-user] | [![][PastelMobileSuit-img]][PastelMobileSuit-user] | [![][rubyist-img]][rubyist-user] | [![][sinbad-img]][sinbad-user] | [![][technoweenie-img]][technoweenie-user] | [![][ttaylorr-img]][ttaylorr-user] |

[andyneff-img]: https://avatars1.githubusercontent.com/u/7596961?v=3&s=100
[PastelMobileSuit-img]: https://avatars2.githubusercontent.com/u/37254014?s=100&v=4
[rubyist-img]: https://avatars1.githubusercontent.com/u/143?v=3&s=100
[sinbad-img]: https://avatars1.githubusercontent.com/u/142735?v=3&s=100
[technoweenie-img]: https://avatars3.githubusercontent.com/u/21?v=3&s=100
[ttaylorr-img]: https://avatars2.githubusercontent.com/u/443245?s=100&v=4
[andyneff-user]: https://github.com/andyneff
[PastelMobileSuit-user]: https://github.com/PastelMobileSuit
[sinbad-user]: https://github.com/sinbad
[rubyist-user]: https://github.com/rubyist
[technoweenie-user]: https://github.com/technoweenie
[ttaylorr-user]: https://github.com/ttaylorr
