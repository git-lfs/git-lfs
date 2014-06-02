We are thrilled to announce large asset support for GitHub. We've seen explosive
growth in the companies and industries on GitHub. We want GitHub to enable more
people to work better together. Unfortunately, projects that track the history
of large files were not a good fit for Git or GitHub. We do this by directing
large files to GitHub Cloud Storage, instead of Git.

To get started, simply update your interface of choice for GitHub. If you use
GitHub for Mac or GitHub for Windows, large files will automatically upload to
GitHub Cloud Storage.

[ picture ]

Files pushed through subversion that identify as large assets are automatically
transferred to and received from GitHub Cloud Storage.

[ picture ]

Command line Git users need to install our custom git command line tool. This
works as a wrapper around git, adding special GitHub-specific features. You can
download binaries for the Mac, Linux, FreeBSD, and Windows. You need Git above
v1.6? to use it.

Once installed, you can rename it to git, and continue working with your
repository in exactly the same way. There is also a new git media command to get
details on the large assets themselves.

```
$ git media status
blah blah

# this is automatically run after pushing, fetching, and cloning.
$ git media sync
```

We believe in developing open protocols and APIs where possible, so that tools
can work well together. The command line tools and the API specification are
available at this GitHub repository.

Have an A-1 Day!
