# Locking API proposal

## POST /locks

| Method  | Accept                         | Content-Type                   | Authorization |
|---------|--------------------------------|--------------------------------|---------------|
| `POST`  | `application/vnd.git-lfs+json` | `application/vnd.git-lfs+json` | Basic         |

### Request

```
> GET https://git-lfs-server.com/locks
> Accept: application/vnd.git-lfs+json
> Authorization: Basic
> Content-Type: application/vnd.git-lfs+json
>
> {
>   path: "/path/to/file",
>   remote: "origin",
>   latest_remote_commit: "d3adbeef",
>   committer: {
>     name: "Jane Doe",
>     email: "jane@example.com"
>   }
> }
```

### Response

* **Successful response**
```
< HTTP/1.1 201 Created
< Content-Type: application/vnd.git-lfs+json
<
< {
<   lock: {
<     id: "some-uuid",
<     path: "/path/to/file",
<     committer: {
<       name: "Jane Doe",
<       email: "jane@example.com"
<     },
<     commit_sha: "d3adbeef",
<     locked_at: "2016-05-17T15:49:06+00:00"
<   }
< }
```

* **Bad request: minimum commit not met**
```
< HTTP/1.1 400 Bad request
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "commit_needed": "other_sha"
< }
```

* **Bad request: lock already present**
```
< HTTP/1.1 409 Conflict
< Content-Type: application/vnd.git-lfs+json
<
< {
<   lock: {
<     /* the previously created lock */
<   },
<   error: "already created lock"
< }
```

* **Bad repsonse: server error**
```
< HTTP/1.1 500 Internal server error
< Content-Type: application/vnd.git-lfs+json
<
< {
<   error: "unable to create lock"
< }
```

## POST /locks/:id/unlock

| Method  | Accept                         | Content-Type | Authorization |
|---------|--------------------------------|--------------|---------------|
| `POST`  | `application/vnd.git-lfs+json` | None         | Basic         |

### Request

```
> POST https://git-lfs-server.com/locks/:id/unlock
> Accept: application/vnd.git-lfs+json
> Authorization: Basic
```

### Repsonse

* **Success: unlocked**
```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
<
< {
<   lock: {
<     id: "some-uuid",
<     path: "/path/to/file",
<     committer: {
<       name: "Jane Doe",
<       email: "jane@example.com"
<     },
<     commit_sha: "d3adbeef",
<     locked_at: "2016-05-17T15:49:06+00:00",
<     unlocked_at: "2016-05-17T15:49:06+00:00"
<   }
< }
}
```

* **Bad response: server error**
```
< HTTP/1.1 500 Internal error
< Content-Type: application/vnd.git-lfs+json
<
< {
<   error: "git-lfs/git-lfs: internal server error"
< }
```

## GET /locks

| Method | Accept                        | Content-Type | Authorization |
|--------|-------------------------------|--------------|---------------|
| `GET`  | `application/vnd.git-lfs+json | None         | Basic         |

### Request

```
> GET https://git-lfs-server.com/locks?filters...&cursor=&limit=
> Accept: application/vnd.git-lfs+json
> Authorization: Basic
```

### Response

* **Success: locks found**

Note: no matching locks yields a payload of `locks: []`, and a status of 200.

```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
<
< {
<   locks: [
<     {
<       id: "some-uuid",
<       path: "/path/to/file",
<       committer": {
<         name: "Jane Doe",
<         email: "jane@example.com"
<       },
<       commit_sha: "1ec245f",
<       locked_at: "2016-05-17T15:49:06+00:00"
<     }
<   ],
<   next_cursor: "optional-next-id",
<   error: "optional error"
< }
```

* **Bad response: some locks may have matched, but the server encountered an error**
```
< HTTP/1.1 500 Internal error
< Content-Type: application/vnd.git-lfs+json
<
< {
<   locks: [],
<   error: "git-lfs/git-lfs: internal server error"
< }
```
