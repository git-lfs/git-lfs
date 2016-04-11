# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## Bugs/Features

* git index issues [#937](https://github.com/github/git-lfs/issues/937)
* `authenticated` property on urls [#960](https://github.com/github/git-lfs/issues/960)
* Use `expires_at` to quickly put objects in the queue to hit the API again to refresh tokens.
* Accept raw remote URLs as valid [#1085](https://github.com/github/git-lfs/issues/1085) (@epriestley) 
* use git proxy settings [#1125](https://github.com/github/git-lfs/issues/1125)
* Not following 301 redirect [#1129](https://github.com/github/git-lfs/issues/1129)
* add all lfs.* git config keys to git lfs env output

## Upcoming Features

* File locking [#666](https://github.com/github/git-lfs/pull/666)
* Resumable uploads and downloads [#414](https://github.com/github/git-lfs/issues/414)
* Wrapped versions of `git pull` & `git checkout` that optimize without filters
like `git lfs clone`
* Remove non-batch API route in client

## Possible Features

* Binary diffing - reduce the amount of content sent over the wire.
* Client side metrics reporting, so the Git LFS server can optionally track
how clients are performing.
* Pure SSH: full API & transfer support for SSH without redirect to HTTP

## Project Related

These are items that don't affect Git LFS end users.

* CI builds for Windows.
* Automated build servers that build Git LFS on native platforms.
* Automated QA test suite for running release candidates through a gauntlet of
open source and proprietary Git LFS environments.
* Automatic updates of the Git LFS client. [#531](https://github.com/github/git-lfs/issues/531)
