# Git LFS Changelog

## 2.6.0 (1 November, 2018)

This release adds better support for redirecting network calls from a Git LFS
API server to one that requires a different authentication mode, builds Git LFS
on Go 1.11, and numerous other bug fixes and modifications.

We would like to extend a special thanks to the following open-source
contributors:

* @andyneff for updating our release targets
* @gtsiolis: for removing the deprecated `git lfs clone` from the listing of
  supported Git LFS commands
* @jsantell for fixing a formatting issue in the INCLUDE AND EXCLUDE man page
  section
* @mmlb for adding a release target for Linux arm64
* @skashyap7 for adding the 'git lfs track -n'
* @Villemoes: for modernizing the Git LFS installation procedure on Debian.

### Features

* commands: list explicitly excluded patterns separately #3320 (@bk2204)
* Uninstall improvements #3326 (@bk2204)
* config: honor GIT_AUTHOR_DATE and GIT_COMMITTER_DATE #3314 (@bk2204)
* Add new `.netrc` credential helper #3307 (@PastelMobileSuit)
* Honor umask and core.sharedRepository #3304 (@bk2204)
* Support listing only filename tracked by git lfs using --name (-n) option
  #3271 (@skashyap7)
* all: use Go 1.11.1 in CI #3298 (@ttaylorr)
* lfsapi/tq: Have DoWithAuth() caller determine URL Access Mode #3293
  (@PastelMobileSuit)
* commands: undeprecate checkout #3303 (@bk2204)
* Checkout options for conflicts #3296 (@bk2204)
* Makefile: build source tarballs for release #3283 (@bk2204)
* Encrypted SSL key support #3270 (@bk2204)
* Add support for core.sshCommand #3235 (@bk2204)
* gitobj-based Object Scanner #3236 (@bk2204)
* README.md: new core team members #3217 (@ttaylorr)
* Add build and releases for linux arm64 #3196 (@mmlb)
* Update packagecloud.rb #3210 (@andyneff)
* all: use Go modules instead of Glide #3208 (@ttaylorr)
* all: use Go 1.11 in CI #3203 (@ttaylorr)

### Bugs

* Fix formatting of INCLUDE AND EXCLUDE (REFS) #3330 (@jsantell)
* go.sum: add missing entries #3319 (@bk2204)
* Ensure correct syntax for commit headers in lfs migrate import #3313 (@bk2204)
* Clean up trailing whitespace #3299 (@bk2204)
* commands: unambiguously resolve remote references #3285 (@ttaylorr)
* Expand custom transfer args by using the shell #3259 (@bk2204)
* Canonicalize paths properly on Windows #3277 (@bk2204)
* debian/prerm: add --system flag #3272 (@Villemoes)
* t: make testsuite run under git rebase -x #3262 (@bk2204)
* git/gitattr: parse 'set' attributes #3255 (@ttaylorr)
* t: avoid panic in lfstest-customadapter #3243 (@bk2204)
* t: avoid using shell variables in printf's first argument #3242 (@bk2204)
* lfsapi: handle SSH hostnames and aliases without users #3230 (@bk2204)
* commands/command_ls_files.go: ignore index with argument #3219 (@ttaylorr)
* commands/command_migrate_import.go: install hooks #3227 (@ttaylorr)
* t: mark test sources as .PHONY #3228 (@ttaylorr)
* Pass GIT_SSH_COMMAND to the shell #3199 (@bk2204)
* Tidy misformatted files #3202 (@bk2204)
* config: expand core.hooksPath #3212 (@ttaylorr)
* locks: manage write permissions of ignored files #3190 (@ttaylorr)

### Misc

* CONTRIBUTING.md: :nail_care: #3325 (@ttaylorr)
* Update CONTRIBUTING #3317 (@bk2204)
* go.mod: depend on tagged gitobj #3311 (@ttaylorr)
* RFC: SSH protocol #3290 (@bk2204)
* Remove `git lfs clone` command from man #3301 (@gtsiolis)
* ROADMAP.md: use GitHub issues instead #3286 (@ttaylorr)
* docs: add note about closing release milestone #3274 (@bk2204)
* CI improvements #3268 (@bk2204)
* docs/howto: document our release process #3261 (@ttaylorr)
* Create new lfshttp package #3244 (@PastelMobileSuit)
* CONTRIBUTING: update required go version #3232 (@PastelMobileSuit)
* go.mod: use latest github.com/olekukonko/ts #3223 (@ttaylorr)
* go.mod: pin github.com/git-lfs/wildmatch to v1.0.0 #3218 (@ttaylorr)
* Update README.md #3193 (@srl295)

## 2.5.2 (17 September, 2018)

### Bugs

* config: Treat [host:port]:path URLs correctly #3226 (@saschpe)
* tq: Always provide a Content-Type when uploading files #3201 (@bk2204)
* commands/track: Properly `lfs track` files with escaped characters in their
  name #3192 (@leonid-s-usov)

### Misc

* packagecloud.rb: remove older versions #3210 (@andyneff)

## 2.5.1 (2 August, 2018)

This release contains miscellaneous bug fixes since v2.5.0. Most notably,
release v2.5.1 allows a user to disable automatic Content-Type detection
(released in v2.5.0) via `git config lfs.contenttype false` for hosts that do
not support it.

### Features

* tq: make Content-Type detection disable-able #3163 (@ttaylorr)

### Bugs

* Makefile: add explicit rule for commands/mancontent_gen.go #3160 (@jj1bdx)
* script/install.sh: mark as executable #3155 (@ttaylorr)
* config: add origin to remote list #3152 (@PastelMobileSuit)

### Misc

* docs/man/mangen.go: don't show non-fatal output without --verbose #3168
  (@ttaylorr)
* LICENSE.md: update copyright year #3156 (@IMJ355)
* Makefile: silence some output #3164 (@ttaylorr)
* Makefile: list prerequisites for resource.syso #3153 (@ttaylorr)

## 2.5.0 (26 July, 2018)

This release adds three new migration modes, updated developer ergonomics, and
a handful of bug fixes to Git LFS.

We would like to extend a special thanks to the following open-source
contributors:

* @calavera for fixing a broken Go test and adding support for custom
  Content-Type headers in #3137 and #3138.
* @cbuehlmann for adding support for encoded character names in filepaths via
  #3093.
* @larsxschneider for changing the default value of lfs.allowincompletepush in
  #3109.
* @NoEffex for supporting TTL in SSH-based authentication tokens via #2867.
* @ssgelm for adding 'go generate' to our Debian packages via #3083.

### Features

* Makefile: replace many scripts with make targets #3144 (@ttaylorr)
* {.travis,appveyor}.yml: upgrade to Go 1.10.3 #3146 (@ttaylorr)
* t: run tests using prove #3125 (@ttaylorr)
* commands/migrate: infer wildmatches with --fixup #3114 (@ttaylorr)
* Retry SSH resolution 5 times #2934 (@stanhu)
* Implement `migrate export` subcommand #3084 (@PastelMobileSuit)
* Add `--no-rewrite` flag to `migrate import` command #3029 (@PastelMobileSuit)

### Bugs

* t: fix contains_same_elements() fn #3145 (@PastelMobileSuit)
* commands: warn if working copy is dirty #3124 (@ttaylorr)
* Ensure provided remote takes precedence over configured pushRemote #3139 (@PastelMobileSuit)
* Fix proxy unit tests. #3138 (@calavera)
* commands/command_migrate.go: loosen meaning of '--everything' #3121 (@ttaylorr)
* lfsapi: don't query askpass for given creds #3126 (@PastelMobileSuit)
* config/git_fetcher.go: mark 'lfs.allowincompletepush' as safe #3113 (@ttaylorr)
* fs: support multiple object alternates #3116 (@ttaylorr)
* commands/checkout: checkout over read-only files #3120 (@ttaylorr)
* test/testhelpers.sh: look for 64 character SHA-256's #3119 (@ttaylorr)
* config/config.go: case-insensitive error search #3098 (@ttaylorr)
* Encoded characters in pathnames #3093 (@cbuehlmann)
* Support default TTL for authentication tokens acquired via SSH #2867 (@NoEffex)
* commands/status.go: relative paths outside of root #3080 (@ttaylorr)
* Run `go generate` on commands in deb build #3083 (@ssgelm)
* lfsapi: prefer proxying from gitconfig before environment #3062 (@ttaylorr)
* commands/track: respect global- and system-level gitattributes #3076 (@ttaylorr)
* git/git.go: pass --multiple to git-fetch(1) when appropriate #3063 (@ttaylorr)
* commands/checkout: fix inaccurate messaging #3055 (@ttaylorr)
* commands/migrate: do not migrate empty commits #3054 (@ttaylorr)
* git/odb: retain trailing newlines in commit messages #3053 (@ttaylorr)

### Misc

* Set original file content type on basic upload. #3137 (@calavera)
* README.md: Git for Windows ships LFS by default #3112 (@larsxschneider)
* change lfs.allowincompletepush default from true to false  #3109 (@larsxschneider)
* *: replace git/odb with vendored copy #3108 (@ttaylorr)
* test/test-ls-files.sh: skip on CircleCI #3101 (@ttaylorr)
* lfsapi/ssh.go: use zero-value sentinels #3099 (@ttaylorr)
* README.md: add link to installation wiki page #3075 (@ttaylorr)
* docs/man/git-lfs.1.ronn: update casing and missing commands #3059 (@ttaylorr)
* commands/checkout: mark 'git lfs checkout' as deprecated #3056 (@ttaylorr)

## 2.4.2 (28 May, 2018)

### Bugs

* lfsapi: re-authenticate HTTP redirects when needed #3028 (@ttaylorr)
* lfsapi: allow unknown keywords in netrc file(s) #3027 (@ttaylorr)

## 2.4.1 (18 May, 2018)

This release fixes a handful of bugs found and fixed since v2.4.0. In
particular, Git LFS no longer panic()'s after invalid API responses, can
correctly run 'fetch' on SHAs instead of references, migrates symbolic links
correctly, and avoids writing to `$HOME/.gitconfig` more than is necessary.

We would like to extend a "thank you" to the following contributors for their
gracious patches:

- @QuLogic fixed an issue with running tests that require credentials
- @patrickmarlier made it possible for 'git lfs migrate import' to work
  correctly with symbolic links.
- @zackse fixed an inconsistency in `CONTRIBUTING.md`
- @zanglang fixed an inconsistency in `README.md`

Git LFS would not be possible without generous contributions from the
open-source community. For these, and many more: thank you!

### Features

* script/packagecloud.rb: release on Ubuntu Bionic #2961 (@ttaylorr)

### Bugs

* lfsapi: canonicalize extra HTTP headers #3010 (@ttaylorr)
* commands/lock: follow symlinks before locking #2996 (@ttaylorr)
* lfs/attribute.go: remove default value from upgradeables #2994 (@ttaylorr)
* git: include SHA1 in ref-less revisions #2982 (@ttaylorr)
* Do not migrate the symlinks to LFS objects. #2983 (@patrickmarlier)
* commands/uninstall: do not log about global hooks with --local #2976 (@ttaylorr)
* commands/run.go: exit 127 on unknown sub-command #2969 (@ttaylorr)
* commands/{un,}track: perform "prefix-agnostic" comparisons #2955 (@ttaylorr)
* commands/migrate: escape paths before .gitattributes  #2933 (@ttaylorr)
* commands/ls-files: do not accept '--all' after '--' #2932 (@ttaylorr)
* tq: prevent uint64 underflow with invalid API response #2902 (@ttaylorr)

### Misc

* test/test-env: skip comparing GIT_EXEC_PATH #3015 (@ttaylorr)
* remove reference to CLA from contributor's guide #2997 (@zackse)
* .gitattributes link is broken #2985 (@zanglang)
* commands: make --version a synonym for 'version' #2968, #3017 (@ttaylorr)
* test: ensure that git-mergetool(1) works with large files #2939 (@ttaylorr)
* README.md: note the correct PackageCloud URL #2960 (@ttaylorr)
* README.md: mention note about `git lfs track` retroactively #2948 (@ttaylorr)
* README.md: reorganize into Core Team, Alumni #2941 (@ttaylorr)
* README.md: :nail_care: #2942 (@ttaylorr)
* circle.yml: upgrade to 'version: 2' syntax #2928 (@ttaylorr)
* Use unique repo name for tests that require credentials. #2901 (@QuLogic)

## 2.4.0 (2 March, 2018)

This release introduces a rewrite of the underlying file matching engine,
expands the API to include relevant refspecs for individual requests,
standardizes the progress output among commands, and more.

Please note: in the next MAJOR release (v3.0.0) the semantic meaning behind
`--include` and `--exclude` flags will change. As the details of exactly which
existing patterns will no longer function as previously are known, we will
indicate them here. Any `--include` or `--exclude` patterns used in v2.3.0 or
earlier are expected to work as previously in this release.

This release would not be possible without the open-source community.
Specifically, we would like to thank:

- @larsxschneider: for contributing fixes to the filter operation in `git lfs
  fsck`, and `git lfs prune`, as well as the bug report leading to the
  filepathfilter changes.
- @yfronto: for adding new Linux release targets.
- @stffabi: for adding support for NTLM with SSPI on Windows.
- @jeffreydwalter: for fixing memory alignment issues with `sync/atomic` on
  32-bit architectures.
- @b4mboo: for adding a LFS configuration key to the list of safe configuration
  options.

Without the aforementioned indviduals, this release would not have been
possible. Thank you!

### Features

* __Support wildmatch-compliant options in `--include`, `--exclude`__
  * filepathfilter: implement using wildmatch #2875 (@ttaylorr)
  * test: add wildmatch migration tests #2888 (@larsxschneider, @ttaylorr)
* __Expand the specification to include relevant refspecs__
  * verify locks against each ref being pushed #2706 (@technoweenie)
  * Batch send refspec take 2 #2809 (@technoweenie)
  * Run 1 TransferQueue per uploaded ref #2806 (@technoweenie)
  * Locks/verify: full refspec #2722 (@technoweenie)
  * send remote refspec for the other lock commands #2773 (@technoweenie)
* __Standardize progress meter output and implementation__
  * tq: standardized progress meter formatting #2811 (@ttaylorr)
  * commands/fetch: unify formatting #2758 (@ttaylorr)
  * commands/prune: unify formatting #2757 (@ttaylorr)
  * progress: use git/githistory/log package for formatting #2732 (@ttaylorr)
  * progress: remove `*progress.Meter` #2762 (@ttaylorr)
  * tasklog: teach `*Logger` how to enqueue new `*SimpleTask`'s #2767 (@ttaylorr)
  * progress: remove spinner.go #2759 (@ttaylorr)
* __Teach new flags, functionality to `git lfs ls-files`__
  * commands: teach '--all' to `git lfs ls-files` #2796 (@ttaylorr)
  * commands/ls-files: show cached, tree-less LFS objects #2795 (@ttaylorr)
  * commands/ls-files: add --include, --exclude #2793 (@ttaylorr)
  * commands/ls-files: add '--size' flag #2764 (@ttaylorr)
* __Add new flags, functionality to `git lfs migrate`__
  * commands/migrate: support '^'-prefix refspec in arguments #2785 (@ttaylorr)
  * commands/migrate: add '--skip-fetch' for offline migrations #2738 (@ttaylorr)
  * git: prefer sending revisions over STDIN than arguments #2739 (@ttaylorr)
* __Release to new operating systems__
  * release lfs for ubuntu/artful too #2704 (@technoweenie)
  * Adding Mint Sylvia to packagecloud.rb script #2829 (@yfronto)
* __New functionality in package `lfsapi`__
  * NTLM authentication with SSPI on windows #2871 (@stffabi)
  * lfsapi/auth: teach DoWithAuth to respect http.extraHeaders #2733 (@ttaylorr)
  * add support for url-specific proxies #2651 (@technoweenie)
* __Code cleanup in git.Config, package `localstorage`__
  * Tracked remote #2700 (@technoweenie)
  * Replace git.Config #2692 (@technoweenie)
  * Replace localstorage #2689 (@technoweenie)
  * Remove last global config #2687 (@technoweenie)
  * Git config refactor #2676 (@technoweenie)

### Bugs

* all: fix 32-bit alignment issues with `sync/atomic` #2883 (@ttaylorr)
* all: memory alignment issues on 32-bit systems. #2880 (@jeffreydwalter)
* command/migrate: don't migrate remote references in bare repositories #2769 (@ttaylorr)
* commands/ls-files: behave correctly before initial commit #2794 (@ttaylorr)
* commands/migrate: allow for ambiguous references in migrations #2734 (@ttaylorr)
* commands: fill in missing printf arg #2678 (@technoweenie)
* config: Add `lfs.locksverify` to safe keys. #2797 (@b4mboo)
* don't replace pointers with objects if clean filter is not configured #2626 (@technoweenie)
* fsck: attach a filter to exclude unfetched items from fsck #2847 (@larsxschneider)
* git/githistory: copy entries from cache, elsewhere #2884 (@ttaylorr)
* git/githistory: migrate annotated tags correctly #2780 (@ttaylorr)
* git/odb: don't print extra newline after commit message #2784 (@ttaylorr)
* git/odb: extract identifiers from commits verbatim #2751 (@wsprent)
* git/odb: implement parsing for annotated `*Tag`'s #2778 (@ttaylorr)
* git/odb: retain newlines when parsing commit messages #2786 (@ttaylorr)
* lfs: PointerScanner is nil after error, so don't close #2699 (@technoweenie)
* lfsapi: Cred helper improvements #2695 (@technoweenie)
* lfsapi: retry requests changing access from none IF Auth header is empty #2621 (@technoweenie)
* prune: always prune excluded paths #2851 (@larsxschneider)
* status: fix incorrect formatting with unpushed objects #2746 (@ttaylorr)
* tasklog: don't drop updates in PercentageTask #2755 (@ttaylorr)
* test: Fix integration test early exit #2735 (@technoweenie)
* test: generate random repo names with fs-safe characters #2698 (@technoweenie)

### Misc

* all: Nitpicks #2821 (@technoweenie)
* all: introduce package 'tlog' #2747 (@ttaylorr)
* all: remove CLA #2870 (@MikeMcQuaid)
* build: Specify the embedded Windows icon as part of versioninfo.json #2770 (@sschuberth)
* config,test: Testlib no global config #2709 (@mathstuf)
* config: add PushRemote() for checking `branch.*.pushRemote` and `remote.pushDefault` first #2715 (@technoweenie)
* docs: Added documentation for git-lfs-ls-files' `*/-` output. #2719 (@bilke)
* docs: Uninstall man page improvements #2730 (@dpursehouse)
* docs: Update usage info for post-checkout #2830 (@proinsias)
* docs: add 'git lfs prune' to main man page #2849 (@larsxschneider)
* docs: use consistent casing for Git #2850 (@larsxschneider)
* git/githistory: have `*RefUpdater` hold `*odb.ObjectDatabase` reference #2779 (@ttaylorr)
* progress: move CopyCallback (& related) to package 'tools' #2749 (@ttaylorr)
* progress: move `*progressLogger` implementation to package 'tools' #2750 (@ttaylorr)
* refspec docs #2820 (@technoweenie)
* script/test: run 'go tool vet' during testing #2788 (@ttaylorr)
* tasklog: introduce `*SimpleTask` #2756 (@ttaylorr)
* test: Ignore comment attr lines #2708 (@mathstuf)
* test: Wait longer for test lfs server to start. #2716 (@QuLogic)
* test: ensure commented attr lines are ignored #2736 (@ttaylorr)
* tools/humanize: add 'FormatByteRate' to format transfer speed #2810 (@ttaylorr)
* vendor: update 'xeipuuv/gojsonpointer' #2846 (@ttaylorr)

## 2.3.4 (18 October, 2017)

### Features

* 'git lfs install' updates filters with 'skip-smudge' option #2673 (@technoweenie)

### Bugs

* FastWalkGitRepo: limit number of concurrent goroutines #2672 (@technoweenie)
* handle scenario where multiple configuration values exist in ~/.gitconfig #2659 (@shiftkey)

## 2.3.3 (9 October, 2017)

### Bugs

* invoke lfs for 'git update-index', fixing 'status' issues #2647 (@technoweenie)
* cache http credential helper output by default #2648 (@technoweenie)

## 2.3.2 (3 October, 2017)

### Features

* bump default activity timeout from 10s -> 30s #2632 (@technoweenie)

### Bugs

* ensure files are marked readonly after unlocking by ID #2642 (@technoweenie)
* add files to index with path relative to current dir #2641 (@technoweenie)
* better Netrc errors #2633 (@technoweenie)
* only use askpass if credential.helper is not configured #2637 (@technoweenie)
* convert backslash to slash when writing to .gitattributes #2625 (@technoweenie)

### Misc

* only copy req headers if there are git-configured extra headers #2622 (@technoweenie)
* update tracerx to add timestamps #2620 (@rubyist)

## 2.3.1 (27 September, 2017)

### Features

* add support for SSH_ASKPASS #2609 (@technoweenie)
* `git lfs migrate --verbose` option #2610 (@technoweenie)
* Support standalone custom transfer based on API URL prefix match #2590 (@sprohaska)

### Bugs

* Improve invalid URL error messages #2614 (@technoweenie)
* Fix double counting progress bug #2608 (@technoweenie)
* trim whitespace from GIT_ASKPASS provided passwords #2607 (@technoweenie)
* remove mmap usage in Packfile reader #2600 (@technoweenie)
* `git lfs clone`: don't fetch for unborn repositories #2598 (@shiftkey)

### Misc

* Windows Installer fixes:
  * Show proper icon in add/remove programs list #2585 (@shiftkey)
  * Make the Inno Setup installer script explicitly check for the binaries #2588 (@sschuberth)
  * Improve compile-win-installer-unsigned.bat a bit #2586 (@sschuberth)
* Update migrate docs example for multiple file types #2596 (@technoweenie)

## 2.3.0 (14 September, 2017)

Git LFS v2.3.0 includes performance optimizations for the `git-lfs-migrate(1)`
and `git-clone(1)` commands, new features, bug-fixes, and more.

This release was made possible by contributors to Git LFS. Specifically:

- @aleb: added support for "standalone" transfer agents, for using `rsync(1)`
  and similar with Git LFS.
- @bozaro: added support for custom `.git/lfs/objects` directories via the
  `lfs.storage` configuration option.
- @larsxschneider: fixed a recursive process leak when shelling out to Git,
  added new features to `git lfs ls-files`, extra information in error
  messages used for debugging, documentation changes and more.
- @mathstuf: contributed a documentation change clarifying LFS's handling of
  empty pointer files.
- @rudineirk and @andyneff: updated our release process to build packages for
  fedora/26.
- @ssgelm: ensured that LFS is able to be released on Ubuntu Universe.

To everyone who has contributed to this or previous releases of Git LFS: Thank
you!

### Features

* git/odb/pack: improve `git lfs migrate` performance
  * git/odb/pack: introduce packed object reassembly #2550 #2551 #2552 #2553 #2554 (@ttaylorr)
  * git/odb/pack: teach packfile index entry lookups #2420 #2421 #2422 #2423 #2437 #2441 #2461 (@ttaylorr)
  * git/{odb,githistory}: don't write unchanged objects #2541 (@ttaylorr)
* commands: improve `git clone` performance with 'delay' capability #2511 #2469 #2468 #2471 #2467 #2476 #2483 (@ttaylorr)
  * commands: mark `git lfs clone` as deprecated #2526 (@ttaylorr)
* commands: enable `lfs.allowincompletepush` by default #2574 (@technoweenie)
* commands: teach '--everything' to `git lfs migrate` #2558 (@ttaylorr)
* commands: teach `git lfs ls-files` a '--debug' option #2540 (@larsxschneider)
* commands,lfs: warn on 4gb size conversion during clean #2510 #2507 #2459 (@ttaylorr)
* lfsapi/creds: teach about GIT_ASKPASS and core.askpass #2500 #2578 (@ttaylorr)
* commands/status: indicate missing objects #2438 (@ttaylorr)
* Allow using custom transfer agents directly #2429 (@aleb)
* Add `lfs.storage` parameter for overriding LFS storage location #2023 (@bozaro)
* lfsapi: enable credential caching by default #2508 (@ttaylorr)
* commands/install: teach `--manual` to `git-lfs-install(1)` #2410 (@ttaylorr)

### Bugs

* migrate: fix migrations with subdirectories in '--include' or '--exclude' #2485 (@ttaylorr)
* commands/migrate: fix hardlinking issue when different filesystem is mounted at `/tmp` #2566 (@ttaylorr)
* commands: make `git lfs migrate` fetch ref updates before migrating #2538 (@ttaylorr)
* commands: remove '--above=1mb' default from `git lfs migrate info` #2460 (@ttaylorr)
* filepathfilter: fix `HasPrefix()` when no '--include' filters present #2579 (@technoweenie)
* git/githistory/log: fix race condition with `git/githistory/log` tests #2495 (@ttaylorr)
* git/odb: fix closing object database test #2457 (@ttaylorr)
* git/githistory: only update local refs after migrations #2559 (@ttaylorr)
* locking: fix unlocking files not removing write flag #2514 (@ttaylorr)
* locks: fix unlocking files in a symlinked directory #2505 (@ttaylorr)
* commands: teach `git lfs unlock` to ignore status errs in appropriate conditions #2475 (@ttaylorr)
* git: expand `GetAttributePaths` check to include non-LFS lockables #2528 (@ttaylorr)
* fix multiple `git update-index` invocations #2531 (@larsxschneider)
* tools: fix SSH credential cacher expiration #2530 (@ttaylorr)
* lfsapi: fix read/write race condition in credential cacher #2493 (@ttaylorr)
* lfs: fix cleaning contents larger than 1024 bytes over stdin #2488 (@ttaylorr)
* fsck only scans current version of objects #2049 (@TheJare)
* progress: fix writing updates to `$GIT_LFS_PROGRESS` #2465 (@ttaylorr)
* commands/track: resolve symlinks before comparing attr paths #2463 (@ttaylorr)
* test: ensure that empty pointers are empty #2458 (@ttaylorr)
* git/githistory/log: prevent 'NaN' showing up in `*PercentageTask` #2455 (@ttaylorr)
* tq: teach Batch() API to retry itself after io.EOF's #2516 (@ttaylorr)

### Misc

* script/packagecloud: release LFS on Fedora/26 #2443 #2509 (@rudineirk, @andyneff)
* git/githistory: change "Rewriting commits" when not updating refs #2577 (@ttaylorr)
* commands: print IP addresses in error logs #2570 (@larsxschneider)
* commands: print current time in UTC to error logs #2571 (@larsxschneider)
* commands: Disable lock verification when using a standalone custom-tr… #2499 (@aleb)
* docs/man: update `git lfs migrate` documentation with EXAMPLES #2580 (@technoweenie)
* docs/man: recommend global per-host locking config #2546 (@larsxschneider)
* commands: use transfer queue's batch size instead of constant #2529 (@ttaylorr)
* add function to invoke Git with disabled LFS filters #2453 (@larsxschneider)
* config: warn on unsafe keys in `.lfsconfig` #2502 (@ttaylorr)
* glide: remove unused dependencies #2501 (@ttaylorr)
* script/build: pass '-{ld,gc}flags' to compiler, if given #2462 (@ttaylorr)
* spec: mention that an empty file is its own LFS pointer #2449 (@mathstuf)
* Update to latest version of github.com/pkg/errors #2426 (@ssgelm)
* Update gitignore to add some temp files that get created when building debs #2425 (@ssgelm)
* lfs: indent contents of `git lfs install`, `update` #2392 (@ttaylorr)
* tq: increase default `lfs.concurrenttransfers` to 8 #2506 (@ttaylorr)

## 2.2.1 (10 July, 2017)

### Bugs

* git lfs status --json only includes lfs files #2374 (@asottile)
* git/odb: remove temporary files after migration #2388 (@ttaylorr)
* git/githistory: fix hanging on empty set of commits #2383 (@ttaylorr)
* migrate: don't checkout HEAD on bare repositories #2389 (@ttaylorr)
* git/odb: prevent cross-volume link error when saving objects #2382 (@ttaylorr)
* commands: only pass --jobs to `git clone` if set #2369 (@technoweenie)

### Misc

* lfs: trace hook install, uninstall, upgrade #2393 (@ttaylorr)
* vendor: remove github.com/cheggaaa/pb #2386 (@ttaylorr)
* Use FormatBytes from git-lfs/tools/humanize instead of cheggaaa/pb #2377 (@ssgelm)


## 2.2.0 (27 June, 2017)

Git LFS v2.2.0 includes bug fixes, minor features, and a brand new `migrate`
command. The `migrate` command rewrites commits, converting large files from
Git blobs to LFS objects. The most common use case will fix a git push rejected
for having large blobs:

```
$ git push origin master
# ...
remote: error: file a.psd is 1.2 gb; this exceeds github's file size limit of 100.00 mb
to github.com:ttaylorr/demo.git
 ! [remote rejected] master -> master (pre-receive hook declined)
error: failed to push some refs to 'git@github.com:ttaylorr/demo.git'

$ git lfs migrate info
*.psd   1.2 GB   27/27 files(s)  100%

$ git lfs migrate import --include="*.psd"
migrate: Sorting commits: ..., done
migrate: Rewriting commits: 100% (810/810), done
  master        f18bb746d44e8ea5065fc779bb1acdf3cdae7ed8 -> 35b0fe0a7bf3ae6952ec9584895a7fb6ebcd498b
migrate: Updating refs: ..., done

$ git push origin
Git LFS: (1 of 1 files) 1.2 GB / 1.2 GB
# ...
To github.com:ttaylorr/demo.git
 * [new branch]      master -> master
```

The `migrate` command has detailed options described in the `git-lfs-migrate(1)`
man page. Keep in mind that this is the first pass at such a command, so we
expect there to be bugs and performance issues (especially on long git histories).
Future updates to the command will be focused on improvements to allow full
LFS transitions on large repositories.

### Features

* commands: add git-lfs-migrate(1) 'import' subcommand #2353 (@ttaylorr)
* commands: add git-lfs-migrate(1) 'info' subcommand #2313 (@ttaylorr)
* Implement status --json #2311 (@asottile)
* commands/uploader: allow incomplete pushes #2199 (@ttaylorr)

### Bugs

* Retry on timeout or temporary errors #2312 (@jakub-m)
* commands/uploader: don't verify locks if verification is disabled #2278 (@ttaylorr)
* Fix tools.TranslateCygwinPath() on MSYS #2277 (@raleksandar)
* commands/clone: add new flags since Git 2.9 #2251, #2252 (@ttaylorr)
* Make pull return non-zero error code when some downloads failed #2237 (@seth2810)
* tq/basic_download: guard against nil HTTP response #2227 (@ttaylorr)
* Bugfix: cannot push to scp style URL #2198 (@jiangxin)
* support lfs.<url>.* values where url does not include .git #2192 (@technoweenie)
* commands: fix logged error not interpolating format qualifiers #2228 (@ttaylorr)
* commands/help: print helptext to stdout for consistency with Git #2210 (@ttaylorr)

### Misc

* Minor cleanups in help index #2248 (@dpursehouse)
* Add git-lfs-lock and git-lfs-unlock to help index #2232 (@dpursehouse)
* packagecloud: add Debian 9 entry to formatted list #2211 (@ttaylorr)
* Update Xenial is to use stretch packages #2212 (@andyneff)

## 2.1.1 (19 May, 2017)

Git LFS v2.1.1 ships with bug fixes and a security patch fixing a remote code
execution vulnerability exploitable by setting a SSH remote via your
repository's `.lfsconfig` to contain the string "-oProxyCommand". This
vulnerability is only exploitable if an attacker has write access to your
repository, or you clone a repository with a `.lfsconfig` file containing that
string.

### Bugs

* Make pull return non-zero error code when some downloads failed #2245 (@seth2810, @technoweenie)
* lfsapi: support cross-scheme redirection #2243 (@ttaylorr)
* sanitize ssh options parsed from ssh:// url #2242 (@technoweenie)
* filepathfilter: interpret as .gitignore syntax #2238 (@technoweenie)
* tq/basic_download: guard against nil HTTP response #2229 (@ttaylorr)
* commands: fix logged error not interpolating format qualifiers #2230 (@ttaylorr)

### Misc

* release: backport Debian 9-related changes #2244 (@ssgelm, @andyneff, @ttaylorr)
* Add git-lfs-lock and git-lfs-unlock to help index #2240 (@dpursehouse)
* config: allow multiple environments when calling config.Unmarshal #2224 (@ttaylorr)

## 2.1.0 (28 April, 2017)

### Features

* commands/track: teach --no-modify-attrs #2175 (@ttaylorr)
* commands/status: add blob info to each entry #2070 (@ttaylorr)
* lfsapi: improve HTTP request/response stats #2184 (@technoweenie)
* all: support URL-style configuration lookups (@ttaylorr)
  * commands: support URL-style lookups for `lfs.{url}.locksverify` #2162 (@ttaylorr)
  * lfsapi: support URL-style lookups for `lfs.{url}.access` #2161 (@ttaylorr)
  * lfsapi/certs: use `*config.URLConfig` to do per-host config lookup #2160 (@ttaylorr)
  * lfsapi: support for http.<url>.extraHeader #2159 (@ttaylorr)
  * config: add prefix to URLConfig type #2158 (@ttaylorr)
  * config: remove dependency on lfsapi package #2156 (@ttaylorr)
  * config: support multi-value lookup on URLConfig #2154 (@ttaylorr)
  * lfsapi: initial httpconfig type #1912 (@technoweenie, @ttaylorr)
* lfsapi,tq: relative expiration support #2130 (@ttaylorr)

### Bugs

* commands: include error in `LoggedError()` #2179 (@ttaylorr)
* commands: cross-platform log formatting to files #2178 (@ttaylorr)
* locks: cross-platform path normalization #2139 (@ttaylorr)
* commands,locking: don't disable locking for auth errors during verify #2110 (@ttaylorr)
* commands/status: show partially staged files twice #2067 (@ttaylorr)

### Misc

* all: build on Go 1.8.1 #2145 (@ttaylorr)
* Polish custom-transfers.md #2171 (@sprohaska)
* commands/push: Fix typo in comment #2170 (@sprohaska)
* config: support multi-valued config entries #2152 (@ttaylorr)
* smudge: use localstorage temp directory, not system #2140 (@ttaylorr)
* locking: send locks limit to server #2107 (@ttaylorr)
* lfs: extract `DiffIndexScanner` #2035 (@ttaylorr)
* status: use DiffIndexScanner to populate results #2042 (@ttaylorr)

## 2.0.2 (29 March, 2017)

### Features

* ssh auth and credential helper caching #2094 (@ttaylorr)
* commands,tq: specialized logging for missing/corrupt objects #2085 (@ttaylorr)
* commands/clone: install repo-level hooks after `git lfs clone` #2074
* (@ttaylorr)
* debian: Support building on armhf and arm64 #2089 (@p12tic)

### Bugs

* commands,locking: don't disable locking for auth errors during verify #2111
* (@ttaylorr)
* commands: show real error while cleaning #2096 (@ttaylorr)
* lfsapi/auth: optionally prepend an empty scheme to Git remote URLs #2092
* (@ttaylorr)
* tq/verify: authenticate verify requests if required #2084 (@ttaylorr)
* commands/{,un}track: correctly escape '#' and ' ' characters #2079 (@ttaylorr)
* tq: use initialized lfsapi.Client instances in transfer adapters #2048
* (@ttaylorr)

### Misc

* locking: send locks limit to server #2109 (@ttaylorr)
* docs: update configuration documentation #2097 #2019 #2102 (@terrorobe)
* docs: update locking API documentation #2099 #2101 (@dpursehouse)
* fixed table markdown in README.md #2095 (@ZaninAndrea)
* remove the the duplicate work #2098 (@grimreaper)

## 2.0.1 (6 March, 2017)

### Misc

* tq: fallback to `_links` if present #2007 (@ttaylorr)

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
