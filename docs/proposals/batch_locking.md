# Batch locking API

When working with huge repositories, git-lfs locking API limitation of one lock per HTTP request can become a performance bottleneck.

See

* [#2978: File locks are extremely slow and basically unusable](https://github.com/git-lfs/git-lfs/issues/2978)
* [#3066: Feature: Add Batch API for locking and unlocking](https://github.com/git-lfs/git-lfs/issues/3066)

This proposal introduces a batch API for locking and unlocking with Subversion-compatible semantics.
See [lock-many](https://github.com/apache/subversion/blob/1.14.5/subversion/libsvn_ra_svn/protocol#L442-L449) and [unlock-many](https://github.com/apache/subversion/blob/1.14.5/subversion/libsvn_ra_svn/protocol#L455-L460) commands.

## Create locks

The client sends the following to create locks by sending a `POST` to `/locks/batch` (appended to the LFS server url).
Server should ensure that user has push access to the repository, and that files are locked exclusively to one user.
Either all or none of specified locks should be created as a result.

```json5
// POST https://lfs-server.com/locks/batch
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ...
{
  "operation": "lock",
  "ref": {
    "name": "refs/heads/my-feature"
  },
  "files": [
    {
      "path": "foo/bar.zip",
    },
    {
      "path": "baz/qux.zip",
    }
  ]
}
```

Properties:

* `operation` - should be `lock`
* `ref` - Optional object describing the server ref that the locks belong to.
    * `name` - Fully-qualified server refspec.
* `files` - Array of files that should be locked.
    * `path` - String path name of the file that is locked.
      This should be relative to the root of the repository working directory.

### Successful Response

Successful responses return the array of created locks.
See the "Create Lock" successful response section of individual git-lfs locking API to see what Lock properties are possible.

If client sent empty `files` list, server should answer with successful response anyway.
Client can use this behavior to determine whether server supports batch locking API.

```json5
// HTTP/1.1 200 Created
// Content-Type: application/vnd.git-lfs+json
{
  "locks": [
    {
      "id": "some-uuid",
      "path": "foo/bar.zip",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    },
    {
      "id": "some-another-uuid",
      "path": "baz/qux.zip",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    },
  ]
}
```

### Not Found Response

Server that lacks batch locking API should return 404.

Any 404 will do, but Git LFS will show a better error message with a json response.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
// HTTP/1.1 404 Not found
// Content-Type: application/vnd.git-lfs+json
{
  "message": "Not found",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Bad Response: Lock Exists

If **any** of the locks already exists, the server should reject the entire request and none of the locks should be created.

* `lock` - The existing Lock that clashes with the request.
* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
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

Server should require that user has push access to the repository before they can create locks.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
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
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "internal server error",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

## Delete locks

The client can delete locks, given their IDs, by sending a `POST` request to `/locks/batch` (appended to the LFS server url).
Server should ensure that caller has push access to the repository.
It should also prevent user from deleting another user's lock, unless the `force` property is given.
Either all or none of specified locks should be deleted as a result.

```json5
// POST https://lfs-server.com/locks/batch
// Accept: application/vnd.git-lfs+json
// Content-Type: application/vnd.git-lfs+json
// Authorization: Basic ...
{
  "operation": "unlock",
  "force": true,
  "ref": {
    "name": "refs/heads/my-feature"
  },
  locks: [
    {
      "id": "some-uuid",
    },
    {
      "id": "some-another-uuid",
    },
  ],
}
```

Properties:

* `operation` - should be `unlock`
* `ref` - Optional object describing the server ref that the locks belong to.
    * `name` - Fully-qualified server refspec.
* `locks` - Array of lock IDs that should be unlocked.
    * `id` - String ID of the Lock.
      Git LFS doesn't enforce what type of ID is used, as long as it's returned as a string.

### Successful Response

Successful deletion returns the list of deleted locks.
See the "Create Lock" successful response section of individual git-lfs locking API to see what Lock properties are possible.

If client sent empty `locks` list, server should answer with successful response anyway.
Client can use this behavior to determine whether server supports batch locking API.

```json5
// HTTP/1.1 200 Ok
// Content-Type: application/vnd.git-lfs+json
{
  "locks": [
    {
      "id": "some-uuid",
      "path": "foo/bar.zip",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    },
    {
      "id": "some-another-uuid",
      "path": "baz/qux.zip",
      "locked_at": "2016-05-17T15:49:06+00:00",
      "owner": {
        "name": "Jane Doe"
      }
    },
  ]
}
```

### Bad Response: Some of the locks could not be deleted

If **any** of the locks could not be deleted, the server should reject the entire request and none of the locks should be deleted.

* `locks` - List of descriptions of each lock that prevented operation from completing successfully.
  * `id` - String ID of the Lock.
  * `error`
    * `code` - HTTP error code for this lock.
      Should be the same as user would get if tried to delete this lock individually.
    * `message` - String error message.
    * `lock` - Optional details of existing lock.
* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report
  errors.

```json5
// HTTP/1.1 409 Conflict
// Content-Type: application/vnd.git-lfs+json
{
  "locks": [
    {
      "id": "some-uuid",
      "error": {
        "code": 403,
        "message": "lock owned by another user",
        "lock": {
          // optional details of existing lock, i.e., "id" and "locked_at" elements
        }
      }
    },
    {
      "id": "some-another-uuid",
      "error": {
        "code": 404,
        "message": "lock does not exist"
      }
    }
  ],
  "message": "no locks deleted due to 2 errors",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Not Found Response

Server that lacks batch locking API should return 404.

Any 404 will do, but Git LFS will show a better error message with a json response.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
// HTTP/1.1 404 Not found
// Content-Type: application/vnd.git-lfs+json
{
  "message": "Not found",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Unauthorized Response

Server should require that user has push access to the repository before they can delete locks.
Also, if the `force` parameter is omitted, or `false`, the user should only be allowed to delete locks that they created.

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
// HTTP/1.1 403 Forbidden
// Content-Type: application/vnd.git-lfs+json
{
  "message": "You must have push access to delete locks",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```

### Error Response

* `message` - String error message.
* `request_id` - Optional String unique identifier for the request.
  Useful for debugging.
* `documentation_url` - Optional String to give the user a place to report errors.

```json5
// HTTP/1.1 500 Internal server error
// Content-Type: application/vnd.git-lfs+json
{
  "message": "internal server error",
  "documentation_url": "https://lfs-server.com/docs/errors",
  "request_id": "123"
}
```
