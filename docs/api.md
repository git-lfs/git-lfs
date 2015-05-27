# Git LFS API

The server implements a simple API for uploading and downloading binary content.
Git repositories that use Git LFS will specify a URI endpoint.  See the
[specification](spec.md) for how Git LFS determines the server endpoint to use.

Use that endpoint as a base, and append the following relative paths to upload
and download from the Git LFS server.

API requests require an Accept header of `application/vnd.git-lfs+json`. The
upload and verify requests need a `application/vnd.git-lfs+json` Content-Type
too.

## Authentication

The Git LFS API uses [git credentials](http://git-scm.com/docs/gitcredentials) to
access the API username and password.  If the Git LFS domain matches the Git
remote's, Git LFS will not prompt you again to enter the password.

If the Git remote is using SSH, Git LFS will execute the `git-lfs-authenticate`
command.  It passes the SSH path, the Git LFS operation (upload or download),
and the object OID as arguments. A successful result outputs a JSON link object
to STDOUT.  This is applied to any Git LFS API request before git credentials
are accessed.

```
# remote: git@github.com:user/repo.git
$ ssh git@github.com git-lfs-authenticate user/repo.git download {oid}
{
  "header": {
    "Authorization": "Basic ..."
  }
  // OPTIONAL key only needed if the Git LFS server is not hosted at the default
  // URL from the Git remote:
  //   https://github.com/user/repo.git/info/lfs
  "href": "https://other-server.com/user/repo",
}
```

If Git LFS detects a non-zero exit status, it displays the command's STDERR:

```
$ ssh git@github.com git-lfs-authenticate user/repo.git wat {oid}
Invalid LFS operation: "wat"
```

NOTE: The Git LFS client currently does not cache the headers from the SSH
command.  This is being considered for a future release.

## API Responses

This specification defines what status codes that API can return.  Look at each
individual API method for more details.  Some of the specific status codes may
trigger specific error messages from the client.

* 200 - The request completed successfully.
* 202 - An upload request has been accepted.  Clients must follow hypermedia
links to actually upload the content.
* 301 - A permanent redirect.  Only supported for GET/HEAD requests.
* 302 - A temporary redirect.  Only supported for GET/HEAD requests.
* 303 - A temporary redirect.  Only supported for GET/HEAD requests.
* 307 - A temporary redirect.  Keeps the original request method intact.
* 400 - General error with the client's request.  Invalid JSON formatting, for
example.
* 401 - The authentication credentials are incorrect.
* 403 - The requesting user has access to see the repository, but not to push
changes to it.
* 404 - Either the user does not have access to see the repository, or the
repository or requested object does not exist.

The following status codes can optionally be returned from the API, depending on
the server implementation.

* 406 - The Accept header needs to be `application/vnd.git-lfs+json`.
* 410 - The requested object used to exist, but was deleted.  The message should
state why (user initiated, legal issues, etc).
* 429 - The user has hit a rate limit with the server.  Though the API does not
specify any rate limits, implementors are encouraged to set some for
availability reasons.
* 501 - The server has not implemented the current method.  Reserved for future
use.
* 509 - Returned if the bandwidth limit for the user or repository has been
exceeded.  The API does not specify any bandwidth limit, but implementors may
track usage.

Some server errors may trigger the client to retry requests, such as 500, 502,
503, and 504.

If the server returns a JSON error object, the client can display this message
to users.

```
> GET https://git-lfs-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.git-lfs+json
>
< HTTP/1.1 200 OK
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "message": "Bad credentials",
<   "documentation_url": "https://git-lfs-server.com/docs/errors",
<   "request_id": "123"
< }
```

The `documentation_url` and `request_id` properties are optional.  If given,
they are displayed to the user.

## Redirections

The Git LFS client follows redirections on the core Git LFS API methods only.
Any of the hypermedia hrefs that are returned must be for the current location.
The client will pass all of the original request headers to the redirected
request, only changing the URL based on the redirect location.  The only
exception is the Authorization header, which is only passed through if the
original request and the new location have a matching URL scheme, host, and
port.

The client will automatically follow redirections for GET or HEAD requests on
a 301, 302, 303, or 307 HTTP status.  It only automatically follows redirections
for other HTTP verbs on a 307 HTTP status.

Note: the 308 HTTP status is not official, and has conflicting proposals for its
intended use.  It is not supported as a redirection.

## Hypermedia

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

## GET /objects/{oid}

This gets the object's meta data.  The OID is the value from the object pointer.

```
> GET https://git-lfs-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "oid": "the-sha-256-signature",
<   "size": 123456,
<   "_links": {
<     "self": {
<       "href": "https://git-lfs-server.com/objects/OID",
<     },
<     "download": {
<       "href": "https://some-download.com",
<       "header": {
<         "Key": "value"
<       }
<     }
<   }
< }
```

The `oid` and `size` properties are required.  A hypermedia `_links` section is
included with a `download` link relation.  Clients can follow this link to
access the object content. See the "Hypermedia" section above for more.

Here's a sample response for a request with an authorization error:

```
> GET https://git-lfs-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 404 Not found
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "message": "Not found"
< }
```

### Responses

* 200 - The object exists and the user has access to download it.
* 401 - The authentication credentials are incorrect.
* 404 - The user does not have access to the object, or it does not exist.
* 410 - The object used to exist, but was deleted. The message should state why
(user initiated, legal issues, etc).

## POST /objects

This request initiates the upload of an object, given a JSON body with the oid
and size of the object to upload.

```
> POST https://git-lfs-server.com/objects HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "oid": "1111111",
>   "size": 123
> }
>
< HTTP/1.1 202 Accepted
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "_links": {
<     "upload": {
<       "href": "https://some-upload.com",
<       "header": {
<         "Key": "value"
<       }
<     },
<     "verify": {
<       "href": "https://some-callback.com",
<       "header": {
<         "Key": "value"
<       }
<     }
<   }
< }
```

A response can include one of multiple link relations, each with an `href`
property and an optional `header` property.

* `upload` - This relation describes how to upload the object.  Expect this with
a 202 status.
* `verify` - The server can specify a URL for the client to hit after
successfully uploading an object.  This is an optional relation for a 202
status.
* `download` - This relation describes how to download the object content.  This
only appears on a 200 status.

### Responses

* 200 - The object already exists.  Don't bother re-uploading.
* 202 - The object is ready to be uploaded.  Follow the "upload" and optional
"verify" links.
* 401 - The authentication credentials are incorrect.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.

## Verification

When Git LFS clients issue a POST request to initiate an object upload, the
response can potentially return a "verify" link relation.  If given, The Git LFS
API expects a POST to the href after a successful upload.  Git LFS clients send:

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
