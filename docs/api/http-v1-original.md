# Git LFS v1 Original API

This describes the original API for Git LFS v0.5.x. It's already deprecated by
the [batch API][batch].  All requests should have:

    Accept: application/vnd.git-lfs+json
    Content-Type: application/vnd.git-lfs+json

[batch]: ./http-v1-batch.md

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
