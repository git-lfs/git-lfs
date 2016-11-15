# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. If you have an idea for a new feature, open an issue for discussion.

## Bugs/Features

| | Name | Ref |
| ------ | ---- | --- |
| | git index issues | [#937](https://github.com/git-lfs/git-lfs/issues/937) |
| :soon: | Add ref information to upload request | [#969](https://github.com/git-lfs/git-lfs/issues/969) |
| :soon: | Socks proxy support | [#1424](https://github.com/git-lfs/git-lfs/issues/1424) |
| :no_entry_sign: | Not following 301 redirect | [#1129](https://github.com/git-lfs/git-lfs/issues/1129) |
| | add all lfs.\* git config keys to git lfs env output | |
| | credential output hidden while transferring files | [#387](https://github.com/git-lfs/git-lfs/pull/387) |
| | Support multiple git alternates | |
| | Investigate `git lfs checkout` hardlinking instead of copying files. | |
| | Investigate `--shared` and `--dissociate` options for `git clone` (similar to `--references`) | |
| | Investigate `GIT_SSH_COMMAND` | [#1142](https://github.com/git-lfs/git-lfs/issues/1142) | |
| | Teach `git lfs install` to use `git config --system` instead of `git config --global` by default | [#1177](https://github.com/git-lfs/git-lfs/pull/1177) |
| | Investigate `git -c lfs.url=... lfs clone` usage | |
| | Test that manpages are built and included | [#1149](https://github.com/git-lfs/git-lfs/pull/1149) |
| | Update CI to build from source outside of git repo | [#1156](https://github.com/git-lfs/git-lfs/issues/1156#issuecomment-211574343) |
| :soon: | Teach `git lfs track` and others to warn when `git lfs install` hasn't been run (or auto-install) | [#1167](https://github.com/git-lfs/git-lfs/issues/1167) |
| | Investigate hanging pushes/pulls when git credential helper is not set | [#197](https://github.com/git-lfs/git-lfs/issues/197) |
| | Support git ssh shorthands | [#278](https://github.com/git-lfs/git-lfs/issues/278) |
| | Support `GIT_CONFIG` | [#318](https://github.com/git-lfs/git-lfs/issues/318) |
| | Warn when Git version is unsupported | [#410](https://github.com/git-lfs/git-lfs/issues/410) |
| | Detect when credential cacher is not setup | [#523](https://github.com/git-lfs/git-lfs/issues/523) |
| | Fix error logging from `git clone` errors | [#513](https://github.com/git-lfs/git-lfs/issues/513) |
| | Investigate cherry picking issues | [#438](https://github.com/git-lfs/git-lfs/issues/438) |
| | dynamic blob size cutoff for pointers | [#524](https://github.com/git-lfs/git-lfs/issues/524) |
| | windows `--help` support | [#394](https://github.com/git-lfs/git-lfs/issues/394) |
| | Investigate git hook installs within git worktree | [#1385](https://github.com/git-lfs/git-lfs/issues/1385) |
| | Support ssh username in ssh config | [#754](https://github.com/git-lfs/git-lfs/issues/754) |
| | Investigate `autocrlf` for lfs objects | [#723](https://github.com/git-lfs/git-lfs/issues/723) |

## Upcoming Features

| | Name | Ref |
| ------ | ---- | --- |
| :construction: | File locking | [#666](https://github.com/git-lfs/git-lfs/pull/666) |
| :ship: | Resumable uploads and downloads | [#414](https://github.com/git-lfs/git-lfs/issues/414) |
| :construction: | Wrapped versions of `git pull` & `git checkout` that optimize without filters like `git lfs clone` | |
| | Remove non-batch API route in client | |

## Possible Features

| | Name | Ref |
| ------ | ---- | --- |
| | Support tracking files by size | [#282](https://github.com/git-lfs/git-lfs/issues/282)
| | Binary diffing - reduce the amount of content sent over the wire. | |
| | Client side metrics reporting, so the Git LFS server can optionally track how clients are performing. | |
| | Pure SSH: full API & transfer support for SSH without redirect to HTTP | |
| | Compression of files in `.git/lfs/objects` | [#260](https://github.com/git-lfs/git-lfs/issues/260) |
| | LFS Migration tool | [#326](https://github.com/git-lfs/git-lfs/issues/326) |
| | Automatic upgrades | [#531](https://github.com/gihtub/git-lfs/issues/531) |
| | Investigate `git add` hash caching | [#574](https://github.com/git-lfs/git-lfs/issues/574) |
| | `git lfs archive` command | [#1322](https://github.com/git-lfs/git-lfs/issues/1322) |
| | Support 507 http responses | [#1327](https://github.com/git-lfs/git-lfs/issues/1327) |
| | Investigate shared object directory | [#766](https://github.com/git-lfs/git-lfs/issues/766) |

## Project Related

These are items that don't affect Git LFS end users.

| | Name | Ref |
| ------ | ---- | --- |
| :ship: | CI builds for Windows. | [#1567](https://github.com/git-lfs/git-lfs/pull/1567) |
| | Automated build servers that build Git LFS on native platforms. | |
| | Automated QA test suite for running release candidates through a gauntlet of open source and proprietary Git LFS environments. | |
| | Automatic updates of the Git LFS client. | [#531](https://github.com/git-lfs/git-lfs/issues/531) |

## Legend

* :ship: - Completed
* :construction: - In Progress
* :soon: - Up next
* :no_entry_sign: - Blocked

## How this works

0. Roadmap items are listed within their category in order of priority.
0. Roadmap items are kept up-to-date with the above legend.
0. Roadmap items are pruned once a release of LFS has been published.
