# Git LFS Agent Instructions

Use these instructions when changing this repository. For contributor
setup, project principles, and the general submission process, also see
[`CONTRIBUTING.md`](CONTRIBUTING.md).

## Plan focused changes

- Search the codebase before proposing a change. Reuse or improve existing
  implementations and helpers; look for related behavior, callers, tests,
  and documentation. Look for parallel or identical problems that should
  be fixed at the same time; use a series of PRs if this would considerably
  expand the scope. Consult Git history when it may explain a design choice
  or defect.
- Keep each PR focused. Put unrelated work in separate PRs. For a large
  refactor or feature, split the work into PRs with discrete purposes,
  unless making closely related, repetitive changes together is clearer.
  For example, a series might first address latent issues, then gradually
  refactor existing code, introduce internal frameworks and tools, and
  finally implement externally visible changes.
- A PR consists of one or more commits. Each commit should make one
  discrete, logical change, or merge recent changes from `main` into the
  topic branch. For logical changes, follow
  [Git's guidelines](https://git-scm.com/docs/SubmittingPatches.html#separate-commits).
  Keep commits as tightly scoped as practical. A relatively large commit
  is appropriate when one logical change requires modifying many identical
  or closely related components, such as changing the return values of a
  commonly used function.
- Order a commit series from preparatory work toward the principal change.
  Preparatory test or helper commits are appropriate when they pass against
  existing behavior.

## Branches and dependent PRs

- Normally base topic branches and PRs on `main`, as described in
  [`CONTRIBUTING.md`](CONTRIBUTING.md).
- Do not flatten, squash, or rebase commits on an open PR's topic branch
  unless a project maintainer advises you to do so.
- Later PRs in a related series may be based on an earlier PR's topic
  branch. Describe the prerequisite PRs and intended merge order. After a
  prerequisite merges to `main`, update dependent topic branches by
  merging the changes from `main` into them as needed.
- Once a PR is approved, a maintainer merges its topic branch into `main`.

## Tests and validation

- After **every commit**, the full Go and shell test suites must pass on
  every platform supported by the CI workflows: Linux, macOS, and Windows.
  Passing only the final PR result does not replace checking each commit.
- From the repository root, run the Go tests:

  ```sh
  make test
  ```

- Run the full shell integration suite, including building its test
  utilities from `t/cmd`:

  ```sh
  make -C t
  ```

- During development, run focused Go tests with a space-separated package
  list, for example:

  ```sh
  make test PKGS='config git/githistory'
  ```

  Run selected shell test scripts with:

  ```sh
  make -C t t-checkout.sh t-fetch.sh
  ```

  Focused tests help iterate quickly; they do not replace the full-suite
  requirement above.
- Run `make fmt` with `goimports` installed to format Go source files, and
  inspect the resulting changes.
- Add or update tests in the same commit as behavior changes, unless
  existing tests already provide sufficient coverage. Preparatory test or
  helper commits may precede the behavior change if they pass against
  existing behavior. Do not leave a failing test in an earlier commit.
- Shell integration tests are generally preferred, and are required for
  changes involving Git client execution, I/O, or file operations. Use Go
  tests for internal package contracts and related small cases as appropriate.
- Reuse existing shell tests, helpers, assertions, and utilities. Shell
  test helper utilities are in `t/cmd`. For a new shell test, follow recent
  examples. When modifying an existing test or adding a closely related
  test in the same file, balance continuity with the existing tests against
  the idioms of newer tests. Use two-space indentation; indent commands
  inside subshells or function definitions, and indent them further inside
  `pushd`/`popd` pairs.

## Go and file style

- Follow the Go [best practices](https://google.github.io/styleguide/go/best-practices.html),
  [Effective Go](https://go.dev/doc/effective_go), and
  [Go code review comments](https://go.dev/wiki/CodeReviewComments).
- Use Unix-style LF line endings for source files, including Asciidoc and
  Markdown. End the last line with a newline and do not add an extra blank
  line at the end of the file.

## Commit messages

 - Commit descriptions must be detailed, accurate, and understandable
   without reading the code. Write in the style of detailed technical
   documentation, avoiding jargon and buzzwords. Follow the relevant
   guidance in
  [Git's Guidelines](https://git-scm.com/docs/SubmittingPatches.html#describe-changes).
- Write in English, using full sentences and logically separated
  paragraphs in the body. Explain the existing behavior and problem, then
  what the change does and why it is appropriate. When relevant, describe
  observable effects, compatibility, tests, alternatives, and historical
  context. Summarize important reasoning rather than relying on links
  alone. References to the commits or PRs that introduced related
  behavior are useful.
- Keep the subject to 50 characters or fewer. Begin with an area prefix,
  followed by a colon and space, then a short subject phrase. Use the
  affected package or directory (for example, `tools` or `git,lfs`), a
  file path, or a glob-style path. List multiple areas alphabetically.
  The prefix may be omitted when a change spans many packages.
- Use lowercase in the subject phrase unless a term requires
  capitalization. Prefer the imperative mood, omit a final period, and
  avoid unnecessary punctuation. Separate the subject from the body with
  a blank line.
- Keep body lines to 72 characters or fewer, except for unbreakable
  strings such as links. Indent hyperlinks and block quotations by two
  spaces.
- Write function and method names with `()` and without quotes, such as
  `main()` and `doCleanup()`. Initially identify Unix system calls and
  POSIX programs with their manual section numbers (for example,
  `socket(2)` and `grep(1)`); later references in the same description
  may be shorter. Use man7.org for section numbers unless a
  Darwin/BSD-specific reference applies. Link Windows system calls to
  their API documentation.
- Put filenames, package and program names, shell commands, Boolean
  values, configuration options and values, and unexported Go names in
  double quotes. For example: `"go.mod"`, `"net/http"`,
  `"git lfs pull"`, `"true"`, and `"apiClient"`. Exported Go names may
  be left unquoted. Write environment variable names in uppercase unless
  the variable is explicitly mixed-case or lowercase.
- Initialisms and acronyms are acceptable, but they should be written out
  as a full phrase with initial capitals the first time they appear
  unless they are commonly used in the context of this project. For
  example, `LFS` and `HTTP` may be used without explication, but use
  `Long-Term Support (LTS)` and `Subject Alternative Name (SAN)`
  when those first appear in a commit description.
- If a commit only partially resolves an issue, reference the issue
  number in the commit's description but do not precede this reference
  with a GitHub issue-closing keyword such as `Fixes` or `Resolves`.
  Otherwise, the issue may close when the PR is merged.  See GitHub's
  [issue-linking guidance](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).

## Pull requests

- Give the PR a concise, descriptive title for its principal goal.
  Because it may appear as a single line in `CHANGELOG.md`, normally keep
  it to about 60 characters or fewer.
- Begin the description with a few short paragraphs summarizing the
  change and its purpose, using full sentences. Explain the problem, the
  intended result, and relevant user-visible or compatibility effects.
- For larger changes, add `####` subsections as useful. Avoid bullet
  lists unless they make a set of closely related items easier to
  understand.
- Link related issues, earlier or parallel work, and prerequisite or
  follow-up PRs. For a PR series, identify dependencies and the intended
  merge order. For multi-commit PRs, say whether commit-by-commit review
  is useful and whether reviewers should ignore whitespace changes.
- Describe the test coverage and any checks that remain outstanding.
- Note when any portion of the PR and its commits were written by
  agentic tools. A human contributor must attest that they have reviewed
  all such materials in the PR.
- If the PR fully resolves an issue, include a closing reference near
  the end of the description, using a standalone line:
  `Resolves #<issue-number>.`
- If the PR only partially resolves an issue, reference the issue number
  in the PR's description but do not precede this reference with a
  GitHub issue-closing keyword, and do not manually link the PR
  to the issue in a way that would close it on merge. See GitHub's
  [issue-linking guidance](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue).
- If an issue raised by someone else is resolved or discussed, credit
  its original reporter on the final line, after any issue reference:
  `/cc @<name> as reporter.`
- When replying to PR reviews or other comments on a pull request, apply
  the same standards of language use and formatting as pertain to
  commit and PR descriptions. Use full English sentences and paragraphs.
- Avoid unnecessary updates to a PR such as automatic merges from the
  `main` branch, unless recent changes in the `main` branch materially
  affect the contents of a PR.

## Security, compatibility, and documentation

- Report potential security vulnerabilities privately, following
  [`SECURITY.md`](SECURITY.md).
- A security-fix commit must include shell integration tests that
  demonstrate the exploit without the fix and show that the commit
  prevents it. The commit description must explain the security change.
  Apply the same high standard of tests and explanation when replacing
  or removing code added for security purposes; explain why the result
  remains secure.
- Clearly identify changes that break backward compatibility; they may
  need to wait for a future major release.
- Where possible without breaking compatibility with previous releases
  of the same major version of Git LFS, respect Git's established choices
  for command names, arguments, configuration, file operations, caching,
  and indexing.
- Update documentation and code comments in the same commit as the
  related functionality. Document supported configuration options and
  command-line arguments in the Asciidoc manual pages under `docs/man`.
  Describe changes to external APIs in Markdown under `docs/api`.
- Propose new major features in a separate PR that adds a Markdown
  document under `docs/proposals`, without accompanying code changes.
