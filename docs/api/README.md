# Git LFS API

The Git LFS client talks to an API to upload and download objects. A typical
flow might look like:

1. The user runs a command to download (`git lfs fetch`) or upload (`git lfs
push`) objects.
2. The client contacts the Git LFS API to get information about transferring
the objects.
3. The client then transfers the objects through the storage API.

## HTTP API

The Git LFS HTTP API is responsible for authenticating the user requests, and
returning the proper info for the Git LFS client to use the storage API. By
default, API endpoint is based on the current Git remote. For example:

```
Git remote: https://git-server.com/user/repo.git
Git LFS endpoint: https://git-server.com/user/repo.git/info/lfs

Git remote: git@git-server.com:user/repo.git
Git LFS endpoint: https://git-server.com/user/repo.git/info/lfs
```

The [specification](spec.md) describes how clients can configure the Git LFS
API endpoint manually.

The [original v1 API][v1] is used for Git LFS v0.5.x. An experimental [v1
batch API][batch] is in the works for v0.6.x.

[v1]: ./http-v1-original.md
[batch]: ./http-v1-batch.md

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

### Hypermedia

The Git LFS API uses hypermedia hints to instruct the client what to do next.
These links are included in a `_links` property.  Possible relations for objects
include:

* `self` - This points to the object's canonical API URL.
* `download` - Follow this link with a GET and the optional header values to
download the object content.
* `upload` - Upload the object content to this link with a PUT.
* `verify` - Optional link for the client to POST after an upload.  If
included, the client assumes this step is required after uploading an object.
See the "Verification" section below for more.

Link relations specify the `href`, and optionally a collection of header values
to set for the request.  These are optional, and depend on the backing object
store that the Git LFS API is using.  

The Git LFS client will automatically send the same credentials to the followed
link relation as Basic Authentication IF:

* The url scheme, host, and port all match the Git LFS API endpoint's.
* The link relation does not specify an Authorization header.

If the host name is different, the Git LFS API needs to send enough information
through the href query or header values to authenticate the request.

The Git LFS client expects a 200 or 201 response from these hypermedia requests.
Any other response code is treated as an error.

## Storage API

The Storage API is a generic API for directly uploading and downloading objects.
Git LFS servers can offload object storage to cloud services like S3, or
implemented natively in the Git LFS server. The only requirement is that 
hypermedia objects from the Git LFS API return the correct headers so clients
can access the storage API properly.

The client downloads objects through individual GET requests. The URL and any
special headers are provided by  a "download" hypermedia link:

```
# the hypermedia object from the Git LFS API
# {
#   "_links": {
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
#   "_links": {
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
