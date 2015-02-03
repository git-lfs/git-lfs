# Hawser

Hawser is a system for handling large blobs in Git repositories.  Instead of
saving the full content of a blob to Git, a pointer file with an ID is written.
A Hawser client can use the ID in the pointer file to download the actual
file from a Hawser server.

* The Client
  * [Specification](spec.md)
  * [Commands](man)
* The Server
  * [API](api.md)

## Getting Started

Download the [latest client](https://github.com/hawser/git-hawser/releases) and run the
included `install.sh` script.

### Configure Git

Run `git hawser init` to set up global Git configuration settings for Hawser.

    # automatically run by the install script.
    $ git hawser init
    git hawser initialized

Git repositories use `.gitattributes` files to configure which files are inserted into
the Hawser server.  Here's a sample one that saves zips and mp3s:

    $ cat .gitattributes
    *.mp3 filter=hawser -crlf
    *.zip filter=hawser -crlf

Hawser can help you manage your `.gitattributes` paths:

    $ git hawser path add "*.mp3"
    Adding path *.mp3

    $ git hawser path add "*.zip"
    Adding path *.zip

    $ git hawser path
    Listing paths
        *.mp3 (.gitattributes)
        *.zip (.gitattributes)

    $ git hawser path remove "*.zip"
    Removing path *.zip

    $ git hawser path
    Listing paths
        *.mp3 (.gitattributes)

### Push a new commit

Once everything is setup, you can clone or create a new repository that uses
Hawser.

```
$ git clone https://github.com/github/gitmediatest
Cloning into 'gitmediatest'...
remote: Counting objects: 22, done.
remote: Compressing objects: 100% (18/18), done.
remote: Total 22 (delta 2), reused 22 (delta 2)
Unpacking objects: 100% (22/22), done.
```

There will be a pause after the objects have been unpacked, while Hawser
downloads the files.  You can tell it worked if the file contains the actual
content, and not a tiny external pointer file:

```
$ cd gitmediatest
$ ls -al
total 24600
drwxr-xr-x   3 rick  staff       204 Oct 31 13:40 .
drwxr-xr-x  70 rick  staff      2414 Oct 31 13:40 ..
drwxr-xr-x   8 rick  staff       442 Oct 31 13:40 .git
-rw-r--r--   1 rick  staff        50 Oct 31 13:40 .gitattributes
-rw-r--r--   1 rick  staff         4 Oct 31 13:40 .gitignore
-rw-r--r--   1 rick  staff  12585968 Oct 31 13:40 mac.zip
```

Now, add a file:

```
$ git add my.zip
$ git commit -m "Add my zip file."

# confirm the zip was added to hawser
$ git hawser ls-files
my.zip
```

When you can see files being added to hawser, you can commit like
normal.  After committing, `git show` will show the file's metadata:

    $ git show
    commit 47b2002173ae56f6a30c67ec46858a932e8f7511
    Author: rick <technoweenie@gmail.com>
    Date:   Thu Oct 31 12:05:32 2013 -0600

        add zip

    diff --git a/my.zip b/my.zip
    new file mode 100644
    index 0000000..fc1f642
    --- /dev/null
    +++ b/my.zip
    @@ -0,0 +1,2 @@
    +# git-media
    +84ff327f80500d3266bd830891ede1e4fd18b9169936a066573f9b230597a696
    \ No newline at end of file

Now, when you run `git push`, added media files will be synced to the
Hawser endpoint.

    $ git push origin master
    Sending my.zip
    12.58 MB / 12.58 MB  100.00 %
    Counting objects: 2, done.
    Delta compression using up to 8 threads.
    Compressing objects: 100% (5/5), done.
    Writing objects: 100% (5/5), 548 bytes | 0 bytes/s, done.
    Total 5 (delta 1), reused 0 (delta 0)
    To https://github.com/github/gitmediatest
       67fcf6a..47b2002  master -> master
