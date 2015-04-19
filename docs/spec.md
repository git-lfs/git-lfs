# Git LFS Specification

This is a general guide for Git LFS clients.  Typically it should be
implemented by a command line `git-lfs` tool, but the details may be useful
for other tools.

## The Pointer

The core Git LFS idea is that instead of writing large blobs to a Git repository,
only a pointer file is written.

```
version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345
(ending \n)
```

The pointer file should be small (less than 200 bytes), and consist of only
ASCII characters.  Libraries that generate this should write the file
identically, so that different implementations write consistent pointers that
translate to the same Git blob OID.  This means:

* Use properties "version", "oid", and "size" in that order.
* Separate the property from its value with a single space.
* Oid has a "sha256:" prefix.  No other hashing methods are currently supported
for Git LFS oids.
* Size is in bytes.

Note: Earlier versions only contained the OID, with a `# comment` above it.
Here's some ruby code to parse older pointer files.

```
# data is a string of the content
# last full line contains the oid
return nil unless data.size < 100
lines = data.
  strip.      # strip ending whitespace
  split("\n") # split by line breaks

# We look for a comment line, and the phrase `git-media` somewhere
lines[0] =~ /# (.*git-media|external)/ && lines.last
```

That code returns the OID, which should be on the last line.  The OID is
generated from the SHA-256 signature of the file's contents.

## The Server

Git LFS needs a URL endpoint to talk to a remote server.  A Git repository
can have different Git LFS endpoints for different remotes.  Here is the list
of rules that Git LFS uses to determine a repository's Git LFS server:

1. The `lfs.url` string.
2. The `remote.{name}.lfsurl` string.
3. Append `/info/lfs` to the remote URL.  Only works with HTTPS URLs.

Git LFS runs two `git config` commands to build up the list of values that it
uses:

1. `git config -l -f .gitconfig` - This file is checked into the repository and
can set defaults for every user that clones the repository.
2. `git config -l` - A user's local git configuration can override any settings
from `.gitconfig`.

Here's a sample Git config file with the optional remote and Git LFS
configuration options:

```
[core]
  repositoryformatversion = 0
[lfs]
  url = "https://github.com/github/git-lfs.git/info/lfs"
[remote "origin"]
  url = https://github.com/github/git-lfs
  fetch = +refs/heads/*:refs/remotes/origin/*
  lfsurl = "https://github.com/github/git-lfs.git/info/lfs"
```

Git LFS uses `git credential` to fetch credentials for HTTPS requests.  Setup
a credential cache helper to save passwords for future users.

## Intercepting Git

Git LFS uses the `clean` and `smudge` filters to decide which files use it.  The
global filters can be set up with `git lfs init`:

```
$ git lfs init
```

These filters ensure that large files aren't written into the repository proper,
instead being stored locally at `.git/lfs/objects/{OID-PATH}` (where `{OID-PATH}`
is a sharded filepath of the form `OID[0:2]/OID[2:4]/OID`), synchronized with
the Git LFS server as necessary.

The `clean` filter runs as files are added to repositories.  Git sends the
content of the file being added as STDIN, and expects the content to write
to Git as STDOUT.

* Stream binary content from STDIN to a temp file, while calculating its SHA-256
signature.
* Check for the file at `.git/lfs/objects/{OID-PATH}`.
* If it does not exist:
  * Queue the OID to be uploaded.
  * Move the temp file to `.git/lfs/objects/{OID-PATH}`.
* Delete the temp file.
* Write the pointer file to STDOUT.

Note that the `clean` filter does not push the file to the server.  Use the
`git lfs sync` command to do that.

The `smudge` filter runs as files are being checked out from the Git repository
to the working directory.  Git sends the content of the Git blob as STDIN, and
expects the content to write to the working directory as STDOUT.

* Read 100 bytes.
* If the content is ASCII and matches the pointer file format:
  * Look for the file in `.git/lfs/objects/{OID-PATH}`.
  * If it's not there, download it from the server.
  * Read its contents to STDOUT
* Otherwise, simply pass the STDIN out through STDOUT.

The `.gitattributes` file controls when the filters run.  Here's a sample file
runs all mp3 and zip files through Git LFS:

```
$ cat .gitattributes
*.mp3 filter=lfs -crlf
*.zip filter=lfs -crlf
```

Use the `git lfs path` command to view and add to `.gitattributes`.
