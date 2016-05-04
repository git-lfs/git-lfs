# Git LFS v1 Batch API

The Git LFS Batch API works like the [original v1 API][v1], but uses a single
endpoint that accepts multiple OIDs. All requests should have the following:

    Accept: application/vnd.git-lfs+json
    Content-Type: application/vnd.git-lfs+json

[v1]: ./http-v1-original.md

This is a newer API introduced in Git LFS v0.5.2, and made the default in
Git LFS v0.6.0. The client automatically detects if the server does not
implement the API yet, and falls back to the legacy API. You can toggle support
manually through the Git config:

    # enable batch support
    $ git config --unset lfs.batch

    # disable batch support
    $ git config lfs.batch false

## Authentication

The Batch API authenticates the same as the original v1 API with one exception:
The client will attempt to make requests without any authentication. This
slight change allows anonymous access to public Git LFS objects. The client
stores the result of this in the `lfs.<url>.access` config setting, where <url>
refers to the endpoint's URL.

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
<           "href": "https://some-upload.com",
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

* `upload` - This relation describes how to upload the object.  If the object
has not previously been uploaded the server should provide this action.  If
the object has been previously uploaded and the object content is known to the
server, it should not provide this action.  When the action is not provided,
the client should assume the server already knows the object content and skip
uploading it.
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

### Successful Responses

The Batch API should always return 200 unless there's an authorization problem
between the requesting user and the repository.

Here is a response to download a single object:

```
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
<           "href": "https://some-download.com"
<         }
<       }
>     }
<   ]
< }
```

It is possible for servers to respond with a 200, and just annotate specific
objects that failed through an `error` property. Here's an example request to
download two objects, with one object that doesn't exist:

```
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
<           "href": "https://some-download.com"
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
```

Object error codes should match HTTP status codes where possible:

* 404 - The object does not exist on the server.
* 410 - The object was removed by the owner.
* 422 - Validation error.

Validation errors can only occur on `upload` requests. Servers must verify
that OIDs are valid SHA-256 strings, and that sizes are positive integers.
Servers may also set an upper bound for the allowed object size too. Here's a
response showing one uploadable object, and one with a validation error:

```
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
<           "href": "https://some-upload.com"
<         }
<       }
>     },
<     {
<       "oid": "2222222",
<       "size": -1,
<       "error": {
<         "code": 422,
<         "message": "Invalid object size"
<       }
>     }
<   ]
< }
```

### Response Errors

Servers can respond with the following HTTP status codes:

* 401 - The authentication credentials are needed, but were not sent.
* 403 - The user has **read**, but not **write** access. Only applicable when
the `operation` in the request is "upload."
* 404 - The repository does not exist for the user.
* 422 - Validation error with one or more of the objects in the request. This
  means that _none_ of the requested objects to upload are valid.

Responses will not have the `objects` property. They must have a `message`
property, and should have `request_id` and `documentation_url` properties to
help users.

```
< HTTP/1.1 403 Forbidden
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "message": "Invalid credentials.",
<   "documentation_url": "https://git-lfs-server.com/docs/errors",
<   "request_id": "123"
< }
```

401 responses should include a `LFS-Authenticate` header to tell the client what
form of authentication it requires. This is a placeholder in case the client
adds support for something other than Basic Authentication. This is meant to
mirror the standard `WWW-Authenticate` header. A new header is used so it
does not trigger the password prompt in browsers.

```
< HTTP/1.1 401 Unauthorized
< Content-Type: application/vnd.git-lfs+json
< LFS-Authenticate: Basic realm="Git LFS"
<
< {
<   "message": "Credentials needed.",
<   "documentation_url": "https://git-lfs-server.com/docs/errors",
<   "request_id": "123"
< }
```

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

## Extended upload & download protocols

By default it is assumed that all transfers (uploads & downloads) will be
performed via a singular HTTP resource and that the URLs provided in the
response are implemented as such. In this case each object is uploaded or
downloaded in its entirety through that one URL.

However, in order to support more advanced transfer features such as resuming,
chunking or delegation to other services, the client can indicate in the request
its ability to handle other transfer mechanisms. 

Here's a possible example:

```json
{
  "operation": "upload",
  "accept-transfers": "tus,resumable.js",
  "objects": [
    {
      "oid": "1111111",
      "size": 123
    }
  ]
}
```

The `accept-transfers` field is a comma-separated list of identifiers which the
client is able to support, in order of preference. In this hypothetical example
the client is indicating it is able to support resumable uploads using either
the tus.io protocol, or the resumable.js protocol. It is implicit that basic
HTTP resources are always supported regardless of the presence or content of
this item.

If the server is able to support one of the extended transfer mechanisms, it can
provide resources specific to that mechanism in the response, with an indicator
of which one it picked:

```json
{
  "transfer": "tus",
  "objects": [
   {
      "oid": "1111111",
      "size": 123,
      "actions": {
        "upload": {
          "href": "https://my.tus.server.com/files/1111111"
        }
      }
    }
  ]
}
```

In this case the server has chosen [tus.io](http://tus.io); in this case the
underlying transport is still HTTP, so the `href` is still a web URL, but the
exact sequence of calls and the headers sent & received are different from a
single resource upload. Other transfers may use other protocols.

__Note__: these API features are provided for future extension and the examples
shown may not represent actual features present in the current client).
