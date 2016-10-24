* a2fab78 2016-10-25 Allow usage of proxies even when contacting localhost (chalstrick[32m (HEAD, chris/fixProxy)[m)
* cbf91a9 2016-10-24 release: v1.4.4 (Taylor Blau[32m (tag: v1.4.4, origin/master, origin/HEAD)[m)
*   8fe768a 2016-10-24 Merge pull request #1603 from github/descriptive-expired-errs (Taylor Blau[32m[m)
[32m|[m[33m\[m  
[32m|[m * 207ad4b 2016-10-24 test: remove "push with continually expired actions" case (Taylor Blau[32m[m)
[32m|[m * eb32041 2016-10-24 transfer: simplify the "object has expired" error message (Taylor Blau[32m[m)
[32m|[m * 1d729a0 2016-10-24 transfer: rename `tt` to `transferTime` (Taylor Blau[32m[m)
[32m|[m * 5e1b97d 2016-10-24 transfer: be more clear about the meaning of `objectExpirationToTransfer` (Taylor Blau[32m[m)
[32m|[m * 67d389e 2016-10-21 test/push: assert on expired_at error message (Taylor Blau[32m (origin/descriptive-expired-errs)[m)
[32m|[m * ca4821d 2016-10-21 api,transfer: teach Object to return when it expired (Taylor Blau[32m[m)
[32m|[m * 8beb41a 2016-10-21 test/gitserver: teach "return-expired-action-forever" (Taylor Blau[32m[m)
[32m|[m[32m/[m  
*   798998b 2016-10-21 Merge pull request #1600 from github/duplicate-oids (Taylor Blau[32m[m)
[34m|[m[35m\[m  
[34m|[m * 0195476 2016-10-20 test: add a test for pushing multiple revs with the same OID (Taylor Blau[32m[m)
[34m|[m * 04c71a3 2016-10-20 test/helpers: teach `pointer` how to optionally include custom versions (Taylor Blau[32m[m)
[34m|[m * d770dfa 2016-10-20 commands/uploader: temporarily use a `tools.StringSet` to avoid duplicate OIDs (Taylor Blau[32m[m)
[34m|[m * 047fe19 2016-10-20 lfs/tq: disallow calling Add() mutliple times for the same OID (Taylor Blau[32m[m)
* [35m|[m   c31d281 2016-10-20 Merge pull request #1599 from github/ssh-result-message (Sebastian Schuberth[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m|[m eb855ba 2016-10-20 Expose the result message in case of an SSH authentication error (Sebastian Schuberth[32m[m)
[36m|[m[36m/[m [35m/[m  
* [35m|[m   3a6f452 2016-10-18 Merge pull request #1594 from github/appveyor_skip_branch_with_pr (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [35m\[m  
[1;32m|[m * [35m|[m 66e2859 2016-10-18 AppVeyor: Do not build branches with open pull requests (Sebastian Schuberth[32m[m)
[1;32m|[m [35m|[m[35m/[m  
* [35m|[m   da7533c 2016-10-18 Merge pull request #1593 from dpursehouse/update-mailmap (Sebastian Schuberth[32m[m)
[35m|[m[1;35m\[m [35m\[m  
[35m|[m [1;35m|[m[35m/[m  
[35m|[m[35m/[m[1;35m|[m   
[35m|[m * 4ae5af6 2016-10-18 Update .mailmap (David Pursehouse[32m[m)
[35m|[m[35m/[m  
*   57234cb 2016-10-17 Merge pull request #1591 from github/release-next (Taylor Blau[32m (tag: v1.4.3, origin/release-1.4)[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 966a9b3 2016-10-17 release: v1.4.3 (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m  
*   76ad7e8 2016-10-17 Merge pull request #1585 from github/storage-retries (Taylor Blau[32m[m)
[32m|[m[33m\[m  
[32m|[m * af1ed6f 2016-10-15 test: batch storage download with retries (Taylor Blau[32m[m)
[32m|[m * 7ac6188 2016-10-15 test: batch storage upload with retries (Taylor Blau[32m[m)
[32m|[m * 6ece33a 2016-10-15 test: legacy storage download with retries (Taylor Blau[32m[m)
[32m|[m * 5f73bf4 2016-10-15 test: legacy storage upload with retries (Taylor Blau[32m[m)
[32m|[m * 864beb8 2016-10-15 test/cmd: teach `storageHandler` how to perform object retries (Taylor Blau[32m[m)
[32m|[m * 4172318 2016-10-15 test/cmd: teach `incrementRetriesFor` which api is being used (Taylor Blau[32m[m)
[32m|[m[32m/[m  
*   018ae22 2016-10-15 Merge pull request #1584 from github/legacy-retries-test (Taylor Blau[32m[m)
[34m|[m[35m\[m  
[34m|[m * 0779577 2016-10-14 test: use `cat` as smudge filter during clone (Taylor Blau[32m[m)
[34m|[m * eb6c4a3 2016-10-14 lfs/tq,test: use Transferable's OID when performing legacy check (Taylor Blau[32m[m)
[34m|[m * 057040c 2016-10-14 test/testhelpers: ignore retry count when checking for OID on server (Taylor Blau[32m[m)
[34m|[m * 0dd6d09 2016-10-14 test/gitserver: include options for legacy check retries (Taylor Blau[32m[m)
* [35m|[m   8d7d552 2016-10-14 Merge pull request #1583 from github/extra-trace-args (Taylor Blau[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m|[m 3e7b872 2016-10-14 lfs/tq: use extra arguments given to tracerx.Printf (Taylor Blau[32m[m)
[36m|[m [35m|[m[35m/[m  
* [35m|[m   7978727 2016-10-14 Merge pull request #1582 from github/legacy-warning-extra (Taylor Blau[32m[m)
[35m|[m[1;33m\[m [35m\[m  
[35m|[m [1;33m|[m[35m/[m  
[35m|[m[35m/[m[1;33m|[m   
[35m|[m * bd0ca46 2016-10-14 api: use fmt.Fprintln to write to os.Stderr (Taylor Blau[32m[m)
[35m|[m[35m/[m  
*   cb80b9d 2016-10-14 Merge pull request #1579 from sschuberth/master (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 98894c4 2016-10-14 docs: Fix a link to the legacy API (Sebastian Schuberth[32m[m)
[1;34m|[m[1;34m/[m  
*   cd68f26 2016-10-13 Merge pull request #1577 from sschuberth/mailmap-new-pr (Taylor Blau[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 1fddc27 2016-10-13 Add a .mailmap file (Sebastian Schuberth[32m[m)
* [31m|[m   db0c919 2016-10-12 Merge pull request #1575 from sschuberth/win-installer-large-wizard-image (Taylor Blau[32m[m)
[31m|[m[33m\[m [31m\[m  
[31m|[m [33m|[m[31m/[m  
[31m|[m[31m/[m[33m|[m   
[31m|[m * 7f7bc70 2016-10-12 Also add a (large) wizard image to the Windows installer (Sebastian Schuberth[32m[m)
[31m|[m * 5ef3a10 2016-10-12 Convert the logo from 16 bits per pixel to 24 bits per pixel (Sebastian Schuberth[32m[m)
[31m|[m[31m/[m  
*   d6a206f 2016-10-12 Merge pull request #1574 from github/appveyor-badge (Taylor Blau[32m[m)
[34m|[m[35m\[m  
[34m|[m * 4b51765 2016-10-11 README: add AppVeyor badge (Taylor Blau[32m[m)
[34m|[m * e7528a6 2016-10-11 README: move Travis badge to newline (Taylor Blau[32m[m)
[34m|[m[34m/[m  
*   6557f9f 2016-10-10 Merge pull request #1569 from github/release-next (Taylor Blau[32m (tag: v1.4.2)[m)
[36m|[m[1;31m\[m  
[36m|[m * c581839 2016-10-10 release: v1.4.2 (Taylor Blau[32m[m)
[36m|[m[36m/[m  
*   5bba96f 2016-10-10 Merge pull request #1571 from github/send-before-wait (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * 6c0b937 2016-10-10 lfs/tq: send error before terminating object (Taylor Blau[32m[m)
[1;32m|[m[1;32m/[m  
*   1328216 2016-10-10 Merge pull request #1570 from github/document-retry-config (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * b339670 2016-10-10 docs/man: document `lfs.transfer.maxretries` (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m  
*   72dc602 2016-10-10 Merge pull request #1568 from github/test-on-1.7 (Taylor Blau[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 72d460d 2016-10-10 travis: run tests on Go 1.7.1 (Taylor Blau[32m[m)
* [31m|[m   b632d4e 2016-10-10 Merge pull request #1567 from sschuberth/windows-sha256sum-appveyor-tests (Taylor Blau[32m[m)
[32m|[m[33m\[m [31m\[m  
[32m|[m * [31m|[m 9e8acb4 2016-10-10 Enable running tests on AppVeyor CI (Sebastian Schuberth[32m[m)
[32m|[m[32m/[m [31m/[m  
* [31m|[m   bcc584c 2016-10-10 Merge pull request #1566 from sschuberth/windows-sha256sum (Taylor Blau[32m[m)
[34m|[m[35m\[m [31m\[m  
[34m|[m * [31m|[m ef802b8 2016-10-10 test: Use "sha256sum" on Windows (Sebastian Schuberth[32m[m)
[34m|[m * [31m|[m c6a9f50 2016-10-09 test: Introduce the calc_oid_file() helper function (Sebastian Schuberth[32m[m)
[34m|[m * [31m|[m d81df7c 2016-10-08 test: Treat MSYS(2) shells also as Windows (Sebastian Schuberth[32m[m)
[34m|[m * [31m|[m 208c0b6 2016-10-08 test: Make IS_{MAC,WINDOWS} checks integer comparisons (Sebastian Schuberth[32m[m)
[34m|[m [31m|[m[31m/[m  
* [31m|[m   b9d2ad3 2016-10-10 Merge pull request #1564 from github/warn-legacy (Taylor Blau[32m[m)
[31m|[m[1;31m\[m [31m\[m  
[31m|[m [1;31m|[m[31m/[m  
[31m|[m[31m/[m[1;31m|[m   
[31m|[m * 657adc8 2016-10-07 lfs: warn on usage of the legacy API (Taylor Blau[32m[m)
* [1;31m|[m   74a10e1 2016-10-07 Merge pull request #1565 from github/windows/fix-fetch-include-exclude (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m  
[1;32m|[m * [1;31m|[m f7536b1 2016-10-07 remove print msg (Rick Olson[32m[m)
[1;32m|[m * [1;31m|[m 67cce7c 2016-10-07 use filepath.Clean() when comparing filenames to include/exclude patterns (Rick Olson[32m[m)
[1;32m|[m [1;31m|[m[1;31m/[m  
* [1;31m|[m   49f3b97 2016-10-07 Merge pull request #1563 from github/track-in-subdirs (Taylor Blau[32m[m)
[1;31m|[m[1;35m\[m [1;31m\[m  
[1;31m|[m [1;35m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m   
[1;31m|[m * c2e6b87 2016-10-07 unused (Rick Olson[32m[m)
[1;31m|[m * b071182 2016-10-07 trim bogus path expansion from git for windows (Rick Olson[32m[m)
[1;31m|[m * 6c771f8 2016-10-07 git: ignore non-root wildcards (Taylor Blau[32m[m)
* [1;35m|[m   f2ce643 2016-10-06 Merge pull request #1557 from sschuberth/brew-upgrade-git (Taylor Blau[32m[m)
[1;35m|[m[31m\[m [1;35m\[m  
[1;35m|[m [31m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[31m|[m   
[1;35m|[m * 3878ebe 2016-10-05 Travis: Only install git if not installed yet (Sebastian Schuberth[32m[m)
[1;35m|[m[1;35m/[m  
*   9600b7c 2016-10-04 Merge pull request #1522 from sschuberth/windows-test-framework-fixes (Taylor Blau[32m[m)
[32m|[m[33m\[m  
[32m|[m * 314a691 2016-09-28 Add a missing testtools tag to git-credential-lfsnoop (Sebastian Schuberth[32m[m)
[32m|[m * 3c8e679 2016-09-28 Make test commands have an ".exe" extension on Windows (Sebastian Schuberth[32m[m)
[32m|[m * e06cd46 2016-09-28 Make building for all OSes include Windows (Sebastian Schuberth[32m[m)
[32m|[m * b43930c 2016-09-28 Make the target OS explicitly default to the host OS (Sebastian Schuberth[32m[m)
[32m|[m * d54a294 2016-09-28 Remove system credential helper specific work-arounds in tests (Sebastian Schuberth[32m[m)
[32m|[m * f89f42a 2016-09-28 Actually export GIT_CONFIG_NOSYSTEM (Sebastian Schuberth[32m[m)
[32m|[m * 296c95e 2016-09-28 Fix determining the path to bash on Windows (Sebastian Schuberth[32m[m)
* [33m|[m   e0db9cc 2016-10-03 Merge pull request #1550 from github/uniq-status (Taylor Blau[32m[m)
[34m|[m[35m\[m [33m\[m  
[34m|[m * [33m|[m 7c1b2f0 2016-09-30 test: ensure status works with clashing OIDs (Taylor Blau[32m[m)
[34m|[m * [33m|[m 015b181 2016-09-30 lfs/scanner: teach ScanIndex, indexFileMap to recognize clashing OIDs (Taylor Blau[32m[m)
[34m|[m * [33m|[m 752b88e 2016-09-30 lfs/scanner: only accept unique revs in ScanIndex (Taylor Blau[32m[m)
[34m|[m[34m/[m [33m/[m  
* [33m|[m   d8cab9e 2016-09-29 Merge pull request #1545 from github/retry-counter (Taylor Blau[32m[m)
[33m|[m[1;31m\[m [33m\[m  
[33m|[m [1;31m|[m[33m/[m  
[33m|[m[33m/[m[1;31m|[m   
[33m|[m * b5cba63 2016-09-28 lfs/tq: extract `RetryCounter` (Taylor Blau[32m[m)
[33m|[m[33m/[m  
*   06b4c14 2016-09-27 Merge pull request #1540 from github/status-initial-commit (Taylor Blau[32m (master)[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * eca20ad 2016-09-27 Change ScanIndex to ref & expose git.RefBeforeFirstCommit for clarity (Steve Streeting[32m[m)
[1;32m|[m * da4b9ef 2016-09-26 Add tests for status before first commit (Steve Streeting[32m[m)
[1;32m|[m * 929f239 2016-09-26 Display added files in git lfs status before initial commit (Steve Streeting[32m[m)
[1;32m|[m * 2179490 2016-09-26 Prevent `git lfs status` from panicking pre initial commit (Steve Streeting[32m[m)
[1;32m|[m[1;32m/[m  
*   fae6810 2016-09-22 Merge pull request #1535 from github/multiple-retries (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 4b19449 2016-09-21 lfs/transfer_queue: remove debug stmt (Taylor Blau[32m[m)
[1;34m|[m * 9b1bd04 2016-09-21 lfs/transfer_queue: support multiple retries per object (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m  
*   878165d 2016-09-20 Merge pull request #1528 from github/batcher-truncate (Taylor Blau[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 1408c85 2016-09-20 lfs/batcher: remove all mentions of 'truncation' from docs (Taylor Blau[32m[m)
[1;36m|[m * 49489c0 2016-09-19 lfs/batcher: :nail_care: docs (Taylor Blau[32m[m)
[1;36m|[m * a27d92d 2016-09-19 lfs/batcher: un-buffer the input channel (again) (Taylor Blau[32m[m)
[1;36m|[m * 65109df 2016-09-19 lfs/batcher: rename Truncate() to Flush() (Taylor Blau[32m[m)
[1;36m|[m * a0d0635 2016-09-19 lfs/batcher: correctly Assert() batcher test cases (Taylor Blau[32m[m)
[1;36m|[m * 0e79937 2016-09-19 lfs/batcher: variadic Add() func (Taylor Blau[32m[m)
[1;36m|[m * 6868390 2016-09-19 lfs/batcher: update `Batcher` documentation (Taylor Blau[32m[m)
[1;36m|[m * a0ba279 2016-09-19 lfs/batcher: use zero-value initialization for exit bool (Taylor Blau[32m[m)
[1;36m|[m * 79e42f3 2016-09-19 lfs/batcher: rename label `Loop` to `Acc` (Taylor Blau[32m[m)
[1;36m|[m * 653a52e 2016-09-19 lfs/batcher: support batch Truncation (Taylor Blau[32m[m)
[1;36m|[m * 98873d8 2016-09-19 lfs/batcher: un-buffer the input channel (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m  
*   b7b1f2b 2016-09-13 Merge pull request #1523 from sschuberth/windows-10-path-comparison (Taylor Blau[32m[m)
[32m|[m[33m\[m  
[32m|[m * b285125 2016-09-13 Make path comparison robust against Windows short / long path issues (Sebastian Schuberth[32m[m)
[32m|[m[32m/[m  
*   59eda37 2016-09-12 Merge pull request #1518 from sschuberth/git-abs-toplevel (Taylor Blau[32m[m)
[34m|[m[35m\[m  
[34m|[m * 5ae1694 2016-09-11 Simplify getting the absolute Git root directory (Sebastian Schuberth[32m[m)
[34m|[m[34m/[m  
*   411305b 2016-09-06 Merge pull request #1505 from github/report-errs (Taylor Blau[32m[m)
[36m|[m[1;31m\[m  
[36m|[m *   9aad5b5 2016-09-06 Merge branch 'master' into report-errs (Taylor Blau[32m[m)
[36m|[m [1;32m|[m[36m\[m  
[36m|[m [1;32m|[m[36m/[m  
[36m|[m[36m/[m[1;32m|[m   
* [1;32m|[m   fa56460 2016-09-06 Merge pull request #1503 from github/clean-plumbing (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m  
[1;34m|[m * [1;32m|[m 6d25314 2016-09-05 docs/man: reference git-lfs-pointer(1) in clean documentation (Taylor Blau[32m[m)
* [1;35m|[m [1;32m|[m   2c5fd74 2016-09-06 Merge pull request #1504 from sschuberth/master (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;32m\[m  
[1;36m|[m * [1;35m|[m [1;32m|[m fc2c9e8 2016-09-06 Add icons to the Windows installer (Sebastian Schuberth[32m[m)
[1;36m|[m[1;36m/[m [1;35m/[m [1;32m/[m  
* [1;35m|[m [1;32m|[m   f54b90f 2016-09-06 Merge pull request #1506 from sschuberth/appveyor (Taylor Blau[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;32m\[m  
[32m|[m * [1;35m|[m [1;32m|[m 63dac50 2016-09-06 Make the AppVeyor CI build work again (Sebastian Schuberth[32m[m)
[32m|[m * [1;35m|[m [1;32m|[m 7261b71 2016-09-06 Do not run gofmt in the src directory (Sebastian Schuberth[32m[m)
[32m|[m [1;35m|[m[1;35m/[m [1;32m/[m  
* [1;35m|[m [1;32m|[m   9713ed8 2016-09-06 Merge pull request #1507 from sschuberth/fetch-without-remote (Taylor Blau[32m[m)
[1;35m|[m[35m\[m [1;35m\[m [1;32m\[m  
[1;35m|[m [35m|[m[1;35m/[m [1;32m/[m  
[1;35m|[m[1;35m/[m[35m|[m [1;32m|[m   
[1;35m|[m * [1;32m|[m 9b9dde4 2016-09-06 Allow fetch to run without a remote configured (Sebastian Schuberth[32m[m)
[1;35m|[m[1;35m/[m [1;32m/[m  
[1;35m|[m * 0a40ef2 2016-09-06 commands: clarify LoggedError documentation (Taylor Blau[32m[m)
[1;35m|[m * 7ccc6b5 2016-09-05 test/push: assert message is logged in a /storage 503 (Taylor Blau[32m[m)
[1;35m|[m * 50ba7eb 2016-09-05 test: add optional $msg parameter to `push_fail_test` (Taylor Blau[32m[m)
[1;35m|[m * 1dabeb5 2016-09-05 commands: make `FullError` print the first line of debugged or fatal errors (Taylor Blau[32m[m)
[1;35m|[m[1;35m/[m  
*   a06acca 2016-09-01 Merge pull request #1495 from github/register-commands-v2 (risk danger olson[32m[m)
[36m|[m[1;31m\[m  
[36m|[m *   5cfe2ad 2016-09-01 Merge branch 'master' into register-commands-v2 (risk danger olson[32m[m)
[36m|[m [1;32m|[m[36m\[m  
[36m|[m [1;32m|[m[36m/[m  
[36m|[m[36m/[m[1;32m|[m   
* [1;32m|[m 9c1aae0 2016-09-01 roadmap: check off upgradable filters (Taylor Blau[32m[m)
* [1;32m|[m   5edb8a7 2016-09-01 Merge pull request #1497 from github/upgrade-old-filters (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m  
[1;34m|[m * [1;32m|[m 87727a2 2016-09-01 Automatically upgrade old filters instead of requiring —force (Steve Streeting[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m  
[1;34m|[m * 60639b0 2016-09-01 put all the command stuff in run.go (risk danger olson[32m[m)
[1;34m|[m * a12cdc8 2016-09-01 RegisterCommand() can no longer disable commands (risk danger olson[32m[m)
[1;34m|[m * 0b0d983 2016-09-01 document NewCommand and RegisterCommand (risk danger olson[32m[m)
[1;34m|[m * e8bf1bd 2016-09-01 move the command init to a separate file (risk danger olson[32m[m)
[1;34m|[m * 093e6d4 2016-09-01 update all command setup to use RegisterCommand (risk danger olson[32m[m)
[1;34m|[m * 9dc64b4 2016-08-31 env is simple enough it doesn't need a callback (risk danger olson[32m[m)
[1;34m|[m * fc4214e 2016-08-31 restore PreRun for env command, update clone command (risk danger olson[32m[m)
[1;34m|[m * a2b1cc5 2016-08-31 commands: try out RegisterCommand() (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m  
*   df4923c 2016-08-26 Merge pull request #1489 from github/allow-lfs.pushurl (risk danger olson[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 8d9f798 2016-08-26 add blank line to appease 'go vet' (risk danger olson[32m[m)
[1;36m|[m * 07bbdda 2016-08-26 add lfs.pushurl to the .lfsconfig safelist (risk danger olson[32m[m)
[1;36m|[m[1;36m/[m  
* 2070e4c 2016-08-26 release: Git LFS v1.4.1 (Taylor Blau[32m (tag: v1.4.1)[m)
*   020b0df 2016-08-26 Merge pull request #1482 from github/no-rewrap (Taylor Blau[32m[m)
[32m|[m[33m\[m  
[32m|[m * ce191e9 2016-08-24 tools/io: fix spelling error (Taylor Blau[32m[m)
[32m|[m * 27cde58 2016-08-24 tools/test: remove '.' imports, increase readability (Taylor Blau[32m[m)
[32m|[m * a9bbe0e 2016-08-24 tools/test: test behavior of RetriableReader (Taylor Blau[32m[m)
[32m|[m * 3406688 2016-08-24 tools: ensure Retriable errors are wrapped only once (Taylor Blau[32m[m)
[32m|[m[32m/[m  
*   89f37f4 2016-08-24 Merge pull request #1454 from larsxschneider/retry-eof (Taylor Blau[32m[m)
[34m|[m[35m\[m  
[34m|[m * a941805 2016-08-24 errors: use github.com/pkg/errors instead of removed errutil package (Lars Schneider[32m[m)
[34m|[m * 8e477ca 2016-08-24 retry if file download failed (Lars Schneider[32m[m)
[34m|[m * 08a780d 2016-08-24 add RetriableReader (Lars Schneider[32m[m)
[34m|[m[34m/[m  
*   4d516f0 2016-08-23 Merge pull request #1478 from github/clone-in-current-directory (Taylor Blau[32m[m)
[36m|[m[1;31m\[m  
[36m|[m * 365d615 2016-08-22 test/clone: refute that the "lfs" directory is created at the top level (Taylor Blau[32m[m)
[36m|[m * 74b4d2c 2016-08-22 test/clone,util: ensure cloning is possible in current directory "." (Taylor Blau[32m[m)
[36m|[m * 0f63130 2016-08-22 commands, lfs: resolve localstorage in PreRun, not init (Taylor Blau[32m[m)
[36m|[m[36m/[m  
* 68c0e18 2016-08-22 Add HTTP 507 (Insufficient Storage) to the list of optional statuses (#1473) (David Pursehouse[32m[m)
* 9f1f93f 2016-08-19 look at me fixing debian errors all by myself (risk danger olson[32m (tag: v1.4.0)[m)
*   ddfaae3 2016-08-19 Merge pull request #1466 from github/errutil-to-errors (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * 831bcc4 2016-08-19 errors: remove Error prefix (risk danger olson[32m[m)
[1;32m|[m * 8624a87 2016-08-19 fix stacktrace logging (risk danger olson[32m[m)
[1;32m|[m *   b42fa38 2016-08-19 Merge branch 'master' into errutil-to-errors (risk danger olson[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m   
* [1;34m|[m e178b59 2016-08-19 release: Git LFS 1.4.0 (Taylor Blau[32m[m)
* [1;34m|[m   ebbe14f 2016-08-19 Merge pull request #1438 from github/triaged-roadmap-updates (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m  
[1;36m|[m * [1;34m|[m 1eda382 2016-08-18 more updates from this morning's triage run. (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m 922d85c 2016-08-17 more updates from triaged issues (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m 77533e1 2016-08-16 ROADMAP: move migration tool to "possible features" section (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m 65368bb 2016-08-12 Update ROADMAP.md (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m 9e23f73 2016-08-12 Update ROADMAP.md (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m 7c0c7ac 2016-08-12 ROADMAP: add LFS Migration tool (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m be318a6 2016-08-12 ROADMAP: add supporting GIT_CONFIG (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m baef8dd 2016-08-11 add ssh shorthands (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m c419e63 2016-08-11 Update ROADMAP.md (risk danger olson[32m[m)
* [31m|[m [1;34m|[m   5ea16db 2016-08-19 Merge pull request #1469 from github/progress-docs (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;34m\[m  
[32m|[m * [31m\[m [1;34m\[m   2f44bc6 2016-08-19 Merge branch 'master' into progress-docs (risk danger olson[32m[m)
[32m|[m [34m|[m[32m\[m [31m\[m [1;34m\[m  
[32m|[m [34m|[m[32m/[m [31m/[m [1;34m/[m  
[32m|[m[32m/[m[34m|[m [31m|[m [1;34m|[m   
* [34m|[m [31m|[m [1;34m|[m   704a8ee 2016-08-19 Merge pull request #1470 from github/fetch-seen-bug (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [31m\[m [1;34m\[m  
[36m|[m * [34m|[m [31m|[m [1;34m|[m 177984f 2016-08-19 don't report "seen" objects as ready to checkout (risk danger olson[32m[m)
[36m|[m[36m/[m [34m/[m [31m/[m [1;34m/[m  
[36m|[m * [31m|[m [1;34m|[m 378b0f1 2016-08-19 docs/man: run spellcheck (Taylor Blau[32m[m)
[36m|[m * [31m|[m [1;34m|[m 2d17873 2016-08-19 docs/man: note GIT_LFS_PROGRESS (Taylor Blau[32m[m)
[36m|[m[36m/[m [31m/[m [1;34m/[m  
[36m|[m [31m|[m *   22d73e5 2016-08-19 Merge branch 'master' into errutil-to-errors (Taylor Blau[32m[m)
[36m|[m [31m|[m [1;32m|[m[36m\[m  
[36m|[m [31m|[m[36m_[m[1;32m|[m[36m/[m  
[36m|[m[36m/[m[31m|[m [1;32m|[m   
* [31m|[m [1;32m|[m   6ff5882 2016-08-19 Merge pull request #1467 from dpursehouse/tidy-up-509-description (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [1;32m\[m  
[1;34m|[m * [31m|[m [1;32m|[m 4e31b03 2016-08-19 Reword the description of HTTP 509 status to make it consistent with the others (David Pursehouse[32m[m)
[1;34m|[m[1;34m/[m [31m/[m [1;32m/[m  
[1;34m|[m [31m|[m * b5a1c07 2016-08-18 errors: split into multiple go files (risk danger olson[32m[m)
[1;34m|[m [31m|[m * 8ecfd22 2016-08-18 errors: remove unused Stack() func (risk danger olson[32m[m)
[1;34m|[m [31m|[m * c999200 2016-08-18 errors: implement fmt.Formatter (risk danger olson[32m[m)
[1;34m|[m [31m|[m * a3a4e5b 2016-08-18 errors: remove GetInnerError() and ErrorWithStack (risk danger olson[32m[m)
[1;34m|[m [31m|[m * 69acb28 2016-08-18 lean into pkg/errors philosophy on wrappin' (risk danger olson[32m[m)
[1;34m|[m [31m|[m * 311cc7a 2016-08-18 api,auth,errors,httputil,lfs,transfer: wrap errors (Taylor Blau[32m[m)
[1;34m|[m [31m|[m * 8fd3774 2016-08-18 errors: Errorf -> Wrapf (Taylor Blau[32m[m)
[1;34m|[m [31m|[m * 2473068 2016-08-18 rename errutil to errors (risk danger olson[32m[m)
[1;34m|[m [31m|[m[1;34m/[m  
[1;34m|[m[1;34m/[m[31m|[m   
* [31m|[m   74a228f 2016-08-18 Merge pull request #1463 from github/errors-next (risk danger olson[32m[m)
[1;36m|[m[31m\[m [31m\[m  
[1;36m|[m * [31m\[m   ab517b6 2016-08-18 Merge branch 'master' into errors-next (risk danger olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [31m\[m  
[1;36m|[m [32m|[m[1;36m/[m [31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [31m|[m   
* [32m|[m [31m|[m   29e1497 2016-08-18 Merge pull request #1458 from github/include-dupe-file (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [31m\[m  
[34m|[m * [32m\[m [31m\[m   f210c3d 2016-08-18 Merge branch 'master' into include-dupe-file (risk danger olson[32m[m)
[34m|[m [36m|[m[34m\[m [32m\[m [31m\[m  
[34m|[m [36m|[m[34m/[m [32m/[m [31m/[m  
[34m|[m[34m/[m[36m|[m [32m|[m [31m|[m   
* [36m|[m [32m|[m [31m|[m   ac5a532 2016-08-18 Merge pull request #1461 from github/check-git-version (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [32m\[m [31m\[m  
[1;32m|[m * [36m\[m [32m\[m [31m\[m   57a5bc8 2016-08-18 Merge branch 'master' into check-git-version (risk danger olson[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [36m\[m [32m\[m [31m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [36m/[m [32m/[m [31m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [36m|[m [32m|[m [31m|[m   
[1;32m|[m * [36m|[m [32m|[m [31m|[m 18f863f 2016-08-17 only re-run Version() if they don't match (risk danger olson[32m[m)
[1;32m|[m * [36m|[m [32m|[m [31m|[m 3d5b99d 2016-08-17 don't ignore a 'git version' error (risk danger olson[32m[m)
[1;32m|[m * [36m|[m [32m|[m [31m|[m a343a11 2016-08-17 check the git version is ok in some key commands (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m 9243c6b 2016-08-17 filter objects in one step (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m cbcb864 2016-08-17 don't send dupe OIDs to the transferqueue (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m 79d7ea6 2016-08-17 revert last change (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m 60fd28b 2016-08-17 transferqueue: ignore added objects with non-unique OIDs (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m 1485571 2016-08-17 use `git ls-tree` to find lfs objects for `lfs fetch` (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m * [32m|[m [31m|[m 76f6a7a 2016-08-16 failing test (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;35m|[m * [31m|[m c5036f6 2016-08-18 fix pointer error messages (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;35m|[m * [31m|[m 162490d 2016-08-18 fix error message tests (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;35m|[m * [31m|[m 4b83600 2016-08-18 remove debug msg (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;35m|[m * [31m|[m 952701d 2016-08-18 errorWrapper interface includes Cause() now (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;35m|[m * [31m|[m c5de285 2016-08-18 errors: use github.com/pkg/errors in errutil package (Taylor Blau[32m[m)
[1;32m|[m [1;35m|[m[1;32m_[m[1;35m|[m[1;32m/[m [31m/[m  
[1;32m|[m[1;32m/[m[1;35m|[m [1;35m|[m [31m|[m   
* [1;35m|[m [1;35m|[m [31m|[m   aaf933a 2016-08-17 Merge pull request #1459 from github/vendor-errs (Taylor Blau[32m[m)
[1;35m|[m[31m\[m [1;35m\[m [1;35m\[m [31m\[m  
[1;35m|[m [31m|[m[1;35m/[m [1;35m/[m [31m/[m  
[1;35m|[m[1;35m/[m[31m|[m [1;35m|[m [31m|[m   
[1;35m|[m * [1;35m|[m [31m|[m 09a2c43 2016-08-17 script/test: skip github.com/pkg/errors (Taylor Blau[32m[m)
[1;35m|[m * [1;35m|[m [31m|[m 46c17a5 2016-08-17 vendor: add github.com/pkg/errs (Taylor Blau[32m[m)
* [31m|[m [1;35m|[m [31m|[m   36b2757 2016-08-17 Merge pull request #1460 from github/javabrett-config-system (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [31m\[m  
[32m|[m * [31m\[m [1;35m\[m [31m\[m   9dd99c8 2016-08-17 fix merge conflicts (risk danger olson[32m[m)
[32m|[m [32m|[m[35m\[m [31m\[m [1;35m\[m [31m\[m  
[32m|[m[32m/[m [35m/[m [31m/[m [1;35m/[m [31m/[m  
[32m|[m * [31m|[m [1;35m|[m [31m|[m df22c0a 2016-06-30 Improved error-reporting in ResolveRef. This is required when SimpleExec stops swallowing the error return. Related test "ls-files: with zero files". (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m d06e97b 2016-06-30 Removed Go 1.6-specific API changes introduced in the previous commit.  This commit can be reverted as soon as Go >= 1.6 is required. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m 7773420 2016-06-30 Stop SimpleExec swallowing errors. Fixed #1183. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m 648d8c7 2016-06-30 Added --system option to install command.  This option is invoked from post-install scripts in rpm/deb packages. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m 82f860a 2016-06-30 Added rpm %post (git lfs install) and %prerm (git lfs uninstall) scripts. These will run as root and so will operate at the new --system scope. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m 7c4db57 2016-06-30 Added postinst (git lfs install) and prerm (git lfs uninstall) scripts to Debian deb. These will run as root and so will operate at the new --system scope. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m fb1e26c 2016-06-30 Config attribute operations now auto-elevate to system scope from global if running as the root user (unless --local is specified), using the new System-level config commands. (Brett Randall[32m[m)
[32m|[m * [31m|[m [1;35m|[m [31m|[m 45712c9 2016-06-30 Added System-scoped (git config --system) versions of Find/Set/Unset/UnsetSection commands. These commands operate on config stored at system scope e.g. /etc/gitconfig. (Brett Randall[32m[m)
* [35m|[m [31m|[m [1;35m|[m [31m|[m   757234a 2016-08-17 Merge pull request #1453 from github/dpursehouse-issue-1356 (Taylor Blau[32m[m)
[1;35m|[m[1;31m\[m [35m\[m [31m\[m [1;35m\[m [31m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[35m|[m[1;35m_[m[31m|[m[1;35m/[m [31m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [35m|[m [31m|[m [31m|[m   
[1;35m|[m * [35m|[m [31m|[m [31m|[m   74a4cb3 2016-08-17 Merge branch 'master' into dpursehouse-issue-1356 (Taylor Blau[32m[m)
[1;35m|[m [1;32m|[m[1;35m\[m [35m\[m [31m\[m [31m\[m  
[1;35m|[m [1;32m|[m[1;35m/[m [35m/[m [31m/[m [31m/[m  
[1;35m|[m[1;35m/[m[1;32m|[m [35m|[m [31m|[m [31m|[m   
* [1;32m|[m [35m|[m [31m|[m [31m|[m   32b9177 2016-08-16 Merge pull request #1455 from ralfthewise/include-exclude-docs (risk danger olson[32m[m)
[1;34m|[m[31m\[m [1;32m\[m [35m\[m [31m\[m [31m\[m  
[1;34m|[m [31m|[m [1;32m|[m[31m_[m[35m|[m[31m/[m [31m/[m  
[1;34m|[m [31m|[m[31m/[m[1;32m|[m [35m|[m [31m|[m   
[1;34m|[m * [1;32m|[m [35m|[m [31m|[m 7926b50 2016-08-16 update fetch include/exclude docs for pattern matching (Tim Garton[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m [35m/[m [31m/[m  
* [1;32m|[m [35m|[m [31m|[m   9c17eb1 2016-08-16 Merge pull request #1451 from github/fetch-remote-urls (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;32m\[m [35m\[m [31m\[m  
[1;36m|[m * [1;32m|[m [35m|[m [31m|[m 8421b5c 2016-08-16 use grep, not ack (risk danger olson[32m[m)
[1;36m|[m * [1;32m|[m [35m|[m [31m|[m   e6dcd3f 2016-08-16 Merge branch 'master' into fetch-remote-urls (risk danger olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [1;32m\[m [35m\[m [31m\[m  
[1;36m|[m [32m|[m[1;36m/[m [1;32m/[m [35m/[m [31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [1;32m|[m [35m|[m [31m|[m   
* [32m|[m [1;32m|[m [35m|[m [31m|[m   fb63bc4 2016-08-16 Merge pull request #1452 from github/allow-skippable-auth (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [1;32m\[m [35m\[m [31m\[m  
[34m|[m * [32m\[m [1;32m\[m [35m\[m [31m\[m   c24e0c5 2016-08-16 Merge branch 'master' into allow-skippable-auth (risk danger olson[32m[m)
[34m|[m [36m|[m[34m\[m [32m\[m [1;32m\[m [35m\[m [31m\[m  
[34m|[m [36m|[m[34m/[m [32m/[m [1;32m/[m [35m/[m [31m/[m  
[34m|[m[34m/[m[36m|[m [32m|[m [1;32m|[m [35m|[m [31m|[m   
[34m|[m * [32m|[m [1;32m|[m [35m|[m [31m|[m   b71f0ef 2016-08-16 Merge branch 'master' into allow-skippable-auth (risk danger olson[32m[m)
[34m|[m [1;32m|[m[1;33m\[m [32m\[m [1;32m\[m [35m\[m [31m\[m  
[34m|[m * [1;33m|[m [32m|[m [1;32m|[m [35m|[m [31m|[m 6b3e1c4 2016-08-15 add object Authenticated property to let servers tell the client to skip the git-credentials check (risk danger olson[32m[m)
[34m|[m [1;33m|[m [1;33m|[m * [1;32m|[m [35m|[m [31m|[m 8c6ed75 2016-08-15 handle trailing slashes from remotes (risk danger olson[32m[m)
[34m|[m [1;33m|[m [1;33m|[m * [1;32m|[m [35m|[m [31m|[m 33134e0 2016-08-15 add happy path tests for push/pull/fetch (risk danger olson[32m[m)
[34m|[m [1;33m|[m [1;33m|[m * [1;32m|[m [35m|[m [31m|[m 7bd194d 2016-08-15 Accept raw remote URLs as valid (epriestley[32m[m)
[34m|[m [1;33m|[m [1;33m|[m[1;33m/[m [1;32m/[m [35m/[m [31m/[m  
[34m|[m [1;33m|[m[1;33m/[m[1;33m|[m [1;32m|[m [35m|[m [31m|[m   
[34m|[m [1;33m|[m [1;33m|[m * [35m|[m [31m|[m 3345723 2016-08-16 clean up duplicate error handling code (risk danger olson[32m[m)
[34m|[m [1;33m|[m [1;33m|[m * [35m|[m [31m|[m   86e6418 2016-08-16 Merge branch 'issue-1356' of https://github.com/dpursehouse/git-lfs into dpursehouse-issue-1356 (risk danger olson[32m[m)
[34m|[m [1;33m|[m [1;33m|[m [34m|[m[1;35m\[m [35m\[m [31m\[m  
[34m|[m [1;33m|[m[34m_[m[1;33m|[m[34m/[m [1;35m/[m [35m/[m [31m/[m  
[34m|[m[34m/[m[1;33m|[m [1;33m|[m [1;35m|[m [35m|[m [31m|[m   
[34m|[m [1;33m|[m [1;33m|[m * [35m|[m [31m|[m d382f1f 2016-08-17 Fixes #1356: Don't show same error message twice (David Pursehouse[32m[m)
[34m|[m [1;33m|[m [1;33m|[m[1;33m/[m [35m/[m [31m/[m  
[34m|[m [1;33m|[m[1;33m/[m[1;33m|[m [35m|[m [31m|[m   
* [1;33m|[m [1;33m|[m [35m|[m [31m|[m 2ccfb7d 2016-08-16 ROADMAP: updates for 1.4.0-pre (Taylor Blau[32m[m)
[1;33m|[m [1;33m|[m[1;33m/[m [35m/[m [31m/[m  
[1;33m|[m[1;33m/[m[1;33m|[m [35m|[m [31m|[m   
* [1;33m|[m [35m|[m [31m|[m   8d3bbfe 2016-08-16 Merge pull request #1450 from github/config-next-wrapped-git (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;33m\[m [35m\[m [31m\[m  
[1;36m|[m * [1;33m\[m [35m\[m [31m\[m   691bbbe 2016-08-16 fix merge conflicts (risk danger olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [1;33m\[m [35m\[m [31m\[m  
[1;36m|[m [32m|[m[1;36m/[m [1;33m/[m [35m/[m [31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [1;33m|[m [35m|[m [31m|[m   
* [32m|[m [1;33m|[m [35m|[m [31m|[m   22c80c2 2016-08-16 Merge pull request #1436 from github/config-next-remove-legacy (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [1;33m\[m [35m\[m [31m\[m  
[34m|[m * [32m\[m [1;33m\[m [35m\[m [31m\[m   44cd61a 2016-08-16 Merge branch 'master' into config-next-remove-legacy (risk danger olson[32m[m)
[34m|[m [36m|[m[34m\[m [32m\[m [1;33m\[m [35m\[m [31m\[m  
[34m|[m [36m|[m[34m/[m [32m/[m [1;33m/[m [35m/[m [31m/[m  
[34m|[m[34m/[m[36m|[m [32m|[m [1;33m|[m [35m|[m [31m|[m   
* [36m|[m [32m|[m [1;33m|[m [35m|[m [31m|[m   fb124de 2016-08-16 Merge pull request #1441 from github/lowercase-event-key (risk danger olson[32m[m)
[1;33m|[m[1;33m\[m [36m\[m [32m\[m [1;33m\[m [35m\[m [31m\[m  
[1;33m|[m [1;33m|[m[1;33m_[m[36m|[m[1;33m_[m[32m|[m[1;33m/[m [35m/[m [31m/[m  
[1;33m|[m[1;33m/[m[1;33m|[m [36m|[m [32m|[m [35m|[m [31m|[m   
[1;33m|[m * [36m|[m [32m|[m [35m|[m [31m|[m   44e7a67 2016-08-15 Merge branch 'master' into lowercase-event-key (Taylor Blau[32m[m)
[1;33m|[m [1;34m|[m[1;33m\[m [36m\[m [32m\[m [35m\[m [31m\[m  
[1;33m|[m [1;34m|[m[1;33m/[m [36m/[m [32m/[m [35m/[m [31m/[m  
[1;33m|[m[1;33m/[m[1;34m|[m [36m|[m [32m|[m [35m|[m [31m|[m   
[1;33m|[m * [36m|[m [32m|[m [35m|[m [31m|[m c1170b7 2016-08-12 transfer/custom: encode "event" as lowercase (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m * [32m|[m [35m|[m [31m|[m 254c11d 2016-08-10 config.etc: remove Getenv utility method (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m * [32m|[m [35m|[m [31m|[m ec50199 2016-08-10 config,etc: remove uses of GetenvBool (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m [1;35m|[m * [35m|[m [31m|[m 14055fc 2016-08-15 config: make GitFetcher lookups case-insensitive (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m [1;35m|[m * [35m|[m [31m|[m d3bc59a 2016-08-15 replace `.GitConfig*` with `.Git.*` (risk danger olson[32m[m)
[1;33m|[m [1;35m|[m [1;35m|[m * [35m|[m [31m|[m 04c2754 2016-08-15 config: introduce `*gitEnvironment` implementation (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m [1;35m|[m * [35m|[m [31m|[m 6c60d2b 2016-08-15 config: demote *Environment to interface (Taylor Blau[32m[m)
[1;33m|[m [1;35m|[m[1;33m_[m[1;35m|[m[1;33m/[m [35m/[m [31m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [1;35m|[m [35m|[m [31m|[m   
* [1;35m|[m [1;35m|[m [35m|[m [31m|[m b546ba4 2016-08-15 ROADMAP: link `config-next` to tracking issue (Taylor Blau[32m[m)
* [1;35m|[m [1;35m|[m [35m|[m [31m|[m   de772a4 2016-08-15 Merge pull request #1443 from github/config-insteadof (risk danger olson[32m[m)
[1;35m|[m[31m\[m [1;35m\[m [1;35m\[m [35m\[m [31m\[m  
[1;35m|[m [31m|[m[1;35m/[m [1;35m/[m [35m/[m [31m/[m  
[1;35m|[m[1;35m/[m[31m|[m [1;35m|[m [35m|[m [31m|[m   
[1;35m|[m * [1;35m|[m [35m|[m [31m|[m 8cff656 2016-08-12 put the insteadof warning in the correct place (risk danger olson[32m[m)
[1;35m|[m * [1;35m|[m [35m|[m [31m|[m db0da8a 2016-08-12 no need to export UrlAliases() (risk danger olson[32m[m)
[1;35m|[m * [1;35m|[m [35m|[m [31m|[m 26f72f5 2016-08-12 handle cases where someone has multiple url.*.insteadof keys with the same alias value (risk danger olson[32m[m)
[1;35m|[m * [1;35m|[m [35m|[m [31m|[m 71fbb41 2016-08-12 add support for url.*.insteadof in git config (risk danger olson[32m[m)
[1;35m|[m[1;35m/[m [1;35m/[m [35m/[m [31m/[m  
* [1;35m|[m [35m|[m [31m|[m   525d07e 2016-08-11 Merge pull request #1434 from github/command-var-scope (risk danger olson[32m[m)
[31m|[m[33m\[m [1;35m\[m [35m\[m [31m\[m  
[31m|[m [33m|[m[31m_[m[1;35m|[m[31m_[m[35m|[m[31m/[m  
[31m|[m[31m/[m[33m|[m [1;35m|[m [35m|[m   
[31m|[m * [1;35m|[m [35m|[m   9aa8781 2016-08-11 Merge branch 'master' into command-var-scope (risk danger olson[32m[m)
[31m|[m [34m|[m[31m\[m [1;35m\[m [35m\[m  
[31m|[m [34m|[m[31m/[m [1;35m/[m [35m/[m  
[31m|[m[31m/[m[34m|[m [1;35m|[m [35m|[m   
* [34m|[m [1;35m|[m [35m|[m   e918500 2016-08-10 Merge pull request #1435 from github/config-next-no-more-setters (Taylor Blau[32m[m)
[1;35m|[m[1;31m\[m [34m\[m [1;35m\[m [35m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[34m|[m[1;35m/[m [35m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [34m|[m [35m|[m   
[1;35m|[m * [34m|[m [35m|[m 31544d0 2016-08-10 config/config: remove `origConfig` field (Taylor Blau[32m[m)
[1;35m|[m * [34m|[m [35m|[m fc23a37 2016-08-10 config/config: remove SetConfig and ResetConfig methods (Taylor Blau[32m[m)
[1;35m|[m[1;35m/[m [34m/[m [35m/[m  
[1;35m|[m * [35m|[m   4fe9f9a 2016-08-10 Merge branch 'master' into command-var-scope (risk danger olson[32m[m)
[1;35m|[m [1;32m|[m[1;35m\[m [35m\[m  
[1;35m|[m [1;32m|[m[1;35m/[m [35m/[m  
[1;35m|[m[1;35m/[m[1;32m|[m [35m|[m   
* [1;32m|[m [35m|[m   68e5a3d 2016-08-10 Merge pull request #1427 from github/config-next-no-git-immut (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [35m\[m  
[1;34m|[m * [1;32m\[m [35m\[m   3c3a0b3 2016-08-10 Merge branch 'master' into config-next-no-git-immut (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;32m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;32m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;32m|[m [35m|[m   
* [1;36m|[m [1;32m|[m [35m|[m   e508fb6 2016-08-10 Merge pull request #1430 from github/transfer-manifest (Taylor Blau[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;32m\[m [35m\[m  
[32m|[m * [1;36m\[m [1;32m\[m [35m\[m   0d4e855 2016-08-10 Merge branch 'master' into transfer-manifest (Taylor Blau[32m[m)
[32m|[m [34m|[m[32m\[m [1;36m\[m [1;32m\[m [35m\[m  
[32m|[m [34m|[m[32m/[m [1;36m/[m [1;32m/[m [35m/[m  
[32m|[m[32m/[m[34m|[m [1;36m|[m [1;32m|[m [35m|[m   
[32m|[m [34m|[m * [1;32m|[m [35m|[m   6cea089 2016-08-10 Merge branch 'master' into config-next-no-git-immut (Taylor Blau[32m[m)
[32m|[m [34m|[m [36m|[m[32m\[m [1;32m\[m [35m\[m  
[32m|[m [34m|[m[32m_[m[36m|[m[32m/[m [1;32m/[m [35m/[m  
[32m|[m[32m/[m[34m|[m [36m|[m [1;32m|[m [35m|[m   
* [34m|[m [36m|[m [1;32m|[m [35m|[m   42eec79 2016-08-09 Merge pull request #1429 from github/config-next-fetch-unmarshal (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [34m\[m [36m\[m [1;32m\[m [35m\[m  
[1;32m|[m * [34m|[m [36m|[m [1;32m|[m [35m|[m 3723cb2 2016-08-09 config: use config.Unmarshal to fill FetchPruneConfig (Taylor Blau[32m[m)
* [1;33m|[m [34m|[m [36m|[m [1;32m|[m [35m|[m   01cd862 2016-08-09 Merge pull request #1428 from github/config-next-load-unmarshal (Taylor Blau[32m[m)
[1;34m|[m[1;33m\[m [1;33m\[m [34m\[m [36m\[m [1;32m\[m [35m\[m  
[1;34m|[m [1;33m|[m[1;33m/[m [34m/[m [36m/[m [1;32m/[m [35m/[m  
[1;34m|[m * [34m|[m [36m|[m [1;32m|[m [35m|[m 1845c0b 2016-08-09 config: ensure .gitconfig is loaded before Unmarshal call (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m [34m/[m [36m/[m [1;32m/[m [35m/[m  
[1;34m|[m [34m|[m * [1;32m|[m [35m|[m ca28d93 2016-08-10 api,transfer: inject custom `*config.Configuration` throughout (Taylor Blau[32m[m)
[1;34m|[m [34m|[m * [1;32m|[m [35m|[m 075fc1e 2016-08-09 api,lfs: eradicate usage of config.Config, SetEnv (Taylor Blau[32m[m)
[1;34m|[m [34m|[m * [1;32m|[m [35m|[m eafa8e0 2016-08-09 config: remove ClearConfig function (Taylor Blau[32m[m)
[1;34m|[m [34m|[m * [1;32m|[m [35m|[m 0c134d4 2016-08-09 httputil: enforce config immutability (Taylor Blau[32m[m)
[1;34m|[m [34m|[m * [1;32m|[m [35m|[m cddd408 2016-08-09 auth: enforce config immutability (Taylor Blau[32m[m)
[1;34m|[m [34m|[m[1;34m/[m [1;32m/[m [35m/[m  
[1;34m|[m[1;34m/[m[34m|[m [1;32m|[m [35m|[m   
[1;34m|[m [34m|[m * [35m|[m 844a1f4 2016-08-10 remove global *transfer.Manifest from commands pkg (risk danger olson[32m[m)
[1;34m|[m [34m|[m * [35m|[m d3c57fc 2016-08-10 use the global *transfer.Manifest setup in commands.Run() (risk danger olson[32m[m)
[1;34m|[m [34m|[m * [35m|[m 6a8e9d3 2016-08-10 teach lfs.Environ() to accept config and manifest values, instead of pulling from global vars (risk danger olson[32m[m)
[1;34m|[m [34m|[m * [35m|[m 98af594 2016-08-10 build commands outside of package init() (risk danger olson[32m[m)
[1;34m|[m [34m|[m[34m/[m [35m/[m  
[1;34m|[m * [35m|[m 9f3a508 2016-08-10 restore these tests (risk danger olson[32m[m)
[1;34m|[m * [35m|[m 8b07e19 2016-08-10 doc tweaks (risk danger olson[32m[m)
[1;34m|[m * [35m|[m 53130bb 2016-08-09 remove global *transfer.Manifest instance (risk danger olson[32m[m)
[1;34m|[m * [35m|[m f4a4753 2016-08-09 transfer: add a Manifest type that stores the upload and download adapters (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m [35m/[m  
* [35m|[m   bbac36e 2016-08-09 Merge pull request #1426 from github/config-next-unmarshalling (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m|[m f0094eb 2016-08-09 config: Unmarshal using default values (Taylor Blau[32m[m)
[1;36m|[m * [35m|[m 09f7d65 2016-08-09 config: introduce func `Unmarshal(v interface{})` (Taylor Blau[32m[m)
* [31m|[m [35m|[m   ca9797a 2016-08-09 Merge pull request #1420 from github/config-next-git-parsing (Taylor Blau[32m[m)
[32m|[m[33m\[m [31m\[m [35m\[m  
[32m|[m * [31m\[m [35m\[m   b27308b 2016-08-09 Merge pull request #1423 from github/config-next-prune (Taylor Blau[32m[m)
[32m|[m [34m|[m[31m\[m [31m\[m [35m\[m  
[32m|[m [34m|[m [31m|[m[31m/[m [35m/[m  
[32m|[m [34m|[m * [35m|[m 34d5552 2016-08-05 cleanup FetchPruneConfig() (risk danger olson[32m[m)
[32m|[m * [35m|[m [35m|[m   55f86e0 2016-08-09 Merge branch 'master' into config-next-git-parsing (risk danger olson[32m[m)
[32m|[m [35m|[m[32m\[m [35m\[m [35m\[m  
[32m|[m [35m|[m[32m/[m [35m/[m [35m/[m  
[32m|[m[32m/[m[35m|[m [35m/[m [35m/[m   
[32m|[m [35m|[m[35m/[m [35m/[m    
* [35m|[m [35m|[m 758604c 2016-08-09 add socks proxy to the roadmap (risk danger olson[32m[m)
[1;31m|[m * [35m|[m 7bcc69b 2016-08-05 change c.GitConfigInt() and c.GitConfigBool() to use c.Git (risk danger olson[32m[m)
[1;31m|[m * [35m|[m 999fb64 2016-08-05 really really fix it (risk danger olson[32m[m)
[1;31m|[m * [35m|[m 84071fc 2016-08-05 fix FetchExcludePaths typo (risk danger olson[32m[m)
[1;31m|[m * [35m|[m 9c727d2 2016-08-05 change (Environment) Get()'s signature to include ok bool (risk danger olson[32m[m)
[1;31m|[m * [35m|[m 8237c1c 2016-08-05 move git config commands to getGitConfigs() (risk danger olson[32m[m)
[1;31m|[m * [35m|[m ee687d7 2016-08-05 get all tests passing (risk danger olson[32m[m)
[1;31m|[m * [35m|[m e73933e 2016-08-05 config: initial take at GitFetcher implementation (Taylor Blau[32m[m)
[1;31m|[m[1;31m/[m [35m/[m  
* [35m|[m   49bf4d5 2016-08-05 Merge pull request #1419 from github/config-next-environments (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [35m\[m  
[1;32m|[m * [35m|[m 38c452b 2016-08-04 config: remove MockFetcher in tests (Taylor Blau[32m[m)
[1;32m|[m * [35m|[m a0acb31 2016-08-04 config/environment_test: use testify/assert (Taylor Blau[32m[m)
[1;32m|[m * [35m|[m 52b7f01 2016-08-04 config,etc: rename EnvFetcher, config.Env to OsFetcher, config.Os (Taylor Blau[32m (origin/config-next-gitconfig-fetcher)[m)
[1;32m|[m * [35m|[m 18cb925 2016-08-04 config: demote Fetcher, introduce Environment (Taylor Blau[32m[m)
[1;32m|[m[1;32m/[m [35m/[m  
* [35m|[m   eea4565 2016-08-04 Merge pull request #1415 from github/config-next (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [35m\[m  
[1;34m|[m * [35m\[m   136c4cf 2016-08-04 Merge pull request #1416 from github/config-next-no-immut (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m[31m\[m [35m\[m  
[1;34m|[m [1;36m|[m * [35m|[m 316f885 2016-08-03 config: remove Setenv and EnvFetcher.Set (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m * [35m|[m 0d25044 2016-08-03 auth: remove usage of Setenv (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m * [35m|[m 4b2417e 2016-08-03 commands: remove usage of config.Setenv (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m * [35m|[m 7aa4126 2016-08-03 config,fetcher: remove SetAllEnv function (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m * [35m|[m 8ca4fcb 2016-08-03 config: eradicate uses of SetAllEnv (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m[1;36m/[m [35m/[m  
[1;34m|[m * [35m|[m   faa3c2d 2016-08-03 Merge branch 'master' into config-next (Taylor Blau[32m[m)
[1;34m|[m [32m|[m[1;34m\[m [35m\[m  
[1;34m|[m [32m|[m[1;34m/[m [35m/[m  
[1;34m|[m[1;34m/[m[32m|[m [35m|[m   
* [32m|[m [35m|[m   9f2ffe7 2016-08-03 Merge pull request #1412 from github/roadmap-emojis (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [35m\[m  
[34m|[m * [32m|[m [35m|[m 079e4bc 2016-08-02 ROADMAP: convert roadmap to markdown table (Taylor Blau[32m[m)
[34m|[m[34m/[m [32m/[m [35m/[m  
[34m|[m * [35m|[m a1e9bae 2016-08-03 config/config: documentation fixes (Taylor Blau[32m[m)
[34m|[m * [35m|[m b665611 2016-08-03 config/env: fix outdated reference to EnvFetcher.String (Taylor Blau[32m[m)
[34m|[m * [35m|[m f6e8c52 2016-08-03 config: integrate EnvFetcher type into `config.Config` (Taylor Blau[32m[m)
[34m|[m * [35m|[m b2272a8 2016-08-03 config: implement type `EnvFetcher` (Taylor Blau[32m[m)
[34m|[m * [35m|[m e32fabe 2016-08-03 config: introduce Fetcher type (Taylor Blau[32m[m)
[34m|[m[34m/[m [35m/[m  
* [35m|[m ebf7fbd 2016-08-02 this ensures that it doesn't prompt for pwd (risk danger olson[32m[m)
* [35m|[m b15ca36 2016-08-02 run against gitlfs packages from docker hub (risk danger olson[32m[m)
* [35m|[m 9c9dffb 2016-08-02 release v1.3.1 (risk danger olson[32m (tag: v1.3.1, origin/release-1.3)[m)
* [35m|[m 083778a 2016-08-02 dont build windows zips, installer is preferred (risk danger olson[32m[m)
* [35m|[m 2113ddb 2016-08-02 remove roadmap items fixed in v1.3.1 (risk danger olson[32m[m)
* [35m|[m   df8f440 2016-08-02 Merge pull request #1411 from github/empty-include-exclude-paths (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m|[m cf3dc4e 2016-08-02 use shared include/exclude flag values across all commands (risk danger olson[32m[m)
[36m|[m * [35m|[m 0bc2f0e 2016-08-02 CleanPathsDefault() is unused (risk danger olson[32m[m)
[36m|[m * [35m|[m   dee8283 2016-08-02 Merge branch 'master' into empty-include-exclude-paths (risk danger olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [35m\[m  
[36m|[m [1;32m|[m[36m/[m [35m/[m  
[36m|[m[36m/[m[1;32m|[m [35m|[m   
* [1;32m|[m [35m|[m   ac301aa 2016-08-02 Merge pull request #1409 from github/install-custom-hooks-path (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [35m\[m  
[1;34m|[m * [1;32m|[m [35m|[m c8d5320 2016-08-02 test: test core.hooksPath doesn't install to .git/ (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m 47245ef 2016-08-02 test: s/preform/perfom/g (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m 223d26d 2016-08-01 docs/man: update git-lfs-install.1 (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m d56b5aa 2016-08-01 lfs/hook: teach `lfs.Hook` about `core.hooksPath` (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m [35m/[m  
[1;34m|[m * [35m|[m d14c36b 2016-08-02 process include/exclude flags for clone and pull correctly (risk danger olson[32m[m)
[1;34m|[m * [35m|[m ee140dd 2016-08-02 extract larger fetch tests to separate files (risk danger olson[32m[m)
[1;34m|[m * [35m|[m 48822ad 2016-08-01 distinguish between empty include/exclude paths (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m [35m/[m  
* [35m|[m   701cc0c 2016-08-01 Merge pull request #1390 from github/commands-config (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m\[m   f436202 2016-08-01 Merge branch 'master' into commands-config (Taylor Blau[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [35m\[m  
[1;36m|[m [32m|[m[1;36m/[m [35m/[m  
[1;36m|[m[1;36m/[m[32m|[m [35m|[m   
* [32m|[m [35m|[m   4ab36e3 2016-08-01 Merge pull request #1404 from dakotahawkins/fix-sslCAInfo-config-lookup (Taylor Blau[32m[m)
[34m|[m[35m\[m [32m\[m [35m\[m  
[34m|[m * [32m|[m [35m|[m 24cab8a 2016-07-28 Fix #1403: sslCAInfo config lookup when host in config doesn't have a trailing slash (Dakota Hawkins[32m[m)
[34m|[m [35m|[m * [35m|[m   21699b4 2016-08-01 Merge branch 'master' into commands-config (risk danger olson[32m[m)
[34m|[m [35m|[m [36m|[m[34m\[m [35m\[m  
[34m|[m [35m|[m[34m_[m[36m|[m[34m/[m [35m/[m  
[34m|[m[34m/[m[35m|[m [36m|[m [35m|[m   
* [35m|[m [36m|[m [35m|[m 4a05f52 2016-08-01 Add some bugs for v1.3.1 (risk danger olson[32m[m)
[35m|[m[35m/[m [36m/[m [35m/[m  
* [36m|[m [35m|[m   d0028c6 2016-07-28 Merge pull request #1405 from dakotahawkins/add-branch-name-to-contributing-readme (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [35m\[m  
[1;32m|[m * [36m|[m [35m|[m d04aa42 2016-07-28 Added branch name (master) to CONTRIBUTING.md (Dakota Hawkins[32m[m)
[1;32m|[m[1;32m/[m [36m/[m [35m/[m  
* [36m|[m [35m|[m 13f3932 2016-07-27 ninja changelog update for v1.3.0 (risk danger olson[32m (tag: v1.3.0)[m)
* [36m|[m [35m|[m   c0a5b4b 2016-07-27 Merge pull request #1400 from github/update-index-combined-output (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [36m\[m [35m\[m  
[1;34m|[m * [36m|[m [35m|[m a95afd2 2016-07-27 import 'bytes' (risk danger olson[32m[m)
[1;34m|[m * [36m|[m [35m|[m 1e3fca0 2016-07-27 don't wait til after the error to capture stdout/stderr (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m [36m/[m [35m/[m  
[1;34m|[m * [35m|[m   74da40a 2016-07-27 merge conflict (risk danger olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [35m|[m   
* [1;36m|[m [35m|[m   1999143 2016-07-27 Merge pull request #1398 from github/linux-build-fixes (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [35m\[m  
[32m|[m * [1;36m\[m [35m\[m   9233523 2016-07-27 Merge branch 'master' into linux-build-fixes (risk danger olson[32m[m)
[32m|[m [34m|[m[32m\[m [1;36m\[m [35m\[m  
[32m|[m [34m|[m[32m/[m [1;36m/[m [35m/[m  
[32m|[m[32m/[m[34m|[m [1;36m|[m [35m|[m   
* [34m|[m [1;36m|[m [35m|[m   6b259fe 2016-07-27 Merge pull request #1399 from github/custom-transfer-trace-newlines (Taylor Blau[32m[m)
[36m|[m[1;31m\[m [34m\[m [1;36m\[m [35m\[m  
[36m|[m * [34m|[m [1;36m|[m [35m|[m d26a6cf 2016-07-27 Trim custom adapter output before tracing to avoid extra newline (Steve Streeting[32m[m)
[36m|[m[36m/[m [34m/[m [1;36m/[m [35m/[m  
[36m|[m * [1;36m|[m [35m|[m e82ec57 2016-07-27 fix golang test excludes for dh_golang (risk danger olson[32m[m)
[36m|[m * [1;36m|[m [35m|[m 69d38ef 2016-07-27 This directory was not a git repo, so `git clean` kept failing (risk danger olson[32m[m)
[36m|[m * [1;36m|[m [35m|[m 95d6745 2016-07-27 don't build the api test cmd for pkg installs (risk danger olson[32m[m)
[36m|[m * [1;36m|[m [35m|[m 6a4f79a 2016-07-27 less noisy builds (risk danger olson[32m[m)
[36m|[m * [1;36m|[m [35m|[m 1b3e949 2016-07-27 fix gopath issues in rpm spec (risk danger olson[32m[m)
[36m|[m[36m/[m [1;36m/[m [35m/[m  
* [1;36m|[m [35m|[m   6ddee2f 2016-07-27 Merge pull request #1397 from github/env-report-transfers (Steve Streeting[32m[m)
[1;32m|[m[1;33m\[m [1;36m\[m [35m\[m  
[1;32m|[m * [1;36m|[m [35m|[m f534b6a 2016-07-27 Sort transfers in env so results are deterministic (Steve Streeting[32m[m)
[1;32m|[m * [1;36m|[m [35m|[m 6dbe55e 2016-07-27 Fix worktree tests (Steve Streeting[32m[m)
[1;32m|[m * [1;36m|[m [35m|[m b5750bd 2016-07-27 Remove redundant part of test, over generous copy & paste (Steve Streeting[32m[m)
[1;32m|[m * [1;36m|[m [35m|[m be266f2 2016-07-27 Report actual list of transfers that will be tried from `git lfs env` (Steve Streeting[32m[m)
[1;32m|[m[1;32m/[m [1;36m/[m [35m/[m  
* [1;36m|[m [35m|[m   953b638 2016-07-26 Merge pull request #1388 from github/next (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;36m\[m [35m\[m  
[1;34m|[m * [1;36m\[m [35m\[m   d60379d 2016-07-26 Merge branch 'master' into next (risk danger olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;36m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;36m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;36m|[m [35m|[m   
* [1;36m|[m [1;36m|[m [35m|[m   e015041 2016-07-26 Merge pull request #1389 from github/skip-transfer-key (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;36m\[m [35m\[m  
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 33ad44a 2016-07-26 better bash equality check so test is skipped correctly (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 2c78673 2016-07-26 fix worktree tests (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 4ba190d 2016-07-26 Fix tus tests too (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m c4f4587 2016-07-26 update env tests with TusTransfers value (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 1270aac 2016-07-26 document `lfs.tustransfers` (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m def174b 2016-07-25 spit out TusTransfers in 'git lfs env' (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 6a35429 2016-07-25 fix BatchTransfer() (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m ebfe51a 2016-07-25 use GitConfigBool() (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 1b5a60c 2016-07-25 quick note about the status of tus.io uploads (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m c2ae16a 2016-07-25 dark ship the tus adapter (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 97044f9 2016-07-22 dont send any transfers if the client only wants to use 'basic' (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;36m|[m [35m|[m 325c673 2016-07-21 Skip sending the `transfers` key if custom transfer adapters are disabled (risk danger olson[32m[m)
* [33m|[m [1;36m|[m [1;36m|[m [35m|[m d96c553 2016-07-25 roadmap: knock out proxy changes (Taylor Blau[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m   1858267 2016-07-25 Merge pull request #1394 from dakotahawkins/bump-next-version-number (Taylor Blau[32m[m)
[33m|[m [33m|[m [34m|[m[35m\[m [1;36m\[m [35m\[m  
[33m|[m [33m|[m [34m|[m * [1;36m|[m [35m|[m 21e1f58 2016-07-25 Bump version number. (Dakota Hawkins[32m[m)
[33m|[m [33m|[m [34m|[m[34m/[m [1;36m/[m [35m/[m  
[33m|[m [33m|[m * [1;36m|[m [35m|[m c1306a2 2016-07-22 Missed one (Steve Streeting[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m 7f46f2c 2016-07-22 Combine transfer features into an easier to understand section (Steve Streeting[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m 409b5ca 2016-07-21 CHANGELOG: condense lock-related entries (Taylor Blau[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m 9e3eee5 2016-07-21 CHANGELOG: move config include/exclude to "features" section (Taylor Blau[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m 6811066 2016-07-21 CHANGELOG: move Travis change to "misc" section (Taylor Blau[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m afa7742 2016-07-21 CHANGELOG: add @jonmagic and @LizzHale to proxy PR (Taylor Blau[32m[m)
[33m|[m [33m|[m * [1;36m|[m [35m|[m 04f3fbb 2016-07-21 release: git-lfs 1.3.0 (Taylor Blau[32m[m)
[33m|[m [33m|[m[33m/[m [1;36m/[m [35m/[m  
[33m|[m [33m|[m * [35m|[m b9c5a10 2016-07-21 replace config.Config references in httputil (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m 4f71d45 2016-07-21 config.NewConfig() => config.New() (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m fecb92d 2016-07-21 command_update no longer needs the config package (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m 698b24e 2016-07-21 a few more config.Config references in the commands pkg (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m 2aa0569 2016-07-21 remove config.Config references in auth pkg (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m 7457375 2016-07-21 rename commands.Config to commands.cfg (risk danger olson[32m[m)
[33m|[m [33m|[m * [35m|[m ed01a95 2016-07-21 Use commands.Config instead of config.Config (risk danger olson[32m[m)
[33m|[m [33m|[m[33m/[m [35m/[m  
[33m|[m[33m/[m[33m|[m [35m|[m   
* [33m|[m [35m|[m   7baff56 2016-07-21 Merge pull request #1176 from javabrett/patch-1 (risk danger olson[32m[m)
[33m|[m[1;31m\[m [33m\[m [35m\[m  
[33m|[m [1;31m|[m[33m/[m [35m/[m  
[33m|[m[33m/[m[1;31m|[m [35m|[m   
[33m|[m * [35m|[m cb661ce 2016-04-25 Added some building-on-RHEL prerequisites. (Brett Randall[32m[m)
* [1;31m|[m [35m|[m   d61aca9 2016-07-21 Merge pull request #1386 from github/disable-lock-commands (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m  
[1;32m|[m * [1;31m|[m [35m|[m e6b29c7 2016-07-21 command/test: guard `unlock` command behind feature flag, too (Taylor Blau[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m ab9842c 2016-07-21 commands: use config's `GetenvBool`, add tests (Taylor Blau[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m 36b0664 2016-07-21 commands/test: make lock commands opt-in using GITLFSLOCKSENABLED=1 (Taylor Blau[32m[m)
* [1;33m|[m [1;31m|[m [35m|[m   e489002 2016-07-21 Merge pull request #1358 from github/jonmagic-use-proxy-from-git-config (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;33m\[m [1;31m\[m [35m\[m   7931683 2016-07-21 Merge branch 'master' into jonmagic-use-proxy-from-git-config (risk danger olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;33m\[m [1;31m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;33m/[m [1;31m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;33m|[m [1;31m|[m [35m|[m   
* [1;36m|[m [1;33m|[m [1;31m|[m [35m|[m 062dc3a 2016-07-21 more fixes (risk danger olson[32m[m)
* [1;36m|[m [1;33m|[m [1;31m|[m [35m|[m 29e8876 2016-07-21 don't export auth type values (risk danger olson[32m[m)
* [1;36m|[m [1;33m|[m [1;31m|[m [35m|[m   7f49746 2016-07-21 Merge pull request #1200 from teo-tsirpanis/patch-1 (risk danger olson[32m[m)
[1;33m|[m[33m\[m [1;36m\[m [1;33m\[m [1;31m\[m [35m\[m  
[1;33m|[m [33m|[m[1;33m_[m[1;36m|[m[1;33m/[m [1;31m/[m [35m/[m  
[1;33m|[m[1;33m/[m[33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[1;33m|[m * [1;36m|[m [1;31m|[m [35m|[m 97448d3 2016-05-04 Fix bug in Windows installer under Win32. (Theodore Tsirpanis[32m[m)
* [33m|[m [1;36m|[m [1;31m|[m [35m|[m   67aa3b2 2016-07-20 Merge pull request #1384 from andyneff/sarah (Taylor Blau[32m[m)
[34m|[m[35m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[34m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 45ab6db 2016-07-20 Added Linux Mint Sarah to package cloud script (Andy Neff[32m[m)
[34m|[m[34m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
* [33m|[m [1;36m|[m [1;31m|[m [35m|[m   b414b58 2016-07-20 Merge pull request #1367 from github/transfers-p5-ext-process (Steve Streeting[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[36m|[m * [33m\[m [1;36m\[m [1;31m\[m [35m\[m   b5a9533 2016-07-20 Merge branch 'master' into transfers-p5-ext-process (Steve Streeting[32m[m)
[36m|[m [1;32m|[m[36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[36m|[m [1;32m|[m[36m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[36m|[m[36m/[m[1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   407b13f 2016-07-18 Merge pull request #1379 from VladimirKhvostov/master (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 5a639a5 2016-07-18 Made changes based on PR feedback. (Vladimir Khvostov[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 4c6f19b 2016-07-17 Updated request.GetAuthType to handle multi-value auth headers (Vladimir Khvostov[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[1;34m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m   51c3320 2016-07-17 Merge branch 'master' into transfers-p5-ext-process (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   cb85ec0 2016-07-17 Merge pull request #1374 from github/windows-fixes (Taylor Blau[32m[m)
[32m|[m[33m\[m [1;36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[32m|[m * [1;36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m   185307a 2016-07-17 Merge branch 'master' into windows-fixes (Taylor Blau[32m[m)
[32m|[m [34m|[m[32m\[m [1;36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[32m|[m [34m|[m[32m/[m [1;36m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[32m|[m[32m/[m[34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   b7bb728 2016-07-17 Merge pull request #1375 from github/tests/mac-race-condition (Taylor Blau[32m[m)
[36m|[m[1;31m\[m [34m\[m [1;36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 0e87780 2016-07-15 all osx builds are good (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m ed20a71 2016-07-15 GIT_TERMINAL_PROMPT=0 is supported on enough versions of git (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 7a7495d 2016-07-15 only skip the osx build with travis' older git version (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 3a3f0e1 2016-07-15 set default test maxprocs back to 4 (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 519b5f6 2016-07-15 print the maxprocs message right before kicking off the test workers (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 18de3ae 2016-07-15 enable credential.usehttppath by default in tests (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 99af66b 2016-07-15 set default test procs on mac to 1 (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m e2046fd 2016-07-15 add uniq id to test server logs (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m c0a2b89 2016-07-15 unreachable (risk danger olson[32m[m)
[36m|[m * [34m|[m [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m d3c9c72 2016-07-15 teach script/integration.go about GIT_LFS_TEST_MAXPROCS (risk danger olson[32m[m)
[36m|[m[36m/[m [34m/[m [1;36m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m bd84dac 2016-07-15 Use filepath.ToSlash() instead of manually replacing (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 40318dd 2016-07-15 Only replace \ with / on Windows (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m ee5df49 2016-07-15 Windows fix: suppress GUI prompts from OpenSSH during integration tests (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 8e3867a 2016-07-15 Windows fix: xeipuuv/gojsonreference tests are buggy on Windows, exclude (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 0209c51 2016-07-15 Windows fix: path cleaning should use '/' not native path separators (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m bee70c0 2016-07-15 Windows fix: exclude vendor pkgs which are not cross-platform from test (Steve Streeting[32m[m)
[36m|[m * [1;36m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 169420e 2016-07-15 Windows fix: don't use filepath, produces backslash path separators (Steve Streeting[32m[m)
[36m|[m[36m/[m [1;36m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m a90c039 2016-07-15 Make docs for error handling consistent (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m   6a62287 2016-07-15 Merge branch 'master' into transfers-p5-ext-process (Steve Streeting[32m[m)
[36m|[m [1;32m|[m[36m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[36m|[m [1;32m|[m[36m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[36m|[m[36m/[m[1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   d7403ce 2016-07-15 Merge pull request #1373 from github/clone-submodule-improvements (Steve Streeting[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m   9da69ed 2016-07-15 Merge branch 'master' into clone-submodule-improvements (Steve Streeting[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   9398386 2016-07-14 Merge pull request #1372 from larsxschneider/patch-3 (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[32m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m f6766e9 2016-07-14 travis-ci: require successful tests against upcoming Git core release (Lars Schneider[32m[m)
* [33m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   bd795b9 2016-07-14 Merge pull request #1371 from github/handle-artifactory-responses (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[34m|[m * [33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m   2a4cbd7 2016-07-14 Merge branch 'master' into handle-artifactory-responses (risk danger olson[32m[m)
[34m|[m [36m|[m[34m\[m [33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[34m|[m [36m|[m[34m/[m [33m/[m [1;36m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[34m|[m[34m/[m[36m|[m [33m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
* [36m|[m [33m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   671556c 2016-07-14 Merge pull request #1359 from github/rev-list-remotes (risk danger olson[32m[m)
[33m|[m[1;33m\[m [36m\[m [33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[33m|[m [1;33m|[m[33m_[m[36m|[m[33m/[m [1;36m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[1;33m|[m [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m * [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   0916c1e 2016-07-14 Merge branch 'master' into rev-list-remotes (Taylor Blau[32m[m)
[33m|[m [1;34m|[m[33m\[m [36m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[33m|[m [1;34m|[m[33m/[m [36m/[m [1;36m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[1;34m|[m [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m * [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m d4ce48c 2016-07-13 update revListArgsRefVsRemote() docs (risk danger olson[32m[m)
[33m|[m * [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 521fa64 2016-07-13 pass --stdin (risk danger olson[32m[m)
[33m|[m * [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   c3b6a37 2016-07-11 Merge branch 'master' into rev-list-remotes (risk danger olson[32m[m)
[33m|[m [1;36m|[m[31m\[m [36m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[33m|[m * [31m|[m [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 1e74819 2016-07-08 document the return args (risk danger olson[32m[m)
[33m|[m * [31m|[m [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 0073bf9 2016-07-08 hit SAVE (risk danger olson[32m[m)
[33m|[m * [31m|[m [36m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m d2d650f 2016-07-08 use `git rev-list --stdin` instead of passing each remote ref (risk danger olson[32m[m)
[33m|[m [31m|[m [31m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m a2380e5 2016-07-14 test/push: fix test failure on old versions of Git (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   dd9ab98 2016-07-13 Merge branch 'master' into handle-artifactory-responses (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m [32m|[m[33m\[m [1;36m\[m [1;32m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[33m|[m [31m|[m[33m_[m[31m|[m[33m_[m[32m|[m[33m/[m [1;36m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [32m|[m [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m [31m|[m [31m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m eee82e8 2016-07-13 transfer: error on invalid object size from batch response (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 44c5031 2016-07-13 docs/v1: clarify actions and batch response (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m * [1;36m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m e4d51fa 2016-07-13 docs/v{1,1.3}: document minimum allowable value for object size (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m fd2ad1d 2016-07-15 Comment fix (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 04931af 2016-07-14 In git 2.9+, run "git lfs pull" in submodules after "git lfs clone" (Steve Streeting[32m[m)
[33m|[m [31m|[m[33m_[m[31m|[m[33m_[m[33m|[m[33m/[m [1;32m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [33m|[m [1;32m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m c018b62 2016-07-15 Rename "id" property to "event" in custom transfer messages (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m bb5549e 2016-07-15 Handle incorrect messages received from custom adapters (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 20b4d8a 2016-07-15 Reduce nesting (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m d063337 2016-07-15 Don't actually need to declare vars anymore for closure capture (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m cb53d50 2016-07-14 Comment fixes (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 74cb09b 2016-07-14 Remove unnecessary dependency on api package (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 6741e9b 2016-07-14 Don't use localstorage, path for upload is already provided in Transfer (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 51e774f 2016-07-14 Add 30s timeout on shutdown request, then abort (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 416fdb0 2016-07-14 Trim unnecessary err test before return (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 24fb19e 2016-07-14 No need to call Error() on err to get desc (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m   69015df 2016-07-14 Merge branch 'master' into transfers-p5-ext-process (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [34m|[m[33m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[33m|[m [31m|[m[33m_[m[31m|[m[33m_[m[33m|[m[33m_[m[34m|[m[33m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [33m|[m [34m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 68749e9 2016-07-13 Test files are moved into lfs objects correctly (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 623bfaf 2016-07-13 Fix weird clone test failures; must not cause config to load too early (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 5fbc31c 2016-07-13 Test for custom transfer uploading and downloading (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 91e00f6 2016-07-13 Improve tracing (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 412ea31 2016-07-13 Remember to send completion message from upload (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 0b896a6 2016-07-13 Make sure we flush stderr consistently from custom adapter (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 2993352 2016-07-13 Tidy up signatures (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 743e647 2016-07-13 First pass of test custom adapter (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m effc7bf 2016-07-13 Don't ignore source files named lfstest-*.go, only compiled binaries (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 94fb71d 2016-07-13 Introduce a message id to each message to disambiguate more easily (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m e8a4586 2016-07-13 Check SHA and move file after download (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 987d3a3 2016-07-13 Capture stderr output of custom adapter to trace (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 0ba7780 2016-07-13 Fix indexes when reading responses (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m ade1bdf 2016-07-13 Simplify request structs a little & reuse with omitempty instead (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 47d7c41 2016-07-13 Completed first cut of custom transfer protocol implementation (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m abd2ba7 2016-07-13 Close stdin/out on shutdown (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 7252ed8 2016-07-13 Added startup & shutdown, protocol messaging utilities (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m aceca9e 2016-07-13 Test to prove push fails correctly when custom path is wrong (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m f380954 2016-07-13 Basic test just to bootstrap & confirm custom adapter running (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m d12dfec 2016-07-13 Better error handling when initialising adapters (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 33fa57d 2016-07-13 Need to include actions from API in custom transfer protocol (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 3c4d4d3 2016-07-13 Make oids match to avoid confusion in docs (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 8f1e943 2016-07-13 Comment fixes (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m d351324 2016-07-13 Extend interfaces and establish scaffold for custom transfer processes (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 357703e 2016-07-13 Read custom transfer configurations & test (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 5fe21cc 2016-07-13 Allow GitConfigBool to have a default of false as well (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 7250e7e 2016-07-13 Remove 'order of preference' from API transfer list (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 7c35e9c 2016-07-13 Remove mention of extended transfers in v1 api docs, in v1.3 docs (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m c25d4dd 2016-07-13 Refactor StringSet into tools package (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 2d5e9b5 2016-07-13 Refined first pass design for custom transfers and convert to markdown (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m 82ca240 2016-07-13 Initial design work on custom transfer processes (Steve Streeting[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m[31m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
[33m|[m [31m|[m [31m|[m[31m/[m[33m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m e50af51 2016-07-21 prefer git config over HTTP_PROXY (risk danger olson[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m 7046d1f 2016-07-14 httputil/proxy: assert.Nil over assert.Equal(t, nil) (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m cb73a19 2016-07-14 lfs,httputil: move proxy test to httputil package (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m f0acdb2 2016-07-14 proxy: assert NO_PROXY testcase (Taylor Blau[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m   adbfa1e 2016-07-13 Merge branch 'master' into jonmagic-use-proxy-from-git-config (risk danger olson[32m[m)
[33m|[m [31m|[m [31m|[m [33m|[m [33m|[m [36m|[m[33m\[m [1;31m\[m [35m\[m  
[33m|[m [31m|[m[33m_[m[31m|[m[33m_[m[33m|[m[33m_[m[33m|[m[33m_[m[36m|[m[33m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [33m|[m [33m|[m [36m|[m [1;31m|[m [35m|[m   
* [31m|[m [31m|[m [33m|[m [33m|[m [36m|[m [1;31m|[m [35m|[m   76186e9 2016-07-13 Merge pull request #1369 from github/git-2.9.1 (risk danger olson[32m[m)
[31m|[m[1;33m\[m [31m\[m [31m\[m [33m\[m [33m\[m [36m\[m [1;31m\[m [35m\[m  
[31m|[m [1;33m|[m[31m_[m[31m|[m[31m/[m [33m/[m [33m/[m [36m/[m [1;31m/[m [35m/[m  
[31m|[m[31m/[m[1;33m|[m [31m|[m [33m|[m [33m|[m [36m|[m [1;31m|[m [35m|[m   
[31m|[m * [31m|[m [33m|[m [33m|[m [36m|[m [1;31m|[m [35m|[m 52c7736 2016-07-13 test/zero-len: `grep -E` and remove `tee` (Taylor Blau[32m[m)
[31m|[m * [31m|[m [33m|[m [33m|[m [36m|[m [1;31m|[m [35m|[m dc42704 2016-07-13 test/zero-len: update test for git v2.9.1 (Taylor Blau[32m[m)
[31m|[m[31m/[m [31m/[m [33m/[m [33m/[m [36m/[m [1;31m/[m [35m/[m  
[31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m 6896c5a 2016-07-13 lfs/proxy: use http.NewRequest to handle request building (Taylor Blau[32m[m)
[31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m 146cdcf 2016-07-11 credit go std lib (risk danger olson[32m[m)
[31m|[m [31m|[m [33m|[m [33m|[m * [1;31m|[m [35m|[m   1392d49 2016-07-11 Merge branch 'master' into jonmagic-use-proxy-from-git-config (risk danger olson[32m[m)
[31m|[m [31m|[m [33m|[m [33m|[m [1;34m|[m[31m\[m [1;31m\[m [35m\[m  
[31m|[m [31m|[m[31m_[m[33m|[m[31m_[m[33m|[m[31m_[m[1;34m|[m[31m/[m [1;31m/[m [35m/[m  
[31m|[m[31m/[m[31m|[m [33m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m   
* [31m|[m [33m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m   2900456 2016-07-11 Merge pull request #1361 from larsxschneider/git-source (Taylor Blau[32m[m)
[33m|[m[31m\[m [31m\[m [33m\[m [33m\[m [1;34m\[m [1;31m\[m [35m\[m  
[33m|[m [31m|[m[33m_[m[31m|[m[33m/[m [33m/[m [1;34m/[m [1;31m/[m [35m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m   
[33m|[m * [31m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m e5756d3 2016-07-11 travis-ci: add a build job to test against upcoming versions of Git (Lars Schneider[32m[m)
[33m|[m[33m/[m [31m/[m [33m/[m [1;34m/[m [1;31m/[m [35m/[m  
* [31m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m   bf5fd83 2016-07-11 Merge pull request #1360 from jasperla/openbsd (Taylor Blau[32m[m)
[31m|[m[33m\[m [31m\[m [33m\[m [1;34m\[m [1;31m\[m [35m\[m  
[31m|[m [33m|[m[31m/[m [33m/[m [1;34m/[m [1;31m/[m [35m/[m  
[31m|[m[31m/[m[33m|[m [33m|[m [1;34m|[m [1;31m|[m [35m|[m   
[31m|[m * [33m|[m [1;34m|[m [1;31m|[m [35m|[m 8d3f881 2016-07-11 Unbreak building httputil on OpenBSD (Jasper Lievisse Adriaanse[32m[m)
[31m|[m[31m/[m [33m/[m [1;34m/[m [1;31m/[m [35m/[m  
[31m|[m [33m|[m * [1;31m|[m [35m|[m 9cc130f 2016-07-08 copy http.ProxyFromEnvironment so we can put our own unique git spin on it (risk danger olson[32m[m)
[31m|[m [33m|[m * [1;31m|[m [35m|[m f3f32b0 2016-07-08 send back ProxyFromEnvironment() errors too (risk danger olson[32m[m)
[31m|[m [33m|[m * [1;31m|[m [35m|[m 08dbfc2 2016-07-08 -a (risk danger olson[32m[m)
[31m|[m [33m|[m * [1;31m|[m [35m|[m   e5ddd5d 2016-07-08 fix merge conflict (risk danger olson[32m[m)
[31m|[m [33m|[m [31m|[m[35m\[m [1;31m\[m [35m\[m  
[31m|[m [33m|[m[31m/[m [35m/[m [1;31m/[m [35m/[m  
[31m|[m[31m/[m[33m|[m [35m|[m [1;31m|[m [35m|[m   
[31m|[m [33m|[m * [1;31m|[m [35m|[m 4bd51e4 2016-06-03 Prefer proxy from env over git config Add three go tests and only one works right now (Jonathan Hoyt[32m[m)
[31m|[m [33m|[m * [1;31m|[m [35m|[m 690262b 2016-04-22 Use proxy from config or else use proxy from environment (Lizz Hale[32m[m)
[31m|[m [33m|[m * [1;31m|[m [35m|[m 3203c62 2016-04-22 Add Proxy function to Configuration (Jonathan Hoyt[32m[m)
[31m|[m [33m|[m [1;31m|[m[1;31m/[m [35m/[m  
* [33m|[m [1;31m|[m [35m|[m   7b8cdca 2016-07-08 Merge branch 'pascalberger-chocolateyupdate' (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;31m\[m [35m\[m  
[36m|[m * [33m\[m [1;31m\[m [35m\[m   c40da4e 2016-07-08 fix merge conflict (risk danger olson[32m[m)
[36m|[m [36m|[m[1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[36m|[m[36m/[m [1;33m/[m [33m/[m [1;31m/[m [35m/[m  
[36m|[m * [33m|[m [1;31m|[m [35m|[m 138755a 2016-04-14 Change to the git-lfs meta package for Chocolatey (Pascal Berger[32m[m)
* [1;33m|[m [33m|[m [1;31m|[m [35m|[m   5c6fecf 2016-07-08 Merge pull request #1255 from github/transferqueue-watch-race (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;33m\[m [33m\[m [1;31m\[m [35m\[m   3718379 2016-07-08 fix merge conflict (risk danger olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;33m/[m [33m/[m [1;31m/[m [35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   
* [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   05495b1 2016-07-07 Merge pull request #1222 from zeldin/gccgo (Taylor Blau[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[32m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 9a8648c 2016-05-18 Create Makefile for building with gccgo (Marcus Comstedt[32m[m)
* [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 2a420e1 2016-07-06 docs/man: remove reference to `--no-touch` (Taylor Blau[32m[m)
* [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   c8484e4 2016-07-06 Merge pull request #1344 from github/track-no-touch-logs (Taylor Blau[32m[m)
[34m|[m[35m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 26d3cc1 2016-07-06 commands/track: clarify verbose log messages (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 0886203 2016-07-06 commands/track: remove --no-touch option (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m f516dfe 2016-07-06 command/track: Print() when searching for files (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 75c40be 2016-07-06 docs/man: fix typo for track.ronn (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 8b8ea26 2016-07-06 commands/track: teach `track` how to --dry-run (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 5e18bf7 2016-07-06 test/track: use log from `git lfs track` instead of atime (Taylor Blau[32m[m)
[34m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m c3f3842 2016-07-06 cmd,doc,test: teach `git lfs track --{no-touch,verbose} (Taylor Blau[32m[m)
[34m|[m[34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m [35m/[m  
* [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 3d11892 2016-07-06 ROADMAP: remove "use expires_at property" item (Taylor Blau[32m[m)
* [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   54be0d9 2016-07-06 Merge pull request #1350 from github/retry-expired-objects (Taylor Blau[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[36m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m a1f5f6f 2016-07-06 transfer/adapterbase: add a short grace period to expiration checks (Taylor Blau[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 28e878c 2016-07-06 api/object: document and test the IsExpired function (Taylor Blau[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   4722fa5 2016-07-05 Merge branch 'master' into retry-expired-objects (Taylor Blau[32m[m)
[36m|[m [1;32m|[m[36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[36m|[m [1;32m|[m[36m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m [35m/[m  
[36m|[m[36m/[m[1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   
* [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 9f73b3c 2016-07-05 roadmap: remove defensive `track` changes (Taylor Blau[32m[m)
* [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   7b67335 2016-07-05 Merge pull request #1346 from github/safe-track-patterns (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 81e525c 2016-07-05 commands, test: change to blocklist (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 5da9560 2016-07-05 git/git: sanitize pattern in git.GetTrackedFiles (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 3ad1222 2016-07-05 commands/track: delay .gitattributes and touch until filters are passed (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 7c41eeb 2016-07-05 command/track: remove TODO comment on prefix blacklist (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m e34f7b3 2016-07-05 test/track: touch and stage files (Taylor Blau[32m[m)
[1;34m|[m * [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m   08afb68 2016-07-05 Merge branch 'master' into safe-track-patterns (Taylor Blau[32m[m)
[1;34m|[m [1;36m|[m[31m\[m [1;32m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [31m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 6c5f738 2016-07-04 commands/track: check tracked files using strings.HasPrefix (Taylor Blau[32m[m)
[1;34m|[m * [31m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m [35m|[m 405eaaf 2016-07-01 commands/track: ignore blacklisted paths by glob and name (Taylor Blau[32m[m)
[1;34m|[m [35m|[m [31m|[m[35m_[m[1;32m|[m[35m_[m[33m|[m[35m_[m[1;36m|[m[35m_[m[1;33m|[m[35m_[m[33m|[m[35m_[m[1;31m|[m[35m/[m  
[1;34m|[m [35m|[m[35m/[m[31m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [35m|[m [31m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   a78d4b8 2016-07-05 Merge pull request #1349 from github/safe-require-stdin (Taylor Blau[32m[m)
[31m|[m[33m\[m [35m\[m [31m\[m [1;32m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [33m|[m[31m_[m[35m|[m[31m/[m [1;32m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[33m|[m [35m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[31m|[m * [35m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   e7594c9 2016-07-05 Merge branch 'master' into safe-require-stdin (Taylor Blau[32m[m)
[31m|[m [34m|[m[31m\[m [35m\[m [1;32m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [34m|[m[31m/[m [35m/[m [1;32m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[34m|[m [35m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[31m|[m * [35m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9fc5f7a 2016-07-04 commands: log error in requireStdin (Taylor Blau[32m[m)
[31m|[m * [35m|[m [1;32m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3088318 2016-07-04 commands: safety check to os.Stdin.Stat() (Taylor Blau[32m[m)
[31m|[m [35m|[m[35m/[m [1;32m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m [35m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m afc28d7 2016-07-05 test/push: uncomment (1 of 1) check (Taylor Blau[32m[m)
[31m|[m [35m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ea60ea0 2016-07-05 transfer/queue: only mark successful transfers as having finished (Taylor Blau[32m[m)
[31m|[m [35m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   fb029fe 2016-07-05 Merge branch 'master' into retry-expired-objects (Taylor Blau[32m[m)
[31m|[m [35m|[m [36m|[m[31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [35m|[m[31m_[m[36m|[m[31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[35m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [35m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   8ee9225 2016-07-05 Merge pull request #1303 from github/experimental/transfer-features-p4 (Steve Streeting[32m[m)
[1;32m|[m[1;33m\[m [35m\[m [36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [35m\[m [36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   1422e8f 2016-07-05 Merge branch 'master' into experimental/transfer-features-p4 (Steve Streeting[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [35m\[m [36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [35m/[m [36m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [35m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;34m|[m [35m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   b44c9f3 2016-07-05 Merge pull request #1352 from github/revert-1262-checkout-unstaged (Steve Streeting[32m[m)
[35m|[m[31m\[m [1;34m\[m [35m\[m [36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[35m|[m [31m|[m[35m_[m[1;34m|[m[35m/[m [36m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[35m|[m[35m/[m[31m|[m [1;34m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[35m|[m * [1;34m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1f386c7 2016-07-05 Revert "Add checkout --unstaged flag" (Steve Streeting[32m[m)
[35m|[m[35m/[m [1;34m/[m [36m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[35m|[m * [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   a2d2c9c 2016-07-04 Merge branch 'master' into experimental/transfer-features-p4 (Steve Streeting[32m[m)
[35m|[m [32m|[m[35m\[m [36m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[35m|[m [32m|[m[35m/[m [36m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[35m|[m[35m/[m[32m|[m [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[35m|[m * [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m db3f17f 2016-07-04 PR comments (Steve Streeting[32m[m)
[35m|[m * [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1a8364b 2016-06-14 Don't include creds by default in tus call, prompt on 401 instead (Steve Streeting[32m[m)
[35m|[m * [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 44cf589 2016-06-10 Add tests for tus.io resume upload support (Steve Streeting[32m[m)
[35m|[m * [36m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m a75d9c4 2016-06-10 First pass of tus.io resumable upload adapter (Steve Streeting[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5f36a03 2016-07-05 test/push: test retrying expired objects (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 2e0ba6b 2016-07-05 test/gitserver: implement once-per-repo expirables (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 022ef14 2016-07-05 api/object: only count non-zero expirations (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 0c4328f 2016-07-05 lfs,transfer: retry transfer before request was made (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b75ac1d 2016-07-05 api: inform Object whether or not it has expired (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 81ce517 2016-07-04 lfs/transfer: retry OOD actions (Taylor Blau[32m[m)
[35m|[m [33m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b22a66f 2016-07-04 api: support ExpiresAt property on Actions (Taylor Blau[32m[m)
[35m|[m [33m|[m[35m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[35m|[m[35m/[m[33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   7ee533f 2016-06-29 Merge pull request #1335 from github/fix-logs-manpage (Taylor Blau[32m[m)
[34m|[m[35m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m * [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m aaf08c2 2016-06-29 docs/man: move "logs" subcommands from OPTIONS to COMMANDS (Taylor Blau[32m[m)
[34m|[m[34m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   22654de 2016-06-28 Merge pull request #1332 from github/tb/add-to-readme (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3210a57 2016-06-28 README: add @ttaylorr to core team (Taylor Blau[32m[m)
[36m|[m[36m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   1dd97dc 2016-06-28 Merge pull request #1298 from javabrett/remove-centos-5-docker (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b560b85 2016-06-09 Removed CentOS 5 from dockers.  Fixed #1295. (Brett Randall[32m[m)
* [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   3ca6d04 2016-06-28 Merge pull request #1292 from javabrett/enforce-packagecloud-ruby-minimum-version (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m * [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 43ef114 2016-06-27 Enforced a minimum gem version of 1.0.4 for packagecloud-ruby (1.0.3 is the current bare minimum). Avoids issues such as encountered in #1288. (Brett Randall[32m[m)
* [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   0532029 2016-06-28 Merge pull request #1262 from orivej/checkout-unstaged (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ef8c178 2016-05-30 Add checkout --unstaged flag (Orivej Desh[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   31af5ef 2016-06-28 Merge pull request #1305 from GabLeRoux/patch-1 (Taylor Blau[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m * [31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   65ea429 2016-06-28 Merge branch 'master' into patch-1 (Gabriel Le Breton[32m[m)
[32m|[m [34m|[m[32m\[m [31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m [34m|[m[32m/[m [31m/[m [1;35m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[32m|[m[32m/[m[34m|[m [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [34m|[m [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   2444c2d 2016-06-28 Merge pull request #1323 from omonnier/fix_fetch_sha1 (Taylor Blau[32m[m)
[36m|[m[1;31m\[m [34m\[m [31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [34m|[m [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m dfe3941 2016-06-28 Enhance fetch integration tests with sha1 (Olivier Monnier[32m[m)
[36m|[m * [34m|[m [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f86997f 2016-06-28 Fix 'git lfs fetch' with a sha1 ref (Olivier Monnier[32m[m)
[36m|[m[36m/[m [34m/[m [31m/[m [1;35m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [31m|[m [1;35m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   94fc09e 2016-06-27 Merge pull request #1331 from github/update-version (Taylor Blau[32m[m)
[1;35m|[m[1;33m\[m [34m\[m [31m\[m [1;35m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;35m|[m [1;33m|[m[1;35m_[m[34m|[m[1;35m_[m[31m|[m[1;35m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;33m|[m [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[1;35m|[m * [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 6834a6f 2016-06-27 release git-lfs v1.2.1 (risk danger olson[32m[m)
[1;35m|[m[1;35m/[m [34m/[m [31m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   7fea47d 2016-06-23 Merge pull request #1321 from github/tb/clone-path-flags (Taylor Blau[32m[m)
[1;34m|[m[1;35m\[m [34m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m * [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 37bfee1 2016-06-23 test/clone: group clone tests, test config+args as well (Taylor Blau[32m[m)
[1;34m|[m * [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 477773b 2016-06-23 man: add new OPTIONS to git-lfs-clone.1 manpage (Taylor Blau[32m[m)
[1;34m|[m * [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 72b52c5 2016-06-23 cmd/clone: add include/exclude via flags and config (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m [34m/[m [31m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   7d8734c 2016-06-23 Merge pull request #1324 from omonnier/test_extra_args (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [34m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [34m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b542186 2016-06-23 script/test: propagate extra args to go test (Olivier Monnier[32m[m)
[1;36m|[m[1;36m/[m [34m/[m [31m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m * [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   ed0f3af 2016-06-23 Merge branch 'master' into patch-1 (Gabriel Le Breton[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m [32m|[m[1;36m/[m [31m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [32m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   3ee47a0 2016-06-21 Merge pull request #1256 from ttaylorr/lock-commands (Taylor Blau[32m[m)
[34m|[m[35m\[m [32m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m * [32m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   f4b3cd7 2016-06-21 Merge branch 'master' into lock-commands (Taylor Blau[32m[m)
[34m|[m [36m|[m[34m\[m [32m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m [36m|[m[34m/[m [32m/[m [31m/[m [1;33m/[m [33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[34m|[m[34m/[m[36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m fe1aa74 2016-06-18 test/fetch: fix mising semi-colon in for-loop (Taylor Blau[32m[m)
* [36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   271b16d 2016-06-16 Merge pull request #1310 from github/fetch-ignore-head-with-all (Taylor Blau[32m[m)
[33m|[m[1;33m\[m [36m\[m [32m\[m [31m\[m [1;33m\[m [33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[33m|[m [1;33m|[m[33m_[m[36m|[m[33m_[m[32m|[m[33m_[m[31m|[m[33m_[m[1;33m|[m[33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[33m|[m * [36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 303e2f0 2016-06-16 test/fetch: test "fetch --all" using a --bare clone (Taylor Blau[32m[m)
[33m|[m * [36m|[m [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 763025c 2016-06-16 commands/fetch: ignore HEAD ref with --all (Taylor Blau[32m[m)
[33m|[m[33m/[m [36m/[m [32m/[m [31m/[m [1;33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8fb4415 2016-06-04 locks: path formatting changes (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 31e5ed0 2016-06-03 test/locks: qualify with --path (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4d8e539 2016-06-03 test: test the `git lfs locks` command (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c4fd57d 2016-06-03 test/server: return three locks at a time (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 48490a6 2016-06-03 commands: teach locks how to follow cursor (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 371e464 2016-06-03 test/server: implement 'path' QSP (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ae533d9 2016-06-03 test/helper: push to origin and avoid ambiguity on git-latest (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m fa407e9 2016-06-03 test: use unique filepaths (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3fbd162 2016-06-03 test: test `git lfs unlock` (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 051a087 2016-06-03 test/helpers: fix refute_server_lock method (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 09c3438 2016-06-03 test/server: handle weird path casing (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1f9bddc 2016-06-03 commmands/unlock: fix incorrect casing (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 21b9238 2016-06-03 commands/lock: properly trim prefix from lock path (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c783561 2016-06-03 test/lock: test the `git lfs lock` command (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4b172af 2016-06-03 test/helpers: add a few testhelpers for lock tests (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7b8427b 2016-06-03 test/server: enforce locks be made against unique paths (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9ccd526 2016-06-03 test/server: temporary workaround for 301 redirects (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1c6ae7e 2016-06-03 test/cmd: reference implementation of the locks API (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d4c6c02 2016-06-03 command_unlock: allow the usage of --force (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 36804ef 2016-06-03 commands_{lock,locks,unlock}: implement lockPath for great good (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m fad3083 2016-06-03 commands_{lock,locks,unlock}: allow custom --remote specification (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b5605e0 2016-06-03 commands/unlock: add --id flag to unlock command (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f74cd86 2016-06-03 commands: update to latest API changes (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 2680afc 2016-06-03 commands_{lock,unlock,locks}: implement lock commands (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m a4264bf 2016-06-03 commands: expose package-local singleton API instance (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d2bb91e 2016-06-03 api/locks_api: implement CurrentCommitter() (Taylor Blau[32m[m)
[33m|[m * [32m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ee97011 2016-06-03 api/http_lifecycle: add a note about handling HTTP errors (Taylor Blau[32m[m)
[33m|[m [1;33m|[m * [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 04a10ed 2016-06-23 Readme update (typos and wording) (Gabriel Le Breton[32m[m)
[33m|[m [1;33m|[m[33m/[m [31m/[m [1;33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   e6600a0 2016-06-10 Merge pull request #1299 from github/transfer-basic-only (Steve Streeting[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m * [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   905bc69 2016-06-10 Merge branch 'master' into transfer-basic-only (Steve Streeting[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;33m/[m [31m/[m [1;33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;36m|[m [1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   93c6d40 2016-06-10 Merge pull request #1297 from github/experimental/transfer-features-p3 (Steve Streeting[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m * [1;36m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   8b5a991 2016-06-10 Merge branch 'master' into experimental/transfer-features-p3 (Steve Streeting[32m[m)
[32m|[m [34m|[m[32m\[m [1;36m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m [34m|[m[32m/[m [1;36m/[m [1;33m/[m [31m/[m [1;33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[32m|[m[32m/[m[34m|[m [1;36m|[m [1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [34m|[m [1;36m|[m [1;33m|[m [31m|[m [1;33m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   9ffc618 2016-06-10 Merge pull request #1296 from github/progress-skip-fixes (Steve Streeting[32m[m)
[1;33m|[m[1;31m\[m [34m\[m [1;36m\[m [1;33m\[m [31m\[m [1;33m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;33m|[m [1;31m|[m[1;33m_[m[34m|[m[1;33m_[m[1;36m|[m[1;33m_[m[1;33m|[m[1;33m_[m[31m|[m[1;33m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[1;31m|[m [34m|[m [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[1;33m|[m * [34m|[m [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 40cd7c4 2016-06-08 Update tests (Steve Streeting[32m[m)
[1;33m|[m * [34m|[m [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 311089c 2016-06-08 Print report from meter when there were files but they were skipped (Steve Streeting[32m[m)
[1;33m|[m * [34m|[m [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d452f0e 2016-06-08 Reduce estimated bytes & files when skipping so completion easier to parse (Steve Streeting[32m[m)
[1;33m|[m * [34m|[m [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 85d1e2d 2016-06-08 Improve reporting of entirely skipped files in progress (Steve Streeting[32m[m)
[1;33m|[m[1;33m/[m [34m/[m [1;36m/[m [1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3528821 2016-06-10 Range header byte ranges are inclusive (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b068c95 2016-06-09 Check callback is present before skipping it forward when resuming (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3574c46 2016-06-09 Support download resuming beyond 2GB boundary (64 bit offsets) (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 19304ae 2016-06-08 Avoid re-request when server returns 200 instead of 206 to Range req (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f082660 2016-06-08 Fix Content-Range parsing, slightly different format to Range (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4a5519e 2016-06-08 Make HTTP Range resume download support the default, fallback handles other cases (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c044ef5 2016-06-07 Validate Content-Range on 206 range response & diagnose failure details (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c8817ff 2016-06-07 Download resume fallback to re-download on 416 error (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b7b95f7 2016-06-07 Fix retry; only confirm that auth signal done when auth func called (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9e12668 2016-06-07 Place incomplete downloads under .git/lfs/objects/incomplete (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c9c457c 2016-06-07 Test http-range resume download (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b727d95 2016-06-07 Add trace info to http-range resume (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m aef3f9c 2016-06-07 First pass of resuming downloads via HTTP Range (Steve Streeting[32m[m)
[1;33m|[m * [1;36m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m bed2540 2016-06-07 Refactor common functionality from basicAdapter for re-use in http range (Steve Streeting[32m[m)
[1;33m|[m [1;31m|[m * [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5d6907d 2016-06-10 Document default behaviour of BasicTransfersOnly() (Steve Streeting[32m[m)
[1;33m|[m [1;31m|[m * [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b051a41 2016-06-09 Add `lfs.basictransfersonly` option to disable non-basic transfer adapters (Steve Streeting[32m[m)
[1;33m|[m [1;31m|[m[1;33m/[m [1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[1;31m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;31m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   12fe249 2016-06-07 Merge pull request #1279 from github/experimental/transfer-features-p2 (Steve Streeting[32m[m)
[1;32m|[m[1;31m\[m [1;31m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m [1;31m|[m[1;31m/[m [1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m * [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   0681774 2016-06-07 Merge branch 'master' into experimental/transfer-features-p2 (Steve Streeting[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;34m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   1789729 2016-06-07 Merge pull request #1265 from github/experimental/transfer-features (Steve Streeting[32m[m)
[1;36m|[m[31m\[m [1;34m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [1;34m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   9429193 2016-06-07 Merge branch 'master' into experimental/transfer-features (Steve Streeting[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [1;34m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m [32m|[m[1;36m/[m [1;34m/[m [1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [1;34m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [32m|[m [1;34m|[m [1;33m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   993fdc6 2016-06-06 Merge pull request #1291 from javabrett/debian-build-vendor-test-excludes (Taylor Blau[32m[m)
[1;33m|[m[35m\[m [32m\[m [1;34m\[m [1;33m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;33m|[m [35m|[m[1;33m_[m[32m|[m[1;33m_[m[1;34m|[m[1;33m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[35m|[m [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[1;33m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m af10cca 2016-06-06 Excluded problematic vendor tests from Debian build. #1290. Excluded those in the CI build plus github.com/spf13/cobra which also seems to fail in dpkg build. (Brett Randall[32m[m)
[1;33m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 539b1dd 2016-06-06 Set DH_GOLANG_INSTALL_ALL=1 to test workaround for missing vendor file in tests, vendor/github.com/xeipuuv/gojsonschema/json_schema_test_suite/type/schema_0.json . #1290. (Brett Randall[32m[m)
[1;33m|[m[1;33m/[m [32m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   bf9bc1f 2016-06-03 Merge pull request #1260 from pdf/git_ssh_command (Steve Streeting[32m[m)
[36m|[m[1;31m\[m [32m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7e46933 2016-06-03 Implement support for GIT_SSH_COMMAND (Peter Fern[32m[m)
[36m|[m[36m/[m [32m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   fbd202d 2016-06-02 Merge pull request #1281 from github/changelog-script (Taylor Blau[32m[m)
[1;32m|[m[1;33m\[m [32m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [32m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   e06e596 2016-06-02 Merge branch 'master' into changelog-script (risk danger olson[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [32m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [32m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;34m|[m [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   cc37270 2016-06-02 Merge pull request #1271 from ttaylorr/gitignore-stuff (Taylor Blau[32m[m)
[1;36m|[m[31m\[m [1;34m\[m [32m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [1;34m|[m [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ce52008 2016-06-02 gitignore: ignore lfstest-* files (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m [1;34m/[m [32m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 566ad5d 2016-06-02 script/changelog: teach '?' operator and handle unknown input gracefully (Taylor Blau[32m[m)
[1;36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m cb56e02 2016-06-02 script/changelog: strip out the 'Backport: ' prefix (Taylor Blau[32m[m)
[1;36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m db744f3 2016-06-02 script/changelog: use h3 headers, insert extra newlines (Taylor Blau[32m[m)
[1;36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 77ccd03 2016-06-02 script/changelog: use * for bullet-points (Taylor Blau[32m[m)
[1;36m|[m * [32m|[m [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 991f85d 2016-06-02 script: add changelog script (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m [32m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   82a9b60 2016-06-02 Merge branch 'master' into experimental/transfer-features-p2 (Steve Streeting[32m[m)
[1;36m|[m [32m|[m [32m|[m[1;36m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m [32m|[m[1;36m_[m[32m|[m[1;36m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;36m|[m[1;36m/[m[32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   28819ce 2016-06-02 Merge pull request #1274 from github/disable-gojsonschema-test (Steve Streeting[32m[m)
[34m|[m[35m\[m [32m\[m [32m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m * [32m\[m [32m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   6e8ea37 2016-06-02 Merge branch 'master' into disable-gojsonschema-test (Steve Streeting[32m[m)
[34m|[m [36m|[m[34m\[m [32m\[m [32m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m [36m|[m[34m/[m [32m/[m [32m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[34m|[m[34m/[m[36m|[m [32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [36m|[m [32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   5cc5278 2016-06-02 Merge pull request #1273 from github/ci/ignore-osx-failures (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [32m\[m [32m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [36m|[m [32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m a207748 2016-06-01 Update .travis.yml (risk danger olson[32m[m)
[1;32m|[m[1;32m/[m [36m/[m [32m/[m [32m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m * [32m|[m [32m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9164f83 2016-06-02 Disable gojsonschema test, causes failures when firewalls block it (Steve Streeting[32m[m)
[1;32m|[m[1;32m/[m [32m/[m [32m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 70073b3 2016-06-02 Update schemas to allow additional fields (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m eb62c08 2016-06-02 Deal gracefully with server changing its mind on adapter between batches (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c5c2a75 2016-06-02 Keep batch structures private, simplify adapter creation (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m defe1f7 2016-06-02 Always use basic adapter in legacy api route (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 05c3ed7 2016-06-02 Deal more elegantly with empty batch requests (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m e0923fd 2016-06-02 WriteHeader() must always come before Write() in ResponseWriter (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f8cc589 2016-06-02 Add transfer adapter negotiation to batch api request / response (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8ba9dfe 2016-06-02 Fix example JSON (Steve Streeting[32m[m)
[1;32m|[m [32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8dabcb2 2016-06-02 Update API docs for v1.3 transfer extensions (Steve Streeting[32m[m)
[1;32m|[m [32m|[m[32m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m * [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   f2678fc 2016-06-02 Merge branch 'master' into experimental/transfer-features (Steve Streeting[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   5bed973 2016-06-01 Merge pull request #1257 from ttaylorr/config-include-exclude (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9e0a30a 2016-06-01 test: include fetch test for unclean paths (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 02236d2 2016-05-31 commands: test determineIncludeExcludePaths (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f9536fc 2016-05-31 config: test that include/exclude filters are loaded correctly (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 05b01a5 2016-05-31 config: introduce NewFromValues (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 2c8d8d8 2016-05-31 tools/filetools: add test for CleanPaths(Default) (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c5ec526 2016-05-31 commands,config: re-introduce determineIncludeExcludePaths() (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 0971cde 2016-05-31 commands: use config-based helpers for path-safe include/exclude operations (Taylor Blau[32m[m)
[1;36m|[m * [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d7ee3d3 2016-05-31 config,lfs: introduce Include/Exclude path helpers to config.Configuration (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m [1;34m/[m [31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [1;34m|[m [31m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   0a16cf8 2016-05-31 Merge pull request #1267 from ttaylorr/noop-creds-helper (risk danger olson[32m[m)
[31m|[m[33m\[m [1;34m\[m [31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [33m|[m[31m_[m[1;34m|[m[31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[33m|[m [1;34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[31m|[m * [1;34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5c694a5 2016-05-31 test: useHttpPath in auth tests (Taylor Blau[32m[m)
[31m|[m * [1;34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ceb0e02 2016-05-31 test: use noop credential helper for auth tests (Taylor Blau[32m[m)
[31m|[m[31m/[m [1;34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 10623f5 2016-05-31 Tests for transfer registration system (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ccf40c7 2016-05-31 Simplify transfer adapter creation, drop factory type & use funcs (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3ee0946 2016-05-31 Rationalise TODO LEGACY API comments so easier to search for later (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 299ed64 2016-05-31 Fix panic when callback passed to smudge is nil (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 956a7ab 2016-05-31 Fix smudge with new transfer adapter (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f4d2a51 2016-05-31 Extra tracing on basic adapter (Steve Streeting[32m[m)
[31m|[m * [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   b3b3118 2016-05-31 Merge branch 'master' into experimental/transfer-features (Steve Streeting[32m[m)
[31m|[m [34m|[m[31m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [34m|[m[31m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   2b1a0b0 2016-05-27 Merge pull request #1259 from ttaylorr/force-unlock-param (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ec6137a 2016-05-27 api/lock_api: use `id`, `force` as explicit arguments (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f5578f3 2016-05-27 api: add Force option to unlock requests (Taylor Blau[32m[m)
[36m|[m[36m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   b2065a9 2016-05-27 Merge pull request #1258 from ttaylorr/fix-spurious-vendor-errs (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ff463d5 2016-05-27 script/{lint,test}: fix spurious vendor-related import errors (Taylor Blau[32m[m)
[1;32m|[m[1;32m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   5e9fcdd 2016-05-25 Merge pull request #1253 from ttaylorr/nuke-test-suites (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 0008d16 2016-05-25 api/http_lifecycle_test: use normal Golang tests over suites (Taylor Blau[32m[m)
[1;34m|[m[1;34m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   9c1b2ac 2016-05-25 Merge pull request #1252 from ttaylorr/api-operations (risk danger olson[32m[m)
[1;36m|[m[31m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 44ac3a9 2016-05-25 api{client,lifecycle}: use config.Endpoint to resolve root (Taylor Blau[32m[m)
[1;36m|[m[1;36m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   d25c0eb 2016-05-25 Merge pull request #1250 from ttaylorr/remove-old-asserts (risk danger olson[32m[m)
[32m|[m[33m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 354f3db 2016-05-25 vendor: remove github.com/technoweenie/assert (Taylor Blau[32m[m)
[32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m fc1daa2 2016-05-25 lfs: use github.com/stretchr/testify for assertions (Taylor Blau[32m[m)
[32m|[m[32m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   b633aa1 2016-05-24 Merge pull request #1248 from github/full-remote-r (risk danger olson[32m[m)
[34m|[m[35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 04db420 2016-05-24 Return a fully remote ref to reduce chances of ref clashes (risk danger olson[32m[m)
[34m|[m[34m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   ac1a109 2016-05-24 Merge pull request #1236 from ttaylorr/api-client (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4bf693e 2016-05-24 api: http.Method{Get,Post} does not exist in Go <1.6 (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 60fad56 2016-05-24 lint: lint all files (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ecdba8d 2016-05-24 api: un-vendor dependencies (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b5cfb3e 2016-05-24 api/http_lifecycle: use httputil.DoHttpRequestWithRedirects() (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f6a0137 2016-05-24 api/http_lifecycle: use httputil's client, update tests (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 00105ac 2016-05-24 api/locks: LockService documentation fixes (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1a9d2ab 2016-05-24 api/lifecycle: documentation tweaks (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c7f2cd0 2016-05-24 api/http_lifecycle: make internal methods, add an ASK for @sinbad (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m a71d5f7 2016-05-24 api/client: minor touch-ups to documentation (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7399b85 2016-05-24 api/http_lifecycle: fix net/http integration (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8887924 2016-05-24 api/{lock,schema} validate LockList against schema (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d352a3f 2016-05-24 api/schema: rename AssertSchema to AssertRequestSchema (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1654834 2016-05-24 api/lock: implement LockService.Search (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 32c557a 2016-05-24 api/{lock,schema}: add lock API schemas, test against the Lock API (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7849ff3 2016-05-24 api/schema: introduce SchemaValidator, Assert() and Refute() (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f7fc823 2016-05-24 api/schema: remove MethodTestCase in favor of AssertSchema() (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7b73f83 2016-05-24 api/schema: teach MethodTestCase how to serialize a repsonse (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c282983 2016-05-24 api/schema: initial take on MethodTestCase (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9234838 2016-05-24 api/http_lifecycle: test HttpLifecycle implementation (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ff2a96b 2016-05-24 api/response: HttpResponse tests (Taylor Blau[32m[m)
[36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 91c1b2d 2016-05-24 api: introduce client, lifecycle and response types (Taylor Blau[32m[m)
[36m|[m[36m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   6d53a5c 2016-05-24 Merge pull request #1243 from ttaylorr/glide-deps (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7ed06cd 2016-05-24 script/lint: update text to refer to Glide (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 28541a5 2016-05-23 script/test: don't test failing dependencies (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 0af911a 2016-05-23 vendor: update technoweenie/assert to latest (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m deab9c3 2016-05-23 script/vendor: remove broken package in go-ntlm (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 460a2a1 2016-05-23 glide,vendor: update go-ntlm to remove log4go dependency (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ef19d83 2016-05-23 glide,vendor: remove broken directories (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m e079493 2016-05-23 script/test: note failure in olekukonko/ts (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 85e080d 2016-05-23 contributing: s/Nut/Glide (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m e1927bd 2016-05-23 script,test,debian: GO15VENDOREXPERIMENT=1 (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5d4ba59 2016-05-23 script/test: exclude vendored packages from tests (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4e617a7 2016-05-23 glide,vendor: remove failing subpackages, modify script/vendor to not overwrite this (Taylor Blau[32m[m)
[1;32m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4593d0a 2016-05-23 vendor: vendor dependencies in vendor/ using Glide (Taylor Blau[32m[m)
[1;32m|[m[1;32m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   3efa04c 2016-05-20 Merge pull request #1234 from github/deps/jsonschema (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;34m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d0b4256 2016-05-20 add the gojsonpointer and gojsonreference packages (risk danger olson[32m[m)
[1;34m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 445a765 2016-05-20 add a jsonschema package (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   841853d 2016-05-19 Merge pull request #1228 from github/add-testify (risk danger olson[32m[m)
[1;36m|[m[31m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 6e946ba 2016-05-18 vendor the testify package (risk danger olson[32m[m)
* [31m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   b2e7e70 2016-05-18 Merge pull request #1229 from github/git-tests-outside-repo (risk danger olson[32m[m)
[31m|[m[33m\[m [31m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[31m|[m [33m|[m[31m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[31m|[m[31m/[m[33m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[31m|[m * [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1371c22 2016-05-18 get git tests passing when run outside of repository (risk danger olson[32m[m)
[31m|[m[31m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   c21de41 2016-05-18 Merge pull request #1226 from github/refactor-1 (Steve Streeting[32m[m)
[34m|[m[35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
* [35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m   355eedc 2016-05-17 Merge pull request #1223 from ttaylorr/locking-commands (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 378aed6 2016-05-17 doc/proposal: clean up locking API's JSON (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b16406b 2016-05-17 doc/proposal: remove Remote field on Lock type (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m e4d837e 2016-05-17 doc/proposal: add CommitSHA field to Lock type (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 946e4a5 2016-05-17 doc/proposal: drop /api/v1/ prefix from locking API (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ce0abcf 2016-05-17 doc/proposal: document locking API (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f649c13 2016-05-17 doc/proposal: pagination updates (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8798f22 2016-05-17 doc/proposal: fix typo in docs (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 2a8bbf7 2016-05-17 doc/proposal: make Locks know about their Id (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 90ceafc 2016-05-17 doc/proposal: change Lock.Creator to Lock.Committer (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 376fd77 2016-05-17 doc/proposal: change ExpiresAt to UnlockedAt (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 81923e0 2016-05-16 doc/proposal: fix JSON tag name in UnlockRequest (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b601a72 2016-05-16 doc/proposal: revise minimum commit negotiation API (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m a7945e0 2016-05-16 doc/proposal: rename `lfs/lock` package to top-level `lock` (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 6864ac5 2016-05-16 doc/proposal: add partial support for minimum commit negotiation (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ee46319 2016-05-16 doc/proposal: propose locking commands and API (Taylor Blau[32m[m)
[36m|[m * [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m cb8e82a 2016-05-16 doc/proposal: clean up line-endings on locking proposal (Taylor Blau[32m[m)
[36m|[m[36m/[m [35m/[m [34m/[m [33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [35m|[m [34m|[m [33m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   7755916 2016-05-16 Merge pull request #666 from sinbad/lock-proposal (risk danger olson[32m[m)
[33m|[m[1;33m\[m [35m\[m [34m\[m [33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[33m|[m [1;33m|[m[33m_[m[35m|[m[33m_[m[34m|[m[33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [35m|[m [34m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[33m|[m * [35m|[m [34m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4b8f061 2015-11-02 Some design notes I wrote a while back when thinking about locking (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 185daf1 2016-05-27 Move to a factory pattern to transfer adapters, better safety without implicit sharing (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 4ae20c6 2016-05-27 Move smudge functions to new transfer approach (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 19d0ae5 2016-05-27 Remove unused member & config dependency (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 6df61c4 2016-05-27 Close results channel (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1b7adf2 2016-05-27 Correctly deal with nil callback / completion channels (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9115a14 2016-05-27 Configure concurrency on transfer adapters explicitly (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 41d23e9 2016-05-27 Add name context to progress callbacks (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 760d4d3 2016-05-27 Make sure all workers stop if no jobs are added (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 9f1fc60 2016-05-27 Fix race condition for auth sequence, make sure worker 0 always takes 1st job (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5e1cd25 2016-05-27 Add trace info (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 940a91a 2016-05-25 Improve temp handling (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 60fc146 2016-05-25 WIP moving transfer queue to use adapters (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 6dc11f6 2016-05-25 ClearTempStorage implementation not needed (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 71ddcce 2016-05-25 Tie jobChan lifecycle to Begin() and End() (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ac90dc7 2016-05-25 More TODO markers (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ee4fbea 2016-05-25 Finish basic adapter download, refactor common bits (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 79fe534 2016-05-25 Signal that auth was ok on first read from content during upload (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 291fb08 2016-05-25 Report progress during upload (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 30999bc 2016-05-24 TODO markers (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 0a7a905 2016-05-24 WIP basicTransferAdapter (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d39c8cf 2016-05-24 Refactor the "verify" functionality into the api package (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 94746c3 2016-05-24 Refine transfer interface (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7cb09c8 2016-05-20 Clarify that callback arguments are optional (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 37e9eb2 2016-05-20 Protect maps against concurrent access (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 868e07d 2016-05-19 Fix outdated comment (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f124f05 2016-05-19 Defining the general transfer interface TransferAdapter & registry (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1ef2d43 2016-05-19 Move CopyWithCallback into general tools for re-use (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 54e86de 2016-05-18 Make the batcher more general and not specific to Transferable (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [35m|[m[35m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ab53c53 2016-05-18 Document that our package structure is not guaranteed to be stable (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 51957f1 2016-05-17 Make script/test automatically tests all pkgs instead of needing a list (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b6c0960 2016-05-17 Fix new package tests (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m f779855 2016-05-17 Fix api tests (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 849ec82 2016-05-17 Move remaining api methods (upload/download) to api pkg (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 5b2a969 2016-05-17 Move Batch() to api pkg, generalise byteCloser (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c23f3eb 2016-05-17 Refactor new(Batch)ApiRequest into api pkg (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 1fc3ff5 2016-05-17 Refactor upload_test and download_test to not need internal lfs access & use NewRepo (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m e14f7b8 2016-05-17 Refactor remaining local store state into localstorage pkg (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 3298d9b 2016-05-16 Refactored many api functions to httputil/api package from lfs (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 633044a 2016-05-16 Rename doStorageRequest to doHttpRequestWithCreds - that's all it does (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 590fd38 2016-05-16 Move NTLM code into `auth` package (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 8b09633 2016-05-16 Rename `credentials` package to `auth` to reflect more general nature (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 7858714 2016-05-16 Move ssh auth into same package as credentials (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m ab1af11 2016-05-16 Refactor errors and credentials into own packages to break cycles (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 70c41ed 2016-05-16 Move ObjectResource, ObjectError, LinkRelation to api package (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 983be18 2016-05-16 Fix remaining uses of lfs.UserAgent outside http, use config.VersionDesc (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   aad4a30 2016-05-13 Merge branch 'master' into experimental/transfer-features (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [1;34m|[m[33m\[m [1;36m\[m [1;33m\[m [33m\[m [1;31m\[m  
[33m|[m [1;33m|[m[33m_[m[1;34m|[m[33m/[m [1;36m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [1;34m|[m [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m   
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 41b9c99 2016-05-13 Fix test server (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 2da226b 2016-05-13 Fix some integration build errors (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m c4bbd37 2016-05-13 Major refactor to pull things into config, httputil, tools (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m d731130 2016-05-10 Refactor CopyCallback/CallbackReader to progress package (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 221a1c4 2016-05-10 Refactor progress meter, spinner & log into own package (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m 333ce45 2016-05-10 Wrote up the plan for transfer adapters as proposal (Steve Streeting[32m[m)
[33m|[m [1;33m|[m * [1;36m|[m [1;33m|[m [33m|[m [1;31m|[m b0a8e55 2016-05-10 Discuss API tweaks to allow alternate upload/download protocols (Steve Streeting[32m[m)
[33m|[m [1;33m|[m [1;35m|[m * [1;33m|[m [33m|[m [1;31m|[m 09532bc 2016-05-26 ensure prune watches the verify queue before adding to it (risk danger olson[32m[m)
[33m|[m [1;33m|[m [1;35m|[m * [1;33m|[m [33m|[m [1;31m|[m d53b841 2016-05-26 watch the channel before adding any objects (risk danger olson[32m[m)
[33m|[m [1;33m|[m[33m_[m[1;35m|[m[33m/[m [1;33m/[m [33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [1;35m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;33m|[m [1;35m|[m [1;33m|[m [33m|[m [1;31m|[m   1e15699 2016-05-12 Merge pull request #1217 from github/env-print-more (Steve Streeting[32m[m)
[1;36m|[m[31m\[m [1;33m\[m [1;35m\[m [1;33m\[m [33m\[m [1;31m\[m  
[1;36m|[m * [1;33m|[m [1;35m|[m [1;33m|[m [33m|[m [1;31m|[m 601deee 2016-05-12 Add missing config details to env command (Steve Streeting[32m[m)
[1;36m|[m[1;36m/[m [1;33m/[m [1;35m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [1;33m|[m [1;35m|[m [1;33m|[m [33m|[m [1;31m|[m   e72efe6 2016-05-11 Merge pull request #1215 from javabrett/patch-2 (risk danger olson[32m[m)
[32m|[m[33m\[m [1;33m\[m [1;35m\[m [1;33m\[m [33m\[m [1;31m\[m  
[32m|[m * [1;33m|[m [1;35m|[m [1;33m|[m [33m|[m [1;31m|[m 499229a 2016-05-11 Link PR #1177 to ROADMAP,md (Brett Randall[32m[m)
[32m|[m [1;35m|[m [1;33m|[m[1;35m/[m [1;33m/[m [33m/[m [1;31m/[m  
[32m|[m [1;35m|[m[1;35m/[m[1;33m|[m [1;33m|[m [33m|[m [1;31m|[m   
* [1;35m|[m [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m 48d0fab 2016-05-11 Small fix to GitConfigBool; use parseConfigBool for better compatability (Steve Streeting[32m[m)
[1;35m|[m[1;35m/[m [1;33m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m   e4e1c5d 2016-05-10 Merge pull request #1213 from github/smudge-optional-fail (Steve Streeting[32m[m)
[34m|[m[35m\[m [1;33m\[m [1;33m\[m [33m\[m [1;31m\[m  
[34m|[m * [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m 2fa1d1c 2016-05-10 Add test for smudge exit code with/without skip download errors (Steve Streeting[32m[m)
[34m|[m * [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m 6d67a1b 2016-05-10 Add SkipDownloadErrors to env & update tests (Steve Streeting[32m[m)
[34m|[m * [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m 76fa85b 2016-05-10 Document lfs.skipdownloaderrors and GIT_LFS_SKIP_DOWNLOAD_ERRORS (Steve Streeting[32m[m)
[34m|[m * [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m f2d53ee 2016-05-10 Add config/env options to not fail smudge filter on download fail #1195 (Steve Streeting[32m[m)
* [35m|[m [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m 5cc78bc 2016-05-10 fix markdown typo (risk danger olson[32m[m)
[35m|[m[35m/[m [1;33m/[m [1;33m/[m [33m/[m [1;31m/[m  
* [1;33m|[m [1;33m|[m [33m|[m [1;31m|[m   ee59987 2016-05-06 Merge pull request #1206 from github/update-code (risk danger olson[32m[m)
[33m|[m[1;31m\[m [1;33m\[m [1;33m\[m [33m\[m [1;31m\[m  
[33m|[m [1;31m|[m[33m_[m[1;33m|[m[33m_[m[1;33m|[m[33m/[m [1;31m/[m  
[33m|[m[33m/[m[1;31m|[m [1;33m|[m [1;33m|[m [1;31m|[m   
[33m|[m * [1;33m|[m [1;33m|[m [1;31m|[m 2c4be72 2016-05-06 embed the open code of conduct since the link is bad now (risk danger olson[32m[m)
[33m|[m[33m/[m [1;33m/[m [1;33m/[m [1;31m/[m  
* [1;33m|[m [1;33m|[m [1;31m|[m   f748d59 2016-05-04 Merge pull request #1198 from teo-tsirpanis/patch-1 (Steve Streeting[32m[m)
[1;32m|[m[1;33m\[m [1;33m\[m [1;33m\[m [1;31m\[m  
[1;32m|[m * [1;33m|[m [1;33m|[m [1;31m|[m 417836c 2016-05-04 Fix installer error on win32. (Theodore Tsirpanis[32m[m)
* [1;33m|[m [1;33m|[m [1;33m|[m [1;31m|[m   8cb192c 2016-05-04 Merge pull request #1182 from github/update-manual (Steve Streeting[32m[m)
[1;33m|[m[1;35m\[m [1;33m\[m [1;33m\[m [1;33m\[m [1;31m\[m  
[1;33m|[m [1;35m|[m[1;33m/[m [1;33m/[m [1;33m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m   
[1;33m|[m * [1;33m|[m [1;33m|[m [1;31m|[m 5be442d 2016-04-26 Add `git lfs update --manual` option & promote it on hook install fail (Steve Streeting[32m[m)
* [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m   c923d50 2016-05-04 Merge pull request #1179 from github/transfer-mutex (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;33m\[m [1;31m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m 982b1f4 2016-04-22 fix concurrent map read and map write (risk danger olson[32m[m)
[1;36m|[m [1;31m|[m [1;35m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;33m|[m[1;31m/[m  
[1;36m|[m [1;31m|[m[1;31m/[m[1;35m|[m [1;33m|[m [1;33m|[m   
* [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   a24530e 2016-05-04 Merge pull request #1193 from javabrett/script-run-ldflags-X-equals (risk danger olson[32m[m)
[32m|[m[33m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m d344012 2016-05-03 Applied same -ldflags -X name value -> name=value fix as in a62e510f. (Brett Randall[32m[m)
[32m|[m[32m/[m [1;31m/[m [1;35m/[m [1;33m/[m [1;33m/[m  
* [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   54c5efa 2016-04-28 Merge pull request #1185 from github/lfs-clone-user-prompts (Steve Streeting[32m[m)
[34m|[m[35m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m 728f013 2016-04-27 Fix problems with user prompts in `git lfs clone` (Steve Streeting[32m[m)
[34m|[m [1;35m|[m [1;31m|[m[1;35m/[m [1;33m/[m [1;33m/[m  
[34m|[m [1;35m|[m[1;35m/[m[1;31m|[m [1;33m|[m [1;33m|[m   
* [1;35m|[m [1;31m|[m [1;33m|[m [1;33m|[m   9b5c2ff 2016-04-27 Merge pull request #1186 from skymoo/macports (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [1;35m\[m [1;31m\[m [1;33m\[m [1;33m\[m  
[1;35m|[m [1;31m|[m[1;35m/[m [1;31m/[m [1;33m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;31m|[m [1;33m|[m [1;33m|[m   
[1;35m|[m * [1;31m|[m [1;33m|[m [1;33m|[m a0aa683 2016-04-25 add instructions to install from MacPorts (Adam Mercer[32m[m)
[1;35m|[m[1;35m/[m [1;31m/[m [1;33m/[m [1;33m/[m  
* [1;31m|[m [1;33m|[m [1;33m|[m   7271269 2016-04-25 Merge pull request #1178 from github/update-install-exit-code (Steve Streeting[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;33m\[m [1;33m\[m  
[1;32m|[m * [1;31m|[m [1;33m|[m [1;33m|[m 6b3629e 2016-04-25 Replace blank hooks without warning (Steve Streeting[32m[m)
[1;32m|[m * [1;31m|[m [1;33m|[m [1;33m|[m 5a081f4 2016-04-25 Fix install test, "Git LFS initialized" now omitted on hook failure (Steve Streeting[32m[m)
[1;32m|[m * [1;31m|[m [1;33m|[m [1;33m|[m 244e8bd 2016-04-25 Move hook install to after config sanitize so that always happens (Steve Streeting[32m[m)
[1;32m|[m * [1;31m|[m [1;33m|[m [1;33m|[m 7d3ec72 2016-04-25 Return non-zero exit code for update & install when hook not updated (Steve Streeting[32m[m)
[1;32m|[m * [1;31m|[m [1;33m|[m [1;33m|[m 5ac6e5d 2016-04-25 Create failing tests (Steve Streeting[32m[m)
[1;32m|[m [1;31m|[m[1;31m/[m [1;33m/[m [1;33m/[m  
* [1;31m|[m [1;33m|[m [1;33m|[m   cd75231 2016-04-25 Merge pull request #1160 from github/clone-flags (Steve Streeting[32m[m)
[1;31m|[m[1;35m\[m [1;31m\[m [1;33m\[m [1;33m\[m  
[1;31m|[m [1;35m|[m[1;31m/[m [1;33m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m [1;33m|[m [1;33m|[m   
[1;31m|[m * [1;33m|[m [1;33m|[m 2000c7a 2016-04-19 Add tests for cloning with flags (Steve Streeting[32m[m)
[1;31m|[m * [1;33m|[m [1;33m|[m f906e06 2016-04-19 LFS clone should work when --origin is used to change remote name (Steve Streeting[32m[m)
[1;31m|[m * [1;33m|[m [1;33m|[m 142785e 2016-04-19 Also use fetch instead of pull when --bare is used to clone (Steve Streeting[32m[m)
[1;31m|[m * [1;33m|[m [1;33m|[m b6b4547 2016-04-19 If `--no-checkout` flag set on `git lfs clone`, fetch not pull (Steve Streeting[32m[m)
[1;31m|[m * [1;33m|[m [1;33m|[m 66eeea4 2016-04-19 Add global inherited flags (Steve Streeting[32m[m)
[1;31m|[m * [1;33m|[m [1;33m|[m 6bd7666 2016-04-19 Support all `git clone` flags in `git lfs clone` #1155 (Steve Streeting[32m[m)
* [1;35m|[m [1;33m|[m [1;33m|[m   110d58f 2016-04-22 Merge pull request #1170 from graingert/patch-1 (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m d662e09 2016-04-22 Add xenial repo (Thomas Grainger[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   0565f77 2016-04-22 Merge pull request #1168 from github/roadmap-install (Steve Streeting[32m[m)
[31m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[31m|[m [33m|[m[31m/[m [1;35m/[m [1;33m/[m [1;33m/[m  
[31m|[m[31m/[m[33m|[m [1;35m|[m [1;33m|[m [1;33m|[m   
[31m|[m * [1;35m|[m [1;33m|[m [1;33m|[m 8271b11 2016-04-22 Roadmap item for warning about lack of `git lfs install` (Steve Streeting[32m[m)
[31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;33m/[m  
* [1;35m|[m [1;33m|[m [1;33m|[m   34ae617 2016-04-22 Merge pull request #1158 from github/technoweenie-patch-1 (Steve Streeting[32m[m)
[34m|[m[35m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [1;35m|[m [1;33m|[m [1;33m|[m e146c81 2016-04-18 Update ROADMAP.md (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;33m|[m [1;33m|[m 2c1176d 2016-04-18 More updates! (risk danger olson[32m[m)
[34m|[m [1;35m|[m[1;35m/[m [1;33m/[m [1;33m/[m  
* [1;35m|[m [1;33m|[m [1;33m|[m   db69cbd 2016-04-20 Merge pull request #1162 from larsxschneider/patch-1 (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[1;35m|[m [1;31m|[m[1;35m/[m [1;33m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;33m|[m [1;33m|[m   
[1;35m|[m * [1;33m|[m [1;33m|[m 013413a 2016-04-20 fix typo (Lars Schneider[32m[m)
[1;35m|[m[1;35m/[m [1;33m/[m [1;33m/[m  
* [1;33m|[m [1;33m|[m   884677e 2016-04-18 Merge pull request #1149 from javabrett/include-man-git-lfs-config-5 (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;33m\[m [1;33m\[m  
[1;32m|[m * [1;33m|[m [1;33m|[m b9ae764 2016-04-15 Fixed #719 missing /usr/share/man/man5/git-lfs-config.5.gz . (Brett Randall[32m[m)
* [1;33m|[m [1;33m|[m [1;33m|[m   33de6ac 2016-04-18 Merge pull request #1154 from larsxschneider/patch-1 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m * [1;33m|[m [1;33m|[m [1;33m|[m 73271c1 2016-04-18 add homebrew update to release process (Lars Schneider[32m[m)
[1;34m|[m[1;34m/[m [1;33m/[m [1;33m/[m [1;33m/[m  
* [1;33m|[m [1;33m|[m [1;33m|[m   896471a 2016-04-15 Merge pull request #1145 from github/technoweenie-patch-1 (risk danger olson[32m[m)
[1;33m|[m[31m\[m [1;33m\[m [1;33m\[m [1;33m\[m  
[1;33m|[m [31m|[m[1;33m/[m [1;33m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[31m|[m [1;33m|[m [1;33m|[m   
[1;33m|[m * [1;33m|[m [1;33m|[m 3c05844 2016-04-15 another one (risk danger olson[32m[m)
[1;33m|[m * [1;33m|[m [1;33m|[m f9faa40 2016-04-14 Update ROADMAP.md (risk danger olson[32m[m)
[1;33m|[m * [1;33m|[m [1;33m|[m c1970b6 2016-04-14 added some small things (risk danger olson[32m[m)
* [31m|[m [1;33m|[m [1;33m|[m 386c5d8 2016-04-14 add missing cert behavior for freebsd (risk danger olson[32m (tag: v1.2.0)[m)
* [31m|[m [1;33m|[m [1;33m|[m   9bd3b8e 2016-04-14 Merge pull request #1143 from strich/feature/win-inst-fix-lfs-paths (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;33m|[m [1;33m|[m 4c72dc4 2016-04-13 Fixed an issue where the newly added Git LFS PATH location wasn't refreshed in time for use later in the installer script and silently failed to initialize Git LFS. If for any reason there is an issue running 'git lfs install', the installer will display a message box asking the user to manually do so. Added a GIT_LFS_PATH global env var for later use. (Scott Richmond[32m[m)
* [33m|[m [31m|[m [1;33m|[m [1;33m|[m 7472995 2016-04-14 remove <br/> usage (risk danger olson[32m[m)
* [33m|[m [31m|[m [1;33m|[m [1;33m|[m   2109397 2016-04-14 Merge pull request #1076 from javabrett/packagecloud-linuxmint (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [31m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [33m|[m [31m|[m [1;33m|[m [1;33m|[m 103c0fa 2016-03-31 Removed hard-coded "github" user reference. (Brett Randall[32m[m)
[34m|[m * [33m|[m [31m|[m [1;33m|[m [1;33m|[m 11ba337 2016-03-31 Added Linux Mint distro to packagecloud push, issue #1074. (Brett Randall[32m[m)
* [35m|[m [33m|[m [31m|[m [1;33m|[m [1;33m|[m 16a3c61 2016-04-14 bump debian version too (risk danger olson[32m[m)
[1;33m|[m [35m|[m[1;33m_[m[33m|[m[1;33m_[m[31m|[m[1;33m_[m[1;33m|[m[1;33m/[m  
[1;33m|[m[1;33m/[m[35m|[m [33m|[m [31m|[m [1;33m|[m   
* [35m|[m [33m|[m [31m|[m [1;33m|[m c5d7a89 2016-04-14 small typos (risk danger olson[32m[m)
* [35m|[m [33m|[m [31m|[m [1;33m|[m   4f53a1a 2016-04-14 Merge branch 'installing-doc' of https://github.com/javabrett/git-lfs into javabrett-installing-doc (risk danger olson[32m[m)
[31m|[m[1;31m\[m [35m\[m [33m\[m [31m\[m [1;33m\[m  
[31m|[m [1;31m|[m[31m_[m[35m|[m[31m_[m[33m|[m[31m/[m [1;33m/[m  
[31m|[m[31m/[m[1;31m|[m [35m|[m [33m|[m [1;33m|[m   
[31m|[m * [35m|[m [33m|[m [1;33m|[m 4b2d5f6 2016-04-07 Documentation as part of #1074 - added INSTALLING.md to clarify aspects of packagecloud installs and linked from README.md. (Brett Randall[32m[m)
* [1;31m|[m [35m|[m [33m|[m [1;33m|[m 16fa138 2016-04-14 Changelog & version bump for v1.2.0 (Steve Streeting[32m[m)
* [1;31m|[m [35m|[m [33m|[m [1;33m|[m   56ee188 2016-04-14 Merge pull request #1144 from github/netrc-no-port (risk danger olson[32m[m)
[33m|[m[1;33m\[m [1;31m\[m [35m\[m [33m\[m [1;33m\[m  
[33m|[m [1;33m|[m[33m_[m[1;31m|[m[33m_[m[35m|[m[33m/[m [1;33m/[m  
[33m|[m[33m/[m[1;33m|[m [1;31m|[m [35m|[m [1;33m|[m   
[33m|[m * [1;31m|[m [35m|[m [1;33m|[m f820622 2016-04-13 fix netrc matching when host has no port (risk danger olson[32m (chris/netrc-no-port)[m)
[33m|[m[33m/[m [1;31m/[m [35m/[m [1;33m/[m  
* [1;31m|[m [35m|[m [1;33m|[m 44903ba 2016-04-13 Add missing 'clone' and 'version' commands to 'git lfs help' (Steve Streeting[32m[m)
* [1;31m|[m [35m|[m [1;33m|[m   fd2d215 2016-04-12 Merge pull request #1127 from github/roadmap-post-v1 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [35m\[m [1;33m\[m  
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m ded0285 2016-04-11 another missed item from 1.2 (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m 05df09d 2016-04-11 more stuff (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m d79aa5a 2016-04-11 Update ROADMAP.md (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m 90b59e8 2016-04-11 Added pull/checkout wrapping, non-batch removal, ssh (Steve Streeting[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m 592f89a 2016-04-06 add some bugs (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m [1;33m|[m b9f81cd 2016-04-06 clear the roadmap of implemented stuff (risk danger olson[32m[m)
[1;34m|[m [1;31m|[m[1;31m/[m [35m/[m [1;33m/[m  
* [1;31m|[m [35m|[m [1;33m|[m   582cc6c 2016-04-12 Merge pull request #1136 from github/add-tests-for-including-credentials-in-remote-or-lfs-url (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [35m\[m [1;33m\[m  
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m 12c8679 2016-04-11 be more specific about the allowed failure (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m 905ebba 2016-04-11 allow those hanging credential tests to fail until we can nail down a fix (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m caee1d5 2016-04-11 send output before killing (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m 676611d 2016-04-11 add timeouts to script/integration, with partial output (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m 2f533e9 2016-04-08 use getCredsForAPI() for everything (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m a1d0ba7 2016-04-08 update credential tests with good and bad tests for pushes and fetches (risk danger olson[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [1;33m|[m 5d076c5 2016-04-08 Add tests showing that credentials in lfs.url or remote url are used (Jonathan Hoyt[32m[m)
[1;36m|[m [1;31m|[m[1;31m/[m [35m/[m [1;33m/[m  
* [1;31m|[m [35m|[m [1;33m|[m   02f5ae5 2016-04-11 Merge pull request #1134 from github/ssh-strict-arguments (risk danger olson[32m[m)
[32m|[m[33m\[m [1;31m\[m [35m\[m [1;33m\[m  
[32m|[m * [1;31m|[m [35m|[m [1;33m|[m 43c1f5c 2016-04-08 SSH should be called more strictly with command as one argument (Steve Streeting[32m[m)
* [33m|[m [1;31m|[m [35m|[m [1;33m|[m   f2b07b2 2016-04-08 Merge pull request #1135 from jonmagic/jonmagic-patch-1 (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;31m\[m [35m\[m [1;33m\[m  
[34m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m 4d44ed6 2016-04-08 You should add the .gitattributes file before committing (Jonathan Hoyt[32m[m)
* [35m|[m [33m|[m [1;31m|[m [35m|[m [1;33m|[m   0ddb6f4 2016-04-08 Merge pull request #1007 from jlehtnie/clone-reference (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [33m\[m [1;31m\[m [35m\[m [1;33m\[m  
[35m|[m [1;31m|[m[35m/[m [33m/[m [1;31m/[m [35m/[m [1;33m/[m  
[35m|[m[35m/[m[1;31m|[m [33m|[m [1;31m|[m [35m|[m [1;33m|[m   
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m ec58694 2016-04-05 Review fixes (Jukka Lehtniemi[32m[m)
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m bfb28d2 2016-04-05 Safety fixes based on review feedback (Jukka Lehtniemi[32m[m)
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m 4450744 2016-04-05 Review fix: copy from reference in command_smudge (Jukka Lehtniemi[32m[m)
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m ec21903 2016-04-05 Style fixes based on review feedback (Jukka Lehtniemi[32m[m)
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m 67b6d7b 2016-04-05 Refactor based on review feedback: keep it dry (Jukka Lehtniemi[32m[m)
[35m|[m * [33m|[m [1;31m|[m [35m|[m [1;33m|[m 5069ecd 2016-04-05 Add support for reference clone (Jukka Lehtniemi[32m[m)
* [1;31m|[m [33m|[m [1;31m|[m [35m|[m [1;33m|[m   3598328 2016-04-08 Merge pull request #1128 from github/skip-pre-push-dupes-2 (risk danger olson[32m[m)
[33m|[m[1;33m\[m [1;31m\[m [33m\[m [1;31m\[m [35m\[m [1;33m\[m  
[33m|[m [1;33m|[m[33m_[m[1;31m|[m[33m/[m [1;31m/[m [35m/[m [1;33m/[m  
[33m|[m[33m/[m[1;33m|[m [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m   
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m be83f64 2016-04-08 rename AddUpload() to SetUploaded() for clarity (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 160ad66 2016-04-07 these assertions should be more specific (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 231686e 2016-04-07 refactor *uploadContext, separate lib responsibilities vs cli responsibilities (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 0935817 2016-04-07 no need for defer since there's only 1 return after cmd.Start() (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m abd9b1e 2016-04-06 introduce an uploadContext struct to share in push and pre-push (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 31dc0c2 2016-04-06 add ResolveRefs() placeholder for future optimization (risk danger olson[32m[m)
[33m|[m * [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 19c0e6c 2016-04-06 teach git package how to get the local refs (risk danger olson[32m[m)
[33m|[m [1;31m|[m [1;31m|[m[1;31m/[m [35m/[m [1;33m/[m  
[33m|[m [1;31m|[m[1;31m/[m[1;31m|[m [35m|[m [1;33m|[m   
* [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m 276892b 2016-04-08 Fix pointer test on Linux (Steve Streeting[32m[m)
* [1;31m|[m [1;31m|[m [35m|[m [1;33m|[m   a65e224 2016-04-07 Merge branch 'sschuberth-master' (risk danger olson[32m[m)
[1;31m|[m[1;35m\[m [1;31m\[m [1;31m\[m [35m\[m [1;33m\[m  
[1;31m|[m [1;35m|[m[1;31m/[m [1;31m/[m [35m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m [1;31m|[m [35m|[m [1;33m|[m   
[1;31m|[m * [1;31m|[m [35m|[m [1;33m|[m 73f1af9 2016-04-07 add tests checking the pointer command's output (risk danger olson[32m[m)
[1;31m|[m * [1;31m|[m [35m|[m [1;33m|[m dcf9249 2016-04-07 fix nasty test bug (risk danger olson[32m[m)
[1;31m|[m * [1;31m|[m [35m|[m [1;33m|[m   2f9bd14 2016-04-07 Merge branch 'master' of https://github.com/sschuberth/git-lfs into sschuberth-master (risk danger olson[32m[m)
[1;31m|[m [1;31m|[m[31m\[m [1;31m\[m [35m\[m [1;33m\[m  
[1;31m|[m[1;31m/[m [31m/[m [1;31m/[m [35m/[m [1;33m/[m  
[1;31m|[m * [1;31m|[m [35m|[m [1;33m|[m da2935d 2016-03-22 pointer: Only write the encoded pointer information to Stdout (Sebastian Schuberth[32m[m)
* [31m|[m [1;31m|[m [35m|[m [1;33m|[m   3671987 2016-04-06 Merge pull request #1122 from rjbell4/allowaccess (risk danger olson[32m[m)
[1;31m|[m[33m\[m [31m\[m [1;31m\[m [35m\[m [1;33m\[m  
[1;31m|[m [33m|[m[1;31m_[m[31m|[m[1;31m/[m [35m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[33m|[m [31m|[m [35m|[m [1;33m|[m   
[1;31m|[m * [31m|[m [35m|[m [1;33m|[m dc4dc8e 2016-04-05 Allow for an access mechanism to be specified in the .lfsconfig file (Bob Bell[32m[m)
* [33m|[m [31m|[m [35m|[m [1;33m|[m 95272fe 2016-04-05 Fix concurrent map issue accessing Configuration.envVars (Steve Streeting[32m[m)
* [33m|[m [31m|[m [35m|[m [1;33m|[m   825cf2e 2016-04-05 Merge pull request #1118 from github/scanner-async-errors (Steve Streeting[32m[m)
[34m|[m[35m\[m [33m\[m [31m\[m [35m\[m [1;33m\[m  
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 5563845 2016-04-05 Add test for delete branch in pre-push (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m ae0af55 2016-04-01 Don't quote ref in ambiguous message, already includes quotes (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m b6f24ad 2016-04-01 Outdated comment (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 23b23cd 2016-03-31 Catch specific case of ambiguous refs in revListShas (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 0bca669 2016-03-31 Fix same pre-push delete branch bug in push (use common var) (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 10a4fb8 2016-03-31 Fix old bug in pre-push hook; would never detect delete branch calls (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 5ad196f 2016-03-31 Disable exit code checking on `git diff-index` to avoid breaking fsck (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 53442b2 2016-03-31 Include more detail in command errors to banish dreaded 'exit code 128' (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m d9751df 2016-03-31 Fix comments (Steve Streeting[32m[m)
[34m|[m * [33m|[m [31m|[m [35m|[m [1;33m|[m 2391896 2016-03-31 Refactor all scanner-related functions to wrap channels & report async errors (Steve Streeting[32m[m)
[34m|[m [35m|[m [33m|[m[35m_[m[31m|[m[35m/[m [1;33m/[m  
[34m|[m [35m|[m[35m/[m[33m|[m [31m|[m [1;33m|[m   
* [35m|[m [33m|[m [31m|[m [1;33m|[m   65f53a6 2016-04-05 Merge pull request #1104 from github/touch-tracked-files (Steve Streeting[32m[m)
[33m|[m[1;31m\[m [35m\[m [33m\[m [31m\[m [1;33m\[m  
[33m|[m [1;31m|[m[33m_[m[35m|[m[33m/[m [31m/[m [1;33m/[m  
[33m|[m[33m/[m[1;31m|[m [35m|[m [31m|[m [1;33m|[m   
[33m|[m * [35m|[m [31m|[m [1;33m|[m 70e5df6 2016-04-04 wait for the command to finish (risk danger olson[32m[m)
[33m|[m * [35m|[m [31m|[m [1;33m|[m 875b4de 2016-03-22 'git lfs track' now automatically touches matching files already in git (Steve Streeting[32m[m)
[33m|[m * [35m|[m [31m|[m [1;33m|[m 21f3812 2016-03-22 Add git.GetTrackedFiles for reporting list of files tracked in git (Steve Streeting[32m[m)
* [1;31m|[m [35m|[m [31m|[m [1;33m|[m   9c7155d 2016-03-31 Merge pull request #1026 from github/remove-old-build-scripts (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m [31m\[m [1;33m\[m  
[1;32m|[m * [1;31m|[m [35m|[m [31m|[m [1;33m|[m 4375e6d 2016-02-23 these scripts are unmaintained (risk danger olson[32m[m)
* [1;33m|[m [1;31m|[m [35m|[m [31m|[m [1;33m|[m 5067543 2016-03-31 fix changelog date abbrev (risk danger olson[32m[m)
* [1;33m|[m [1;31m|[m [35m|[m [31m|[m [1;33m|[m   c88674c 2016-03-31 Merge tag 'v1.1.2' (risk danger olson[32m[m)
[35m|[m[1;35m\[m [1;33m\[m [1;31m\[m [35m\[m [31m\[m [1;33m\[m  
[35m|[m [1;35m|[m[35m_[m[1;33m|[m[35m_[m[1;31m|[m[35m/[m [31m/[m [1;33m/[m  
[35m|[m[35m/[m[1;35m|[m [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m bf6a6c9 2016-03-01 v1.1.2 (risk danger olson[32m (tag: v1.1.2)[m)
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m ed38201 2016-03-01 ignore another connection error (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m d79d479 2016-03-01 ignore docker related connection errors (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   10b7575 2016-02-24 Merge pull request #1042 from github/more-zombies (risk danger olson[32m[m)
[35m|[m [1;36m|[m[31m\[m [1;33m\[m [1;31m\[m [31m\[m [1;33m\[m  
[35m|[m [1;36m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m 5a8dbb8 2016-02-24 defer a couple more cmd.Wait() calls (risk danger olson[32m[m)
[35m|[m [1;36m|[m[1;36m/[m [1;33m/[m [1;31m/[m [31m/[m [1;33m/[m  
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   cba4d00 2016-02-24 Merge pull request #1039 from github/1.1/more-better-errors (risk danger olson[32m[m)
[35m|[m [32m|[m[33m\[m [1;33m\[m [1;31m\[m [31m\[m [1;33m\[m  
[35m|[m [32m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m d52b983 2016-02-24 better handling on push errors (risk danger olson[32m[m)
[35m|[m [32m|[m[32m/[m [1;33m/[m [1;31m/[m [31m/[m [1;33m/[m  
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   4b3b9d5 2016-02-23 Merge pull request #1030 from github/release-1.1-backport-1025 (risk danger olson[32m[m)
[35m|[m [34m|[m[35m\[m [1;33m\[m [1;31m\[m [31m\[m [1;33m\[m  
[35m|[m [34m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m 17db7ed 2016-02-23 Backport go-1.6 from #1025 to release-1.1 (risk danger olson[32m[m)
[35m|[m [34m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m e232b3d 2016-02-23 update test so it passes regardless of git version or batch api (risk danger olson[32m[m)
[35m|[m [34m|[m[34m/[m [1;33m/[m [1;31m/[m [31m/[m [1;33m/[m  
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m e862509 2016-02-23 remove concurrent map access issue (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   00111e7 2016-02-23 Merge pull request #1028 from github/release-1.1-backport-1023 (risk danger olson[32m[m)
[35m|[m [36m|[m[1;31m\[m [1;33m\[m [1;31m\[m [31m\[m [1;33m\[m  
[35m|[m [36m|[m * [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m 6812d0b 2016-02-23 Backport better-errors from #1023 to release-1.1 (risk danger olson[32m[m)
[35m|[m [36m|[m[36m/[m [1;33m/[m [1;31m/[m [31m/[m [1;33m/[m  
* [36m|[m [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m 0265e27 2016-03-30 Usage doc fix, docker images are called centos_x not lfs_centos_x (Steve Streeting[32m[m)
* [36m|[m [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m   9ef0b45 2016-03-30 Merge pull request #1102 from javabrett/report-packagecloud-api-errors (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [1;33m\[m [1;31m\[m [31m\[m [1;33m\[m  
[1;32m|[m * [36m|[m [1;33m|[m [1;31m|[m [31m|[m [1;33m|[m 2e0b159 2016-03-22 Report errors generated by packagecloud API put_package and exit on failure. (Brett Randall[32m[m)
[1;32m|[m [31m|[m [36m|[m[31m_[m[1;33m|[m[31m_[m[1;31m|[m[31m/[m [1;33m/[m  
[1;32m|[m [31m|[m[31m/[m[36m|[m [1;33m|[m [1;31m|[m [1;33m|[m   
* [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m   cf637d0 2016-03-30 Merge pull request #1115 from javabrett/debian-dpkg-go-1.6 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [36m\[m [1;33m\[m [1;31m\[m [1;33m\[m  
[1;34m|[m * [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m 0b9609e 2016-03-30 Added GO15VENDOREXPERIMENT=0 to dpkg dh build, follow-on from #1025. (Brett Randall[32m[m)
[1;34m|[m[1;34m/[m [31m/[m [36m/[m [1;33m/[m [1;31m/[m [1;33m/[m  
* [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m cc41ca5 2016-03-24 Fix auto tempdir for GNU mktemp, needs X's (BSD doesn't) (Steve Streeting[32m[m)
* [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m 50b0602 2016-03-24 Be consistent in error reporting from bufferDownloadedFile result (Steve Streeting[32m[m)
* [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m   3117af8 2016-03-23 Merge pull request #1107 from github/automatic-git-lfs-test-dir (Steve Streeting[32m[m)
[1;36m|[m[31m\[m [31m\[m [36m\[m [1;33m\[m [1;31m\[m [1;33m\[m  
[1;36m|[m * [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m e1c2c8a 2016-03-18 Always run integration tests in a separate GIT_LFS_TEST_DIR (Steve Streeting[32m[m)
* [31m|[m [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m   709a471 2016-03-23 Merge pull request #1091 from github/symlink-issues (Steve Streeting[32m[m)
[32m|[m[31m\[m [31m\[m [31m\[m [36m\[m [1;33m\[m [1;31m\[m [1;33m\[m  
[32m|[m [31m|[m[31m/[m [31m/[m [36m/[m [1;33m/[m [1;31m/[m [1;33m/[m  
[32m|[m * [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m 3d41190 2016-03-18 Skip blank paths in ResolveSymlinks (Steve Streeting[32m[m)
[32m|[m * [31m|[m [36m|[m [1;33m|[m [1;31m|[m [1;33m|[m 6da7fe8 2016-03-18 Fully resolve symlinks in all cases for consistency (Steve Streeting[32m[m)
[32m|[m [1;31m|[m [31m|[m[1;31m_[m[36m|[m[1;31m_[m[1;33m|[m[1;31m/[m [1;33m/[m  
[32m|[m [1;31m|[m[1;31m/[m[31m|[m [36m|[m [1;33m|[m [1;33m|[m   
* [1;31m|[m [31m|[m [36m|[m [1;33m|[m [1;33m|[m   9a7cfed 2016-03-22 Merge pull request #1093 from anatolyborodin/contributing-releng-access (Steve Streeting[32m[m)
[34m|[m[35m\[m [1;31m\[m [31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [1;31m|[m [31m|[m [36m|[m [1;33m|[m [1;33m|[m 1d6ddee 2016-03-19 Mention releng access rights to the git-lfs.github.com repository (Anatoly Borodin[32m[m)
[34m|[m [31m|[m [1;31m|[m[31m/[m [36m/[m [1;33m/[m [1;33m/[m  
[34m|[m [31m|[m[31m/[m[1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   
* [31m|[m [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   95ccb90 2016-03-22 Merge pull request #1094 from anatolyborodin/pat-your-self (Steve Streeting[32m[m)
[36m|[m[1;31m\[m [31m\[m [1;31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[36m|[m * [31m|[m [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m b4fc7f7 2016-03-19 Replace "pat your self" with "pat yourself" (Anatoly Borodin[32m[m)
[36m|[m [31m|[m[31m/[m [1;31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [31m|[m [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   163c07e 2016-03-22 Merge pull request #1096 from epriestley/rev-list-ambiguous (Steve Streeting[32m[m)
[31m|[m[1;33m\[m [31m\[m [1;31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[31m|[m [1;33m|[m[31m/[m [1;31m/[m [36m/[m [1;33m/[m [1;33m/[m  
[31m|[m[31m/[m[1;33m|[m [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   
[31m|[m * [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m ea639a5 2016-03-19 Fix `git rev-list` when a file exists with the same name as the branch (epriestley[32m[m)
[31m|[m[31m/[m [1;31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   8f90f59 2016-03-18 Merge pull request #1087 from epriestley/batch-upload-download (Steve Streeting[32m[m)
[1;31m|[m[1;35m\[m [1;31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;31m|[m [1;35m|[m[1;31m/[m [36m/[m [1;33m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m [36m|[m [1;33m|[m [1;33m|[m   
[1;31m|[m * [36m|[m [1;33m|[m [1;33m|[m 0c02383 2016-03-18 Correct the batch upload/download skip detection documentation (epriestley[32m[m)
[1;31m|[m[1;31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [36m|[m [1;33m|[m [1;33m|[m   15592a4 2016-03-18 Merge pull request #1081 from github/sslverify-per-host (Steve Streeting[32m[m)
[1;36m|[m[31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;36m|[m * [36m|[m [1;33m|[m [1;33m|[m 3a59d82 2016-03-17 Support gitconfig sslverify for specific host (Steve Streeting[32m[m)
* [31m|[m [36m|[m [1;33m|[m [1;33m|[m   91937fd 2016-03-17 Merge pull request #1084 from epriestley/swapped-trace-string (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [31m|[m [36m|[m [1;33m|[m [1;33m|[m 9fe715f 2016-03-17 Fix "Uploading refs <remote> to remote <refs>" log string (epriestley[32m[m)
[32m|[m [31m|[m[31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [31m|[m [36m|[m [1;33m|[m [1;33m|[m   890f4ea 2016-03-17 Merge pull request #1083 from epriestley/unusual-filenames2 (risk danger olson[32m[m)
[34m|[m[35m\[m [31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [31m|[m [36m|[m [1;33m|[m [1;33m|[m fa2ec2f 2016-03-17 Fix smudge/clean filters to work with filenames that have leading dashes (epriestley[32m[m)
[34m|[m * [31m|[m [36m|[m [1;33m|[m [1;33m|[m 1577510 2016-03-17 Add a failing test for filenames starting with dashes (epriestley[32m[m)
[34m|[m [31m|[m[31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [31m|[m [36m|[m [1;33m|[m [1;33m|[m   c3c7e90 2016-03-17 Merge pull request #1082 from epriestley/patch-2 (risk danger olson[32m[m)
[31m|[m[1;31m\[m [31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[31m|[m [1;31m|[m[31m/[m [36m/[m [1;33m/[m [1;33m/[m  
[31m|[m[31m/[m[1;31m|[m [36m|[m [1;33m|[m [1;33m|[m   
[31m|[m * [36m|[m [1;33m|[m [1;33m|[m 5e2e61c 2016-03-17 Explain a subtlety of client handling for "download" batch action (Evan Priestley[32m[m)
[31m|[m[31m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [36m|[m [1;33m|[m [1;33m|[m   b419488 2016-03-17 Merge pull request #1080 from github/refactor-tty-to-subprocess (Steve Streeting[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;32m|[m * [36m|[m [1;33m|[m [1;33m|[m e0972e4 2016-03-17 Move TTY wrapper to subprocess package, makes more sense now it exists (Steve Streeting[32m[m)
[1;32m|[m[1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [36m|[m [1;33m|[m [1;33m|[m   cb88941 2016-03-17 Merge pull request #1067 from sinbad/self-signed-certs (Steve Streeting[32m[m)
[1;34m|[m[1;35m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m * [36m\[m [1;33m\[m [1;33m\[m   c2095b9 2016-03-17 Merge master into self-signed-certs to resolve conflicts (Steve Streeting[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [36m/[m [1;33m/[m [1;33m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [36m|[m [1;33m|[m [1;33m|[m   
* [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m   1ca7615 2016-03-16 Merge pull request #1070 from sinbad/pty_windows_compat (Steve Streeting[32m[m)
[32m|[m[33m\[m [1;36m\[m [36m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m f1e87af 2016-03-15 Fix netrc tests on Windows (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m e8deb6e 2016-03-15 Skip track absolute test on Windows, MinGW bash doesn't pass rooted correctly (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m c8a6320 2016-03-14 Fix running integration tests on Windows (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m d52eb78 2016-03-14 Fix git lfs clone on Windows; Stdout shouldn't be set AND copied (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m 27fb43b 2016-03-14 Fix Windows build; +build directive needs blank line after (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m 66ec2c8 2016-03-14 Fix Linux stall, close tty not pty (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m a6f9120 2016-03-11 Copy/paste error (still works due to other code but wrong) (Steve Streeting[32m[m)
[32m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m b01297c 2016-03-11 Fix windows build; referencing kr/pty breaks it, hide in extra layer (Steve Streeting[32m[m)
* [33m|[m [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m   7cca781 2016-03-15 Merge pull request #1078 from epriestley/patch-1 (risk danger olson[32m[m)
[33m|[m[35m\[m [33m\[m [1;36m\[m [36m\[m [1;33m\[m [1;33m\[m  
[33m|[m [35m|[m[33m/[m [1;36m/[m [36m/[m [1;33m/[m [1;33m/[m  
[33m|[m[33m/[m[35m|[m [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m   
[33m|[m * [1;36m|[m [36m|[m [1;33m|[m [1;33m|[m 04e5666 2016-03-15 Fix an alignment issue in batch API documentation (Evan Priestley[32m[m)
[33m|[m[33m/[m [1;36m/[m [36m/[m [1;33m/[m [1;33m/[m  
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m d2c1c0f 2016-03-17 Skip SSL integration tests on Travis (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 3824349 2016-03-16 Export the golang httptest cert as PEM data for better compatibility with git (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 745b9df 2016-03-14 Wait for cert file (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m b6f74b6 2016-03-11 Protect http client cache with mutex (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 3329f1d 2016-03-11 Cache http client per host (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 19a2e9b 2016-03-10 Simplify appendRootCAsForHostFromGitconfig (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 8f91a1b 2016-03-10 Fix tests (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 452198a 2016-03-10 Support binary certs (ASN.1 DER) as well as PEM data (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 720fca6 2016-03-09 Set ssl cert globally so all clones can pick it up easily (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 2872d1b 2016-03-09 Set http.sslcainfo in gitconfig after clone_repo_ssl (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 77cdf5c 2016-03-09 Don't switch to localhost, golang cert is actually valid for 127.0.0.1 (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 4c64e82 2016-03-09 Trying to get SSL integration test working, having cert issues (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m f793172 2016-03-09 Tests for gitconfig & env cases of sslcainfo/sslcapath (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 0ef7ed3 2016-03-08 Fix signatures on Linux/Windows specific builds (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m aa7babf 2016-03-08 Rework certs, only return custom root CAs if we find matching for host (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m 6f62801 2016-03-08 Temp WIP - thinking of doing this differently (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m a0adc82 2016-03-08 Support for SSL certs in login & system keychains on Mac (Steve Streeting[32m[m)
[33m|[m * [36m|[m [1;33m|[m [1;33m|[m d27a09e 2016-03-08 Refactor execCommand and simpleExec into separate package (Steve Streeting[32m[m)
[33m|[m[33m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [36m|[m [1;33m|[m [1;33m|[m   2cf082f 2016-03-03 Merge pull request #715 from github/netrc (risk danger olson[32m[m)
[36m|[m[1;31m\[m [36m\[m [1;33m\[m [1;33m\[m  
[36m|[m * [36m\[m [1;33m\[m [1;33m\[m   194cc26 2016-03-03 merge master (risk danger olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [36m\[m [1;33m\[m [1;33m\[m  
[36m|[m [1;32m|[m[36m/[m [36m/[m [1;33m/[m [1;33m/[m  
[36m|[m[36m/[m[1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   
* [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   7d8a312 2016-03-03 Merge pull request #1054 from github/docker-test-issues (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m a6cd196 2016-03-01 ignore certain docker network test errors (risk danger olson[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   5d5202e 2016-03-01 Merge pull request #1051 from sinbad/remote-branch-check (Steve Streeting[32m[m)
[1;36m|[m[31m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m bb59098 2016-03-01 Reduce nesting (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 721b41d 2016-03-01 Split RemoteRefs into 2 functions, one for cached, one for real remote call (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 1712f88 2016-02-29 Doc fix (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 919b89d 2016-02-29 Add test for pushing when server branch deleted & objects GC'd (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 9aa57fe 2016-02-29 Fix misleading comment (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m d2107a0 2016-02-29 Fix bad index (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m bb12398 2016-02-29 Check actual remote refs when scanning excluding remote refs (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m ca8aad2 2016-02-29 Change to RemoteRefs and return more standard refs instead (Steve Streeting[32m[m)
[1;36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m b23462a 2016-02-29 Add RemoteBranchList helper func (Steve Streeting[32m[m)
[1;36m|[m[1;36m/[m [1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   17355b4 2016-02-25 Merge pull request #1048 from github/add-core-team (risk danger olson[32m[m)
[32m|[m[33m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 25eed4e 2016-02-25 describe label for core team issues (risk danger olson[32m[m)
[32m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 2461db1 2016-02-25 mention the core team (risk danger olson[32m[m)
* [33m|[m [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   2818626 2016-02-25 Merge pull request #1045 from andyneff/debian_man (risk danger olson[32m[m)
[33m|[m[35m\[m [33m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[33m|[m [35m|[m[33m/[m [1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
[33m|[m[33m/[m[35m|[m [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   
[33m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m c860854 2016-02-25 Fix for #719 and #995 (Andy Neff[32m[m)
[33m|[m[33m/[m [1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
* [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 6b00ebc 2016-02-23 update test so it passes regardless of git version or batch api (risk danger olson[32m (tag: v1.1.1-pre-push-tracing)[m)
* [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   466ef52 2016-02-23 Merge pull request #862 from github/extract-localstorage-1 (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 9024843 2016-02-23 remove concurrent map access issue (risk danger olson[32m[m)
[36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m b8a3cd6 2016-02-23 remove some more `ls.objects` references (risk danger olson[32m[m)
[36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 89d9382 2016-02-23 don't expose the localstorage object from the lfs package (risk danger olson[32m[m)
[36m|[m * [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   b35e385 2016-02-23 merge master (risk danger olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[36m|[m [1;32m|[m[36m/[m [1;32m/[m [36m/[m [1;33m/[m [1;33m/[m  
[36m|[m[36m/[m[1;32m|[m [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   
* [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m 51df853 2016-02-23 update backport-pr docs since i use it so infrequently (risk danger olson[32m[m)
* [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m [1;33m|[m   bb68c35 2016-02-23 Merge pull request #1023 from github/better-errors (risk danger olson[32m[m)
[1;33m|[m[1;35m\[m [1;32m\[m [1;32m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;33m|[m [1;35m|[m[1;33m_[m[1;32m|[m[1;33m_[m[1;32m|[m[1;33m_[m[36m|[m[1;33m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m   
[1;33m|[m * [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m 3daf3d4 2016-02-23 add tests for GetInnerError() (risk danger olson[32m[m)
[1;33m|[m * [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m   bc60e40 2016-02-23 Merge branch 'master' into better-errors (risk danger olson[32m[m)
[1;33m|[m [1;36m|[m[1;33m\[m [1;32m\[m [1;32m\[m [36m\[m [1;33m\[m  
[1;33m|[m [1;36m|[m[1;33m/[m [1;32m/[m [1;32m/[m [36m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[1;36m|[m [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m   
* [1;36m|[m [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m   01d56f1 2016-02-23 Merge pull request #1025 from github/go-1.6 (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;32m\[m [1;32m\[m [36m\[m [1;33m\[m  
[32m|[m * [1;36m|[m [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m bfb5759 2016-02-23 test git lfs on go 1.5 and go 1.6 (risk danger olson[32m[m)
[32m|[m * [1;36m|[m [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m 5c48919 2016-02-23 update scripts to build with vendor experiment off (risk danger olson[32m[m)
[32m|[m[32m/[m [1;36m/[m [1;32m/[m [1;32m/[m [36m/[m [1;33m/[m  
[32m|[m * [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m 256ba65 2016-02-23 tweak the error from ensureFile() (risk danger olson[32m[m)
[32m|[m * [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m 090944e 2016-02-23 actually print non-fatal errors (risk danger olson[32m[m)
[32m|[m[32m/[m [1;32m/[m [1;32m/[m [36m/[m [1;33m/[m  
* [1;32m|[m [1;32m|[m [36m|[m [1;33m|[m   e17f239 2016-02-22 Merge pull request #988 from sinbad/git-lfs-clone (risk danger olson[32m[m)
[36m|[m[35m\[m [1;32m\[m [1;32m\[m [36m\[m [1;33m\[m  
[36m|[m [35m|[m[36m_[m[1;32m|[m[36m_[m[1;32m|[m[36m/[m [1;33m/[m  
[36m|[m[36m/[m[35m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m d0e9536 2016-02-15 Instead of failing when git version <2.2, use slower cat fallback (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 45c4ec3 2016-02-12 Clean up comments (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m a61be12 2016-02-12 Use pseudo-tty to ensure we get full git clone output (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m df70bf5 2016-02-12 Skip clone test when git version < 2.2 (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 48c5cd1 2016-02-12 Only send filtered stderr output to trace to avoid duplication (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 87cdc58 2016-02-12 Send all git clone stderr output to trace so it's visible if needed (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 65b0652 2016-02-12 Document issues with output in git lfs clone (Steve Streeting[32m[m)
[36m|[m * [1;32m|[m [1;32m|[m [1;33m|[m   a19b63b 2016-02-11 Merge master to resolve conflicts in Travis config (Steve Streeting[32m[m)
[36m|[m [36m|[m[1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m e7d27dd 2016-02-11 Use Exit instead of Panic in lfs clone (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 657949e 2016-02-11 git lfs clone requires git version 2.2.0 (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 0a6248b 2016-02-11 Attempting to fix Travis with newer git, cred helper issue (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 74b3f79 2016-02-11 Get a more recent version of git (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m b37fe0f 2016-02-11 Added tests for 'git lfs clone' (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m a1fd022 2016-02-11 Remove copy/paste error (Steve Streeting[32m[m)
[36m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 894794b 2016-02-11 Implemented 'git lfs clone' (Steve Streeting[32m[m)
* [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m e49c3ae 2016-02-22 send the PATH too (risk danger olson[32m[m)
* [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   d69ccc1 2016-02-22 Merge pull request #1006 from github/fix-scanner-panic (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m * [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m   66d3014 2016-02-22 merge (risk danger olson[32m (chris/fix-scanner-panic)[m)
[1;32m|[m [1;34m|[m[1;32m\[m [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [1;31m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   68dffe0 2016-02-22 Merge pull request #1016 from github/lstree-parsing (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m d9e84ff 2016-02-21 add 'lfs pull' test with unicode filename (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m ad022c9 2016-02-19 no need to export this right now (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m fa1cd32 2016-02-19 split on whitespace manually instead of using a regex (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 7acc064 2016-02-19 write a quick benchmark (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 8244c51 2016-02-19 actually add -z to ls-tree command (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 7f66b77 2016-02-19 split on null (risk danger olson[32m[m)
[1;36m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 9088cc3 2016-02-19 extract ls tree parser (risk danger olson[32m[m)
* [31m|[m [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 526de52 2016-02-22 typo (risk danger olson[32m[m)
* [31m|[m [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   a8811b2 2016-02-22 Merge pull request #1017 from github/kill-more-zombies (risk danger olson[32m[m)
[31m|[m[33m\[m [31m\[m [1;34m\[m [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [33m|[m[31m/[m [1;34m/[m [1;31m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[33m|[m [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[31m|[m * [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 2bc5ec8 2016-02-19 make sure we're calling Wait() like we're supposed to (risk danger olson[32m[m)
[31m|[m[31m/[m [1;34m/[m [1;31m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;34m|[m [1;31m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   a2781ff 2016-02-18 Merge pull request #1012 from github/rlaakkol-master (risk danger olson[32m[m)
[1;31m|[m[35m\[m [1;34m\[m [1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;31m|[m [35m|[m[1;31m_[m[1;34m|[m[1;31m_[m[1;31m|[m[1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[35m|[m [1;34m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;31m|[m * [1;34m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 7f63e58 2016-02-17 wait for command to finish before closing the channel (risk danger olson[32m[m)
[1;31m|[m * [1;34m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m fc9658e 2016-02-16 Moved exec.Wait()s to better places (Riku Lääkkölä[32m[m)
[1;31m|[m * [1;34m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 5cf6935 2016-02-12 Added zombie cleanups for external git command calls. (Riku Lääkkölä[32m[m)
[1;31m|[m[1;31m/[m [1;34m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m af6c982 2016-02-15 skip lines that are too short (risk danger olson[32m[m)
[1;31m|[m[1;31m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   05aa87d 2016-02-11 Merge pull request #990 from larsxschneider/travis (risk danger olson[32m[m)
[1;31m|[m[1;31m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;31m|[m [1;31m|[m[1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m dc358a0 2016-02-11 travis-ci: declare global variable in 'env' section (Lars Schneider[32m[m)
[1;31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m ed2b9f6 2016-02-11 travis-ci: use multi line 'before_install' phase to improve readability (Lars Schneider[32m[m)
[1;31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m bc9ed62 2016-02-11 travis-ci: run tests against default Git and latest Git (Lars Schneider[32m[m)
* [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m e801c47 2016-02-10 push centos/7 packages to fedora too (risk danger olson[32m[m)
* [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   6120422 2016-02-10 Merge pull request #989 from github/base64-stdencoding (risk danger olson[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;31m|[m [1;33m|[m[1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 60346d2 2016-02-09 The standard base64 encoding is more common (risk danger olson[32m[m)
[1;31m|[m[1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;32m|[m [1;32m|[m [1;33m|[m   8fa4963 2016-02-05 Merge pull request #980 from larsxschneider/travis-latest-git (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;34m|[m * [1;32m|[m [1;32m|[m [1;33m|[m a477140 2016-02-05 fix 'test-credentials-no-prompt' to work with the LFS batch API (Lars Schneider[32m[m)
[1;34m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 7338338 2016-02-05 install latest Git on Travis-CI Linux to run all tests (Lars Schneider[32m[m)
* [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m   69d1419 2016-02-05 Merge pull request #981 from github/build-all-temp-env (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m fa75a07 2016-02-02 dont shadow the os package name (risk danger olson[32m[m)
[1;36m|[m * [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 131132a 2016-02-02 copy TEMP vars over (risk danger olson[32m[m)
* [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 1337e1d 2016-02-05 update package cloud uploader (risk danger olson[32m[m)
* [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 6684825 2016-02-05 fix test on windows (risk danger olson[32m[m)
* [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 7de0397 2016-02-04 release v1.1.1 (risk danger olson[32m (tag: v1.1.1)[m)
[1;35m|[m [31m|[m[1;35m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   94d356c 2016-02-04 Merge branch 'bozaro-btrfs-clone' (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 600a577 2016-02-04 no need to test unused/removed method anymore (risk danger olson[32m[m)
[32m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m e5d16d3 2016-02-04 goimports (risk danger olson[32m[m)
[32m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   d884dad 2016-02-04 Merge branch 'btrfs-clone' of https://github.com/bozaro/git-lfs into bozaro-btrfs-clone (risk danger olson[32m[m)
[32m|[m [34m|[m[35m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
* [34m|[m [35m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m   289c273 2016-02-04 Merge pull request #952 from bozaro/btrfs-clone (risk danger olson[32m[m)
[34m|[m[35m\[m [34m\[m [35m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[34m|[m [35m|[m[34m/[m [35m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[34m|[m[34m/[m[35m|[m [35m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m   
[34m|[m [35m|[m[35m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m    
[34m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m a2f72c9 2016-01-30 Fix non linux with cgo build condition (Artem V. Navrotskiy[32m[m)
[34m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m ea16fd5 2016-01-26 Add copy-on-write support for Linux BTRFS filesystem (Artem V. Navrotskiy[32m[m)
* [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   4cca3f8 2016-02-04 Merge pull request #977 from github/tunable-timeouts (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m * [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m fcd732f 2016-02-04 document new git config settings (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 886dc23 2016-02-04 set max idle conns to the concurrent transfers value (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 938bf76 2016-02-04 add tunable http client timeouts. also increase defaults. (risk danger olson[32m[m)
[1;32m|[m[1;32m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   870c81e 2016-02-04 Merge pull request #976 from andyneff/centos5_cgo (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;34m|[m * [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 6006acc 2016-02-03 Disable CGO for Centos 5 to assist #952 (Andy Neff[32m[m)
* [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   5a37cf5 2016-02-04 Merge pull request #975 from github/fix-track-patterns (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 51e2819 2016-02-03 pass gitattributes patterns (mostly) unchanged (risk danger olson[32m[m)
[1;36m|[m [1;35m|[m[1;35m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   4857922 2016-02-04 Merge pull request #974 from github/push-with-invalid-remote (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m * [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 3f89436 2016-02-03 validate remote name in pre-push too (risk danger olson[32m[m)
[32m|[m * [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m dceb661 2016-02-03 guard against invalid remotes passed to `git lfs push` (risk danger olson[32m[m)
[32m|[m [1;35m|[m[1;35m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   5677342 2016-02-04 Merge pull request #971 from github/skip-trace (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[34m|[m * [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 6f682a8 2016-02-03 remove dumb debug messages (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m c4874df 2016-02-03 don't pass GIT_TRACE to exec.Command calls in the git package (risk danger olson[32m[m)
[34m|[m [1;35m|[m[1;35m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;35m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   d9f99fa 2016-02-04 Merge pull request #972 from github/fix-pull-bug (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [1;35m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;35m|[m [1;31m|[m[1;35m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;35m|[m * [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 90112cf 2016-02-03 setup the transfer queue watcher BEFORE processing any LFS pointers (risk danger olson[32m[m)
[1;35m|[m[1;35m/[m [1;31m/[m [31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;31m|[m [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   0640430 2016-02-02 Merge pull request #875 from strich/win-installer-upgrade (risk danger olson[32m[m)
[31m|[m[1;33m\[m [1;31m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [1;33m|[m[31m_[m[1;31m|[m[31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 1651e55 2016-01-28 Git LFS now installs to Prog Files/Git LFS/. Note that for this to work, the old Git LFS must be uninstalled - This installer will attempt to do this for you silently. Added support for compiling x86 and x64 binaries into a single 32bit installer. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m b7bf2ce 2015-12-01 Moved the prefix to the define list, for future expansion. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 016a11d 2015-12-01 The output installer filename now dynamically sets the version and respects existing naming conventions. Set the existing version number to 3 digits instead of 4. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m df0ebba 2015-11-30 Added a batch script to compile the win-installer. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m c9c79ec 2015-11-30 Renamed a few things to be more descriptive. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 11ec5c8 2015-11-30 Removed old nsis-based installer scripts. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m ae5caa3 2015-11-29 Added code that runs 'git lfs uninstall' at uninstall time. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m c9bce8e 2015-11-28 Installer now compiles to the base directory along-side git-lfs.exe. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 94f3064 2015-11-28 Fixed a bug where changing the install location didn't actually do so. Added post-install git lfs initialization. (Scott Richmond[32m[m)
[31m|[m * [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 74a249f 2015-11-28 Committing initial working version of the INNO Setup script file. Replicates existing install behavior. (Scott Richmond[32m[m)
* [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 295fd47 2016-02-02 build the test server api command with upload and download tests (risk danger olson[32m[m)
* [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   21bdc9a 2016-02-02 merge master (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m cd8bba7 2015-11-26 Build API test binary in integration tests to ensure stays up to date (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 41e7996 2015-11-26 Upload tests (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m f6eb937 2015-11-26 Summarise test results in console & exit code (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 39102e9 2015-11-26 Generalise interleaving of test data (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m b655a15 2015-11-26 Fixed test mistake when exist/missing lists are different lengths (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 2b62524 2015-11-26 Fix error message when error missing (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m a613919 2015-11-26 Mixed download test (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m d995ad5 2015-11-25 Initial download tests (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 773465e 2015-11-25 Expose ObjectError struct since part of exposed API (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 110e02e 2015-11-25 Allow generated data to be saved for future use (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m f820483 2015-11-25 Make it easy to add tests (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 3a0ee06 2015-11-25 Report construction of test data (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m eb777fa 2015-11-25 Make sure we force the loading of gitconfig before altering anything (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 5bfed86 2015-11-25 Set up test oids using actual data so content will validate, & upload (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 44c8bc7 2015-11-25 Need to include size in test data (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m ab21ec8 2015-11-25 Allow Endpoint to be manually specified instead of from gitconfig (Steve Streeting[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m eca6d99 2015-11-25 Initial docs & scaffold for server API tests (Steve Streeting[32m[m)
* [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   d56c5fd 2016-02-02 Merge pull request #964 from github/git-remotes (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 703a4c3 2016-02-02 add lfs.gitprotocol to safe config keys (risk danger olson[32m[m)
[1;36m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 2282a34 2016-02-02 add lfs.gitprotocol to configure default protocol for git:// lfs server (risk danger olson[32m[m)
[1;36m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m cc8524f 2016-01-29 convert git:// remotes to the lfs default (risk danger olson[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   96d86cd 2016-02-02 Merge branch 'nathanhi-0byte-files' (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 2dd108f 2016-02-02 add an integration test for zero len files (risk danger olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m f25452c 2016-02-02 simpler early return (risk danger olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   6847e45 2016-02-02 Merge branch '0byte-files' of https://github.com/nathanhi/git-lfs into nathanhi-0byte-files (risk danger olson[32m[m)
[32m|[m [32m|[m[35m\[m [31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m[32m/[m [35m/[m [31m/[m [1;35m/[m [1;33m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 763ed1f 2016-02-02 Return empty buffer early on empty string (Nathan-J. Hirschauer[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m d1b9376 2016-02-01 Test for empty file pointer generation (Nathan-J. Hirschauer[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m a59f59a 2016-02-01 Don't create pointer for empty files (Nathan-J. Hirschauer[32m[m)
[32m|[m [31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   45b0574 2016-02-02 Merge pull request #963 from github/fix-submod-lfsconfig (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[36m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m a6df20a 2016-02-01 always use 'git rev-parse' to figure out git and working tree dirs (risk danger olson[32m[m)
[36m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 7128d90 2016-01-29 treat GIT_WORK_TREE as a relative of the current working directory (risk danger olson[32m[m)
[36m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 5e816ea 2016-01-29 test lfsconfig behavior with `git submodule update --init --remote` (risk danger olson[32m[m)
[36m|[m [31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   20560c5 2016-02-01 Merge pull request #965 from github/tweak-errors (risk danger olson[32m[m)
[31m|[m[1;33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [1;33m|[m[31m/[m [1;35m/[m [1;33m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[1;33m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[31m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 07a9e6e 2016-01-29 caught this in the syntax highlighted diff (risk danger olson[32m[m)
[31m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 479713d 2016-01-29 better errors than "relation not found" (risk danger olson[32m[m)
[31m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 38b520a 2016-01-29 don't spit out uri query params, just too noisy (risk danger olson[32m[m)
[31m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m b01eaf7 2016-01-29 "media" word is left over from "git media" days (risk danger olson[32m[m)
[31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   b5e636e 2016-01-29 Merge pull request #949 from sinbad/support-remote-pushurl (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;34m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 8a5d271 2016-01-25 Fix doc comment (Steve Streeting[32m[m)
[1;34m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 21dbd49 2016-01-25 Support remote.name.pushurl correctly to fix #945 (Steve Streeting[32m[m)
[1;34m|[m [1;31m|[m [1;35m|[m[1;31m_[m[1;33m|[m[1;31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;34m|[m [1;31m|[m[1;31m/[m[1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [1;31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   70d6ef4 2016-01-29 Merge pull request #860 from github/remove-git-current-branch (risk danger olson[32m[m)
[1;31m|[m[31m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;31m|[m [31m|[m[1;31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;31m|[m[1;31m/[m[31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;31m|[m * [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m d717011 2015-11-24 remove git.CurrentBranch() and its last internal uses (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m cc7da10 2015-11-24 remove unnecessary git.CurrentBranch() call in status command (risk danger olson[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   c57bcf0 2016-01-19 Merge pull request #891 from andyneff/rpm_require (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m acfc0dc 2015-12-06 Got old self signing tests working again (Andy Neff[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m d2d90a8 2015-12-06 Trying the IUS repo for git for centos 6 #878 (Andy Neff[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m 6d61f5b 2015-12-06 Updates to support the new automated build docker images (Andy Neff[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m 047bc59 2015-12-05 Clean up and update to gpg_agent via docker_hub (Andy Neff[32m[m)
* [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   5cbe389 2016-01-14 Merge pull request #933 from larsxschneider/osx-travis (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[34m|[m * [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m a43b3c2 2016-01-14 enable OSX build on TravisCI (Lars Schneider[32m[m)
[34m|[m[34m/[m [33m/[m [31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   cc90378 2016-01-05 Merge pull request #900 from ro31337/patch-1 (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[36m|[m * [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m 6a60c1a 2015-12-15 Update README.md (Roman Pushkin[32m[m)
[36m|[m [33m|[m[33m/[m [31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   2c77a36 2015-12-24 Merge pull request #909 from sinbad/fix-ssh-error-fallthrough (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m * [33m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m a79cb1c 2015-12-22 Fix fallthrough when git-lfs-authenticate returns an error (Steve Streeting[32m[m)
[1;32m|[m [31m|[m [33m|[m[31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;32m|[m [31m|[m[31m/[m[33m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [31m|[m [33m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   b5892fd 2015-12-24 Merge pull request #907 from wpsmith/packagecloud-link (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [33m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;34m|[m * [31m|[m [33m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m c79afdc 2015-12-21 Update PackageCloud Link (Travis Smith[32m[m)
[1;34m|[m[1;34m/[m [31m/[m [33m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [31m|[m [33m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   5466fc0 2015-12-20 Merge pull request #883 from github/progress-bar-col-fix (risk danger olson[32m[m)
[33m|[m[31m\[m [31m\[m [33m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[33m|[m [31m|[m[33m_[m[31m|[m[33m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[33m|[m[33m/[m[31m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[33m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m bff3d07 2015-12-02 --amend (risk danger olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m abd88a1 2015-12-02 handle cases where ts package can't calculate the terminal width (risk danger olson[32m[m)
* [31m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   e467a11 2015-12-04 Merge pull request #861 from github/lfs-config-warnings (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[32m|[m * [31m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m dcae83d 2015-11-24 only show git config warnings for lfs keys (risk danger olson[32m[m)
[32m|[m [31m|[m [31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[32m|[m [31m|[m[31m/[m[31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [31m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   9617da0 2015-12-04 Merge pull request #882 from github/readme-updates (risk danger olson[32m[m)
[31m|[m[35m\[m [31m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [35m|[m[31m_[m[31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[35m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[31m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   36551c4 2015-12-02 merge master (risk danger olson[32m[m)
[31m|[m [36m|[m[31m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [36m|[m[31m/[m [31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[36m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
* [36m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   79e0090 2015-12-02 Merge pull request #879 from ssgelm/add-git-version-dep-to-deb (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [36m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m * [36m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m c54f301 2015-12-02 Add git version dependency to debian package (Stephen Gelman[32m[m)
[1;32m|[m[1;32m/[m [36m/[m [31m/[m [1;35m/[m [1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [36m|[m [31m|[m [1;35m|[m [1;33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   2b28501 2015-12-01 Merge pull request #877 from pabloguerrero/patch-1 (risk danger olson[32m[m)
[1;33m|[m[1;35m\[m [36m\[m [31m\[m [1;35m\[m [1;33m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;33m|[m [1;35m|[m[1;33m_[m[36m|[m[1;33m_[m[31m|[m[1;33m_[m[1;35m|[m[1;33m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [36m|[m [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;33m|[m * [36m|[m [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 0923d80 2015-12-02 Update README.md (Pablo Guerrero[32m[m)
[1;33m|[m * [36m|[m [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 5de8e15 2015-12-02 Update README.md (Pablo Guerrero[32m[m)
[1;33m|[m[1;33m/[m [36m/[m [31m/[m [1;35m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;33m|[m * [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m 9e8cd56 2015-12-02 big updates to the readme and contributing guide (risk danger olson[32m[m)
[1;33m|[m[1;33m/[m [31m/[m [1;35m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [31m|[m [1;35m|[m [1;32m|[m [1;32m|[m [1;33m|[m   24166bd 2015-11-25 Merge pull request #863 from sinbad/export-object-resource (risk danger olson[32m[m)
[1;36m|[m[1;35m\[m [31m\[m [1;35m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[1;36m|[m [1;35m|[m [31m|[m[1;35m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[1;36m|[m [1;35m|[m[1;35m/[m[31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[1;36m|[m * [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m 3c86ea1 2015-11-25 Export ObjectResource; it's returned by several exported functions (Steve Streeting[32m[m)
[1;36m|[m [31m|[m[31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
* [31m|[m [1;32m|[m [1;32m|[m [1;33m|[m   c95ba38 2015-11-25 Merge pull request #851 from noamt/ruby-upgrade (risk danger olson[32m[m)
[31m|[m[33m\[m [31m\[m [1;32m\[m [1;32m\[m [1;33m\[m  
[31m|[m [33m|[m[31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m[31m/[m[33m|[m [1;32m|[m [1;32m|[m [1;33m|[m   
[31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 83f18f6 2015-11-24 Depending on ruby and ruby-dev packages instead of version specific packages. There's no dependency that would limit us to a specific version and depending on the generic will be less fragile in the long run (noamt[32m[m)
[31m|[m * [1;32m|[m [1;32m|[m [1;33m|[m 1cc4868 2015-11-20 Upgrading to the Debian build script to Ruby 2.2 because Ubuntu repositories no longer offer the Ruby2.0 package, thus breaking the script (noamt[32m[m)
[31m|[m[31m/[m [1;32m/[m [1;32m/[m [1;33m/[m  
[31m|[m * [1;32m|[m [1;33m|[m 766da7e 2016-01-29 fix a few more references (risk danger olson[32m[m)
[31m|[m * [1;32m|[m [1;33m|[m 54b8eeb 2016-01-29 rename LocalStorage var to Objects (risk danger olson[32m[m)
[31m|[m * [1;32m|[m [1;33m|[m a94c78f 2015-11-24 extra temp object cleanup (risk danger olson[32m[m)
[31m|[m * [1;32m|[m [1;33m|[m 50a3980 2015-11-24 teach localstorage how to init its dirs (risk danger olson[32m[m)
[31m|[m * [1;32m|[m [1;33m|[m 8517888 2015-11-24 extract localstorage path building (risk danger olson[32m[m)
[31m|[m * [1;32m|[m [1;33m|[m 6a03511 2015-11-24 extract localstorage object scanning to separate package (risk danger olson[32m[m)
[31m|[m[31m/[m [1;32m/[m [1;33m/[m  
* [1;32m|[m [1;33m|[m   fbbfd7d 2015-11-18 Merge pull request #843 from github/package-uploads (risk danger olson[32m[m)
[34m|[m[35m\[m [1;32m\[m [1;33m\[m  
[34m|[m * [1;32m|[m [1;33m|[m aa1bba7 2015-11-18 upload ubuntu packages and generate some markdown package download links (risk danger olson[32m[m)
[34m|[m[34m/[m [1;32m/[m [1;33m/[m  
* [1;32m|[m [1;33m|[m 258acf1 2015-11-18 release v1.1.0 (risk danger olson[32m (tag: v1.1.0)[m)
* [1;32m|[m [1;33m|[m 469c23e 2015-11-18 fix some more 'lfs init' references (risk danger olson[32m[m)
* [1;32m|[m [1;33m|[m   7b94764 2015-11-18 Merge pull request #750 from glensc/patch-2 (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;32m\[m [1;33m\[m  
[36m|[m * [1;32m|[m [1;33m|[m 68308d5 2015-11-18 Update git-lfs.spec (Elan Ruusamäe[32m[m)
[36m|[m * [1;32m|[m [1;33m|[m 37c68fb 2015-10-21 restore perl-Digest-SHA (Elan Ruusamäe[32m[m)
[36m|[m * [1;32m|[m [1;33m|[m 6edeae7 2015-10-17 Update git-lfs.spec (Elan Ruusamäe[32m[m)
* [1;31m|[m [1;32m|[m [1;33m|[m   17a9e3f 2015-11-18 Merge pull request #842 from github/cred-fixes (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;32m\[m [1;33m\[m  
[1;32m|[m [1;33m|[m [1;31m|[m * [1;33m|[m 2a601cb 2015-11-21 Use a build flagged var instead of var + init() for setting netrc name (rubyist[32m[m)
[1;32m|[m [1;33m|[m [1;31m|[m * [1;33m|[m   a262559 2015-11-18 Merge branch 'cred-fixes' into netrc (risk danger olson[32m[m)
[1;32m|[m [1;33m|[m [1;31m|[m [1;34m|[m[1;33m\[m [1;33m\[m  
[1;32m|[m [1;33m|[m [1;31m|[m[1;33m_[m[1;34m|[m[1;33m/[m [1;33m/[m  
[1;32m|[m [1;33m|[m[1;33m/[m[1;31m|[m [1;34m|[m [1;33m|[m   
[1;32m|[m * [1;31m|[m [1;34m|[m [1;33m|[m 8bb4d4f 2015-11-18 send json response bodies to tracerx (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m * [1;33m|[m f145021 2015-11-18 gofmt (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m * [1;33m|[m e7b4348 2015-11-18 cache the parsed netrc file so it's not checked all the time (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m * [1;33m|[m 69ff7b8 2015-11-18 save merge conflict fix, doh (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m * [1;33m|[m   e5ff9c9 2015-11-18 merge cred-fixes (risk danger olson[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m [1;36m|[m[1;35m\[m [1;33m\[m  
[1;32m|[m [1;35m|[m [1;31m|[m[1;35m_[m[1;36m|[m[1;35m/[m [1;33m/[m  
[1;32m|[m [1;35m|[m[1;35m/[m[1;31m|[m [1;36m|[m [1;33m|[m   
[1;32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 88d1f7a 2015-11-18 hide the problematic test behind a git 2.3 check (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m adfea0d 2015-11-18 add full http debug mode (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m f596625 2015-11-18 move credential error handling to fillCredentials() (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m   695d0fc 2015-11-17 Merge branch 'advertise' into cred-fixes (risk danger olson[32m[m)
[1;32m|[m [32m|[m[33m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[1;32m|[m [32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 39d9304 2015-11-17 cleaner http dump output (risk danger olson[32m (chris/advertise)[m)
[1;32m|[m [32m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 97b92b5 2015-11-17 add "git lfs" to the progress meter (risk danger olson[32m[m)
[1;32m|[m [32m|[m[1;32m/[m [1;31m/[m [1;36m/[m [1;33m/[m  
[1;32m|[m[1;32m/[m[32m|[m [1;31m|[m [1;36m|[m [1;33m|[m   
* [32m|[m [1;31m|[m [1;36m|[m [1;33m|[m   b8ac3a7 2015-11-17 Merge pull request #749 from glensc/patch-1 (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[34m|[m * [32m|[m [1;31m|[m [1;36m|[m [1;33m|[m 1a51bf0 2015-10-21 update url version (Elan Ruusamäe[32m[m)
[34m|[m * [32m|[m [1;31m|[m [1;36m|[m [1;33m|[m a83dd75 2015-10-17 Update git-lfs.spec (Elan Ruusamäe[32m[m)
[34m|[m [1;31m|[m [32m|[m[1;31m/[m [1;36m/[m [1;33m/[m  
[34m|[m [1;31m|[m[1;31m/[m[32m|[m [1;36m|[m [1;33m|[m   
[34m|[m [1;31m|[m * [1;36m|[m [1;33m|[m 47725d0 2015-11-17 don't return any empty creds (risk danger olson[32m[m)
[34m|[m [1;31m|[m * [1;36m|[m [1;33m|[m da37a83 2015-11-17 halt if creds are needed but not found (risk danger olson[32m[m)
[34m|[m [1;31m|[m * [1;36m|[m [1;33m|[m 09aff9e 2015-11-17 add tracer messages when creds are filled or not found (risk danger olson[32m[m)
[34m|[m [1;31m|[m[34m/[m [1;36m/[m [1;33m/[m  
[34m|[m[34m/[m[1;31m|[m [1;36m|[m [1;33m|[m   
* [1;31m|[m [1;36m|[m [1;33m|[m   02b89e3 2015-11-17 Merge pull request #838 from github/init-to-install (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[36m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 1dba9c4 2015-11-16 annotate TODOs with issues (risk danger olson[32m[m)
[36m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 8089b9e 2015-11-16 add 2.0 deprecation todos (risk danger olson[32m[m)
[36m|[m * [1;31m|[m [1;36m|[m [1;33m|[m 74bf7f3 2015-11-16 rename uninit => uninstall (risk danger olson[32m[m)
[36m|[m * [1;31m|[m [1;36m|[m [1;33m|[m c9fe096 2015-11-16 rename init => install (risk danger olson[32m[m)
* [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m   14aeecc 2015-11-17 Merge pull request #837 from github/use-lfsconfig (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m 826f8e3 2015-11-16 annotate todo with issue ref (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m e13f5d3 2015-11-16 only show the .lfsconfig warning IF .gitconfig exists (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m c7fe776 2015-11-16 nicer terror message (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m dcb2d4e 2015-11-16 warn if two similar git config values are seen (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m 265feee 2015-11-16 only show the warning for 'git lfs env' (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m 3ce5770 2015-11-16 add a warning message for .gitconfig users (risk danger olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m e836c1f 2015-11-16 use .lfsconfig (with fallback to .gitconfig) (risk danger olson[32m[m)
[1;32m|[m [1;31m|[m[1;31m/[m [1;31m/[m [1;36m/[m [1;33m/[m  
* [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m   16103be 2015-11-17 Merge pull request #784 from andyneff/appveyor_poc (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[1;34m|[m * [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m f1c0ce4 2015-10-23 Appveyor Proof of concept (Andy Neff[32m[m)
* [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m   b1baae0 2015-11-17 Merge pull request #807 from andyneff/test_dir_fix (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m c9db386 2015-10-31 Fixed GIT_LFS_TEST_DIR to work on the outside tests (Andy Neff[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m 028f340 2015-10-30 Undo change to bootstrap (Andy Neff[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m aa8b753 2015-10-30 Broke, can't build at home (Andy Neff[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m ac9df2d 2015-10-29 Removed logging redirect for building rpms (Andy Neff[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;33m|[m 26a9879 2015-10-29 Added GIT_LFS_TEST_DIR to rpm spec (Andy Neff[32m[m)
[1;36m|[m [1;33m|[m [1;35m|[m[1;33m_[m[1;31m|[m[1;33m_[m[1;31m|[m[1;33m_[m[1;36m|[m[1;33m/[m  
[1;36m|[m [1;33m|[m[1;33m/[m[1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m   
* [1;33m|[m [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m   e2060f1 2015-11-16 Merge pull request #831 from ajohnson23/master (risk danger olson[32m[m)
[32m|[m[33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [1;33m|[m [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m 71bfd4e 2015-11-11 smudge --skip will now smudge if object is in cache. (Andrew Johnson[32m[m)
* [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m 090222f 2015-11-16 unnecessary trace msg (risk danger olson[32m[m)
* [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m   256e95b 2015-11-16 Merge pull request #818 from sinbad/fix-track-in-symlinked-dir (risk danger olson[32m[m)
[1;31m|[m[35m\[m [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;31m\[m [1;36m\[m  
[1;31m|[m [35m|[m[1;31m_[m[33m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;35m|[m[1;31m/[m [1;31m/[m [1;36m/[m  
[1;31m|[m[1;31m/[m[35m|[m [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;31m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 8af895c 2015-11-06 Fix "outside repository" error when "git lfs track" in a symlinked dir (Steve Streeting[32m[m)
[1;31m|[m [33m|[m[33m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   a4dca09 2015-11-16 Merge pull request #742 from sinbad/prune (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m   3fb0150 2015-11-11 Merge branch 'sinbad-prune' into prune (Steve Streeting[32m[m)
[36m|[m [1;32m|[m[1;33m\[m [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m [1;32m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m bf297da 2015-11-09 use the lfs.NewScanRefsOptions constructor (risk danger olson[32m (chris/sinbad-prune)[m)
[36m|[m [1;32m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   9c33079 2015-11-09 try a merge out (risk danger olson[32m[m)
[36m|[m [1;32m|[m [33m|[m[1;32m\[m [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m [1;32m|[m [33m|[m[1;32m/[m [33m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[36m|[m [1;32m|[m[1;32m/[m[33m|[m [33m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m   
[36m|[m [1;32m|[m [33m|[m[33m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m    
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7a615bb 2015-10-16 Be explicit that the reflog isn't grounds for retention, only commits (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 262d37c 2015-10-16 Make progress type field name less ambiguous (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 97434fe 2015-10-15 Implement `fetch --prune` + test (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m d7b23e2 2015-10-15 Move prune worktree test into its own file to improve git version skip (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 459f1a6 2015-10-15 Tests for prune with worktrees (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m ab48608 2015-10-15 Remove unnecessary test output (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m a020f9d 2015-10-15 Test verify remote via config, and overriding it on cmdline (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m aa9a240 2015-10-15 Tests for the --verify-remote option (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m b8cdfaf 2015-10-15 Remove superfluous newlines in trace output (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5863dff 2015-10-15 Tidy up tests for remotes (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m f83f761 2015-10-15 Add delete_server_object from another PR, I need it here too (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 00637f1 2015-10-15 Improve trace info around verifying pruned objects with remote (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 817d865 2015-10-15 Verify if asked even in dry-run, results should match just not delete (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m fcdd67b 2015-10-15 Clarify that the pruneoffsetdays is only used if fetch days > 0 (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7db5a01 2015-10-15 Prune tests for unpushed with variable remote conditions (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0b537eb 2015-10-14 Test turning off the recent options & make sure more data pruned (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 16cac5e 2015-10-14 Test for prune with recent options (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 473f41e 2015-10-14 Fix typo in test setting, should be pruneoffsetdays not pruneoffset (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 06bb90e 2015-10-14 Add extra trace output for recent refs (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 6801922 2015-10-14 Only retain recent refs if fetch recent setting > 0 (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m d377e7a 2015-10-14 Use fetch prefs to determine whether to include remote branches in recent (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5ead773 2015-10-14 Make retention information available via GIT_TRACE (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 9d5d2a0 2015-10-14 Fix unpushed merge test (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m c8778a7 2015-10-12 Add merge stage to unpushed test (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m af27d5a 2015-10-12 Use "Nothing to prune" message in dry run mode too (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 58dfa23 2015-10-12 Test for retaining unpushed data (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m baa324d 2015-10-12 When lfs.fetchrecentcommitsdays=0 prune should prune previous commits (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 00ca2d7 2015-10-12 Special case message for "Nothing to prune" (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 03b85e0 2015-10-12 Moving to smaller tests for prune so cause for retention is clearer (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5574bac 2015-10-12 Make verbose output visually a sub detail of summary printed at end (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 9d50c0f 2015-10-12 Actually print verbose output (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m eeb61d9 2015-10-12 Fix prune args (were all connected to dry run, doh) (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 489cf08 2015-10-12 Fix verbose typo (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m db57e0f 2015-10-12 Eliminate duplicate retentions from progress log (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 978899e 2015-10-09 WIP prune tests (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 50be5ab 2015-10-09 Implement reachable objects for remote checking (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0a004dc 2015-10-09 Retain files for other worktree checkouts (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 963d0d4 2015-10-08 Add methods for determining the refs that all worktrees are using (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m a9a62fe 2015-10-08 Add utility functions for testing git version (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 444ae6c 2015-10-02 Retain recent refs (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7d62be9 2015-10-02 Implement retention of unpushed objects (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 11cea4d 2015-10-02 Export channel version of ScanUnpushed (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 74f37dd 2015-10-02 Enable ScanUnpushed to report for named remotes instead of all remotes (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 552e8ab 2015-10-02 Implement retention of current checkout (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m d57dbed 2015-10-02 Add error collection goroutine so we can abort on fatal errors (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5c13b75 2015-10-02 Sync progress to ensure we always finish printing it (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m ef0a1af 2015-10-02 Delete files implemented (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 509c6cc 2015-10-02 Implemented main prune routine & waits, progress meters, reporting (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0965b58 2015-10-02 Change AllLocalObjects to return Pointers so we have access to size (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 3288c9d 2015-10-02 Skeleton of algorithm & main goroutines (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4e27696 2015-10-02 Test data is fixed now (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m cdee499 2015-10-02 Add StringSet to replace use of map[string]struct{} directly (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m b3f1baf 2015-10-02 Added function to scan for all locally stored objects (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 3374f7e 2015-10-02 Skeleton for prune command & algorithm in pseudocode (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 9b1a15d 2015-10-02 Add new prune config settings (Steve Streeting[32m[m)
[36m|[m * [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m bd72983 2015-10-02 Documenting the `prune` command and its config settings (Steve Streeting[32m[m)
* [1;35m|[m [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m b5fa32f 2015-11-16 add an env test to confirm (risk danger olson[32m[m)
* [1;35m|[m [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   d29a402 2015-11-16 Merge branch 'jx/env-error-reports-remote-endpoint' of https://github.com/jiangxin/git-lfs into jiangxin-jx/env-error-reports-remote-endpoint (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [33m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [33m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m e864529 2015-11-11 env: Show proper SshUserAndHost for remoteEndpoint (Jiang Xin[32m[m)
[1;36m|[m [33m|[m [1;35m|[m[33m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;36m|[m [33m|[m[33m/[m[1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [33m|[m [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   cf92541 2015-11-16 Merge pull request #820 from github/WillHipschman-ntlm (risk danger olson[32m[m)
[33m|[m[33m\[m [33m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[33m|[m [33m|[m[33m/[m [1;35m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[33m|[m[33m/[m[33m|[m [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   7225d04 2015-11-16 Merge pull request #821 from github/ntlm-cloneable-body (risk danger olson[32m[m)
[33m|[m [34m|[m[35m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[33m|[m [34m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 8106843 2015-11-16 better var names (risk danger olson[32m[m)
[33m|[m [34m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 45b29a2 2015-11-06 only buffer up to 1MB in memory when copying http request bodies (risk danger olson[32m[m)
[33m|[m [34m|[m[34m/[m [1;35m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m d82607b 2015-11-06 clone the transfer encoding property (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 192b106 2015-11-06 extract cloneRequestBody() (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m aa21be5 2015-11-06 never want to TEST for a panic (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 045ee3c 2015-11-06 a const works just fine here for now (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 46448e8 2015-11-06 prefer ioutil.NopCloser() (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4e7ba74 2015-11-06 no need to defer here (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 42bf10b 2015-11-06 protect against potential go panic (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0f54f9a 2015-11-06 remove some unnecessary helper funcs (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m a51a655 2015-11-06 force ConcurrentTransfers() to 1 for ntlm (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 41db1f9 2015-11-06 another spot where the res is double closed (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 1833fa8 2015-11-06 no need to close the body again (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m bacc69f 2015-11-06 fix the md formatting. linebreak waaaaar (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m ac9219c 2015-11-06 ntlm isn't special (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m d5622de 2015-11-06 use a standard tracer msg when toggling the auth type (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 16835c3 2015-11-06 don't repeat !Config.NtlmAccess() as much (risk danger olson[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   6a463b4 2015-11-06 Merge branch 'ntlm' of https://github.com/WillHipschman/git-lfs into WillHipschman-ntlm (risk danger olson[32m[m)
[33m|[m [33m|[m[1;31m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[33m|[m[33m/[m [1;31m/[m [1;35m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   6599c1f 2015-10-30 Merge From Master (William Hipschman[32m[m)
[33m|[m [1;32m|[m[1;33m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[33m|[m [1;32m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m ac4306f 2015-10-30 Update NTLM Proposal (William Hipschman[32m[m)
[33m|[m [1;32m|[m [1;33m|[m [1;35m|[m[1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[33m|[m [1;32m|[m [1;33m|[m[1;33m/[m[1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 279cd1a 2015-10-16 Formatting (Will[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 22c4b5a 2015-10-16 Update Authenticatin Docs To Include NTLM Info (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b8caba5 2015-10-16 Add NTLM Unit Tests (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 41fd037 2015-10-13 Remove Unneeeded Comments (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m fae7766 2015-10-13 Try Removing Stream Closing (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m d4fdb02 2015-10-09 Fix Casing Bug (Will[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4740930 2015-10-08 Fix Stream Close Ordering Bug (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a594128 2015-10-07 Fix Auth Tests (Will[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4ee9881 2015-10-07 Add Nut Dependencies for NTLM (Will[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b1a9b08 2015-10-07 Update NTLM Nut Location (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a7f7977 2015-10-07 Add log4go Dependency (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 310710a 2015-10-07 Update NTLM Import String To Vendor Location (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 794735d 2015-10-07 Add ThomsonReuters NTLM Library (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7e21b52 2015-10-07 Remove Submodule and Add Nut Dependency (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 2579100 2015-10-07 Response to PR Feedback (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 3797654 2015-10-05 Fix NTLM Domain Handling Logic (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 15b1cac 2015-10-05 Clean Up Trace Statements (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 6665e1a 2015-10-05 Adding Log To Split NTLM Domain and User (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m d71daee 2015-10-05 Git Lfs NTLM Work (William Hipschman[32m[m)
[33m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   21159f8 2015-10-05 Merge from remote (William Hipschman[32m[m)
[33m|[m [1;34m|[m[1;35m\[m [1;33m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4c3976f 2015-09-01 More Trace Cleanup (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m db07670 2015-09-01 More Trace Cleanup (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 2666d2e 2015-09-01 Clean Up Debugging Traces (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 75307c3 2015-08-31 NTLM Fetch (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 619fa88 2015-08-28 NTLM Push (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0f7d782 2015-08-27 NTLM Toggle Switch and Basic Challenge/Response Code (William Hipschman[32m[m)
[33m|[m [1;34m|[m * [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 21ff0f6 2015-08-26 Add NTLM Library to git-lfs (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a0b8ebb 2015-10-05 More Trace Cleanup (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b4910ee 2015-10-05 More Trace Cleanup (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m c16c169 2015-10-05 Clean Up Debugging Traces (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 9bdca02 2015-10-05 NTLM Fetch (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 21f5379 2015-10-05 NTLM Push (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 383ac27 2015-10-05 NTLM Toggle Switch and Basic Challenge/Response Code (William Hipschman[32m[m)
[33m|[m * [1;35m|[m [1;33m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 624307a 2015-10-05 Add NTLM Library to git-lfs (William Hipschman[32m[m)
[33m|[m [1;35m|[m [1;35m|[m[1;35m_[m[1;33m|[m[1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[33m|[m [1;35m|[m[1;35m/[m[1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [1;35m|[m [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   b0318b2 2015-11-03 Merge pull request #812 from lbarbisan-ullink/master (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m 18f90c6 2015-11-03 Update nsis to repsect silent mode #789 (Laurent Barbisan[32m[m)
[1;36m|[m[1;36m/[m [1;35m/[m [1;35m/[m [1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;33m|[m [1;35m|[m [1;31m|[m [1;36m|[m   9a9ab03 2015-11-02 Merge pull request #806 from cjs/fix-api-spec-link (risk danger olson[32m[m)
[1;33m|[m[33m\[m [1;35m\[m [1;35m\[m [1;33m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;33m|[m [33m|[m[1;33m_[m[1;35m|[m[1;33m_[m[1;35m|[m[1;33m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;33m|[m[1;33m/[m[33m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;33m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 431d763 2015-10-30 Update spec location (Craig Steinberger[32m[m)
[1;33m|[m[1;33m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0566698 2015-10-28 ship v1.0.2 (risk danger olson[32m (tag: v1.0.2)[m)
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   552ffe7 2015-10-28 Merge pull request #801 from github/fix-rev-list-race (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[34m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m c3bc64b 2015-10-28 Fixes a race condition accessing the maps inside goroutines (risk danger olson[32m[m)
* [35m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   86e72c8 2015-10-28 Merge pull request #795 from github/smudge-send-size (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[35m|[m [1;31m|[m[35m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[35m|[m[35m/[m[1;31m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[35m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 25bbe79 2015-10-27 teach 'git lfs smudge' to send the expected size to the LFS api (risk danger olson[32m[m)
* [1;31m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   ca4b6d3 2015-10-27 Merge pull request #793 from github/fix-windows-integration (risk danger olson[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;31m|[m [1;33m|[m[1;31m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m afab33a 2015-10-27 update changelog to match release notes (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 16e362e 2015-10-27 fix the packagecloud script (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 2182d86 2015-10-27 remove exit and put tests in the right order (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b89d912 2015-10-27 update integration script to look for bash (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 589ab9b 2015-10-27 backslashes (risk danger olson[32m[m)
[1;31m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 626190f 2015-10-27 add debugging (risk danger olson[32m[m)
[1;31m|[m[1;31m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   a999185 2015-10-27 Merge pull request #772 from github/rel-1.0.1 (risk danger olson[32m (tag: v1.0.1)[m)
[1;34m|[m[1;35m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m * [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m   10084d5 2015-10-27 merge master (risk danger olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [1;36m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   ac8dd2e 2015-10-27 Merge pull request #792 from sinbad/test-no-orig-files (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m f12d409 2015-10-27 Make sure we don't accidentally run *.sh.orig (merge remnants) in test (Steve Streeting[32m[m)
[32m|[m[32m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[32m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   dea98b5 2015-10-26 merge master with test/docker improvements (risk danger olson[32m[m)
[32m|[m [34m|[m[32m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[32m|[m [34m|[m[32m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[32m|[m[32m/[m[34m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [34m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   b921667 2015-10-26 Merge pull request #790 from github/golang-test-runner (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [34m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m   c4f3a11 2015-10-26 merge master (risk danger olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [34m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m [1;32m|[m[36m/[m [34m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[36m|[m[36m/[m[1;32m|[m [34m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [1;32m|[m [34m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   81269bf 2015-10-26 Merge pull request #783 from andyneff/docker_hub (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [34m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m * [1;32m|[m [34m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b100472 2015-10-23 Change the run script to download docker instead of build (Andy Neff[32m[m)
[1;34m|[m [1;35m|[m [1;32m|[m[1;35m_[m[34m|[m[1;35m_[m[1;35m|[m[1;35m_[m[1;35m|[m[1;35m/[m [1;31m/[m [1;36m/[m  
[1;34m|[m [1;35m|[m[1;35m/[m[1;32m|[m [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;34m|[m [1;35m|[m * [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m cbafdbc 2015-10-26 update tests so they pass if GIT_LFS_TEST_DIR is unset (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m * [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m da3ced1 2015-10-26 don't format go files in these non-src dirs (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m * [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 47d167f 2015-10-26 return exit code 1 if any of the tests fail (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m * [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 807bb3e 2015-10-26 rewrite test runner in go to reduce output race conditions (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m[1;34m/[m [34m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;34m|[m[1;34m/[m[1;35m|[m [34m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;34m|[m [1;35m|[m * [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a5b2f07 2015-10-26 this is worthy of a mention (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m * [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   4e4dac2 2015-10-26 merge latest master (risk danger olson[32m[m)
[1;34m|[m [1;35m|[m [1;36m|[m[1;34m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m [1;35m|[m[1;34m_[m[1;36m|[m[1;34m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;34m|[m[1;34m/[m[1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   c2a0141 2015-10-26 Merge pull request #788 from github/whitelist-include-exclude-keys (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 585d496 2015-10-26 Allow lfs.fetchinclude and lfs.fetchexclude in .gitconfig (risk danger olson[32m[m)
* [33m|[m [1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m dba4c22 2015-10-26 skip lfs.batch in legacy download tests (risk danger olson[32m[m)
[33m|[m[33m/[m [1;35m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   8c655a8 2015-10-26 Merge pull request #771 from sinbad/windows-tests (risk danger olson[32m[m)
[1;35m|[m[35m\[m [1;35m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;35m|[m [35m|[m[1;35m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m e21d81b 2015-10-23 Fix worktree test on Windows when GIT_LFS_TEST_DIR set to native path (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 91379fe 2015-10-23 Use diff to compare elements for ease of use (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4b67bd7 2015-10-22 Fix worktree test on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7d9c2be 2015-10-22 Fix track test on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 47f1f11 2015-10-22 Fix submodule tests on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 2f2344d 2015-10-22 Actually stdin being inherited is a general problem (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7eb4b9b 2015-10-22 File open error messages slightly different on Windows vs Mac/Linux (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 9084bda 2015-10-22 Skip detached stdin test on Windows, is not a separate case (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0a29dfd 2015-10-22 Fix init test on Windows (may be general bug) (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 58332fc 2015-10-22 Fixed fsck tests on Windows - must use native_path (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 680199d 2015-10-22 Windows fsck test: output from shasum is slightly different in MinGW (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 8f89a0b 2015-10-22 Fix bare repo env test on Windows (new after rebase) (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 2a94f13 2015-10-22 Missed a string test (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 02c9020 2015-10-22 printf needs quotes (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m f2e1078 2015-10-22 No need to sort environ in LFS now that test handles it (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 22d546a 2015-10-22 Use order/blanks agnostic comparator for env tests (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 1671fc4 2015-10-22 Fix env tests again; now there can be 0 GIT_* env vars (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5964ec7 2015-10-22 Don't print trace at the end under Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m f91a4fc 2015-10-22 Fix stalling tests on Windows, custom fd for GIT_TRACE doesn't work (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0c2ad37 2015-10-22 Rename IS_MINGW_CYGWIN to IS_WINDOWS, simpler (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 49bf10d 2015-10-22 All env tests working on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 06c5bea 2015-10-22 3rd env test working on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 491515e 2015-10-22 Fix native_path on Mac/Linux (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 930f5ae 2015-10-22 Another env test fixed on Windows (Steve Streeting[32m[m)
[1;35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 4c1e5ac 2015-10-22 Fix the first Windows integration test error, paths (Steve Streeting[32m[m)
* [35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   9fdc0bc 2015-10-23 Merge pull request #776 from github/safe-config (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 6e6eb62 2015-10-23 change var name (risk danger olson[32m[m)
[36m|[m[36m/[m [35m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   0d09a0d 2015-10-23 Merge pull request #775 from github/fix-push-panic (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [35m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;32m|[m * [35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m f425cab 2015-10-23 guard against res being nil (risk danger olson[32m (chris/fix-push-panic)[m)
[1;32m|[m[1;32m/[m [35m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   97d4c13 2015-10-22 Merge pull request #763 from github/atomic-writes (risk danger olson[32m[m)
[35m|[m[1;35m\[m [35m\[m [1;36m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[35m|[m [1;35m|[m[35m/[m [1;36m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[35m|[m[35m/[m[1;35m|[m [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 24bef56 2015-10-22 lowest exit code here should be 129 (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b82f749 2015-10-22 dont clear temp objects when outside of a git repository (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 566a72d 2015-10-22 use Readdirnames instead of filepath.Walk (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a3a2369 2015-10-22 be specific about the signals we're watching (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7416550 2015-10-21 same exit code that commands.Exit() uses (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m fae68d8 2015-10-21 actually exit on interrupt/kill signals (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 3256b77 2015-10-21 lower the modtime check to 1 hour (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 0294185 2015-10-21 remove temp files if they dont match the '{oid}-{rand}' pattern (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m a0704be 2015-10-21 cleanup temp files automatically (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m d178505 2015-10-21 re-use test setup from earlier tests (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m cc5e258 2015-10-21 move smudge test, and exit line so allllll smudge tests run (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7194804 2015-10-21 yay go 1.5 supports (mostly) atomic renames on windows (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 43833d9 2015-10-20 add atomic package (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5d32559 2015-10-20 remove contentaddressable package (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 5848641 2015-10-20 echo messages help us track down the cause of test failures (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7ad041d 2015-10-20 put the pid in the temp filename (risk danger olson[32m[m)
[35m|[m * [1;36m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 65935dd 2015-10-20 rewrite the atomic file writes using 3rd party atomic.WriteFile() as a guide (risk danger olson[32m[m)
[35m|[m [1;35m|[m * [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 78697ff 2015-10-22 release 1.0.1 (risk danger olson[32m[m)
[35m|[m [1;35m|[m[35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
[35m|[m[35m/[m[1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   4cad027 2015-10-22 Merge pull request #770 from sinbad/fix-clone-exclude (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 417fd73 2015-10-22 Fix clone fail when fetch is excluded globally (#759) (Steve Streeting[32m[m)
[1;36m|[m[1;36m/[m [1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   761a35f 2015-10-21 Merge pull request #760 from github/config-whitelist (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m ef3a44a 2015-10-20 whitelist the valid keys from .gitconfig (risk danger olson[32m[m)
[32m|[m [1;35m|[m[1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   6cbc358 2015-10-21 Merge pull request #761 from github/doc-updates (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[34m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m b50a813 2015-10-21 mention build instructions (risk danger olson[32m (chris/doc-updates)[m)
[34m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m e77d5cc 2015-10-20 add build instructions and missing init flags (risk danger olson[32m[m)
[34m|[m [1;35m|[m[1;35m/[m [1;35m/[m [1;35m/[m [1;31m/[m [1;36m/[m  
* [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m   f56cd54 2015-10-21 Merge pull request #711 from github/smudge-new-api (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 7c2ecb8 2015-10-19 lfs.Download() should not attempt batch transfers every time (risk danger olson[32m[m)
[36m|[m * [1;35m|[m [1;35m|[m [1;35m|[m [1;31m|[m [1;36m|[m 1563366 2015-10-06 Use batch API for Download, fallback to legacy if unsupported (rubyist[32m[m)
[36m|[m [1;31m|[m [1;35m|[m[1;31m_[m[1;35m|[m[1;31m_[m[1;35m|[m[1;31m/[m [1;36m/[m  
[36m|[m [1;31m|[m[1;31m/[m[1;35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   
* [1;31m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   6c2bf06 2015-10-21 Merge pull request #765 from sinbad/attribute-global (risk danger olson[32m[m)
[1;35m|[m[1;33m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;35m|[m [1;33m|[m[1;35m_[m[1;31m|[m[1;35m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[1;33m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 95d6872 2015-10-21 Use --global to determine if needs to set a global attribute (Steve Streeting[32m[m)
[1;35m|[m[1;35m/[m [1;31m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
* [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   3a673d6 2015-10-20 Merge pull request #756 from github/fix-init-update-tests (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;34m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 7aa9c59 2015-10-19 fix lfs.InRepo() (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 7969729 2015-10-19 only skip these tests when run through docker suite (risk danger olson[32m[m)
* [1;35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   4915979 2015-10-20 Merge pull request #734 from sinbad/fetch-return-code (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m dddd170 2015-10-13 Make sure transfer errors are given enough context to be understandable (Steve Streeting[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 32f56cd 2015-10-13 Make fetch return non-zero error code when some downloads failed (Steve Streeting[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 835b40b 2015-10-13 Test server should return 404 from batch for "download" if missing (Steve Streeting[32m[m)
[1;36m|[m [1;31m|[m [1;35m|[m[1;31m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
[1;36m|[m [1;31m|[m[1;31m/[m[1;35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   
* [1;31m|[m [1;35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   91f3128 2015-10-20 Merge pull request #713 from sinbad/default-remote-behaviour (risk danger olson[32m[m)
[1;35m|[m[33m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;35m|[m [33m|[m[1;35m_[m[1;31m|[m[1;35m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[33m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 927eba0 2015-10-08 Refactor remote/ref derivation to allow for remotes with '/' in name (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 6ec4b57 2015-10-07 Use Exit() instead of Panic() for validation (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 1997c74 2015-10-06 Test for friendly fail case when default remote cannot be established (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 1e7cb85 2015-10-06 Validate user-supplied remote names in fetch/pull (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m d2f15d0 2015-10-06 Eliminate unnecessary duplicate retrieval of remote for reporting (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 28e6f28 2015-10-06 Implement improved DefaultRemote() method with better defaults & errors (Steve Streeting[32m[m)
[1;35m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 5dec48e 2015-10-06 Rename CurrentRemote() to RemoteForCurrentBranch() for clarity (Steve Streeting[32m[m)
[1;35m|[m [1;31m|[m[1;31m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
* [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   4cd71f6 2015-10-19 Merge pull request #692 from sinbad/fix-bare-repo (risk danger olson[32m[m)
[34m|[m[35m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[34m|[m * [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 9511e57 2015-09-30 Fix resolve of git dirs finding a parent non-bare repo if run in a bare repo (Steve Streeting[32m[m)
* [35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m   ecd754a 2015-10-19 Merge pull request #732 from github/panic-smudge (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [1;31m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[36m|[m * [35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m 26af3bc 2015-10-12 Keep logging the error, just exit 2 (rubyist[32m[m)
[36m|[m * [35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m f7c0a16 2015-10-12 Use Exit() instead of Panic() (rubyist[32m[m)
[36m|[m * [35m|[m [1;31m|[m [1;35m|[m [1;35m|[m [1;36m|[m d3e2430 2015-10-12 If a smudge operation fails, the smudge command should exit 1 (rubyist[32m[m)
[36m|[m [1;31m|[m [35m|[m[1;31m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
[36m|[m [1;31m|[m[1;31m/[m[35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   
* [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   239c331 2015-10-19 Merge pull request #735 from sinbad/checkout-index-errors (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;32m|[m * [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m 1c69c82 2015-10-13 Use LoggedError instead of Panic if update-index fails in checkout #725 (Steve Streeting[32m[m)
[1;32m|[m [1;31m|[m[1;31m/[m [35m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
* [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m 81563f2 2015-10-19 goimports (risk danger olson[32m[m)
* [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   5b1915e 2015-10-19 Merge pull request #718 from bozaro/stderr (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [35m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;34m|[m * [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m c5376d1 2015-10-08 Use separate buffers for stdout and stderr on executing git-lfs-authenticate (Artem V. Navrotskiy[32m[m)
[1;34m|[m [1;31m|[m[1;31m/[m [35m/[m [1;35m/[m [1;35m/[m [1;36m/[m  
* [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m   cf7b715 2015-10-19 Merge pull request #691 from sinbad/fix-smudge-skip-test (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [35m\[m [1;35m\[m [1;35m\[m [1;36m\[m  
[1;36m|[m * [1;31m|[m [35m|[m [1;35m|[m [1;35m|[m [1;36m|[m cf3fe60 2015-10-01 "smudge with skip" test would fail if GIT_LFS_TEST_DIR was not set (Steve Streeting[32m[m)
[1;36m|[m [1;35m|[m [1;31m|[m[1;35m_[m[35m|[m[1;35m/[m [1;35m/[m [1;36m/[m  
[1;36m|[m [1;35m|[m[1;35m/[m[1;31m|[m [35m|[m [1;35m|[m [1;36m|[m   
* [1;35m|[m [1;31m|[m [35m|[m [1;35m|[m [1;36m|[m   e08632f 2015-10-19 Merge pull request #710 from github/403 (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;31m\[m [35m\[m [1;35m\[m [1;36m\[m  
[32m|[m * [1;35m|[m [1;31m|[m [35m|[m [1;35m|[m [1;36m|[m 5c6e41c 2015-10-05 Always reveal underlying error (rubyist[32m[m)
[32m|[m * [1;35m|[m [1;31m|[m [35m|[m [1;35m|[m [1;36m|[m 6f5c95d 2015-10-05 Return the error in the 403 case (rubyist[32m[m)
[32m|[m [1;31m|[m [1;35m|[m[1;31m/[m [35m/[m [1;35m/[m [1;36m/[m  
[32m|[m [1;31m|[m[1;31m/[m[1;35m|[m [35m|[m [1;35m|[m [1;36m|[m   
* [1;31m|[m [1;35m|[m [35m|[m [1;35m|[m [1;36m|[m   b8241db 2015-10-19 Merge pull request #690 from WillHipschman/AuthFix (risk danger olson[32m[m)
[1;31m|[m[35m\[m [1;31m\[m [1;35m\[m [35m\[m [1;35m\[m [1;36m\[m  
[1;31m|[m [35m|[m[1;31m/[m [1;35m/[m [35m/[m [1;35m/[m [1;36m/[m  
[1;31m|[m[1;31m/[m[35m|[m [1;35m|[m [35m|[m [1;35m|[m [1;36m|[m   
[1;31m|[m * [1;35m|[m [35m|[m [1;35m|[m [1;36m|[m 5ae638e 2015-10-06 Add Test To Validate Capital Letters Work in URLs (Will[32m[m)
[1;31m|[m * [1;35m|[m [35m|[m [1;35m|[m [1;36m|[m e5b4775 2015-10-01 Clean Up Whitespace (William Hipschman[32m[m)
[1;31m|[m * [1;35m|[m [35m|[m [1;35m|[m [1;36m|[m 17ceb93 2015-10-01 Fix For Infinite Auth Redirect (William Hipschman[32m[m)
[1;31m|[m [35m|[m [1;35m|[m [35m|[m [1;35m|[m * 7a86b2b 2015-10-07 Test with bad credentials in netrc (rubyist[32m[m)
[1;31m|[m [35m|[m [1;35m|[m [35m|[m [1;35m|[m * 91195f6 2015-10-07 Add integration tests for netrc (rubyist[32m[m)
[1;31m|[m [35m|[m [1;35m|[m [35m|[m [1;35m|[m * 8422b0e 2015-10-06 Basic netrc support for LFS API requests (rubyist[32m[m)
[1;31m|[m [35m|[m[1;31m_[m[1;35m|[m[1;31m_[m[35m|[m[1;31m_[m[1;35m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[35m|[m [1;35m|[m [35m|[m [1;35m|[m   
* [35m|[m [1;35m|[m [35m|[m [1;35m|[m   f32da15 2015-10-05 Merge pull request #702 from sinbad/fetch-include-exclude-doc-fix (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [1;35m\[m [35m\[m [1;35m\[m  
[36m|[m * [35m|[m [1;35m|[m [35m|[m [1;35m|[m 0c92ee9 2015-10-05 Fixed fetch include/exclude examples, should not have '=' (Steve Streeting[32m[m)
[36m|[m [1;35m|[m [35m|[m[1;35m/[m [35m/[m [1;35m/[m  
[36m|[m [1;35m|[m[1;35m/[m[35m|[m [35m|[m [1;35m|[m   
* [1;35m|[m [35m|[m [35m|[m [1;35m|[m   49a4ed1 2015-10-05 Merge pull request #704 from revi/typofix (risk danger olson[32m[m)
[1;35m|[m[1;33m\[m [1;35m\[m [35m\[m [35m\[m [1;35m\[m  
[1;35m|[m [1;33m|[m[1;35m/[m [35m/[m [35m/[m [1;35m/[m  
[1;35m|[m[1;35m/[m[1;33m|[m [35m|[m [35m|[m [1;35m|[m   
[1;35m|[m * [35m|[m [35m|[m [1;35m|[m 2ccc084 2015-10-06 Fix typo (Yongmin Hong[32m[m)
[1;35m|[m[1;35m/[m [35m/[m [35m/[m [1;35m/[m  
* [35m|[m [35m|[m [1;35m|[m   e98e470 2015-10-01 Merge pull request #689 from kaleworsley/fix_typos (Scott Barron[32m[m)
[35m|[m[1;35m\[m [35m\[m [35m\[m [1;35m\[m  
[35m|[m [1;35m|[m[35m/[m [35m/[m [1;35m/[m  
[35m|[m[35m/[m[1;35m|[m [35m|[m [1;35m|[m   
[35m|[m * [35m|[m [1;35m|[m eef8c9b 2015-10-02 resmoves -> removes. (Kale Worsley[32m[m)
[35m|[m * [35m|[m [1;35m|[m 3953a90 2015-10-02 commmands_pre_push.go -> command_pre_push.go. (Kale Worsley[32m[m)
[35m|[m * [35m|[m [1;35m|[m 33b1524 2015-10-02 attribtues -> attributes. (Kale Worsley[32m[m)
[35m|[m[35m/[m [35m/[m [1;35m/[m  
* [35m|[m [1;35m|[m   7379fa9 2015-09-30 Merge pull request #686 from github/release-1.0-master (risk danger olson[32m (tag: v1.0.0)[m)
[1;36m|[m[31m\[m [35m\[m [1;35m\[m  
[1;36m|[m * [35m\[m [1;35m\[m   3b33670 2015-09-30 Merge branch 'master' into release-1.0-master (Rick Olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [35m\[m [1;35m\[m  
[1;36m|[m [32m|[m[1;36m/[m [35m/[m [1;35m/[m  
[1;36m|[m[1;36m/[m[32m|[m [35m|[m [1;35m|[m   
* [32m|[m [35m|[m [1;35m|[m   a93b963 2015-09-30 Merge pull request #685 from sinbad/fix-worktree-test (risk danger olson[32m[m)
[35m|[m[35m\[m [32m\[m [35m\[m [1;35m\[m  
[35m|[m [35m|[m[35m_[m[32m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[35m|[m [32m|[m [1;35m|[m   
[35m|[m * [32m|[m [1;35m|[m a2174c9 2015-09-30 Fix worktree test (Steve Streeting[32m[m)
[35m|[m [35m|[m * [1;35m|[m 7429458 2015-09-30 we need the rootdir, actually (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m 2d4da5f 2015-09-30 hardcode docker's build path (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m 23964ec 2015-09-30 look for GIT_LFS_BUILD_DIR (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m f80238a 2015-09-30 skip the init --local test too (Rick Olson[32m (chris/release-1.0-hotfix, chris/release-1.0)[m)
[35m|[m [35m|[m * [1;35m|[m e608164 2015-09-30 skip failing tests for docker scripts (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m b231160 2015-09-30 push GIT_TERMINAL_PROMPT tests to separate file so they run only on git 2.3+ (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m 49c4522 2015-09-30 set GIT_LFS_TEST_DIR for docker tests (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m bb9e848 2015-09-30 bump version (Rick Olson[32m[m)
[35m|[m [35m|[m * [1;35m|[m 08da1f8 2015-09-30 fix CoC link (Rick Olson[32m[m)
[35m|[m [35m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[35m|[m [1;35m|[m   
* [35m|[m [1;35m|[m   8d3e39a 2015-09-30 Merge pull request #681 from github/ls-files-output (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [1;35m\[m  
[35m|[m [1;31m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[1;31m|[m [1;35m|[m   
[35m|[m * [1;35m|[m acc99eb 2015-09-25 update man page for new ls-files option (Rick Olson[32m[m)
[35m|[m * [1;35m|[m 2082b26 2015-09-25 tweak ls-files output (Rick Olson[32m[m)
[35m|[m * [1;35m|[m 21bc33e 2015-09-25 ls-files show duplicate with OID(s) (Bhuridech Sudsee[32m[m)
[35m|[m * [1;35m|[m 9f88f2f 2015-09-25 ls-files show duplicate (Bhuridech Sudsee[32m[m)
* [1;31m|[m [1;35m|[m   a1bef30 2015-09-29 Merge pull request #684 from github/smudge-tiny-files (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;35m\[m  
[1;32m|[m * [1;31m|[m [1;35m|[m f58db7f 2015-09-25 return a NotAPointerError if it has an empty version (Rick Olson[32m[m)
[1;32m|[m * [1;31m|[m [1;35m|[m 4e7bc74 2015-09-25 Better handling for tiny files below the blob size threshold that contain 'git-lfs' (Rick Olson[32m[m)
* [1;33m|[m [1;31m|[m [1;35m|[m   b4fd1c9 2015-09-28 Merge pull request #682 from github/better-readme (risk danger olson[32m[m)
[1;33m|[m[1;35m\[m [1;33m\[m [1;31m\[m [1;35m\[m  
[1;33m|[m [1;35m|[m[1;33m/[m [1;31m/[m [1;35m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [1;31m|[m [1;35m|[m   
[1;33m|[m * [1;31m|[m [1;35m|[m 2f5b120 2015-09-25 add loranallensmith's video (Rick Olson[32m (chris/better-readme)[m)
[1;33m|[m * [1;31m|[m [1;35m|[m bd2f4bc 2015-09-25 Add some more details on installation and documentation (Rick Olson[32m[m)
[1;33m|[m[1;33m/[m [1;31m/[m [1;35m/[m  
* [1;31m|[m [1;35m|[m   d402f5b 2015-09-25 Merge pull request #679 from github/smudge-passthrough (risk danger olson[32m[m)
[1;31m|[m[31m\[m [1;31m\[m [1;35m\[m  
[1;31m|[m [31m|[m[1;31m/[m [1;35m/[m  
[1;31m|[m[1;31m/[m[31m|[m [1;35m|[m   
[1;31m|[m * [1;35m|[m 7a5e296 2015-09-25 update vars (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m c096ad3 2015-09-23 update docs in man page (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m 8d1875a 2015-09-23 add filter.lfs configs to 'git lfs env' (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m 8cd51c0 2015-09-23 rename the options to SKIP_SMUDGE and --skip-smudge (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m bc025a8 2015-09-23 better tests (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m f8715e5 2015-09-23 teach init about --smudge-passthrough (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m 82ac4ac 2015-09-23 teach init about --local flag (Rick Olson[32m[m)
[1;31m|[m * [1;35m|[m 5729eac 2015-09-23 teach smudge to skip lfs activities with GIT_LFS_SMUDGE_PASSTHROUGH=1 (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m [1;35m/[m  
* [1;35m|[m   68f98da 2015-09-23 Merge pull request #659 from github/packagecloud-script (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m 10016f4 2015-09-11 Quick script to upload all the packages to packagecloud (Rick Olson[32m[m)
* [33m|[m [1;35m|[m   3feaca1 2015-09-23 Merge pull request #670 from github/docgen-fixes (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;35m\[m  
[34m|[m * [33m|[m [1;35m|[m a0514b4 2015-09-17 show the entire section title (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 795a1eb 2015-09-17 remove all Short messages (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m f604acf 2015-09-17 update help template so cmd.Short is not shown (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m ff8c102 2015-09-17 clean up surrounding whitespace in the help output (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 50cf12e 2015-09-17 remove html linebreaks (<br>) too. (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 62b1df8 2015-09-17 remove empty section (rick[32m[m)
* [35m|[m [33m|[m [1;35m|[m   59fad47 2015-09-23 Merge pull request #671 from github/update-in-bare-repo (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m * [35m|[m [33m|[m [1;35m|[m 380c46f 2015-09-17 set GIT_LFS_TEST_DIR for travis tests (rick[32m[m)
[36m|[m * [35m|[m [33m|[m [1;35m|[m 072e778 2015-09-17 more info to debug failures (rick[32m[m)
[36m|[m * [35m|[m [33m|[m [1;35m|[m 4b1cd58 2015-09-17 teach init too (rick[32m[m)
[36m|[m * [35m|[m [33m|[m [1;35m|[m be6cc73 2015-09-17 teach 'update' cmd how to install hooks in bare repos (rick[32m[m)
* [1;31m|[m [35m|[m [33m|[m [1;35m|[m   d6ba089 2015-09-22 Merge pull request #677 from github/nicer-error (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[1;32m|[m * [1;31m|[m [35m|[m [33m|[m [1;35m|[m d4df865 2015-09-22 The original error string provides this info, and more (rubyist[32m[m)
[1;32m|[m[1;32m/[m [1;31m/[m [35m/[m [33m/[m [1;35m/[m  
* [1;31m|[m [35m|[m [33m|[m [1;35m|[m   8e2638b 2015-09-21 Merge pull request #676 from github/better-server-failure-handling-tests (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[1;34m|[m * [1;31m\[m [35m\[m [33m\[m [1;35m\[m   2cdac01 2015-09-21 merge master (rick[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [1;31m/[m [35m/[m [33m/[m [1;35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [1;31m|[m [35m|[m [33m|[m [1;35m|[m   
* [1;36m|[m [1;31m|[m [35m|[m [33m|[m [1;35m|[m   d3b149c 2015-09-21 Merge pull request #675 from github/legacy-retries (risk danger olson[32m[m)
[1;31m|[m[33m\[m [1;36m\[m [1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[1;31m|[m [33m|[m[1;31m_[m[1;36m|[m[1;31m/[m [35m/[m [33m/[m [1;35m/[m  
[1;31m|[m[1;31m/[m[33m|[m [1;36m|[m [35m|[m [33m|[m [1;35m|[m   
[1;31m|[m * [1;36m|[m [35m|[m [33m|[m [1;35m|[m 8a5c0da 2015-09-21 Move the batcher nil check down so retries work in "legacy" mode (rubyist[32m[m)
[1;31m|[m[1;31m/[m [1;36m/[m [35m/[m [33m/[m [1;35m/[m  
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 9e6a5d3 2015-09-21 add tests for bad dns name in lfs endpoint (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 3671cc7 2015-09-21 check for errors before checking for a link action (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m b85954d 2015-09-21 use bash instead of shell (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m b3a8eab 2015-09-21 update legacy test server to return http error responses instead (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 0aceedf 2015-09-21 dont share a remote repo (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m a8aa41f 2015-09-21 add more push tests for the batch and legacy APIs (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 8be2dc4 2015-09-21 forgot to add a content handler for storage 403 responses (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 2c58589 2015-09-21 initial storage server tests (rick[32m[m)
[1;31m|[m * [35m|[m [33m|[m [1;35m|[m 54601c6 2015-09-21 introduce calc_oid() helper for integration tests (rick[32m[m)
[1;31m|[m[1;31m/[m [35m/[m [33m/[m [1;35m/[m  
* [35m|[m [33m|[m [1;35m|[m a62e510 2015-09-17 use -X correctly for go 1.5 (rick[32m[m)
[35m|[m[35m/[m [33m/[m [1;35m/[m  
* [33m|[m [1;35m|[m   53bcb75 2015-09-17 Merge pull request #665 from sinbad/help-from-man (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;35m\[m  
[34m|[m * [33m|[m [1;35m|[m 0ea92d8 2015-09-17 Also format out links to other manpages so they're easier to read (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 625f4a9 2015-09-17 Reformat links as well (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 84a3ef9 2015-09-15 More extensive re-formatting of man page to look better for --help (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m b558b71 2015-09-15 Revise how we determine the command name to deal with all variants (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 34faace 2015-09-15 Convert error output from go generate to string before printing (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m e970264 2015-09-14 Use man content for `help <cmd>' and '<cmd> --help' (Steve Streeting[32m[m)
* [35m|[m [33m|[m [1;35m|[m   f0fe845 2015-09-17 Merge pull request #669 from github/ls-files-tests (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m * [35m|[m [33m|[m [1;35m|[m 54753c8 2015-09-17 add an integration test to prove #668 works (rick[32m[m)
[36m|[m[36m/[m [35m/[m [33m/[m [1;35m/[m  
* [35m|[m [33m|[m [1;35m|[m   aaed3a9 2015-09-17 Merge pull request #668 from Aorjoa/master (risk danger olson[32m[m)
[35m|[m[1;33m\[m [35m\[m [33m\[m [1;35m\[m  
[35m|[m [1;33m|[m[35m/[m [33m/[m [1;35m/[m  
[35m|[m[35m/[m[1;33m|[m [33m|[m [1;35m|[m   
[35m|[m * [33m|[m [1;35m|[m   e711c99 2015-09-17 Merge for fixed #664 (Bhuridech Sudsee[32m[m)
[35m|[m [35m|[m[1;35m\[m [33m\[m [1;35m\[m  
[35m|[m[35m/[m [1;35m/[m [33m/[m [1;35m/[m  
[35m|[m * [33m|[m [1;35m|[m 9bf3cff 2015-09-17 Fixed #664 check empty ref of ResolveRef (Bhuridech Sudsee[32m[m)
[35m|[m[35m/[m [33m/[m [1;35m/[m  
* [33m|[m [1;35m|[m   ccbc6f6 2015-09-14 Merge pull request #658 from github/ci-go1.5 (risk danger olson[32m[m)
[33m|[m[31m\[m [33m\[m [1;35m\[m  
[33m|[m [31m|[m[33m/[m [1;35m/[m  
[33m|[m[33m/[m[31m|[m [1;35m|[m   
[33m|[m * [1;35m|[m 492a7c3 2015-09-11 Stop testing Go 1.3 and 1.4.2 (risk danger olson[32m[m)
[33m|[m[33m/[m [1;35m/[m  
* [1;35m|[m   6dcc6d0 2015-09-11 Merge pull request #657 from andyneff/rpm_remove_git_req (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m dd60f0e 2015-09-11 Clean up (Andy Neff[32m[m)
[32m|[m * [1;35m|[m aa260f1 2015-09-11 Remove git requirement to make custom installs work (Andy Neff[32m[m)
[32m|[m[32m/[m [1;35m/[m  
* [1;35m|[m   684898b 2015-09-11 Merge pull request #654 from andyneff/docker_golang_update (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m  
[34m|[m * [1;35m|[m 88430de 2015-09-11 Added go 1.5.1 to docker scripts (via binary distro) (Andy Neff[32m[m)
* [35m|[m [1;35m|[m   51d462b 2015-09-11 Merge pull request #656 from sinbad/testutils-fixes2 (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [1;35m\[m  
[35m|[m [1;31m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[1;31m|[m [1;35m|[m   
[35m|[m * [1;35m|[m e52ce9c 2015-09-11 Make sure random test data is seeded differently between commits (Steve Streeting[32m[m)
[35m|[m[35m/[m [1;35m/[m  
* [1;35m|[m 183bdc8 2015-09-10 prioritize the builds i typically care most (rick[32m (tag: v0.6.0)[m)
* [1;35m|[m feda7d6 2015-09-10 remove old windows installer (rick[32m[m)
* [1;35m|[m   0de6370 2015-09-10 Merge pull request #652 from github/ghe-token-hack (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;35m\[m  
[1;32m|[m * [1;35m|[m 53256ad 2015-09-10 update changelog to match (rick[32m[m)
[1;32m|[m * [1;35m|[m 228bea7 2015-09-10 Skip credential check if url has ?token (rick[32m[m)
[1;32m|[m[1;32m/[m [1;35m/[m  
* [1;35m|[m b8ed855 2015-09-10 add the checkout/pull man pages too (rick[32m[m)
* [1;35m|[m   ea89e46 2015-09-10 Merge pull request #651 from github/release-v0.6.0 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;35m\[m  
[1;34m|[m * [1;35m|[m d9e9ea3 2015-09-10 unnecessary function (rick[32m[m)
[1;34m|[m * [1;35m|[m 2343592 2015-09-10 dont upload windows builds from script/release (rick[32m[m)
[1;34m|[m * [1;35m|[m 1bace61 2015-09-10 link up the man pages (rick[32m[m)
[1;34m|[m * [1;35m|[m 3b1099e 2015-09-10 poing to the updated man pages (rick[32m[m)
[1;34m|[m * [1;35m|[m e07e5c2 2015-09-10 update the changelog with @sinbad's suggestions (rick[32m[m)
[1;34m|[m * [1;35m|[m 6480bff 2015-09-10 promote extensions to real feature (rick[32m[m)
[1;34m|[m * [1;35m|[m c47cd71 2015-09-10 add installation items (rick[32m[m)
[1;34m|[m * [1;35m|[m b78cd7a 2015-09-10 initial changelog (rick[32m[m)
* [1;35m|[m [1;35m|[m   bfe39ab 2015-09-10 Merge pull request #650 from github/batch-access-check (risk danger olson[32m[m)
[1;35m|[m[31m\[m [1;35m\[m [1;35m\[m  
[1;35m|[m [31m|[m[1;35m/[m [1;35m/[m  
[1;35m|[m[1;35m/[m[31m|[m [1;35m|[m   
[1;35m|[m * [1;35m|[m 35cb151 2015-09-10 always consult PrivateAccess() for batch API calls (rick[32m[m)
[1;35m|[m[1;35m/[m [1;35m/[m  
* [1;35m|[m 19105ed 2015-09-10 better wording in the man page (rick[32m[m)
* [1;35m|[m   63371a3 2015-09-10 Merge pull request #642 from github/nsis (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m fe52565 2015-09-09 $INSTDIR should always be a path (risk danger olson[32m[m)
[32m|[m * [1;35m|[m 9dcb73a 2015-09-08 copy to the 'proxy git' dir too (Rick Olson[32m[m)
[32m|[m * [1;35m|[m 48212a3 2015-09-08 fix the version and copyright metadata (Rick Olson[32m[m)
[32m|[m * [1;35m|[m 6f4ff26 2015-09-06 modify text (Rick Olson[32m[m)
[32m|[m * [1;35m|[m eee4bd6 2015-09-06 initial nsis script (Rick Olson[32m[m)
* [33m|[m [1;35m|[m   aa9d204 2015-09-10 Merge pull request #646 from github/push-all-to-remote (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;35m\[m  
[34m|[m * [33m|[m [1;35m|[m a3ed316 2015-09-10 update push docs (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 8214aee 2015-09-09 extract scanAll() from fetchAll() so the push cmd can use it too (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m 0a7e94c 2015-09-09 hide the download check queue. (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m 75d754a 2015-09-09 use the channel scanner when pushing ALL objects (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m bff5aa7 2015-09-09 'grep -c' is less brittle than 'wc -l' (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m 0c6ae0e 2015-09-09 unnecessary trace msg (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m df2007d 2015-09-09 teach 'push --all' to skip objects not in .git/lfs/objects but already on the server (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m 1ff901b 2015-09-09 rework 'push --all' so it's consistent with fetch (Rick Olson[32m[m)
[34m|[m * [33m|[m [1;35m|[m e680792 2015-09-08 update man pages for push (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 174cda9 2015-09-08 add a test for --all (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m 6bd3b9f 2015-09-08 add --all flag to push (rick[32m[m)
[34m|[m * [33m|[m [1;35m|[m ffca2e0 2015-09-08 update push dry-run output so paths and oids line up. (rick[32m[m)
* [35m|[m [33m|[m [1;35m|[m   16a277d 2015-09-10 Merge pull request #595 from github/retries (Scott Barron[32m[m)
[36m|[m[1;31m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m * [35m|[m [33m|[m [1;35m|[m 9008651 2015-09-09 fix another documentation again (rubyist[32m[m)
[36m|[m * [35m|[m [33m|[m [1;35m|[m d8916d8 2015-09-09 fix another documentation (rubyist[32m[m)
[36m|[m * [35m|[m [33m|[m [1;35m|[m   e052e0e 2015-09-09 Merge branch 'master' into retries (rubyist[32m[m)
[36m|[m [1;32m|[m[1;33m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m * [1;33m|[m [35m|[m [33m|[m [1;35m|[m 5f183ce 2015-09-09 Update some documentation, reset the retrying flag in TQ (rubyist[32m[m)
[36m|[m * [1;33m|[m [35m|[m [33m|[m [1;35m|[m e48d6fe 2015-09-09 The batcher Reset can be implicit (rubyist[32m[m)
[36m|[m * [1;33m|[m [35m|[m [33m|[m [1;35m|[m 66aa279 2015-09-08 Fix a logic error, make sure the batcher exits on retry, add some tracing (rubyist[32m[m)
[36m|[m * [1;33m|[m [35m|[m [33m|[m [1;35m|[m 826e519 2015-09-04 Withhold retries until the end, then retry (rubyist[32m[m)
[36m|[m * [1;33m|[m [35m|[m [33m|[m [1;35m|[m   21bd924 2015-09-04 Merge branch 'master' into retries (rubyist[32m[m)
[36m|[m [1;34m|[m[33m\[m [1;33m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m [1;34m|[m [33m|[m [1;33m|[m[33m_[m[35m|[m[33m/[m [1;35m/[m  
[36m|[m [1;34m|[m [33m|[m[33m/[m[1;33m|[m [35m|[m [1;35m|[m   
[36m|[m * [33m|[m [1;33m|[m [35m|[m [1;35m|[m b5d02fc 2015-09-04 Push retry counting into the transferables, remove mutex wrapped map (rubyist[32m[m)
[36m|[m * [33m|[m [1;33m|[m [35m|[m [1;35m|[m a45802c 2015-09-03 Retry uploads that get a 403 due to token expiry (rubyist[32m[m)
[36m|[m * [33m|[m [1;33m|[m [35m|[m [1;35m|[m   375c056 2015-09-02 Merge branch 'master' into retries (rubyist[32m[m)
[36m|[m [1;36m|[m[31m\[m [33m\[m [1;33m\[m [35m\[m [1;35m\[m  
[36m|[m * [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m 0556a99 2015-09-02 Limit the number of batch endpoint retries (rubyist[32m[m)
[36m|[m * [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m a3676a1 2015-09-02 Move this retriable error up higher (rubyist[32m[m)
[36m|[m * [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m   d5d33a3 2015-09-02 Merge branch 'master' into retries (rubyist[32m[m)
[36m|[m [32m|[m[33m\[m [31m\[m [33m\[m [1;33m\[m [35m\[m [1;35m\[m  
[36m|[m * [33m|[m [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m 477ab63 2015-09-02 Reintroduce the retriable error under the new error system (rubyist[32m[m)
[36m|[m * [33m|[m [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m   e617264 2015-09-01 Merge branch 'master' into retries (rubyist[32m[m)
[36m|[m [34m|[m[35m\[m [33m\[m [31m\[m [33m\[m [1;33m\[m [35m\[m [1;35m\[m  
[36m|[m * [35m|[m [33m|[m [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m d06ec80 2015-08-18 Start a retriable error (rubyist[32m[m)
[36m|[m * [35m|[m [33m|[m [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m 6acadd5 2015-08-17 Retry transfer failures (rubyist[32m[m)
* [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;33m|[m [35m|[m [1;35m|[m   e8a5cfb 2015-09-09 Merge pull request #647 from sinbad/testutils-fixes (risk danger olson[32m[m)
[1;33m|[m[1;31m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;33m\[m [35m\[m [1;35m\[m  
[1;33m|[m [1;31m|[m[1;33m_[m[35m|[m[1;33m_[m[35m|[m[1;33m_[m[33m|[m[1;33m_[m[31m|[m[1;33m_[m[33m|[m[1;33m/[m [35m/[m [1;35m/[m  
[1;33m|[m[1;33m/[m[1;31m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [35m|[m [1;35m|[m   
[1;33m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [35m|[m [1;35m|[m 7216ffc 2015-09-09 Refactor initialisation of global dirs into reusable function (Steve Streeting[32m[m)
[1;33m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [35m|[m [1;35m|[m 6a2fdac 2015-09-09 Make sure data in AddCommits really is uniquely seeded for all files (Steve Streeting[32m[m)
[1;33m|[m[1;33m/[m [35m/[m [35m/[m [33m/[m [31m/[m [33m/[m [35m/[m [1;35m/[m  
* [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [35m|[m [1;35m|[m   8d50ce4 2015-09-08 Merge pull request #644 from github/in-repo-check (risk danger olson[32m[m)
[35m|[m[1;33m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [35m\[m [1;35m\[m  
[35m|[m [1;33m|[m[35m_[m[35m|[m[35m_[m[35m|[m[35m_[m[33m|[m[35m_[m[31m|[m[35m_[m[33m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[1;33m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m f6d1f99 2015-09-08 skip these tests if GIT_LFS_TEST_DIR is unset (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 416eb28 2015-09-08 wrap the common behavior in requireInRepo() (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 4174867 2015-09-08 test the exit code of the commands (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 322ace4 2015-09-08 use fd 5, since end_test uses 3/4 already (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 027f33c 2015-09-08 add extra line after displaying stdout/err/trace logs (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m ba21e8e 2015-09-08 it's a unix system. i know this (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 62e9e7d 2015-09-08 reset the trace file between tests (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 43e9f37 2015-09-08 fix behavior of 7 commands when run outside git repo (rick[32m[m)
[35m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m cf9b7b5 2015-09-08 don't show the verbose output (server logs and env) when run locally (rick[32m[m)
* [1;33m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   88dbe9f 2015-09-08 Merge pull request #620 from rtyley/tree-entry-for-pointer-file-preserves-exec-bit (risk danger olson[32m[m)
[1;33m|[m[1;35m\[m [1;33m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;35m\[m  
[1;33m|[m [1;35m|[m[1;33m/[m [35m/[m [35m/[m [33m/[m [31m/[m [33m/[m [1;35m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   
[1;33m|[m * [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 8d075a8 2015-08-31 Fix spec: tree entry for pointer file preserves exec bit (Roberto Tyley[32m[m)
* [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   bd1a6d7 2015-09-08 Merge pull request #634 from github/401legacy (Scott Barron[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;35m\[m  
[1;36m|[m * [1;35m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;35m\[m   daeaa3d 2015-09-08 Merge branch 'master' into 401legacy (rubyist[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;35m\[m  
[1;36m|[m [32m|[m[1;36m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [31m/[m [33m/[m [1;35m/[m  
[1;36m|[m[1;36m/[m[32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   
* [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m   60eba2d 2015-09-07 Merge pull request #629 from github/access-types (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [31m\[m [33m\[m [1;35m\[m  
[34m|[m * [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 5d4e201 2015-09-02 no need to pass the local .git/config path to SetLocal() anymore (Rick Olson[32m[m)
[34m|[m * [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 9ee483b 2015-09-02 teach "git lfs update" how to update the private access values (Rick Olson[32m[m)
[34m|[m * [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 7dda3c8 2015-09-02 teach SetLocal() and UnsetLocalKey() how to make config changes to .git/config (Rick Olson[32m[m)
[34m|[m * [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m 1eda75a 2015-09-02 add some Access() tests (Rick Olson[32m[m)
[34m|[m * [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [31m|[m [33m|[m [1;35m|[m ae23360 2015-09-02 remove AuthTypes (Rick Olson[32m[m)
[34m|[m [31m|[m [32m|[m[31m_[m[1;35m|[m[31m_[m[35m|[m[31m_[m[35m|[m[31m_[m[33m|[m[31m/[m [33m/[m [1;35m/[m  
[34m|[m [31m|[m[31m/[m[32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m   
* [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m   c879e6d 2015-09-07 Merge pull request #635 from sinbad/pre-push-multiple-branches (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [33m\[m [1;35m\[m  
[36m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 6c98f39 2015-09-07 Refactor to reduce size of nested code in for loop (Steve Streeting[32m[m)
[36m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 822249a 2015-09-07 Continue rather than return when encountering delete branch in pre-push (Steve Streeting[32m[m)
[36m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m d758975 2015-09-07 Fix pre-push hook when multiple branches are pushed in one `git push` (Steve Streeting[32m[m)
[36m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m e593529 2015-09-07 Failing test (Steve Streeting[32m[m)
* [1;31m|[m [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m   db615bd 2015-09-07 Merge pull request #633 from sinbad/fetch-all (risk danger olson[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m [31m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [33m\[m [1;35m\[m  
[1;31m|[m [1;33m|[m[1;31m/[m [31m/[m [32m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [33m/[m [1;35m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m   
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 73d01bc 2015-09-07 Use `*` instead of `Done` on spinner completion on Windows (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 799ee1e 2015-09-07 Test for fetch --all (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m abfbca3 2015-09-07 Improve messages (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 4787266 2015-09-07 Fix validation (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 15ec9f0 2015-09-07 Remove duplicate validation (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 51f13d6 2015-09-07 Introduce spinner & use in ScanRefs part of fetch --all (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 05ee956 2015-09-07 Add the option to receive scan results on demand (Steve Streeting[32m[m)
[1;31m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m 8a36ba8 2015-09-07 Implement basic fetch --all (Steve Streeting[32m[m)
[1;31m|[m[1;31m/[m [31m/[m [32m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [33m/[m [1;35m/[m  
* [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [33m|[m [1;35m|[m c1ac3d4 2015-09-07 change the version so pre-releases don't look like 0.5.x releases (rick[32m[m)
[33m|[m [31m|[m[33m_[m[32m|[m[33m_[m[1;35m|[m[33m_[m[35m|[m[33m_[m[35m|[m[33m_[m[33m|[m[33m/[m [1;35m/[m  
[33m|[m[33m/[m[31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
* [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   d4462c9 2015-09-04 Merge pull request #638 from github/init-test-fix (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;34m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m cf43c70 2015-09-04 test for bad clean or smudge attribute (Your Name[32m[m)
[1;34m|[m[1;34m/[m [31m/[m [32m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
* [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   10dbfb5 2015-09-04 Merge pull request #637 from github/errorwait (Scott Barron[32m[m)
[1;36m|[m[31m\[m [31m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;36m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 512bad1 2015-09-04 TransferQueue's Wait() needs to wait until all errors have been collected (rubyist[32m[m)
[1;36m|[m[1;36m/[m [31m/[m [32m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[1;36m|[m [31m|[m * [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   eb7e210 2015-09-04 Merge branch 'master' into 401legacy (rubyist[32m[m)
[1;36m|[m [31m|[m [32m|[m[1;36m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;36m|[m [31m|[m[1;36m_[m[32m|[m[1;36m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[1;36m|[m[1;36m/[m[31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
* [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   9cbf64e 2015-09-04 Merge pull request #636 from github/tracetests (Scott Barron[32m[m)
[34m|[m[35m\[m [31m\[m [32m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[34m|[m * [31m|[m [32m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m ae9f89f 2015-09-04 Add GIT_TRACE output to integration tests, display on failure (rubyist[32m[m)
[34m|[m[34m/[m [31m/[m [32m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[34m|[m [31m|[m * [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m ee41de0 2015-09-04 Handle 401 for the legacy api path the same way batch handles it (rubyist[32m[m)
[34m|[m [31m|[m[34m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[34m|[m[34m/[m[31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
* [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   05fea2e 2015-09-03 Merge pull request #631 from sinbad/fix-worktree-test (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[36m|[m * [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 2e07d0c 2015-09-03 Fix worktree test now that batch is enabled by default (Steve Streeting[32m[m)
* [1;31m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   7255a51 2015-09-03 Merge pull request #628 from github/config-bool-checks (risk danger olson[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m [31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;31m|[m [1;33m|[m[1;31m/[m [31m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
[1;31m|[m * [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 6f3d070 2015-09-02 write a common config bool parser (Rick Olson[32m[m)
[1;31m|[m * [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m bec37b0 2015-09-02 simpler batch transfer bool check (Rick Olson[32m[m)
[1;31m|[m [31m|[m[31m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
* [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   411f88b 2015-09-02 Merge pull request #630 from github/travis-go-versions (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;34m|[m * [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m b09a95a 2015-09-01 test with multiple go versions (risk danger olson[32m[m)
* [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   a227c22 2015-09-02 Merge pull request #619 from ttaylorr/promote-hooks (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;36m|[m * [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m e56d630 2015-09-02 lfs/attribute, hook: rename Path to Section, nuke HookType type (Taylor Blau[32m[m)
[1;36m|[m * [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 3b9b0f2 2015-09-01 lfs/setup, commands: revert Setup/Teardown to Install/Uninstall (Taylor Blau[32m[m)
[1;36m|[m * [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 0aba239 2015-09-01 lfs/filters+attributes: drop the `Filter` type for `Attribute` (Taylor Blau[32m[m)
[1;36m|[m * [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m a046790 2015-09-01 lfs/setup: use `git-lfs` instead of `git lfs` to ensure tests pass (Taylor Blau[32m[m)
[1;36m|[m * [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m c7cb8e0 2015-09-01 lfs: promote Hooks and Filters to types (Taylor Blau[32m[m)
[1;36m|[m [1;35m|[m[1;35m/[m [31m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
* [1;35m|[m [31m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   9bf4cc8 2015-09-02 Merge pull request #626 from sinbad/creds-username (risk danger olson[32m[m)
[31m|[m[33m\[m [1;35m\[m [31m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[31m|[m [33m|[m[31m_[m[1;35m|[m[31m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[31m|[m[31m/[m[33m|[m [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
[31m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m ecf2dbe 2015-09-02 Tests for disambiguating creds by user (thanks @technoweenie) (Steve Streeting[32m[m)
[31m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m d806269 2015-09-02 Include the username in the creds call if present (Steve Streeting[32m[m)
[31m|[m [1;35m|[m[1;35m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
* [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   3226d2a 2015-09-02 Merge pull request #624 from github/prefer-batch-api (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 06e1f20 2015-09-02 update config docs (Rick Olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 3b4898b 2015-09-02 whitespace (Rick Olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 180ff62 2015-09-02 merge all batch tests into a single test func (Rick Olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 99f1956 2015-09-01 make the ssh auth call with the correct operation (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   ce18a1d 2015-09-01 merge master (risk danger olson[32m[m)
[34m|[m [36m|[m[1;35m\[m [1;35m\[m [1;35m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[34m|[m [36m|[m [1;35m|[m[1;35m/[m [1;35m/[m [35m/[m [35m/[m [33m/[m [1;35m/[m  
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 7440f60 2015-09-01 fix tests (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 920d6a2 2015-09-01 empty strings are not a valid auth type (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 5930c7f 2015-09-01 show the endpoint auth type in 'git lfs env' (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m f22a19f 2015-09-01 don't use '--add' for single value keys (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m e44fe9d 2015-09-01 set lfs.{endpoint}.access to "basic", not "private" (risk danger olson[32m[m)
[34m|[m * [1;35m|[m [1;35m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 8491e60 2015-09-01 default to the batch API if lfs.batch is not set (risk danger olson[32m[m)
[34m|[m [35m|[m [1;35m|[m[35m_[m[1;35m|[m[35m_[m[35m|[m[35m/[m [33m/[m [1;35m/[m  
[34m|[m [35m|[m[35m/[m[1;35m|[m [1;35m|[m [35m|[m [33m|[m [1;35m|[m   
* [35m|[m [1;35m|[m [1;35m|[m [35m|[m [33m|[m [1;35m|[m   72c5222 2015-09-02 Merge pull request #599 from github/batch-api-validation-spec (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [35m\[m [1;35m\[m [1;35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;32m|[m * [35m|[m [1;35m|[m [1;35m|[m [35m|[m [33m|[m [1;35m|[m 45b8048 2015-09-02 optional, not recommended (risk danger olson[32m[m)
[1;32m|[m * [35m|[m [1;35m|[m [1;35m|[m [35m|[m [33m|[m [1;35m|[m 05c4756 2015-09-01 re-arrange the response documentation, mention LFS-Authenticate (risk danger olson[32m[m)
[1;32m|[m * [35m|[m [1;35m|[m [1;35m|[m [35m|[m [33m|[m [1;35m|[m 8b40b39 2015-08-21 describe how validation errors are returned from the batch api (Rick Olson[32m[m)
[1;32m|[m [35m|[m [35m|[m[35m_[m[1;35m|[m[35m_[m[1;35m|[m[35m/[m [33m/[m [1;35m/[m  
[1;32m|[m [35m|[m[35m/[m[35m|[m [1;35m|[m [1;35m|[m [33m|[m [1;35m|[m   
* [35m|[m [35m|[m [1;35m|[m [1;35m|[m [33m|[m [1;35m|[m   9f58bce 2015-09-02 Merge pull request #627 from sinbad/batch-test-issue-1 (risk danger olson[32m[m)
[33m|[m[1;35m\[m [35m\[m [35m\[m [1;35m\[m [1;35m\[m [33m\[m [1;35m\[m  
[33m|[m [1;35m|[m[33m_[m[35m|[m[33m_[m[35m|[m[33m_[m[1;35m|[m[33m_[m[1;35m|[m[33m/[m [1;35m/[m  
[33m|[m[33m/[m[1;35m|[m [35m|[m [35m|[m [1;35m|[m [1;35m|[m [1;35m|[m   
[33m|[m * [35m|[m [35m|[m [1;35m|[m [1;35m|[m [1;35m|[m 972efa6 2015-09-02 Go 1.5 fix: Use sync channel, not just queue.Wait() to ensure xfers done (Steve Streeting[32m[m)
[33m|[m * [35m|[m [35m|[m [1;35m|[m [1;35m|[m [1;35m|[m d4f6358 2015-09-02 Test disabled batch support more locally not globally on the server (Steve Streeting[32m[m)
[33m|[m [1;35m|[m [35m|[m[1;35m_[m[35m|[m[1;35m/[m [1;35m/[m [1;35m/[m  
[33m|[m [1;35m|[m[1;35m/[m[35m|[m [35m|[m [1;35m|[m [1;35m|[m   
* [1;35m|[m [35m|[m [35m|[m [1;35m|[m [1;35m|[m   6307432 2015-09-02 Merge pull request #615 from ttaylorr/refactor-batcher (Scott Barron[32m[m)
[1;35m|[m[31m\[m [1;35m\[m [35m\[m [35m\[m [1;35m\[m [1;35m\[m  
[1;35m|[m [31m|[m[1;35m/[m [35m/[m [35m/[m [1;35m/[m [1;35m/[m  
[1;35m|[m[1;35m/[m[31m|[m [35m|[m [35m|[m [1;35m|[m [1;35m|[m   
[1;35m|[m * [35m|[m [35m|[m [1;35m|[m [1;35m|[m e947838 2015-08-28 lfs/batcher: remove Lot type for []Transferable (Taylor Blau[32m[m)
[1;35m|[m * [35m|[m [35m|[m [1;35m|[m [1;35m|[m 79f939c 2015-08-28 lfs/batcher: remove ambiguity in argument to IsFull func (Taylor Blau[32m[m)
[1;35m|[m * [35m|[m [35m|[m [1;35m|[m [1;35m|[m 37a82cb 2015-08-28 lfs/batcher: refactor `run` goroutine, introduce type `Lot` (Taylor Blau[32m[m)
[1;35m|[m [1;35m|[m [35m|[m[1;35m_[m[35m|[m[1;35m/[m [1;35m/[m  
[1;35m|[m [1;35m|[m[1;35m/[m[35m|[m [35m|[m [1;35m|[m   
* [1;35m|[m [35m|[m [35m|[m [1;35m|[m   d7aa084 2015-09-01 Merge pull request #611 from github/creds-for-lfs-only (risk danger olson[32m[m)
[35m|[m[33m\[m [1;35m\[m [35m\[m [35m\[m [1;35m\[m  
[35m|[m [33m|[m[35m_[m[1;35m|[m[35m_[m[35m|[m[35m/[m [1;35m/[m  
[35m|[m[35m/[m[33m|[m [1;35m|[m [35m|[m [1;35m|[m   
[35m|[m * [1;35m|[m [35m|[m [1;35m|[m 71813f6 2015-09-01 push the error wrapping down to the lower level credential functions (risk danger olson[32m[m)
[35m|[m * [1;35m|[m [35m|[m [1;35m|[m   5deede1 2015-09-01 merge master (risk danger olson[32m[m)
[35m|[m [34m|[m[35m\[m [1;35m\[m [35m\[m [1;35m\[m  
[35m|[m [34m|[m[35m/[m [1;35m/[m [35m/[m [1;35m/[m  
[35m|[m[35m/[m[34m|[m [1;35m|[m [35m|[m [1;35m|[m   
* [34m|[m [1;35m|[m [35m|[m [1;35m|[m   86f9d3d 2015-09-01 Merge pull request #600 from github/errors (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [1;35m\[m [35m\[m [1;35m\[m  
[36m|[m * [34m\[m [1;35m\[m [35m\[m [1;35m\[m   ca2a93a 2015-09-01 Merge branch 'master' into errors (rubyist[32m[m)
[36m|[m [1;32m|[m[36m\[m [34m\[m [1;35m\[m [35m\[m [1;35m\[m  
[36m|[m [1;32m|[m[36m/[m [34m/[m [1;35m/[m [35m/[m [1;35m/[m  
[36m|[m[36m/[m[1;32m|[m [34m|[m [1;35m|[m [35m|[m [1;35m|[m   
* [1;32m|[m [34m|[m [1;35m|[m [35m|[m [1;35m|[m   1902456 2015-09-01 Merge pull request #610 from sinbad/fetch-recent (risk danger olson[32m[m)
[1;35m|[m[1;35m\[m [1;32m\[m [34m\[m [1;35m\[m [35m\[m [1;35m\[m  
[1;35m|[m [1;35m|[m[1;35m_[m[1;32m|[m[1;35m_[m[34m|[m[1;35m/[m [35m/[m [1;35m/[m  
[1;35m|[m[1;35m/[m[1;35m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m   
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 7519896 2015-09-01 Change default fetch recent to include remote refs from current remote (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m fc72090 2015-09-01 Typo (case) (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 0e4faf6 2015-09-01 Move git-lfs-config man to section 5 (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 8cf2be7 2015-08-28 Add an additional test to prove that fetch does snapshot unchanged files (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3c62a2a 2015-08-27 Add tests for fetch --recent with remote branches (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 88b362c 2015-08-27 Tests for fetch --recent (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 58decd0 2015-08-27 Tweak fetch recent output a little (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m 5a9aaa5 2015-08-27 Config was skipping fetch settings when 0, must read & set 0's (Steve Streeting[32m[m)
[1;35m|[m * [1;32m|[m [34m|[m [35m|[m [1;35m|[m   66e92b0 2015-08-27 Merge branch 'master' into fetch-recent (Steve Streeting[32m[m)
[1;35m|[m [1;36m|[m[31m\[m [1;32m\[m [34m\[m [35m\[m [1;35m\[m  
[1;35m|[m * [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 6f97398 2015-08-26 Don't allocate unless required; also data consistency with PointerClean (Steve Streeting[32m[m)
[1;35m|[m * [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m   6a2ce62 2015-08-26 Merge branch 'push-other-branch' into fetch-recent (Steve Streeting[32m[m)
[1;35m|[m [32m|[m[33m\[m [31m\[m [1;32m\[m [34m\[m [35m\[m [1;35m\[m  
[1;35m|[m [32m|[m * [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 5d92a5d 2015-08-26 Fix failure to push non-current branch #606 (Steve Streeting[32m[m)
[1;35m|[m [32m|[m * [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m be37e93 2015-08-26 Add oids to push --dry-run output otherwise can't see modifications (Steve Streeting[32m[m)
[1;35m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 111e674 2015-08-26 First fetch --recent integration test (Steve Streeting[32m[m)
[1;35m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 2fa370e 2015-08-26 Slightly more tracing in fetch (Steve Streeting[32m[m)
[1;35m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m b7323b3 2015-08-26 Alter call to ScanRefs in fetch to use --no-walk (snapshot only) (Steve Streeting[32m[m)
[1;35m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m bcc14cf 2015-08-26 Push now works properly for other branches (Steve Streeting[32m[m)
[1;35m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m   aeca33c 2015-08-26 Merge branch 'push-other-branch' into fetch-recent (Steve Streeting[32m[m)
[1;35m|[m [34m|[m[35m\[m [33m\[m [31m\[m [1;32m\[m [34m\[m [35m\[m [1;35m\[m  
[1;35m|[m [34m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m c65ed36 2015-08-26 Fix failure to push non-current branch #606 (Steve Streeting[32m[m)
[1;35m|[m [34m|[m * [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 31d887d 2015-08-26 Add oids to push --dry-run output otherwise can't see modifications (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 049bb52 2015-08-25 Fix go tests for commit structure changes (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m aba8f6d 2015-08-25 Real fetch test - currently failing because of https://github.com/github/git-lfs/issues/606 (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 1b52eb3 2015-08-25 Terminate return json output with newline (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 0968ac0 2015-08-25 Encode incoming son data as string for ease of use (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3f18feb 2015-08-25 Make testutils actually write objects too (PointerClean only creates temp) (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m fd22663 2015-08-25 Allow file data to be provided as reader or []byte (useful for json) (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 034e382 2015-08-25 Return UTC time not local time as TZ format is different for Go vs date (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 21d5b3a 2015-08-25 Added a date offset helper which works with BSD(Mac) & GNU date tools (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3d2ed18 2015-08-25 More representative example of feeding JSON into lfstest-testutils (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 2fa44e0 2015-08-24 Create lfstest-testutils tool to create more complex test cases easily in shell (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 0ebd8ba 2015-08-24 Don't use testing.T directly, create compatible subset interface & use that (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 21596a0 2015-08-24 Test for ScanPreviousVersions (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 9a31822 2015-08-24 Initialise pointer extension to empty array in clean to be consistent with pointer decode (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 178e568 2015-08-24 Add full test for ScanUnpushed now scaffolding is there (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m abf0f93 2015-08-21 Sort refs by name to be able to assert.Equal() consistently (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m cdafe05 2015-08-21 Refine errors (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m a60a958 2015-08-21 Fix Popd() (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 8138379 2015-08-21 Make sure username/email are always supplied to commit so work in any env (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 0740f59 2015-08-21 More error checking on commitAtDate (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m d134b94 2015-08-21 More info from git show failure (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m dadeb34 2015-08-21 First pass of fetching recent commits (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 5de6e8b 2015-08-21 Test the git package in script/test (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 75a9115 2015-08-21 Fix error (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3967619 2015-08-20 Make ref resolution return a complete Ref rather than just SHA, & test (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 75cdf02 2015-08-20 Use execCommand not exec.Command (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 03be359 2015-08-20 Remove some unnecessary verbosity (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m c5032d3 2015-08-20 Test RecentBranches with remote branches (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 4089286 2015-08-20 Allow test repos to have remotes which are automatically cleaned (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 7e6e297 2015-08-20 Store *testing.T in test.Repo to reduce number of args needed to funcs (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3dd6017 2015-08-19 Simplify the test types a bit, already in the test package (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3510af3 2015-08-19 First test of RecentBranches using new test utils (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 1538c89 2015-08-19 Change RecentRefs to RecentBranches & optimise so we can get date in 1 cmd (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 11ab2c1 2015-08-19 Add Pushd/Popd functions to TestRepo for convenience (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 8202a78 2015-08-19 Support creating tags on test commits (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 241152d 2015-08-19 Fix placeholder data reader, need to return io.EOF at end of stream (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 895bad2 2015-08-18 When cleaning up, make sure cwd isn't inside folder to be deleted (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m d8cad4e 2015-08-18 Attach setup func to TestRepo instance, be safe with cwd (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 02c5c64 2015-08-18 Working on some "go test" utilities to create non-trivial repos for granular tests (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m e7e150c 2015-08-18 Add examples of fetch include/exclude (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 5e0f81d 2015-08-18 Adding utility methods to get recent refs (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m fa34aa5 2015-08-18 Fix test with changed config name (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m b759b89 2015-08-18 Create a dedicated man page containing all config options (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 8758276 2015-08-18 Add "--recent" fetch option & update docs (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 3c795c0 2015-08-18 Change the default for recent commits to 0 days (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 7422b4b 2015-08-18 Rename lfs.fetchrecentrefsincluderemotes to lfs.fetchrecentremoterefs (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 28fd6c1 2015-08-18 Add lfs.fetchrecentalways config option (Steve Streeting[32m[m)
[1;35m|[m * [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [35m|[m [1;35m|[m 1688043 2015-08-18 Adding config options for fetch/prune (Steve Streeting[32m[m)
[1;35m|[m [35m|[m [35m|[m[35m_[m[33m|[m[35m_[m[31m|[m[35m_[m[1;32m|[m[35m_[m[34m|[m[35m/[m [1;35m/[m  
[1;35m|[m [35m|[m[35m/[m[35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [1;35m|[m   
* [35m|[m [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [1;35m|[m 3f1eec2 2015-08-27 typo (Rick Olson[32m[m)
* [35m|[m [35m|[m [33m|[m [31m|[m [1;32m|[m [34m|[m [1;35m|[m   5d4b5cf 2015-08-27 Merge pull request #613 from ttaylorr/batcher-test-cases (risk danger olson[32m[m)
[31m|[m[1;31m\[m [35m\[m [35m\[m [33m\[m [31m\[m [1;32m\[m [34m\[m [1;35m\[m  
[31m|[m [1;31m|[m[31m_[m[35m|[m[31m_[m[35m|[m[31m_[m[33m|[m[31m/[m [1;32m/[m [34m/[m [1;35m/[m  
[31m|[m[31m/[m[1;31m|[m [35m|[m [35m|[m [33m|[m [1;32m|[m [34m|[m [1;35m|[m   
[31m|[m * [35m|[m [35m|[m [33m|[m [1;32m|[m [34m|[m [1;35m|[m a213908 2015-08-27 lfs/batcher: rename assertAll to runBatcherTests (Taylor Blau[32m[m)
[31m|[m * [35m|[m [35m|[m [33m|[m [1;32m|[m [34m|[m [1;35m|[m d796539 2015-08-27 lfs/batcher: drop prefixes from Godoc comments in test (Taylor Blau[32m[m)
[31m|[m * [35m|[m [35m|[m [33m|[m [1;32m|[m [34m|[m [1;35m|[m 4a41bf1 2015-08-27 lfs/batcher: introduce test cases (Taylor Blau[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 2d93c0c 2015-08-31 Update a doc string (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m c57a899 2015-08-28 Error context can hold any kind of data, simplify the cleanPointerError (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 433ccdd 2015-08-28 Errorf works properly now, remove these TODOs (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 4517851 2015-08-27 Fix up Errorf(), provide some default messages for lfs error types (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 3a1f8f5 2015-08-27 Export CleanPointerError for now, to satisfy linting (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 80beead 2015-08-27 Bring DownloadDeclinedError into errors system. (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m dbeb15e 2015-08-27 Prevent error data loss, give not a pointer error a default message (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 136a68f 2015-08-27 NotInARepositoryError => IsInvalidRepoError() (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 290cc83 2015-08-27 That should be a Fatalf (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m   e434980 2015-08-27 Merge branch 'master' into errors (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m [1;32m|[m[31m\[m [34m\[m [1;35m\[m  
[31m|[m [1;31m|[m[31m_[m[35m|[m[31m_[m[35m|[m[31m_[m[33m|[m[31m_[m[1;32m|[m[31m/[m [34m/[m [1;35m/[m  
[31m|[m[31m/[m[1;31m|[m [35m|[m [35m|[m [33m|[m [1;32m|[m [34m|[m [1;35m|[m   
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 492e5f6 2015-08-22 Refactor error type out (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 1f1821c 2015-08-22 replace NotAPointerError (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 554aef8 2015-08-21 tests, fix a bug, docs (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 187c6cb 2015-08-21 wErr -> err (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 0c02eda 2015-08-21 add some docs (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m cd754db 2015-08-21 First pass at getting rid of WrappedError (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [34m|[m [1;35m|[m 354cce9 2015-08-21 base error behaviors (rubyist[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m[35m_[m[33m|[m[35m/[m [34m/[m [1;35m/[m  
[31m|[m [1;31m|[m [35m|[m[35m/[m[35m|[m [33m|[m [34m|[m [1;35m|[m   
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m   6655726 2015-08-31 Merge pull request #618 from github/more-cred-tests (risk danger olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m [1;34m|[m[1;35m\[m [1;35m\[m  
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m [1;34m|[m * [1;35m|[m c3996ef 2015-08-28 setup a reusable test matrix (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m [1;34m|[m * [1;35m|[m 278d4f2 2015-08-28 add Config.Reset() (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m [1;34m|[m[1;34m/[m [1;35m/[m  
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 446ffb2 2015-08-28 run "verify" API requests with credentials too (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 5b77244 2015-08-28 backing out of this test (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m a3a0307 2015-08-28 test the output if git-credential doesnt fail (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m e406d70 2015-08-28 more info (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 5801efa 2015-08-28 check git-credential on storage api requests (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 79a82a5 2015-08-28 bake credential handling into doHttpRequest() and handleResponse() (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 9829d26 2015-08-28 simplify creds (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m d42ad12 2015-08-28 rename credentials() => fillCredentials() (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m f71969d 2015-08-28 ask git-credential for API creds even if the host is different (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 74b6e2d 2015-08-28 re-arrange some methods (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m 55b1fdf 2015-08-28 rename getCreds() to getCredsForAPI() to describe it's LFS-API-specific behavior (Rick Olson[32m[m)
[31m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m * [1;35m|[m ed376b4 2015-08-27 only check git-credentials for lfs api requests (Rick Olson[32m[m)
[31m|[m [1;31m|[m[31m_[m[35m|[m[31m_[m[35m|[m[31m_[m[33m|[m[31m/[m [1;35m/[m  
[31m|[m[31m/[m[1;31m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   
* [1;31m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m   965ddbb 2015-08-26 Merge pull request #608 from sinbad/push-other-branch (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [35m\[m [35m\[m [33m\[m [1;35m\[m  
[1;36m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 1cacc95 2015-08-26 Fix failure to push non-current branch #606 (Steve Streeting[32m[m)
[1;36m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m [1;35m|[m 41cdd01 2015-08-26 Add oids to push --dry-run output otherwise can't see modifications (Steve Streeting[32m[m)
[1;36m|[m [1;35m|[m [1;31m|[m[1;35m_[m[35m|[m[1;35m_[m[35m|[m[1;35m_[m[33m|[m[1;35m/[m  
[1;36m|[m [1;35m|[m[1;35m/[m[1;31m|[m [35m|[m [35m|[m [33m|[m   
* [1;35m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m   385b811 2015-08-26 Merge pull request #603 from andyneff/docker_windows (risk danger olson[32m[m)
[1;35m|[m[33m\[m [1;35m\[m [1;31m\[m [35m\[m [35m\[m [33m\[m  
[1;35m|[m [33m|[m[1;35m/[m [1;31m/[m [35m/[m [35m/[m [33m/[m  
[1;35m|[m[1;35m/[m[33m|[m [1;31m|[m [35m|[m [35m|[m [33m|[m   
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m 5381439 2015-08-23 Fixed newlines in extensions.md (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m b1bf586 2015-08-22 Fix permissions problem for Debian docker from Windows (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m f382189 2015-08-22 Enable autocrlf (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m 8f78b05 2015-08-22 Fixed Centos 6 failing due to new epel golang package (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m a842afd 2015-08-22 Patch for windows (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m 5ea3f9b 2015-08-22 Fix to handle #563 (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m 21b0df7 2015-08-22 Fixed small bug for when epel is updating thier servers (Andy Neff[32m[m)
[1;35m|[m * [1;31m|[m [35m|[m [35m|[m [33m|[m 5f033e5 2015-08-22 Added DOCKER_OTHER_OPTIONS and started patching for windows (Andy Neff[32m[m)
[1;35m|[m [35m|[m [1;31m|[m[35m/[m [35m/[m [33m/[m  
[1;35m|[m [35m|[m[35m/[m[1;31m|[m [35m|[m [33m|[m   
* [35m|[m [1;31m|[m [35m|[m [33m|[m   e8a2161 2015-08-26 Merge pull request #607 from sinbad/test-enhancements (risk danger olson[32m[m)
[1;31m|[m[33m\[m [35m\[m [1;31m\[m [35m\[m [33m\[m  
[1;31m|[m [33m|[m[1;31m_[m[35m|[m[1;31m/[m [35m/[m [33m/[m  
[1;31m|[m[1;31m/[m[33m|[m [35m|[m [35m|[m [33m/[m   
[1;31m|[m [33m|[m [35m|[m[33m_[m[35m|[m[33m/[m    
[1;31m|[m [33m|[m[33m/[m[35m|[m [35m|[m     
[1;31m|[m * [35m|[m [35m|[m 847d005 2015-08-26 Undo extensions change, not required by this PR (Steve Streeting[32m[m)
[1;31m|[m [35m|[m [35m|[m[35m/[m  
[1;31m|[m [35m|[m[35m/[m[35m|[m   
[1;31m|[m * [35m|[m 766c22d 2015-08-25 Add enhanced test setup tools for 'go test' and integration tests (Steve Streeting[32m[m)
[1;31m|[m[1;31m/[m [35m/[m  
* [35m|[m   592f5c5 2015-08-25 Merge pull request #605 from nbrew/patch-1 (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m|[m 5cb8e7e 2015-08-24 Format CentOS bullet points in linux-build.md (Nathan Hyde[32m[m)
[36m|[m[36m/[m [35m/[m  
* [35m|[m   0cca12b 2015-08-24 Merge pull request #594 from github/batch-hang (Scott Barron[32m[m)
[35m|[m[1;33m\[m [35m\[m  
[35m|[m [1;33m|[m[35m/[m  
[35m|[m[35m/[m[1;33m|[m   
[35m|[m * d9b8178 2015-08-18 Tidy up the error handling (rubyist[32m[m)
[35m|[m * fd417b2 2015-08-17 Write a test that triggers and tests for bad batch status codes (rubyist[32m[m)
[35m|[m * e5d3fce 2015-08-17 Display errors when fetching (rubyist[32m[m)
[35m|[m * 84a34c2 2015-08-17 Decrease the waitgroup when batch errors (rubyist[32m[m)
[35m|[m * 2005013 2015-08-17 Return an actual error instead of nil (rubyist[32m[m)
[35m|[m[35m/[m  
*   4457d7c 2015-08-17 Merge pull request #578 from billygor/major-pre-push-optimization-for-current-remote (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * f3bf4ea 2015-08-17 major pre-push optimization (Billy Gor[32m[m)
* [1;35m|[m   78d208c 2015-08-17 Merge pull request #583 from sinbad/fetch-pull-remote-args (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m  
[1;36m|[m * [1;35m|[m b5df63c 2015-08-13 Add optional remote arg to fetch & pull, and default to tracking remote (Steve Streeting[32m[m)
* [31m|[m [1;35m|[m   0f63d9b 2015-08-17 Merge pull request #573 from sinbad/fetch-include-exclude (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m  
[32m|[m * [31m\[m [1;35m\[m   dc71cb1 2015-08-14 Incorporate panic prevention from PR #570 to avoid merge conflicts (Steve Streeting[32m[m)
[32m|[m [31m|[m[35m\[m [31m\[m [1;35m\[m  
[32m|[m [31m|[m [35m|[m[31m/[m [1;35m/[m  
[32m|[m [31m|[m[31m/[m[35m|[m [1;35m|[m   
[32m|[m * [35m|[m [1;35m|[m bf8bade 2015-08-10 Don't download on smudge if file is filtered by include/exclude config (Steve Streeting[32m[m)
[32m|[m * [35m|[m [1;35m|[m 6f847ea 2015-08-10 Do not call update-index until we know we have 1 file to give it as an arg (Steve Streeting[32m[m)
[32m|[m * [35m|[m [1;35m|[m 451b03e 2015-08-10 Adding --include and --exclude args to fetch/pull (Steve Streeting[32m[m)
[32m|[m * [35m|[m [1;35m|[m 4675404 2015-08-10 Add lfs.fetchinclude and lfs.fetchexclude gitconfig options (Steve Streeting[32m[m)
* [35m|[m [35m|[m [1;35m|[m   57eec4f 2015-08-17 Merge pull request #587 from github/bump-cobra (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [35m\[m [35m\[m [1;35m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[35m|[m[1;35m_[m[35m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [35m|[m [35m|[m   
[1;35m|[m * [35m|[m [35m|[m 2be3811 2015-08-14 my typo fix got merged pretty quickly (Rick Olson[32m[m)
[1;35m|[m * [35m|[m [35m|[m ffb83ae 2015-08-14 update the cobra/pflag packages (Rick Olson[32m[m)
[1;35m|[m[1;35m/[m [35m/[m [35m/[m  
* [35m|[m [35m|[m   5214daa 2015-08-14 Merge pull request #585 from github/sinbad-pre-push-exit-code (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [35m\[m [35m\[m  
[1;32m|[m * [35m|[m [35m|[m f6d9680 2015-08-14 update pre-push to be more explicit about the path (Rick Olson[32m[m)
[1;32m|[m * [35m|[m [35m|[m 5b7990c 2015-08-13 teach the init command how to update pre-push hooks with force (Rick Olson[32m[m)
[1;32m|[m * [35m|[m [35m|[m 87b557d 2015-08-13 teach the upgrade command how to update pre-push hooks (Rick Olson[32m[m)
[1;32m|[m * [35m|[m [35m|[m   bfceb3e 2015-08-13 Merge branch 'pre-push-exit-code' of https://github.com/sinbad/git-lfs into sinbad-pre-push-exit-code (Rick Olson[32m[m)
[1;32m|[m [1;34m|[m[1;35m\[m [35m\[m [35m\[m  
[1;32m|[m [1;34m|[m * [35m|[m [35m|[m 3920ef5 2015-08-13 Pre-push hook must exit with a non-zero result when git-lfs isn't on the path (Steve Streeting[32m[m)
* [1;34m|[m [1;35m|[m [35m|[m [35m|[m   179224c 2015-08-14 Merge pull request #581 from sinbad/pre-push-missing-local-present-server (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m [1;35m\[m [35m\[m [35m\[m  
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m 524d450 2015-08-13 Eliminate duplicate code in test server for batch upload/download (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m a7d0a36 2015-08-13 Re-use code between DownloadCheckable and Downloadable (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m a94e2a9 2015-08-12 Change the approach, listen on q.Watch() instead since errors not reliable in batch mode (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m f2c2b60 2015-08-12 Make test server work in batch for download and not just upload (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m 9f54872 2015-08-12 Rename verify_queue to check_queue to eliminate ambiguity (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m d8a7fb0 2015-08-12 Make error reporting consistent with previous so tests pass (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m 92ef7a1 2015-08-11 Working on using a VerifyQueue to check oid server presence without downloading (Steve Streeting[32m[m)
[1;36m|[m * [1;34m|[m [1;35m|[m [35m|[m [35m|[m 88cd8f3 2015-08-11 Make failing test for the condition we want to fix (Steve Streeting[32m[m)
[1;36m|[m [1;35m|[m [1;34m|[m[1;35m/[m [35m/[m [35m/[m  
[1;36m|[m [1;35m|[m[1;35m/[m[1;34m|[m [35m|[m [35m|[m   
* [1;35m|[m [1;34m|[m [35m|[m [35m|[m   22dca3b 2015-08-14 Merge pull request #570 from github/index-progress (Scott Barron[32m[m)
[1;34m|[m[35m\[m [1;35m\[m [1;34m\[m [35m\[m [35m\[m  
[1;34m|[m [35m|[m[1;34m_[m[1;35m|[m[1;34m/[m [35m/[m [35m/[m  
[1;34m|[m[1;34m/[m[35m|[m [1;35m|[m [35m|[m [35m/[m   
[1;34m|[m [35m|[m [1;35m|[m[35m_[m[35m|[m[35m/[m    
[1;34m|[m [35m|[m[35m/[m[1;35m|[m [35m|[m     
[1;34m|[m * [1;35m|[m [35m|[m e909264 2015-08-07 The loop that feeds git update-index MUST NOT panic (rubyist[32m[m)
[1;34m|[m * [1;35m|[m [35m|[m 711f483 2015-08-07 Don't use make here (rubyist[32m[m)
[1;34m|[m [35m|[m [1;35m|[m[35m/[m  
[1;34m|[m [35m|[m[35m/[m[1;35m|[m   
* [35m|[m [1;35m|[m f241c2c 2015-08-13 hyperlinks! (risk danger olson[32m[m)
* [35m|[m [1;35m|[m   9a99504 2015-08-11 Merge pull request #579 from github/cred-tests-without-ldflags (risk danger olson[32m[m)
[1;35m|[m[35m\[m [35m\[m [1;35m\[m  
[1;35m|[m [35m|[m[1;35m_[m[35m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[35m|[m [35m|[m   
[1;35m|[m * [35m|[m 63f388f 2015-08-11 don't use ldflags in credential tests (risk danger olson[32m[m)
[1;35m|[m[1;35m/[m [35m/[m  
* [35m|[m   3f2a526 2015-08-11 Merge pull request #577 from sinbad/queue-error-race-condition (Scott Barron[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m|[m edaecc2 2015-08-11 Fix race condition meaning TransferQueue might not report errors in time (Steve Streeting[32m[m)
[36m|[m[36m/[m [35m/[m  
* [35m|[m   a1d9be9 2015-08-10 Merge pull request #572 from billygor/fix-non-batch-transfer (Scott Barron[32m[m)
[35m|[m[1;33m\[m [35m\[m  
[35m|[m [1;33m|[m[35m/[m  
[35m|[m[35m/[m[1;33m|[m   
[35m|[m * 433150c 2015-08-10 fix non-batch transfer (Billy Gor[32m[m)
[35m|[m[35m/[m  
*   93cdb52 2015-08-07 Merge pull request #553 from github/newbatchapi (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * b8a082f 2015-08-07 Tracer all api errors for the batch call (rubyist[32m[m)
[1;34m|[m * 8eaa760 2015-08-07 clean that up a little (rubyist[32m[m)
[1;34m|[m * 714e848 2015-08-07 handle individual object errors (rubyist[32m[m)
[1;34m|[m * aecb40d 2015-07-31 Still need to support _links for the time being (rubyist[32m[m)
[1;34m|[m * 34cec85 2015-07-31 Start with _links => actions (rubyist[32m[m)
* [1;35m|[m   24633a7 2015-08-07 Merge pull request #569 from sinbad/fetch-pull-checkout-docs (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m  
[1;36m|[m * [1;35m|[m 6210876 2015-08-07 Added missing top-level man entries for fetch, pull & checkout (Steve Streeting[32m[m)
[1;36m|[m[1;36m/[m [1;35m/[m  
* [1;35m|[m   53e0579 2015-08-07 Merge pull request #566 from sinbad/checkout-dont-download (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m e0e56ca 2015-08-06 Never download anything when using the checkout command (Steve Streeting[32m[m)
* [33m|[m [1;35m|[m   69f8241 2015-08-07 Merge pull request #559 from sinbad/scan-log-diff (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;35m\[m  
[34m|[m * [33m|[m [1;35m|[m d543277 2015-08-07 Add tests including pointer extension data for parseLogOutputToPointers (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 30cbc8e 2015-08-05 Increase context to 12 to cope with extreme edge cases using 10 extensions (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 051db88 2015-08-05 Pass extensions to pointer decode too now they're supported (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m b5d098a 2015-08-05 Make sure we have enough diff context in pointer text to cope with extensions (Steve Streeting[32m[m)
[34m|[m * [33m|[m [1;35m|[m 90ce638 2015-08-05 Adding parsing methods to scan git log for LFS references which have been introduced or removed (Steve Streeting[32m[m)
[34m|[m [33m|[m[33m/[m [1;35m/[m  
* [33m|[m [1;35m|[m   72e8dd6 2015-08-06 Merge pull request #555 from andyneff/docker-improvements (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;35m\[m  
[36m|[m * [33m|[m [1;35m|[m fa35014 2015-08-05 Minor bug and documentation (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 540b270 2015-08-05 Update README.md (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m f32a0ac 2015-08-04 Made more Mac/BSD Friendly (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 53905b5 2015-07-31 Fix permissions when using sudo (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 226224a 2015-07-31 Readme update (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 9b27b4f 2015-07-31 Removed signing.key shortcut. too complicated (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m e41cc41 2015-07-31 GPG working correctly now (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 56ffe42 2015-07-31 Stuff (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 43624a3 2015-07-31 rpm_build.bsh gets version from git-lfs version instead of parsing code (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 8e5636f 2015-07-31 In the middle of fixing gpg keys (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m ec50fd2 2015-07-31 Cleaned up multiple keys some (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m d2e8da1 2015-07-31 Broke something, private key no longer loads in container (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m e41c14d 2015-07-31 Added GPG docker instead of preload_key.bsh (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 5becfbc 2015-07-31 Fixed bug so 32 and 64 bit debs can coexist and coupdate (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m b5b8eda 2015-07-31 Fixed repo dir paths (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 0cd1c6c 2015-07-31 32 bit deb's cross compiling correctly (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m c0962f4 2015-07-31 32 bit debian build WORKING, but a little complicated (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 26c6e2e 2015-07-31 Converted to docker+ dockerfiles (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 3207b46 2015-07-31 Removed directories that were pointless in this pattern (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 796ec71 2015-07-31 Fixed SUDO test for boot2docker (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;35m|[m 5ef85e4 2015-07-31 Cleaning up/improving docker performance (Andy Neff[32m[m)
* [1;31m|[m [33m|[m [1;35m|[m   af42c0c 2015-08-06 Merge pull request #565 from sinbad/fetch-test-local-storage (risk danger olson[32m[m)
[33m|[m[1;33m\[m [1;31m\[m [33m\[m [1;35m\[m  
[33m|[m [1;33m|[m[33m_[m[1;31m|[m[33m/[m [1;35m/[m  
[33m|[m[33m/[m[1;33m|[m [1;31m|[m [1;35m|[m   
[33m|[m * [1;31m|[m [1;35m|[m 5e654f2 2015-08-06 Fetch tests should be checking that content exists, not pointer exists (Steve Streeting[32m[m)
[33m|[m[33m/[m [1;31m/[m [1;35m/[m  
* [1;31m|[m [1;35m|[m   5df23bf 2015-08-05 Merge pull request #563 from hoolio/patch-1 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;35m\[m  
[1;34m|[m * [1;31m|[m [1;35m|[m 285389d 2015-08-04 Rename LICENSE to LICENSE.md (Julio Avalos[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m bb8702b 2015-08-04 Update LFS license file to reflect MIT license use (Julio Avalos[32m[m)
[1;34m|[m[1;34m/[m [1;31m/[m [1;35m/[m  
* [1;31m|[m [1;35m|[m   df4be34 2015-08-04 Merge pull request #560 from github/sort-extension-refactor (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [1;35m\[m  
[1;36m|[m * [1;31m|[m [1;35m|[m 36946ec 2015-08-04 refactor lfs.SortExtensions (Rick Olson[32m[m)
* [31m|[m [1;31m|[m [1;35m|[m   1a18ff6 2015-08-04 Merge pull request #551 from sinbad/checkout-progress (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;31m\[m [1;35m\[m  
[32m|[m * [31m|[m [1;31m|[m [1;35m|[m 9818b8b 2015-08-03 Call progress.Start() (Steve Streeting[32m[m)
[32m|[m * [31m|[m [1;31m|[m [1;35m|[m 0cb1dcf 2015-08-03 Add progress meter to checkout (Steve Streeting[32m[m)
[32m|[m [1;31m|[m [31m|[m[1;31m/[m [1;35m/[m  
[32m|[m [1;31m|[m[1;31m/[m[31m|[m [1;35m|[m   
* [1;31m|[m [31m|[m [1;35m|[m   243b734 2015-08-04 Merge pull request #561 from github/credential-tests (risk danger olson[32m[m)
[31m|[m[35m\[m [1;31m\[m [31m\[m [1;35m\[m  
[31m|[m [35m|[m[31m_[m[1;31m|[m[31m/[m [1;35m/[m  
[31m|[m[31m/[m[35m|[m [1;31m|[m [1;35m|[m   
[31m|[m * [1;31m|[m [1;35m|[m   557bdbb 2015-08-04 Merge pull request #562 from github/usehttppath-tests (risk danger olson[32m[m)
[31m|[m [36m|[m[1;31m\[m [1;31m\[m [1;35m\[m  
[31m|[m [36m|[m * [1;31m|[m [1;35m|[m 22d7c9f 2015-08-04 dry up code for setting auth on a req based on url creds (Rick Olson[32m[m)
[31m|[m [36m|[m * [1;31m|[m [1;35m|[m 74cafba 2015-08-04 return any errors from parsing the remote url (Rick Olson[32m[m)
[31m|[m [36m|[m * [1;31m|[m [1;35m|[m   7ee5f66 2015-08-04 Merge branch '554-master' into usehttppath-tests (Rick Olson[32m[m)
[31m|[m [36m|[m [1;32m|[m[1;33m\[m [1;31m\[m [1;35m\[m  
[31m|[m [36m|[m [1;32m|[m * [1;31m|[m [1;35m|[m 5bcb160 2015-08-02 Use the git remote's credentials if the scheme and host match, so that users don't have to cache the same credentials twice (Clare Liguori[32m[m)
[31m|[m [36m|[m [1;32m|[m * [1;31m|[m [1;35m|[m 5dc0697 2015-07-31 Trim leading slash from the path passed to credential fill (Clare Liguori[32m[m)
[31m|[m [36m|[m [1;32m|[m * [1;31m|[m [1;35m|[m 1f08ac9 2015-07-31 Support credentials from multiple accounts when useHttpPath=true is set in the git config (Clare Liguori[32m[m)
[31m|[m [36m|[m [1;32m|[m [1;31m|[m[1;31m/[m [1;35m/[m  
[31m|[m [36m|[m * [1;31m|[m [1;35m|[m 69aee53 2015-08-04 add more credential tests (Rick Olson[32m[m)
[31m|[m [36m|[m[36m/[m [1;31m/[m [1;35m/[m  
[31m|[m * [1;31m|[m [1;35m|[m abb94ba 2015-08-04 fewer loc (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;35m|[m 2929402 2015-08-04 add rudimentary auth check (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;35m|[m b2e7ba2 2015-08-04 teach the git credential helper how to read passwords from disk (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;35m|[m 087542a 2015-08-04 write to stderr so its clear when the credential helper is running (Rick Olson[32m[m)
[31m|[m[31m/[m [1;31m/[m [1;35m/[m  
* [1;31m|[m [1;35m|[m   e01eff6 2015-08-04 Merge pull request #486 from ryansimmen/extensions (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;35m\[m  
[1;34m|[m * [1;31m|[m [1;35m|[m 6a3a2c3 2015-08-01 LFS Extensions: use lfs.extension instead of lfs-ext (Ryan Simmen[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m 8d7769a 2015-07-31 LFS Extensions: smudge command (Ryan Simmen[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m d010bf9 2015-07-31 LFS Extensions: clean command (Ryan Simmen[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m 214983d 2015-07-31 LFS Extensions: pointer file manipulation (Ryan Simmen[32m[m)
[1;34m|[m * [1;31m|[m [1;35m|[m ffd52e6 2015-07-31 LFS Extensions: git-lfs ext command (Ryan Simmen[32m[m)
[1;34m|[m [1;31m|[m[1;31m/[m [1;35m/[m  
* [1;31m|[m [1;35m|[m 69345c3 2015-08-04 add the gitter chat room to the readme (Rick Olson[32m[m)
* [1;31m|[m [1;35m|[m   23a78d9 2015-08-03 Merge pull request #552 from github/git-version-tests (risk danger olson[32m[m)
[1;31m|[m[31m\[m [1;31m\[m [1;35m\[m  
[1;31m|[m [31m|[m[1;31m/[m [1;35m/[m  
[1;31m|[m[1;31m/[m[31m|[m [1;35m|[m   
[1;31m|[m * [1;35m|[m 332a9db 2015-07-31 build a helper for confirming the git version in tests (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m [1;35m/[m  
* [1;35m|[m   f021ede 2015-07-31 Merge pull request #546 from sinbad/worktree-support (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m 72b3fb5 2015-07-31 Doh, forgot the exit (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m 4f93045 2015-07-31 Simplify skipping of worktree tests when git < 2.5.0 (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m d1728f8 2015-07-31 Comment revisions per review (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m 7161afe 2015-07-31 Add git worktree tests when git version >= 2.5.0 (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m cab91e4 2015-07-31 Fix tests (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m 66ebac2 2015-07-31 Fixed support for worktree (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m 4169507 2015-07-31 Initial stab at worktree support (Steve Streeting[32m[m)
[32m|[m [1;35m|[m[1;35m/[m  
* [1;35m|[m   3abb3ea 2015-07-31 Merge pull request #547 from github/delay-progress (Scott Barron[32m[m)
[34m|[m[35m\[m [1;35m\[m  
[34m|[m * [1;35m|[m 87fbb85 2015-07-31 Delay progress output until the transfer starts (rubyist[32m[m)
[34m|[m [1;35m|[m[1;35m/[m  
* [1;35m|[m   fd67366 2015-07-31 Merge pull request #549 from github/allow-no-commit (Scott Barron[32m[m)
[1;35m|[m[1;31m\[m [1;35m\[m  
[1;35m|[m [1;31m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m   
[1;35m|[m * 483d38b 2015-07-31 err can just be ignored completely (rubyist[32m[m)
[1;35m|[m * 3d954e9 2015-07-31 Don't panic if we can't find the git commit for the version (rubyist[32m[m)
[1;35m|[m[1;35m/[m  
*   cc0bf41 2015-07-30 Merge pull request #543 from sinbad/checkout-dotpath (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * 9ec60d8 2015-07-30 Move local dir set out for pkg-level construction instead of method-level (Steve Streeting[32m[m)
[1;32m|[m * b123acd 2015-07-30 Use a set to detect `.` `./` `.\` instead of 3 conditions (Steve Streeting[32m[m)
[1;32m|[m * 6deebd3 2015-07-30 Support `git lfs checkout .` and similar correctly (Steve Streeting[32m[m)
* [1;33m|[m   13323eb 2015-07-30 Merge pull request #544 from github/suppress-dry-run-progress (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m  
[1;34m|[m * [1;33m|[m 46249c9 2015-07-30 While I'm in here, this value is always the same, push it down (rubyist[32m[m)
[1;34m|[m * [1;33m|[m b87386d 2015-07-30 Bake in dry-run-ability and we'll never forget to suppress output (rubyist[32m[m)
[1;34m|[m * [1;33m|[m e464e65 2015-07-30 Suppression should happen immediately (rubyist[32m[m)
[1;34m|[m * [1;33m|[m 3a33a5f 2015-07-30 Suppress progress output for a push dry run (rubyist[32m[m)
[1;34m|[m [1;33m|[m[1;33m/[m  
* [1;33m|[m   bcf15d9 2015-07-30 Merge pull request #540 from github/fix-test-race-condition (risk danger olson[32m[m)
[1;33m|[m[31m\[m [1;33m\[m  
[1;33m|[m [31m|[m[1;33m/[m  
[1;33m|[m[1;33m/[m[31m|[m   
[1;33m|[m * bc8a085 2015-07-29 use the standard bash shebang in script/debian-build (Rick Olson[32m[m)
[1;33m|[m * dbf09f7 2015-07-29 update install.sh to use the standard shebang (Rick Olson[32m[m)
[1;33m|[m * 202ae12 2015-07-29 update the shebang lines to use bash (Rick Olson[32m[m)
[1;33m|[m * 42ae63f 2015-07-29 remove test counting code (Rick Olson[32m[m)
[1;33m|[m * 791d100 2015-07-29 compare ran tests vs how many begin_test calls are in the test file (Rick Olson[32m[m)
[1;33m|[m * 18759e5 2015-07-29 no need to create a temp home since each test has its own (Rick Olson[32m[m)
[1;33m|[m * 1efb30a 2015-07-29 rebuild HOME for each test (Rick Olson[32m[m)
* [31m|[m   5410d18 2015-07-30 Merge pull request #542 from sinbad/fetch-multiple-refs (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m  
[32m|[m * [31m|[m dbf4a8b 2015-07-30 Support multiple refs in fetch command (Steve Streeting[32m[m)
[32m|[m [31m|[m[31m/[m  
* [31m|[m   da95c18 2015-07-30 Merge pull request #541 from sinbad/filter-tests (risk danger olson[32m[m)
[31m|[m[35m\[m [31m\[m  
[31m|[m [35m|[m[31m/[m  
[31m|[m[31m/[m[35m|[m   
[31m|[m * eeaacc1 2015-07-30 Eliminate duplication on filter tests (Steve Streeting[32m[m)
[31m|[m[31m/[m  
*   fa57c15 2015-07-29 Merge pull request #527 from sinbad/fetch-pull-checkout (risk danger olson[32m[m)
[36m|[m[1;31m\[m  
[36m|[m *   0b49d6b 2015-07-29 Bring fetch-pull-checkout up to date with master & resolve conflicts (Steve Streeting[32m[m)
[36m|[m [1;32m|[m[1;33m\[m  
[36m|[m * [1;33m|[m a2be01d 2015-07-29 Avoid deadlock when output channel is nil in fetchAndReportToChan (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 0517dc2 2015-07-29 Fix Travis - damn this sh vs bash issue (on OS X sh==bash) (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 409c68e 2015-07-29 Correctly return NotAPointerError when file < blobSizeCutoff (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m d719b6f 2015-07-29 Test checkout with path args and in subdirs (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 50191c5 2015-07-29 Fix vendor import (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m d3d5999 2015-07-29 Make pull parallelise fetch/checkout again (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m a44c595 2015-07-28 ConvertCwdFilesRelativeToRepo should have passthrough early-out too (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 760c7d7 2015-07-28 Fix a bunch of issues related to checkout in non-root directories (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 1ac8ed6 2015-07-28 Be sure to create all folders required to smudge a file (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m de8c59a 2015-07-28 Use goimports to format imports (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m a381655 2015-07-28 Tests for FilenamePassesIncludeExcludeFilter (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 7bd4b37 2015-07-28 Cache platform detection & make it a bit more useful in other cases (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 9a793dd 2015-07-28 Fix typo (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 45f3ae6 2015-07-28 Refactor to use PointerSmudgeToFile and defer the file close to avoid Panic leaving handle open (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 7df8db2 2015-07-28 Externalise goroutine and wait from checkoutWithChan (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 05de710 2015-07-27 Travis fails when all is well here, sh/bash issue again? (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m f71ece1 2015-07-27 Fix pre-push test; there is now no output from push when nothing to do (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 1d05552 2015-07-27 Add checkout integration test (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 23e4eeb 2015-07-27 Introduce new ScanTree func which reports files with same content & use in checkout (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 85823eb 2015-07-27 Add support for checking out specific paths (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m e5e4dd9 2015-07-27 Fix pull test (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 0159f6c 2015-07-27 Change pull to do a simple full fetch & checkout instead of checkout in parallel of fetched items (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m da9a48a 2015-07-27 Defer the wait.Done() for greater resilience (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 096b6da 2015-07-24 Add pull test (failing initially) (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m cc6032a 2015-07-24 Need to pass the wait group by pointer to checkout (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 203e483 2015-07-24 Early-out the transfer queue when there's nothing to do (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 16dab09 2015-07-24 Only add oids to the fetch queue which are actually missing or wrong size locally (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m afed34a 2015-07-24 Add short command info (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m f55d805 2015-07-24 Fix fetch test, only downloads now, doesn't check out (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m 001ddcd 2015-07-24 Work in progress refactoring fetch into fetch/checkout/pull (Steve Streeting[32m[m)
* [1;33m|[m [1;33m|[m e72b187 2015-07-29 0.5.3 released in July, not June :) (Rick Olson[32m[m)
* [1;33m|[m [1;33m|[m   4549581 2015-07-29 Merge pull request #532 from github/roadmap-installation (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m * [1;33m|[m [1;33m|[m c1032f9 2015-07-28 add install experience to the roadmap (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [1;33m|[m 64b541e 2015-07-28 describe the common issue labels we use (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [1;33m|[m   277f8d6 2015-07-29 Merge pull request #537 from github/reorg-error-logs (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m 3ac1324 2015-07-28 fix test (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m a10947e 2015-07-28 re-organize the error logs to make them more useful (Rick Olson[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   18b9fa2 2015-07-29 Merge pull request #536 from github/sha-in-user-agent (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m da9368a 2015-07-28 pass the ldflag correctly (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m 0cf7c47 2015-07-28 pass the ldflags in script/bootstrap (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m 22bb4ab 2015-07-28 remove unnecessary script/fmt call in script/build.go (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m 960966b 2015-07-28 dump git commit from `script/run version` (Rick Olson[32m[m)
[32m|[m [31m|[m[31m/[m [1;35m/[m [1;33m/[m [1;33m/[m  
* [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   3bfde33 2015-07-29 Merge pull request #530 from github/fix-uninit (risk danger olson[32m[m)
[34m|[m[35m\[m [31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[34m|[m * [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m eb03c46 2015-07-27 teach 'uninit' how to remove git config sections (Rick Olson[32m[m)
* [35m|[m [31m|[m [1;35m|[m [1;33m|[m [1;33m|[m   9bf08c8 2015-07-29 Merge pull request #538 from sinbad/detect-ronn (risk danger olson[32m[m)
[1;33m|[m[1;31m\[m [35m\[m [31m\[m [1;35m\[m [1;33m\[m [1;33m\[m  
[1;33m|[m [1;31m|[m[1;33m_[m[35m|[m[1;33m_[m[31m|[m[1;33m_[m[1;35m|[m[1;33m_[m[1;33m|[m[1;33m/[m  
[1;33m|[m[1;33m/[m[1;31m|[m [35m|[m [31m|[m [1;35m|[m [1;33m|[m   
[1;33m|[m * [35m|[m [31m|[m [1;35m|[m [1;33m|[m b1ff4e1 2015-07-29 Correctly detect the absence of `ronn` (Steve Streeting[32m[m)
[1;33m|[m[1;33m/[m [35m/[m [31m/[m [1;35m/[m [1;33m/[m  
* [35m|[m [31m|[m [1;35m|[m [1;33m|[m   bbd6dbc 2015-07-29 Merge pull request #476 from github/batches-of-n (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [35m\[m [31m\[m [1;35m\[m [1;33m\[m  
[1;32m|[m * [35m|[m [31m|[m [1;35m|[m [1;33m|[m 8b31fde 2015-07-29 fix typo (rubyist[32m[m)
[1;32m|[m * [35m|[m [31m|[m [1;35m|[m [1;33m|[m   e17c3cc 2015-07-29 Merge branch 'master' into batches-of-n (rubyist[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m [35m\[m [31m\[m [1;35m\[m [1;33m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m [35m/[m [31m/[m [1;35m/[m [1;33m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m [35m|[m [31m|[m [1;35m|[m [1;33m|[m   
* [1;34m|[m [35m|[m [31m|[m [1;35m|[m [1;33m|[m   312350b 2015-07-29 Merge pull request #534 from github/ssh-credentials-fix (risk danger olson[32m[m)
[31m|[m[31m\[m [1;34m\[m [35m\[m [31m\[m [1;35m\[m [1;33m\[m  
[31m|[m [31m|[m[31m_[m[1;34m|[m[31m_[m[35m|[m[31m/[m [1;35m/[m [1;33m/[m  
[31m|[m[31m/[m[31m|[m [1;34m|[m [35m|[m [1;35m|[m [1;33m|[m   
[31m|[m * [1;34m|[m [35m|[m [1;35m|[m [1;33m|[m 6f8a333 2015-07-28 Remove redundant nil check for a map (Rick Olson[32m[m)
[31m|[m * [1;34m|[m [35m|[m [1;35m|[m [1;33m|[m 622d973 2015-07-28 newClientRequest() now accepts an initial header (Rick Olson[32m[m)
[31m|[m[31m/[m [1;34m/[m [35m/[m [1;35m/[m [1;33m/[m  
* [1;34m|[m [35m|[m [1;35m|[m [1;33m|[m   c8649e4 2015-07-28 Merge pull request #511 from andyneff/docker-scripts (risk danger olson[32m[m)
[1;35m|[m[33m\[m [1;34m\[m [35m\[m [1;35m\[m [1;33m\[m  
[1;35m|[m [33m|[m[1;35m_[m[1;34m|[m[1;35m_[m[35m|[m[1;35m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[33m|[m [1;34m|[m [35m|[m [1;33m|[m   
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m daeca88 2015-07-28 Update README.md (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 8de97e0 2015-07-27 Fixed markdown (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m c2d25ee 2015-07-23 Added 32 bit RPM build support (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m c1b093b 2015-07-23 Bump for 0.5.3 (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m adac578 2015-07-23 Fixed Array length call (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 76bbb61 2015-07-23 Cleanup (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 2ec4d2b 2015-07-23 Update README.md (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 183ad44 2015-07-23 Move the repos directory (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 4d60430 2015-07-23 Readme update (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 5440bdb 2015-07-23 Fixed CentOS 5 bug in test (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m f74e3d5 2015-07-23 Fixed all signing and testing bugs (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 7ca335b 2015-07-23 Dockers working for centos and debian with GPG signing (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 9d2aef4 2015-07-23 First draft of GPG signing working in CentOS (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m cb0d7d4 2015-07-23 Need blank public.key at least (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m bb21163 2015-07-23 Added Repo Tests (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 68b29af 2015-07-23 Got Debian signing packages (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 28acd96 2015-07-23 First attempt at GPG integration (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m a5dfb17 2015-07-23 Just checking files in to test repo generator (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 195a82f 2015-07-23 Centos full builds work (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m bf6516b 2015-07-23 Centos repos now build (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 82fc36c 2015-07-23 Clean up (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 66f9455 2015-07-23 Building debian 8 repo successfully (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m a359dc3 2015-07-23 First working draft of docker repository scripts working (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 563f6d4 2015-07-23 I've accepted this is a failed attemped to try to easily build all OSes (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 8f0f70c 2015-07-23 Split script into two files (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 0e2b566 2015-07-23 Rough draft of build script (Andy Neff[32m[m)
[1;35m|[m * [1;34m|[m [35m|[m [1;33m|[m 3a228d5 2015-07-23 Initial docker files (Andy Neff[32m[m)
[1;35m|[m [1;33m|[m [1;34m|[m[1;33m_[m[35m|[m[1;33m/[m  
[1;35m|[m [1;33m|[m[1;33m/[m[1;34m|[m [35m|[m   
* [1;33m|[m [1;34m|[m [35m|[m 4fee3dd 2015-07-28 some more test env details to help debug intermittent travis issues (Rick Olson[32m[m)
* [1;33m|[m [1;34m|[m [35m|[m   2277237 2015-07-28 Merge branch 'batch-api-schema' (Rick Olson[32m[m)
[35m|[m[35m\[m [1;33m\[m [1;34m\[m [35m\[m  
[35m|[m [35m|[m[35m_[m[1;33m|[m[35m_[m[1;34m|[m[35m/[m  
[35m|[m[35m/[m[35m|[m [1;33m|[m [1;34m|[m   
[35m|[m * [1;33m|[m [1;34m|[m 751d412 2015-07-28 pull in schema changes from #528 (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [1;34m|[m aecd968 2015-07-27 updates the batch api docs (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [1;34m|[m   9c947e1 2015-07-27 Merge branch 'master' into batch-api-schema (Rick Olson[32m[m)
[35m|[m [36m|[m[1;31m\[m [1;33m\[m [1;34m\[m  
[35m|[m * [1;31m|[m [1;33m|[m [1;34m|[m 3202118 2015-07-20 more definitions: (Rick Olson[32m[m)
[35m|[m * [1;31m|[m [1;33m|[m [1;34m|[m 9bde1bf 2015-06-24 fix son error (Rick Olson[32m[m)
[35m|[m * [1;31m|[m [1;33m|[m [1;34m|[m 8cf0c9f 2015-06-24 specify size on response objects (Rick Olson[32m[m)
[35m|[m * [1;31m|[m [1;33m|[m [1;34m|[m a7e15e6 2015-06-24 initial json schemas (Rick Olson[32m[m)
* [1;31m|[m [1;31m|[m [1;33m|[m [1;34m|[m   990ffe1 2015-07-27 Merge pull request #518 from sinbad/integration-test-arguments (risk danger olson[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m [1;31m\[m [1;33m\[m [1;34m\[m  
[1;31m|[m [1;33m|[m[1;31m_[m[1;31m|[m[1;31m/[m [1;33m/[m [1;34m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m [1;31m|[m [1;33m|[m [1;34m|[m   
[1;31m|[m * [1;31m|[m [1;33m|[m [1;34m|[m a8b3b16 2015-07-25 Require bash for array support (Steve Streeting[32m[m)
[1;31m|[m * [1;31m|[m [1;33m|[m [1;34m|[m 16ce825 2015-07-24 Allow a subset of tests to be run by script/integration with arguments (Steve Streeting[32m[m)
[1;31m|[m [1;33m|[m [1;31m|[m[1;33m/[m [1;34m/[m  
[1;31m|[m [1;33m|[m[1;33m/[m[1;31m|[m [1;34m|[m   
* [1;33m|[m [1;31m|[m [1;34m|[m   fac4434 2015-07-27 Merge pull request #526 from andyneff/debain-cross-compile (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;31m\[m [1;34m\[m  
[1;34m|[m * [1;33m|[m [1;31m|[m [1;34m|[m 432e1fb 2015-07-26 32 bit deb's cross compiling correctly (Andy Neff[32m[m)
[1;34m|[m * [1;33m|[m [1;31m|[m [1;34m|[m 9248d6a 2015-07-26 32 bit debian build WORKING, but a little complicated (Andy Neff[32m[m)
[1;34m|[m[1;34m/[m [1;33m/[m [1;31m/[m [1;34m/[m  
[1;34m|[m [1;33m|[m [1;31m|[m * 4fbb155 2015-07-27 Add some docs, minor refactor (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * 20e8a58 2015-07-27 refine progress output (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * b16cb98 2015-07-27 Add transfer to progress bar when it's added to the channel (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * 4539fcf 2015-07-27 Remove some stale code from batcher (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * 951f305 2015-07-27 Had parameters swapped (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * ab5b35a 2015-07-27 Progress meter needs to be more dynamic for batches of n (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m * 1c4b0aa 2015-07-27 remove spelling error I reintroduced (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m *   f758553 2015-07-27 Merge branch 'master' into batches-of-n (rubyist[32m[m)
[1;34m|[m [1;33m|[m [1;31m|[m [1;36m|[m[1;34m\[m  
[1;34m|[m [1;33m|[m[1;34m_[m[1;31m|[m[1;34m_[m[1;36m|[m[1;34m/[m  
[1;34m|[m[1;34m/[m[1;33m|[m [1;31m|[m [1;36m|[m   
* [1;33m|[m [1;31m|[m [1;36m|[m   c2b7f94 2015-07-24 Merge pull request #522 from ssgelm/fix-typos (risk danger olson[32m[m)
[32m|[m[33m\[m [1;33m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [1;33m|[m [1;31m|[m [1;36m|[m 445c0bb 2015-07-24 Fix typos in transfer_queue.go (Stephen Gelman[32m[m)
* [33m|[m [1;33m|[m [1;31m|[m [1;36m|[m   a9b3ed4 2015-07-24 Merge pull request #521 from ssgelm/update-deb-version-0.5.3 (risk danger olson[32m[m)
[33m|[m[35m\[m [33m\[m [1;33m\[m [1;31m\[m [1;36m\[m  
[33m|[m [35m|[m[33m/[m [1;33m/[m [1;31m/[m [1;36m/[m  
[33m|[m[33m/[m[35m|[m [1;33m|[m [1;31m|[m [1;36m|[m   
[33m|[m * [1;33m|[m [1;31m|[m [1;36m|[m 9c67bef 2015-07-24 Update deb version to 0.5.3 (Stephen Gelman[32m[m)
[33m|[m[33m/[m [1;33m/[m [1;31m/[m [1;36m/[m  
* [1;33m|[m [1;31m|[m [1;36m|[m   20fc8dc 2015-07-24 Merge pull request #520 from sinbad/integration-test-fix-osx-keychain-errors (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;33m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [1;33m|[m [1;31m|[m [1;36m|[m 07936c1 2015-07-24 Remove verbose output (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;36m|[m 9fe112d 2015-07-24 Fix issues with LocalItems keychain after trying to use replacement login.keychain (Steve Streeting[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;36m|[m 332b229 2015-07-24 Fix annoying keychain errors when running integration tests on OS X (Steve Streeting[32m[m)
[36m|[m [1;33m|[m[1;33m/[m [1;31m/[m [1;36m/[m  
* [1;33m|[m [1;31m|[m [1;36m|[m   04d9b01 2015-07-24 Merge pull request #519 from nuta/fix-install-sh (risk danger olson[32m[m)
[1;33m|[m[1;33m\[m [1;33m\[m [1;31m\[m [1;36m\[m  
[1;33m|[m [1;33m|[m[1;33m/[m [1;31m/[m [1;36m/[m  
[1;33m|[m[1;33m/[m[1;33m|[m [1;31m|[m [1;36m|[m   
[1;33m|[m * [1;31m|[m [1;36m|[m 2bab2fd 2015-07-24 don't use install(1) with -D option (Seiya Nuta[32m[m)
[1;33m|[m[1;33m/[m [1;31m/[m [1;36m/[m  
* [1;31m|[m [1;36m|[m c98bc5b 2015-07-23 update release guide with the correct script for running all tests (Rick Olson[32m (tag: v0.5.3)[m)
* [1;31m|[m [1;36m|[m 05579fd 2015-07-23 update changelog (Rick Olson[32m[m)
* [1;31m|[m [1;36m|[m   b4ade67 2015-07-23 Merge branch 'release-0.5.3' (Rick Olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m * [1;31m|[m [1;36m|[m ef4fdb2 2015-07-22 yay v0.5.3 (Rick Olson[32m[m)
* [1;35m|[m [1;31m|[m [1;36m|[m   c3756f9 2015-07-23 Merge pull request #491 from github/roadmap-local-storage (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m 9bff45d 2015-07-23 add a quick overview of the next major releases (risk danger olson[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m 4921655 2015-07-23 Add centos/apt to the list (risk danger olson[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m fe94492 2015-07-22 update roadmap to point to the new local storage management issue (Rick Olson[32m[m)
* [31m|[m [1;35m|[m [1;31m|[m [1;36m|[m   f223cce 2015-07-23 Merge pull request #510 from andyneff/test_clean_newline_fix (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[32m|[m * [31m|[m [1;35m|[m [1;31m|[m [1;36m|[m c832b33 2015-07-22 Fixed pseudo pointer with extra data to work on CentOS (Andy Neff[32m[m)
[32m|[m [1;35m|[m [31m|[m[1;35m/[m [1;31m/[m [1;36m/[m  
[32m|[m [1;35m|[m[1;35m/[m[31m|[m [1;31m|[m [1;36m|[m   
* [1;35m|[m [31m|[m [1;31m|[m [1;36m|[m   0e49d5a 2015-07-23 Merge pull request #509 from github/doc-push-object-id (risk danger olson[32m[m)
[1;35m|[m[35m\[m [1;35m\[m [31m\[m [1;31m\[m [1;36m\[m  
[1;35m|[m [35m|[m[1;35m/[m [31m/[m [1;31m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[35m|[m [31m|[m [1;31m|[m [1;36m|[m   
[1;35m|[m * [31m|[m [1;31m|[m [1;36m|[m be24b11 2015-07-22 documents git lfs push --object-id (Rick Olson[32m[m)
[1;35m|[m[1;35m/[m [31m/[m [1;31m/[m [1;36m/[m  
* [31m|[m [1;31m|[m [1;36m|[m   67b9cf0 2015-07-22 Merge pull request #489 from github/backport-script (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [31m|[m [1;31m|[m [1;36m|[m 3f9ea9c 2015-07-22 update this to push against the oss repo (Rick Olson[32m[m)
[36m|[m * [31m|[m [1;31m|[m [1;36m|[m 27417d2 2015-07-21 uniq (Rick Olson[32m[m)
[36m|[m * [31m|[m [1;31m|[m [1;36m|[m cda1fd4 2015-07-21 initial backport script (Rick Olson[32m[m)
[36m|[m [31m|[m[31m/[m [1;31m/[m [1;36m/[m  
* [31m|[m [1;31m|[m [1;36m|[m   9b1c989 2015-07-22 Merge pull request #488 from github/http-nil-res-body (risk danger olson[32m[m)
[31m|[m[1;33m\[m [31m\[m [1;31m\[m [1;36m\[m  
[31m|[m [1;33m|[m[31m/[m [1;31m/[m [1;36m/[m  
[31m|[m[31m/[m[1;33m|[m [1;31m|[m [1;36m|[m   
[31m|[m * [1;31m|[m [1;36m|[m bbf2d16 2015-07-21 ensure responses always have a body (risk danger olson[32m[m)
[31m|[m[31m/[m [1;31m/[m [1;36m/[m  
* [1;31m|[m [1;36m|[m   1180e5e 2015-07-21 Merge pull request #485 from github/linter (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;36m\[m  
[1;34m|[m * [1;31m|[m [1;36m|[m 47f0129 2015-07-21 bash bash bash (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [1;36m|[m 3e6b15c 2015-07-21 run script/lint at the end of script/fmt (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [1;36m|[m b2bf353 2015-07-21 add another helpful msg to script/lint (risk danger olson[32m[m)
[1;34m|[m * [1;31m|[m [1;36m|[m 0f007b4 2015-07-21 Script to warn about non-vendored third party packages (rubyist[32m[m)
* [1;35m|[m [1;31m|[m [1;36m|[m   a1efa58 2015-07-21 Merge pull request #483 from github/proposals (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m 8c53479 2015-07-20 lfs-extension => lfs.ext (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m a4657d2 2015-07-20 spacing (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;31m|[m [1;36m|[m becbd14 2015-07-20 document proposal process (Rick Olson[32m[m)
* [31m|[m [1;35m|[m [1;31m|[m [1;36m|[m   02d779d 2015-07-21 Merge pull request #484 from michael-k/naming (risk danger olson[32m[m)
[1;35m|[m[33m\[m [31m\[m [1;35m\[m [1;31m\[m [1;36m\[m  
[1;35m|[m [33m|[m[1;35m_[m[31m|[m[1;35m/[m [1;31m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[33m|[m [31m|[m [1;31m|[m [1;36m|[m   
[1;35m|[m * [31m|[m [1;31m|[m [1;36m|[m 20bd658 2015-07-20 Name uninit file according to other commands (Michael Käufl[32m[m)
[1;35m|[m[1;35m/[m [31m/[m [1;31m/[m [1;36m/[m  
* [31m|[m [1;31m|[m [1;36m|[m   2ebafda 2015-07-20 Merge pull request #465 from github/better-init (risk danger olson[32m[m)
[31m|[m[35m\[m [31m\[m [1;31m\[m [1;36m\[m  
[31m|[m [35m|[m[31m/[m [1;31m/[m [1;36m/[m  
[31m|[m[31m/[m[35m|[m [1;31m|[m [1;36m|[m   
[31m|[m * [1;31m|[m [1;36m|[m 572b949 2015-07-20 only one pre-push hook to worry about (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 403d2f7 2015-07-10 better wording (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m f3c4c07 2015-07-10 dont update global gitconfig in tests (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m c97c5dc 2015-07-10 Revert 'this tempdir home stuff is unnecessary' (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 6eaa77c 2015-07-10 only remove pre-push if it's a known git lfs pre-push hook (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 8407e01 2015-07-10 add uninit docs (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 0e30dba 2015-07-10 this tempdir home stuff is unnecessary (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m bf2691c 2015-07-10 add an uninit command (risk danger olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 3187e0e 2015-07-07 document init --force (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 37f62bd 2015-07-07 learn grep -c (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m 82df715 2015-07-06 test that (MISSING) is not in the init message anymore (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m fb4cc48 2015-07-06 test --force (Rick Olson[32m[m)
[31m|[m * [1;31m|[m [1;36m|[m be80114 2015-07-06 add init --force (Rick Olson[32m[m)
* [35m|[m [1;31m|[m [1;36m|[m   d77bc03 2015-07-20 Merge pull request #444 from sanoursa/ExtendingLFS (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [1;31m\[m [1;36m\[m  
[36m|[m * [35m|[m [1;31m|[m [1;36m|[m 64f1c89 2015-07-13 Moved the extensions design to a separate file under docs/proposals (Saeed Noursalehi[32m[m)
[36m|[m * [35m|[m [1;31m|[m [1;36m|[m ff18881 2015-07-13 Added a section on error handling, and made some other minor tweaks. (Saeed Noursalehi[32m[m)
[36m|[m * [35m|[m [1;31m|[m [1;36m|[m 63024ce 2015-06-26 Design proposal for making Git LFS extensible (Saeed Noursalehi[32m[m)
[36m|[m [1;31m|[m [35m|[m[1;31m/[m [1;36m/[m  
[36m|[m [1;31m|[m[1;31m/[m[35m|[m [1;36m|[m   
* [1;31m|[m [35m|[m [1;36m|[m   e46f9b8 2015-07-16 Merge pull request #480 from andyneff/git_for_centos (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m [1;36m\[m  
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m a4c9db8 2015-07-13 Added shasum to download list so tests can pass (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m 304c7d8 2015-07-12 Added git fixes to cento-build too (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m 6beee4f 2015-07-12 Fixed GIT_VERSION parsing (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m 4431db5 2015-07-12 Docs update (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m 5ecc286 2015-07-12 Added Bad owner/group workaround for build SRPMS (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m a7e8bd5 2015-07-12 Added Centos 5 support for git (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m 182ae63 2015-07-12 Simplified the SUDO variable out, since most people shouldn't need it (Andy Neff[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [1;36m|[m cd3e92e 2015-07-12 Added git building for CentOS 6 (Andy Neff[32m[m)
[1;32m|[m[1;32m/[m [1;31m/[m [35m/[m [1;36m/[m  
* [1;31m|[m [35m|[m [1;36m|[m   2eef603 2015-07-10 Merge pull request #475 from github/crlf-to-text (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [35m\[m [1;36m\[m  
[1;34m|[m * [1;31m|[m [35m|[m [1;36m|[m 7146bdb 2015-07-09 Use -text instead of -crlf (Rick Olson[32m[m)
* [1;35m|[m [1;31m|[m [35m|[m [1;36m|[m 0fb9c7b 2015-07-09 add more to the release checklist (Rick Olson[32m[m)
* [1;35m|[m [1;31m|[m [35m|[m [1;36m|[m   3d50543 2015-07-09 Merge pull request #469 from ssgelm/deb-cleanup (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [35m\[m [1;36m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [35m|[m [1;36m|[m dbc190e 2015-07-08 Some cleanup to make debian package pass lintian checks (Stephen Gelman[32m[m)
* [31m|[m [1;35m|[m [1;31m|[m [35m|[m [1;36m|[m   eaf8b13 2015-07-09 Merge pull request #471 from github/code-of-conduct (risk danger olson[32m[m)
[1;35m|[m[33m\[m [31m\[m [1;35m\[m [1;31m\[m [35m\[m [1;36m\[m  
[1;35m|[m [33m|[m[1;35m_[m[31m|[m[1;35m/[m [1;31m/[m [35m/[m [1;36m/[m  
[1;35m|[m[1;35m/[m[33m|[m [31m|[m [1;31m|[m [35m|[m [1;36m|[m   
[1;35m|[m * [31m|[m [1;31m|[m [35m|[m [1;36m|[m 32ef884 2015-07-08 Update url to code of conduct (Brandon Keepers[32m[m)
[1;35m|[m * [31m|[m [1;31m|[m [35m|[m [1;36m|[m c5c0173 2015-07-08 Add code of conduct to contributing guidelines (Brandon Keepers[32m[m)
[1;35m|[m [31m|[m[31m/[m [1;31m/[m [35m/[m [1;36m/[m  
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * f16c765 2015-07-16 Done() the waitgroup for objects that don't get added to the transfer queue (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 565f8f4 2015-07-15 keep batch errors (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * c20272c 2015-07-15 Remove batcher timeout (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 710862b 2015-07-10 Combine GIT_LFS_PROGRESS logger with ProgressMeter (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 38d14c2 2015-07-10 First pass at extracting progress meter (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * dcd975f 2015-07-09 Fix import again (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 5bcde76 2015-07-09 Fix import (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * b7420e4 2015-07-09 Add some tracing (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 7b48834 2015-07-09 Set a batchSize const (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * ea7e5fc 2015-07-09 Fallback to individual when server doesn't implement batching (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * a7742bf 2015-07-09 The apiEvent mechanism is no longer needed (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * c1fc449 2015-07-09 Remove excessive debug tracing (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 1def203 2015-07-09 Add back the progress output (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * 94ecd16 2015-07-09 Start integrating batcher into transfer queue (rubyist[32m[m)
[1;35m|[m [31m|[m [1;31m|[m [35m|[m * b9c4b3a 2015-07-09 Introduce a batcher that batches things into groups of n (rubyist[32m[m)
[1;35m|[m [31m|[m[1;35m_[m[1;31m|[m[1;35m_[m[35m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[31m|[m [1;31m|[m [35m|[m   
* [31m|[m [1;31m|[m [35m|[m   70297ff 2015-07-09 Merge pull request #472 from github/bodynil (Scott Barron[32m[m)
[34m|[m[35m\[m [31m\[m [1;31m\[m [35m\[m  
[34m|[m * [31m|[m [1;31m|[m [35m|[m 58154c3 2015-07-08 Check doHttpRequest() error in UploadObject() (rubyist[32m[m)
[34m|[m [31m|[m[31m/[m [1;31m/[m [35m/[m  
* [31m|[m [1;31m|[m [35m|[m   855f1af 2015-07-09 Merge pull request #468 from github/batch401fix (Scott Barron[32m[m)
[31m|[m[1;31m\[m [31m\[m [1;31m\[m [35m\[m  
[31m|[m [1;31m|[m[31m/[m [1;31m/[m [35m/[m  
[31m|[m[31m/[m[1;31m|[m [1;31m|[m [35m|[m   
[31m|[m * [1;31m|[m [35m|[m 214c352 2015-07-09 Fix import line (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m e00d337 2015-07-09 Add an extra tracer (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 931d0fb 2015-07-07 Fully re-submit batch requests on 401 (rubyist[32m[m)
[31m|[m[31m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [35m|[m   f57eb5e 2015-07-07 Merge pull request #467 from github/gopath-setup (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m  
[1;32m|[m * [1;31m|[m [35m|[m f2138f9 2015-07-07 bootstrap: fix syntax error. (Mike McQuaid[32m[m)
[1;32m|[m[1;32m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [35m|[m   a87daa6 2015-07-07 Merge pull request #458 from github/gopath-setup (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;31m|[m [35m|[m 075bead 2015-07-05 bootstrap: only set GOPATH on non-Windows. (Mike McQuaid[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m b45aeef 2015-07-03 bootstrap: setup GOPATH if it isn't already set. (Mike McQuaid[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m feebd87 2015-07-03 script/bootstrap: remove unneeded `pwd` (Mike McQuaid[32m[m)
[1;34m|[m * [1;31m|[m [35m|[m 53f3d2d 2015-07-03 script/bootstrap: exit on any errors (Mike McQuaid[32m[m)
* [1;35m|[m [1;31m|[m [35m|[m   a53565a 2015-07-07 Merge pull request #464 from github/ssh-refactor (risk danger olson[32m[m)
[35m|[m[31m\[m [1;35m\[m [1;31m\[m [35m\[m  
[35m|[m [31m|[m[35m_[m[1;35m|[m[35m_[m[1;31m|[m[35m/[m  
[35m|[m[35m/[m[31m|[m [1;35m|[m [1;31m|[m   
[35m|[m * [1;35m|[m [1;31m|[m 41f1b02 2015-07-06 Refactor some of the code in #404 (Rick Olson[32m[m)
[35m|[m[35m/[m [1;35m/[m [1;31m/[m  
* [1;35m|[m [1;31m|[m   5299e0c 2015-07-06 Merge branch 'sinbad-ssh_urls' (Rick Olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;31m\[m  
[32m|[m * [1;35m\[m [1;31m\[m   e9d4cf6 2015-07-06 Merge branch 'ssh_urls' of https://github.com/sinbad/git-lfs into sinbad-ssh_urls (Rick Olson[32m[m)
[32m|[m [32m|[m[35m\[m [1;35m\[m [1;31m\[m  
[32m|[m[32m/[m [35m/[m [1;35m/[m [1;31m/[m  
[32m|[m * [1;35m|[m [1;31m|[m 2892978 2015-06-22 In the rare case that url.Parse() returns an error, return an explicit EndpointUrlUnknown (Steve Streeting[32m[m)
[32m|[m * [1;35m|[m [1;31m|[m 72795a3 2015-06-22 Improved SSH support: non-bare urls, custom ports, and GIT_SSH/plinks (Steve Streeting[32m[m)
* [35m|[m [1;35m|[m [1;31m|[m   35970a5 2015-07-06 Merge pull request #461 from github/push-deleted-files (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [1;35m\[m [1;31m\[m  
[36m|[m * [35m|[m [1;35m|[m [1;31m|[m c8fe922 2015-07-06 update assert_server_object usage (Rick Olson[32m[m)
[36m|[m * [35m|[m [1;35m|[m [1;31m|[m   0345ca2 2015-07-06 Merge branch 'master' into push-deleted-files (Rick Olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [35m\[m [1;35m\[m [1;31m\[m  
[36m|[m [1;32m|[m[36m/[m [35m/[m [1;35m/[m [1;31m/[m  
[36m|[m[36m/[m[1;32m|[m [35m|[m [1;35m|[m [1;31m|[m   
* [1;32m|[m [35m|[m [1;35m|[m [1;31m|[m   4136153 2015-07-06 Merge pull request #463 from github/test-helper-fixes (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m [35m\[m [1;35m\[m [1;31m\[m  
[1;34m|[m * [1;32m|[m [35m|[m [1;35m|[m [1;31m|[m 27cf051 2015-07-06 update test server to store unique lfs objects per repo (Rick Olson[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m [1;35m|[m [1;31m|[m 89885bd 2015-07-06 turn largeObjects into a struct with methods (Rick Olson[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m [1;35m|[m [1;31m|[m 9928f73 2015-07-06 implement refute_server_object and assert_server_object (Rick Olson[32m[m)
[1;34m|[m * [1;32m|[m [35m|[m [1;35m|[m [1;31m|[m ec1e0f8 2015-07-06 actually call 'which' (Rick Olson[32m[m)
[1;34m|[m[1;34m/[m [1;32m/[m [35m/[m [1;35m/[m [1;31m/[m  
[1;34m|[m * [35m|[m [1;35m|[m [1;31m|[m 9b8fa2f 2015-07-06 fix ls-files so it doesnt list deleted objects (Rick Olson[32m[m)
[1;34m|[m * [35m|[m [1;35m|[m [1;31m|[m 821e8b1 2015-07-06 Don't stat the file in the working dir (Rick Olson[32m[m)
[1;34m|[m * [35m|[m [1;35m|[m [1;31m|[m   a7db703 2015-07-06 merge master (Rick Olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [35m\[m [1;35m\[m [1;31m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [35m/[m [1;35m/[m [1;31m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [35m|[m [1;35m|[m [1;31m|[m   
* [1;36m|[m [35m|[m [1;35m|[m [1;31m|[m   1f24533 2015-07-06 Merge pull request #447 from github/clean-fix (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [35m\[m [1;35m\[m [1;31m\[m  
[32m|[m * [1;36m\[m [35m\[m [1;35m\[m [1;31m\[m   4b1dd8a 2015-07-06 merge master again (Rick Olson[32m[m)
[32m|[m [34m|[m[32m\[m [1;36m\[m [35m\[m [1;35m\[m [1;31m\[m  
[32m|[m [34m|[m[32m/[m [1;36m/[m [35m/[m [1;35m/[m [1;31m/[m  
[32m|[m[32m/[m[34m|[m [1;36m|[m [35m|[m [1;35m|[m [1;31m|[m   
* [34m|[m [1;36m|[m [35m|[m [1;35m|[m [1;31m|[m   63f9449 2015-07-06 Merge pull request #449 from larsxschneider/lars/push-object-ids (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [34m\[m [1;36m\[m [35m\[m [1;35m\[m [1;31m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[34m|[m[1;35m_[m[1;36m|[m[1;35m_[m[35m|[m[1;35m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [34m|[m [1;36m|[m [35m|[m [1;31m|[m   
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m bf9ab31 2015-07-06 Support dry run for object ID upload. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m 08c00bb 2015-07-06 Add upload trace message to object ID upload. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m 853e114 2015-07-06 Improve trace message for file upload. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m ce431b0 2015-07-06 Fix push "--object-id" usage message. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m dee8383 2015-07-06 Revert "assert_file_line_count" test helper (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m 431db15 2015-07-06 Rename `pathForStats` to `statsPath` (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m 92ea3b5 2015-07-06 Use shorthand variable declaration. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m c84c7d0 2015-07-06 Add assertion to check the number of output lines. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m cd3844a 2015-07-06 Cleanup test case. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m d52b188 2015-07-06 Remove unnecessary parameters from NewUploadable function. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m 29f6944 2015-07-06 Remove superfluous NewUploadableWithoutFilename function. (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m d55f813 2015-07-06 Rename command line parameter "objectid" to "object-id". https://github.com/github/git-lfs/pull/449/files#r33897536 (Lars Schneider[32m[m)
[1;35m|[m * [34m|[m [1;36m|[m [35m|[m [1;31m|[m d24ab63 2015-06-28 Add option to push object ID(s) directly. (Lars Schneider[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m 73fff90 2015-07-06 teardown cleaned object asap (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m   42f0bb5 2015-07-06 merge master (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;32m|[m[1;35m\[m [1;36m\[m [35m\[m [1;31m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[1;32m|[m[1;35m/[m [1;36m/[m [35m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;32m|[m [1;36m|[m [35m|[m [1;31m|[m   
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m 8b3fa6a 2015-06-28 change message when lfs object is missing (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m 6c0a9af 2015-06-28 add a push test containing git and lfs objects (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m d75291f 2015-06-28 no need to pass this file around if callers just immediately close it (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m fe5dffc 2015-06-28 remove unused attr (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m * [1;36m|[m [35m|[m [1;31m|[m a251b2e 2015-06-28 parallelize clean tests (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;31m|[m [1;36m|[m[1;31m_[m[35m|[m[1;31m/[m  
[1;35m|[m [1;31m|[m [1;31m|[m[1;31m/[m[1;36m|[m [35m|[m   
[1;35m|[m [1;31m|[m [1;31m|[m * [35m|[m 5480801 2015-07-05 add options object to ref scanner (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;31m|[m * [35m|[m d008ac6 2015-07-05 failing test (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;31m|[m * [35m|[m ee16ae5 2015-07-05 we just need the file size. the file doesn't need to be in the working dir (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;31m|[m * [35m|[m 561f12b 2015-07-05 this forces the scanner to look for deleted objects (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;31m|[m * [35m|[m 25fe0c9 2015-07-05 failing test that proves git lfs misses deleted files (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m[1;35m_[m[1;31m|[m[1;35m/[m [35m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;31m|[m [35m|[m   
* [1;31m|[m [1;31m|[m [35m|[m   96c94aa 2015-07-01 Merge pull request #386 from ryansimmen/chunkedTransferEncoding (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;31m|[m [1;31m|[m [35m|[m 962be37 2015-07-01 Support chunked Transfer-Encoding (Ryan Simmen[32m[m)
[1;34m|[m[1;34m/[m [1;31m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [1;31m|[m [35m|[m   363fc71 2015-06-29 Merge pull request #451 from github/error-format (Scott Barron[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [1;31m\[m [35m\[m  
[1;36m|[m * [1;31m\[m [1;31m\[m [35m\[m   2e8361e 2015-06-29 Merge branch 'master' into error-format (rubyist[32m[m)
[1;36m|[m [32m|[m[1;36m\[m [1;31m\[m [1;31m\[m [35m\[m  
[1;36m|[m [32m|[m[1;36m/[m [1;31m/[m [1;31m/[m [35m/[m  
[1;36m|[m[1;36m/[m[32m|[m [1;31m|[m [1;31m|[m [35m|[m   
* [32m|[m [1;31m|[m [1;31m|[m [35m|[m   0f8e358 2015-06-29 Merge pull request #452 from github/fix-batch-test (Scott Barron[32m[m)
[34m|[m[35m\[m [32m\[m [1;31m\[m [1;31m\[m [35m\[m  
[34m|[m * [32m|[m [1;31m|[m [1;31m|[m [35m|[m 1dbc5a1 2015-06-29 Isolate init tests (rubyist[32m[m)
[34m|[m * [32m|[m [1;31m|[m [1;31m|[m [35m|[m 3a03aa6 2015-06-29 Update test server to support operation field in batch request (rubyist[32m[m)
[34m|[m[34m/[m [32m/[m [1;31m/[m [1;31m/[m [35m/[m  
[34m|[m * [1;31m|[m [1;31m|[m [35m|[m 84f4ba4 2015-06-29 Don't Sprintf if there are no args (rubyist[32m[m)
[34m|[m[34m/[m [1;31m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [1;31m|[m [35m|[m   9783bff 2015-06-29 Merge pull request #415 from github/401batch (Scott Barron[32m[m)
[36m|[m[1;31m\[m [1;31m\[m [1;31m\[m [35m\[m  
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m 8f14f36 2015-06-24 Update docs, use the correct key when setting config (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m 697fe6a 2015-06-24 Scope the lfs.access setting to the endpoint url (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m c705747 2015-06-24 Call the variable its original name (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m 978d397 2015-06-24 Don't need an extra struct here (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m 916fad6 2015-06-24 Doc typo (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m c376065 2015-06-24 Shutdown lfstest-gitserver before removing trash dir (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m f43441e 2015-06-24 Add some code docs (rubyist[32m[m)
[36m|[m * [1;31m|[m [1;31m|[m [35m|[m   07234e0 2015-06-24 Merge branch 'master' into 401batch (rubyist[32m[m)
[36m|[m [1;32m|[m[1;33m\[m [1;31m\[m [1;31m\[m [35m\[m  
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m aaea561 2015-06-24 Simplify checking code (rubyist[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m d55a377 2015-06-24 Just re-run the request, don't call the function again (rubyist[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m f109ded 2015-06-24 Add operation to batch payload (rubyist[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 9bda6e2 2015-06-23 Load local config, set access to private for 401, rerun Batch() (rubyist[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 86539d4 2015-06-18 Step two: Use credentials if access has been marked private (rubyist[32m[m)
[36m|[m * [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 647e719 2015-06-18 Step one: auth-less batch api (rubyist[32m[m)
* [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m   a2091fb 2015-06-29 Merge pull request #421 from github/test-with-built-lfs (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;33m\[m [1;31m\[m [1;31m\[m [35m\[m  
[1;34m|[m * [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 4e032fb 2015-06-19 allow a custom rootdir (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 2d7cf0e 2015-06-19 add the git version (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m df2eb7c 2015-06-19 teach script/integration how to run against a specific git-lfs build (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m   8cfb491 2015-06-29 Merge pull request #442 from sinbad/307-relative-location (Scott Barron[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;33m\[m [1;31m\[m [1;31m\[m [35m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 1a363df 2015-06-29 Add test for 307 redirects (Steve Streeting[32m[m)
[1;36m|[m * [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [1;31m|[m [35m|[m 4c90ff4 2015-06-26 Support 307 redirects using relative Location headers (Steve Streeting[32m[m)
[1;36m|[m [1;31m|[m [1;35m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;31m|[m[1;31m/[m [35m/[m  
[1;36m|[m [1;31m|[m[1;31m/[m[1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [35m|[m   
* [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [35m|[m   f56893c 2015-06-29 Merge pull request #446 from github/init-tests (Scott Barron[32m[m)
[32m|[m[33m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m [1;31m\[m [35m\[m  
[32m|[m * [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [35m|[m 7955046 2015-06-28 add some basic init tests (Rick Olson[32m[m)
[32m|[m [1;31m|[m[1;31m/[m [1;35m/[m [1;33m/[m [1;33m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [1;31m|[m [35m|[m   fc319de 2015-06-29 Merge pull request #450 from github/atomic-blast (Scott Barron[32m[m)
[1;31m|[m[35m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m [1;31m\[m [35m\[m  
[1;31m|[m [35m|[m[1;31m_[m[1;31m|[m[1;31m_[m[1;35m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;33m|[m[1;31m/[m [35m/[m  
[1;31m|[m[1;31m/[m[35m|[m [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m   
[1;31m|[m * [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m cfa3ad0 2015-06-29 Use int32 for these counters to avoid alignment bugs (rubyist[32m[m)
[1;31m|[m[1;31m/[m [1;31m/[m [1;35m/[m [1;33m/[m [1;33m/[m [35m/[m  
* [1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m   fb38e88 2015-06-27 Merge pull request #441 from devcurmudgeon/devcurmudgeon-fix-api-doc-link (risk danger olson[32m[m)
[1;31m|[m[1;31m\[m [1;31m\[m [1;35m\[m [1;33m\[m [1;33m\[m [35m\[m  
[1;31m|[m [1;31m|[m[1;31m/[m [1;35m/[m [1;33m/[m [1;33m/[m [35m/[m  
[1;31m|[m[1;31m/[m[1;31m|[m [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m   
[1;31m|[m * [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m b7992c8 2015-06-26 Fix broken link to api/README.md (Paul Sherwood[32m[m)
[1;31m|[m[1;31m/[m [1;35m/[m [1;33m/[m [1;33m/[m [35m/[m  
* [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m e2e04ac 2015-06-24 fix typo (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [1;33m|[m [35m|[m   8fafe93 2015-06-24 Merge pull request #428 from andyneff/centos_fixup (risk danger olson[32m[m)
[1;33m|[m[1;33m\[m [1;35m\[m [1;33m\[m [1;33m\[m [35m\[m  
[1;33m|[m [1;33m|[m[1;33m_[m[1;35m|[m[1;33m_[m[1;33m|[m[1;33m/[m [35m/[m  
[1;33m|[m[1;33m/[m[1;33m|[m [1;35m|[m [1;33m|[m [35m|[m   
[1;33m|[m * [1;35m|[m [1;33m|[m [35m|[m 1e60e72 2015-06-20 Fixed many errors in INSTALL.md (Andy Neff[32m[m)
[1;33m|[m * [1;35m|[m [1;33m|[m [35m|[m e2f770c 2015-06-19 Added compatibility between build-centos and build_rpms scripts (Andy Neff[32m[m)
[1;33m|[m * [1;35m|[m [1;33m|[m [35m|[m f8aa369 2015-06-19 Minor changes/fixes to CentOS building (Andy Neff[32m[m)
[1;33m|[m * [1;35m|[m [1;33m|[m [35m|[m acd6092 2015-06-18 Minor tweaks to build (Andy Neff[32m[m)
* [1;33m|[m [1;35m|[m [1;33m|[m [35m|[m   a0e23b4 2015-06-23 Merge pull request #429 from github/fetchbug (Scott Barron[32m[m)
[35m|[m[1;35m\[m [1;33m\[m [1;35m\[m [1;33m\[m [35m\[m  
[35m|[m [1;35m|[m[35m_[m[1;33m|[m[35m_[m[1;35m|[m[35m_[m[1;33m|[m[35m/[m  
[35m|[m[35m/[m[1;35m|[m [1;33m|[m [1;35m|[m [1;33m|[m   
[35m|[m * [1;33m|[m [1;35m|[m [1;33m|[m fe6f541 2015-06-23 Use a WaitGroup to be more explicit that we're waiting for a goroutine to finish and not waiting on any data (rubyist[32m[m)
[35m|[m * [1;33m|[m [1;35m|[m [1;33m|[m 0486305 2015-06-20 Fix race condition in fetch, add integration tests (rubyist[32m[m)
[35m|[m[35m/[m [1;33m/[m [1;35m/[m [1;33m/[m  
* [1;33m|[m [1;35m|[m [1;33m|[m   b351a8b 2015-06-19 Merge pull request #420 from github/ship-v0.5.2 (risk danger olson[32m (tag: v0.5.2)[m)
[1;36m|[m[31m\[m [1;33m\[m [1;35m\[m [1;33m\[m  
[1;36m|[m * [1;33m|[m [1;35m|[m [1;33m|[m eb2743c 2015-06-19 bump version (Rick Olson[32m[m)
[1;36m|[m * [1;33m|[m [1;35m|[m [1;33m|[m bb4c09e 2015-06-19 update changelog (Rick Olson[32m[m)
* [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m   a5b832a 2015-06-19 Merge pull request #424 from github/fix-logs (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;33m\[m [1;35m\[m [1;33m\[m  
[32m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m 714a061 2015-06-19 use append since its simpler (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m a85ab75 2015-06-19 can't modify the names array if its len is 0 (Rick Olson[32m[m)
[32m|[m[32m/[m [31m/[m [1;33m/[m [1;35m/[m [1;33m/[m  
* [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m   8782fc9 2015-06-19 Merge pull request #423 from github/progressfix (Scott Barron[32m[m)
[34m|[m[35m\[m [31m\[m [1;33m\[m [1;35m\[m [1;33m\[m  
[34m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m c8ea92d 2015-06-19 Don't close in shutdown, the caller still closes (rubyist[32m[m)
[34m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m 2f0ed97 2015-06-19 Check the log instead of the flag so close can still work after an error (rubyist[32m[m)
[34m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m b45132a 2015-06-19 Fix embarassing omission, return the proper amount of things (rubyist[32m[m)
[34m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m 7a961fe 2015-06-19 Combine Write()/Sync(), handle write errors, add some docs (rubyist[32m[m)
[34m|[m * [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m   ac7f2f0 2015-06-19 Merge branch 'master' into progressfix (rubyist[32m[m)
[34m|[m [36m|[m[34m\[m [31m\[m [1;33m\[m [1;35m\[m [1;33m\[m  
[34m|[m [36m|[m[34m/[m [31m/[m [1;33m/[m [1;35m/[m [1;33m/[m  
[34m|[m[34m/[m[36m|[m [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m   
* [36m|[m [31m|[m [1;33m|[m [1;35m|[m [1;33m|[m   f771cbc 2015-06-19 Merge pull request #422 from github/objectfix (risk danger olson[32m[m)
[1;35m|[m[1;33m\[m [36m\[m [31m\[m [1;33m\[m [1;35m\[m [1;33m\[m  
[1;35m|[m [1;33m|[m[1;35m_[m[36m|[m[1;35m_[m[31m|[m[1;35m_[m[1;33m|[m[1;35m/[m [1;33m/[m  
[1;35m|[m[1;35m/[m[1;33m|[m [36m|[m [31m|[m [1;33m|[m [1;33m|[m   
[1;35m|[m * [36m|[m [31m|[m [1;33m|[m [1;33m|[m ac6b245 2015-06-19 ensure the oid/size are set (rubyist[32m[m)
[1;35m|[m[1;35m/[m [36m/[m [31m/[m [1;33m/[m [1;33m/[m  
* [36m|[m [31m|[m [1;33m|[m [1;33m|[m 973649e 2015-06-19 period (risk danger olson[32m[m)
[31m|[m [36m|[m[31m/[m [1;33m/[m [1;33m/[m  
[31m|[m[31m/[m[36m|[m [1;33m|[m [1;33m|[m   
* [36m|[m [1;33m|[m [1;33m|[m   6ca3d37 2015-06-19 Merge pull request #413 from github/reorganize-api-docs (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [36m\[m [1;33m\[m [1;33m\[m  
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m e40793a 2015-06-19 update batch api response codes (risk danger olson[32m[m)
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m b6530ce 2015-06-19 Update README.md (risk danger olson[32m[m)
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m f050759 2015-06-18 remove the oid from the ssh command args (Rick Olson[32m[m)
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m 99ce648 2015-06-18 add the api readme (Rick Olson[32m[m)
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m 9837c6d 2015-06-18 add docs for the v1 http api (Rick Olson[32m[m)
[1;34m|[m * [36m|[m [1;33m|[m [1;33m|[m 96bc2b5 2015-06-18 add batch api docs (Rick Olson[32m[m)
[1;34m|[m [1;33m|[m [36m|[m[1;33m_[m[1;33m|[m[1;33m/[m  
[1;34m|[m [1;33m|[m[1;33m/[m[36m|[m [1;33m|[m   
* [1;33m|[m [36m|[m [1;33m|[m   f0f2d86 2015-06-19 Merge pull request #412 from github/update-root-docs (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;33m\[m [36m\[m [1;33m\[m  
[1;36m|[m * [1;33m|[m [36m|[m [1;33m|[m 92cb88e 2015-06-19 link roadmap issue (risk danger olson[32m[m)
[1;36m|[m * [1;33m|[m [36m|[m [1;33m|[m f03b070 2015-06-19 minor change (risk danger olson[32m[m)
[1;36m|[m * [1;33m|[m [36m|[m [1;33m|[m 59a49b0 2015-06-18 streamline the readme/roadmap, add first principles to contributing (Rick Olson[32m[m)
[1;36m|[m [1;33m|[m[1;33m/[m [36m/[m [1;33m/[m  
* [1;33m|[m [36m|[m [1;33m|[m   c7fb140 2015-06-19 Merge pull request #408 from github/endpoints-with-url-auth (risk danger olson[32m[m)
[1;33m|[m[33m\[m [1;33m\[m [36m\[m [1;33m\[m  
[1;33m|[m [33m|[m[1;33m/[m [36m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[33m|[m [36m|[m [1;33m|[m   
[1;33m|[m * [36m|[m [1;33m|[m ee62503 2015-06-17 const formatting nitpick (Rick Olson[32m[m)
[1;33m|[m * [36m|[m [1;33m|[m 9f54dad 2015-06-17 set request auth from optional endpoint url username/password (Rick Olson[32m[m)
[1;33m|[m [33m|[m * [1;33m|[m 4c9edba 2015-06-19 Support concurrent transfers when outputting progress for desktop apps (rubyist[32m[m)
[1;33m|[m [33m|[m[1;33m/[m [1;33m/[m  
[1;33m|[m[1;33m/[m[33m|[m [1;33m|[m   
* [33m|[m [1;33m|[m   80564d3 2015-06-18 Merge pull request #406 from github/env-with-version (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;33m\[m  
[34m|[m * [33m|[m [1;33m|[m 11d6130 2015-06-18 Updated env tests (Michael Käufl[32m[m)
[34m|[m * [33m|[m [1;33m|[m 9a2dc26 2015-06-17 add the git lfs and git version to the top of 'git lfs env' (Rick Olson[32m[m)
* [35m|[m [33m|[m [1;33m|[m 6cf2619 2015-06-18 slight fsck man page update (Rick Olson[32m[m)
* [35m|[m [33m|[m [1;33m|[m   955a06d 2015-06-18 Merge pull request #373 from michael-k/fsck (risk danger olson[32m[m)
[1;33m|[m[1;31m\[m [35m\[m [33m\[m [1;33m\[m  
[1;33m|[m [1;31m|[m[1;33m_[m[35m|[m[1;33m_[m[33m|[m[1;33m/[m  
[1;33m|[m[1;33m/[m[1;31m|[m [35m|[m [33m|[m   
[1;33m|[m * [35m|[m [33m|[m 3bd4a68 2015-06-18 Provided a rudimentary man page for git-lfs-fsck (Michael Käufl[32m[m)
[1;33m|[m * [35m|[m [33m|[m 22365d7 2015-06-18 Switched fsck tests to new test environment (Michael Käufl[32m[m)
[1;33m|[m * [35m|[m [33m|[m 7a22db4 2015-06-07 Move corrupt to a separate directory instead of deleting them (Michael Käufl[32m[m)
[1;33m|[m * [35m|[m [33m|[m ad1f2e9 2015-06-07 Used lfs.LocalMediaDir instead of lfs.LocalGitDir (Michael Käufl[32m[m)
[1;33m|[m * [35m|[m [33m|[m b8e594d 2015-06-07 manually write the object files since the clean filter probably isn't running in ci (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m c802222 2015-06-07 store the scanned pointers in a map to remove dupes (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m 7cfaa63 2015-06-07 add fsck --dry-run (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m 97ea7da 2015-06-07 Do not panic if an lfs object does not exist (Michael Käufl[32m[m)
[1;33m|[m * [35m|[m [33m|[m e209ddf 2015-06-07 delete corrupt lfs objects so they re-download on the next checkout (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m 0fb75ed 2015-06-07 write fsck results to stdout instead of logging an error (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m c6ee2e2 2015-06-07 actually test the fsck command (Rick Olson[32m[m)
[1;33m|[m * [35m|[m [33m|[m 58af282 2015-06-07 Add LFS fsck command (zeroshirts[32m[m)
* [1;31m|[m [35m|[m [33m|[m   92114fc 2015-06-17 Merge branch 'thpang67-master' (Rick Olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [35m\[m [33m\[m  
[1;32m|[m * [1;31m|[m [35m|[m [33m|[m 4eb1660 2015-06-17 dont export the file mode constants, not needed by lfs package users (Rick Olson[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [33m|[m   fed78fd 2015-06-17 Merge branch 'master' of https://github.com/thpang67/git-lfs into thpang67-master (Rick Olson[32m[m)
[1;32m|[m [1;32m|[m[1;35m\[m [1;31m\[m [35m\[m [33m\[m  
[1;32m|[m[1;32m/[m [1;35m/[m [1;31m/[m [35m/[m [33m/[m  
[1;32m|[m * [1;31m|[m [35m|[m [33m|[m 250e374 2015-06-04 Update lfs.go (Thomas S. Pangborn[32m[m)
[1;32m|[m * [1;31m|[m [35m|[m [33m|[m   1177462 2015-06-03 Merge pull request #1 from thpang67/thpang67-patch-1 (Thomas S. Pangborn[32m[m)
[1;32m|[m [1;36m|[m[31m\[m [1;31m\[m [35m\[m [33m\[m  
[1;32m|[m [1;36m|[m * [1;31m|[m [35m|[m [33m|[m 03a7590 2015-06-02 Update lfs.go (Thomas S. Pangborn[32m[m)
[1;32m|[m [1;36m|[m[1;36m/[m [1;31m/[m [35m/[m [33m/[m  
* [1;36m|[m [1;31m|[m [35m|[m [33m|[m   61d0bc7 2015-06-17 Merge pull request #383 from ssgelm/fix-deb (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [1;31m\[m [35m\[m [33m\[m  
[32m|[m * [1;36m|[m [1;31m|[m [35m|[m [33m|[m 5c3de3e 2015-06-12 Fix the deb package build and bump the version to 0.5.2 (Stephen Gelman[32m[m)
* [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m   3ba9307 2015-06-17 Merge pull request #393 from sinbad/travis-on-forks (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m [33m\[m  
[34m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m 4fd1d0e 2015-06-15 Make it possible to enable Travis on forks (Steve Streeting[32m[m)
[34m|[m [33m|[m[33m/[m [1;36m/[m [1;31m/[m [35m/[m [33m/[m  
* [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m   38a1caf 2015-06-17 Merge pull request #332 from andyneff/centos_rpms (risk danger olson[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m [33m\[m  
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m bc2d929 2015-05-29 Added tests to building rpms (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m adca28e 2015-05-29 Fixes to handle #330/#331 (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m 0be2781 2015-05-28 Added INSTALL.md (Andy Neff[32m[m)
[36m|[m * [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m 4a71627 2015-05-28 RPMs for Centos 5/6/7 (Andy Neff[32m[m)
* [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m [33m|[m   f1350be 2015-06-17 Merge pull request #313 from andyneff/andyneff_centos_ruby (risk danger olson[32m[m)
[33m|[m[1;33m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m [33m\[m  
[33m|[m [1;33m|[m[33m_[m[1;31m|[m[33m_[m[33m|[m[33m_[m[1;36m|[m[33m_[m[1;31m|[m[33m_[m[35m|[m[33m/[m  
[33m|[m[33m/[m[1;33m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 567032e 2015-05-29 Fixes to handle #330/#331 (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 53979f7 2015-05-29 Changed check for git-lfs to look just for bootstrap (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m caaf7a5 2015-05-29 Simpler support for CentOS 5 (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 1a82ce2 2015-05-29 CentOS 5 uses EPEL SRPM to build instead of download binary (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m b769b35 2015-05-29 Moved to sf (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 2a86523 2015-05-29 Now successfully builds on CentOS 5 (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m ae7482b 2015-05-29 Added CentOS 5 support (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m 01a976c 2015-05-29 Uses current dir if its looks like an archive (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m e329c84 2015-05-29 Uses current dir if its a git repo as mentioned in #299 (Andy Neff[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m b8acf77 2015-05-29 Install ruby on CentOS 6 (Andy Neff[32m[m)
[33m|[m [1;31m|[m[1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [35m/[m  
* [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [35m|[m   29643c4 2015-06-17 Merge pull request #403 from sinbad/endpoint_consistency (risk danger olson[32m[m)
[35m|[m[1;35m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [35m\[m  
[35m|[m [1;35m|[m[35m_[m[1;31m|[m[35m_[m[33m|[m[35m_[m[1;36m|[m[35m_[m[1;31m|[m[35m/[m  
[35m|[m[35m/[m[1;35m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m   
[35m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m 7342df5 2015-06-17 Consistently construct Endpoint from urls the same way (Steve Streeting[32m[m)
* [1;35m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m   34b9ef3 2015-06-17 Merge pull request #382 from bozaro/git-version (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m  
[1;36m|[m * [1;35m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m f3c116b 2015-06-11 Remove useless fork per every file for "git version" Because Ms Windows has very expensive CreateProcess this operation takes too long time on large file count. (a.navrotskiy[32m[m)
[1;36m|[m [33m|[m [1;35m|[m[33m_[m[1;31m|[m[33m/[m [1;36m/[m [1;31m/[m  
[1;36m|[m [33m|[m[33m/[m[1;35m|[m [1;31m|[m [1;36m|[m [1;31m|[m   
* [33m|[m [1;35m|[m [1;31m|[m [1;36m|[m [1;31m|[m   07c3844 2015-06-17 Merge pull request #381 from bozaro/blinking (risk danger olson[32m[m)
[32m|[m[33m\[m [33m\[m [1;35m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[32m|[m * [33m|[m [1;35m|[m [1;31m|[m [1;36m|[m [1;31m|[m ab20cae 2015-06-17 Windows: hide git application window. (Artem V. Navrotskiy[32m[m)
[32m|[m [1;35m|[m [33m|[m[1;35m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[32m|[m [1;35m|[m[1;35m/[m[33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   
* [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   a444aaa 2015-06-17 Merge pull request #370 from michael-k/env (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[34m|[m * [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 5bd7026 2015-06-17 Added tests for env vars GIT_DIR and GIT_WORK_TREE (Michael Käufl[32m[m)
[34m|[m * [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m f0a9fa3 2015-06-17 Added support for relative paths in environment variable GIT_WORK_TREE (Michael Käufl[32m[m)
[34m|[m * [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 646e502 2015-06-17 Added support for environment variables GIT_DIR and GIT_WORK_TREE (Michael Käufl[32m[m)
[34m|[m [1;35m|[m[1;35m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
* [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   b27296a 2015-06-17 Merge pull request #405 from michael-k/cleanup (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;35m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[36m|[m * [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 56cba97 2015-06-01 Cleaned up old integration test leftovers (Michael Käufl[32m[m)
* [1;31m|[m [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   3cd402e 2015-06-17 Merge pull request #401 from michael-k/boolEnv (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;35m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[1;32m|[m * [1;31m|[m [1;35m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 410830a 2015-06-17 Parse boolean environment variables (Michael Käufl[32m[m)
[1;32m|[m [1;35m|[m [1;31m|[m[1;35m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[1;32m|[m [1;35m|[m[1;35m/[m[1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   
* [1;35m|[m [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   757200a 2015-06-17 Merge pull request #402 from sinbad/callback_doc (risk danger olson[32m[m)
[1;35m|[m[1;35m\[m [1;35m\[m [1;31m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[1;35m|[m [1;35m|[m[1;35m/[m [1;31m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;35m|[m [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   
[1;35m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m fd2fc10 2015-06-17 Document the arguments to CopyCallback (Steve Streeting[32m[m)
[1;35m|[m[1;35m/[m [1;31m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
* [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   968faba 2015-06-16 Merge pull request #400 from github/statstweak (Scott Barron[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[1;36m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m f252577 2015-06-16 Build object outside of lock, pull URL from response (rubyist[32m[m)
[1;36m|[m[1;36m/[m [1;31m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
* [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m   0fc2f0b 2015-06-16 Merge pull request #399 from github/michael-k-cleanupConfig (risk danger olson[32m[m)
[32m|[m[33m\[m [1;31m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[32m|[m * [1;31m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m   e19ce8b 2015-06-16 merge master (Rick Olson[32m[m)
[32m|[m [32m|[m[35m\[m [1;31m\[m [33m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[32m|[m[32m/[m [35m/[m [1;31m/[m [33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 790d2b2 2015-06-14 Check environment variables only once per git-lfs process (Michael Käufl[32m[m)
[32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m d3b9eaa 2015-06-14 Followed convention of grouping mutex with guarded stuff (Michael Käufl[32m[m)
[32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 191e071 2015-06-14 Extracted function (Michael Käufl[32m[m)
[32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 81cec75 2015-06-14 Moved ObjectUrl's next to each other (Michael Käufl[32m[m)
[32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;36m|[m [1;31m|[m 63e713c 2015-06-14 Removed unused struct (Michael Käufl[32m[m)
[32m|[m [33m|[m [1;31m|[m[33m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[32m|[m [33m|[m[33m/[m[1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m   
* [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m b756d56 2015-06-16 ensure untrack tests operate in trashdir sub dirs (Rick Olson[32m[m)
* [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m   8ae4dac 2015-06-16 Merge pull request #398 from michael-k/untrack (Scott Barron[32m[m)
[36m|[m[1;31m\[m [33m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 92cb8eb 2015-06-16 Extracted function (Michael Käufl[32m[m)
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 26dfdc4 2015-06-16 Break out of loop (Michael Käufl[32m[m)
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m b83fe22 2015-06-16 Used defer to close the attributes file (Michael Käufl[32m[m)
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 7f9275d 2015-06-16 Don't remove lines without `filter=lfs` (Michael Käufl[32m[m)
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 1db4857 2015-06-16 Only untrack if we're in a repo's work tree (Michael Käufl[32m[m)
[36m|[m * [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 47bf091 2015-06-16 Added failing untrack tests (Michael Käufl[32m[m)
* [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m   257f909 2015-06-16 Merge pull request #366 from github/statistician (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [33m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[1;32m|[m * [1;31m\[m [33m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;31m\[m   55b6b6f 2015-06-15 Merge branch 'master' into statistician (rubyist[32m[m)
[1;32m|[m [1;34m|[m[1;31m\[m [1;31m\[m [33m\[m [1;31m\[m [1;31m\[m [1;36m\[m [1;31m\[m  
[1;32m|[m [1;34m|[m [1;31m|[m[1;31m/[m [33m/[m [1;31m/[m [1;31m/[m [1;36m/[m [1;31m/[m  
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 61bc0b0 2015-06-05 Logs commands should skip directories (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 0e29183 2015-06-05 Log to an http subdir to not interfere with `git lfs logs` (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m f710bfd 2015-06-05 Log the URL with the transfer (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 08dc45c 2015-06-05 Put some meta info in the first line of the log (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 50f9742 2015-06-05 Calculate the response header size sooner and outside of the mutex lock (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m d4f70f4 2015-06-04 Dump http stats into a k/v format log file (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 3fbc2c0 2015-06-04 Add some locking around stats data structures (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 161e2fd 2015-06-04 Use a different env var to flip on http stats logging (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m eb686c4 2015-06-03 fractions of a byte! (rubyist[32m[m)
[1;32m|[m * [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;36m|[m [1;31m|[m 6fafade 2015-06-03 Modify the http tracing code to collect stats on all http transfers (rubyist[32m[m)
[1;32m|[m [1;36m|[m [1;31m|[m[1;36m_[m[33m|[m[1;36m_[m[1;31m|[m[1;36m_[m[1;31m|[m[1;36m/[m [1;31m/[m  
[1;32m|[m [1;36m|[m[1;36m/[m[1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   
* [1;36m|[m [1;31m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   8b7e650 2015-06-15 Merge pull request #396 from michael-k/race-condition (Scott Barron[32m[m)
[1;31m|[m[31m\[m [1;36m\[m [1;31m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[1;31m|[m [31m|[m[1;31m_[m[1;36m|[m[1;31m/[m [33m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
[1;31m|[m[1;31m/[m[31m|[m [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   
[1;31m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m ab77185 2015-06-16 Resolved race condition (Michael Käufl[32m[m)
[1;31m|[m[1;31m/[m [1;36m/[m [33m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
* [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 0f544e5 2015-06-15 Fix simpleExec args (rubyist[32m[m)
* [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   0e80e80 2015-06-15 Merge pull request #358 from github/fail-batch-gracefully (Scott Barron[32m[m)
[32m|[m[33m\[m [1;36m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[32m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 8a04052 2015-06-04 When automatically turning off batch, use .git/config instead of .gitconfig (rubyist[32m[m)
[32m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 636a167 2015-06-02 Int. tests for disabling batch when server doesn't support it (rubyist[32m[m)
[32m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 76d267a 2015-06-02 If batch transfer is enabled but not supported by the server, turn it off (rubyist[32m[m)
[32m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 316533b 2015-06-02 Update tracerx (rubyist[32m[m)
[32m|[m * [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 34c3078 2015-06-01 If the batch operation fails, fall back to individual (rubyist[32m[m)
* [33m|[m [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   509651a 2015-06-15 Merge branch 'bozaro-non-exists-config' (rubyist[32m[m)
[34m|[m[35m\[m [33m\[m [1;36m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[34m|[m * [33m\[m [1;36m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m   0dbe235 2015-06-15 Merge branch 'non-exists-config' of https://github.com/bozaro/git-lfs into bozaro-non-exists-config (rubyist[32m[m)
[34m|[m [34m|[m[1;31m\[m [33m\[m [1;36m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[34m|[m[34m/[m [1;31m/[m [33m/[m [1;36m/[m [33m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
[34m|[m * [33m|[m [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 8586140 2015-06-15 Do not read the configuration from nonexistent file (remove redundant git call) (a.navrotskiy[32m[m)
* [1;31m|[m [33m|[m [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m   62352a1 2015-06-15 Merge pull request #384 from michael-k/defer-in-loop (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [33m\[m [1;36m\[m [33m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[1;32m|[m * [1;31m|[m [33m|[m [1;36m|[m [33m|[m [1;31m|[m [1;31m|[m [1;31m|[m 9afb570 2015-06-12 Do not defer close in loop (Michael Käufl[32m[m)
[1;32m|[m [33m|[m [1;31m|[m[33m_[m[33m|[m[33m_[m[1;36m|[m[33m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
[1;32m|[m [33m|[m[33m/[m[1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m   
* [33m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m   74b4dfa 2015-06-15 Merge pull request #392 from michael-k/cleanupArgs (Scott Barron[32m[m)
[33m|[m[1;35m\[m [33m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[33m|[m [1;35m|[m[33m/[m [1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
[33m|[m[33m/[m[1;35m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m   
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m bed4653 2015-06-15 Unified naming of varargs (Michael Käufl[32m[m)
[33m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m 3a46210 2015-06-15 Removed unused function arguments (Michael Käufl[32m[m)
[33m|[m[33m/[m [1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [1;31m/[m [1;31m/[m  
* [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m   296dddf 2015-06-10 Merge pull request #377 from bozaro/more-fast-filter (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[1;36m|[m * [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m ee0c877 2015-06-10 Remove unnecessary exec the process by filter wrapper This change improve git-lfs performance on Windows (Artem V. Navrotskiy[32m[m)
* [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m   c9c3be8 2015-06-08 Merge pull request #374 from michael-k/spec (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [1;31m\[m [1;31m\[m  
[32m|[m * [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m [1;31m|[m 1643608 2015-06-08 Fixed some (markdown) issues in spec (Michael Käufl[32m[m)
[32m|[m [1;31m|[m [31m|[m[1;31m_[m[1;31m|[m[1;31m_[m[33m|[m[1;31m_[m[1;36m|[m[1;31m_[m[1;31m|[m[1;31m_[m[1;31m|[m[1;31m/[m  
[32m|[m [1;31m|[m[1;31m/[m[31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m   
* [1;31m|[m [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m   295a709 2015-06-08 Merge pull request #372 from michael-k/fail-tests-on-build-fail (risk danger olson[32m[m)
[1;31m|[m[35m\[m [1;31m\[m [31m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [1;31m\[m  
[1;31m|[m [35m|[m[1;31m/[m [31m/[m [1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [1;31m/[m  
[1;31m|[m[1;31m/[m[35m|[m [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m   
[1;31m|[m * [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m 2dd5e68 2015-06-07 Return a non-zero exit status if build fails (Michael Käufl[32m[m)
[1;31m|[m[1;31m/[m [31m/[m [1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [1;31m/[m  
* [31m|[m [1;31m|[m [33m|[m [1;36m|[m [1;31m|[m [1;31m|[m   afbc581 2015-06-03 Merge pull request #361 from michael-k/config (risk danger olson[32m[m)
[1;36m|[m[1;31m\[m [31m\[m [1;31m\[m [33m\[m [1;36m\[m [1;31m\[m [1;31m\[m  
[1;36m|[m [1;31m|[m [31m|[m[1;31m/[m [33m/[m [1;36m/[m [1;31m/[m [1;31m/[m  
[1;36m|[m [1;31m|[m[1;31m/[m[31m|[m [33m|[m [1;36m/[m [1;31m/[m [1;31m/[m   
[1;36m|[m [1;31m|[m[1;36m_[m[31m|[m[1;36m_[m[33m|[m[1;36m/[m [1;31m/[m [1;31m/[m    
[1;36m|[m[1;36m/[m[1;31m|[m [31m|[m [33m|[m [1;31m|[m [1;31m|[m     
[1;36m|[m * [31m|[m [33m|[m [1;31m|[m [1;31m|[m 67c3484 2015-06-01 Used .gitconfig in LocalWorkingDir instead of PWD (Michael Käufl[32m[m)
[1;36m|[m * [31m|[m [33m|[m [1;31m|[m [1;31m|[m 6e0e7ab 2015-06-01 Added failing env test with .gitconfig file (Michael Käufl[32m[m)
[1;36m|[m * [31m|[m [33m|[m [1;31m|[m [1;31m|[m 1b55516 2015-06-01 Added missing env test in .git directory (Michael Käufl[32m[m)
[1;36m|[m * [31m|[m [33m|[m [1;31m|[m [1;31m|[m 0b5527d 2015-06-01 Fixed typo (Michael Käufl[32m[m)
[1;36m|[m [33m|[m [31m|[m[33m/[m [1;31m/[m [1;31m/[m  
[1;36m|[m [33m|[m[33m/[m[31m|[m [1;31m|[m [1;31m|[m   
* [33m|[m [31m|[m [1;31m|[m [1;31m|[m   27ea25e 2015-06-02 Merge pull request #364 from michael-k/fileModeBits (Scott Barron[32m[m)
[33m|[m[1;33m\[m [33m\[m [31m\[m [1;31m\[m [1;31m\[m  
[33m|[m [1;33m|[m[33m/[m [31m/[m [1;31m/[m [1;31m/[m  
[33m|[m[33m/[m[1;33m|[m [31m|[m [1;31m|[m [1;31m|[m   
[33m|[m * [31m|[m [1;31m|[m [1;31m|[m d95b2ce 2015-06-02 Used 0755 as file mode bits when creating directories (Michael Käufl[32m[m)
[33m|[m[33m/[m [31m/[m [1;31m/[m [1;31m/[m  
* [31m|[m [1;31m|[m [1;31m|[m   87264d0 2015-06-01 Merge pull request #356 from michael-k/track (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [1;31m\[m [1;31m\[m  
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 8beb5ea 2015-06-01 Added test for pure helper function (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 3145f2d 2015-06-01 Don't track files outside repository (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m ec756f6 2015-06-01 Extended test of tracking files outside the repo (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m ebaef3d 2015-06-01 Track the relative representation of a path (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m cb9ff2a 2015-06-01 Added a failing test for github/git-lfs#230 (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 6e87008 2015-06-01 Improved detection of duplicate paths (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m bf0e10a 2015-06-01 Added test concerning recognition of duplicate paths (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 213241b 2015-06-01 Removed unnecessary if statement (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 87622df 2015-06-01 If attributes file not readable, continue with next one (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m e0000cf 2015-06-01 Used a label to continue the outer loop (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m 5c42ca3 2015-06-01 Used defer to close the attributes file (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m a5cb344 2015-06-01 Made sure that .git/info/attributes is not a directory (Michael Käufl[32m[m)
[1;34m|[m * [31m|[m [1;31m|[m [1;31m|[m a381b54 2015-05-31 Rearranged track's helper functions (Michael Käufl[32m[m)
[1;34m|[m [1;31m|[m [31m|[m[1;31m/[m [1;31m/[m  
[1;34m|[m [1;31m|[m[1;31m/[m[31m|[m [1;31m|[m   
* [1;31m|[m [31m|[m [1;31m|[m   7d1defa 2015-06-01 Merge pull request #354 from michael-k/dotGit (Scott Barron[32m[m)
[1;31m|[m[31m\[m [1;31m\[m [31m\[m [1;31m\[m  
[1;31m|[m [31m|[m[1;31m/[m [31m/[m [1;31m/[m  
[1;31m|[m[1;31m/[m[31m|[m [31m|[m [1;31m|[m   
[1;31m|[m * [31m|[m [1;31m|[m 233b4c7 2015-05-30 Allowed track only if we're in a working dir (Michael Käufl[32m[m)
[1;31m|[m * [31m|[m [1;31m|[m 73d54e1 2015-05-30 Added failing test for github/git-lfs#353 (Michael Käufl[32m[m)
[1;31m|[m * [31m|[m [1;31m|[m 2c08cd7 2015-05-30 Added some comments (Michael Käufl[32m[m)
[1;31m|[m * [31m|[m [1;31m|[m b779f54 2015-05-30 Extracted detection of LocalWorkingDir from .git file processing (Michael Käufl[32m[m)
[1;31m|[m * [31m|[m [1;31m|[m d81345c 2015-05-30 Return early if the .git file doesn't tell us anything (Michael Käufl[32m[m)
[1;31m|[m * [31m|[m [1;31m|[m 42228cd 2015-05-30 Moved the special case in the if block (Michael Käufl[32m[m)
[1;31m|[m[1;31m/[m [31m/[m [1;31m/[m  
* [31m|[m [1;31m|[m   2bea23b 2015-05-29 Merge branch 'michael-k-workingDir' (Rick Olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;31m\[m  
[32m|[m * [31m|[m [1;31m|[m c8f7fa9 2015-05-29 fix the echo statement (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;31m|[m   e17aaf2 2015-05-29 Merge branch 'workingDir' of https://github.com/michael-k/git-lfs into michael-k-workingDir (Rick Olson[32m[m)
[32m|[m [32m|[m[35m\[m [31m\[m [1;31m\[m  
[32m|[m[32m/[m [35m/[m [31m/[m [1;31m/[m  
[32m|[m * [31m|[m [1;31m|[m a0dc285 2015-05-30 Fixed LocalWorkingDir for directories in submodules (Michael Käufl[32m[m)
[32m|[m * [31m|[m [1;31m|[m 04df6a4 2015-05-30 Added another env check in the submodules subdirectory (Michael Käufl[32m[m)
[32m|[m * [31m|[m [1;31m|[m 946e2d4 2015-05-30 Made submodule test more precise by checking for line endings (Michael Käufl[32m[m)
* [35m|[m [31m|[m [1;31m|[m   cd76905 2015-05-29 Merge pull request #352 from michael-k/defer (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [31m\[m [1;31m\[m  
[35m|[m [1;31m|[m[35m/[m [31m/[m [1;31m/[m  
[35m|[m[35m/[m[1;31m|[m [31m|[m [1;31m|[m   
[35m|[m * [31m|[m [1;31m|[m aaa9929 2015-05-30 Keep defer close (Michael Käufl[32m[m)
[35m|[m * [31m|[m [1;31m|[m 639beb6 2015-05-30 Only defer Close() if err is nil (Michael Käufl[32m[m)
[35m|[m[35m/[m [31m/[m [1;31m/[m  
* [31m|[m [1;31m|[m   3e0cc8b 2015-05-29 Merge pull request #346 from github/clean-empty-file (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [31m\[m [1;31m\[m  
[1;32m|[m * [31m|[m [1;31m|[m aa0c3c9 2015-05-28 fix issues with decoding a pointer file from an empty buffer (Rick Olson[32m[m)
[1;32m|[m * [31m|[m [1;31m|[m 077ec4c 2015-05-28 always send size, even if its zero (Rick Olson[32m[m)
[1;32m|[m [1;31m|[m [31m|[m[1;31m/[m  
[1;32m|[m [1;31m|[m[1;31m/[m[31m|[m   
* [1;31m|[m [31m|[m   d71bab6 2015-05-29 Merge pull request #349 from github/lfsfetch (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m [31m\[m  
[1;34m|[m * [1;31m|[m [31m|[m a57d47a 2015-05-29 Rename `get` to `fetch` (rubyist[32m[m)
[1;34m|[m[1;34m/[m [1;31m/[m [31m/[m  
* [1;31m|[m [31m|[m   858c595 2015-05-29 Merge pull request #348 from github/batchapitypo (Scott Barron[32m[m)
[1;31m|[m[31m\[m [1;31m\[m [31m\[m  
[1;31m|[m [31m|[m[1;31m/[m [31m/[m  
[1;31m|[m[1;31m/[m[31m|[m [31m|[m   
[1;31m|[m * [31m|[m 2ce0189 2015-05-29 The batch api example json is missing some braces (rubyist[32m[m)
[1;31m|[m[1;31m/[m [31m/[m  
* [31m|[m   47f7350 2015-05-28 Merge pull request #331 from github/nut (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m  
[32m|[m * [31m|[m 230026e 2015-05-28 no need to rewrite imports in comments (Rick Olson[32m[m)
[32m|[m * [31m|[m 1fa1f2e 2015-05-28 removed result of a bad master merge (Rick Olson[32m[m)
[32m|[m * [31m|[m   e2c7d0d 2015-05-28 merge master (Rick Olson[32m[m)
[32m|[m [34m|[m[32m\[m [31m\[m  
[32m|[m [34m|[m[32m/[m [31m/[m  
[32m|[m[32m/[m[34m|[m [31m|[m   
* [34m|[m [31m|[m   ac67b68 2015-05-28 Merge pull request #285 from github/multitransfer (risk danger olson[32m[m)
[36m|[m[1;31m\[m [34m\[m [31m\[m  
[36m|[m * [34m|[m [31m|[m 7df3999 2015-05-28 mention in the API spec that the batch api is still experimental (risk danger olson[32m[m)
[36m|[m * [34m|[m [31m|[m 2fbe6ea 2015-05-27 update test to use tee (Rick Olson[32m[m)
[36m|[m * [34m|[m [31m|[m efdcb06 2015-05-27 Remove unused Upload() function, fix upload tests to use new functions (rubyist[32m[m)
[36m|[m * [34m|[m [31m|[m 71ebdfd 2015-05-27 fix push and pre-push tests (Rick Olson[32m[m)
[36m|[m * [34m|[m [31m|[m e3173e2 2015-05-27 fix batch tests (Rick Olson[32m[m)
[36m|[m * [34m|[m [31m|[m 589592c 2015-05-27 fix env tests (Rick Olson[32m[m)
[36m|[m * [34m|[m [31m|[m   3eb8e24 2015-05-27 merge master (Rick Olson[32m[m)
[36m|[m [1;32m|[m[1;33m\[m [34m\[m [31m\[m  
[36m|[m * [1;33m|[m [34m|[m [31m|[m ebc81ae 2015-05-27 Run update-index as a single background process (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m d059f10 2015-05-27 Add some perf tracing to lfs get, run a single update-index process (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m b44af51 2015-05-27 Update tracerx so it can run perf tracing without regular tracing (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 36246c5 2015-05-27 Add the concurrent and batch settings to `lfs env` output (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 135d59b 2015-05-26 More robust config parsing for concurrent/batch vals (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 565928c 2015-05-26 Update some documentation (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 8ed808c 2015-05-26 Remove debugging print (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 8d2d826 2015-05-22 Update the index when replacing files in the wd with `get` (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m 740ba62 2015-05-21 Add an integration test that pushes with batch enabled (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m cc960a6 2015-05-21 Add batch endpoint to integration test server (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m e9646c0 2015-05-21 Update happy path output test for batched progress output (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m ee68f55 2015-05-21 The test server needs to give the oid and size back in the json (rubyist[32m[m)
[36m|[m * [1;33m|[m [34m|[m [31m|[m   22d3f4e 2015-05-21 Merge branch 'master' into multitransfer (rubyist[32m[m)
[36m|[m [1;34m|[m[1;35m\[m [1;33m\[m [34m\[m [31m\[m  
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m a3bb997 2015-05-21 Pull common queue code into a TransferQueue, reduce duplicated code (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m f8f4ad2 2015-05-21 Start a man page for lfs get (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m e445b45 2015-05-21 If the ref we 'get' is the ref we're on, we can write to the working directory (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 685e877 2015-05-21 Add some git code to resolve a ref to a sha (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m f905afa 2015-05-21 Make PointerSmudgeObject only download to the media directory (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 1d87b73 2015-05-18 Don't need to open if we're just using Stat() (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m de200c6 2015-05-18 Using structs here is unnecessary, a map will do (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 3bb2846 2015-05-14 Can batch downloads (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m c869fc6 2015-05-14 Basic "lfs get" that takes a ref, queues downloads (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m bf10519 2015-05-13 Naming consistency (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m f2ecad0 2015-05-13 ConcurrentUploads => ConcurrentTransfers (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m bf966e0 2015-05-13 Should be part of the queues (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m b4a4844 2015-05-13 Start a download queue (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 93381bf 2015-05-12 Batch takes objectResource instead of Uploadable (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m a870cc8 2015-05-07 Flip batch with config (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 52dac4f 2015-05-07 Initial rough draft of client code for the Batch endpoint and batch uploads (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 978471f 2015-05-05 accept and return an object with an array instead of an array (rubyist[32m[m)
[36m|[m * [1;35m|[m [1;33m|[m [34m|[m [31m|[m 694a6fd 2015-05-05 Draft an endpoint for batch upload/download operations (rubyist[32m[m)
[36m|[m [31m|[m [1;35m|[m[31m_[m[1;33m|[m[31m_[m[34m|[m[31m/[m  
[36m|[m [31m|[m[31m/[m[1;35m|[m [1;33m|[m [34m|[m   
* [31m|[m [1;35m|[m [1;33m|[m [34m|[m   5ebb953 2015-05-28 Merge pull request #339 from github/prepush (Scott Barron[32m[m)
[1;36m|[m[31m\[m [31m\[m [1;35m\[m [1;33m\[m [34m\[m  
[1;36m|[m * [31m|[m [1;35m|[m [1;33m|[m [34m|[m e65b992 2015-05-27 update shell tests for hook change (Rick Olson[32m[m)
[1;36m|[m * [31m|[m [1;35m|[m [1;33m|[m [34m|[m   f20d129 2015-05-27 merge master (Rick Olson[32m[m)
[1;36m|[m [32m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [34m\[m  
[1;36m|[m * [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m b363c56 2015-05-27 Make the pre-push hook check that git-lfs exists first (rubyist[32m[m)
* [33m|[m [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m   5e028d0 2015-05-28 Merge pull request #340 from github/contributing (Scott Barron[32m[m)
[33m|[m[35m\[m [33m\[m [33m\[m [31m\[m [1;35m\[m [1;33m\[m [34m\[m  
[33m|[m [35m|[m[33m_[m[33m|[m[33m/[m [31m/[m [1;35m/[m [1;33m/[m [34m/[m  
[33m|[m[33m/[m[35m|[m [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m   
[33m|[m * [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m 709c916 2015-05-27 Add a link to the CLA (rubyist[32m[m)
[33m|[m [33m|[m[33m/[m [31m/[m [1;35m/[m [1;33m/[m [34m/[m  
* [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m 4df64eb 2015-05-27 cat the readme if grep fails (Rick Olson[32m[m)
* [33m|[m [31m|[m [1;35m|[m [1;33m|[m [34m|[m 6dcf22d 2015-05-27 simpler grep usage (Rick Olson[32m[m)
[1;33m|[m [33m|[m[1;33m_[m[31m|[m[1;33m_[m[1;35m|[m[1;33m/[m [34m/[m  
[1;33m|[m[1;33m/[m[33m|[m [31m|[m [1;35m|[m [34m|[m   
[1;33m|[m [33m|[m [31m|[m [1;35m|[m * 80930f6 2015-05-28 update tracerx (Rick Olson[32m[m)
[1;33m|[m [33m|[m [31m|[m [1;35m|[m *   1598eda 2015-05-27 merge master (Rick Olson[32m[m)
[1;33m|[m [33m|[m [31m|[m [1;35m|[m [36m|[m[1;33m\[m  
[1;33m|[m [33m|[m[1;33m_[m[31m|[m[1;33m_[m[1;35m|[m[1;33m_[m[36m|[m[1;33m/[m  
[1;33m|[m[1;33m/[m[33m|[m [31m|[m [1;35m|[m [36m|[m   
* [33m|[m [31m|[m [1;35m|[m [36m|[m   fae85dc 2015-05-27 Merge pull request #342 from github/test-tweaks (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [31m\[m [1;35m\[m [36m\[m  
[1;32m|[m * [33m|[m [31m|[m [1;35m|[m [36m|[m 4422ddf 2015-05-27 dont bother changing the global git config anymore (Rick Olson[32m[m)
[1;32m|[m * [33m|[m [31m|[m [1;35m|[m [36m|[m 5ec74b9 2015-05-27 don't compile before running git tests, they dont use the compiled binary (Rick Olson[32m[m)
[1;32m|[m * [33m|[m [31m|[m [1;35m|[m [36m|[m 086d04b 2015-05-27 simplify happy-path tests a bit (Rick Olson[32m[m)
[1;32m|[m[1;32m/[m [33m/[m [31m/[m [1;35m/[m [36m/[m  
* [33m|[m [31m|[m [1;35m|[m [36m|[m   7f11548 2015-05-27 Merge pull request #336 from github/port-tests (risk danger olson[32m[m)
[33m|[m[1;35m\[m [33m\[m [31m\[m [1;35m\[m [36m\[m  
[33m|[m [1;35m|[m[33m/[m [31m/[m [1;35m/[m [36m/[m  
[33m|[m[33m/[m[1;35m|[m [31m|[m [1;35m|[m [36m|[m   
[33m|[m * [31m|[m [1;35m|[m [36m|[m 9cb9d86 2015-05-26 update env tests for different GIT_* env vars (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 271bb2e 2015-05-26 fix git env tests with submodules (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m e2b2716 2015-05-26 == is bash only, use = for string comparisons (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 8b54aee 2015-05-26 port clean, the last of the old tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m e92a3d1 2015-05-26 simplify happy-path tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 8207c91 2015-05-26 port update tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 663d58c 2015-05-26 port status tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 645dc9a 2015-05-26 port smudge tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m aa56320 2015-05-26 port push tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 736ed89 2015-05-26 port pre-push test (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 2b32063 2015-05-26 remove old pointer tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m df46ca3 2015-05-26 port pointer tests (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 6eb340b 2015-05-26 port ls-files test (Rick Olson[32m[m)
[33m|[m * [31m|[m [1;35m|[m [36m|[m 793d894 2015-05-26 port env tests (Rick Olson[32m[m)
[33m|[m[33m/[m [31m/[m [1;35m/[m [36m/[m  
[33m|[m [31m|[m [1;35m|[m * 85f4ed1 2015-05-26 bring script/bootstrap back? (Rick Olson[32m[m)
[33m|[m [31m|[m [1;35m|[m *   b21d912 2015-05-26 Merge branch 'master' into nut (Rick Olson[32m[m)
[33m|[m [31m|[m [1;35m|[m [1;36m|[m[33m\[m  
[33m|[m [31m|[m[33m_[m[1;35m|[m[33m_[m[1;36m|[m[33m/[m  
[33m|[m[33m/[m[31m|[m [1;35m|[m [1;36m|[m   
* [31m|[m [1;35m|[m [1;36m|[m   ca05b90 2015-05-26 Merge pull request #335 from github/new-home (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;35m\[m [1;36m\[m  
[32m|[m * [31m|[m [1;35m|[m [1;36m|[m 6063148 2015-05-26 set the user values (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;35m|[m [1;36m|[m 1feee58 2015-05-26 create a new HOME with a global git config just for tests (Rick Olson[32m[m)
[32m|[m[32m/[m [31m/[m [1;35m/[m [1;36m/[m  
[32m|[m [31m|[m [1;35m|[m * 1507424 2015-05-25 run script/integration first, which runs script/bootstrap (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * 06ce5aa 2015-05-25 unnecessary (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * 67e13fa 2015-05-25 update scripts (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * d678d4b 2015-05-25 update import paths (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * eab3892 2015-05-25 alphabetical (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * 6a80e12 2015-05-25 get assert fork which fixes issue with 'nut install' (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * ac2d1f2 2015-05-25 fix example package import (Rick Olson[32m[m)
[32m|[m [31m|[m [1;35m|[m * 9a68a80 2015-05-25 setup nut (Rick Olson[32m[m)
[32m|[m [31m|[m[32m_[m[1;35m|[m[32m/[m  
[32m|[m[32m/[m[31m|[m [1;35m|[m   
* [31m|[m [1;35m|[m   8ce6043 2015-05-25 Merge pull request #329 from joerg/fix-http-charset (Scott Barron[32m[m)
[34m|[m[35m\[m [31m\[m [1;35m\[m  
[34m|[m * [31m|[m [1;35m|[m 9526889 2015-05-25 Fix charset http Content-Type header (Jörg Herzinger[32m[m)
[34m|[m[34m/[m [31m/[m [1;35m/[m  
* [31m|[m [1;35m|[m   c8cc948 2015-05-22 Merge pull request #277 from ddanier/ddanier-ssh-href-doc-fix (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m [1;35m\[m  
[36m|[m * [31m|[m [1;35m|[m b0b4ecd 2015-05-05 No trailing slash for SSH API URLs (David Danier[32m[m)
[36m|[m * [31m|[m [1;35m|[m cb23f82 2015-05-02 Fixed api.md not containing the correct information about href (David Danier[32m[m)
[36m|[m [31m|[m[31m/[m [1;35m/[m  
* [31m|[m [1;35m|[m   3e9499e 2015-05-22 Merge pull request #327 from github/track-space (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [31m\[m [1;35m\[m  
[1;32m|[m * [31m|[m [1;35m|[m 7dc4ddf 2015-05-22 spaces (Rick Olson[32m[m)
[1;32m|[m * [31m|[m [1;35m|[m ec27bf7 2015-05-22 support tracking paths with spaces in the name (Rick Olson[32m[m)
[1;32m|[m[1;32m/[m [31m/[m [1;35m/[m  
* [31m|[m [1;35m|[m   9e22b23 2015-05-22 Merge pull request #323 from github/track-tests (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [31m\[m [1;35m\[m  
[1;34m|[m * [31m|[m [1;35m|[m f033ce9 2015-05-22 single grep (Rick Olson[32m[m)
[1;34m|[m * [31m|[m [1;35m|[m 7bbd60d 2015-05-21 update test to try adding the file twice (Rick Olson[32m[m)
[1;34m|[m * [31m|[m [1;35m|[m 4e27a32 2015-05-21 always use 'git lfs' in the tests.  no env var necessary because the bin path is added to PATH (Rick Olson[32m[m)
[1;34m|[m * [31m|[m [1;35m|[m 801dfc3 2015-05-21 fail early if the track command runs outside a git repo (Rick Olson[32m[m)
[1;34m|[m * [31m|[m [1;35m|[m   ae117f9 2015-05-21 Merge branch 'master' into track-tests (Rick Olson[32m[m)
[1;34m|[m [1;36m|[m[1;34m\[m [31m\[m [1;35m\[m  
[1;34m|[m [1;36m|[m[1;34m/[m [31m/[m [1;35m/[m  
[1;34m|[m[1;34m/[m[1;36m|[m [31m|[m [1;35m|[m   
* [1;36m|[m [31m|[m [1;35m|[m   38657a8 2015-05-21 Merge pull request #321 from github/configurable-test-dirs (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [31m\[m [1;35m\[m  
[32m|[m * [1;36m|[m [31m|[m [1;35m|[m f0b48a9 2015-05-21 configurable maxprocs too (Rick Olson[32m[m)
[32m|[m * [1;36m|[m [31m|[m [1;35m|[m f88f366 2015-05-21 configurable TMPDIR (Rick Olson[32m[m)
[32m|[m * [1;36m|[m [31m|[m [1;35m|[m b5a69cf 2015-05-21 use REMOTEDIR env var everywhere (Rick Olson[32m[m)
[32m|[m[32m/[m [1;36m/[m [31m/[m [1;35m/[m  
[32m|[m * [31m|[m [1;35m|[m 14e4d00 2015-05-21 replace track tests (Rick Olson[32m[m)
[32m|[m[32m/[m [31m/[m [1;35m/[m  
* [31m|[m [1;35m|[m   615b4d8 2015-05-21 fix merge conflict (Rick Olson[32m[m)
[34m|[m[35m\[m [31m\[m [1;35m\[m  
[34m|[m * [31m|[m [1;35m|[m c6a6c68 2015-05-15 Add instructions for bulding a deb (Stephen Gelman[32m[m)
[34m|[m * [31m|[m [1;35m|[m a5acd51 2015-05-15 Add debian directory and gitignore temporary build files (Stephen Gelman[32m[m)
* [35m|[m [31m|[m [1;35m|[m   09f72a4 2015-05-21 Merge pull request #319 from github/michael-k-panic-fix (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [35m\[m [31m\[m [1;35m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[35m|[m[1;35m_[m[31m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [35m|[m [31m|[m   
[1;35m|[m * [35m|[m [31m|[m 515f27e 2015-05-21 add failing test for https://github.com/github/git-lfs/issues/310 (Rick Olson[32m[m)
[1;35m|[m * [35m|[m [31m|[m be7581d 2015-05-21 add SKIPCOMPILE to tests (Rick Olson[32m[m)
[1;35m|[m * [35m|[m [31m|[m   d71ad5b 2015-05-21 Merge branch 'panic-fix' of https://github.com/michael-k/git-lfs into michael-k-panic-fix (Rick Olson[32m[m)
[1;35m|[m [1;35m|[m[1;33m\[m [35m\[m [31m\[m  
[1;35m|[m[1;35m/[m [1;33m/[m [35m/[m [31m/[m  
[1;35m|[m * [35m|[m [31m|[m 6d343ba 2015-05-16 Fixed detection of LocalGitDir within submodules (Michael Käufl[32m[m)
* [1;33m|[m [35m|[m [31m|[m   6ba0e21 2015-05-21 Merge pull request #306 from github/integration-tests-3 (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [35m\[m [31m\[m  
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 3bdfa6c 2015-05-21 bring the old credential helper back to see if ci is more reliable (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m c01bb2c 2015-05-21 report the git lfs version at the beginning of the script (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 6d811de 2015-05-21 dump the git lfs version if it fails (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 98407dd 2015-05-21 integration tests cleanup after themselves unless told not to (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 1f7a6e1 2015-05-20 remove the custom git credential helper (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 2bfc8a3 2015-05-20 use a really really simple credential helper (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 68f61f0 2015-05-20 update test to clone empty repos first (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 904145f 2015-05-19 lots of test docs (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m cc0ba8a 2015-05-19 better go conventions (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m d9e0b22 2015-05-19 dont exit 0, this affects the exit status of the test (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 1b36588 2015-05-18 dont show the logs and env if the test passes (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m b2a6fba 2015-05-18 fail fast if script/test fails (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m de1e0a0 2015-05-18 run tests in parallel with xargs (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m eb54207 2015-05-18 single = (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m d17948c 2015-05-18 double brackets (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m b4a7bf8 2015-05-18 teach script/integration how to clean up (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 2310602 2015-05-18 unsaved (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 7b6d89d 2015-05-18 teach script/integration how to setup and run the test suite (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 3f15399 2015-05-18 better test output (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 5914d7d 2015-05-18 implement the full push/clone workflow test (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 856779f 2015-05-18 use a while loop (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 8a97acc 2015-05-18 add a single command to run integration tests (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 9523561 2015-05-18 run the happy path test in the ci suite (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 667cde0 2015-05-18 test/local is not used anymore (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m b1f5025 2015-05-18 import testlib with as few changes as possible (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m b685624 2015-05-15 rename tests => test (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m e2afd37 2015-05-15 get test working with a git and lfs server (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 5fc4fde 2015-05-14 initial crappy sh test (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m da2a961 2015-05-14 initial readme (Rick Olson[32m[m)
[1;34m|[m * [1;33m|[m [35m|[m [31m|[m 5e78bce 2015-05-14 only build lfs and commands sub packages by default. (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [35m|[m [31m|[m   24c12a4 2015-05-19 Merge pull request #315 from sinbad/master (risk danger olson[32m[m)
[1;33m|[m[31m\[m [1;35m\[m [1;33m\[m [35m\[m [31m\[m  
[1;33m|[m [31m|[m[1;33m_[m[1;35m|[m[1;33m/[m [35m/[m [31m/[m  
[1;33m|[m[1;33m/[m[31m|[m [1;35m|[m [35m|[m [31m|[m   
[1;33m|[m * [1;35m|[m [35m|[m [31m|[m 081e089 2015-05-18 Fix spec error - there is no 'git lfs sync' command (Steve Streeting[32m[m)
[1;33m|[m[1;33m/[m [1;35m/[m [35m/[m [31m/[m  
* [1;35m|[m [35m|[m [31m|[m   c91bb6f 2015-05-16 Merge pull request #311 from michael-k/trace (Scott Barron[32m[m)
[35m|[m[33m\[m [1;35m\[m [35m\[m [31m\[m  
[35m|[m [33m|[m[35m_[m[1;35m|[m[35m/[m [31m/[m  
[35m|[m[35m/[m[33m|[m [1;35m|[m [31m|[m   
[35m|[m * [1;35m|[m [31m|[m d2f9880 2015-05-14 Traced performance in case of error too (Michael Käufl[32m[m)
* [33m|[m [1;35m|[m [31m|[m   0854875 2015-05-15 Merge pull request #305 from michael-k/errorhandling (risk danger olson[32m[m)
[1;35m|[m[35m\[m [33m\[m [1;35m\[m [31m\[m  
[1;35m|[m [35m|[m[1;35m_[m[33m|[m[1;35m/[m [31m/[m  
[1;35m|[m[1;35m/[m[35m|[m [33m|[m [31m|[m   
[1;35m|[m * [33m|[m [31m|[m e8c6175 2015-05-14 Write error to stderr, even if os.MkdirAll fails (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m a56881b 2015-05-14 Got rid of the recursive call (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m 5476c87 2015-05-14 Extracted function (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m 75b224a 2015-05-14 Reordered functions to improve top-down readability (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m dc614fc 2015-05-14 Don't try to log panic to file in recursive calls (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m eb12f25 2015-05-14 Don't return a file name we didn't write to (Michael Käufl[32m[m)
[1;35m|[m * [33m|[m [31m|[m af86976 2015-05-14 Kept error handling in one place (Michael Käufl[32m[m)
[1;35m|[m[1;35m/[m [33m/[m [31m/[m  
* [33m|[m [31m|[m   b94361f 2015-05-14 Merge pull request #303 from github/goimports (risk danger olson[32m[m)
[33m|[m[1;31m\[m [33m\[m [31m\[m  
[33m|[m [1;31m|[m[33m/[m [31m/[m  
[33m|[m[33m/[m[1;31m|[m [31m|[m   
[33m|[m * [31m|[m 92672e3 2015-05-14 Fixed contentaddressable import (Michael Käufl[32m[m)
[33m|[m * [31m|[m 3140957 2015-05-13 dont install goimports (Rick Olson[32m[m)
[33m|[m * [31m|[m 2166231 2015-05-13 Use goimports instead of gofmt (Rick Olson[32m[m)
[33m|[m[33m/[m [31m/[m  
* [31m|[m   29630a5 2015-05-13 Merge pull request #300 from michael-k/codestyle (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [31m\[m  
[1;32m|[m * [31m|[m 43b4deb 2015-05-13 Checked for err != nil instead of == nil (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 8890a09 2015-05-13 Checked for err != nil instead of == nil (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m fca70c0 2015-05-13 Checked for err != nil instead of == nil (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 3897869 2015-05-13 Decreased nesting (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 00e70ad 2015-05-13 Used fmt.Fprintln() to print line breaks (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 1b04957 2015-05-13 Removed else after return (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m d7fa0de 2015-05-13 Print error messages to stderr (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 4e8bd02 2015-05-11 Switched functions to improve top-down readability (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 08bb66e 2015-05-11 Return nil explicitly in newClientRequest (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 8ce6038 2015-05-11 Followed error handling style guide (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 3327bfb 2015-05-11 Decreased nesting (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 744f387 2015-05-11 Moved initialization statement in if statement (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m f8c2fa8 2015-05-11 Return nil as payload in case of error (Michael Käufl[32m[m)
[1;32m|[m * [31m|[m 6eef120 2015-05-11 Return nil explicitly in doApiRequest (Michael Käufl[32m[m)
* [1;33m|[m [31m|[m 9f118cd 2015-05-13 let's clone git-lfs from github :) (Rick Olson[32m (tag: v0.5.1-tracing)[m)
* [1;33m|[m [31m|[m 690a9b7 2015-05-13 80c (Rick Olson[32m[m)
* [1;33m|[m [31m|[m 300cf8b 2015-05-12 Build scripts for CentOS and Debian/Ubuntu (Jeff Haemer[32m[m)
* [1;33m|[m [31m|[m cbb1ed4 2015-05-12 Use "install -D" instead of "mkdir -p" and "cp". (Jeff Haemer[32m[m)
* [1;33m|[m [31m|[m 0a67e7a 2015-05-11 Add instructions for CentOS7 (Jeff Haemer[32m[m)
* [1;33m|[m [31m|[m 7ca03df 2015-05-11 Add instructions on how to get prereqs. (Jeff Haemer[32m[m)
* [1;33m|[m [31m|[m 6194fa4 2015-05-11 Case-harden Linux install script. (Jeff Haemer[32m[m)
[1;33m|[m[1;33m/[m [31m/[m  
* [31m|[m   923be66 2015-05-05 Merge pull request #283 from catphish/patch-1 (risk danger olson[32m[m)
[31m|[m[1;35m\[m [31m\[m  
[31m|[m [1;35m|[m[31m/[m  
[31m|[m[31m/[m[1;35m|[m   
[31m|[m * 3e355b3 2015-05-05 Remove trailing period when printing log file location (Charles Horatio Bernstrom[32m[m)
[31m|[m[31m/[m  
* 5e47f06 2015-04-30 Release v0.5.1 (Rick Olson[32m (tag: v0.5.1)[m)
*   8fd3fde 2015-04-30 Merge pull request #274 from github/release-0.5.1 (risk danger olson[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 65c535f 2015-04-30 mention windows installer fix (risk danger olson[32m[m)
[1;36m|[m * 368b97b 2015-04-30 mention the CHANGELOG in the release guide (risk danger olson[32m[m)
[1;36m|[m * 62c9712 2015-04-30 add a changelog (risk danger olson[32m[m)
* [31m|[m   085e17a 2015-04-30 Merge pull request #258 from github/concurrent-uploads (Scott Barron[32m[m)
[31m|[m[33m\[m [31m\[m  
[31m|[m [33m|[m[31m/[m  
[31m|[m[31m/[m[33m|[m   
[31m|[m * 2b1d10d 2015-04-30 Drain readers when not uploading file so progress bars can update (rubyist[32m[m)
[31m|[m * 6f549a3 2015-04-29 lfs.Scan() => lfs.ScanRefs() (rubyist[32m[m)
[31m|[m * 83c19bc 2015-04-29 Add some documentation to the upload queue and make some better names (rubyist[32m[m)
[31m|[m * e904708 2015-04-29 Use the queue in the pre push command as well (rubyist[32m[m)
[31m|[m *   0a9ebaf 2015-04-29 Merge branch 'master' into concurrent-uploads (rubyist[32m[m)
[31m|[m [34m|[m[35m\[m  
[31m|[m * [35m|[m 13612f0 2015-04-29 API can broadcast success/fail events to control queue synchronization (rubyist[32m[m)
[31m|[m * [35m|[m 5619c5c 2015-04-29 Syncrhonize uploads until one is successful, then allow concurrent (rubyist[32m[m)
[31m|[m * [35m|[m 23ea409 2015-04-24 Display all upload errors that occur (rubyist[32m[m)
[31m|[m * [35m|[m 7a77a30 2015-04-24 Display any errors when creating the CopyCallbackFile (rubyist[32m[m)
[31m|[m * [35m|[m f951e3e 2015-04-24 Update things from master (rubyist[32m[m)
[31m|[m * [35m|[m   2a8b026 2015-04-24 Merge branch 'master' into concurrent-uploads (rubyist[32m[m)
[31m|[m [36m|[m[1;31m\[m [35m\[m  
[31m|[m * [1;31m|[m [35m|[m d977daf 2015-04-24 Use pb's new Prefix() to count completed transfers (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m eb341f7 2015-04-24 Update pb (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m d9c36fa 2015-04-24 Fix tests, don't run the queue if it's a dry run push (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m a1ed0f2 2015-04-24 Initial pass at showing a total queue progress bar (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m f12efa2 2015-04-23 Add to the wg before writing upload to channel, some renames (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 2881870 2015-04-23 Add some docs to the upload queue code (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 3eec3e2 2015-04-23 Some refactoring and struct unification around the upload queue (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 0cacca1 2015-04-23 Move pointer and scanner packages into lfs (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 7fcdb6c 2015-04-23 Hash out a config value, default to 3 for now (rubyist[32m[m)
[31m|[m * [1;31m|[m [35m|[m 88b5de3 2015-04-22 Base concurrent upload queue (rubyist[32m[m)
* [1;31m|[m [1;31m|[m [35m|[m 8bb6b26 2015-04-30 list more commands in the git-lfs(1) man page (risk danger olson[32m[m)
* [1;31m|[m [1;31m|[m [35m|[m 0ccea94 2015-04-30 add a note to git lfs push man page (risk danger olson[32m[m)
* [1;31m|[m [1;31m|[m [35m|[m e3b9635 2015-04-29 fix test (Rick Olson[32m[m)
[35m|[m [1;31m|[m[35m_[m[1;31m|[m[35m/[m  
[35m|[m[35m/[m[1;31m|[m [1;31m|[m   
* [1;31m|[m [1;31m|[m   3321ccc 2015-04-29 Merge pull request #271 from github/dont-clean-pointers (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m [1;31m\[m  
[1;32m|[m * [1;31m|[m [1;31m|[m 31219e5 2015-04-28 bring over hook tests (Rick Olson[32m[m)
[1;32m|[m * [1;31m|[m [1;31m|[m e09e5e1 2015-04-28 pass pointers straight through 'git lfs clean' (Rick Olson[32m[m)
* [1;33m|[m [1;31m|[m [1;31m|[m   8a39372 2015-04-28 Merge pull request #267 from github/smudge-zero-len-fix (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [1;31m\[m [1;31m\[m  
[1;34m|[m * [1;33m|[m [1;31m|[m [1;31m|[m dc367b2 2015-04-28 fix smudge attempts from zero len file (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [1;31m|[m [1;31m|[m c366638 2015-04-28 set the connect and tls timeouts to just 5s (Rick Olson[32m[m)
* [1;35m|[m [1;33m|[m [1;31m|[m [1;31m|[m   5fc3150 2015-04-28 Merge pull request #215 from Mistobaan/feature/add-http-timeout (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;31m\[m  
[1;36m|[m * [1;35m|[m [1;33m|[m [1;31m|[m [1;31m|[m b4db8b6 2015-04-10 add default timeout to http client (Fabrizio (Misto) Milo[32m[m)
* [31m|[m [1;35m|[m [1;33m|[m [1;31m|[m [1;31m|[m   8dad4e8 2015-04-28 Merge pull request #263 from github/split-push-cmd (risk danger olson[32m[m)
[1;33m|[m[33m\[m [31m\[m [1;35m\[m [1;33m\[m [1;31m\[m [1;31m\[m  
[1;33m|[m [33m|[m[1;33m_[m[31m|[m[1;33m_[m[1;35m|[m[1;33m/[m [1;31m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[33m|[m [31m|[m [1;35m|[m [1;31m|[m [1;31m|[m   
[1;33m|[m * [31m|[m [1;35m|[m [1;31m|[m [1;31m|[m   f023179 2015-04-28 Merge branch 'master' into split-push-cmd (Rick Olson[32m[m)
[1;33m|[m [34m|[m[1;33m\[m [31m\[m [1;35m\[m [1;31m\[m [1;31m\[m  
[1;33m|[m [34m|[m[1;33m/[m [31m/[m [1;35m/[m [1;31m/[m [1;31m/[m  
[1;33m|[m[1;33m/[m[34m|[m [31m|[m [1;35m|[m [1;31m|[m [1;31m|[m   
* [34m|[m [31m|[m [1;35m|[m [1;31m|[m [1;31m|[m eac36d7 2015-04-28 add an old alpha alias (Rick Olson[32m[m)
* [34m|[m [31m|[m [1;35m|[m [1;31m|[m [1;31m|[m   56bd3c8 2015-04-28 Merge pull request #265 from github/update-track-attribs (risk danger olson[32m[m)
[1;35m|[m[1;31m\[m [34m\[m [31m\[m [1;35m\[m [1;31m\[m [1;31m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[34m|[m[1;35m_[m[31m|[m[1;35m/[m [1;31m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [34m|[m [31m|[m [1;31m|[m [1;31m|[m   
[1;35m|[m * [34m|[m [31m|[m [1;31m|[m [1;31m|[m 2aa1f7e 2015-04-24 add diff/merge properties to gitattributes (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [34m|[m[1;31m_[m[31m|[m[1;31m_[m[1;31m|[m[1;31m/[m  
[1;35m|[m [1;31m|[m[1;31m/[m[34m|[m [31m|[m [1;31m|[m   
[1;35m|[m [1;31m|[m * [31m|[m [1;31m|[m   ae3f873 2015-04-27 merge master (Rick Olson[32m[m)
[1;35m|[m [1;31m|[m [1;32m|[m[1;35m\[m [31m\[m [1;31m\[m  
[1;35m|[m [1;31m|[m[1;35m_[m[1;32m|[m[1;35m/[m [31m/[m [1;31m/[m  
[1;35m|[m[1;35m/[m[1;31m|[m [1;32m|[m [31m|[m [1;31m|[m   
* [1;31m|[m [1;32m|[m [31m|[m [1;31m|[m 97ef181 2015-04-27 mention possible 410 response in api (Rick Olson[32m[m)
* [1;31m|[m [1;32m|[m [31m|[m [1;31m|[m   eecc9b0 2015-04-24 Merge pull request #264 from github/require-stdin (risk danger olson[32m[m)
[1;31m|[m[1;35m\[m [1;31m\[m [1;32m\[m [31m\[m [1;31m\[m  
[1;31m|[m [1;35m|[m[1;31m/[m [1;32m/[m [31m/[m [1;31m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m [1;32m|[m [31m|[m [1;31m|[m   
[1;31m|[m * [1;32m|[m [31m|[m [1;31m|[m 596c5f6 2015-04-24 requires stdin content on certain commands (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m [1;32m/[m [31m/[m [1;31m/[m  
[1;31m|[m * [31m|[m [1;31m|[m   d5f8cb7 2015-04-24 Merge branch 'master' into split-push-cmd (Rick Olson[32m[m)
[1;31m|[m [1;36m|[m[1;31m\[m [31m\[m [1;31m\[m  
[1;31m|[m [1;36m|[m[1;31m/[m [31m/[m [1;31m/[m  
[1;31m|[m[1;31m/[m[1;36m|[m [31m|[m [1;31m|[m   
* [1;36m|[m [31m|[m [1;31m|[m   2e6bb2a 2015-04-24 Merge pull request #247 from github/0.5.0-ssh-spec (risk danger olson[32m[m)
[32m|[m[33m\[m [1;36m\[m [31m\[m [1;31m\[m  
[32m|[m * [1;36m|[m [31m|[m [1;31m|[m b1850da 2015-04-22 use the href property from the git-lfs-authenticate ssh command (Rick Olson[32m[m)
[32m|[m * [1;36m|[m [31m|[m [1;31m|[m   a825d49 2015-04-22 Merge branch 'master' into 0.5.0-ssh-spec (Rick Olson[32m[m)
[32m|[m [34m|[m[35m\[m [1;36m\[m [31m\[m [1;31m\[m  
[32m|[m * [35m|[m [1;36m|[m [31m|[m [1;31m|[m 5f48c4c 2015-04-19 add the optional HREF key (risk danger olson[32m[m)
* [35m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m   632a436 2015-04-24 Merge pull request #246 from github/0.5.0-pointer-spec (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [35m\[m [1;36m\[m [31m\[m [1;31m\[m  
[35m|[m [1;31m|[m[35m_[m[35m|[m[35m/[m [1;36m/[m [31m/[m [1;31m/[m  
[35m|[m[35m/[m[1;31m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m   
[35m|[m * [35m|[m [1;36m|[m [31m|[m [1;31m|[m bc62c32 2015-04-22 add man page for the pointer command (Rick Olson[32m[m)
[35m|[m * [35m|[m [1;36m|[m [31m|[m [1;31m|[m   e9627e1 2015-04-22 Merge branch 'master' into 0.5.0-pointer-spec (Rick Olson[32m[m)
[35m|[m [1;32m|[m[1;33m\[m [35m\[m [1;36m\[m [31m\[m [1;31m\[m  
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 9da6255 2015-04-22 implement 'git lfs pointer --stdin' (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m bd1b1ed 2015-04-22 change the pointer command's switches (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m c6b722d 2015-04-22 implement the pointer command (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 430e547 2015-04-22 don't close the cleaned object if its nil (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m a1a730c 2015-04-22 the pointer command isn't meant to be a plumbing command (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m df39074 2015-04-22 sketch out a 'git lfs pointer' command (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 035aafc 2015-04-22 don't mention the really old pointers (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 1c39eba 2015-04-22 mention the pre-release version url (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m c601997 2015-04-22 mention the string format for the oid field (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 80a1ff7 2015-04-22 add support for pre-release pointers, remove support for really old ones (Rick Olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 4bdecff 2015-04-20 fix typo (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 5c52dea 2015-04-20 path command is gone (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m 2df1c72 2015-04-20 this is important (risk danger olson[32m[m)
[35m|[m * [1;33m|[m [35m|[m [1;36m|[m [31m|[m [1;31m|[m c2b8dce 2015-04-19 stricter pointer spec (risk danger olson[32m[m)
[35m|[m [35m|[m [1;33m|[m[35m/[m [1;36m/[m [31m/[m [1;31m/[m  
[35m|[m [35m|[m[35m/[m[1;33m|[m [1;36m|[m [31m|[m [1;31m|[m   
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m 491fecd 2015-04-24 better error message for push with no args (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m cd4ccd3 2015-04-24 remove more repo/refspec references (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m 145cdf7 2015-04-24 tweak docs (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m ad416ef 2015-04-24 silently upgrade the old push hooks (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m 394d7db 2015-04-24 teach the update command how to update known git lfs hooks (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m 847a547 2015-04-24 split the push command code (Rick Olson[32m[m)
[35m|[m [35m|[m [1;33m|[m * [31m|[m [1;31m|[m 9fe3a5b 2015-04-24 split pre-push cmd from push (Rick Olson[32m[m)
[35m|[m [35m|[m[35m_[m[1;33m|[m[35m/[m [31m/[m [1;31m/[m  
[35m|[m[35m/[m[35m|[m [1;33m|[m [31m|[m [1;31m|[m   
* [35m|[m [1;33m|[m [31m|[m [1;31m|[m   f4b10e0 2015-04-22 Merge pull request #257 from github/better-terminal-prompt-message (risk danger olson[32m[m)
[1;31m|[m[1;35m\[m [35m\[m [1;33m\[m [31m\[m [1;31m\[m  
[1;31m|[m [1;35m|[m[1;31m_[m[35m|[m[1;31m_[m[1;33m|[m[1;31m_[m[31m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[1;35m|[m [35m|[m [1;33m|[m [31m|[m   
[1;31m|[m * [35m|[m [1;33m|[m [31m|[m ac5e6b2 2015-04-22 better error message if you call git-lfs with GIT_TERMINAL_PROMPT (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m [35m/[m [1;33m/[m [31m/[m  
* [35m|[m [1;33m|[m [31m|[m   383fced 2015-04-22 Merge pull request #252 from github/link-roadmap (risk danger olson[32m[m)
[1;33m|[m[31m\[m [35m\[m [1;33m\[m [31m\[m  
[1;33m|[m [31m|[m[1;33m_[m[35m|[m[1;33m/[m [31m/[m  
[1;33m|[m[1;33m/[m[31m|[m [35m|[m [31m|[m   
[1;33m|[m * [35m|[m [31m|[m f05b68c 2015-04-20 Link to roadmap in readme (Brandon Keepers[32m[m)
[1;33m|[m[1;33m/[m [35m/[m [31m/[m  
* [35m|[m [31m|[m   301131e 2015-04-20 Merge pull request #228 from github/man-tidy (risk danger olson[32m[m)
[32m|[m[33m\[m [35m\[m [31m\[m  
[32m|[m * [35m|[m [31m|[m 4592dd5 2015-04-20 update docs to mention references, not refspec (risk danger olson[32m[m)
[32m|[m * [35m|[m [31m|[m 769f542 2015-04-20 Remove the mention of the old push queue (risk danger olson[32m[m)
[32m|[m * [35m|[m [31m|[m 1b1e135 2015-04-13 `git lfs push`: make pre-push stdin content easier to understand (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 4eaad23 2015-04-13 Fix link to `git-lfs-env` command (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m ce2c73f 2015-04-13 Improve `git-lfs` manpage (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 5408c31 2015-04-13 `git lfs status`: improve manpage (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 66e59f1 2015-04-13 `git lfs track` and `git lfs untrack`: improve manpages (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 0a5d4ab 2015-04-13 Explain `git lfs clean` in a more user-relevant way (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m a66f021 2015-04-13 `git lfs smudge`: improve man page text (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 312dc41 2015-04-13 `git lfs clean` reads from standard input, not standard output (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m 56b7408 2015-04-13 Expand "STDIN/STDOUT" to "standard input/output" (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m e151e98 2015-04-13 Man pages: change command descriptions to the imperative voice (Michael Haggerty[32m[m)
[32m|[m * [35m|[m [31m|[m ac1451e 2015-04-11 Tidy up the man page summary lines (Michael Haggerty[32m[m)
* [33m|[m [35m|[m [31m|[m   a42f6bb 2015-04-20 Merge pull request #245 from github/0.5.0-config-spec (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [35m\[m [31m\[m  
[34m|[m * [33m|[m [35m|[m [31m|[m f18a260 2015-04-20 parse the git config from the local .gitconfig file first (risk danger olson[32m[m)
[34m|[m * [33m|[m [35m|[m [31m|[m 434647d 2015-04-19 clarify how git config is loaded (risk danger olson[32m[m)
[34m|[m [35m|[m [33m|[m[35m/[m [31m/[m  
[34m|[m [35m|[m[35m/[m[33m|[m [31m|[m   
* [35m|[m [33m|[m [31m|[m   42bac87 2015-04-20 Merge pull request #244 from github/0.5.0-spec-oid-path (risk danger olson[32m[m)
[35m|[m[1;31m\[m [35m\[m [33m\[m [31m\[m  
[35m|[m [1;31m|[m[35m/[m [33m/[m [31m/[m  
[35m|[m[35m/[m[1;31m|[m [33m|[m [31m|[m   
[35m|[m * [33m|[m [31m|[m 3f833de 2015-04-19 clarify OID-PATH (risk danger olson[32m[m)
[35m|[m[35m/[m [33m/[m [31m/[m  
* [33m|[m [31m|[m   2944655 2015-04-17 Merge pull request #241 from github/travistweak (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [31m\[m  
[1;32m|[m * [33m|[m [31m|[m 6ee6f2a 2015-04-17 Skip the install step in travis, everything is vendored (rubyist[32m[m)
[1;32m|[m[1;32m/[m [33m/[m [31m/[m  
* [33m|[m [31m|[m   d9624a5 2015-04-17 Merge pull request #240 from github/specfixes (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [33m\[m [31m\[m  
[1;34m|[m * [33m|[m [31m|[m 518c18b 2015-04-17 Update spec to reflect lfs.url and lfsurl (rubyist[32m[m)
[1;34m|[m[1;34m/[m [33m/[m [31m/[m  
* [33m|[m [31m|[m   0a47c6b 2015-04-16 Merge pull request #237 from github/fix-lfs-url-key (risk danger olson[32m[m)
[1;36m|[m[31m\[m [33m\[m [31m\[m  
[1;36m|[m * [33m|[m [31m|[m 543b5ab 2015-04-15 use remote.{name}.lfsurl to specify an endpoint per remote (Rick Olson[32m[m)
* [31m|[m [33m|[m [31m|[m   568166c 2015-04-16 Merge pull request #232 from github/create-roadmap (risk danger olson[32m[m)
[31m|[m[33m\[m [31m\[m [33m\[m [31m\[m  
[31m|[m [33m|[m[31m/[m [33m/[m [31m/[m  
[31m|[m[31m/[m[33m|[m [33m|[m [31m|[m   
[31m|[m * [33m|[m [31m|[m a8f34e9 2015-04-16 mention automatic updates (risk danger olson[32m[m)
[31m|[m * [33m|[m [31m|[m fd91160 2015-04-14 add a rough roadmap (risk danger olson[32m[m)
[31m|[m [33m|[m[33m/[m [31m/[m  
* [33m|[m [31m|[m   7efbd3e 2015-04-15 Merge pull request #185 from github/upload-progress-fix (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [31m\[m  
[34m|[m * [33m|[m [31m|[m 54cd2cc 2015-04-08 dont need to print the size (Rick Olson[32m[m)
[34m|[m * [33m|[m [31m|[m 42013d6 2015-04-08 fix the progress bar output by calling Finish() (Rick Olson[32m[m)
* [35m|[m [33m|[m [31m|[m   c5f6c6f 2015-04-15 Merge pull request #225 from thekafkaf/command_untrack-docs (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [33m\[m [31m\[m  
[36m|[m * [35m|[m [33m|[m [31m|[m e2d50bf 2015-04-11 Grammer (kafkaf-[32m[m)
[36m|[m * [35m|[m [33m|[m [31m|[m 37d03f6 2015-04-11 Small doc refactor (kafkaf-[32m[m)
[36m|[m * [35m|[m [33m|[m [31m|[m 21815f1 2015-04-11 Small refactor (kafkaf-[32m[m)
[36m|[m * [35m|[m [33m|[m [31m|[m 6da9459 2015-04-11 logical doc inside untrackCommand (kafkaf-[32m[m)
[36m|[m * [35m|[m [33m|[m [31m|[m 34f1d39 2015-04-11 untrackCommand doc (kafkaf-[32m[m)
[36m|[m [33m|[m [35m|[m[33m/[m [31m/[m  
[36m|[m [33m|[m[33m/[m[35m|[m [31m|[m   
* [33m|[m [35m|[m [31m|[m 7b0abef 2015-04-15 refer to it as 'Git LFS' always. (Rick Olson[32m[m)
* [33m|[m [35m|[m [31m|[m   94794c9 2015-04-15 Merge pull request #212 from rtyley/patch-1 (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [35m\[m [31m\[m  
[1;32m|[m * [33m|[m [35m|[m [31m|[m 3e94cd2 2015-04-11 Move definition of sharded path to above 1st use (Roberto Tyley[32m[m)
[1;32m|[m * [33m|[m [35m|[m [31m|[m 7f260e5 2015-04-10 Spec: clarify that filepath sharding is used in .git/lfs/objects/ (Roberto Tyley[32m[m)
* [1;33m|[m [33m|[m [35m|[m [31m|[m   495db0d 2015-04-15 Merge pull request #235 from mhagger/pre-push-arg-quoting (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;33m\[m [33m\[m [35m\[m [31m\[m  
[1;34m|[m * [1;33m|[m [33m|[m [35m|[m [31m|[m 0dd0e17 2015-04-15 pre-push hook: properly quote args that are passed along (Michael Haggerty[32m[m)
[1;34m|[m[1;34m/[m [1;33m/[m [33m/[m [35m/[m [31m/[m  
* [1;33m|[m [33m|[m [35m|[m [31m|[m   428285f 2015-04-14 Merge pull request #213 from github/gpm-license (risk danger olson[32m[m)
[33m|[m[31m\[m [1;33m\[m [33m\[m [35m\[m [31m\[m  
[33m|[m [31m|[m[33m_[m[1;33m|[m[33m/[m [35m/[m [31m/[m  
[33m|[m[33m/[m[31m|[m [1;33m|[m [35m|[m [31m|[m   
[33m|[m * [1;33m|[m [35m|[m [31m|[m c65f637 2015-04-14 link to the actual project (risk danger olson[32m[m)
[33m|[m * [1;33m|[m [35m|[m [31m|[m 29f234b 2015-04-10 Add gpm's license to its file (rubyist[32m[m)
[33m|[m [1;33m|[m[1;33m/[m [35m/[m [31m/[m  
* [1;33m|[m [35m|[m [31m|[m   03a0d4c 2015-04-10 Merge pull request #223 from PeterDaveHello/patch-3 (Scott Barron[32m[m)
[32m|[m[33m\[m [1;33m\[m [35m\[m [31m\[m  
[32m|[m * [1;33m|[m [35m|[m [31m|[m 788ef52 2015-04-11 Update install.bat.example (Peter Dave Hello[32m[m)
[32m|[m[32m/[m [1;33m/[m [35m/[m [31m/[m  
* [1;33m|[m [35m|[m [31m|[m   2eb12ec 2015-04-10 Merge pull request #221 from github/media-removal (Scott Barron[32m[m)
[34m|[m[35m\[m [1;33m\[m [35m\[m [31m\[m  
[34m|[m * [1;33m|[m [35m|[m [31m|[m c7b71c1 2015-04-10 Remove a few more git media refs (rubyist[32m[m)
[34m|[m [1;33m|[m[1;33m/[m [35m/[m [31m/[m  
* [1;33m|[m [35m|[m [31m|[m   a46d0f3 2015-04-10 Merge pull request #218 from thekafkaf/refactor (risk danger olson[32m[m)
[31m|[m[1;31m\[m [1;33m\[m [35m\[m [31m\[m  
[31m|[m [1;31m|[m[31m_[m[1;33m|[m[31m_[m[35m|[m[31m/[m  
[31m|[m[31m/[m[1;31m|[m [1;33m|[m [35m|[m   
[31m|[m * [1;33m|[m [35m|[m 05d4692 2015-04-10 Breaking when a known path was found, no need to continue the iteration (kafkaf-[32m[m)
[31m|[m * [1;33m|[m [35m|[m 68c1249 2015-04-10 Removed a redundant assignment of a variable by moving it upper (kafkaf-[32m[m)
[31m|[m[31m/[m [1;33m/[m [35m/[m  
* [1;33m|[m [35m|[m   1cb54bc 2015-04-10 Merge pull request #205 from PeterDaveHello/patch-1 (Brandon Keepers[32m[m)
[1;33m|[m[1;33m\[m [1;33m\[m [35m\[m  
[1;33m|[m [1;33m|[m[1;33m/[m [35m/[m  
[1;33m|[m[1;33m/[m[1;33m|[m [35m|[m   
[1;33m|[m * [35m|[m a0aca3a 2015-04-10 Add travis CI badge (Peter Dave Hello[32m[m)
[1;33m|[m[1;33m/[m [35m/[m  
* [35m|[m   d6b1e6d 2015-04-09 Merge pull request #194 from mattn/go-gettable (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m [35m\[m  
[1;34m|[m * [35m|[m 320a2c6 2015-04-09 fix build script (Yasuhiro Matsumoto[32m[m)
[1;34m|[m * [35m|[m 6813971 2015-04-09 fix script (Yasuhiro Matsumoto[32m[m)
[1;34m|[m * [35m|[m 4c3d468 2015-04-09 possible to go get (Yasuhiro Matsumoto[32m[m)
[1;34m|[m[1;34m/[m [35m/[m  
* [35m|[m   5eb9bb0 2015-04-08 Merge pull request #192 from pborreli/typos (Timothy Clem[32m[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m|[m ce6dcd3 2015-04-09 Fixed typos (Pascal Borreli[32m[m)
[1;36m|[m[1;36m/[m [35m/[m  
* [35m|[m   e3bc1a3 2015-04-08 Merge pull request #188 from fj/patch-1 (Scott Barron[32m[m)
[32m|[m[33m\[m [35m\[m  
[32m|[m * [35m|[m 3e4b678 2015-04-08 Specify that `size` property is in bytes (John Feminella[32m[m)
* [33m|[m [35m|[m   78bfc39 2015-04-08 Merge pull request #191 from github/travis-ci-nosudo (Scott Barron[32m[m)
[34m|[m[35m\[m [33m\[m [35m\[m  
[34m|[m * [33m|[m [35m|[m 61e13e2 2015-04-08 Sudo may not be necessary (Brandon Keepers[32m[m)
* [35m|[m [33m|[m [35m|[m   44f668d 2015-04-08 Merge pull request #181 from github/travis-ci (Brandon Keepers[32m[m)
[36m|[m[35m\[m [35m\[m [33m\[m [35m\[m  
[36m|[m [35m|[m[35m/[m [33m/[m [35m/[m  
[36m|[m * [33m|[m [35m|[m 243766a 2015-04-08 Set git config in script/cibuild (Brandon Keepers[32m[m)
[36m|[m * [33m|[m [35m|[m 596d13d 2015-04-08 See if setting these globally fixes travis (Brandon Keepers[32m[m)
[36m|[m * [33m|[m [35m|[m 4e6175b 2015-04-08 Set a default git name/email (Brandon Keepers[32m[m)
[36m|[m * [33m|[m [35m|[m ac953f8 2015-04-08 require sudo (Brandon Keepers[32m[m)
[36m|[m * [33m|[m [35m|[m d68ac27 2015-04-08 Add Travis CI build config (Brandon Keepers[32m[m)
[36m|[m [33m|[m[33m/[m [35m/[m  
* [33m|[m [35m|[m   ddd4d54 2015-04-08 Merge pull request #189 from lucaswerkmeister/patch-1 (Brandon Keepers[32m[m)
[33m|[m[1;33m\[m [33m\[m [35m\[m  
[33m|[m [1;33m|[m[33m/[m [35m/[m  
[33m|[m[33m/[m[1;33m|[m [35m|[m   
[33m|[m * [35m|[m 2a6a35d 2015-04-08 CONTRIBUTING.md: Fix link to man pages (Lucas Werkmeister[32m[m)
[33m|[m[33m/[m [35m/[m  
* [35m|[m   a0d3893 2015-04-08 Merge pull request #187 from github/nit-pickin (Lee Reilly[32m[m)
[1;34m|[m[1;35m\[m [35m\[m  
[1;34m|[m * [35m|[m 12022f1 2015-04-08 Colons consistently before code (Lee Reilly[32m[m)
[1;34m|[m[1;34m/[m [35m/[m  
* [35m|[m   a871542 2015-04-08 Merge pull request #186 from chrissiebrodigan/patch-1 (Timothy Clem[32m[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m|[m bae3135 2015-04-08 Add https to link (Timothy Clem[32m[m)
[1;36m|[m * [35m|[m 79244ba 2015-04-08 Update README.md (Timothy Clem[32m[m)
[1;36m|[m * [35m|[m bbed5a3 2015-04-08 List known implementations (Timothy Clem[32m[m)
[1;36m|[m * [35m|[m 1733b01 2015-04-08 Adding copy for the early access program (Chrissie Brodigan[32m[m)
[1;36m|[m[1;36m/[m [35m/[m  
* [35m|[m 594150c 2015-04-08 Say support on GitHub.com is coming in README (Timothy Clem[32m[m)
* [35m|[m   f8e5979 2015-04-08 Merge branch 'readme-tweaks' (Rick Olson[32m[m)
[35m|[m[33m\[m [35m\[m  
[35m|[m [33m|[m[35m/[m  
[35m|[m[35m/[m[33m|[m   
[35m|[m * c74f78a 2015-04-08 :scissors: (Brandon Keepers[32m[m)
[35m|[m * 4ea5e41 2015-04-03 :pencil2: (Brandon Keepers[32m[m)
[35m|[m * 6801f8d 2015-04-02 add releasing docs (risk danger olson[32m[m)
[35m|[m * c64b513 2015-04-02 typo (risk danger olson[32m[m)
[35m|[m * ebca3d2 2015-04-02 80c (risk danger olson[32m[m)
[35m|[m *   6bcdc30 2015-04-02 merge master into readme-tweaks (risk danger olson[32m[m)
[35m|[m [34m|[m[35m\[m  
[35m|[m * [35m|[m 6a768d0 2015-03-24 :pencil2: (Brandon Keepers[32m[m)
[35m|[m * [35m|[m a39015d 2015-03-24 Add features (Brandon Keepers[32m[m)
[35m|[m * [35m|[m 1f8ae39 2015-03-24 Link to client and sever specs (Brandon Keepers[32m[m)
* [35m|[m [35m|[m 07d86f5 2015-04-08 add docs (Rick Olson[32m[m)
* [35m|[m [35m|[m 5050f05 2015-04-08 ship v0.5.0 (Rick Olson[32m (tag: v0.5.0)[m)
* [35m|[m [35m|[m   77dba8f 2015-04-08 Merge pull request #183 from github/ssh (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m [35m\[m  
[36m|[m * [35m\[m [35m\[m   a72262b 2015-04-03 merge master into ssh (risk danger olson[32m[m)
[36m|[m [1;32m|[m[1;33m\[m [35m\[m [35m\[m  
[36m|[m * [1;33m|[m [35m|[m [35m|[m f417fa5 2015-03-30 print ssh error to tracerx so it can try hitting the api directly (Rick Olson[32m[m)
[36m|[m * [1;33m|[m [35m|[m [35m|[m 6b95a6e 2015-03-29 merge ssh header if the git remote is an ssh endpoint (Rick Olson[32m[m)
[36m|[m * [1;33m|[m [35m|[m [35m|[m 9c67ba5 2015-03-29 promote ObjectUrl to a package level function (Rick Olson[32m[m)
[36m|[m * [1;33m|[m [35m|[m [35m|[m 514d947 2015-03-27 teach config.Endpoint() to return an Endpoint struct with optional ssh info (Rick Olson[32m[m)
[36m|[m * [1;33m|[m [35m|[m [35m|[m f131898 2015-03-27 rename these test methods (Rick Olson[32m[m)
[36m|[m * [1;33m|[m [35m|[m [35m|[m 9291213 2015-03-27 tweak release script temporarily (Rick Olson[32m[m)
[36m|[m [35m|[m [1;33m|[m[35m_[m[35m|[m[35m/[m  
[36m|[m [35m|[m[35m/[m[1;33m|[m [35m|[m   
* [35m|[m [1;33m|[m [35m|[m   edc52fc 2015-04-07 Merge pull request #184 from github/hawser-references (risk danger olson[32m[m)
[1;33m|[m[1;35m\[m [35m\[m [1;33m\[m [35m\[m  
[1;33m|[m [1;35m|[m[1;33m_[m[35m|[m[1;33m/[m [35m/[m  
[1;33m|[m[1;33m/[m[1;35m|[m [35m|[m [35m|[m   
[1;33m|[m * [35m|[m [35m|[m 94f238c 2015-04-06 Replace Hawser reference with Git LFS in comment (Lee Reilly[32m[m)
[1;33m|[m * [35m|[m [35m|[m 8b0e55a 2015-04-06 Replace Haswer ref with Git LFS in man pages (Lee Reilly[32m[m)
[1;33m|[m * [35m|[m [35m|[m 4e0d0c2 2015-04-06 Replace Hawser reference with Git LFS in comment (Lee Reilly[32m[m)
[1;33m|[m * [35m|[m [35m|[m 166d70c 2015-04-06 Replace Haswer ref with Git LFS in man pages (Lee Reilly[32m[m)
[1;33m|[m[1;33m/[m [35m/[m [35m/[m  
* [35m|[m [35m|[m b3ebec0 2015-04-02 update the readme to use the latest track command (risk danger olson[32m[m)
[35m|[m[35m/[m [35m/[m  
* [35m|[m dd17828 2015-03-27 first v0.5.0 build (Rick Olson[32m[m)
* [35m|[m   6d2e84e 2015-03-27 Merge pull request #4 from github/rewrite-client (risk danger olson[32m (tag: v0.5.0.pre1)[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m|[m 16f882d 2015-03-27 remove shaky language (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 10b0258 2015-03-27 just dup the line (Rick Olson[32m[m)
[1;36m|[m * [35m|[m f134d93 2015-03-26 remove debugging prints (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 430d956 2015-03-26 support redirects on POST requests (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 29aa653 2015-03-26 don't set auth on redirections if the scheme or host do not match (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 588a4d9 2015-03-26 support multiple redirects (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 6b6a718 2015-03-26 support GET redirects (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 474ba75 2015-03-26 remove unused error (Rick Olson[32m[m)
[1;36m|[m * [35m|[m a56b11d 2015-03-26 we don't need a non-redirecting http client anymore (Rick Olson[32m[m)
[1;36m|[m * [35m|[m b4f99ea 2015-03-22 dont show tracing messages without GIT_CURL_VERBOSE (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 5b2e9c3 2015-03-22 change remote.{name}.lfs config to remote.{name}.lfs_url (Rick Olson[32m[m)
[1;36m|[m * [35m|[m   abd0fd1 2015-03-22 Merge branch 'rename-path-to-track' into rewrite-client (Rick Olson[32m[m)
[1;36m|[m [32m|[m[33m\[m [35m\[m  
[1;36m|[m [32m|[m * [35m\[m   495137f 2015-03-22 Merge branch 'interim' into rename-path-to-track (Rick Olson[32m[m)
[1;36m|[m [32m|[m [34m|[m[35m\[m [35m\[m  
[1;36m|[m [32m|[m [34m|[m [35m|[m[35m/[m  
[1;36m|[m [32m|[m * [35m|[m   4c985db 2015-03-22 merge (Rick Olson[32m[m)
[1;36m|[m [32m|[m [36m|[m[1;31m\[m [35m\[m  
[1;36m|[m [32m|[m * [1;31m|[m [35m|[m 51f6e81 2015-03-22 update man pages (Rick Olson[32m[m)
[1;36m|[m [32m|[m * [1;31m|[m [35m|[m   042cc64 2015-03-22 Merge branch 'rewrite-client' into rename-path-to-track (Rick Olson[32m[m)
[1;36m|[m [32m|[m [1;32m|[m[1;33m\[m [1;31m\[m [35m\[m  
[1;36m|[m [32m|[m * [1;33m|[m [1;31m|[m [35m|[m b500178 2015-03-22 rename path => track (Rick Olson[32m[m)
[1;36m|[m * [1;33m|[m [1;33m|[m [1;31m|[m [35m|[m   e0bdba0 2015-03-22 Merge branch 'interim' into rewrite-client (Rick Olson[32m[m)
[1;36m|[m [1;31m|[m[35m\[m [1;33m\[m [1;33m\[m [1;31m\[m [35m\[m  
[1;36m|[m [1;31m|[m [35m|[m[1;31m_[m[1;33m|[m[1;31m_[m[1;33m|[m[1;31m/[m [35m/[m  
[1;36m|[m [1;31m|[m[1;31m/[m[35m|[m [1;33m|[m [1;33m|[m [35m/[m   
[1;36m|[m [1;31m|[m [35m|[m [1;33m|[m[35m_[m[1;33m|[m[35m/[m    
[1;36m|[m [1;31m|[m [35m|[m[35m/[m[1;33m|[m [1;33m|[m     
[1;36m|[m * [35m|[m [1;33m|[m [1;33m|[m   9d406f64 2015-03-22 merge (Rick Olson[32m[m)
[1;36m|[m [1;33m|[m[31m\[m [35m\[m [1;33m\[m [1;33m\[m  
[1;36m|[m [1;33m|[m [31m|[m[1;33m_[m[35m|[m[1;33m_[m[1;33m|[m[1;33m/[m  
[1;36m|[m [1;33m|[m[1;33m/[m[31m|[m [35m|[m [1;33m|[m   
[1;36m|[m * [31m|[m [35m|[m [1;33m|[m   c90cf24 2015-03-22 Merge branch 'rename-again' into rewrite-client (Rick Olson[32m[m)
[1;36m|[m [1;33m|[m[33m\[m [31m\[m [35m\[m [1;33m\[m  
[1;36m|[m [1;33m|[m [33m|[m[1;33m_[m[31m|[m[1;33m_[m[35m|[m[1;33m/[m  
[1;36m|[m [1;33m|[m[1;33m/[m[33m|[m [31m|[m [35m|[m   
[1;36m|[m * [33m|[m [31m|[m [35m|[m 7369c01 2015-03-22 fix weird variable name (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 1f60f98 2015-03-22 count all uploaded and downloaded bodies when tracing http calls (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 2c57372 2015-03-22 make sure we close response bodies correctly (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 61175c3 2015-03-22 trace the counted bytes when Close() is called (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 3fc67ed 2015-03-22 embed the structs (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 2f8d1fc 2015-03-22 move smudge message before the download actually starts (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 3f1bd97 2015-03-22 print a msg when downloading a file (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 2550678 2015-03-21 set the content length for all of the Upload() requests (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m bff10be 2015-03-21 redundant tracerx messages (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m d743ff4 2015-03-21 Upload() looks for http 200, not 202 (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 1d9edbb 2015-03-20 use New64 so we can pass an int64 to the progress bar (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m cbf8a83 2015-03-20 test CopyCallback in Upload() (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m e95c19c 2015-03-20 replace fmt.Printf with tracerx.Printf (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 0f40a18 2015-03-20 pass just the oid to Download() (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 3b408a0 2015-03-20 test that Upload() skips the PUT on a 202 status (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m c5aa13a 2015-03-20 allow application/json responses (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 4872f51 2015-03-19 more client upload and download tests (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 91f3491 2015-03-19 check authorization header in download tests (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 81d7639 2015-03-19 unused (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m cd87658 2015-03-19 implement uploads and downloads (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m 55fe5b9 2015-03-19 implement http response handling (Rick Olson[32m[m)
[1;36m|[m * [33m|[m [31m|[m [35m|[m d5a6bf8 2015-03-19 restart client with simple getCreds() (Rick Olson[32m[m)
* [33m|[m [33m|[m [31m|[m [35m|[m   e370b10 2015-03-26 Merge pull request #6 from github/license (Brandon Keepers[32m[m)
[35m|[m[35m\[m [33m\[m [33m\[m [31m\[m [35m\[m  
[35m|[m [35m|[m[35m_[m[33m|[m[35m_[m[33m|[m[35m_[m[31m|[m[35m/[m  
[35m|[m[35m/[m[35m|[m [33m|[m [33m|[m [31m|[m   
[35m|[m * [33m|[m [33m|[m [31m|[m 14e6631 2015-03-24 Create LICENSE (Brandon Keepers[32m[m)
[35m|[m[35m/[m [33m/[m [33m/[m [31m/[m  
* [33m|[m [33m|[m [31m|[m   ac33382 2015-03-22 Merge pull request #1 from github/rename-again (risk danger olson[32m[m)
[36m|[m[31m\[m [33m\[m [33m\[m [31m\[m  
[36m|[m [31m|[m [33m|[m[31m_[m[33m|[m[31m/[m  
[36m|[m [31m|[m[31m/[m[33m|[m [33m|[m   
[36m|[m * [33m|[m [33m|[m e69c476 2015-03-22 fix some places where we're still referring to 'git media' (Rick Olson[32m[m)
[36m|[m [33m|[m [33m|[m[33m/[m  
[36m|[m [33m|[m[33m/[m[33m|[m   
[36m|[m * [33m|[m c0846b1 2015-03-22 rename man pages (Rick Olson[32m[m)
[36m|[m [33m|[m[33m/[m  
[36m|[m * 88a2231 2015-03-19 ObjectsUrl() can now return an error (Rick Olson[32m[m)
[36m|[m * 476657d 2015-03-19 config_media_url.git => config_lfs_url.git (used for integration tests) (Rick Olson[32m[m)
[36m|[m * ce06758 2015-03-19 change info/media suffix to info/lfs (Rick Olson[32m[m)
[36m|[m * e37b695 2015-03-19 rename hawser => git-lfs (Rick Olson[32m[m)
[36m|[m[36m/[m  
*   550e2a9 2015-03-19 Merge pull request #176 from hawser/ensure-cleaned-objects (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m *   5d14e02 2015-03-19 Merge branch 'master' into ensure-cleaned-objects (Rick Olson[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m   
* [1;34m|[m   12575e0 2015-03-19 Merge pull request #174 from hawser/simplify-api (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m  
[1;36m|[m * [1;34m|[m e20a566 2015-03-19 add request_id to docs (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 9827837 2015-03-19 client error returns docs and/or request id if given (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 7377f3b 2015-03-19 dont reject creds on 404 (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m c5339c3 2015-03-19 include DocumentationUrl on the error object (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m c269041 2015-03-19 describe the error object (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 3f0d1d3 2015-03-19 specify more error codes (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 6476bca 2015-03-10 specify what statuses the POST relations appear (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m d601942 2015-02-27 specify (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 50ab3a8 2015-02-27 rip this out.  return 404 for objects that dont exist (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 9e111b0 2015-02-27 define a response for a media blob that doesnt exist yet (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 92ea321 2015-02-27 redundant paragraph (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 9001b87 2015-02-26 dont forget ports (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 4e4778e 2015-02-26 remove all the redirection stuff.  simple, consistent api +1 (Rick Olson[32m[m)
* [31m|[m [1;34m|[m   140b755 2015-03-05 Merge pull request #175 from hawser/http-trace (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [1;34m\[m  
[32m|[m * [31m|[m [1;34m|[m d4a1b55 2015-03-05 dead function (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m 0ad03b3 2015-03-05 extra linebreak for readability (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m d298867 2015-03-05 count how many bytes are uploaded (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m cda5bab 2015-03-05 dont bother supporting GIT_HTTP_VERBOSE (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m 5847f6b 2015-03-05 dont need to tweak where curl traces go (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m 57e6538 2015-03-05 missed some output (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m 2ca960f 2015-03-05 lets write curl verbose to stderr (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m bba0738 2015-03-05 let the commands package tell the hawser package where to send output (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m abcb4ba 2015-03-05 add a Configuration constructor (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m 6e3b35a 2015-03-05 implement part of curl verbose (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m b8ad7ba 2015-03-05 httpTransportFor() is only used once (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m a916c95 2015-03-05 rename doRequest so it's clear why you'd use it (Rick Olson[32m[m)
[32m|[m * [31m|[m [1;34m|[m c009ff4 2015-03-05 trace http calls in DoHTTP() (Rick Olson[32m[m)
[32m|[m[32m/[m [31m/[m [1;34m/[m  
[32m|[m [31m|[m * 5d239d6 2015-03-05 ensure cleaned hawser objects exist before pushing (Rick Olson[32m[m)
[32m|[m [31m|[m[32m/[m  
[32m|[m[32m/[m[31m|[m   
* [31m|[m   cb666d2 2015-03-05 Merge pull request #173 from hawser/download-from-metadata (risk danger olson[32m[m)
[31m|[m[35m\[m [31m\[m  
[31m|[m [35m|[m[31m/[m  
[31m|[m[31m/[m[35m|[m   
[31m|[m * 2dd5983 2015-02-25 teach Download() how to follow hypermedia (Rick Olson[32m[m)
[31m|[m * 7728ac2 2015-02-25 teach objectResource how to build http requests (Rick Olson[32m[m)
[31m|[m * 22ebd69 2015-02-25 better names for the internal structs (Rick Olson[32m[m)
[31m|[m * b7ec4a7 2015-02-25 unused (Rick Olson[32m[m)
[31m|[m * 80e9855 2015-02-25 document the media types (Rick Olson[32m[m)
[31m|[m[31m/[m  
* b366856 2015-02-18 v0.4.1 (Rick Olson[32m[m)
*   f56f240 2015-02-18 Merge pull request #168 from hawser/v0.4.0-bugs (risk danger olson[32m (tag: v0.4.1)[m)
[36m|[m[1;31m\[m  
[36m|[m * 887117c 2015-02-17 add a trailing line break to .gitattributes before adding new entries (Rick Olson[32m[m)
[36m|[m * ec0d0ed 2015-02-17 get the creds from the verify request (Rick Olson[32m[m)
[36m|[m[36m/[m  
*   0cc22d2 2015-02-17 Merge pull request #165 from hawser/dont-panic (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m *   2959942 2015-02-17 Merge pull request #166 from hawser/creds-for-external-put (risk danger olson[32m[m)
[1;32m|[m [1;34m|[m[1;35m\[m  
[1;32m|[m [1;34m|[m * a3fed8e 2015-02-17 ensure we always send User-Agent (Rick Olson[32m[m)
[1;32m|[m [1;34m|[m * 466bca7 2015-02-17 some tests around getRequestCreds() (Rick Olson[32m[m)
[1;32m|[m [1;34m|[m * 88acbc3 2015-02-17 get/set credentials for external put and verify requests (Rick Olson[32m[m)
[1;32m|[m [1;34m|[m[1;34m/[m  
[1;32m|[m * 03994bf 2015-02-17 don't panic for common http errors (Rick Olson[32m[m)
[1;32m|[m[1;32m/[m  
*   e605ae0 2015-02-17 Merge pull request #163 from hawser/download-redirect (risk danger olson[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m *   7252be8 2015-02-17 merge master (Rick Olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m  
[1;36m|[m [32m|[m[1;36m/[m  
[1;36m|[m[1;36m/[m[32m|[m   
* [32m|[m   5f716e5 2015-02-17 Merge pull request #154 from hawser/alternate-remotes (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m  
[34m|[m * [32m|[m df17f10 2015-02-02 add support for alternate remotes by grabbing it out of the `git hawser push` args (Rick Olson[32m[m)
[34m|[m [35m|[m * 6aee89e 2015-02-13 fix download redirections by creating a separate http client for GET/HEAD (Rick Olson[32m[m)
[34m|[m [35m|[m * 2f6767a 2015-02-13 test downloads with legacy git media header (Rick Olson[32m[m)
[34m|[m [35m|[m * 00a381d 2015-02-13 put tests in their expected places (Rick Olson[32m[m)
[34m|[m [35m|[m[34m/[m  
[34m|[m[34m/[m[35m|[m   
* [35m|[m   8f6c7e8 2015-02-13 Merge pull request #162 from hawser/tweak-api (risk danger olson[32m[m)
[36m|[m[1;31m\[m [35m\[m  
[36m|[m * [35m\[m   e5194a8 2015-02-13 merge (Rick Olson[32m[m)
[36m|[m [1;32m|[m[36m\[m [35m\[m  
[36m|[m [1;32m|[m[36m/[m [35m/[m  
[36m|[m[36m/[m[1;32m|[m [35m|[m   
* [1;32m|[m [35m|[m 76d1579 2015-02-13 remove the *Request structs, just pass args (Rick Olson[32m[m)
[1;33m|[m * [35m|[m 78cc969 2015-02-13 implement spec changes in the client (Rick Olson[32m[m)
[1;33m|[m * [35m|[m c39298c 2015-02-13 make some api spec changes. (Rick Olson[32m[m)
[1;33m|[m[1;33m/[m [35m/[m  
* [35m|[m   cff5715 2015-02-13 Merge pull request #161 from hawser/kill-hawserclient (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [35m\[m  
[1;34m|[m * [35m|[m fd63c06 2015-02-12 don't export the internal client methods (Rick Olson[32m[m)
[1;34m|[m * [35m|[m e652927 2015-02-12 split up legacy client tests (Rick Olson[32m[m)
[1;34m|[m * [35m|[m f8b861d 2015-02-12 move ObjectUrl() to a func on Configuration (Rick Olson[32m[m)
[1;34m|[m * [35m|[m 3488d5a 2015-02-12 change hawser.Get() to hawser.Download() (Rick Olson[32m[m)
[1;34m|[m * [35m|[m b07899e 2015-02-12 move hawserclient code into hawser package (Rick Olson[32m[m)
[1;34m|[m[1;34m/[m [35m/[m  
* [35m|[m   70ab696 2015-02-12 Merge pull request #160 from hawser/hawser-client-tests (risk danger olson[32m[m)
[1;36m|[m[31m\[m [35m\[m  
[1;36m|[m * [35m|[m 178f4c3 2015-02-11 add an Upload() test (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 7996b3f 2015-02-11 less if nesting (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 6241246 2015-02-11 refactor most of `git hawser push` into `hawser client.Upload()` (Rick Olson[32m[m)
[1;36m|[m * [35m|[m ecc617c 2015-02-11 test external Put() (Rick Olson[32m[m)
[1;36m|[m * [35m|[m ede7014 2015-02-11 test Post() (Rick Olson[32m[m)
[1;36m|[m * [35m|[m 28a0c90 2015-02-11 test Get() (Rick Olson[32m[m)
[1;36m|[m * [35m|[m b2a64e5 2015-02-11 test Put() (Rick Olson[32m[m)
[1;36m|[m * [35m|[m acb1e48 2015-02-11 test Options() (Rick Olson[32m[m)
[1;36m|[m * [35m|[m c729aea 2015-02-11 make it easy to replace the credential fetcher in tests (Rick Olson[32m[m)
* [31m|[m [35m|[m   a2e59bf 2015-02-12 Merge pull request #158 from hawser/windows-install-script (Brendan Forster[32m[m)
[31m|[m[33m\[m [31m\[m [35m\[m  
[31m|[m [33m|[m[31m/[m [35m/[m  
[31m|[m[31m/[m[33m|[m [35m|[m   
[31m|[m * [35m|[m b119581 2015-02-11 hide setx output (joshvera[32m[m)
[31m|[m * [35m|[m ac82cd4 2015-02-09 fix setting the path (joshvera[32m[m)
[31m|[m * [35m|[m 1a7d45b 2015-02-09 only set path if we've never created the git media directory (joshvera[32m[m)
[31m|[m * [35m|[m 6c0effe 2015-02-09 fix set syntax (joshvera[32m[m)
[31m|[m * [35m|[m 19e1430 2015-02-09 don't show delete and copy file output (joshvera[32m[m)
[31m|[m * [35m|[m a998c8e 2015-02-09 only mkdir GIT_MEDIA_BIN_PATH if it doesn't exist (joshvera[32m[m)
[31m|[m * [35m|[m b595ae7 2015-02-09 check if git.exe exists (joshvera[32m[m)
[31m|[m * [35m|[m ff54ef6 2015-02-09 no need to delete git-hawser (Josh Vera[32m[m)
[31m|[m * [35m|[m 3cc6182 2015-02-09 auto elevate script (joshvera[32m[m)
[31m|[m * [35m|[m 0d058af 2015-02-06 Strip quotes and copy the right exe (joshvera[32m[m)
[31m|[m * [35m|[m c69f5a9 2015-02-06 dont forget to init (joshvera[32m[m)
[31m|[m * [35m|[m 34b031d 2015-02-06 update script (joshvera[32m[m)
* [33m|[m [35m|[m cbd08bd 2015-02-07 teach script/test how to pass args to `go test` (Rick Olson[32m[m)
* [33m|[m [35m|[m 442d4d3 2015-02-07 update the contributing doc based on the template (Rick Olson[32m[m)
* [33m|[m [35m|[m bf8d037 2015-02-07 sample (Rick Olson[32m[m)
* [33m|[m [35m|[m 16d2621 2015-02-07 move the doc overview to the readme, add a stub contributing.md file (Rick Olson[32m[m)
[33m|[m[33m/[m [35m/[m  
* [35m|[m f01b34d 2015-02-03 check ls-files after committing (risk danger olson[32m[m)
* [35m|[m   349ec30 2015-02-03 Merge pull request #155 from hawser/update-links (risk danger olson[32m[m)
[35m|[m[35m\[m [35m\[m  
[35m|[m [35m|[m[35m/[m  
[35m|[m[35m/[m[35m|[m   
[35m|[m * 1e92865 2015-02-03 Update links to repo (Brandon Keepers[32m[m)
[35m|[m[35m/[m  
* 4367c72 2015-02-02 yay v0.4.0 (Rick Olson[32m (tag: v0.4.0)[m)
* abc0066 2015-02-02 add "rm" alias for "git hawser remove" (Rick Olson[32m[m)
*   609d405 2015-02-02 Merge branch 'enterprise-pre-release' (Rick Olson[32m[m)
[36m|[m[1;31m\[m  
[36m|[m * d77bded 2015-02-01 disable redirects for enterprise support during pre-release (Rick Olson[32m[m)
* [1;31m|[m 3d7a390 2015-02-01 Split add/remove from the path command into top-level commands (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m  
* 1cb3b27 2015-01-31 Fix links in readme (Timothy Clem[32m[m)
* cade8ce 2015-01-31 Use the right url in the release script (rubyist[32m[m)
*   fda5cae 2015-01-31 Merge pull request #151 from hawser/hawserify (Scott Barron[32m (tag: v0.3.6)[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * f73f91b 2015-01-31 First do a 0.3.6 that uses the old media type for compatibility (rubyist[32m[m)
[1;32m|[m * c52c77b 2015-01-30 uprev (rubyist[32m[m)
[1;32m|[m * 97edd11 2015-01-30 GIT_MEDIA_PROGRESS => HAWSER_PROGRESS (rubyist[32m[m)
[1;32m|[m * 50ed75c 2015-01-30 Loosen blobSizeCutoff to account for longer pointer file url (rubyist[32m[m)
[1;32m|[m * 874700d 2015-01-30 A smattering of missed instances (rubyist[32m[m)
[1;32m|[m * f13f15a 2015-01-30 .git/media => .git/hawser (rubyist[32m[m)
[1;32m|[m * ba5ec2a 2015-01-30 filter=media => filter=hawser (rubyist[32m[m)
[1;32m|[m * 9e81fd3 2015-01-30 Update push (rubyist[32m[m)
[1;32m|[m * 82df4a7 2015-01-30 Update ls files (rubyist[32m[m)
[1;32m|[m * de0b6a8 2015-01-30 update pointer version url with https (Rick Olson[32m[m)
[1;32m|[m * 28b13ed 2015-01-30 remove unused gitMediaHeader (rubyist[32m[m)
[1;32m|[m * 928000f 2015-01-30 updating go import paths (Rick Olson[32m[m)
[1;32m|[m * 6cad195 2015-01-30 Go code level changes (rubyist[32m[m)
[1;32m|[m * 56d657f 2015-01-30 Update the media type (rubyist[32m[m)
[1;32m|[m * dda4e5b 2015-01-30 Update some misc pieces (rubyist[32m[m)
[1;32m|[m * a70990a 2015-01-30 update scripts (rubyist[32m[m)
[1;32m|[m * 8911507 2015-01-30 update linux build instructions (rubyist[32m[m)
[1;32m|[m * 0e57111 2015-01-30 Update the other README (rubyist[32m[m)
[1;32m|[m * adaca22 2015-01-30 Update README (rubyist[32m[m)
[1;32m|[m * 98beb8d 2015-01-30 Update man pages (rubyist[32m[m)
[1;32m|[m * 69f427e 2015-01-30 Update pre-push hook (rubyist[32m[m)
[1;32m|[m * c118fd4 2015-01-30 Update version string (rubyist[32m[m)
[1;32m|[m * 1cdb897 2015-01-30 Update the user agent (rubyist[32m[m)
[1;32m|[m * c1afb37 2015-01-30 Rename the command (rubyist[32m[m)
[1;32m|[m * cc6d5a5 2015-01-30 Start rename in the specs (rubyist[32m[m)
[1;32m|[m[1;32m/[m  
*   776a3e9 2015-01-30 Merge pull request #140 from github/redirection-proposal (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 8f844b1 2015-01-13 update callback description (Rick Olson[32m[m)
[1;34m|[m * 47aa1e5 2015-01-13 sketch out a POST request (Rick Olson[32m[m)
[1;34m|[m * 2863e79 2015-01-13 remove OPTIONS response body stuff for backwards compatibility (Rick Olson[32m[m)
[1;34m|[m * da96dd4 2015-01-12 clarify response codes (Rick Olson[32m[m)
[1;34m|[m * ae63207 2015-01-12 describe 405 response (Rick Olson[32m[m)
[1;34m|[m * 67d74a3 2015-01-12 remote put redirection stuff (Rick Olson[32m[m)
[1;34m|[m * 7615629 2015-01-12 flesh out GET and OPTIONS output (Rick Olson[32m[m)
[1;34m|[m * 7422a57 2015-01-08 consistent example urls (Rick Olson[32m[m)
[1;34m|[m * e2cf414 2015-01-08 update with details on PUT redirections (Rick Olson[32m[m)
[1;34m|[m * fc75a0f 2014-12-29 remove custom header for downloads (risk danger olson[32m[m)
[1;34m|[m * 14d2248 2014-12-17 Updates the spec to support redirects for GET and PUT (Rick Olson[32m[m)
* [1;35m|[m   e4efe6c 2015-01-30 Merge pull request #139 from github/tracerup (Scott Barron[32m[m)
[1;36m|[m[31m\[m [1;35m\[m  
[1;36m|[m * [1;35m|[m 9cfbc0f 2014-11-02 Update tracerx (rubyist[32m[m)
[1;36m|[m [1;35m|[m[1;35m/[m  
* [1;35m|[m   576606c 2015-01-30 Merge pull request #150 from github/redirect (Scott Barron[32m[m)
[32m|[m[33m\[m [1;35m\[m  
[32m|[m * [1;35m|[m ea6a5b6 2015-01-29 Consistently use `res` for the http response object name (rubyist[32m[m)
[32m|[m * [1;35m|[m 24939d8 2015-01-27 Accept header for the callback will be provided by the server (rubyist[32m[m)
[32m|[m * [1;35m|[m ffa86bb 2015-01-27 Simplify the pushAsset code (rubyist[32m[m)
[32m|[m * [1;35m|[m 5cb0520 2015-01-26 When getting a 405 on POST, handle it properly so we can fall back to PUT (rubyist[32m[m)
[32m|[m * [1;35m|[m 2c62c37 2015-01-23 Handle the new redirecting API (rubyist[32m[m)
[32m|[m[32m/[m [1;35m/[m  
* [1;35m|[m   e961481 2015-01-16 Merge pull request #149 from github/update-release-link (risk danger olson[32m[m)
[34m|[m[35m\[m [1;35m\[m  
[34m|[m * [1;35m|[m 237f0f5 2015-01-15 Mention install.sh so noobs know to run it (Emily Nakashima[32m[m)
[34m|[m * [1;35m|[m c8eb0cc 2015-01-15 Link to the releases page instead of a release (Emily Nakashima[32m[m)
[34m|[m[34m/[m [1;35m/[m  
* [1;35m|[m   2088d01 2015-01-12 Merge pull request #147 from github/metadata (risk danger olson[32m[m)
[36m|[m[1;31m\[m [1;35m\[m  
[36m|[m * [1;35m|[m 8ce236d 2015-01-12 Meta data => metadata (Lee Reilly[32m[m)
[36m|[m[36m/[m [1;35m/[m  
* [1;35m|[m   cfad541 2015-01-12 Merge pull request #146 from github/fix-man-broken-links (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [1;35m\[m  
[1;32m|[m * [1;35m|[m dd6e146 2015-01-12 Fix broken links to the man pages (Michael Haggerty[32m[m)
[1;32m|[m[1;32m/[m [1;35m/[m  
* [1;35m|[m 45ad21a 2015-01-11 bump (risk danger olson[32m (tag: v0.3.5)[m)
* [1;35m|[m   6efb9c8 2015-01-11 Merge pull request #145 from github/skip-https (risk danger olson[32m[m)
[1;35m|[m[1;35m\[m [1;35m\[m  
[1;35m|[m [1;35m|[m[1;35m/[m  
[1;35m|[m[1;35m/[m[1;35m|[m   
[1;35m|[m * 7a5bd75 2015-01-11 support http.sslVerify too (risk danger olson[32m[m)
[1;35m|[m * f83b028 2014-12-10 skip https verification if GIT_SSL_NO_VERIFY is set (Rick Olson[32m[m)
[1;35m|[m[1;35m/[m  
* 7a08d70 2014-10-29 bump version to 0.3.4 (rubyist[32m (tag: v0.3.4)[m)
*   1efb45f 2014-10-29 Merge pull request #137 from github/status (Scott Barron[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * 9f95d6f 2014-10-28 Scan index with --cached as well, de-dup results, show all statuses (rubyist[32m[m)
[1;36m|[m * 4b681fe 2014-10-28 Simplify current branch/remote finding, fix bug in branch finding, add a status test (rubyist[32m[m)
[1;36m|[m * 58aa383 2014-10-28 Add some code documentation and man pages for status and ls-files (rubyist[32m[m)
[1;36m|[m * faf5102 2014-10-27 Clean up humanizeBytes a bit (rubyist[32m[m)
[1;36m|[m * 27ac222 2014-10-27 Humanize the byte sizes (rubyist[32m[m)
[1;36m|[m * 19a9ccf 2014-10-27 Indicate renames/copies, add parseable output for porcelain (rubyist[32m[m)
[1;36m|[m * aa8cea0 2014-10-27 Show modified files too (rubyist[32m[m)
[1;36m|[m * 7f11591 2014-10-27 Format the output (rubyist[32m[m)
[1;36m|[m * 3b93c39 2014-10-27 Add a scanner for staging (rubyist[32m[m)
[1;36m|[m * 1900c10 2014-10-27 Store the size of the media blob, not the size of the pointer file (rubyist[32m[m)
[1;36m|[m * 4dbb8ac 2014-10-27 Calculate the remote ref rather than hard coding "^origin/HEAD" (which was wrong) (rubyist[32m[m)
[1;36m|[m * 56b8ce0 2014-10-23 Store the size when scanning (rubyist[32m[m)
[1;36m|[m * 5b29a98 2014-10-21 I totally didn't copy and paste that ... (rubyist[32m[m)
[1;36m|[m * d947831 2014-10-21 Start a status command (rubyist[32m[m)
[1;36m|[m[1;36m/[m  
*   673856c 2014-10-20 Merge pull request #136 from github/pushscan (Scott Barron[32m[m)
[32m|[m[33m\[m  
[32m|[m * e0bf70a 2014-10-20 With push using the scanner we don't need link files, and all the associated things (rubyist[32m[m)
[32m|[m * 4f84e07 2014-10-20 Make `push` use the scanner (rubyist[32m[m)
[32m|[m[32m/[m  
*   d221011 2014-10-13 Merge pull request #130 from github/ascannerdarkly (Scott Barron[32m[m)
[34m|[m[35m\[m  
[34m|[m * 788351f 2014-10-13 Document some constants (rubyist[32m[m)
[34m|[m * 90da729 2014-10-11 Clean up a few code things (rubyist[32m[m)
[34m|[m * 95b2408 2014-10-09 These are no longer used, removing (rubyist[32m[m)
[34m|[m * 2582e3e 2014-10-09 Remove the `scan` command (rubyist[32m[m)
[34m|[m * 47e134f 2014-10-08 Give just a little more room for the size comparison (rubyist[32m[m)
[34m|[m * 533eb1f 2014-10-08 --create-links for scan (rubyist[32m[m)
[34m|[m * 921ffd8 2014-10-08 Track file names when scanning files (rubyist[32m[m)
[34m|[m * 38dbac3 2014-10-07 Handle some errors, buffer some channels (rubyist[32m[m)
[34m|[m *   5ee311f 2014-10-07 Merge branch 'master' into ascannerdarkly (rubyist[32m[m)
[34m|[m [36m|[m[34m\[m  
[34m|[m [36m|[m[34m/[m  
[34m|[m[34m/[m[36m|[m   
* [36m|[m 7816cd5 2014-10-06 bump version (rubyist[32m (tag: v0.3.3)[m)
* [36m|[m   cba64af 2014-10-06 Merge pull request #134 from github/absfabpath (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [36m\[m  
[1;32m|[m * [36m|[m 2fd52c3 2014-10-06 Use an absolute path for GIT_MEDIA_PROGRESS rather than a file:/// URI (rubyist[32m[m)
[1;32m|[m[1;32m/[m [36m/[m  
* [36m|[m   1a0edcc 2014-10-02 Merge pull request #132 from github/removelock (Scott Barron[32m (tag: v0.3.2)[m)
[1;34m|[m[1;35m\[m [36m\[m  
[1;34m|[m * [36m|[m acbf20f 2014-10-02 Bump version (rubyist[32m[m)
[1;34m|[m * [36m|[m 91d14de 2014-10-02 Remove the original file before mv'ing to it, Windows errors if the file exists (rubyist[32m[m)
[1;34m|[m[1;34m/[m [36m/[m  
[1;34m|[m * ef31750 2014-10-07 Add some tracing to the scanner (rubyist[32m[m)
[1;34m|[m * 6cda0a7 2014-10-07 When scanner is given a ref, it should not traverse the ancesorts Doing so will e.g. show deleted files in `git ls-files` (rubyist[32m[m)
[1;34m|[m * 014d35e 2014-10-07 Show names again on ls-files (rubyist[32m[m)
[1;34m|[m * 5ccda38 2014-10-07 Add some documentation (rubyist[32m[m)
[1;34m|[m * 6503061 2014-10-07 Buffer subprocess stdouts (rubyist[32m[m)
[1;34m|[m * d0a174f 2014-10-07 Extract subcommand creation (rubyist[32m[m)
[1;34m|[m * a0ff519 2014-10-07 Rework scanner with channels, saves huge memory. Needs clean up and error handling. (rubyist[32m[m)
[1;34m|[m * d68ae59 2014-10-05 Add some performance tracing to the scanner (rubyist[32m[m)
[1;34m|[m * 084655b 2014-10-04 try this again (rubyist[32m[m)
[1;34m|[m * d73ac37 2014-10-04 wat (rubyist[32m[m)
[1;34m|[m * 46841cc 2014-10-04 Stricter parsing of OID in old alpha pointer formats The old pointer format was pretty loose, this can cause false positives for the scanner. An example can be found in this repository, docs/man/index.txt will be considered a git media file. Validing that the line that should be an OID is an OID should reduce these false positives. (rubyist[32m[m)
[1;34m|[m * 2ba6ccf 2014-10-04 Use tracerx (rubyist[32m[m)
[1;34m|[m * 3b13dbf 2014-10-03 ls-files will use a scanner and honor the current branch, or take a branch on the command line (rubyist[32m[m)
[1;34m|[m * 91b0999 2014-10-03 Make scanner its own package (rubyist[32m[m)
[1;34m|[m * d669f77 2014-10-03 Add a tracer that honors GIT_TRACE (rubyist[32m[m)
[1;34m|[m * 494a547 2014-10-03 Avoid shelling out once per object, things go much faster (rubyist[32m[m)
[1;34m|[m * fa922d7 2014-10-02 Step 1: Write a crude scanner that finds all git media files (rubyist[32m[m)
[1;34m|[m[1;34m/[m  
*   7c07c70 2014-10-01 Merge pull request #129 from github/pushbail (Scott Barron[32m (tag: v0.3.1)[m)
[1;36m|[m[31m\[m  
[1;36m|[m * ba14b6e 2014-10-01 Bump version (rubyist[32m[m)
[1;36m|[m * 45d77fd 2014-10-01 Update test to meet arg requirements (rubyist[32m[m)
[1;36m|[m * a0601f0 2014-10-01 If we receive no command line args, the pre-push hook is out of date. Bail. (rubyist[32m[m)
[1;36m|[m[1;36m/[m  
*   8a9d65a 2014-10-01 Merge pull request #122 from github/smarterpush (Scott Barron[32m (tag: v0.3.0)[m)
[32m|[m[33m\[m  
[32m|[m * ae5d667 2014-10-01 Ensure that `git media push` called on its own doesn't just sit there waiting for stdin. (rubyist[32m[m)
[32m|[m * 344f0e9 2014-10-01 Ensure we're testing with the bin in the repo and not any installed bins (rubyist[32m[m)
[32m|[m * 6d7460f 2014-09-30 This should have a ^ (rubyist[32m[m)
[32m|[m * e6b143a 2014-09-29 Using ls-remote for media push --dry-run is more accurate (rubyist[32m[m)
[32m|[m * 321c1a5 2014-09-29 Make sure git-media is in the path for tests (rubyist[32m[m)
[32m|[m * dba825f 2014-09-29 Add some more push tests (rubyist[32m[m)
[32m|[m * 11128a8 2014-09-29 Add a repo.GitCmd, tidy up ls-files test (rubyist[32m[m)
[32m|[m * 67e8f72 2014-09-26 basic ls-files test (rubyist[32m[m)
[32m|[m * 0491067 2014-09-26 Ensure link files are created atomically (rubyist[32m[m)
[32m|[m * be349c5 2014-09-24 Add some docs (rubyist[32m[m)
[32m|[m * 141bd6f 2014-09-24 Make --dry-run a bit smarter and more useful (rubyist[32m[m)
[32m|[m * 17ce00d 2014-09-24 This should just be called HashObject (rubyist[32m[m)
[32m|[m * 49c9c5b 2014-09-23 Provide a `git media update` command that updates the pre-push hook and migrates upload queue (rubyist[32m[m)
[32m|[m * 5222604 2014-09-23 RevListObjects returns real objects (rubyist[32m[m)
[32m|[m * b925e87 2014-09-23 Unbreak ls-files (rubyist[32m[m)
[32m|[m * e5354bc 2014-09-23 extract linksFromRefs, simplify pushCommand (rubyist[32m[m)
[32m|[m * b850e5e 2014-09-23 extract decodeRefs, simplify pushCommand (rubyist[32m[m)
[32m|[m * edd94a3 2014-09-23 Consolidate pointer link functions, simplify push (rubyist[32m[m)
[32m|[m * 493808b 2014-09-23 Move git config runner to git/, simpleExec takes a stdin reader, use simpleExec for everything (rubyist[32m[m)
[32m|[m * 781d8bb 2014-09-23 Extract some git commands, start cleaning up push (rubyist[32m[m)
[32m|[m * 0e217df 2014-09-23 kill the queue (rubyist[32m[m)
[32m|[m * 069e96f 2014-09-22 git media ls-files command This will trawl through .git/media/objects and list out the files we know about. (rubyist[32m[m)
[32m|[m * d8ade10 2014-09-22 Write link files when smudging (rubyist[32m[m)
[32m|[m * 567907c 2014-09-22 Handle a couple errors around link files (rubyist[32m[m)
[32m|[m * 1be35b2 2014-09-22 Need to append to sub-process environment, not overwrite it. If the subprocess then launches any commands itself, it will have no environment. e.g. PATH, GOPATH, etc are all gone. (rubyist[32m[m)
[32m|[m * 60f43a9 2014-09-22 If the pre-push OPTIONS returns a 200, the server has the file, don't push it again. (rubyist[32m[m)
[32m|[m * dfbbd0f 2014-09-19 Tests do catch my horrible errors (rubyist[32m[m)
[32m|[m * 53aaad1 2014-09-19 Store filename in pointer link, display during push (rubyist[32m[m)
[32m|[m * e9092a8 2014-09-19 When deleting a branch, don't send any objects (rubyist[32m[m)
[32m|[m * 0d6b35d 2014-09-18 Add a --dry-run flag This only works when running `git media push`, I dont' think we can get any args down to the pre-push hook from `git push` (rubyist[32m[m)
[32m|[m * af7f92c 2014-09-18 Handle new branch case (rubyist[32m[m)
[32m|[m * 18fbfbb 2014-09-18 pre-push needs to pass its command line args to git media push (rubyist[32m[m)
[32m|[m * c8e49b3 2014-09-18 Error prone but functioning happy path. Doesn't handle many cases yet (rubyist[32m[m)
[32m|[m * 326790e 2014-09-18 what even are tests anyway: :/ (rubyist[32m[m)
[32m|[m * 0375015 2014-09-18 Don't queue the file when cleaning (rubyist[32m[m)
[32m|[m * 52a139a 2014-09-18 Create the link file when cleaning (rubyist[32m[m)
[32m|[m * adfe84c 2014-09-18 A git hasher that calculates the git sha1 for a stream of data (rubyist[32m[m)
[32m|[m[32m/[m  
* 3dc9a64 2014-09-15 v0.2.3 (rubyist[32m (tag: v0.2.3)[m)
*   256dcff 2014-09-15 Merge pull request #120 from github/useragent (Scott Barron[32m[m)
[34m|[m[35m\[m  
[34m|[m * 6d7cb11 2014-09-15 Add vendor info to useragent string (rubyist[32m[m)
[34m|[m * c04feb7 2014-09-15 Include architecture in useragent string (rubyist[32m[m)
[34m|[m * 4b103e6 2014-09-15 Update user agent string Example: git-media/0.2.2 (darwin; git 1.9.1; go 1.3) (rubyist[32m[m)
[34m|[m[34m/[m  
*   e7fc3ef 2014-09-15 Merge pull request #119 from github/ensureprefix (Scott Barron[32m[m)
[36m|[m[1;31m\[m  
[36m|[m * fe6bba2 2014-09-15 Ensure the installation prefix exists before installing into it (rubyist[32m[m)
* [1;31m|[m   7ba9249 2014-09-15 Merge pull request #118 from github/credsfix (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m [1;31m\[m  
[1;32m|[m * [1;31m|[m ebf3be7 2014-09-15 It's easier if we just don't hook stderr up at all, since we can never use it anyway (rubyist[32m[m)
[1;32m|[m * [1;31m|[m 4fad516 2014-09-11 Only provide stdout/stderr buffers for the fill credentials subcommand There is a bug in the way the git credential helpers launch the daemon if it is not already running such that the stderr of the grandchild does not appear to be getting closed, causing the git media client to not receive EOF on the pipe and wait forever. (rubyist[32m[m)
[1;32m|[m [1;31m|[m[1;31m/[m  
* [1;31m|[m a3f1578 2014-09-09 v0.2.2 (Rick Olson[32m (tag: v0.2.2)[m)
* [1;31m|[m   63ba097 2014-09-09 Merge pull request #114 from github/user-agent (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [1;31m\[m  
[1;34m|[m * [1;31m|[m 4416cd0 2014-09-09 fix version tests (Rick Olson[32m[m)
[1;34m|[m * [1;31m|[m c64db8b 2014-09-09 show the git version with the git media version (Rick Olson[32m[m)
[1;34m|[m * [1;31m|[m 8adbef6 2014-09-09 create a UserAgent var, use it in the version command and User-Agent (Rick Olson[32m[m)
[1;34m|[m [1;31m|[m[1;31m/[m  
* [1;31m|[m   2d97253 2014-09-09 Merge pull request #115 from github/fix-release (risk danger olson[32m[m)
[1;31m|[m[31m\[m [1;31m\[m  
[1;31m|[m [31m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[31m|[m   
[1;31m|[m * 86fed57 2014-09-09 Change -d to --data-binary (Rick Olson[32m[m)
[1;31m|[m[1;31m/[m  
* eb22d74 2014-09-08 forgot to push the actual version bump (Rick Olson[32m[m)
*   6043db4 2014-09-08 Merge pull request #113 from github/script-release (risk danger olson[32m[m)
[32m|[m[33m\[m  
[32m|[m * 6c2f61e 2014-09-05 build.go and release.go get compiled together (Rick Olson[32m[m)
[32m|[m * 4c422b4 2014-09-05 add a release script (Rick Olson[32m[m)
[32m|[m * a363a65 2014-09-05 real exit codes for script/bootstrap (Rick Olson[32m[m)
[32m|[m * 14f0f1d 2014-09-05 dump a build_matrix.json after cross compiling (Rick Olson[32m[m)
[32m|[m[32m/[m  
*   d5255d7 2014-09-05 Merge pull request #112 from github/smudgepassthrough (Scott Barron[32m (tag: v0.2.1)[m)
[34m|[m[35m\[m  
[34m|[m * 0255d1f 2014-09-05 Check for an error on Copy (rubyist[32m[m)
[34m|[m * 56a6029 2014-09-05 If decoding pointer file errors, pass data to stdout (rubyist[32m[m)
[34m|[m[34m/[m  
*   26fe77e 2014-09-03 Merge pull request #111 from github/windows-root-path (risk danger olson[32m (tag: v0.2.1-p1)[m)
[36m|[m[1;31m\[m  
[36m|[m * 39271ac 2014-09-03 the root directory always ends in a path separator (Brendan Forster[32m[m)
[36m|[m[36m/[m  
* 690c8da 2014-08-21 ignore close errors (Rick Olson[32m (tag: v0.2.0)[m)
* 84e0690 2014-08-20 cut v0.2.0 release (Rick Olson[32m[m)
*   10badbc 2014-08-19 Merge pull request #106 from github/accept (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * fc5d6e0 2014-08-19 Fix reversed logic in error handling, call Accept (rubyist[32m[m)
[1;32m|[m[1;32m/[m  
*   c957761 2014-08-18 Merge pull request #105 from github/contentaddressable (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 8a862f3 2014-08-18 use go-contentaddressable when smudging media content (Rick Olson[32m[m)
[1;34m|[m * a4354c5 2014-08-18 add go-contentaddressable files (Rick Olson[32m[m)
[1;34m|[m * 196b026 2014-08-18 add go-contentaddressable (Rick Olson[32m[m)
[1;34m|[m[1;34m/[m  
*   4aee116 2014-08-18 Merge pull request #100 from github/progress-callback (risk danger olson[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * bdb20ba 2014-08-14 dont send initial line marking 0 bytes (Rick Olson[32m[m)
[1;36m|[m * 87b7552 2014-08-14 require a uri scheme for GIT_MEDIA_PROGRESS (Rick Olson[32m[m)
[1;36m|[m * 36fa23d 2014-08-14 Send real stats in the progress callback instead of % (Rick Olson[32m[m)
* [31m|[m   5338630 2014-08-18 Merge pull request #101 from github/pointer-v3 (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m  
[32m|[m * [31m|[m e4b0a21 2014-08-14 more strict key parsing (Rick Olson[32m[m)
[32m|[m * [31m|[m 825cf6d 2014-08-14 pointer files need git-media somewhere (Rick Olson[32m[m)
[32m|[m * [31m|[m 64b8331 2014-08-14 simplify pointer format (Rick Olson[32m[m)
[32m|[m [31m|[m[31m/[m  
* [31m|[m d6879cf 2014-08-16 /cc @pengwynn (risk danger olson[32m[m)
[31m|[m[31m/[m  
* ad57cf7 2014-08-12 fix copy pasta issue (Rick Olson[32m[m)
* 9ac7816 2014-08-12 buffer, bub (Rick Olson[32m[m)
* 8ca48f8 2014-08-12 write progress to the callback while downloading a file (Rick Olson[32m[m)
* 23b40ba 2014-08-12 sync progress file writes (Rick Olson[32m[m)
*   bdcc42c 2014-08-08 Merge pull request #98 from github/windows-fixes (risk danger olson[32m[m)
[34m|[m[35m\[m  
[34m|[m * 1e2e976 2014-08-08 hide Authorization headers from logs (Rick Olson[32m[m)
[34m|[m * cb9a80e 2014-08-08 Don't cast a *gitmedia.WrappedError to an error (Rick Olson[32m[m)
[34m|[m * f69546c 2014-08-08 use gitmedia.WrappedError (Rick Olson[32m[m)
[34m|[m * a80796d 2014-08-08 revert this unintended change (Rick Olson[32m[m)
[34m|[m * c6635d8 2014-08-08 keep using filepath.Base() on these **fs paths** (Rick Olson[32m[m)
[34m|[m * d530af1 2014-08-08 build the path manually (Rick Olson[32m[m)
[34m|[m * 4bc848d 2014-08-08 use path.Join() since we're not building fs paths (Rick Olson[32m[m)
[34m|[m * a3a665d 2014-08-08 way more response context (Rick Olson[32m[m)
[34m|[m * 890a42c 2014-08-08 add an error context (Rick Olson[32m[m)
[34m|[m * beccfaa 2014-08-08 bubble the stack all the way up (Rick Olson[32m[m)
[34m|[m * a724373 2014-08-08 don't reject auth for upstream service issues (Rick Olson[32m[m)
[34m|[m * f94c1fd 2014-08-07 use a gitmedia wrapped error for boomtown (Rick Olson[32m[m)
[34m|[m * b24232a 2014-08-07 log the WrappedError's stack if available (Rick Olson[32m[m)
[34m|[m * 055ec90 2014-08-07 wrap client errors so we can assign more context (like the stack) (Rick Olson[32m[m)
[34m|[m * f960932 2014-08-07 log smudge errors (Rick Olson[32m[m)
[34m|[m * a3795ff 2014-08-07 close the temp file after cleaning (Rick Olson[32m[m)
[34m|[m * 212f3a2 2014-08-07 remove punctuation from log filenames (Rick Olson[32m[m)
[34m|[m * a2b9a06 2014-08-07 log the error that was returned while trying to log an error (Rick Olson[32m[m)
[34m|[m * e38d42f 2014-08-07 write panic to stderr if it can't write to the file (Rick Olson[32m[m)
[34m|[m[34m/[m  
*   3cf4399 2014-08-07 Merge pull request #97 from github/filter-progress (risk danger olson[32m[m)
[36m|[m[1;31m\[m  
[36m|[m * 0f81693 2014-08-07 send the correct index to push actions (Rick Olson[32m[m)
[36m|[m * df2c5a9 2014-08-07 show the current index and total number of files of the current operation (Rick Olson[32m[m)
[36m|[m * defae1c 2014-08-07 implement a copy callback for pushing files (Rick Olson[32m[m)
[36m|[m * d7a787d 2014-08-07 use a callback reader instead (Rick Olson[32m[m)
[36m|[m * f5298c2 2014-08-07 extract a smaller pushAsset() function (Rick Olson[32m[m)
[36m|[m * 0392ea4 2014-08-07 implement progress logs for clean (Rick Olson[32m[m)
[36m|[m * d349fa0 2014-08-07 encapsulate what makes a good callback func for io.Copy (Rick Olson[32m[m)
[36m|[m * 25fa6f5 2014-08-07 unnecessary, empty string is already the default (Rick Olson[32m[m)
[36m|[m * ff583eb 2014-08-07 guess the size if it's not provided by the pointer (Rick Olson[32m[m)
[36m|[m * c362a81 2014-08-07 add the event to the front of each line (Rick Olson[32m[m)
[36m|[m * b1582de 2014-08-07 write the percentage first (Rick Olson[32m[m)
[36m|[m * 101cc6e 2014-08-07 don't send duplicate progress events (Rick Olson[32m[m)
[36m|[m * b954a1e 2014-08-07 put the linebreaks at the end (Rick Olson[32m[m)
[36m|[m * dc464ed 2014-08-07 update the test to use a pointer file that specifies the size (Rick Olson[32m[m)
[36m|[m * 203a86e 2014-08-07 handle cases where we dont know the total (Rick Olson[32m[m)
[36m|[m * 0204cb2 2014-08-07 almost complete the logging, lol %s (Rick Olson[32m[m)
[36m|[m * fd41906 2014-08-06 add CopyWithCallback() (Rick Olson[32m[m)
[36m|[m * 98b216d 2014-08-06 Smudge takes a *Pointer instead of just an Oid string (Rick Olson[32m[m)
[36m|[m[36m/[m  
*   a2aba3a 2014-08-06 Merge pull request #96 from github/force-hook-dir (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * f92767f 2014-08-06 create a .git/hooks dir if it doesn't exist (Rick Olson[32m[m)
[1;32m|[m[1;32m/[m  
*   8a55f8c 2014-08-05 Merge pull request #94 from github/reliable-smudge (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 5c43ba9 2014-08-05 better error messages when smudge fails (Rick Olson[32m[m)
[1;34m|[m * d2b824c 2014-08-05 better cleanup (Rick Olson[32m[m)
[1;34m|[m * 5f928b2 2014-08-05 use os rename to write to .git/media with consistency (Rick Olson[32m[m)
[1;34m|[m * c6bfb6f 2014-08-05 the gitmedia package can create temp files in the .git/media dir (Rick Olson[32m[m)
[1;34m|[m * 7e5ff6b 2014-07-28 don't write the media file until the last minute (Rick Olson[32m[m)
[1;34m|[m * 1125c63 2014-07-28 gitmediaclient doesn't need to check .git/media dir (Rick Olson[32m[m)
[1;34m|[m * 5fbaa6c 2014-07-28 write to a temp file first before streaming to stdout (Rick Olson[32m[m)
[1;34m|[m * 5d53adb 2014-07-28 approve credentials after successful requests (Rick Olson[32m[m)
[1;34m|[m * da8001c 2014-07-28 be more explicit about where the error came from (Rick Olson[32m[m)
[1;34m|[m * 1b0762c 2014-07-28 cleanup the opened media file (Rick Olson[32m[m)
[1;34m|[m * 65a021c 2014-07-28 if smudge errors, write the pointer file to stdout (Rick Olson[32m[m)
[1;34m|[m * a52a287 2014-07-28 don't shadow `err` (Rick Olson[32m[m)
[1;34m|[m * db8dd88 2014-07-28 be sure to cleanup tempfile if there's an error (Rick Olson[32m[m)
[1;34m|[m * a1a99f3 2014-07-28 use the consistent file writer in Smudge() (Rick Olson[32m[m)
[1;34m|[m * 2e06323 2014-07-28 add a consistent file writer (Rick Olson[32m[m)
[1;34m|[m * 8fb3928 2014-07-28 oid, not sha (Rick Olson[32m[m)
[1;34m|[m * 0a19587 2014-07-28 break smudge into downloadFile() and readLocalFile() (Rick Olson[32m[m)
[1;34m|[m[1;34m/[m  
* 53e5ed6 2014-07-28 dont expose pointer.CleanedAsset (Rick Olson[32m[m)
* 1235797 2014-07-28 combine metafile and filters packages in a pointer package (Rick Olson[32m[m)
*   59db431 2014-07-25 Merge pull request #93 from github/smudge-info (risk danger olson[32m[m)
[1;36m|[m[31m\[m  
[1;36m|[m * bb111b8 2014-07-25 remove unnecessary var (risk danger olson[32m[m)
[1;36m|[m *   21548f5 2014-07-25 merge master (risk danger olson[32m[m)
[1;36m|[m [32m|[m[1;36m\[m  
[1;36m|[m [32m|[m[1;36m/[m  
[1;36m|[m[1;36m/[m[32m|[m   
* [32m|[m   c394a46 2014-07-25 Merge pull request #87 from github/versioned-pointer (risk danger olson[32m[m)
[34m|[m[35m\[m [32m\[m  
[34m|[m [35m|[m * 0737611 2014-07-25 use the real file size where possible (risk danger olson[32m[m)
[34m|[m [35m|[m * d4a254e 2014-07-25 add `smudge --info` (risk danger olson[32m[m)
[34m|[m [35m|[m * 6b8b82d 2014-07-25 document smudge --info (risk danger olson[32m[m)
[34m|[m [35m|[m[35m/[m  
[34m|[m * 5120f93 2014-07-24 write the new pointer format (Rick Olson[32m[m)
[34m|[m * 2864673 2014-07-24 more pointer file rules in the spec (Rick Olson[32m[m)
[34m|[m * 7beeca9 2014-07-24 some failing tests for the ini format (Rick Olson[32m[m)
[34m|[m * a2e5dae 2014-07-24 re-arrange pointer decoding tests (Rick Olson[32m[m)
[34m|[m * 6e20e28 2014-07-24 add ini decoding (Rick Olson[32m[m)
[34m|[m * 75388f0 2014-07-24 add goini lib (Rick Olson[32m[m)
[34m|[m *   d733e35 2014-07-24 Merge branch 'master' into versioned-pointer (Rick Olson[32m[m)
[34m|[m [36m|[m[34m\[m  
[34m|[m [36m|[m[34m/[m  
[34m|[m[34m/[m[36m|[m   
* [36m|[m a78c4e2 2014-07-24 update cobra (Rick Olson[32m[m)
* [36m|[m a2d7a8b 2014-07-24 update pflag (Rick Olson[32m[m)
* [36m|[m 18b2f00 2014-07-24 update kr/pretty (Rick Olson[32m[m)
* [36m|[m 8f15228 2014-07-24 update pb (Rick Olson[32m[m)
* [36m|[m 1b70201 2014-07-24 update cobra (Rick Olson[32m[m)
[1;31m|[m * 5b89247 2014-07-24 split out alpha pointer decoder (Rick Olson[32m[m)
[1;31m|[m * f72a2eb 2014-07-24 metafile.Encode takes a *metafile.Pointer (Rick Olson[32m[m)
[1;31m|[m * a31b736 2014-07-24 metafile.Decode now returns a *metafile.Pointer (Rick Olson[32m[m)
[1;31m|[m *   49f6bf4 2014-07-24 Merge branch 'master' into versioned-pointer (Rick Olson[32m[m)
[1;31m|[m [1;32m|[m[1;31m\[m  
[1;31m|[m [1;32m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[1;32m|[m   
* [1;32m|[m   81ba22e 2014-07-24 Merge branch 'cobraaaaa' (Rick Olson[32m[m)
[1;34m|[m[1;35m\[m [1;32m\[m  
[1;34m|[m * [1;32m|[m 3d43579 2014-06-26 use built-in Print() (Rick Olson[32m[m)
* [1;35m|[m [1;32m|[m   e2dc929 2014-07-24 Merge pull request #91 from github/man (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;35m\[m [1;32m\[m  
[1;36m|[m * [1;35m|[m [1;32m|[m 85e48a3 2014-07-24 git addddd (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;32m|[m 774ff46 2014-07-24 we need that index.txt file too (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;32m|[m f6b7888 2014-07-24 treat ./man as a dir for built man/html docs (Rick Olson[32m[m)
[1;36m|[m * [1;35m|[m [1;32m|[m a764b20 2014-07-24 move ronn files to docs/man (Rick Olson[32m[m)
[1;36m|[m[1;36m/[m [1;35m/[m [1;32m/[m  
* [1;35m|[m [1;32m|[m   c2546a2 2014-07-24 Merge pull request #81 from github/paths-v-path (risk danger olson[32m[m)
[32m|[m[33m\[m [1;35m\[m [1;32m\[m  
[32m|[m * [1;35m|[m [1;32m|[m fa7c119 2014-06-26 Kill the s :fire: (Mike Skalnik[32m[m)
[32m|[m * [1;35m|[m [1;32m|[m e9edcf5 2014-06-26 Be consistent about path vs paths (Mike Skalnik[32m[m)
* [33m|[m [1;35m|[m [1;32m|[m   c128fe9 2014-07-24 Merge pull request #84 from github/cobraaaaa (risk danger olson[32m[m)
[33m|[m[1;35m\[m [33m\[m [1;35m\[m [1;32m\[m  
[33m|[m [1;35m|[m[33m/[m [1;35m/[m [1;32m/[m  
[33m|[m[33m/[m[1;35m|[m [1;35m/[m [1;32m/[m   
[33m|[m [1;35m|[m[1;35m/[m [1;32m/[m    
[33m|[m * [1;32m|[m 989f831 2014-06-26 split logs into 5 commands (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 768b290 2014-06-26 split init into 2 commands (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 53b43b8 2014-06-26 split path into 3 commands (Rick Olson[32m[m)
[33m|[m * [1;32m|[m d9b5475 2014-06-26 remove redundant 'paths' command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 25b3487 2014-06-26 no periods in the USAGE (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 9f3a7bf 2014-06-26 Usage() has one less line break (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 9e95c6e 2014-06-26 root cmd shows help and version (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 4c27931 2014-06-26 is this all that changed (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 726ca34 2014-06-26 remove old command crap (Rick Olson[32m[m)
[33m|[m * [1;32m|[m b731bed 2014-06-26 update smudge command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 46426ee 2014-06-26 update queues command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 9799953 2014-06-26 update push command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 9abcbde 2014-06-26 update paths command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 05561b6 2014-06-26 update logs command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m c421039 2014-06-26 update the init command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 35cfe51 2014-06-26 update clean command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 57ef4f8 2014-06-26 dont run old commands (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 42e0177 2014-06-26 update the env command (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 7c5cd5f 2014-06-26 update version to use cobra (Rick Olson[32m[m)
[33m|[m * [1;32m|[m 5ac1854 2014-06-26 import cobra (Rick Olson[32m[m)
[33m|[m[33m/[m [1;32m/[m  
[33m|[m * eda60cc 2014-07-21 update spec with new pointer format (risk danger olson[32m[m)
[33m|[m[33m/[m  
*   4843542 2014-06-25 Merge pull request #80 from github/doc-review (risk danger olson[32m (tag: v0.1.0)[m)
[36m|[m[1;31m\[m  
[36m|[m * 6c08210 2014-06-24 prepare for v0.1.0 tag (Rick Olson[32m[m)
[36m|[m * b291e22 2014-06-24 more slight doc tweaks (Rick Olson[32m[m)
[36m|[m * 87bd98b 2014-06-24 no one cares about this fake blog post I wrote (Rick Olson[32m[m)
[36m|[m * 72e36d5 2014-06-24 run 'git media init' during installation (Rick Olson[32m[m)
[36m|[m[36m/[m  
*   7b22098 2014-06-24 Merge pull request #79 from github/doc-update (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * 5ceb05c 2014-06-24 Insert a newline for better output formatting (rubyist[32m[m)
[1;32m|[m * edd8a77 2014-06-24 Update docs to remove references to `sync`, show `push` (rubyist[32m[m)
[1;32m|[m[1;32m/[m  
*   ec88a4e 2014-06-16 Merge pull request #76 from github/submodules (Scott Barron[32m[m)
[1;34m|[m[1;35m\[m  
[1;34m|[m * 1675064 2014-06-16 Absolute paths for submodule env output (rubyist[32m[m)
[1;34m|[m *   dffbfc9 2014-06-13 Merge branch 'master' into submodules (rubyist[32m[m)
[1;34m|[m [1;36m|[m[31m\[m  
[1;34m|[m * [31m|[m 928b412 2014-06-13 Avoid reading the entire file, just in case (rubyist[32m[m)
[1;34m|[m * [31m|[m 2a0df64 2014-06-13 Don't need the dots (rubyist[32m[m)
[1;34m|[m * [31m|[m 55a13ef 2014-06-13 Test that .gitconfig functions as expected in submodules (rubyist[32m[m)
[1;34m|[m * [31m|[m d5b646f 2014-06-11 Process .git files that are not directories In some cases, .git is not a directory but is a file pointing to the real .git directory. This is true in the case of submodules, and may be true for other odd set ups. (rubyist[32m[m)
[1;34m|[m * [31m|[m ece4a01 2014-06-11 Iteration order of maps is randomized in Go, building an output string like this can sometimes fail (rubyist[32m[m)
* [31m|[m [31m|[m   bbe946c 2014-06-16 Merge pull request #78 from github/linux-build-docs (risk danger olson[32m[m)
[32m|[m[33m\[m [31m\[m [31m\[m  
[32m|[m * [31m|[m [31m|[m 62f752f 2014-06-14 :lipstick: (Sergio Rubio[32m[m)
[32m|[m * [31m|[m [31m|[m 28a8f06 2014-06-14 Added man page build/install docs (Sergio Rubio[32m[m)
[32m|[m * [31m|[m [31m|[m cd98de2 2014-06-14 Quick build instructions for Linux (Sergio Rubio[32m[m)
[32m|[m [31m|[m [31m|[m[31m/[m  
[32m|[m [31m|[m[31m/[m[31m|[m   
* [31m|[m [31m|[m   5cf91b1 2014-06-14 Merge pull request #77 from github/bin-bash-test (risk danger olson[32m[m)
[31m|[m[35m\[m [31m\[m [31m\[m  
[31m|[m [35m|[m[31m/[m [31m/[m  
[31m|[m[31m/[m[35m|[m [31m|[m   
[31m|[m * [31m|[m 2c59064 2014-06-14 Remove bashisms from script/man (Sergio Rubio[32m[m)
[31m|[m * [31m|[m 1f52205 2014-06-14 Fix Bashisms (Sergio Rubio[32m[m)
[31m|[m[31m/[m [31m/[m  
* [31m|[m   c00f4f9 2014-06-12 Merge pull request #72 from github/git-media-tmp (risk danger olson[32m[m)
[36m|[m[1;31m\[m [31m\[m  
[36m|[m * [31m|[m c5e3786 2014-06-05 put git-media's tmp in the media dir (Rick Olson[32m[m)
* [1;31m|[m [31m|[m 485ab44 2014-06-10 see the man pages instead (risk danger olson[32m[m)
* [1;31m|[m [31m|[m 7191ec6 2014-06-10 Update README.md (risk danger olson[32m[m)
* [1;31m|[m [31m|[m 41dfce3 2014-06-10 Update README.md (risk danger olson[32m[m)
[31m|[m [1;31m|[m[31m/[m  
[31m|[m[31m/[m[1;31m|[m   
* [1;31m|[m   86de3ce 2014-06-09 Merge pull request #63 from github/error-cleanup (Scott Barron[32m[m)
[1;31m|[m[1;33m\[m [1;31m\[m  
[1;31m|[m [1;33m|[m[1;31m/[m  
[1;31m|[m[1;31m/[m[1;33m|[m   
[1;31m|[m *   7ee80a7 2014-06-09 Merge pull request #69 from github/error-cleanup-merged (Scott Barron[32m[m)
[1;31m|[m [1;34m|[m[1;35m\[m  
[1;31m|[m [1;34m|[m * bdee457 2014-06-05 doh, fix errors (Rick Olson[32m[m)
[1;31m|[m [1;34m|[m *   40d2731 2014-06-05 merge master (Rick Olson[32m[m)
[1;31m|[m [1;34m|[m [1;34m|[m[31m\[m  
[1;31m|[m [1;34m|[m[1;34m/[m [31m/[m  
[1;31m|[m * [31m|[m b901903 2014-06-05 ci doesn't have cover installed (rubyist[32m[m)
[1;31m|[m * [31m|[m 022449b 2014-06-05 Show some test coverage stats (rubyist[32m[m)
[1;31m|[m * [31m|[m 27f615f 2014-06-05 Restore showing file names for `queues` and `push`. I think this got changed with unifying things into one binary (rubyist[32m[m)
[1;31m|[m * [31m|[m 2862ab1 2014-06-04 Be friendlier about panicing, don't mask the error, control exit code ourselves (rubyist[32m[m)
[1;31m|[m * [31m|[m   82df9fe 2014-06-04 Merge branch 'master' into error-cleanup (rubyist[32m[m)
[1;31m|[m [32m|[m[33m\[m [31m\[m  
[1;31m|[m * [33m|[m [31m|[m d1f05fb 2014-06-04 If a meta file heater was present but no sha, a panic would occur. Return an error instead. (rubyist[32m[m)
[1;31m|[m * [33m|[m [31m|[m f1276a1 2014-06-04 If the metafile can't be decoded, return an error (rubyist[32m[m)
[1;31m|[m * [33m|[m [31m|[m b3227b2 2014-06-04 If there are no logs to show this code would panic. Show a message to the user instead. (rubyist[32m[m)
* [33m|[m [33m|[m [31m|[m   4180e06 2014-06-05 Merge pull request #71 from github/test-with-submodules (risk danger olson[32m[m)
[34m|[m[35m\[m [33m\[m [33m\[m [31m\[m  
[34m|[m * [33m|[m [33m|[m [31m|[m c30183a 2014-06-05 test that submodules don't mess up the env (Rick Olson[32m[m)
[34m|[m * [33m|[m [33m|[m [31m|[m 167ab85 2014-06-05 update repo cleaning scripts to clean submodules (Rick Olson[32m[m)
[34m|[m[34m/[m [33m/[m [33m/[m [31m/[m  
* [33m|[m [33m|[m [31m|[m   91851e3 2014-06-05 Merge pull request #70 from github/add-ts-package (risk danger olson[32m[m)
[31m|[m[1;31m\[m [33m\[m [33m\[m [31m\[m  
[31m|[m [1;31m|[m[31m_[m[33m|[m[31m_[m[33m|[m[31m/[m  
[31m|[m[31m/[m[1;31m|[m [33m|[m [33m|[m   
[31m|[m * [33m|[m [33m|[m 9cf9bfb 2014-06-05 add ts package, needed for windows (Rick Olson[32m[m)
[31m|[m[31m/[m [33m/[m [33m/[m  
* [33m|[m [33m|[m   5467466 2014-06-05 Merge pull request #68 from github/logging-in-commands (risk danger olson[32m[m)
[1;32m|[m[1;33m\[m [33m\[m [33m\[m  
[1;32m|[m * [33m|[m [33m|[m 7b4b576 2014-06-05 remove another panic() (Rick Olson[32m[m)
[1;32m|[m * [33m|[m [33m|[m efb3af7 2014-06-05 unnecessary panic (Rick Olson[32m[m)
[1;32m|[m * [33m|[m [33m|[m 92cf6ea 2014-06-05 remove the old logging (Rick Olson[32m[m)
[1;32m|[m * [33m|[m [33m|[m f83d26f 2014-06-05 move logging functions to commands package (Rick Olson[32m[m)
[1;32m|[m[1;32m/[m [33m/[m [33m/[m  
* [33m|[m [33m|[m   37af77c 2014-06-05 Merge pull request #67 from github/install-hooks-with-path (risk danger olson[32m[m)
[1;34m|[m[1;35m\[m [33m\[m [33m\[m  
[1;34m|[m * [33m|[m [33m|[m 6c464cc 2014-06-05 install hooks with `git media path` (Rick Olson[32m[m)
* [1;35m|[m [33m|[m [33m|[m   613a9b2 2014-06-05 Merge pull request #66 from github/git-media-env (risk danger olson[32m[m)
[1;35m|[m[31m\[m [1;35m\[m [33m\[m [33m\[m  
[1;35m|[m [31m|[m[1;35m/[m [33m/[m [33m/[m  
[1;35m|[m[1;35m/[m[31m|[m [33m|[m [33m|[m   
[1;35m|[m * [33m|[m [33m|[m d9935fa 2014-06-05 rename 'config' to 'env' (Rick Olson[32m[m)
[1;35m|[m[1;35m/[m [33m/[m [33m/[m  
* [33m|[m [33m|[m   698f15f 2014-06-05 Merge pull request #65 from github/auto-install-hooks (risk danger olson[32m[m)
[32m|[m[33m\[m [33m\[m [33m\[m  
[32m|[m * [33m|[m [33m|[m 4248b1b 2014-06-04 don't print extra output during clean/smudge commands (Rick Olson[32m[m)
[32m|[m * [33m|[m [33m|[m e0b1433 2014-06-04 install hooks in clean/smudge filters (Rick Olson[32m[m)
[32m|[m[32m/[m [33m/[m [33m/[m  
* [33m|[m [33m|[m 81e499b 2014-06-04 forgot one (Rick Olson[32m[m)
* [33m|[m [33m|[m 645690b 2014-06-04 use defer to run tests (Rick Olson[32m[m)
* [33m|[m [33m|[m 2467a7b 2014-06-04 add both the clean and smudge tests" (Rick Olson[32m[m)
* [33m|[m [33m|[m 183b7fd 2014-06-04 test the clean command (Rick Olson[32m[m)
[33m|[m [33m|[m[33m/[m  
[33m|[m[33m/[m[33m|[m   
* [33m|[m 7ec7f41 2014-06-04 manformatting (Rick Olson[32m[m)
* [33m|[m 71e2976 2014-06-04 document the logs and queues commands (Rick Olson[32m[m)
* [33m|[m 2f08f52 2014-06-04 remove the need to use filepath.Join for AddPath() (Rick Olson[32m[m)
* [33m|[m 4832356 2014-06-04 move integration tests to ./commands (Rick Olson[32m[m)
* [33m|[m   18f166b 2014-06-04 Merge pull request #62 from github/pre-push-hook-2 (Scott Barron[32m[m)
[33m|[m[35m\[m [33m\[m  
[33m|[m [35m|[m[33m/[m  
[33m|[m[33m/[m[35m|[m   
[33m|[m * e718716 2014-06-04 Add the pre-push hook setup (rubyist[32m[m)
[33m|[m[33m/[m  
*   f5a0158 2014-06-04 Merge pull request #61 from github/keep-em-separated (Scott Barron[32m[m)
[36m|[m[1;31m\[m  
[36m|[m * fc9768b 2014-06-03 No need to export SimpleExec (rubyist[32m[m)
[36m|[m * cffa7d3 2014-06-03 Move filter installation under gitmedia package. Shave some yaks to make that happen. (rubyist[32m[m)
[36m|[m * 2719233 2014-06-03 Use gitmedia.Print instead of fmt.Print (rubyist[32m[m)
[36m|[m * 56f8c84 2014-06-03 consolidate notion of 'in repo' (rubyist[32m[m)
[36m|[m * 5d44b53 2014-06-03 Simplify in-repo init, move hook instalation under gitmedia package (rubyist[32m[m)
[36m|[m * e3be932 2014-06-03 init hooks should bail if not in a repo (rubyist[32m[m)
[36m|[m * 000aaed 2014-06-03 Provide separate initialization paths for global and in-repo initialiazation (rubyist[32m[m)
[36m|[m[36m/[m  
*   66f3e90 2014-06-03 Merge pull request #59 from github/no-more-local-import-paths (Scott Barron[32m[m)
[1;32m|[m[1;33m\[m  
[1;32m|[m * 66463b6 2014-06-03 go away .test artifacts (rubyist[32m[m)
[1;32m|[m *   2ab8c3d 2014-06-03 Merge branch 'master' into no-more-local-import-paths (rubyist[32m[m)
[1;32m|[m [1;34m|[m[1;32m\[m  
[1;32m|[m [1;34m|[m[1;32m/[m  
[1;32m|[m[1;32m/[m[1;34m|[m   
* [1;34m|[m   1f73d01 2014-06-03 Merge pull request #58 from github/man (risk danger olson[32m[m)
[1;36m|[m[31m\[m [1;34m\[m  
[1;36m|[m * [1;34m|[m 68b3718 2014-06-02 link to the external gitattributes(5) manual (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m a01f633 2014-06-02 add script for building man files (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 4781a38 2014-06-02 git-media-path(1) (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m cade956 2014-06-02 add git clean/smudge man pages (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m ac0e0bb 2014-06-02 put man files in ./man (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 2b46d73 2014-06-02 git-media-push(1) (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m b8bc66d 2014-06-02 git-media-init(1) (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 9c653e8 2014-06-02 add git media config man page (Rick Olson[32m[m)
[1;36m|[m * [1;34m|[m 79518af 2014-06-02 add git-media(1) (Rick Olson[32m[m)
[1;36m|[m[1;36m/[m [1;34m/[m  
[1;36m|[m * 8c57ade 2014-06-03 Missing formatting directive in error message (rubyist[32m[m)
[1;36m|[m * ff7bed0 2014-06-03 remove redundant clean up (rubyist[32m[m)
[1;36m|[m * 49c248f 2014-06-03 Don't depend on existing $GOPATH, everything under .vendor (rubyist[32m[m)
[1;36m|[m * 3178428 2014-06-03 Tests run, still depends on normal $GOPATH (rubyist[32m[m)
[1;36m|[m * f6d2ee3 2014-06-03 First pass at reorg. Builds, tests don't run yet (rubyist[32m[m)
[1;36m|[m[1;36m/[m  
* dd326fe 2014-06-02 import spec (Rick Olson[32m[m)
* f9c5a27 2014-06-02 Update integration tests to look for required line (rubyist[32m[m)
* 19966b6 2014-06-02 setup script/test to run individual package tests (Rick Olson[32m[m)
* 4de8743 2014-06-02 remove ruby test suite (Rick Olson[32m[m)
* 120efb7 2014-06-02 write integration tests in go (Rick Olson[32m[m)
* 6ec56aa 2014-06-02 More rubby differences? (rubyist[32m[m)
* f1a7970 2014-06-02 ruby syntax - ci is not using 2.0 (rubyist[32m[m)
* c77a64f 2014-06-02 Test for attributes in .git/info/attributes. Add before block, t.write() (rubyist[32m[m)
* 4949f28 2014-06-02 Integration tests for listing from all .gitattributes files (rubyist[32m[m)
* f393202 2014-06-02 rename @repositories => @paths (Rick Olson[32m[m)
* 6a6c4c1 2014-06-02 get rid of config object (Rick Olson[32m[m)
* 0602cbc 2014-06-02 test 'git media path' (Rick Olson[32m[m)
* aec93ec 2014-06-02 clone the repos before each test (Rick Olson[32m[m)
* 5d02b22 2014-06-02 expand /var symlink (Rick Olson[32m[m)
* 5492e78 2014-06-02 fix tests (Rick Olson[32m[m)
* 37a845f 2014-06-02 prevent panic when remote is not a url (Rick Olson[32m[m)
* 45bc740 2014-06-02 all tests are against fixture repositories (Rick Olson[32m[m)
* 9ea90d7 2014-06-02 commit the fixture cleaning scripts (Rick Olson[32m[m)
* c677732 2014-06-02 always clean git test fixtures (Rick Olson[32m[m)
* f7cc510 2014-06-02 setup git repo fixtures (Rick Olson[32m[m)
* 7316645 2014-06-02 Ruby and Go get the temp dir the same way (Rick Olson[32m[m)
* 9b47027 2014-06-02 only run integration tests if unit tests pass (Rick Olson[32m[m)
* c860db8 2014-06-02 separate integration tests from unit tests (Rick Olson[32m[m)
* 3d71602 2014-06-02 docs (Rick Olson[32m[m)
* 5fa549c 2014-06-02 a more declarative syntax for complex command tests (Rick Olson[32m[m)
* a3462d5 2014-06-02 add a `init` failing test case (Rick Olson[32m[m)
* 2b69298 2014-06-02 don't need block syntax if we're just checking output (Rick Olson[32m[m)
* 18637bb 2014-06-02 try passing GOPATH to the go scripts (Rick Olson[32m[m)
* 53d220b 2014-06-02 use the ruby integration test suite (Rick Olson[32m[m)
* dc65cef 2014-06-02 set correct exit codes (Rick Olson[32m[m)
* c350867 2014-06-02 integration tests in ruby instead... (Rick Olson[32m[m)
* fa1fdac 2014-06-02 grammar (rubyist[32m[m)
* c4653d1 2014-06-02 Set the required flag for media filters (rubyist[32m[m)
* c6d5cb8 2014-06-02 unused vars (Rick Olson[32m[m)
* bab1552 2014-06-02 unset before setting the clean/smudge filters (Rick Olson[32m[m)
* 3425ae8 2014-06-02 Make queues list output consistent with git status (rubyist[32m[m)
* 2d5a920 2014-06-02 only call Getwd once (rubyist[32m[m)
* e1ede04 2014-06-02 Make paths output consistent with git status, i.e. relative to cwd (rubyist[32m[m)
* 09fe958 2014-06-02 Use LocalWorkingDir and LocalGitDir as bases for searching attributes (rubyist[32m[m)
* b0ad898 2014-06-02 When searching for the git directory, bail out of we reach '/'. Repo not found. (rubyist[32m[m)
* b99c999 2014-06-02 Also parse .git/info/attributes if it exists (rubyist[32m[m)
* 23b77d2 2014-06-02 Walk the project looking for .gitattributes files when listing paths (rubyist[32m[m)
* d52b3f5 2014-06-02 rename  to  while we're at it (Rick Olson[32m[m)
* 7be700d 2014-06-02 actually set the new media filter values (Rick Olson[32m[m)
* d7f09b0 2014-06-02 git media init resets config (Rick Olson[32m[m)
* 24b1939 2014-06-02 `git media [clean | smudge]` vs `git-media-clean` (Rick Olson[32m[m)
* 4e24860 2014-06-02 send GOROOT to `go build` (Rick Olson[32m[m)
* dc4a9ec 2014-06-02 create the release directory during installation (Rick Olson[32m[m)
* b972a9f 2014-06-02 Show the source of a path (currently just .gitattributes). Also able to use 'paths' (rubyist[32m[m)
* 05ffdeb 2014-06-02 Create the .gitattributes file if it doesn't exist. Make sure there's a newline after added paths (rubyist[32m[m)
* 1002a41 2014-06-02 type -> path (rubyist[32m[m)
* 4fd0399 2014-06-02 remove type (rubyist[32m[m)
* 20b9214 2014-06-02 add a media type (rubyist[32m[m)
* 4d09bbd 2014-06-02 List file types known by git media for the repo (rubyist[32m[m)
* 2041927 2014-06-02 clean up existing git-media binaries after installs (Rick Olson[32m[m)
* 00c455e 2014-06-02 fix usage (Rick Olson[32m[m)
* 3eacb03 2014-06-02 fix default endpoint for ssh remotes (Rick Olson[32m[m)
* 355349d 2014-06-02 only write this crap once (Rick Olson[32m[m)
* ea5bee7 2014-06-02 add existing GIT env to expected config in integration tests (Rick Olson[32m[m)
* 44e1679 2014-06-02 integration test failures should fail ci build (Rick Olson[32m[m)
* 90e476c 2014-06-02 cibuild only needs to invoke script/test once (risk danger olson[32m[m)
* 76b22cb 2014-06-02 stricter tests for decoding (rubyist[32m[m)
* cbddeb1 2014-06-02 stricter pointer file matching, use git-media instead of external (rubyist[32m[m)
* d2e1403 2014-06-02 remove unneeded variable (rubyist[32m[m)
* 143ba72 2014-06-02 Fix bugs in header validation, move into its own function (rubyist[32m[m)
* e51149c 2014-06-02 more strict parsing of media header (rubyist[32m[m)
* e6cd37c 2014-06-02 Parse out git media header (rubyist[32m[m)
* 1d93b80 2014-06-02 Validate access with OPTIONS before doing a PUT (rubyist[32m[m)
* 11cdd54 2014-06-02 Check for the git media header and strip it off (rubyist[32m[m)
* c7b35e4 2014-06-02 Handle ssh urls (rubyist[32m[m)
* 4248de4 2014-06-02 Handle repos that end with and without ".git" (rubyist[32m[m)
* ad87d48 2014-06-02 that was an inflexible way of doing it (rubyist[32m[m)
* 2fdafc7 2014-06-02 Start an `init` command, checks for and installs clean/smudge filter config (rubyist[32m[m)
* 8aebb2b 2014-06-02 remove unused code (rubyist[32m[m)
* ceb0fc1 2014-06-02 Remove output that conflicts with git's own output (rubyist[32m[m)
* daebd31 2014-06-02 Download progress bar (needs some clean up, but it works) (rubyist[32m[m)
* 66e3d13 2014-06-02 Display filenames when syncing up (rubyist[32m[m)
* 54139e6 2014-06-02 Show file name in `git media queues` if it is present, otherwise show hash (rubyist[32m[m)
* d7af583 2014-06-02 If given, add the filename to the upload queue (rubyist[32m[m)
* 4cab1cd 2014-06-02 add gopack binary for Windows (half-ogre[32m[m)
* aa8e2d2 2014-06-02 kill toml (Rick Olson[32m[m)
* c184f7d 2014-06-02 some (*Configuration) Endpoint() tests (Rick Olson[32m[m)
* d505766 2014-06-02 use the origin endpoint as the default (Rick Olson[32m[m)
* 2e25dcc 2014-06-02 kill toml, git config does what we need (Rick Olson[32m[m)
* 9f4a8e2 2014-06-02 allow configurable media endpoints per remote (Rick Olson[32m[m)
* 2f12609 2014-06-02 use LocalWorkingDir to find the toml file (Rick Olson[32m[m)
* 46ab1cd 2014-06-02 use append() to add GIT_* env vars (Rick Olson[32m[m)
* 2064283 2014-06-02 track the working dir and the git dir (Rick Olson[32m[m)
* 8949720 2014-06-02 find correct git dir from sub dir (Rick Olson[32m[m)
* e75799e 2014-06-02 confirm tests on git-media/.git dir (Rick Olson[32m[m)
* 4cf3bc5 2014-06-02 initial integration tests spike (Rick Olson[32m[m)
* 16d67c8 2014-06-02 `git-media config` shows the config settings (Rick Olson[32m[m)
* 3c722cd 2014-06-02 log env keys matching GIT_*, and important git media config values (Rick Olson[32m[m)
* 110d41c 2014-06-02 fix script/test (Rick Olson[32m[m)
* e36f46f 2014-06-02 update scripts for ci (rick[32m[m)
* b4f4134 2014-06-02 update install.sh to use %f flag (risk danger olson[32m[m)
* 9d60c76 2014-06-02 update install.bat to use %f flag (risk danger olson[32m[m)
* 1ccee54 2014-06-02 dont try to fmt the old server code (risk danger olson[32m[m)
* e1758ca 2014-06-02 remove server stuff (risk danger olson[32m[m)
* b517620 2014-06-02 dont print smudge output to stdout (risk danger olson[32m[m)
* 099adaa 2014-06-02 send accept/content-type headers on git-media PUT requests (risk danger olson[32m[m)
* 163529c 2014-06-02 check PREFIX or BOXEN_HOME for a better prefix (risk danger olson[32m[m)
* 20f1248 2014-06-02 make install.sh executable (risk danger olson[32m[m)
* b26765d 2014-06-02 be flexible about missing log directories (risk danger olson[32m[m)
* 2d1e739 2014-06-02 ensure -debug option is in the usage (risk danger olson[32m[m)
* 6838e69 2014-06-02 make it clear where the logs are going (risk danger olson[32m[m)
* 290e143 2014-06-02 version easter egg should still print the version (risk danger olson[32m[m)
* 152d131 2014-06-02 rename errors to logs command (risk danger olson[32m[m)
* 30a7361 2014-06-02 default errors subcommand reads the log file (risk danger olson[32m[m)
* 097f150 2014-06-02 add commands for listing errors and showing the latest (risk danger olson[32m[m)
* f565dd3 2014-06-02 rename Print() to Error(), add a Print() for STDOUT.  Also remove more fmt.Print calls (risk danger olson[32m[m)
* bfe3e33 2014-06-02 FlagSet errors should go to the panic log too (risk danger olson[32m[m)
* 96ff9d5 2014-06-02 use gitmedia.Print() instead of fmt.Print (risk danger olson[32m[m)
* ab32d19 2014-06-02 rename boomtown to errors command. (risk danger olson[32m[m)
* 56afd22 2014-06-02 docs on the logging functions (risk danger olson[32m[m)
* c8e63fd 2014-06-02 Don't exit before printing the panic log (risk danger olson[32m[m)
* 7742c44 2014-06-02 typos (risk danger olson[32m[m)
* ba04873 2014-06-02 add a Print for stderr messages (risk danger olson[32m[m)
* 934408e 2014-06-02 add simple Exit function if there is no error (risk danger olson[32m[m)
* ba62ee4 2014-06-02 use > instead of $ (risk danger olson[32m[m)
* 297ec76 2014-06-02 Panic() now writes to a log file (risk danger olson[32m[m)
* fc2fa92 2014-06-02 welcome to boomtown (risk danger olson[32m[m)
* d6fbaf5 2014-06-02 add -debug flag to commands (risk danger olson[32m[m)
* 4bf11bd 2014-06-02 enforce correct usage of gitmedia.Panic() everywhere (risk danger olson[32m[m)
* 69057e0 2014-06-02 we should be using gitmedia.Panic() instead of panic() everywhere (risk danger olson[32m[m)
* d6066ce 2014-06-02 get the path if its available (rick[32m[m)
* 5f8c073 2014-06-02 add stderr messages /cc @jbarnette (rick[32m[m)
* 8aa5f52 2014-06-02 account for the trailing linebreak (rick[32m[m)
* 7c22a43 2014-06-02 add an ending linebreak to the meta file that's written to git (risk danger olson[32m[m)
* e219224 2014-06-02 Create media file when it doesn't exist, use local copy when it does (rubyist[32m[m)
* 72bfbaa 2014-06-02 Move media file creation to smudge filter rather than client (rubyist[32m[m)
* 509c7bb 2014-06-02 When the client gets media and the media file doesn't exist, ensure it is created (rubyist[32m[m)
* 8d21a00 2014-06-02 get rid of the awful panic.  `git media config` to see your current config (rick[32m[m)
* b0d77f1 2014-06-02 read toml file if it exists (rick[32m[m)
* e38b7cb 2014-06-02 need some cibuild action (rick[32m[m)
* 330c8cb 2014-06-02 test that ObjectUrl behaves with various endpoint values (rick[32m[m)
* 550c5c7 2014-06-02 initial config example (rick[32m[m)
* a4c420a 2014-06-02 rename dir to releases (risk danger olson[32m[m)
* 436cf44 2014-06-02 put all the releases in the same dir (risk danger olson[32m[m)
* 297c2ba 2014-06-02 fix windows zip creation (risk danger olson[32m[m)
* 8e94d5f 2014-06-02 use exec.Cmd CombinedOutput() (risk danger olson[32m[m)
* d7a4e9f 2014-06-02 basic windows installer (risk danger olson[32m[m)
* f89ba31 2014-06-02 exe extension for windows (risk danger olson[32m[m)
* 96afa8d 2014-06-02 rename the install.sh sample (risk danger olson[32m[m)
* ca30a16 2014-06-02 add the global git config commands (risk danger olson[32m[m)
* cb162ce 2014-06-02 really dumb installer for *nix (risk danger olson[32m[m)
* dd69f82 2014-06-02 old, unused prototype scripts (risk danger olson[32m[m)
* 34b02e5 2014-06-02 white space (risk danger olson[32m[m)
* 41dbdab 2014-06-02 the only thing tha really changes is the dir (risk danger olson[32m[m)
* fb01d2e 2014-06-02 be sure that GOPATH is set (risk danger olson[32m[m)
* 9059e2b 2014-06-02 put cross-compiled binaries in separate dirs (risk danger olson[32m[m)
* 75c180a 2014-06-02 Cannot defer Close here or it crashes The connection will be closed when doRequest() returns, but isn't being read until Get() returns to smudge.go. This causes a crash in git-media-smudge (rubyist[32m[m)
* e8de7a3 2014-06-02 approve the credentials if the request succeeds, or reject them (risk danger olson[32m[m)
* 0074d2f 2014-06-02 better credential error messages (risk danger olson[32m[m)
* 9981470 2014-06-02 split credential functions to a separate file (risk danger olson[32m[m)
* 9d406f69 2014-06-02 handle get errors (risk danger olson[32m[m)
* cc1c530 2014-06-02 return the credentials (risk danger olson[32m[m)
* 4dd79c5 2014-06-02 extract the credential logic to a function and an embedded struct (risk danger olson[32m[m)
* aad597d 2014-06-02 add support for pulling credentials out of git-credentials (rick[32m[m)
* 4315cf2 2014-06-02 make a common clientRequest() helper (rick[32m[m)
* 5c1e7df 2014-06-02 don't print this to stdout. (rick[32m[m)
* dc6afd9 2014-06-02 set the content length (rick[32m[m)
* f416951 2014-06-02 halt immediately if the alambic api returns an unsuccessful status (rick[32m[m)
* b59b531 2014-06-02 only hash the file content (rick[32m[m)
* ae02e8c 2014-06-02 use the file size to calculate the oid (rick[32m[m)
* b896dcb 2014-06-02 use sha256 because why not (rick[32m[m)
* 9b8ef9b 2014-06-02 don't ever shadow errors (rick[32m[m)
* 200acc7 2014-06-02 smudge uses the gitmediaclient to download files if necessary (rick[32m[m)
* c40853e 2014-06-02 using http verbs for the client method names (rick[32m[m)
* b3a4ae6 2014-06-02 there can be only one smudger (rick[32m[m)
* fb40a32 2014-06-02 return the error (rick[32m[m)
* 4fb75f8 2014-06-02 "git media sync" now walks to the upload queue (rick[32m[m)
* 74c19bf 2014-06-02 walk the queue (rick[32m[m)
* de78a0a 2014-06-02 add `git media queues` for inspecting the queues (rick[32m[m)
* 3aca460 2014-06-02 doh, filepath.Ext(".git") always == ".git" (rick[32m[m)
* 1c4e794 2014-06-02 make it easier to queue stuff (rick[32m[m)
* dfa80b4 2014-06-02 add cleaned files to the clean queue (rick[32m[m)
* 7bf0709 2014-06-02 add Queue.Walk() (rick[32m[m)
* e0018ef 2014-06-02 add assert (rick[32m[m)
* bf631d2 2014-06-02 remove concept of tmp queue (rick[32m[m)
* dfa2422 2014-06-02 implement Queue.Move() (rick[32m[m)
* c5ea396 2014-06-02 teach Queue how to delete stuff (rick[32m[m)
* 37cae35 2014-06-02 add AddString() and AddBytes() helpers (rick[32m[m)
* ddda5a3 2014-06-02 use simpleuuid (rick[32m[m)
* 8e1a1a1 2014-06-02 start on a basic queuedir package (rick[32m[m)
* f531a54 2014-06-02 clean large assets with  header (risk danger olson[32m[m)
* 93bd77b 2014-06-02 use `git diff-index`, separated by NULs. (risk danger olson[32m[m)
* 7f4e95e 2014-06-02 more debugging to figure out the benchmarking woes (rick[32m[m)
* 099c2f1 2014-06-02 better panics and debugging (rick[32m[m)
* 06676e2 2014-06-02 return early if the file exists (rick[32m[m)
* 138ead9 2014-06-02 add a pseudo shebang header for identification (rick[32m[m)
* fbdb915 2014-06-02 remove the clean writers (rick[32m[m)
* a3ad1a5 2014-06-02 add a basic server for syncing (rick[32m[m)
* bd003df 2014-06-02 don't bother writing the metadata file, all we need is the sha (rick[32m[m)
* 0dc4b1b 2014-06-02 don't hardcode os.Stdout (rick[32m[m)
* 27fd6ab 2014-06-02 add gopack (rick[32m[m)
* 0c7f659 2014-06-02 use a LocalSmudger (rick[32m[m)
* c5de76d 2014-06-02 rename gitmediaclean => gitmediafilters (rick[32m[m)
* 9cff651 2014-06-02 let PipeMediaCommand() figure out where the media commands live (rick[32m[m)
* 8b18f62 2014-06-02 pass git-media-clean and git-media-smudge commands straight through (rick[32m[m)
* 7440cc5 2014-06-02 add script/test (risk danger olson[32m[m)
* 13caafa 2014-06-02 metaencoding tests (risk danger olson[32m[m)
* 0dcab39 2014-06-02 write only a sha to the meta data (risk danger olson[32m[m)
* 51d054c 2014-06-02 remove that warning header length (risk danger olson[32m[m)
* 04c2ed6 2014-06-02 just pass all args to the inner script (risk danger olson[32m[m)
* 5b1b34d 2014-06-02 add a cross compilation build script (risk danger olson[32m[m)
* ac20e8b 2014-06-02 go not git (risk danger olson[32m[m)
* b8f7cb4 2014-06-02 add script for benchmarking git-media on a full git repo's working dir (risk danger olson[32m[m)
* e44bff9 2014-06-02 some comments (risk danger olson[32m[m)
* 6329a4a 2014-06-02 extract cleaning logic (risk danger olson[32m[m)
* 5b67411 2014-06-02 extract clean writer to the gitmediaclean package (risk danger olson[32m[m)
* 6d81cb7 2014-06-02 move MediaWarning where it makes sense (risk danger olson[32m[m)
* f45828c 2014-06-02 extract large media encoder/decoder (risk danger olson[32m[m)
* c22d346 2014-06-02 don't re-write the large asset file if it exists (risk danger olson[32m[m)
* f827793 2014-06-02 quick smudge implementation (risk danger olson[32m[m)
* 6a1cfb6 2014-06-02 write a warning at the top of the file (risk danger olson[32m[m)
* e7dae8e 2014-06-02 extra Println (risk danger olson[32m[m)
* 0694316 2014-06-02 add a git-media clean script (risk danger olson[32m[m)
* b355ef7 2014-06-02 move common logic to the gitmedia package (risk danger olson[32m[m)
* 8005bee 2014-06-02 script/fmt formats commands/* too (risk danger olson[32m[m)
* 4b9690c 2014-06-02 add script/build (risk danger olson[32m[m)
* b716249 2014-06-02 use the command name (risk danger olson[32m[m)
* 4d92c26 2014-06-02 basic Usage() (risk danger olson[32m[m)
* 0671b84 2014-06-02 separate Setup() and Run() steps (risk danger olson[32m[m)
* a95a2d7 2014-06-02 separate commands out (risk danger olson[32m[m)
* a76a186 2014-06-02 move commands to cmd/* (risk danger olson[32m[m)
* dca7f73 2014-06-02 really basic pre-commit hook that rejects commits with files > 5MB (risk danger olson[32m[m)
* 55c5c5e 2014-06-02 add a basic pre-commit hook that lists changed files (risk danger olson[32m[m)
* 2ba1fee 2014-06-02 initial version subcommand with easter egg (risk danger olson[32m[m)
* 3ccf614 2014-06-02 add script/fmt (risk danger olson[32m[m)
* d8f7803 2014-06-02 rip off internal prototype's README (risk danger olson[32m[m)
