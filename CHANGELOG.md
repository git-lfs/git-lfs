# Git LFS Changelog

## 2.0.0 (1 March, 2017)

Git LFS v2.0.0 brings a number of important bug fixes, some new features, and
a lot of internal refactoring. It also completely removes old APIs that were
deprecated in Git LFS v0.6.

### Locking

File Locking is a brand new feature that lets teams communicate when they are
working on files that are difficult to merge. Users are not able to edit or push
changes to any files that are locked by other users. While the feature has been
in discussion for a year, we are releasing a basic Locking implementation to
solicit feedback from the community.

### Transfer Queue

LFS 2.0 introduces a new Git Scanner, which walks a range of Git commits looking
for LFS objects to transfer. The Git Scanner is now asynchronous, initiating
large uploads or downloads in the Transfer Queue immediately once an LFS object
is found. Previously, the Transfer Queue waited until all of the Git commits
have been scanned before initiating the transfer. The Transfer Queue also
automatically retries failed uploads and downloads more often.

### Deprecations

Git LFS v2.0.0 also drops support for the legacy API in v0.5.0. If you're still
using LFS servers on the old API, you'll have to stick to v1.5.6.

### Features

* Mid-stage locking support #1769 (@sinbad)
* Define lockable files, make read-only in working copy #1870 (@sinbad)
* Check that files are not uncommitted before unlock #1896 (@sinbad)
* Fix `lfs unlock --force` on a missing file #1927 (@technoweenie)
* locking: teach pre-push hook to check for locks #1815 (@ttaylorr)
* locking: add `--json` flag #1814 (@ttaylorr)
* Implement local lock cache, support querying it #1760 (@sinbad)
* support for client certificates pt 2 #1893 (@technoweenie)
* Fix clash between progress meter and credential helper #1886 (@technoweenie)
* Teach uninstall cmd about --local and --system #1887 (@technoweenie)
* Add `--skip-repo` option to `git lfs install` & use in tests #1868 (@sinbad)
* commands: convert push, pre-push to use async gitscanner #1812 (@ttaylorr)
* tq: prioritize transferring retries before new items #1758 (@ttaylorr)

### Bugs

* ensure you're in the correct directory when installing #1793 (@technoweenie)
* locking: make API requests relative to repository, not root #1818 (@ttaylorr)
* Teach 'track' about CRLF #1914 (@technoweenie)
* Teach 'track' how to handle empty lines in .gitattributes #1921 (@technoweenie)
* Closing stdout pipe before function return #1861 (@monitorjbl)
* Custom transfer terminate #1847 (@sinbad)
* Fix Install in root problems #1727 (@technoweenie)
* cat-file batch: read all of the bytes #1680 (@technoweenie)
* Fixed file paths on cygwin. #1820, #1965 (@creste, @ttaylorr)
* tq: decrement uploaded bytes in basic_upload before retry #1958 (@ttaylorr)
* progress: fix never reading bytes with sufficiently small files #1955 (@ttaylorr)
* tools: fix truncating string fields between balanced quotes in GIT_SSH_COMMAND #1962 (@ttaylorr)
* commands/smudge: treat empty pointers as empty files #1954 (@ttaylorr)

### Misc

* all: build using Go 1.8 #1952 (@ttaylorr)
* Embed the version information into the Windows executable #1689 (@sschuberth)
* Add more meta-data to the Windows installer executable #1752 (@sschuberth)
* docs/api: object size must be positive #1779 (@ttaylorr)
* build: omit DWARF tables by default #1937 (@ttaylorr)
* Add test to prove set operator [] works in filter matching #1768 (@sinbad)
* test: add ntlm integration test #1840 (@technoweenie)
* lfs/tq: completely remove legacy support #1686 (@ttaylorr)
* remove deprecated features #1679 (@technoweenie)
* remove legacy api support #1629 (@technoweenie)

## 1.5.6 (16 February, 2017)

## Bugs

* Spool malformed pointers to avoid deadlock #1932 (@ttaylorr)

## 1.5.5 (12 January, 2017)

### Bugs

* lfs: only buffer first 1k when creating a CleanPointerError #1856 (@ttaylorr)

## 1.5.4 (27 December, 2016)

### Bugs

* progress: guard negative padding width, panic in `strings.Repeat` #1807 (@ttaylorr)
* commands,lfs: handle malformed pointers #1805 (@ttaylorr)

### Misc

* script/packagecloud: release LFS on fedora/25 #1798 (@ttaylorr)
* backport filepathfilter to v1.5.x #1782 (@technoweenie)

## 1.5.3 (5 December, 2016)

### Bugs

* Support LFS installations at filesystem root #1732 (@technoweenie)
* git: parse filter process header values containing '=' properly #1733 (@larsxschneider)
* Fix SSH endpoint parsing #1738 (@technoweenie)

### Misc

* build: release on Go 1.7.4 #1741 (@ttaylorr)

## 1.5.2 (22 November, 2016)

### Features

* Release LFS on Fedora 24 #1685 (@technoweenie)

### Bugs

* filter-process: fix reading 1024 byte files #1708 (@ttaylorr)
* Support long paths on Windows #1705 (@technoweenie)

### Misc

* filter-process: exit with error if we detect an unknown command from Git #1707 (@ttaylorr)
* vendor: remove contentaddressable lib #1706 (@technoweenie)

## 1.5.1 (18 November, 2016)

### Bugs

* cat-file --batch parser errors on non-lfs git blobs #1680 (@technoweenie)

## 1.5.0 (17 November, 2016)

### Features

* Filter Protocol Support #1617 (@ttaylorr, @larsxschneider)
* Fast directory walk #1616 (@sinbad)
* Allow usage of proxies even when contacting localhost #1605 (@chalstrick)

### Bugs

* start reading off the Watch() channel before sending any input #1671 (@technoweenie)
* wait for remote ref commands to exit before returning #1656 (@jjgod, @technoweenie)

### Misc

* rewrite new catfilebatch implementation for upcoming gitscanner pkg #1650 (@technoweenie)
* refactor testutils.FileInput so it's a little more clear #1666 (@technoweenie)
* Update the lfs track docs #1642 (@technoweenie)
* Pre push tracing #1638 (@technoweenie)
* Remove `AllGitConfig()` #1634 (@technoweenie)
* README: set minimal required Git version to 1.8.5 #1636 (@larsxschneider)
* 'smudge --info' is deprecated in favor of 'ls-files' #1631 (@technoweenie)
* travis-ci: test GitLFS with ancient Git version #1626 (@larsxschneider)

## 1.4.4 (24 October, 2016)

### Bugs

* transfer: more descriptive "expired at" errors #1603 (@ttaylorr)
* commands,lfs/tq: Only send unique OIDs to the Transfer Queue #1600 (@ttaylorr)
* Expose the result message in case of an SSH authentication error #1599 (@sschuberth)

### Misc

* AppVeyor: Do not build branches with open pull requests #1594 (@sschuberth)
* Update .mailmap #1593 (@dpursehouse)

## 1.4.3 (17 October, 2016)

### Bugs

* lfs/tq: use extra arguments given to tracerx.Printf #1583 (@ttaylorr)
* api: correctly print legacy API warning to Stderr #1582 (@ttaylorr)

### Misc

* Test storage retries #1585 (@ttaylorr)
* Test legacy check retries behavior #1584 (@ttaylorr)
* docs: Fix a link to the legacy API #1579 (@sschuberth)
* Add a .mailmap file #1577 (@sschuberth)
* Add a large wizard image to the Windows installer #1575 (@sschuberth)
* Appveyor badge #1574 (@ttaylorr)

## 1.4.2 (10 October, 2016)

v1.4.2 brings a number of bug fixes and usability improvements to LFS. This
release also adds support for multiple retries within the transfer queue, making
transfers much more reliable. To enable this feature, see the documentation for
`lfs.transfer.maxretries` in `git-lfs-config(5)`.

We'd also like to extend a special thank-you to @sschuberth who undertook the
process of making LFS's test run on Windows through AppVeyor. Now all pull
requests run tests on macOS, Linux, and Windows.

### Features

* lfs: warn on usage of the legacy API #1564 (@ttaylorr)
* use filepath.Clean() when comparing filenames to include/exclude patterns #1565 (@technoweenie)
* lfs/transfer_queue: support multiple retries per object #1505, #1528, #1535, #1545 (@ttaylorr)
* Automatically upgrade old filters instead of requiring —force #1497 (@sinbad)
* Allow lfs.pushurl in .lfsconfig #1489 (@technoweenie)

### Bugs

* Use "sha256sum" on Windows  #1566 (@sschuberth)
* git: ignore non-root wildcards #1563 (@ttaylorr)
* Teach status to recognize multiple files with identical contents #1550 (@ttaylorr)
* Status initial commit #1540 (@sinbad)
* Make path comparison robust against Windows short / long path issues #1523 (@sschuberth)
* Allow fetch to run without a remote configured #1507 (@sschuberth)

### Misc

* travis: run tests on Go 1.7.1 #1568 (@ttaylorr)
* Enable running tests on AppVeyor CI #1567 (@sschuberth)
* Travis: Only install git if not installed yet #1557 (@sschuberth)
* Windows test framework fixes #1522 (@sschuberth)
* Simplify getting the absolute Git root directory #1518 (@sschuberth)
* Add icons to the Windows installer #1504 (@sschuberth)
* docs/man: reference git-lfs-pointer(1) in clean documentation #1503 (@ttaylorr)
* Make AppVeyor CI for Windows work again #1506 (@sschuberth)
* commands: try out RegisterCommand() #1495 (@technoweenie)

## 1.4.1 (26 August, 2016)

### Features

* retry if file download failed #1454 (@larsxschneider)
* Support wrapped clone in current directory #1478 (@ttaylorr)

### Misc

* Test `RetriableReader` #1482 (@ttaylorr)

## 1.4.0 (19 August, 2016)

### Features

* Install LFS at the system level when packaged #1460 (@javabrett)
* Fetch remote urls #1451 (@technoweenie)
* add object Authenticated property #1452 (@technoweenie)
* add support for `url.*.insteadof` in git config #1117, #1443 (@artagnon, @technoweenie)

### Bugs

* fix --include bug when multiple files have same lfs content #1458 (@technoweenie)
* check the git version is ok in some key commands #1461 (@technoweenie)
* fix duplicate error reporting #1445, #1453 (@dpursehouse, @technoweenie)
* transfer/custom: encode "event" as lowercase #1441 (@ttaylorr)

### Misc

* docs/man: note GIT_LFS_PROGRESS #1469 (@ttaylorr)
* Reword the description of HTTP 509 status #1467 (@dpursehouse)
* Update fetch include/exclude docs for pattern matching #1455 (@ralfthewise)
* config-next: API changes to the `config` package #1425 (@ttaylorr)
* errors-next: Contextualize error messages #1463 (@ttaylorr, @technoweenie)
* scope commands to not leak instances of themselves #1434 (@technoweenie)
* Transfer manifest #1430 (@technoweenie)

## 1.3.1 (2 August 2016)

### Features

* lfs/hook: teach `lfs.Hook` about `core.hooksPath` #1409 (@ttaylorr)

### Bugs

* distinguish between empty include/exclude paths #1411 (@technoweenie)
* Fix sslCAInfo config lookup when host in config doesn't have a trailing slash #1404 (@dakotahawkins)

### Misc

* Use commands.Config instead of config.Config #1390 (@technoweenie)

## 1.3.0 (21 July 2016)

### Features

* use proxy from git config #1173, #1358 (@jonmagic, @LizzHale, @technoweenie)
* Enhanced upload/download of LFS content: #1265 #1279 #1297 #1303 #1367 (@sinbad)
  * Resumable downloads using HTTP range headers
  * Resumable uploads using [tus.io protocol](http://tus.io)
  * Pluggable [custom transfer adapters](https://github.com/git-lfs/git-lfs/blob/master/docs/custom-transfers.md)
* In git 2.9+, run "git lfs pull" in submodules after "git lfs clone" #1373 (@sinbad)
* cmd,doc,test: teach `git lfs track --{no-touch,verbose,dry-run}` #1344 (@ttaylorr)
* ⏳ Retry transfers with expired actions #1350 (@ttaylorr)
* Safe track patterns #1346 (@ttaylorr)
* Add checkout --unstaged flag #1262 (@orivej)
* cmd/clone: add include/exclude via flags and config #1321 (@ttaylorr)
* Improve progress reporting when files skipped #1296 (@sinbad)
* Experimental file locking commands #1236, #1259, #1256, #1386 (@ttaylorr)
* Implement support for GIT_SSH_COMMAND #1260 (@pdf)
* Recognize include/exclude filters from config #1257 (@ttaylorr)

### Bugs

* Fix bug in Windows installer under Win32. #1200 (@teo-tsirpanis)
* Updated request.GetAuthType to handle multi-value auth headers #1379 (@VladimirKhvostov)
* Windows fixes #1374 (@sinbad)
* Handle artifactory responses #1371 (@ttaylorr)
* use `git rev-list --stdin` instead of passing each remote ref #1359 (@technoweenie)
* docs/man: move "logs" subcommands from OPTIONS to COMMANDS #1335 (@ttaylorr)
* test/zero-len: update test for git v2.9.1 #1369 (@ttaylorr)
* Unbreak building httputil on OpenBSD #1360 (@jasperla)
* WIP transferqueue race fix #1255 (@technoweenie)
* Safety check to `comands.requireStdin` #1349 (@ttaylorr)
* Removed CentOS 5 from dockers. Fixed #1295. #1298 (@javabrett)
* Fix 'git lfs fetch' with a sha1 ref #1323 (@omonnier)
* Ignore HEAD ref when fetching with --all #1310 (@ttaylorr)
* Return a fully remote ref to reduce chances of ref clashes #1248 (@technoweenie)
* Fix reporting of `git update-index` errors in `git lfs checkout` and `git lfs pull` #1400 (@technoweenie)

### Misc

* Added Linux Mint Sarah to package cloud script #1384 (@andyneff)
* travis-ci: require successful tests against upcoming Git core release #1372 (@larsxschneider)
* travis-ci: add a build job to test against upcoming versions of Git #1361 (@larsxschneider)
* Create Makefiles for building with gccgo #1222 (@zeldin)
* README: add @ttaylorr to core team #1332 (@ttaylorr)
* Enforced a minimum gem version of 1.0.4 for packagecloud-ruby #1292 (@javabrett)
* I think this should be "Once installed" and not "One installed", but … #1305 (@GabLeRoux)
* script/test: propagate extra args to go test #1324 (@omonnier)
* Add `lfs.basictransfersonly` option to disable non-basic transfer adapters #1299 (@sinbad)
* Debian build vendor test excludes #1291 (@javabrett)
* gitignore: ignore lfstest-\* files #1271 (@ttaylorr)
* Disable gojsonschema test, causes failures when firewalls block it #1274 (@sinbad)
* test: use noop credential helper for auth tests #1267 (@ttaylorr)
* get git tests passing when run outside of repository #1229 (@technoweenie)
* Package refactor no.1 #1226 (@sinbad)
* vendor: vendor dependencies in vendor/ using Glide #1243 (@ttaylorr)

## 1.2.1 (2 June 2016)

### Features

* Add missing config details to `env` command #1217 (@sinbad)
* Allow smudge filter to return 0 on download failure #1213 (@sinbad)
* Add `git lfs update --manual` option & promote it on hook install fail #1182 (@sinbad)
* Pass `git lfs clone` flags through to `git clone` correctly, respect some options #1160 (@sinbad)

### Bugs

* Clean trailing `/` from include/exclude paths #1278 (@ttaylorr)
* Fix problems with user prompts in `git lfs clone` #1185 (@sinbad)
* Fix failure to return non-zero exit code when lfs install/update fails to install hooks #1178 (@sinbad)
* Fix missing man page #1149 (@javabrett)
* fix concurrent map read and map write #1179 (@technoweenie)

### Misc

* Allow additional fields on request & response schema #1276 (@sinbad)
* Fix installer error on win32. #1198 (@teo-tsirpanis)
* Applied same -ldflags -X name value -> name=value fix #1193 (@javabrett)
* add instructions to install from MacPorts #1186 (@skymoo)
* Add xenial repo #1170 (@graingert)

## 1.2.0 (14 April 2016)

### Features

* netrc support #715 (@rubyist)
* `git lfs clone` command #988 (@sinbad)
* Support self-signed certs #1067 (@sinbad)
* Support sslverify option for specific hosts #1081 (@sinbad)
* Stop transferring duplicate objects on major push or fetch operations on multiple refs. #1128 (@technoweenie)
* Touch existing git tracked files when tracked in LFS so they are flagged as modified #1104 (@sinbad)
* Support for git reference clones #1007 (@jlehtnie)

### Bugs

* Fix clean/smudge filter string for files starting with - #1083 (@epriestley)
* Fix silent failure to push LFS objects when ref matches a filename in the working copy #1096 (@epriestley)
* Fix problems with using LFS in symlinked folders #818 (@sinbad)
* Fix git lfs push silently misbehaving on ambiguous refs; fail like git push instead #1118 (@sinbad)
* Whitelist `lfs.*.access` config in local ~/.lfsconfig #1122 (@rjbell4)
* Only write the encoded pointer information to Stdout #1105 (@sschuberth)
* Use hardcoded auth from remote or lfs config when accessing the storage api #1136 (@technoweenie, @jonmagic)
* SSH should be called more strictly with command as one argument #1134 (@sinbad)

## 1.1.2 (1 March, 2016)

* Fix Base64 issues with `?` #989 (@technoweenie)
* Fix zombie git proc issue #1012 (@rlaakkol)
* Fix problems with files containing unicode characters #1016 (@technoweenie)
* Fix panic in `git cat-file` parser #1006 (@technoweenie)
* Display error messages in non-fatal errors #1028 #1039 #1042 (@technoweenie)
* Fix concurrent map access in progress meter (@technoweenie)

## 1.1.1 (4 February, 2016)

### Features

* Add copy-on-write support for Linux BTRFS filesystem #952 (@bozaro)
* convert `git://` remotes to LFS servers automatically #964 (@technoweenie)
* Fix `git lfs track` handling of absolute paths. #975  (@technoweenie)
* Allow tunable http client timeouts #977 (@technoweenie)

### Bugs

* Suppress git config warnings for non-LFS keys #861 (@technoweenie)
* Fix fallthrough when `git-lfs-authenticate` returns an error #909 (@sinbad)
* Fix progress bar issue #883 (@pokehanai)
* Support `remote.name.pushurl` config #949 (@sinbad)
* Fix handling of `GIT_DIR` and `GIT_WORK_TREE` #963, #971 (@technoweenie)
* Fix handling of zero length files #966 (@nathanhi)
* Guard against invalid remotes passed to `push` and `pre-push` #974 (@technoweenie)
* Fix race condition in `git lfs pull` #972 (@technoweenie)

### Extra

* Add server API test tool #868 (@sinbad)
* Redo windows installer with innosetup #875 (@strich)
* Pre-built packages are built with Go v1.5.3

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

See [git-lfs-fetch(1)](https://github.com/git-lfs/git-lfs/blob/v0.6.0/docs/man/git-lfs-fetch.1.ronn),
[git-lfs-checkout(1)](https://github.com/git-lfs/git-lfs/blob/v0.6.0/docs/man/git-lfs-checkout.1.ronn),
and [git-lfs-pull(1)](https://github.com/git-lfs/git-lfs/blob/v0.6.0/docs/man/git-lfs-pull.1.ronn)
 for details.

### Push

* Support pushing multiple branches in the pre-push hook. #635 (@sinbad)
* Fix pushing objects from a branch that's not HEAD. #608 (@sinbad)
* Check server for objects before failing push because local is missing. #581 (@sinbad)
* Filter out commits from remote refs when pushing. #578 (@billygor)
* Support pushing all objects to the server, regardless of the remote ref. #646 (@technoweenie)
* Fix case where pre-push git hook exits with 0. #582 (@sinbad)

See [git-lfs-push(1)](https://github.com/git-lfs/git-lfs/blob/v0.6.0/docs/man/git-lfs-push.1.ronn) for details.

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

* Documented Git config values used by Git LFS in [git-lfs-config(5)](https://github.com/git-lfs/git-lfs/blob/v0.6.0/docs/man/git-lfs-config.5.ronn). #610 (@sinbad)
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
