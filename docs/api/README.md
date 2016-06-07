# Git LFS API

The Git LFS client talks to an API to upload and download objects. A typical
flow might look like:

1. The user runs a command to download (`git lfs fetch`) or upload (`git lfs
push`) objects.
2. The client contacts the Git LFS API to get information about transferring
the objects.
3. The client then transfers the objects through the transfer API.

## HTTP API

The Git LFS HTTP API is responsible for authenticating the user requests, and
returning the proper info for the Git LFS client to use the transfer API. By
default, API endpoint is based on the current Git remote. For example:

```
Git remote: https://git-server.com/user/repo.git
Git LFS endpoint: https://git-server.com/user/repo.git/info/lfs

Git remote: git@git-server.com:user/repo.git
Git LFS endpoint: https://git-server.com/user/repo.git/info/lfs
```

The [specification](/docs/spec.md) describes how clients can configure the Git LFS
API endpoint manually.

The [legacy v1 API][legacy] was used for Git LFS v0.5.x. From 0.6.x the 
[batch API][batch] should always be used where available. 

[legacy]: ./v1/http-v1-legacy.md
[batch]: ./v1/http-v1-batch.md

From v1.3 there are [optional extensions to the batch API][batch v1.3] for more 
flexible transfers.

[batch v1.3]: ./v1.3/http-v1.3-batch.md


### Authentication

The Git LFS API uses HTTP Basic Authentication to authorize requests. The
credentials can come from the following places:

1. Specified in the URL: `https://user:password@git-server.com/user/repo.git/info/lfs`.
This is not recommended for security reasons because it relies on the
credentials living in your local git config.
2. `git-credential` will either retrieve the stored credentials for your Git
host, or ask you to provide them. Successful requests will store the credentials
for later if you have a [good git credential cacher](https://help.github.com/articles/caching-your-github-password-in-git/).
3. SSH

If the Git remote is using SSH, Git LFS will execute the `git-lfs-authenticate`
command.  It passes the SSH path and the Git LFS operation (upload or download),
as arguments. A successful result outputs a JSON link object to STDOUT.  This is
applied to any Git LFS API request before git credentials are accessed.

NOTE: Git LFS v0.5.x clients using the [original v1 HTTP API][v1] also send the
OID as the 3rd argument to `git-lfs-authenticate`. SSH servers that want to
support both the original and the [batch][batch] APIs should ignore the OID
argument.

```
# remote: git@git-server.com:user/repo.git
$ ssh git@git-server.com git-lfs-authenticate user/repo.git download
{
  "header": {
    "Authorization": "Basic ..."
  }
  // OPTIONAL key only needed if the Git LFS server is not hosted at the default
  // URL from the Git remote:
  //   https://git-server.com/user/repo.git/info/lfs
  "href": "https://other-server.com/user/repo",
}
```

If Git LFS detects a non-zero exit status, it displays the command's STDERR:

```
$ ssh git@git-server.com git-lfs-authenticate user/repo.git wat
Invalid LFS operation: "wat"
```

HTTPS is strongly encouraged for all production Git LFS servers.

If your Git LFS server authenticates with NTLM then you must provide your credentials to `git-credential`
in the form `username:DOMAIN\user password:password`.

## Transfer API

The transfer API is a generic API for directly uploading and downloading objects.
Git LFS servers can offload object storage to cloud services like S3, or
implemented natively in the Git LFS server. The only requirement is that
hypermedia objects from the Git LFS API return the correct headers so clients
can access the transfer API properly.

As of v1.3 there can be multiple ways files can be uploaded or downloaded, see
the [v1.3 API doc](v1.3/http-v1.3-batch.md) for details. The following section
describes the basic transfer method which is the default.

### The basic transfer API

The client downloads objects through individual GET requests. The URL and any
special headers are provided by  a "download" hypermedia link:

```
# the hypermedia object from the Git LFS API
# {
#   "actions": {
#     "download": {
#       "href": "https://storage-server.com/OID",
#       "header": {
#         "Authorization": "Basic ...",
#       }
#     }
#   }
# }

# This request expects the raw object contents in the response.
> GET https://storage-server.com/OID
> Authorization: Basic ...
<
< HTTP/1.1 200 OK
< Content-Type: application/octet-stream
< Content-Length: 123
<
< {contents}
```

The client uploads objects through individual PUT requests. The URL and headers
are provided by an "upload" hypermedia link:

```
# the hypermedia object from the Git LFS API
# {
#   "actions": {
#     "upload": {
#       "href": "https://storage-server.com/OID",
#       "header": {
#         "Authorization": "Basic ...",
#       }
#     }
#   }
# }

# This sends the raw object contents as the request body.
> PUT https://storage-server.com/OID
> Authorization: Basic ...
> Content-Type: application/octet-stream
> Content-Length: 123
>
> {contents}
>
< HTTP/1.1 200 OK
```

## Verification

The Git LFS API can optionally return a "verify" hypermedia link in addition to
an "upload" link. If given, The Git LFS API expects a POST to the href after a
successful upload.  Git LFS clients send:

* `oid` - The String OID of the Git LFS object.
* `size` - The integer size of the Git LFS object, in bytes.

```
> POST https://git-lfs-server.com/callback
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Content-Length: 123
>
> {"oid": "{oid}", "size": 10000}
>
< HTTP/1.1 200 OK
```

A 200 response means that the object exists on the server.
