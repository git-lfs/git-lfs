# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## v1.0

These are the features that we feel are important for a v1 release of Git LFS,
and we have a good idea how they could work.

* Fast, efficient uploading and downloading ([#414](https://github.com/github/git-lfs/issues/414)).
* Improved local storage management ([#490](https://github.com/github/git-lfs/issues/490)).
* Ability to remove objects from the command line through the API.
* [Extensions](docs/proposals/extensions.md).
* Official packages for CentOS, Apt.
* Go 1.5+

## Possible Features

These are features that require some more research. It's very possible that
these can make it in for v1.0 if there's a great proposal.

* File locking
* Binary diffing - reduce the amount of content sent over the wire.
* Client side metrics reporting, so the Git LFS server can optionally track
how clients are performing.

## Project Related

These are items that don't affect Git LFS end users.

* Releases through common package repositories: RPM, Apt, Chocolatey, Homebrew.
* CI builds for Windows.
* Automated build servers that build Git LFS on native platforms.
* Automated QA test suite for running release candidates through a gauntlet of
open source and proprietary Git LFS environments.
* Automatic updates of the Git LFS client.
