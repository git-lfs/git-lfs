---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**System environment**
The version of your operating system, plus any relevant information about platform or configuration (e.g., container or CI usage, Cygwin, WSL, or non-Basic authentication).  If relevant, include the output of `git config -l` as a code block.

Please also mention the usage of any proxy, including any TLS MITM device or non-default antivirus or firewall.

**Output of `git lfs env`**
The output of running `git lfs env` as a code block.

**Additional context**
Any other relevant context about the problem here.

If you're having problems trying to push or pull data, please run the command with `GIT_TRACE=1 GIT_TRANSFER_TRACE=1 GIT_CURL_VERBOSE=1` and include it inline or attach it as a text file.  In a bash or other POSIX shell, you can simply prepend this string and a space to the command.

<!--
Please note: if you're receiving a message from the server side (including a
`batch response` message), please contact your Git hosting provider.  This
repository is for the Git LFS client only; problems with GitHub's server-side
LFS support should be reported to them as described in the `CONTRIBUTING.md`
file.
-->
