# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## Releases

* 0.5 - Initial release using the [original HTTP API](docs/api/http-v1-original.md)
* 0.6 - First release using the [batch HTTP API](docs/api/http-v1-batch.md),
        with a fallback to the original API.
* 0.7 - Drops support for the original API.

## v1.0

These are the features that we feel are important for a v1 release of Git LFS,
and we have a good idea how they could work.

* Fast, efficient uploading and downloading ([#414](https://github.com/github/git-lfs/issues/414)).
* Improved local storage management ([#490](https://github.com/github/git-lfs/issues/490)).
* [Extensions](docs/proposals/extensions.md) (#486).
* Improved installation/upgrade experience (#531).
* Ability to remove objects from the command line through the API.
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
