# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## Upcoming Features

* File locking [#666](https://github.com/github/git-lfs/pull/666)
* Resumable uploads and downloads [#414](https://github.com/github/git-lfs/issues/414)

## Possible Features

* Binary diffing - reduce the amount of content sent over the wire.
* Client side metrics reporting, so the Git LFS server can optionally track
how clients are performing.

## Project Related

These are items that don't affect Git LFS end users.

* CI builds for Windows.
* Automated build servers that build Git LFS on native platforms.
* Automated QA test suite for running release candidates through a gauntlet of
open source and proprietary Git LFS environments.
* Automatic updates of the Git LFS client. [#531](https://github.com/github/git-lfs/issues/531)
