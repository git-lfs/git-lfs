# Git LFS Specification

This is a general guide for Git LFS clients.  Typically it should be
implemented by a command line `git-lfs` tool, but the details may be useful
for other tools.

## The Pointer

The core Git LFS idea is that instead of writing large blobs to a Git repository,
only a pointer file is written.

* Pointer files are text files which MUST contain only UTF-8 characters.
* Each line MUST be of the format `{key} {value}\n` (trailing unix newline).
* Only a single space character between `{key}` and `{value}`.
* Keys MUST only use the characters `[a-z] [0-9] . -`.
* The first key is _always_ `version`.
* Lines of key/value pairs MUST be sorted alphabetically in ascending order
(with the exception of `version`, which is always first).
* Values MUST NOT contain return or newline characters.
* Pointer files SHOULD NOT have the executable bit set when checked out from Git.

The required keys are:

* `version` is a URL that identifies the pointer file spec.  Parsers MUST use
simple string comparison on the version, without any URL parsing or
normalization.  It is case sensitive, and %-encoding is discouraged.
* `oid` tracks the unique object id for the file, prefixed by its hashing
method: `{hash-method}:{hash}`.  Currently, only `sha256` is supported.
* `size` is in bytes.

Example of a v1 text pointer:

```
version https://git-lfs.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345
(ending \n)
```

Blobs created with the pre-release version of the tool generated files with
a different version URL.  Git LFS can read these files, but writes them using
the version URL above.

```
version https://hawser.github.com/spec/v1
oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
size 12345
(ending \n)
```

For testing compliance of any tool generating its own pointer files, the
reference is this official Git LFS tool:

**NOTE:** exact pointer command behavior TBD!

* Tools that parse and regenerate pointer files MUST preserve keys that they
don't know or care about.
* Run the `pointer` command to generate a pointer file for the given local
file:

    ```
    $ git lfs pointer --file=path/to/file
    Git LFS pointer for path/to/file:

    version https://git-lfs.github.com/spec/v1
    oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
    size 12345
    ```

* Run `pointer` to compare the blob OID of a pointer file built by Git LFS with
a pointer built by another tool.

  * Write the other implementation's pointer to "other/pointer/file":

    ```
    $ git lfs pointer --file=path/to/file --pointer=other/pointer/file
    Git LFS pointer for path/to/file:

    version https://git-lfs.github.com/spec/v1
    oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
    size 12345

    Blob OID: 60c8d8ab2adcf57a391163a7eeb0cdb8bf348e44

    Pointer from other/pointer/file
    version https://git-lfs.github.com/spec/v1
    oid sha256 4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
    size 12345

    Blob OID: 08e593eeaa1b6032e971684825b4b60517e0638d

    Pointers do not match
    ```

  * It can also read STDIN to get the other implementation's pointer:

    ```
    $ cat other/pointer/file | git lfs pointer --file=path/to/file --stdin
    Git LFS pointer for path/to/file:

    version https://git-lfs.github.com/spec/v1
    oid sha256:4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
    size 12345

    Blob OID: 60c8d8ab2adcf57a391163a7eeb0cdb8bf348e44

    Pointer from STDIN
    version https://git-lfs.github.com/spec/v1
    oid sha256 4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393
    size 12345

    Blob OID: 08e593eeaa1b6032e971684825b4b60517e0638d

    Pointers do not match
    ```

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
the Git LFS server as necessary.  Here is a sample path to a file:

    .git/lfs/objects/4d/7a/4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393

The `clean` filter runs as files are added to repositories.  Git sends the
content of the file being added as STDIN, and expects the content to write
to Git as STDOUT.

* Stream binary content from STDIN to a temp file, while calculating its SHA-256
signature.
* Atomically move the temp file to `.git/lfs/objects/{OID-PATH}` if it does not
exist, and the sha-256 signature of the contents matches the given OID.
* Delete the temp file.
* Write the pointer file to STDOUT.

Note that the `clean` filter does not push the file to the server.  Use the
`git push` command to do that (lfs files are pushed before commits in a pre-push hook).

The `smudge` filter runs as files are being checked out from the Git repository
to the working directory.  Git sends the content of the Git blob as STDIN, and
expects the content to write to the working directory as STDOUT.

* Read 100 bytes.
* If the content is ASCII and matches the pointer file format:
  * Look for the file in `.git/lfs/objects/{OID-PATH}`.
  * If it's not there, download it from the server.
  * Write its contents to STDOUT
* Otherwise, simply pass the STDIN out through STDOUT.

The `.gitattributes` file controls when the filters run.  Here's a sample file that
runs all mp3 and zip files through Git LFS:

```
$ cat .gitattributes
*.mp3 filter=lfs -crlf
*.zip filter=lfs -crlf
```

Use the `git lfs track` command to view and add to `.gitattributes`.

## Extending LFS

Teams who use Git LFS often have custom requirements for how the pointer files and 
blobs should be handled.  Some examples of extensions that could be built:

* Compress large files on clean, uncompress them on smudge/fetch
* Encrypt files on clean, decrypt on smudge/fetch
* Scan files on clean to make sure they don't contain sensitive information

The basic extensibilty model is that LFS extensions must be registered explicitly, and
they will be invoked on clean and smudge to manipulate the contents of the files as 
needed.  On clean, LFS itself ensures that the pointer file is updated with all the 
information needed to be able to smudge correctly, and the extensions never write to the 
pointer file.

Note that LFS is currently transitioning away from using the Git smudge filter, in favor 
of smudging all files using "git-lfs fetch" post checkout.  However, that detail should 
be transparent to extensions, since they are still invoked on a per-file basis.

### Registration

To register an LFS extension, it must be added to the Git config.  Each extension needs 
to define:

* Its unique name.  This will be used as part of the key in the pointer file.
* The command to run on clean
* The command to run on smudge/fetch
* The priority of the extension, which must be a unique, positive integer

The sequence "%f" in the clean and smudge commands will be replaced by the filename being 
processed.

Here's an example extension registration in the Git config:

```
[lfs-extension "foo"]
  clean = git-lfs-foo clean %f
  smudge = git-lfs-foo smudge %f
  priority = 1
[lfs-extension "bar"]
  clean = git-lfs-bar clean %f
  smudge = git-lfs-bar smudge %f
  priority = 2
```

### Clean

When staging a file, Git invokes the LFS clean filter, as described earlier.  If no 
extensions are installed, the LFS clean filter reads bytes from STDIN, calculates the 
SHA-256 signature, and writes the bytes to a temp file.  It then moves the temp file into 
the appropriate place in .git/lfs/objects and writes a valid pointer file to STDOUT.

When an extension is installed, LFS will invoke the extension to do additional processing 
on the bytes before writing them into the temp file.  If multiple extensions are 
installed, they are invoked in the order defined by their priority.  LFS will also insert 
a key in the pointer file for each extension that was invoked, indicating both the order 
that the extension was invoked and the oid of the file before that extension was invoked.  
All of that information is required to be able to reliably smudge the file later.  Each 
new line in the pointer file will be of the form

`extension-{priority}-{name} {hash-method}:{hash-of-input-to-extension} `

This naming ensures that all extensions are written in both alphabetical and priority 
order, and also shows the progression of changes to the oid as it is processed by the 
extensions.

Here's an example sequence, assuming extensions foo and bar are installed, as shown in 
the previous section.

* Git passes the original contents of the file to LFS clean over STDIN
* LFS reads those bytes and calculates the original SHA-256 signature as it does so
* LFS streams the bytes to STDIN of lfs-extension.foo.clean, which is expected to write 
those bytes, modified or not, to its STDOUT
* LFS reads the bytes from STDOUT of lfs-extension.foo.clean, calculates the SHA-256 
signature, and writes them to STDIN of lfs-extension.bar.clean, which then writes those
bytes, modified or not, to its STDOUT
* LFS reads the bytes from STDOUT of lfs-extension.bar.clean, calculates the SHA-256 
signature, and writes the bytes to a temp flie
* When finished, LFS atomically moves the temp file into .git/lfs/objects, as before
* LFS generates the pointer file, with some changes:
 * The oid and size keys are calculated from the final bytes written into the LFS storage
 * LFS also writes keys named extension-1-foo and extension-2-bar into the pointer, along 
 with their respective input oid's

Here's an example pointer file, for a file processed by extensions foo and bar:
  
```
version https://git-lfs.github.com/spec/v1
extension-1-foo sha256:{original hash}
extension-2-bar sha256:{hash after foo}
oid sha256:{hash after bar}
size 123
(ending \n)
```

Note: as an optimization, if an extension just does a pass-through, its key can be 
omitted from the pointer file.  This will make smudging the file a bit more efficient 
since that extension can be skipped.  LFS can detect a pass-through extension because the 
input and output oid's will be the same.


### Smudge

When a file is checked out, Git invokes the LFS smudge filter, as described earlier. If 
no extensions are installed, the LFS smudge filter inspects the first 100 bytes of the 
bytes off STDIN, and if it is a pointer file, uses the oid to find the correct object in 
the LFS storage, and writes those bytes to STDOUT so that Git can write them to the 
working directory.

If the pointer file indicates that extensions were invoked on that file, then those 
extensions must be installed in order to smudge.  If they are not installed, not found, 
or unusable for any reason, LFS will fail to smudge the file, and outputs an error 
indicating which extension is missing.

Each of the extensions indicated in the pointer file must be invoked in reverse order to 
undo the changes they made to the contents of the file.  After each extension is invoked, 
LFS will compare the SHA-256 signature of the bytes output by the extension with the oid 
stored in the pointer file as the original input to that same extension.  Those 
signatures must match, otherwise the extension did not undo its changes correctly.  In 
that case, LFS fails to smudge the file, and outputs an error indicating which extension 
is failing.

Here's an example sequence, indicating how LFS will smudge the pointer file shown in the 
previous section:

* Git passes the bytes of the pointer file to LFS smudge over STDIN.  Note that when 
using "git lfs fetch", LFS reads the files directly from disk rather than off STDIN.  The 
rest of the steps are unaffected either way.
* LFS reads those bytes and inspects them to see if this is a pointer file.  If it was 
not, the bytes would just be passed through to STDOUT.
* Since it is a pointer file, LFS reads the whole file off STDIN, parses it, and 
determines that extensions foo and bar both processed the file, in that order.
* LFS uses the value of the oid key to find the blob in the .git/lfs/objects folder, or 
download from the server as needed
* LFS writes the contents of the blob to STDIN of lfs-extension.bar.smudge, which
modifies them as needed and writes them to its STDOUT
* LFS reads the bytes from STDOUT of lfs-extension.bar.smudge, calculates the SHA-256 
signature, and writes the bytes to STDIN of lfs-extension.foo.smudge, which modifies them 
as needed and writes to them its STDOUT
* LFS reads the bytes from STDOUT of lfs-extension.foo.smudge, calculates the SHA-256 
signature, and writes the bytes to its own STDOUT
* At the end, ensure that the hashes calculated on the outputs of foo and bar match their 
corresponding input hashes from the pointer file.  If not, write a descriptive error 
message indicating which extension failed to undo its changes.
 * Question: On error, should we overwrite the file in the working directory with the 
 original pointer file?  Can this be done reliably?

