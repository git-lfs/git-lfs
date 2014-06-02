# Git Media API

The server implements a simple API for uploading and downloading binary content.
Git repositories that use Git Media will specify a URI endpoint.  See the
[specification](spec.md) for how Git Media determines the server endpoint to use.

Use that endpoint as a base, and append the following relative paths to upload
and download from the Git Media server.

## GET objects/{oid}

The OID is the value from the pointer file.

```
> GET objects/{oid} HTTP/1.1
> Accept: application/vnd.git-media
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK
< Content-Type: application/vnd.git-media; header=git-media.265b3cb3f0530aae9010780a30d92a898400a5582081b21a099d51941eff
<
< --git-media.265b3cb3f0530aae9010780a30d92a898400a5582081b21a099d51941eff
< {binary contents}
```

The `Content-Type` header in the response will add a `header` parameter.  This
identifies a unique header that the server writes before sending the content.

The server returns a 404 if the file is not found.

You can also request just the JSON meta data of the files:

```
> GET objects/{OID} HTTP/1.1
> Accept: application/vnd.git-media+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK
< Content-Type: application/vnd.git-media+json
<
< {
<   "oid": "the-sha-256-signature",
<   "size": 123456
< }
```

The `oid` and `size` properties are required.  The server can extend the output
with custom properties.

## PUT objects/{oid}

This writes the file contents to the Git Media server.

```
> PUT objects/{oid} HTTP/1.1
> Accept: application/vnd.git-media
> Content-Type: application/octet-stream
> Authorization: Basic ...
> Content-Length: 123
>
> {binary contents}
>
< HTTP/1.1 200 OK
```

Responses:

* 200 - The file already exists.
* 201 - The file was uploaded successfully.
* 409 - The file contents do not match the OID.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.

## OPTIONS objects/{oid}

This is a pre-flight request to verify credentials before sending the file
contents.

```
> OPTIONS objects/{oid} HTTP/1.1
> Accept: application/vnd.git-media
> Authorization: Basic ...
>
< HTTP/1.1 200 OK
```

* 200 - The user is able to send the file.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.
