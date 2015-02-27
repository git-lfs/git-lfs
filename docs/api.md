# Hawser API

The server implements a simple API for uploading and downloading binary content.
Git repositories that use Hawser will specify a URI endpoint.  See the
[specification](spec.md) for how Git Media determines the server endpoint to use.

Use that endpoint as a base, and append the following relative paths to upload
and download from the Hawser server.

All requests should send an Accept header of `application/vnd.hawser+json`.
This may change in the future as the API evolves.

## Hypermedia

The Hawser API uses hypermedia hints to instruct the client what to do next.
These links are included in a `_links` property.  Possible relations for objects
include:

* `self` - This points to the object's canonical URL.
* `download` - Follow this link with a GET and the optional header values to
download the object content.
* `upload` - Upload the object content to this link with a PUT.
* `verify` - Optional link for the client to POST after an upload.  If
included, the client assumes this step is required after uploading an object.
See the "Verification" section below for more.

Link relations specify the `href`, and optionally a collection of header values
to set for the request.  These are optional, and depend on the backing object
store that the Hawser API is using.  

The Hawser client will automatically send the same credentials to the followed
link relation as Basic Authentication IF:

* The url scheme, host, and port all match the Hawser API endpoint's.
* The link relation does not specify an Authorization header.

If the host name is different, the Hawser API needs to send enough information
through the href query or header values to authenticate the request.

## GET /objects/{oid}

This gets the object's meta data.  The OID is the value from the object pointer.

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
<     "self": {
<       "href": "https://hawser-server.com/objects/OID",
<     },
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
included link relations for the object.  An object will either return a
`download` or an `upload` link.  The `verify` link is optional, and depends on
the server's configuration.  See the "Hypermedia" section above for more.

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

### Responses

* 200 - The object contents or meta data is in the response.
* 404 - The user does not have access to the object, or it does not exist.

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
* 202 - The object is ready to be uploaded.  Follow the "upload" and optional
"verify" links.
* 403 - The user has **read**, but not **write** access.
* 404 - The repository does not exist for the user.

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
