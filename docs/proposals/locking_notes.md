# Capture Locking Notes during Locks creation and retrieve same during List Locks.

## Concept

The git-lfs REST API for Locks creation should be able to accept additonal attribute (message/notes) which would be easy to have some information related to lock creation. This same information can be retrieved back with the git-lfs List Locks REST API.

  - Allow to have additional attribute to store the lock message/notes during lock creation
  - Include lock message/notes in the git-lfs List Locks REST API response

## API extensions

The current Git LFS File Locking API [https://github.com/git-lfs/git-lfs/blob/v3.2.0/docs/api/locking.md] doesn't have a field to capture some information related to Locks creation which would be useful to understand why and from where the lock was acquired. To get more understanding, below is the use case:

Consider an application where source control management is one of the feature and GitHub is the repository being used. Here there are 3 workspaces with 2 users working on a single Repository containing 6 files.

Here the User1 is working on FileA, FileB under Workspace W1, User2 is working on FileC, FileD under Workspace W2 and also User1 is working on FileE, FileF under Workspace W3.

So we acquire locks before working on these files respectively. Now we can see that W1 has locks on FileA, FileB. W2 has locks on FileC, FileD. W3 has locks on FileE, FileF.

Repository
Wl
-- User1
------ FileA
------ FileB
W2
-- User2
------ FileC
------ FileD
W3
-- User1
------ FileE
------ FileF

Now when we run List Locks request from W1, we will get all locks on the Repo as given below. Here we doesn't know from which workspaces, the locks were acquired.

User Workspace File
User1 -- FileA
User1 -- FileB
User1 -- FileE
User1 -- FileF
User2 -- FileC
User2 -- FileD

Our requirement is to get the workspace information along with locks, to know from where the locks were acquired as given below (say When fetchLocks is run from W1):

User Workspace File
User1 W1 FileA
User1 W1 FileB
User1 W3 FileE
User1 W3 FileF
User2 W2 FileC
User2 W2 FileD

So if we capture this workspace info (as per above use case) as part of custom attribute during lock creation, same can be retrieved back and displayed.

Sample pre-defined comment during acquiring lock is as given below:

<SCMComment xmlns='http://schemas.cordys.com/cws/1.0'> <Comment>Automatically locked by the web interface upon editing</Comment><UserDN>sysadmin</UserDN><Workspace>'Workspace A' from organization 'system'</Workspace><DateTime>1656321671318</DateTime></SCMComment>

Here UserDN element has the Logged in username of the Application. The owner attribute from the List Locks response would give us the Git Repo user. The Workspace element would give us the workspace information.

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

