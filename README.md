Git Hawser
======

Git command line extension for managing large files.  Hawser replaces large
files with text pointers inside Git, while storing the actual files in a remote
Hawser server.

The Git Hawser client is written in Go, with pre-compiled binaries available for
Mac, Windows, Linux, and FreeBSD.

See [CONTRIBUTING.md](CONTRIBUTING.md) for info on working on Hawser and sending
patches.

## Getting Started

Download the [latest client][rel] and run the included install script.  The
installer should run `git hawser init` for you, which sets up Git's global
configuration settings for Hawser.

[rel]: https://github.com/hawser/git-hawser/releases

### Configuration

Hawser uses `.gitattributes` files to configure which are managed by Hawser.
Here is a smaple one that saves zips and mp3s:

    $ cat .gitattributes
    *.mp3 filter=hawser -crlf
    *.zip filter=hawser -crlf

Hawser can help manage the paths:

    $ git hawser add "*.mp3"
    Adding path *.mp3

    $ git hawser add "*.zip"
    Adding path *.zip

    $ git hawser path
    Listing paths
        *.mp3 (.gitattributes)
        *.zip (.gitattributes)

    $ git hawser remove "*.zip"
    Removing path *.zip

    $ git hawser path
    Listing paths
        *.mp3 (.gitattributes)

### Pushing commits

Once setup, you're ready to push some commits.

    $ git add my.zip
    $ git commit -m "add zip"

You can confirm that Hawser is managing your zip file:

    $ git hawser ls-files
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
    To https://github.com/hawser/hawser-test
       67fcf6a..47b2002  master -> master

See the [Hawser overview](https://github.com/hawser/git-hawser/tree/master/docs) and [man pages](https://github.com/hawser/git-hawser/tree/master/docs/man).
