# Git LFS Changelog

## 1.1.0 (18 November, 2015)

* NTLM auth support #820 (@WillHipschman, @technoweenie)
* Add `prune` command #742 (@sinbad)
* Use .lfsconfig instead of .gitconfig #837 (@technoweenie)
* Rename "init" command #838 (@technoweenie)
* Raise error if credentials are needed #842 (@technoweenie)
* Support git repos in symlinked directories #818 (@sinbad, @difro, @jiangxin)
* Fix "git lfs env" to show correct SSH remote info #828 (@jiangxin)

## 1.0.2 (28 October, 2015)

* Fix issue with 'git lfs smudge' and the batch API. #795 (@technoweenie)
* Fix race condition in the git scanning code. #801 (@technoweenie)

## 1.0.1 (23 October, 2015)

* Downcase git config keys (prevents Auth loop) #690 (@WillHipschman)
* Show more info for unexpected http responses #710 (@rubyist)
* Use separate stdout/stderr buffers for `git-lfs-authenticate` #718 (@bozaro)
* Use LoggedError instead of Panic if update-index fails in checkout #735 (@sinbad)
* `smudge` command exits with non-zero if the download fails #732 (@rubyist)
* Use `git rev-parse` to find the git working dir #692 (@sinbad)
* Improved default remote behaviour & validation for fetch/pull #713 (@sinbad)
* Make fetch return error code when 1+ downloads failed #734 (@sinbad)
* Improve lfs.InRepo() detection in `init`/`update` #756 (@technoweenie)
* Teach smudge to use the batch api #711 (@rubyist)
* Fix not setting global attribute when needed to b/c of local state #765 (@sinbad)
* Fix clone fail when fetch is excluded globally #770 (@sinbad)
* Fix for partial downloads problem #763 (@technoweenie)
* Get integration tests passing on Windows #771 (@sinbad)

### Security

* Whitelist the valid keys read from .gitconfig #760 (@technoweenie)

This prevents unsafe git configuration values from being used by Git LFS.

## v1.0 (1 October, 2015)

* Manual reference is integrated into the "help" options #665 @sinbad
* Fix `ls-files` when run from an empty repository #668 @Aorjoa
* Support listing duplicate files in `ls-files` #681 @Aorjoa @technoweenie
* `update` and `init` commands can install the pre-push hook in bare repositories #671 @technoweenie
* Add `GIT_LFS_SKIP_SMUDGE` and `init --skip-smudge` #679 @technoweenie

## v0.6.0 (10 September, 2015)

This is the first release that uses the new Batch API by default, while still
falling back to the Legacy API automatically. Also, new fetch/checkout/push
commands have been added.

Run `git lfs update` in any local repositories to make sure all config settings
are updated.

### Fetch

* Rename old `git lfs fetch` command to `git lfs pull`. #527 (@sinbad)
* Add `git lfs checkout` #527 #543 #551 #566 (@sinbad)
* Add `git lfs fetch` for _just_ downloading objects. #527 (@sinbad)
  * Add `remote` arg, and default to tracking remote instead of "origin". #583 (@sinbad)
  * Support fetching multiple refs #542 (@sinbad)
  * Add `--include` and `--exclude` flag for `git lfs fetch` #573 (@sinbad)
  * Add `--recent` flag for downloading recent files outside of the current
    checkout. #610 (@sinbad)
  * Add `--all` option for download all objects from the server. #633 (@sinbad)
* Fix error handling while `git update-index` is running. #570 (@rubyist)

See [git-lfs-fetch(1)](https://github.com/github/git-lfs/blob/v0.6.0/docs/man/git-lfs-fetch.1.ronn),
[git-lfs-checkout(1)](https://github.com/github/git-lfs/blob/v0.6.0/docs/man/git-lfs-checkout.1.ronn),
and [git-lfs-pull(1)](https://github.com/github/git-lfs/blob/v0.6.0/docs/man/git-lfs-pull.1.ronn)
 for details.

### Push

* Support pushing multiple branches in the pre-push hook. #635 (@sinbad)
* Fix pushing objects from a branch that's not HEAD. #608 (@sinbad)
* Check server for objects before failing push because local is missing. #581 (@sinbad)
* Filter out commits from remote refs when pushing. #578 (@billygor)
* Support pushing all objects to the server, regardless of the remote ref. #646 (@technoweenie)
* Fix case where pre-push git hook exits with 0. #582 (@sinbad)

See [git-lfs-push(1)](https://github.com/github/git-lfs/blob/v0.6.0/docs/man/git-lfs-push.1.ronn) for details.

### API Clients

* Fix some race conditions in the Batch API client. #577 #637 (@sinbad, @rubyist)
* Support retries in the Batch API client. #595 (@rubyist)
* Fix hanging batch client in certain error conditions. #594 (@rubyist)
* Treat 401 responses as errors in the Legacy API client. #634 (@rubyist)
* Fix bug in the Legacy API client when the object already exists on the server. #572 (@billygor)

### Credentials

* Fix how git credentials are checked in certain edge cases. #611 #650 #652 (@technoweenie)
* Send URI user to git credentials. #626 (@sinbad)
* Support git credentials with useHttpPath enabled. #554 (@clareliguori)

### Installation

* Docker images and scripts for building and testing linux packages. #511 #526 #555 #603 (@andyneff, @ssgelm)
* Create Windows GUI installer. #642 (@technoweenie)
* Binary releases use Go 1.5, which includes fix for Authorization when the
  request URL includes just the username. [golang/go#11399](https://github.com/golang/go/issues/11399)

### Misc

* Documented Git config values used by Git LFS in [git-lfs-config(5)](https://github.com/github/git-lfs/blob/v0.6.0/docs/man/git-lfs-config.5.ronn). #610 (@sinbad)
* Experimental support for Git worktrees (in Git 2.5+) #546 (@sinbad)
* Experimental extension support. #486 (@ryansimmen)

## v0.5.4 (30 July, 2015)

* Ensure `git lfs uninit` cleans your git config thoroughly. #530 (@technoweenie)
* Fix issue with asking `git-credentials` for auth details after getting them
from the SSH command. #534 (@technoweenie)

## v0.5.3 (23 July, 2015)

* `git lfs fetch` bugs #429 (@rubyist)
* Push can crash on 32 bit architectures #450 (@rubyist)
* Improved SSH support #404, #464 (@sinbad, @technoweenie)
* Support 307 redirects with relative url #442 (@sinbad)
* Fix `init` issues when upgrading #446 #451 #452 #465 (@technoweenie, @rubyist)
* Support chunked Transfer-Encoding #386 (@ryansimmen)
* Fix issue with pushing deleted objects #461 (@technoweenie)
* Teach `git lfs push` how to send specific objects #449 (@larsxschneider)
* Update error message when attempting to push objects that don't exist in `.git/lfs/objects` #447 (@technoweenie)
* Fix bug in HTTP client when response body is nil #472 #488 (@rubyist, @technoweenie)
* `-crlf` flag in gitattributes is deprecated #475 (@technoweenie)
* Improvements to the CentOS and Debian build and package scripts (@andyneff, @ssgelm)

## v0.5.2 (19 June, 2015)

* Add `git lfs fetch` command for downloading objects. #285 (@rubyist)
* Fix `git lfs track` issues when run outside of a git repository #312, #323 (@michael-k, @Aorjoa)
* Fix `git lfs track` for paths with spaces in them #327 (@technoweenie)
* Fix `git lfs track` by writing relative paths to .gitattributes #356 (@michael-k)
* Fix `git lfs untrack` so it doesn't remove entries incorrectly from .gitattributes #398 (@michael-k)
* Fix `git lfs clean` bug with zero length files #346 (@technoweenie)
* Add `git lfs fsck` #373 (@zeroshirts, @michael-k)
* The Git pre-push warns if Git LFS is not installed #339 (@rubyist)
* Fix Content-Type header sent by the HTTP client #329 (@joerg)
* Improve performance tracing while scanning refs #311 (@michael-k)
* Fix detection of LocalGitDir and LocalWorkingDir #312 #354 #361 (@michael-k)
* Fix inconsistent file mode bits for directories created by Git LFS #364 (@michael-k)
* Optimize shell execs #377, #382, #391 (@bozaro)
* Collect HTTP transfer stats #366, #400 (@rubyist)
* Support GIT_DIR and GIT_WORK_TREE #370 (@michael-k)
* Hide Git application window in Windows #381 (@bozaro)
* Add support for configured URLs containing credentials per RFC1738 #408 (@ewbankkit, @technoweenie)
* Add experimental support for batch API calls #285 (@rubyist)
* Improve linux build instructions for CentOS and Debian. #299 #309 #313 #332 (@jsh, @ssgelm, @andyneff)

## v0.5.1 (30 April, 2015)

* Fix Windows install.bat script.  #223 (@PeterDaveHello)
* Fix bug where `git lfs clean` will clean Git LFS pointers too #271 (@technoweenie)
* Better timeouts for the HTTP client #215 (@Mistobaan)
* Concurrent uploads through `git lfs push` #258 (@rubyist)
* Fix `git lfs smudge` behavior with zero-length file in `.git/lfs/objects` #267 (@technoweenie)
* Separate out pre-push hook behavior from `git lfs push` #263 (@technoweenie)
* Add diff/merge properties to .gitattributes #265 (@technoweenie)
* Respect `GIT_TERMINAL_PROMPT ` #257 (@technoweenie)
* Fix CLI progress bar output #185 (@technoweenie)
* Fail fast in `clean` and `smudge` commands when run without STDIN #264 (@technoweenie)
* Fix shell quoting in pre-push hook.  #235 (@mhagger)
* Fix progress bar output during file uploads.  #185 (@technoweenie)
* Change `remote.{name}.lfs_url` to `remote.{name}.lfsurl` #237 (@technoweenie)
* Swap `git config` order.  #245 (@technoweenie)
* New `git lfs pointer` command for generating and comparing pointers #246 (@technoweenie)
* Follow optional "href" property from git-lfs-authenticate SSH command #247 (@technoweenie)
* `.git/lfs/objects` spec clarifications: #212 (@rtyley), #244 (@technoweenie)
* man page updates: #228 (@mhagger)
* pointer spec clarifications: #246 (@technoweenie)
* Code comments for the untrack command: #225 (@thekafkaf)

## v0.5.0 (10 April, 2015)

* Initial public release
