# Authentication

The Git LFS API uses HTTP Basic Authentication to authorize requests. Therefore,
HTTPS is strongly encouraged for all production Git LFS servers. The credentials
can come from the following places:

## SSH

Git LFS will add any HTTP headers returned from the `git-lfs-authenticate`
command to any Batch API requests. If servers are returning expiring tokens,
they can add an `expires_in` (or `expires_at`) property to hint when the token
will expire.

```bash
# Called for remotes like:
#   * git@git-server.com:foo/bar.git
#   * ssh://git@git-server.com/foo/bar.git
$ ssh git@git-server.com git-lfs-authenticate foo/bar.git download
{
  "header": {
    "Authorization": "RemoteAuth some-token"
  },

  # optional, for expiring tokens, preferred over expires_at
  "expires_in": 86400

  # optional, for expiring tokens
  "expires_at": "2016-11-10T15:29:07Z"
}
```

See the SSH section in the [Server Discovery doc](./server-discovery.md) for
more info about `git-lfs-authenticate`.

## Git Credentials

Git provides a [`credentials` command](https://git-scm.com/docs/gitcredentials)
for storing and retrieving credentials through a customizable credential helper.
By default, it associates the credentials with a domain. You can enable
`credential.useHttpPath` so different repository paths have different
credentials.

Git ships with a really basic credential cacher that stores passwords in memory,
so you don't have to enter your password frequently. However, you are encouraged
to setup a [custom git credential cacher](https://help.github.com/articles/caching-your-github-password-in-git/),
if a better one exists for your platform.

If your Git LFS server authenticates with NTLM then you must provide your credentials to `git-credential`
in the form `username:DOMAIN\user password:password`.

## Specified in URL

You can hardcode credentials into your Git remote or LFS url properties in your
git config. This is not recommended for security reasons because it relies on
the credentials living in your local git config.

```bash
$ git remote add origin https://user:password@git-server.com/foo/bar.git
```
