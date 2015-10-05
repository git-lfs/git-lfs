# Extending LFS

Teams who use Git LFS often have custom requirements for how the pointer files
and blobs should be handled.  Some examples of extensions that could be built:

* Compress large files on clean, uncompress them on smudge/fetch
* Encrypt files on clean, decrypt on smudge/fetch
* Scan files on clean to make sure they don't contain sensitive information

The basic extensibility model is that LFS extensions must be registered
explicitly, and they will be invoked on clean and smudge to manipulate the
contents of the files as needed.  On clean, LFS itself ensures that the pointer
file is updated with all the information needed to be able to smudge correctly,
and the extensions never modify the pointer file directly.

NOTE: This feature is considered experimental, and included so developers can
work on extensions. Exact details of how extensions work are subject to change
based on feedback. It is possible for buggy extensions to leave your repository
in a bad state, so don't rely on them with a production git repository without
extensive testing.

## Registration

To register an LFS extension, it must be added to the Git config.  Each
extension needs to define:

* Its unique name.  This will be used as part of the key in the pointer file.
* The command to run on clean (when files are added to git).
* The command to run on smudge (when files are downloaded and checked out).
* The priority of the extension, which must be a unique, non-negative integer.

The sequence `%f` in the clean and smudge commands will be replaced by the
filename being processed.

Here's an example extension registration in the Git config:

```
[lfs "extension.foo"]
  clean = foo clean %f
  smudge = foo smudge %f
  priority = 0
[lfs "extension.bar"]
  clean = bar clean %f
  smudge = bar smudge %f
  priority = 1
```

## Clean

When staging a file, Git invokes the LFS clean filter, as described earlier.  If
no extensions are installed, the LFS clean filter reads bytes from STDIN,
calculates the SHA-256 signature, and writes the bytes to a temp file.  It then
moves the temp file into the appropriate place in .git/lfs/objects and writes a
valid pointer file to STDOUT.

When an extension is installed, LFS will invoke the extension to do additional
processing on the bytes before writing them into the temp file.  If multiple
extensions are installed, they are invoked in the order defined by their
priority.  LFS will also insert a key in the pointer file for each extension
that was invoked, indicating both the order that the extension was invoked and
the oid of the file before that extension was invoked. All of that information
is required to be able to reliably smudge the file later.  Each new line in the
pointer file will be of the form:

`ext-{order}-{name} {hash-method}:{hash-of-input-to-extension}`

This naming ensures that all extensions are written in both alphabetical and
priority order, and also shows the progression of changes to the oid as it is
processed by the extensions.

Here's an example sequence, assuming extensions foo and bar are installed, as
shown in the previous section.

* Git passes the original contents of the file to LFS clean over STDIN.
* LFS reads those bytes and calculates the original SHA-256 signature.
* LFS streams the bytes to STDIN of `foo clean`, which is expected to write
  those bytes, modified or not, to its STDOUT.
* LFS reads the bytes from STDOUT of `foo clean`, calculates the SHA-256
  signature, and writes them to STDIN of `bar clean`, which then writes those
  bytes, modified or not, to its STDOUT.
* LFS reads the bytes from STDOUT of `bar clean`, calculates the SHA-256
  signature, and writes the bytes to a temp file.
* When finished, LFS atomically moves the temp file into `.git/lfs/objects`.
* LFS generates the pointer file, with some changes:
* The oid and size keys are calculated from the final bytes written to LFS
  local storage.
* LFS also writes keys named `ext-0-foo` and `ext-1-bar` into the pointer, along
  with their respective input oids.

Here's an example pointer file, for a file processed by extensions foo and bar:

```
version https://git-lfs.github.com/spec/v1
ext-0-foo sha256:{original hash}
ext-1-bar sha256:{hash after foo}
oid sha256:{hash after bar}
size 123
(ending \n)
```

Note: as an optimization, if an extension just does a pass-through, its key can
be omitted from the pointer file.  This will make smudging the file a bit more
efficient since that extension can be skipped.  LFS can detect a pass-through
extension because the input and output oids will be the same.

This implies that extensions must have no side effects other than writing to
their STDOUT. Otherwise LFS has no way to know what extensions modified a file.

## Smudge

When a file is checked out, Git invokes the LFS smudge filter, as described
earlier. If no extensions are installed, the LFS smudge filter inspects the
first 100 bytes of the bytes off STDIN, and if it is a pointer file, uses the
oid to find the correct object in the LFS storage, and writes those bytes to
STDOUT so that Git can write them to the working directory.

If the pointer file indicates that extensions were invoked on that file, then
those extensions must be installed in order to smudge.  If they are not
installed, not found, or unusable for any reason, LFS will fail to smudge the
file, and outputs an error indicating which extension is missing.

Each of the extensions indicated in the pointer file must be invoked in reverse
order to undo the changes they made to the contents of the file.  After each
extension is invoked, LFS will compare the SHA-256 signature of the bytes output
by the extension with the oid stored in the pointer file as the original input
to that same extension.  Those signatures must match, otherwise the extension
did not undo its changes correctly.  In that case, LFS fails to smudge the file,
and outputs an error indicating which extension is failing.

Here's an example sequence, indicating how LFS will smudge the pointer file
shown in the previous section:

* Git passes the bytes of the pointer file to LFS smudge over STDIN.  Note that
  when using `git lfs checkout`, LFS reads the files directly from disk rather
  than off STDIN.  The rest of the steps are unaffected either way.
* LFS reads those bytes and inspects them to see if this is a pointer file.  If
  it was not, the bytes would just be passed through to STDOUT.
* Since it is a pointer file, LFS reads the whole file off STDIN, parses it, and
  determines that extensions foo and bar both processed the file, in that order.
* LFS uses the value of the oid key to find the blob in the `.git/lfs/objects`
  folder, or download from the server as needed.
* LFS writes the contents of the blob to STDIN of `bar smudge`, which modifies
  them as needed and writes them to its STDOUT.
* LFS reads the bytes from STDOUT of `bar smudge`, calculates the SHA-256
  signature, and writes the bytes to STDIN of `foo smudge`, which modifies them
  as needed and writes to them its STDOUT.
* LFS reads the bytes from STDOUT of `foo smudge`, calculates the SHA-256
  signature, and writes the bytes to its own STDOUT.
* At the end, ensure that the hashes calculated on the outputs of foo and bar
  match their corresponding input hashes from the pointer file.  If not, write a
  descriptive error message indicating which extension failed to undo its
  changes.
* Question: On error, should we overwrite the file in the working directory with
  the original pointer file?  Can this be done reliably?

## Handling errors

If there are errors in the configuration of LFS extensions, such as invalid
extension names, duplicate priorities, etc, then any LFS commands that rely on
them will abort with a descriptive error message.

If an extension is unable to perform its task, it can indicate this error by
returning a non-zero error code and writing a descriptive error message to its
STDERR. The behavior on an error depends on whether we are cleaning or smudging.

### Clean

If an extension fails to clean a file, it will return a non-zero error code and
write an error message to its STDERR.  Because the file was not cleaned
correctly, it can't be added to the index.  LFS will ensure that no pointer file
is added or updated for failed files.  In addition, it will display the error
messages for any files that could not be cleaned (and keep those errors in a
log), so that the user can diagnose the failure, and then rerun "git add" on
those files.

### Smudge

If an extension fails to smudge a file, it will return a non-zero error code and
write an error message to its STDERR.  Because the file was not smudged
correctly, LFS cannot update that file in the working directory.  LFS will
ensure that the pointer file is written to both the index and working directory.
In addition, it will display the error messages for any files that could not be
smudged (and keep those errors in a log), so that the user can diagnose the
failure and then rerun `git-lfs checkout` to fix up any remaining pointer files.
