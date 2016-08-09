# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## Bugs/Features

| | Name | Ref |
| ------ | ---- | --- |
| | git index issues | [#937](https://github.com/github/git-lfs/issues/937) |
| :soon: | `authenticated` property on urls | [#960](https://github.com/github/git-lfs/issues/960) |
| :soon: | Add ref information to upload request | [#969](https://github.com/github/git-lfs/issues/969) |
| :soon: | Accept raw remote URLs as valid | [#1085](https://github.com/github/git-lfs/issues/1085) |
| :construction: | `config` refactoring | |
| :soon: | Socks proxy support | [#1424](https://github.com/github/git-lfs/issues/1424) |
| :no_entry_sign: | Not following 301 redirect | [#1129](https://github.com/github/git-lfs/issues/1129) |
| | add all lfs.\* git config keys to git lfs env output | |
| | Teach `git lfs update` how to update the clean/smudge filter values | [#1083](https://github.com/github/git-lfs/pull/1083) |
| | Support multiple git alternates | |
| | Investigate `git lfs checkout` hardlinking instead of copying files. | |
| | Investigate `--shared` and `--dissociate` options for `git clone` (similar to `--references`) | |
| | Investigate `GIT_SSH_COMMAND` | [#1142](https://github.com/github/git-lfs/issues/1142) | |
| | Teach `git lfs install` to use `git config --system` instead of `git config --global` by default | [#1177](https://github.com/github/git-lfs/pull/1177) |
| | Investigate `git -c lfs.url=... lfs clone` usage | |
| | Test that manpages are built and included | [#1149](https://github.com/github/git-lfs/pull/1149) |
| | Update CI to build from source outside of git repo | [#1156](https://github.com/github/git-lfs/issues/1156#issuecomment-211574343) |
| :soon: | Teach `git lfs track` and others to warn when `git lfs install` hasn't been run (or auto-install) | [#1167](https://github.com/github/git-lfs/issues/1167) |

## Upcoming Features

| | Name | Ref |
| ------ | ---- | --- |
| :construction: | File locking [#666](https://github.com/github/git-lfs/pull/666) | |
| :ship: | Resumable uploads and downloads [#414](https://github.com/github/git-lfs/issues/414) | |
| :construction: | Wrapped versions of `git pull` & `git checkout` that optimize without filters like `git lfs clone` | |
| | Remove non-batch API route in client | |

## Possible Features

| | Name | Ref |
| ------ | ---- | --- |
| | Binary diffing - reduce the amount of content sent over the wire. | |
| | Client side metrics reporting, so the Git LFS server can optionally track how clients are performing. | |
| | Pure SSH: full API & transfer support for SSH without redirect to HTTP | |

## Project Related

These are items that don't affect Git LFS end users.

| | Name | Ref |
| ------ | ---- | --- |
| | CI builds for Windows. | |
| | Automated build servers that build Git LFS on native platforms. | |
| | Automated QA test suite for running release candidates through a gauntlet of open source and proprietary Git LFS environments. | |
| | Automatic updates of the Git LFS client. | [#531](https://github.com/github/git-lfs/issues/531) |

## Legend

* :ship: - Completed
* :construction: - In Progress
* :soon: - Up next
* :no_entry_sign: - Blocked

## How this works

0. Roadmap items are listed within their category in order of priority.
0. Roadmap items are kept up-to-date with the above legend.
0. Roadmap items are pruned once a release of LFS has been published.
