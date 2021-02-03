# Git LFS Batch API

Added: v0.6

The Batch API is used to request the ability to transfer LFS objects with the
LFS server. The Batch URL is built by adding `/objects/batch` to the LFS server
URL.

Git remote: https://git-server.com/foo/bar</br>
LFS server: https://git-server.com/foo/bar.git/info/lfs<br>
Batch API: https://git-server.com/foo/bar.git/info/lfs/objects/batch

See the [Server Discovery doc](./server-discovery.md) for more info on how LFS
builds the LFS server URL.

All Batch API requests use the POST verb, and require the following HTTP
headers. The request and response bodies are JSON.

    Accept: application/vnd.git-lfs+json
    Content-Type: application/vnd.git-lfs+json

See the [Authentication doc](./authentication.md) for more info on how LFS
gets authorizes Batch API requests.

## Requests

The client sends the following information to the Batch endpoint to transfer
some objects:

* `operation` - Should be `download` or `upload`.
* `transfers` - An optional Array of String identifiers for transfer adapters
that the client has configured. If omitted, the `basic` transfer adapter MUST
be assumed by the server.
* `ref` - Optional object describing the server ref that the objects belong to. Note: Added in v2.4.
  * `name` - Fully-qualified server refspec.
* `objects` - An Array of objects to download.
  * `oid` - String OID of the LFS object.
  * `size` - Integer byte size of the LFS object. Must be at least zero.

Note: Git LFS currently only supports the `basic` transfer adapter. This
property was added for future compatibility with some experimental transfer
adapters. See the [API README](./README.md) for a list of the documented
transfer adapters.

```js
// POST https://lfs-server.com/objects/batch
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ... (if needed)
{
  "operation": "download",
  "transfers": [ "basic" ],
  "ref": { "name": "refs/heads/main" },
  "objects": [
    {
      "oid": "12345678",
      "size": 123
    }
  ]
}
```

#### Ref Property

The Batch API added the `ref` property in LFS v2.4 to support Git server authentication schemes that take the refspec into account. Since this is
a new addition to the API, servers should be able to operate with a missing or null `ref` property.

Some examples will illustrate how the `ref` property can be used.

* User `owner` has full access to the repository.
* User `contrib` has readonly access to the repository, and write access to `refs/heads/contrib`.

```js
{
  "operation": "download",
  "transfers": [ "basic" ],
  "objects": [
    {
      "oid": "12345678",
      "size": 123
    }
  ]
}
```

With this payload, both `owner` and `contrib` can download the requested object, since they both have read access.

```js
{
  "operation": "upload",
  "transfers": [ "basic" ],
  "objects": [
    {
      "oid": "12345678",
      "size": 123
    }
  ]
}
```

With this payload, only `owner` can upload the requested object.

```js
{
  "operation": "upload",
  "transfers": [ "basic" ],
  "ref": { "name": "refs/heads/contrib" },
  "objects": [
    {
      "oid": "12345678",
      "size": 123
    }
  ]
}
```

Both `owner` and `contrib` can upload the request object.

### Successful Responses

The Batch API should always return with a 200 status, unless there are some
issues with the request (bad authorization, bad json, etc). See below for examples of response errors. Check out the documented transfer adapters in the
[API README](./README.md) to see how Git LFS handles successful Batch responses.

Successful responses include the following properties:

* `transfer` - String identifier of the transfer adapter that the server
prefers. This MUST be one of the given `transfer` identifiers from the request.
Servers can assume the `basic` transfer adapter if none were given. The Git LFS
client will use the `basic` transfer adapter if the `transfer` property is
omitted.
* `objects` - An Array of objects to download.
  * `oid` - String OID of the LFS object.
  * `size` - Integer byte size of the LFS object. Must be at least zero.
  * `authenticated` - Optional boolean specifying whether the request for this
  specific object is authenticated. If omitted or false, Git LFS will attempt
  to [find credentials for this URL](./authentication.md).
  * `actions` - Object containing the next actions for this object. Applicable
  actions depend on which `operation` is specified in the request. How these
  properties are interpreted depends on which transfer adapter the client will
  be using.
    * `href` - String URL to download the object.
    * `header` - Optional hash of String HTTP header key/value pairs to apply
    to the request.
    * `expires_in` - Whole number of seconds after local client time when
      transfer will expire. Preferred over `expires_at` if both are provided.
      Maximum of 2147483647, minimum of -2147483647.
    * `expires_at` - String uppercase RFC 3339-formatted timestamp with second
      precision for when the given action expires (usually due to a temporary
      token).

Download operations MUST specify a `download` action, or an object error if the
object cannot be downloaded for some reason. See "Response Errors" below.

Upload operations can specify an `upload` and a `verify` action. The `upload`
action describes how to upload the object. If the object has a `verify` action,
the LFS client will hit this URL after a successful upload. Servers can use this
for extra verification, if needed. If a client requests to upload an object that
the server already has, the server should omit the `actions` property
completely. The client will then assume the server already has it.

```js
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "transfer": "basic",
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "authenticated": true,
      "actions": {
        "download": {
          "href": "https://some-download.com",
          "header": {
            "Key": "value"
          },
          "expires_at": "2016-11-10T15:29:07Z"
        }
      }
    }
  ]
}
```

If there are problems accessing individual objects, servers should continue to
return a 200 status code, and provide per-object errors. Here is an example:

```js
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "transfer": "basic",
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "error": {
        "code": 404,
        "message": "Object does not exist"
      }
    }
  ]
}
```

LFS object error codes should match HTTP status codes where possible:

* 404 - The object does not exist on the server.
* 410 - The object was removed by the owner.
* 422 - Validation error.

### Response Errors

LFS servers can respond with these other HTTP status codes:

* 401 - The authentication credentials are needed, but were not sent. Git LFS
will attempt to [get the authentication](./authentication.md) for the request
and retry immediately.
* 403 - The user has **read**, but not **write** access. Only applicable when
the `operation` in the request is "upload."
* 404 - The Repository does not exist for the user.
* 422 - Validation error with one or more of the objects in the request. This
  means that _none_ of the requested objects to upload are valid.

Error responses will not have an `objects` property. They will only have:

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 404 Not Found
// Content-Type: application/vnd.git-lfs+json

{
  "message": "Not found",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

HTTP 401 responses should include an `LFS-Authenticate` header to tell the
client what form of authentication it requires. If omitted, Git LFS will assume
Basic Authentication. This mirrors the standard `WWW-Authenticate` header with
a custom header key so it does not trigger password prompts in browsers.

```js
// HTTP/1.1 401 Unauthorized
// Content-Type: application/vnd.git-lfs+json
// LFS-Authenticate: Basic realm="Git LFS"

{
  "message": "Credentials needed",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

The following status codes can optionally be returned from the API, depending on
the server implementation.

* 406 - The Accept header needs to be `application/vnd.git-lfs+json`.
* 413 - The batch API request contained too many objects or the request was
otherwise too large.
* 429 - The user has hit a rate limit with the server.  Though the API does not
specify any rate limits, implementors are encouraged to set some for
availability reasons.
* 501 - The server has not implemented the current method.  Reserved for future
use.
* 507 - The server has insufficient storage capacity to complete the request.
* 509 - The bandwidth limit for the user or repository has been exceeded.  The
API does not specify any bandwidth limit, but implementors may track usage.

Some server errors may trigger the client to retry requests, such as 500, 502,
503, and 504.
