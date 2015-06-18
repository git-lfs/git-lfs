# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## v1.0

These are the features that we feel are important for a v1 release of Git LFS,
and we have a good idea how they could work.

* Fast, efficient uploading and downloading.
* `git lfs fetch` command for downloading large files.
* Automatic GC for the `.git/lfs/objects` directory.
* Client side metrics reporting, so the Git LFS server can optionally track
how clients are performing.
* Ability to remove objects from the command line through the API.

## Possible Features

These are features that require some more research. It's very possible that
these can make it in for v1.0 if there's a great proposal.

* Narrow clones - Allow clients to specify which large files to download
automatically.
* File locking
* Binary diffing - reduce the amount of content sent over the wire.

## Project Related

These are items that don't affect Git LFS end users.

* Releases through common package repositories: RPM, Apt, Chocolatey, Homebrew.
* CI builds for Windows.
* Automated build servers that build Git LFS on native platforms.
* Automated QA test suite for running release candidates through a gauntlet of
open source and proprietary Git LFS environments.
* Automatic updates of the Git LFS client.
