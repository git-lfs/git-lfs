# Git LFS Roadmap

This is a high level overview of some of the big changes we want to make for
Git LFS. Nothing here is final. Anything can be added or moved around at any
time. Also, the items will be annotated with issue references to show the
current state of the feature.

## v1.0

These are the features that we feel are important for a v1 release of Git LFS.

* Resumable, chunked downloads
* Resumable, chunked uploads
* Concurrent uploads. Though chunking may take care of this.
* New command for replacing pointers with large files outside of the Git smudge
and clean filters (`git lfs get path/to/file`)
* Automatic GC for the `.git/lfs/objects` directory
* Client side metrics reporting, so the Git LFS server can optionally track
how clients are performing.

## Possible Features

These are features that require some more research. It's very possible that
these can make it in for v1.0 if there's a great proposal.

* Narrow clones - Allow clients to specify which large files to download
automatically.
* File locking
* Binary diffing - reduce the amount of content sent over the wire.
* Concurrent downloads - Difficult to implement due to how git smudge is used.

## Project Related

These are items that don't affect Git LFS end users.

* Cross platform integration tests in shell
* Build and CI servers for Linux, Windows, and Mac.
