# Capture Locking Notes during Locks creation and retrieve same during List Locks.

## Concept

The git-lfs REST API for Locks creation should be able to accept additonal attribute (message/notes) which would be easy to have some information related to lock creation. This same information can be retrieved back with the git-lfs List Locks REST API.

  - Allow to have additional attribute to store the lock message/notes during lock creation
  - Include lock message/notes in the git-lfs List Locks REST API response

## API extensions

The current Git LFS File Locking API [https://github.com/git-lfs/git-lfs/blob/v3.2.0/docs/api/locking.md] doesn't have a field to capture some information related to Locks creation which would be useful to understand why and from where the lock was acquired.

With this enhancement, we can have some predefined comment as part of lock creation and get back same with the List Locks REST API so that it will be useful to differentiate and get more information on the File lock.

# Create Locks Enhancement API proposal

### Request

```
> POST https://lfs-server.com/locks
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ...
> {
>  "path": "foo/bar.zip",
>  "ref": {
>    "name": "refs/heads/my-feature"
>  },
>  "notes": "Lock applied from Workspace A"
> }
```

### Response

* **Successful response**
```
< HTTP/1.1 201 Created
< Content-Type: application/vnd.git-lfs+json
< {
<  "lock": {
<    "id": "some-uuid",
<    "path": "foo/bar.zip",
<    "locked_at": "2022-05-17T15:49:06+00:00",
<    "owner": {
<      "name": "Jane Doe"
<    },
<    "notes": "Lock applied from Workspace A"
<  }
< }
```

# List Locks Enhancement API proposal

### Request (with notes -- notes=true)

```
> GET https://lfs-server.com/locks?path=&id&cursor=limit&**notes=true**&refspec=
> Accept: application/vnd.git-lfs+json
> Authorization: Basic ... (if needed)
```

### Response

* **Successful response**
```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
< {
<  "locks": [
<    {
<      "id": "some-uuid",
<      "path": "foo/bar.zip",
<      "locked_at": "2022-05-17T15:49:06+00:00",
<      "owner": {
<        "name": "Jane Doe"
<      },
<      "notes": "Lock applied from Workspace A"
<    }
<  ],
<  "next_cursor": "optional next ID"
< }
```

### Request (with out notes)

```
> GET https://lfs-server.com/locks?path=&id&cursor=limit&refspec=
> Accept: application/vnd.git-lfs+json
> Authorization: Basic ... (if needed)
```

### Response

* **Successful response**
```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
< {
<  "locks": [
<    {
<      "id": "some-uuid",
<      "path": "foo/bar.zip",
<      "locked_at": "2022-05-17T15:49:06+00:00",
<      "owner": {
<        "name": "Jane Doe"
<      }
<    }
<  ],
<  "next_cursor": "optional next ID"
< }
```

# List Locks for Verification Enhancement API proposal

### Request (with notes)

```
> POST https://lfs-server.com/locks/verify
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ...
> {
>  "cursor": "optional cursor",
>  "limit": 100, // also optional
>  "ref": {
>    "name": "refs/heads/my-feature"
>  },
>  "notes" : true, // also optional
> }
```

### Response

* **Successful response**
```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
< {
<  "ours": [
<    {
<      "id": "some-uuid",
<      "path": "/path/to/file",
<      "locked_at": "2016-05-17T15:49:06+00:00",
<      "owner": {
<        "name": "Jane Doe"
<      },
<      "notes": "Lock applied from Workspace A"
<    }
<  ],
<  "theirs": [],
<  "next_cursor": "optional next ID"
< }
```

### Request (with out notes)

```
> POST https://lfs-server.com/locks/verify
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ...
> {
>  "cursor": "optional cursor",
>  "limit": 100, // also optional
>  "ref": {
>    "name": "refs/heads/my-feature"
>  }
> }
```

### Response

* **Successful response**
```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
< {
<  "ours": [
<    {
<      "id": "some-uuid",
<      "path": "/path/to/file",
<      "locked_at": "2016-05-17T15:49:06+00:00",
<      "owner": {
<        "name": "Jane Doe"
<      }
<    }
<  ],
<  "theirs": [],
<  "next_cursor": "optional next ID"
< }
```

