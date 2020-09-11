## Security

Git LFS is a public, open-source project supported by GitHub and a
broad community of other organizations and individual contributors.
The Git LFS community takes the security of our project seriously,
including the all of source code repositories managed through
our GitHub [organization](https://github.com/git-lfs).

If you believe you have found a security vulnerability in any Git LFS
client software repository, please report it to us as described below.

If you believe you have found a security vulnerability in a Git LFS API
service, please report it to the relevant hosting company (e.g., Atlassian,
GitLab, GitHub, etc.) by following their preferred security report process.

## Reporting Security Issues

*Please do not report security vulnerabilities through public GitHub issues.*

If you believe you have found a security vulnerability in the Git LFS
client software, including any of our Go modules such as
[gitobj](https://github.com/git-lfs/gitobj) or
[pktline](https://github.com/git-lfs/pktline), please report it
by email to one of the Git LFS [core team members](https://github.com/git-lfs/git-lfs#core-team).

Email addresses for core team members may be found either on their
personal GitHub pages or simply by searching through the Git history
for this project; all commits from core team members should have their
email address in the `Author` Git log field.

If possible, encrypt your message with the core team member's PGP key.
These may be located by searching a public keyserver or from the
team member [list](https://github.com/git-lfs/git-lfs#core-team)
on our home page.

If you do not receive a timely response (generally within 24 hours of the
first working day after your submission), please follow up by email
with them and another core team member as well.

Please include the requested information listed below (as much as you can provide) to help us better understand the nature and scope of the possible issue:

  * Type of issue (e.g. buffer overflow, cross-site scripting, etc.)
  * Full paths of source file(s) related to the manifestation of the issue
  * The location of the affected source code (tag/branch/commit or direct URL)
  * Any special configuration required to reproduce the issue
  * Step-by-step instructions to reproduce the issue
  * Proof-of-concept or exploit code (if possible)
  * Impact of the issue, including how an attacker might exploit the issue

This information will help us triage your report more quickly.

We also recommend reviewing our [guidelines](CONTRIBUTING.md) for
contributors and our [Open Code of Conduct](CODE-OF-CONDUCT.md).

Note that because the Git LFS client is a public open-source project,
it is not enrolled in any bug bounty programs; however, implementations
of the Git LFS API service may be, depending on the hosting provider.

## Preferred Languages

We prefer all communications to be in English.
