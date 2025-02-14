# Releasing Git LFS

The core team of Git LFS maintainers publishes releases on a cadence of their
determining.

## Release Naming

We follow Semantic Versioning standards as follows:

  * `MAJOR` releases are done on a scale of 2-4 years. These encompass breaking,
    incompatible API changes, or command-line interface changes that would
    cause existing programs or use-cases scripted against Git LFS to break.

  * `MINOR` releases are done on a scale of 2-6 months. These encompass new
    features, bug fixes, and other "medium"-sized changes into a semi-regular
    release schedule.

  * `PATCH` releases are done on the scale of weeks to months. These encompass
    critical bug fixes, but lack new features. They are amended to a `MINOR`
    release "series", or, if serious enough (e.g., security vulnerabilities,
    etc.) are backported to previous versions.

## Release Artifacts

We package several artifacts for each tagged release. They are:

  1. `git-lfs-@{os}-v@{release}-@{arch}.tar.gz` for the following values:

      |     | operating system | architecture |
      | --- | ---------------- | ------------ |
      | git-lfs-darwin-amd64-v@{version}.tar.gz | darwin | amd64 |
      | git-lfs-darwin-arm64-v@{version}.tar.gz | darwin | arm64 |
      | git-lfs-freebsd-386-v@{version}.tar.gz | freebsd | 386 |
      | git-lfs-freebsd-amd64-v@{version}.tar.gz | freebsd | amd64 |
      | git-lfs-linux-386-v@{version}.tar.gz | linux (generic) | 386 |
      | git-lfs-linux-amd64-v@{version}.tar.gz | linux (generic) | amd64 |
      | git-lfs-linux-arm-v@{version}.tar.gz | linux (generic) | arm |
      | git-lfs-linux-arm64-v@{version}.tar.gz | linux (generic) | arm64 |
      | git-lfs-linux-loong64-v@{version}.tar.gz | linux (generic) | loong64 |
      | git-lfs-linux-ppc64le-v@{version}.tar.gz | linux (generic) | ppc64le |
      | git-lfs-linux-riscv64-v@{version}.tar.gz | linux (generic) | riscv64 |
      | git-lfs-linux-s390x-v@{version}.tar.gz | linux (generic) | s390x |

  2. `git-lfs-windows-v@{release}-@{arch}.zip` for the following values:

      |     | operating system | architecture |
      | --- | ---------------- | ------------ |
      | git-lfs-windows-386-v@{version}.zip | windows | 386 |
      | git-lfs-windows-amd64-v@{version}.zip | windows | amd64 |
      | git-lfs-windows-arm64-v@{version}.zip | windows | arm64 |

  3. `git-lfs-windows-v@{release}.exe`, a signed Windows installer that contains
     copies of both `-x86` and `-x64` copies of Git LFS.

  4. `*.deb`, and `*.rpm` packages for all of the distributions named in
     `script/packagegcloud.rb`.

## Development Philosophy

We do all major development on the `main` branch, and assume it to be passing
tests at all times. New features are added via the feature-branch workflow, or
(optionally) from a contributor's fork.

This is done so that `main` can progress and grow new features, while
historical releases such as `vM.N.0` can receive bug fixes as they are applied
to `main`, eventually culminating in a `vM.N.1` (and so on) release.

## Building a release

Let release `vM.N.P` denote the version that we are _releasing_.

When `P` is equal to zero, we say that we are releasing a MINOR version of
Git LFS in the `vM.N`-series, unless `N` is also equal to zero, in which
case we are releasing a MAJOR version.  Conversely, if `P` is not equal
to zero, we are releasing a PATCH version.

  1. Upgrade the Go version used in the `git-lfs/build-dockers` repository
     to the latest available PATCH release with the same major and minor
     version numbers as the Go version used in this repository's GitHub
     Actions workflows.

  2. Write the release notes and do the housekeeping required to
     indicate a new version.  For a MAJOR or MINOR version, we start with
     a `main` branch which is up to date with the latest changes from the
     remote and then checkout a new `release-next` branch from that base.

     If we are releasing a PATCH version, we create a `release-M.N` branch
     with cherry-picked merges from the `main` branch, as described in
     the [instructions](#building-patch-versions) below, and then checkout
     the `release-next` branch from that base.

     We next perform the following steps to prepare the `release-next` branch:

     * Run `script/changelog` and categorize each merge commit as a feature,
       bug fix, miscellaneous change, or a change to be skipped and ignored.
       Ensure that your `~/.netrc` credentials are up-to-date in order to
       make requests to the GitHub API, or use a `GITHUB_TOKEN` environment
       variable.

       The `changelog` script will write a portion of the new CHANGELOG to
       stdout, which you should copy and paste into `CHANGELOG.md`, along with
       an H2-level heading containing the new version and the expected release
       date.  This heading should be consistent with the existing style in the
       document.

       For a MAJOR release, use `script/changelog v(M-1).L.0...HEAD`, where
       `(M-1)` is the previous MAJOR release number and `L` is the final
       MINOR release number in that series.  For a MINOR release, use
       `script/changelog vM.(N-1).0...HEAD`, where `(N-1)` is the previous
       MINOR release number, and for a PATCH release, use
       `script/changelog --patch vM.N.(P-1)...HEAD`, where `(P-1)` is the
       previous PATCH release number.

       * Optionally write 1-2 paragraphs summarizing the release and calling
         out community contributions.

       * Call out any changes in the operating system versions required
         by the new release, as well as any differences in the set of Linux
         platforms for which we build release packages.  Check for any
         new platform requirements from the version of Go in use.

       * If we are releasing a MAJOR or MINOR version and not a PATCH, and
         if the most recent non-PATCH release was followed by a series of one
         or more PATCH releases, include any changes listed in the CHANGELOG
         of that series' release branch in the new release's CHANGELOG.
         (For a new MAJOR version, the prior release branch would be named
         `release-(M-1).L`, following the terminology defined above, while
         for a new MINOR version the prior release branch would be named
         `release-M.(N-1)`.)

     * Run `script/update-version vM.N.P` to update the version number in all
       of the relevant files.  Note that this script requires a version of
       `sed(1)` compatible with the GNU implementation.

     * Adjust the date in the `debian/changelog` entry to reflect the
       expected release date rather than the current date.

     * Commit all the files changed in the steps above in a single new commit:
       ```ShellSession
       $ git commit -m 'release: vM.N.P'
       ```

  3. Push the `release-next` branch and create a pull request with your
     changes from the branch.  If you're building a MAJOR or MINOR release,
     set the base to the `main` branch.  Otherwise, set the base to the
     `release-M.N` branch.

     * Add the `release` label to the PR.

     * In the PR description, consider uploading builds for implementers to
       use in testing.  These can be generated from a local tag, which
       does not need to be signed (but must be annotated).  Check that the
       local version of Go is equivalent to the most recent one used by the
       GitHub Actions workflows for the release branch, which may be different
       from that used on the `main` branch.  For a PATCH release in particular
       you may need to downgrade your local Go version.  The build artifacts
       will be placed in the `bin/releases` directory and may be uploaded
       into the PR from there:

       ```ShellSession
       $ git tag -m vM.N.P-pre vM.N.P-pre
       $ make release
       $ ls bin/releases
       ```

     * Notify the `@git-lfs/releases` and `@git-lfs/implementers` teams,
       collections of humans who are interested in Git LFS releases.

     * Ensure that the normal Continuous Integration workflow for PRs
       that runs automatically in GitHub Actions succeeds fully.

     * As the GitHub Actions release workflow will not run for PRs, consider
       creating an annotated tag with the `-pre` suffix and pushing the tag,
       which will trigger a run of the release workflow that does not upload
       artifacts to Packagecloud.  Alternatively, in a private clone of
       the repository, create such a tag from the `release-next` branch plus
       one commit to change the repository name in `script/upload`, and
       push the tag so Actions will run the release workflow.  Ensure that
       the workflow succeeds (excepting the Packagecloud upload step, which
       will be skipped).

  4. Once approved and verified, merge the pull request you created in the
     previous step. Locally, create a GPG-signed tag on the merge commit called
     `vM.N.P`:

     ```ShellSession
     $ git show -q --pretty=%s%n%b HEAD
     Merge pull request #xxxx from git-lfs/release-next
     release: vM.N.P

     $ git tag -s vM.N.P -m vM.N.P

     $ git describe HEAD
     vM.N.P

     $ git show -q --pretty=%s%d%n%b vM.N.P
     tag vM.N.P
     Tagger: ...

     vM.N.P
     -----BEGIN PGP SIGNATURE-----
     ...
     -----END PGP SIGNATURE-----
     Merge pull request #xxxx from git-lfs/release-next (tag: vM.N.P)
     release: vM.N.P
     ```

  5. Push the tag, via:

     ```ShellSession
     $ git push origin vM.N.P
     ```

     Validate the pending `production` environment deployment rule in the UI
     of the GitHub management application for macOS Developer ID certificates.

     This will kick off the process of building the release artifacts.  This
     process will take somewhere between 45 minutes and an hour.  When it's
     done, you'll end up with a set of release assets available for download,
     and Linux RPM and Debian packages for the new release will be available
     on Packagecloud.

  6. Download the `release-assets` archive file from the "Artifacts" section
     of the "Summary" page for the most recent run of the release GitHub
     Actions
     [workflow](https://github.com/git-lfs/git-lfs/actions/workflows/release.yml),
     and then unpack the file at the top level of the repository:

     ```ShellSession
     $ rm -rf bin/releases
     $ unzip /path/to/release-assets.zip
     ```

     The release assets should now be present in the `bin/releases` directory.
     The `script/upload` utility will create a new draft release announcement
     and then upload the release assets and attach them to the announcement:

     ```ShellSession
     $ script/upload --skip-verify vM.N.P
     ```

     Note that this script requires GnuPG, with your signing key configured,
     as well as Ruby 3.x, the OpenSSL Ruby gem, and several other tools,
     including the GNU coreutils version of `b2sum(1)`.  You will need to
     provide your GitHub credentials in your `~/.netrc` file or via a
     `GITHUB_TOKEN` environment variable.

  7. Finalize the release process using the same `script/upload` utility but
     with the `--finalize` option.  The script will add GPG signatures to the
     `hashes` and `sha256sums` files and then upload the signed files:

     ```ShellSession
     $ script/upload --finalize vM.N.P
     ```

     If you want to inspect the data before approving it, pass the `--inspect`
     option, which will drop you to a shell and let you look at things.  If the
     shell exits successfully, the build will be signed; otherwise, the process
     will be aborted.

  8. Publish the release on GitHub, assuming it looks correct.

  9. Move any remaining items out of the milestone for the current release to a
     future release and close the milestone.

 10. Update the `_config.yml` file in
     [`git-lfs/git-lfs.github.com`](https://github.com/git-lfs/git-lfs.github.com),
     similar to the following:

     ```diff
     --- _config.yml
     +++ _config.yml
     @@ -1,7 +1,7 @@
      # Site settings
      title: "Git Large File Storage"
      description: "Git Large File Storage (LFS) replaces large files such as audio samples, videos, datasets, and graphics with text pointers inside Git, while storing the file contents on a remote server like GitHub.com or GitHub Enterprise."
     -git-lfs-release: M.(N-1).0
     +git-lfs-release: M.N.0

      url: "https://git-lfs.com"
     ```

 11. If Homebrew does not automatically update within a few hours,
     create a GitHub PR to update the Homebrew formula for Git LFS with
     the `brew bump-formula-pr` command on a macOS system.  The SHA-256 value
     should correspond with the packaged artifact containing the new
     release's source files which is available at the given URL:

     ```ShellSession
     $ brew tap homebrew/core
     $ brew bump-formula-pr \
         --url https://github.com/git-lfs/git-lfs/releases/download/vM.N.P/git-lfs-vM.N.P.tar.gz \
         --sha256 <SHA-256> \
         git-lfs
     ```

### Building PATCH versions

When building a PATCH release, we cherry-pick merges from `main` to the
`vM.N` release branch, creating the branch first if it does not exist,
and then use that branch as the base for the PATCH release.

  1. Upgrade or downgrade the Go version used in the `git-lfs/build-dockers`
     repository to the latest available PATCH release with the same major
     and minor version numbers as the Go version used to build the Git LFS
     `vM.N.(P-1)` release.

  2. If the `release-M.N` branch does not already exist, create it from
     the corresponding MINOR release tag (or MAJOR release tag, if no
     MINOR releases have been made since the last MAJOR release):

     ```ShellSession
     $ git checkout -b release-M.N vM.N.0
     ```

     If the release branch already exists because this is not the first
     PATCH release for the given MINOR (or MAJOR) release, simply checkout
     the `release-M.N` branch, and ensure that you have the latest changes
     from the remote.

  3. Gather a set of potential candidates to backport to the `release-M.N`
     branch with:

     ```ShellSession
     $ git log --merges --first-parent vM.N.(P-1)...main
     ```

  4. For each merge that you want to backport, run:

     ```ShellSession
     $ git cherry-pick -m1 <SHA-1>
     ```

     This will cherry-pick the merge onto your release branch, using
     the `-m1` option to specify that the first parent of the merge
     corresponds to the mainline.

  5. Then follow the [guidelines](#building-a-release) above, using the
     `release-M.N` branch as the base for the new PATCH release.

### Building security PATCH versions

In our public security [policy](../../SECURITY.md) we request that
potential vulnerabilities in the Git LFS client be reported to us
via email.  Users sometimes also choose to open draft GitHub security
advisories, although we do not encourage that option.

If we determine that such a report is valid, we develop a PATCH release
version of the client with a remediation for the security issue,
following the steps below.

  1. Open a new draft security advisory, fill out the relevant
     details in the
     [template](https://github.com/git-lfs/git-lfs/security/advisories/new),
     and request a CVE identifier.

  2. Create a temporary private fork of this repository from the draft
     security advisory page, and use the fork to develop a resolution
     of the vulnerability.

  3. Create two PRs in the private fork, one to remediate the vulnerability
     in the `main` development branch of the public `git-lfs/git-lfs`
     repository, and one to do so in the appropriate `release-M.N` branch.

     The PR which targets the release branch should have a final, extra
     commit which adds an entry for the new PATCH release version in the
     `CHANGELOG.md` file, describing the release and the security fix it
     contains.  This commit should also include the changes generated
     by the `script/update-version vM.N.P` command, as per the second step
     in our standard release process.

     If the `release-M.N` branch does not exist in the public repository,
     create it as per our regular PATCH release process, and then create
     the second PR in the private fork against the new branch.  Be prepared
     to proceed relatively quickly, since the appearance of the new branch
     serves as a public notice that we may be publishing a PATCH release
     shortly.

  4. From the draft security advisory page, use the "Merge pull requests"
     option to merge both PRs simultaneously.

     Note that both the release branch and the `main` branch in the private
     fork must be up-to-date with the corresponding branches in the public
     repository in order for the GitHub UI to determine whether the PRs have
     no merge conflicts.

  5. Publish the security advisory.

  6. Follow our standard release process, starting with the step in which
     we create a GPG-signed tag named `vM.N.P` on the merge commit to the
     `release-M.N` branch.

  7. After publication of the new release, update the `git-lfs.com` home
     page with a banner message regarding the security PATCH release.
     Place the banner at the top of the `_includes/home/secondary.html`
     file in the `git-lfs/git-lfs.github.com` repository.  Use the GHSA
     identifier from the security advisory, and remove any previous
     security message banners.

     ```html
     <div class="column">
       <a class="dot-com-announcement" data-ga-params="banner"
         href="https://github.com/git-lfs/git-lfs/security/advisories/GHSA-abcd-1234-wxyz">
         Git LFS security update: <span>All users should update to M.N.P or newer</span>.</a>
       </a>
     </div>
     ```

  8. Create a new discussion in the "Announcements" section of the GitHub
     discussion [forum](https://github.com/git-lfs/git-lfs/discussions)
     describing the security PATCH release, and pin the discussion for
     all categories in the forum.

### Building security PATCH versions under embargo

When coordinating the release of a security patch with one or more
other projects, we must not follow the processes described above as
they depend on the GitHub Actions release workflow in our public
`git-lfs/git-lfs` repository.  Instead, we use a private repository
with a modified release workflow to build the binaries and packages
for our new version so that we can share them with the other projects
in advance of the coordinated release date.

  1. Open a draft security advisory, apply for a CVE identifier, and
     create a temporary private fork from the advisory page.  Develop a
     resolution of the vulnerability in the private fork.  These initial
     steps are the same as those of our non-embargoed security PATCH
     release process.

  2. Create one PR in the private fork to remediate the vulnerability
     in the `main` development branch of the public repository.

  3. Create a branch in the private fork from the appropriate
     `release-M.N` branch, or if that does not yet exist, from the
     appropriate `vM.N.0` tag, and add the necessary changes to remediate
     the vulnerability.

     This branch should have one extra, final commit which adds an entry
     for the new PATCH release version in the `CHANGELOG.md` file,
     describing the release and the security fix it contains.  This
     commit should also include the changes generated by the
     `script/update-version vM.N.P` command, as per the second step
     in our standard release process.

  4. Pull this branch from the private fork of the public repository
     and push it into a separate private repository which has a fresh copy
     of the public `git-lfs/git-lfs` repository, and for which our full
     CI and release GitHub Actions workflows are configured and enabled.

     Note that the private fork created from the draft security advisory
     will not execute GitHub Actions jobs, and so we require the use of
     a separate private repository to run a modified release workflow.

  5. If the `release-M.N` branch does not exist, create it in the
     separate private repository from the `vM.N.0` tag, as described
     in the second step of our regular PATCH release process.

  6. Merge the branch containing the security fix and the commit with the
     new `CHANGELOG.md` entry (and `script/update-version vM.N.P` changes)
     into the `release-M.N` branch, and sign the merge commit with your
     GPG key:

     ```ShellSession
     $ git checkout release-M.N
     $ git merge --no-ff -S \
         -m "Merge pull request from GHSA-abcd-1234-wxyz" \
         -m "release: M.N.P" \
         branch-with-fix-and-changelog
     ```

     Use the GHSA identifier from the draft security advisory in the
     merge commit's description.

  7. Create a GPG-signed tag named `vM.N.P` on the merge commit, using
     the same command from the equivalent step of our standard release
     process:

     ```ShellSession
     $ git tag -s vM.N.P -m vM.N.P
     ```

  8. Check out the `main` branch of the private repository and make
     the following revisions to the `.github/workflows/release.yml` file,
     and then push these changes back to the private repository's `main`
     branch.

     First, change the `on` value to `workflow_dispatch`:

     ```diff
     -on:
     -  push:
     -    tags: '*'
     +on: workflow_dispatch
     ```

     Set the version of Go in each `matrix` context to the same version
     used in the Dockerfiles from our `git-lfs/build-dockers` repository.

     Replace the `${{ github.ref }}` context wherever it appears with the
     `vM.N.P` tag:

     ```diff
     -    ref: ${{ github.ref }}
     +    ref: vM.N.P
     ```

     Revise the `build-docker` and `build-docker-arm` jobs so they do not
     upload the Linux packages generated by the `docker/run_dockers.bsh`
     script to Packagecloud by running the `script/packagecloud.rb` utility.
     Change the jobs to instead upload the Linux packages as job artifacts;
     for instance, for the `build-docker` job:

     ```yaml
         - uses: actions/upload-artifact@v4
           with:
             name: docker-assets
             path: |
               repos/**/*.deb
               repos/**/*.rpm
     ```

     Make sure to use distinct artifact `name`s for each of these jobs,
     i.e., `docker-assets` and `docker-arm-assets`.

  9. Push the `vM.N.P` tag to the private repository, and then cancel
     the GitHub Actions job which runs from the release workflow.

     (This workflow will run because the tag includes our normal
     GitHub Actions workflow definitions without the changes made in
     the previous step, since those exist only on the `main` branch,
     and we do not want to include these temporary workflow changes in
     the security release itself.)

     Confirm that the CI workflow job succeeds.

 10. From the private repository's GitHub Actions page, manually dispatch
     the release workflow.

 11. When the manually-dispatched job is complete, download the
     `release-assets`, `docker-assets`, and `docker-arm-assets` archive
     files from the "Artifacts" section of the job's "Summary" page.

     Share the relevant binaries from the archive files with the other
     collaborating projects during the embargo period.

 12. Before the embargo is due to the lifted, unpack the `release-assets`
     archive file at the top level of this repository and use the
     `script/upload` utility to create a draft release announcement and
     attach the release assets to it, just as for a standard release:

     ```ShellSession
     $ rm -rf bin/releases
     $ unzip /path/to/release-assets.zip

     $ script/upload --skip-verify vM.N.P
     ```

 13. When the embargo is lifted, use the "Merge pull requests" option
     on the draft security advisory page to merge the PR with the security
     fix into the `main` branch of the public repository.

     Note that the `main` branch in the private fork must be up-to-date
     with the corresponding branch in the public repository in order for
     the GitHub UI to determine whether the PR has no merge conflicts.

 14. Publish the security advisory.

 15. Pull the `release-M.N` branch and `vM.N.P` tag from the private
     repository where the release workflow was run manually, and push
     them into the public `git-lfs/git-lfs` repository.

     Immediately cancel the GitHub Actions job which runs from the
     release workflow in the public repository.

     Note that cancelling the release workflow job is important, since
     it will otherwise build new versions of the RPM and Debian Linux
     packages and publish them to Packagecloud.

 16. Upload to Packagecloud the RPM and Debian Linux packages in the
     `docker-assets` and `docker-arm-assets` archive files created earlier
     by the manual release workflow job.

     Note that the `script/packagecloud.rb` utility requires the
     `PACKAGECLOUD_TOKEN` environment variable to contain the current
     Packagecloud account credential token.

     ```ShellSession
     $ mkdir repos
     $ unzip -d repos /path/to/docker-assets.zip
     $ unzip -d repos /path/to/docker-arm-assets.zip

     $ gem install packagecloud-ruby
     $ PACKAGECLOUD_TOKEN="<token>" script/packagecloud.rb
     ```

 17. Finalize the release process using the `script/upload` utility with
     the `--finalize` option.  The script will add GPG signatures to the
     `hashes` and `sha256sums` files and then upload the signed files:

     ```ShellSession
     $ script/upload --finalize vM.N.P
     ```

 18. Publish the release announcement.

 19. Update the `_config.yml` file in the `git-lfs/git-lfs.github.com`
     repository with the new `M.N.P` release version and push the change,
     as described in the final steps of our standard release process.

     In addition, update the `_includes/home/secondary.html` file with
     a banner message regarding the security PATCH release, and push
     the change.

 20. Create a new discussion in the "Announcements" section of the GitHub
     discussion [forum](https://github.com/git-lfs/git-lfs/discussions)
     describing the security PATCH release, and pin the discussion for
     all categories in the forum.
