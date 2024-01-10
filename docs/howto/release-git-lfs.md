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
      | git-lfs-linux-ppc64le-v@{version}.tar.gz | linux (generic) | ppc64le |
      | git-lfs-linux-s390x-v@{version}.tar.gz | linux (generic) | s390x |
      | git-lfs-linux-loong64-v@{version}.tar.gz | linux (generic) | loong64 |

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

  1. First, we write the release notes and do the housekeeping required to
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
       date.  This heading should be consistent with the exising style in the
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

  2. Then, push the `release-next` branch and create a pull request with your
     changes from the branch.  If you're building a MAJOR or MINOR release,
     set the base to the `main` branch.  Otherwise, set the base to the
     `release-M.N` branch.

     * Add the `release` label to the PR.

     * In the PR description, consider uploading builds for implementers to
       use in testing.  These can be generated from a local tag, which
       does not need to be signed (but must be annotated).  Check that the
       local version of Go is equivalent to the most recent one used by the
       GitHub Actions workflows for the release branch, which may be different
       from that used on the `main` branch.  For a patch release in particular
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

  3. Once approved and verified, merge the pull request you created in the
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

  4. Push the tag, via:

     ```ShellSession
     $ git push origin vM.N.P
     ```

     This will kick off the process of building the release artifacts.  This
     process will take somewhere between 45 minutes and an hour.  When it's
     done, you'll end up with a draft release in the repository for the version
     in question.

  5. From the command line, finalize the release process by signing the release:

     ```ShellSession
     $ script/upload --finalize vM.N.P
     ```

     Note that this script requires GnuPG as well as Ruby (with the OpenSSL
     gem) and several other tools.  You will need to provide your GitHub
     credentials in your `~/.netrc` file or via a `GITHUB_TOKEN` environment
     variable.

     If you want to inspect the data before approving it, pass the `--inspect`
     option, which will drop you to a shell and let you look at things.  If the
     shell exits successfully, the build will be signed; otherwise, the process
     will be aborted.

  6. Publish the release on GitHub, assuming it looks correct.

  7. Move any remaining items out of the milestone for the current release to a
     future release and close the milestone.

  8. Update the `_config.yml` file in
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

  9. Create a GitHub PR to update the Homebrew formula for Git LFS with
     the `brew bump-formula-pr` command on a macOS system.  The SHA-256 value
     should correspond with the packaged artifact containing the new
     release's source files which is available at the given URL:

     ```
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

  1. If the `release-M.N` branch does not already exist, create it from
     the corresponding MINOR release tag (or MAJOR release tag, if no
     MINOR releases have been made since the last MAJOR release):

     ```ShellSession
     $ git checkout -b release-M.N vM.N.0
     ```

     If the release branch already exists because this is not the first
     patch release for the given MINOR (or MAJOR) release, simply checkout
     the `release-M.N` branch, and ensure that you have the latest changes
     from the remote.

  2. Gather a set of potential candidates to backport to the `release-M.N`
     branch with:

     ```ShellSession
     $ git log --merges --first-parent vM.N.(P-1)...main
     ```

   3. For each merge that you want to backport, run:

      ```ShellSession
      $ git cherry-pick -m1 <SHA-1>
      ```

      This will cherry-pick the merge onto your release branch, using
      the `-m1` option to specify that the first parent of the merge
      corresponds to the mainline.

   4. Then follow the [guidelines](#building-a-release) above, using the
      `release-M.N` branch as the base for the new PATCH release.
