# Hawser API

The server implements a simple API for uploading and downloading binary content.
Git repositories that use Hawser will specify a URI endpoint.  See the
[specification](spec.md) for how Git Media determines the server endpoint to use.

Use that endpoint as a base, and append the following relative paths to upload
and download from the Hawser server.

## GET /objects/{oid}

This gets either the object content, or the object's meta data.  The OID is the
value from the object pointer.

### Getting the content

To download the object content, send an Accept header of `application/vnd.hawser`.
The server returns the raw content back with a `Content-Type` of
`application/octet-stream`.

```
> GET https://hawser-server.com/objects/{oid} HTTP/1.1
> Accept: application/octet-stream
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
> GET https://hawser-server.com/objects/{oid} HTTP/1.1
> Accept: application/vnd.hawser
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 302 Found
< Location: https://storage-server.com/{oid}
<
< {binary contents}
```

### Responses

* 200 - The object contents or meta data is in the response.
* 302 - Temporary redirect to a new location.
* 404 - The user does not have access to the object, or it does not exist.

### Getting meta data.

You can also request just the JSON meta data with an `Accept` header of
`application/vnd.hawser+json`.  Here's an example successful request:

```
> GET https://hawser-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.hawser+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK
< Content-Type: application/vnd.hawser+json
<
< {
<   "oid": "the-sha-256-signature",
<   "size": 123456,
<   "_links": {
<     "download": {
<       "href": "https://some-download.com",
<       "header": {
<         "Key": "value"
<       }
<     }
<   }
< }
```

The `oid` and `size` properties are required.  A hypermedia `_links` section is
included with a `download` link relation, which describes how to download the
object content.  If the GET request to download an object (with `Accept:
application/octet-stream`) redirects somewhere else, a similar URL should be
used with the `download` relation.

Here's a sample response for a request with an authorization error:

```
> GET https://hawser-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.hawser+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 404 Not found
< Content-Type: application/vnd.hawser+json
<
< {
<   "message": "Not found"
< }
```

There are what the HTTP status codes mean:

* 200 - The user is able to read the object.
* 404 - The repository does not exist for the user, or the user does not have
access to it.

## OPTIONS /objects/{oid}

This is a pre-flight request to verify credentials before sending the file
contents.  Note: The `OPTIONS` method is only supported in pre-1.0 Hawser
clients.  After 1.0, clients should use the `GET` with the
`application/vnd.hawser+json` Accept header.

Here's an example successful request:

```
> OPTIONS https://hawser-server.com/objects/{OID} HTTP/1.1
> Accept: application/vnd.hawser+json
> Authorization: Basic ... (if authentication is needed)
>
< HTTP/1.1 200 OK

(no response body)
```

There are what the HTTP status codes mean:

* 200 - The user is able to read the object.
* 204 - The user is able to PUT the object to the same URL.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.
* 405 - OPTIONS not supported, use a GET request with a `application/vnd.hawser+json`
Accept header.

## POST /objects

This request initiates the upload of an object, given a JSON body with the oid
and size of the object to upload.

```
> POST https://hawser-server.com/objects/ HTTP/1.1
> Accept: application/vnd.hawser+json
> Content-Type: application/vnd.hawser+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "oid": "1111111",
>   "size": 123
> }
>
< HTTP/1.1 202 Accepted
< Content-Type: application/vnd.hawser+json
<
< {
<   "_links": {
<     "upload": {
<       "href": "https://some-upload.com",
<       "header": {
<         "Key": "value"
<       }
<     },
<     "verify": {
<       "href": "https://some-callback.com",
<       "header": {
<         "Key": "value"
<       }
<     }
<   }
< }
```

A response can include one of multiple link relations, each with an `href`
property and an optional `header` property.

* `upload` - This relation describes how to upload the object.
* `verify` - The server can specify a URL for the client to hit after
successfully uploading an object.
* `download` - This relation describes how to download the object content.

### Responses

* 200 - The object already exists.  Don't bother re-uploading.
* 202 - The object is ready to be uploaded.Follow the "upload" and optional
"verify" links.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.

## PUT /objects/{oid}

This writes the object contents to the Git Media server.

```
> PUT https://hawser-server.com/objects/{oid} HTTP/1.1
> Accept: application/vnd.hawser
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
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.
* 405 - PUT method is not allowed.  Use an OPTIONS or GET pre-flight request to
get the current URL to send a file.

## Verification

When Hawser clients issue a POST request to initiate an object upload, the
response can potentially return a "verify" link relation.  If given, The Hawser
server expects a POST to the href after a successful upload.  Hawser
clients send:

* `oid` - The String OID of the Git Media object.
* `size` - The integer size of the Git Media object.

```
> POST https://hawser-server.com/callback
> Accept: application/vnd.hawser
> Content-Type: application/vnd.hawser+json
> Content-Length: 123
>
> {"oid": "{oid}", "size": 10000}
>
< HTTP/1.1 200 OK
```

A 200 response means that the object exists on the server.
