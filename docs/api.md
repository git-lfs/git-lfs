# Git Media API

The server implements a simple API for uploading and downloading binary content.
Git repositories that use Git Media will specify a URI endpoint.  See the
[specification](spec.md) for how Git Media determines the server endpoint to use.

Use that endpoint as a base, and append the following relative paths to upload
and download from the Git Media server.

## GET objects/{oid}

This gets either the object content, or the object's meta data.  The OID is the
value from the object pointer.

### Getting the content

To download the object content, send an Accept header of `application/vnd.git-media`.
The server returns the raw content back with a `Content-Type` of
`application/octet-stream`.

```
> GET objects/{oid} HTTP/1.1
> Accept: application/vnd.git-media
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK
< Content-Type: application/octet-stream
<
< {binary contents}
```

The server can also redirect to another location.  This is useful in cases where
you do not want to render user content on a domain with important cookies.  
Request headers like `Range` or `Accept` should be passed through.  The
`Authorization` header must _not_ be passed through if the location's host or
scheme differs from the original request uri.

```
> GET objects/{oid} HTTP/1.1
> Accept: application/vnd.git-media
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 302 Found
< Location: https://some-other-site.com/{oid}
<
< --git-media.265b3cb3f0530aae9010780a30d92a898400a5582081b21a099d51941eff
< {binary contents}
```

### Responses

* 200 - The object contents or meta data is in the response.
* 302 - Temporary redirect to a new location.
* 404 - The user does not have access to the object, or it does not exist.

### Getting meta data.

You can also request just the JSON meta data with an `Accept` header of
`application/vnd.git-media+json`.

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
<   "size": 123456,
<   "md5": "the-md5-signature",
<   "sha1": "the-sha1-signature"
< }
```

The `oid` and `size` properties are required.  The server can extend the output
with custom properties.

## PUT objects/{oid}

This writes the object contents to the Git Media server.

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

### Responses

* 200 - The object already exists.
* 201 - The object was uploaded successfully.
* 307 - The server is returning a different location to send the object.  Clients
  should retry the PUT against the URL.  The `Authorization` header must _not_ be
  passed through if the location's host or scheme differs from the original
  request uri.
* 409 - The object contents do not match the OID.
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

* 200 - The user is able to send the object but the server already has it.
* 204 - The user is able to send the object and the server does not have it.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.
