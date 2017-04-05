# Server Discovery

One of the Git LFS goals is to work with supporting Git remotes with as few
required configuration properties as possible. Git LFS will attempt to use
your Git remote to determine the LFS server. You can also configure a custom
LFS server if your Git remote doesn't support one, or you just want to use a
separate one.

Look for the `Endpoint` properties in `git lfs env` to see your current LFS
servers.

## Guessing the Server

By default, Git LFS will append `.git/info/lfs` to the end of a Git remote url
to build the LFS server URL it will use:

Git Remote: `https://git-server.com/foo/bar`<br>
LFS Server: `https://git-server.com/foo/bar.git/info/lfs`

Git Remote: `https://git-server.com/foo/bar.git`<br>
LFS Server: `https://git-server.com/foo/bar.git/info/lfs`

Git Remote: `git@git-server.com:foo/bar.git`<br>
LFS Server: `https://git-server.com/foo/bar.git/info/lfs`

Git Remote: `ssh://git-server.com/foo/bar.git`<br>
LFS Server: `https://git-server.com/foo/bar.git/info/lfs`

## SSH

If Git LFS detects an SSH remote, it will run the `git-lfs-authenticate`
command. This allows supporting Git servers to give the Git LFS client
alternative authentication so the user does not have to setup a git credential
helper.

Git LFS runs the following command:

    $ ssh [{user}@]{server} git-lfs-authenticate {path} {operation}

The `user`, `server`, and `path` properties are taken from the SSH remote. The
`operation` can either be "download" or "upload". The SSH command can be
tweaked with the `GIT_SSH` or `GIT_SSH_COMMAND` environment variables. The
output for successful commands is JSON, and matches the schema as an `action`
in a Batch API response. Git LFS will dump the STDERR from the `ssh` command if
it returns a non-zero exit code.

Examples:

The `git-lfs-authenticate` command can even suggest an LFS endpoint that does
not match the Git remote by specifying an `href` property.

```bash
# Called for remotes like:
#   * git@git-server.com:foo/bar.git
#   * ssh://git@git-server.com/foo/bar.git
$ ssh git@git-server.com git-lfs-authenticate foo/bar.git download
{
  "href": "https://lfs-server.com/foo/bar",
  "header": {
    "Authorization": "RemoteAuth some-token"
  },
  "expires_in": 86400
}
```

Git LFS will output the STDERR if `git-lfs-authenticate` returns a non-zero
exit code:

```bash
$ ssh git@git-server.com git-lfs-authenticate foo/bar.git wat
Invalid LFS operation: "wat"
```

## Custom Configuration

If Git LFS can't guess your LFS server, or you aren't using the
`git-lfs-authenticate` command, you can specify the LFS server using Git config.

Set `lfs.url` to set the LFS server, regardless of Git remote.

```bash
$ git config lfs.url https://lfs-server.com/foo/bar
```

You can set `remote.{name}.lfsurl` to set the LFS server for that specific
remote only:

```bash
$ git config remote.dev.lfsurl http://lfs-server.dev/foo/bar
$ git lfs env
...

Endpoint=https://git-server.com/foo/bar.git/info/lfs (auth=none)
Endpoint (dev)=http://lfs-server.dev/foo/bar (auth=none)
```

Git LFS will also read these settings from a `.lfsconfig` file in the root of
your repository. This lets you commit it to the repository so that all users
can use it, if you wish.

```bash
$ git config --file=.lfsconfig lfs.url https://lfs-server.com/foo/bar
```
