# Git Large File Storage

Git LFS is a command line extension and [specification](docs/spec.md) for
managing large files with Git. The client is written in Go, with pre-compiled
binaries available for Mac, Windows, Linux, and FreeBSD.

**Support for Git LFS on GitHub.com is coming soon!**

## Features

By design, every git repository contains every version of every file. But
for some types of projects, this is not reasonable or even practical.
Multiple revisions of a large file take up space quickly, slowing down
repository operations and making fetches unwieldy.

Git LFS overcomes this limitation by storing the metadata for large files in
Git and syncing the file contents to a configurable [Git LFS
server](docs/api.md). Some of the key features include:

* Tight integration with Git means you don't have to change your workflow after
the initial configuration.

* Large files are synced separately to a configurable Git LFS server over HTTPS,
so you are not limited in where you push your Git repository.

* Large files are only synced from the server when they are checked out, so your
local repository doesn't carry the weight of every version of every file when it
is not needed.

* The meta data stored in Git is extensible for future use. It currently
includes a hash of the contents of the file, and the file size so clients can
display a progress bar while downloading or opt out of a large download.

* Clients and servers can make use of all the features of HTTPS, such as caching
content locally on a CDN, resumable uploads and downloads, or performing
requests in parallel for faster transfers.

## Getting Started

Download the [latest client][rel] and run the included install script.  The
installer should run `git lfs init` for you, which sets up Git's global
configuration settings for Git LFS.

[rel]: https://github.com/github/git-lfs/releases

### Configuration

Git LFS uses `.gitattributes` files to configure which are managed by Git LFS.
Here is a sample one that saves zips and mp3s:

    $ cat .gitattributes
    *.mp3 filter=lfs -crlf
    *.zip filter=lfs -crlf

Git LFS can manage `.gitattributes` for you:

    $ git lfs track "*.mp3"
    Tracking *.mp3

    $ git lfs track "*.zip"
    Tracking *.zip

    $ git lfs track
    Listing tracked paths
        *.mp3 (.gitattributes)
        *.zip (.gitattributes)

    $ git lfs untrack "*.zip"
    Untracking *.zip

    $ git lfs track
    Listing tracked paths
        *.mp3 (.gitattributes)

### Pushing commits

Once setup, you're ready to push some commits.

    $ git add my.zip
    $ git commit -m "add zip"

You can confirm that Git LFS is managing your zip file:

    $ git lfs ls-files
    my.zip

Once you've made your commits, push your files to the Git remote.

    $ git push origin master
    Sending my.zip
    12.58 MB / 12.58 MB  100.00 %
    Counting objects: 2, done.
    Delta compression using up to 8 threads.
    Compressing objects: 100% (5/5), done.
    Writing objects: 100% (5/5), 548 bytes | 0 bytes/s, done.
    Total 5 (delta 1), reused 0 (delta 0)
    To https://github.com/github/git-lfs-test
       67fcf6a..47b2002  master -> master

See the [Git LFS overview](https://github.com/github/git-lfs/tree/master/docs)
and [man pages](https://github.com/github/git-lfs/tree/master/docs/man).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for info on working on Git LFS and
sending patches.
