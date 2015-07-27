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
stores the result of this in the `lfs.<url>.access` config setting, where <url>
refers to the endpoint's URL.

## API Responses

This specification defines what status codes that API can return.  Look at each
individual API method for more details.  Some of the specific status codes may
trigger specific error messages from the client.

* 200 - The request completed successfully.
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

* If `lfs.<url>.access` is not set, make an unauthenticated request.
  1. If it returns 401, set `lfs.<url>.access` to `private`.
* If `lfs.<url>.access` is `private`, always send authentication. Ask the user if
authentication information is not readily available.

When uploading objects through `git lfs push`, Git LFS will always send
authentication info, regardless of how `lfs.<url>.access` is configured.

```
> POST https://git-lfs-server.com/objects/batch HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "operation": "upload",
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
<       "size": 123,
<       "actions": {
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
multiple actions, each with an `href` property and an optional `header`
property. The requests and responses need to validate with the included JSON
schemas:

* [Batch request](./http-v1-batch-request-schema.json)
* [Batch response](./http-v1-batch-response-schema.json)

Here are the valid actions:

* `upload` - This relation describes how to upload the object.  Expect this with
when the object has not been previously uploaded.
* `verify` - The server can specify a URL for the client to hit after
successfully uploading an object.  This is an optional relation for the case that
the server has not verified the object.
* `download` - This relation describes how to download the object content.  This
only appears if an object has been previously uploaded.

An action can optionally include an `expires_at`, which is an ISO 8601 formatted
timestamp for when the given action expires (usually due to a temporary token).

```json
{
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "actions": {
        "download": {
          "href": "https://some-download.com?token=abc123",
          "expires_at": "2015-07-27T21:15:01Z"
        }
      }
    }
  ]
}
```

### Responses

The Batch API should always return 200 unless there's an authorization problem
between the requesting user and the repository.

* 200 - OK
* 401 - The authentication credentials are needed, but were not sent.
* 403 - The user has **read**, but not **write** access. Only applicable when
the `operation` in the request is "upload."
* 404 - The repository does not exist for the user.

### Object Errors

The server can return specific errors for objects in the batch request. Here's
a sample request and response that includes one missing object:

```
> POST https://git-lfs-server.com/objects/batch HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "operation": "download",
>   "objects": [
>     {
>       "oid": "1111111",
>       "size": 123
>     },
>     {
>       "oid": "2222222",
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
<       "size": 123,
<       "actions": {
<         "download": {
<          "href": "https://some-download.com"
<         }
<       }
>     },
<     {
<       "oid": "2222222",
<       "size": 123,
<       "error": {
<         "code": 404,
<         "message": "Object does not exist on the server"
<       }
>     }
<   ]
< }

The error codes are integers that map to similar HTTP status codes where
possible. Specifics can be included by the server in the `message` property.

Here's a list of defined error codes for objects:

* 404 - Returned when trying to download an object missing from the server.
* 500 - Describes a fatal server error in cases where other HTTP statuses don't
cover it.
