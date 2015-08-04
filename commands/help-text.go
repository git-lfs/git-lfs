package commands 

const (
git_lfs_checkout_HelpText = `git-lfs-checkout(1) -- Update working copy with file content if available
=========================================================================

## SYNOPSIS

'git lfs checkout' <filespec>...

## DESCRIPTION

Try to ensure that the working copy contains file content for Git LFS objects
for the current ref, if the object data is available. Does not download any
content, see git-lfs-fetch(1) for that. 

Checkout scans the current ref for all LFS objects that would be required, then
where a file is either missing in the working copy, or contains placeholder
pointer content with the same SHA, the real file content is written, provided
we have it in the local store. Modified files are never overwritten.

Filespecs can be provided as arguments to restrict the files which are updated.

## EXAMPLES

* Checkout all files that are missing or placeholders

  'git lfs checkout'

* Checkout a specific couple of files

  'git lfs checkout path/to/file1.png path/to.file2.png'

## SEE ALSO

git-lfs-fetch(1), git-lfs-pull(1).

Part of the git-lfs(1) suite.

`
git_lfs_clean_HelpText = `git-lfs-clean(1) -- Git clean filter that converts large files to pointers
==========================================================================

## SYNOPSIS

'git lfs clean' <path>

## DESCRIPTION

Read the contents of a large file from standard input, and write a Git
LFS pointer file for that file to standard output.

Clean is typically run by Git's clean filter, configured by the repository's
Git attributes.

## SEE ALSO

git-lfs-init(1), git-lfs-push(1), gitattributes(5).

Part of the git-lfs(1) suite.
`
git_lfs_env_HelpText = `git-lfs-env(1) -- Display the Git LFS environment
=================================================

## SYNOPSIS

'git lfs env'

## DESCRIPTION

Display the current Git LFS environment.

## SEE ALSO

Part of the git-lfs(1) suite.
`
git_lfs_fetch_HelpText = `git-lfs-fetch(1) -- Download all Git LFS files for a given ref
==============================================================

## SYNOPSIS

'git lfs fetch' [<ref>...]

## DESCRIPTION

Download any Git LFS objects for the given refs. If no refs are given,
the currently checked out ref will be used.

This does not update the working copy.

## EXAMPLES

* Fetch the LFS objects for the current ref

  'git lfs fetch'

* Fetch the LFS objects for a branch

  'git lfs fetch mybranch'

* Fetch the LFS objects for 2 branches and a commit

  'git lfs fetch master mybranch e445b45c1c9c6282614f201b62778e4c0688b5c8'

## SEE ALSO

git-lfs-checkout(1), git-lfs-pull(1).

Part of the git-lfs(1) suite.

`
git_lfs_fsck_HelpText = `git-lfs-fsck(1) -- Check GIT LFS files for consistency
======================================================

## SYNOPSIS

'git lfs fsck'

## DESCRIPTION

Checks all GIT LFS files in the current HEAD for consistency.

Corrupted files are moved to ".git/lfs/bad".

## SEE ALSO

git-lfs-ls-files(1), git-lfs-status(1).

Part of the git-lfs(1) suite.
`
git_lfs_init_HelpText = `git-lfs-init(1) -- Ensure Git LFS is configured properly
========================================================

## SYNOPSIS

'git lfs init'<br>
'git lfs init' --force

## DESCRIPTION

Perform the following actions to ensure that Git LFS is setup properly:

* Set up the clean and smudge filters under the name "lfs" in the global Git
  config.
* Install a pre-push hook to run git-lfs-pre-push(1) for the current repository,
  if run from inside one.

## OPTIONS

Without any options, 'git lfs init' will only setup the "lfs" smudge and clean
filters if they are not already set.

* '--force':
    Sets the "lfs" smudge and clean filters, overwriting existing values.

## SEE ALSO

git-lfs-uninit(1).

Part of the git-lfs(1) suite.
`
git_lfs_logs_HelpText = `git-lfs-logs(1) - Show errors from the git-lfs command
======================================================

## SYNOPSIS

'git lfs logs'<br>
'git lfs logs' <file><br>
'git lfs logs' --clear<br>
'git lfs logs' --boomtown<br>

## DESCRIPTION

Display errors from the git-lfs command.  Any time it crashes, the details are
saved to ".git/lfs/logs".

## OPTIONS

Without any options, 'git lfs logs' simply shows the list of error logs.

* <file>:
    Shows the specified error log.  Use "last" to show the most recent error.

* '--clear':
    Clears all of the existing logged errors.

* '--boomtown':
    Triggers a dummy exception.

## SEE ALSO

Part of the git-lfs(1) suite.
`
git_lfs_ls_files_HelpText = `git-lfs-ls-files(1) -- Show information about Git LFS files in the index and working tree
=========================================================================================

## SYNOPSIS

'git lfs ls-files' [<reference>]

## DESCRIPTION

Display paths of Git LFS files that are found in the given reference.  If no
reference is given, scan the currently checked-out branch.

## SEE ALSO

git-lfs-status(1).

Part of the git-lfs(1) suite.
`
git_lfs_pointer_HelpText = `git-lfs-pointer(1) -- Build and compare pointers
================================================

## SYNOPSIS

'git lfs pointer --file=path/to/file'<br>
'git lfs pointer --file=path/to/file --pointer=path/to/pointer'<br>
'git lfs pointer --file=path/to/file --stdin'

## Description

Builds and optionally compares generated pointer files to ensure consistency
between different Git LFS implementations.

## OPTIONS

* '--file':
    A local file to build the pointer from.

* '--pointer':
    A local file including the contents of a pointer generated from another
    implementation.  This is compared to the pointer generated from '--file'.

* '--stdin':
    Reads the pointer from STDIN to compare with the pointer generated from
    '--file'.

## SEE ALSO

Part of the git-lfs(1) suite.
`
git_lfs_pre_push_HelpText = `git-lfs-pre-push(1) -- Git pre-push hook implementation
=======================================================

## SYNOPSIS

'git lfs pre-push' <remote> [remoteurl]

## DESCRIPTION

Responds to Git pre-hook events. It reads the range of commits from STDIN, in
the following format:

    <local-ref> SP <local-sha1> SP <remote-ref> SP <remote-sha1> \n

It also takes the remote name and URL as arguments.

## SEE ALSO

git-lfs-clean(1), git-lfs-push(1).

Part of the git-lfs(1) suite.
`
git_lfs_pull_HelpText = `git-lfs-pull(1) -- Download all Git LFS files for current ref & checkout
========================================================================

## SYNOPSIS

'git lfs pull'

## DESCRIPTION

Download Git LFS objects for the currently checked out ref, and update
the working copy with the downloaded content if required.

This is equivalent to running the following 2 commands:

git lfs fetch
git lfs checkout

## EXAMPLES

## SEE ALSO

git-lfs-fetch(1), git-lfs-checkout(1).

Part of the git-lfs(1) suite.

`
git_lfs_push_HelpText = `git-lfs-push(1) -- Push queued large files to the Git LFS endpoint
==================================================================

## SYNOPSIS

'git lfs push' <remote> [branch]<br>
'git lfs push' --object-id <remote> <oid1> <oid2> ...

## DESCRIPTION

Upload Git LFS files to the configured endpoint for the current Git remote.

This command shouldn't be necessary since Git LFS automatically sets up a
pre-push hook for each repository.

## OPTIONS

* '--object-id':
    This pushes only the object OIDs listed at the end of the command, separated
    by spaces.

* '--dry-run':
    Print the files that would be pushed, without actually pushing them.

* '--stdin':
    Read the remote and branch on stdin. This is used in conjunction with the
    pre-push hook and must be in the format used by the pre-push hook:
    <local-ref> <local-sha1> <remote-ref> <remote-sha1>. If --stdin is used
    the command line arguments are ignored.  NOTE: This is deprecated in favor
    of the 'pre-push' command.

## SEE ALSO

git-lfs-clean(1), git-lfs-pre-push(1).

Part of the git-lfs(1) suite.
`
git_lfs_smudge_HelpText = `git-lfs-smudge(1) -- Git smudge filter that converts pointer in blobs to the actual content
===========================================================================================

## SYNOPSIS

'git lfs smudge' [<path>]

## DESCRIPTION

Read a Git LFS pointer file from standard input and write the contents
of the corresponding large file to standard output.  If needed,
download the file's contents from the Git LFS endpoint.  The <path>
argument, if provided, is only used for a progress bar.

Smudge is typically run by Git's smudge filter, configured by the repository's
Git attributes.

## OPTIONS

Without any options, 'git lfs smudge' outputs the raw Git LFS content to
standard output.

* '--info':
    Display the file size and the local path to the Git LFS file.  If the file
    does not exist, show '--'.

## SEE ALSO

git-lfs-init(1), gitattributes(5).

Part of the git-lfs(1) suite.
`
git_lfs_status_HelpText = `git-lfs-status(1) -- Show the status of Git LFS files in the working tree
=========================================================================

## SYNOPSIS

'git lfs status' [<options>]

## DESCRIPTION

Display paths of Git LFS objects that

* have not been pushed to the Git LFS server.  These are large files
  that would be uploaded by 'git push'.

* have differences between the index file and the current HEAD commit.
  These are large files that would be committed by 'git commit'.

* have differences between the working tree and the index file.  These
  are files that could be staged using 'git add'.

## OPTIONS

* '--porcelain':
    Give the output in an easy-to-parse format for scripts.

## SEE ALSO

git-lfs-ls-files(1).

Part of the git-lfs(1) suite.
`
git_lfs_track_HelpText = `git-lfs-track(1) - View or add Git LFS paths to Git attributes
==============================================================

## SYNOPSIS

'git lfs track' [<path>...]

## DESCRIPTION

Start tracking the given path(s) through Git LFS.  The <path> argument
can be a pattern or a file path.  If no paths are provided, simply list
the currently-tracked paths.

## EXAMPLES

* List the paths that Git LFS is currently tracking:

    'git lfs track'

* Configure Git LFS to track GIF files:

    'git lfs track '*.gif''

## SEE ALSO

git-lfs-untrack(1), git-lfs-init(1), gitattributes(5).

Part of the git-lfs(1) suite.
`
git_lfs_uninit_HelpText = `git-lfs-uninit(1) -- Remove Git LFS configuration
=================================================

## SYNOPSIS

'git lfs uninit'

## DESCRIPTION

Perform the following actions to remove the Git LFS configuration:

* Remove the "lfs" clean and smudge filters from the global Git config.
* Uninstall the Git LFS pre-push hook if run from inside a Git repository.

## SEE ALSO

git-lfs-init(1).

Part of the git-lfs(1) suite.
`
git_lfs_untrack_HelpText = `git-lfs-untrack(1) - Remove Git LFS paths from Git Attributes
=============================================================

## SYNOPSIS

'git lfs untrack' <path>...

## DESCRIPTION

Stop tracking the given path(s) through Git LFS.  The <path> argument
can be a glob pattern or a file path.

## EXAMPLES

* Configure Git LFS to stop tracking GIF files:

    'git lfs untrack '*.gif''

## SEE ALSO

git-lfs-track(1), git-lfs-init(1), gitattributes(5).

Part of the git-lfs(1) suite.
`
git_lfs_update_HelpText = `git-lfs-update(1) -- Update Git hooks
=====================================

## SYNOPSIS

'git lfs update' [--force]

## DESCRIPTION

Updates the Git hooks used by Git LFS. Silently upgrades known hook contents.
Pass '--force' to upgrade the hooks, clobbering any existing contents.

## SEE ALSO

Part of the git-lfs(1) suite.
`
git_lfs_HelpText = `git-lfs(1) -- Work with large files in Git repositories
=======================================================

## SYNOPSIS

'git lfs' <command> [<args>]

## DESCRIPTION

Git LFS is a system for managing and versioning large files in
association with a Git repository.  Instead of storing the large files
within the Git repository as blobs, Git LFS stores special "pointer
files" in the repository, while storing the actual file contents on a
Git LFS server.  The contents of the large file are downloaded
automatically when needed, for example when a Git branch containing
the large file is checked out.

Git LFS works by using a "smudge" filter to look up the large file
contents based on the pointer file, and a "clean" filter to create a
new version of the pointer file when the large file's contents change.
It also uses a 'pre-push' hook to upload the large file contents to
the Git LFS server whenever a commit containing a new large file
version is about to be pushed to the corresponding Git server.

## COMMANDS

Like Git, Git LFS commands are separated into high level ("porcelain")
commands and low level ("plumbing") commands.

### High-level commands (porcelain)

* git-lfs-env(1):
    Display the Git LFS environment.
* git-lfs-fsck(1):
    Check GIT LFS files for consistency.
* git-lfs-init(1):
    Ensure Git LFS is configured properly.
* git-lfs-logs(1):
    Show errors from the git-lfs command.
* git-lfs-ls-files(1):
    Show information about Git LFS files in the index and working tree.
* git-lfs-push(1):
    Push queued large files to the Git LFS endpoint.
* git-lfs-status(1):
    Show the status of Git LFS files in the working tree.
* git-lfs-track(1):
    View or add Git LFS paths to Git attributes.
* git-lfs-untrack(1):
    Remove Git LFS paths from Git Attributes.
* git-lfs-update(1):
    Update Git hooks for the current Git repository.

### Low level commands (plumbing)

* git-lfs-clean(1):
    Git clean filter that converts large files to pointers.
* git-lfs-pointer(1):
    Build and compare pointers.
* git-lfs-pre-push(1):
    Git pre-push hook implementation.
* git-lfs-smudge(1):
    Git smudge filter that converts pointer in blobs to the actual content.
`
)
