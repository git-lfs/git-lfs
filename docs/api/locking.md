# Git LFS File Locking API

Added: v2.0

The File Locking API is used to create, list, and delete locks, as well as
verify that locks are respected in Git pushes. The locking URLs are built
by adding a suffix to the LFS Server URL.

Git remote: https://git-server.com/foo/bar<br>
LFS server: https://git-server.com/foo/bar.git/info/lfs<br>
Locks API: https://git-server.com/foo/bar.git/info/lfs/locks<br>

See the [Server Discovery doc](./server-discovery.md) for more info on how LFS
builds the LFS server URL.

All File Locking requests require the following HTTP headers:

    Accept: application/vnd.git-lfs+json
    Content-Type: application/vnd.git-lfs+json

See the [Authentication doc](./authentication.md) for more info on how LFS
gets authorizes Batch API requests.

Note: This is the first version of the File Locking API, supporting only the
simplest use case: single branch locking. The API is designed to be extensible
as we experiment with more advanced locking scenarios, as defined in the
[original proposal](/docs/proposals/locking.md).

The [Batch API's `ref` property docs](./batch.md#ref-property) describe how the `ref` property can be used to support auth schemes that include the server ref. Locking API implementations should also only use it for authentication, until advanced locking scenarios have been developed.

## Create Lock

The client sends the following to create a lock by sending a `POST` to `/locks`
(appended to the LFS server url, as described above). Servers should ensure that
users have push access to the repository, and that files are locked exclusively
to one user.

* `path` - String path name of the file that is locked. This should be
relative to the root of the repository working directory.
* `ref` - Optional object describing the server ref that the locks belong to. Note: Added in v2.4.
  * `name` - Fully-qualified server refspec.

```js
// POST https://lfs-server.com/locks
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ...
{
  "path": "foo/bar.zip",
  "ref": {
    "name": "refs/heads/my-feature"
  }
}
```

### Successful Response

Successful responses return the created lock:

* `id` - String ID of the Lock. Git LFS doesn't enforce what type of ID is used,
as long as it's returned as a string.
* `path` - String path name of the locked file. This should be relative to the
root of the repository working directory.
* `locked_at` - The timestamp the lock was created, as an uppercase
RFC 3339-formatted string with second precision.
* `owner` - Optional name of the user that created the Lock. This should be set from
the user credentials posted when creating the lock.

```js
// HTTP/1.1 201 Created
// Content-Type: application/vnd.git-lfs+json
{
  "lock": {
    "id": "some-uuid",
    "path": "foo/bar.zip",
    "locked_at": "2016-05-17T15:49:06+00:00",
    "owner": {
      "name": "Jane Doe"
    }
  }
}
```

### Bad Response: Lock Exists

Lock services should reject lock creations if one already exists for the given
path on the current repository.

* `lock` - The existing Lock that clashes with the request.
* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 409 Conflict
// Content-Type: application/vnd.git-lfs+json
{
  "lock": {
    // details of existing lock
  },
  "message": "already created lock",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Unauthorized Response

Lock servers should require that users have push access to the repository before
they can create locks.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 403 Forbidden
// Content-Type: application/vnd.git-lfs+json
{
  "message": "You must have push access to create a lock",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Error Response

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "internal server error",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

## List Locks

The client can request the current active locks for a repository by sending a
`GET` to `/locks` (appended to the LFS server url, as described above).  LFS
Servers should ensure that users have at least pull access to the repository.

The properties are sent as URI query values, instead of through a JSON body:

* `path` - Optional string path to match against locks on the server.
* `id` - Optional string ID to match against a lock on the server.
* `cursor` - The optional string value to continue listing locks. This value
should be the `next_cursor` from a previous request.
* `limit` - The integer limit of the number of locks to return. The server
should have its own upper and lower bounds on the supported limits.
* `refspec` - Optional fully qualified server refspec
from which to search for locks.

```js
// GET https://lfs-server.com/locks?path=&id=&cursor=&limit=&refspec=
// Accept: application/vnd.git-lfs+json
// Authorization: Basic ... (if needed)
```

### Successful Response

A successful response will list the matching locks:

* `locks` - Array of matching Lock objects. See the "Create Lock" successful
response section to see what Lock properties are possible.
* `next_cursor` - Optional string cursor that the server can return if there
are more locks matching the given filters. The client will re-do the request,
setting the `?cursor` query value with this `next_cursor` value.

Note: If the server has no locks, it must return an empty `locks` array.

```js
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "locks": [
    {
      "id": "some-uuid",
      "path": "/path/to/file",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    }
  ],
  "next_cursor": "optional next ID"
}
```

### Unauthorized Response

Lock servers should require that users have pull access to the repository before
they can list locks.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 403 Forbidden
// Content-Type: application/vnd.git-lfs+json
{
  "message": "You must have pull access to list locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Error Response

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "unable to list locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

## List Locks for Verification

The client can use the Lock Verification endpoint to check for active locks
that can affect a Git push. For a caller, this endpoint is very similar to the
"List Locks" endpoint above, except:

* Verification requires a `POST` request.
* The `cursor`, `ref` and `limit` values are sent as properties in the json
request body.
* The response includes locks partitioned into `ours` and `theirs` properties.

LFS Servers should ensure that users have push access to the repository.

Clients send the following to list locks for verification by sending a `POST`
to `/locks/verify` (appended to the LFS server url, as described above):

* `ref` - Optional object describing the server ref that the locks belong to. Note: Added in v2.4.
  * `name` - Fully-qualified server refspec.
* `cursor` - Optional cursor to allow pagination. Servers can determine how cursors are formatted based on how they are stored internally.
* `limit` - Optional limit to how many locks to
return.

```js
// POST https://lfs-server.com/locks/verify
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ...
{
  "cursor": "optional cursor",
  "limit": 100, // also optional
  "ref": {
    "name": "refs/heads/my-feature"
  }
}
```

Note: As more advanced locking workflows are implemented, more details will
likely be added to this request body in future iterations.

### Successful Response

A successful response will list the relevant locks:

* `ours` - Array of Lock objects currently owned by the authenticated user.
modify.
* `theirs` - Array of Lock objects currently owned by other users.
* `next_cursor` - Optional string cursor that the server can return if there
are more locks matching the given filters. The client will re-do the request,
setting the `cursor` property with this `next_cursor` value.

If a Git push updates any files matching any of "our" locks, Git LFS will list
them in the push output, in case the user will want to unlock them after the
push. However, any updated files matching one of "their" locks will halt the
push. At this point, it is up to the user to resolve the lock conflict with
their team.

Note: If the server has no locks, it must return an empty array in the `ours` or
`theirs` properties.

```js
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "ours": [
    {
      "id": "some-uuid",
      "path": "/path/to/file",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    }
  ],
  "theirs": [],
  "next_cursor": "optional next ID"
}
```

### Not Found Response

By default, an LFS server that doesn't implement any locking endpoints should
return 404. This response will not halt any Git pushes.

Any 404 will do, but Git LFS will show a better error message with a json
response.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 404 Not found
// Content-Type: application/vnd.git-lfs+json
{
  "message": "Not found",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Unauthorized Response

Lock servers should require that users have push access to the repository before
they can get a list of locks to verify a Git push.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 403 Forbidden
// Content-Type: application/vnd.git-lfs+json
{
  "message": "You must have push access to verify locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Error Response

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "unable to list locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

## Delete Lock

The client can delete a lock, given its ID, by sending a `POST` to
`/locks/:id/unlock` (appended to the LFS server url, as described above). LFS
servers should ensure that callers have push access to the repository. They
should also prevent a user from deleting another user's lock, unless the `force`
property is given.

Properties:

* `force` - Optional boolean specifying that the user is deleting another user's
lock.
* `ref` - Optional object describing the server ref that the locks belong to. Note: Added in v2.4.
  * `name` - Fully-qualified server refspec.

```js
// POST https://lfs-server.com/locks/:id/unlock
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ...

{
  "force": true,
  "ref": {
    "name": "refs/heads/my-feature"
  }
}
```

### Successful Response

Successful deletions return the deleted lock. See the "Create Lock" successful
response section to see what Lock properties are possible.

```js
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "lock": {
    "id": "some-uuid",
    "path": "/path/to/file",
    "locked_at": "2016-05-17T15:49:06+00:00",
    "owner": {
      "name": "Jane Doe"
    }
  }
}
```

### Unauthorized Response

Lock servers should require that users have push access to the repository before
they can delete locks. Also, if the `force` parameter is omitted, or false,
the user should only be allowed to delete locks that they created.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 403 Forbidden
// Content-Type: application/vnd.git-lfs+json
{
  "message": "You must have push access to delete locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Error response

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request. Useful for
debugging.
* `documentation_url` - Optional String to give the user a place to report
errors.

```js
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "unable to delete lock",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```
