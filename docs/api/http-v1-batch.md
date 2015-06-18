# Git LFS v1 Batch API

The Git LFS Batch API works like the [original v1 API][v1], but uses a single
endpoint that accepts multiple OIDs. All requests should have the following:

    Accept: application/vnd.git-lfs+json
    Content-Type: application/vnd.git-lfs+json

[v1]: ./http-v1-original.md

This is an experimental API introduced in Git LFS v0.5.2, and only used if the
`lfs.batch` config value is true. You can toggle support for any local
repository like this:

    # enable batch support
    $ git config lfs.batch true

    # disable batch support
    $ git config --unset lfs.batch

The Batch API is subject to change until it becomes the _default_ API used by
Git LFS v0.6.0.

## Authentication

The Batch API authenticates the same as the original v1 API with one exception:
The client will attempt to make requests without any authentication. This
slight change allows anonymous access to public Git LFS objects. The client
stores the result of this in the `lfs.access` config setting.

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
* 401 - The authentication credentials are needed, but were not sent.
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

## POST /objects/batch

This request retrieves the metadata for a batch of objects, given a JSON body
containing an object with an array of objects with the oid and size of each
object. While the API endpoint can support requests to download AND upload
objects in one batch, the client will usually stick to one or the other.

When downloading objects through a command such as `git lfs fetch`, the client
will initially skip authentication if it doesn't know the access level of the
repository.

* If `lfs.access` is not set, make an unauthenticated request.
  1. If it returns 200, set `lfs.access` to `public`.
  2. If it returns 401, set `lfs.access` to `private`.
* If `lfs.access` is `public`, don't ask the user for authentication. If
authentication is available in `git credential`, or through `ssh`, then use it.
* If `lfs.access` is `private`, always send authentication. Ask the user if
authentication information is not readily available.

When uploading objects through `git lfs push`, Git LFS will always send
authentication info, regardless of how `lfs.access` is configured.

```
> POST https://git-lfs-server.com/objects/batch HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "objects": [
>     {
>       "oid": "1111111",
>       "size": 123
>     }
>   ]
> }
>
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "objects": [
<     {
<       "oid": "1111111",
<       "_links": {
<         "upload": {
<          "href": "https://some-upload.com",
<           "header": {
<             "Key": "value"
<           }
<         },
<         "verify": {
<           "href": "https://some-callback.com",
<           "header": {
<             "Key": "value"
<           }
<         }
<       }
>     }
<   ]
< }
```

The response will be an object containing an array of objects with one of
multiple link relations, each with an `href` property and an optional `header`
property.

* `upload` - This relation describes how to upload the object.  Expect this with
when the object has not been previously uploaded.
* `verify` - The server can specify a URL for the client to hit after
successfully uploading an object.  This is an optional relation for the case that
the server has not verified the object.
* `download` - This relation describes how to download the object content.  This
only appears if an object has been previously uploaded.

### Responses

* 200 - OK
* 401 - The authentication credentials are needed, but were not sent.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.
