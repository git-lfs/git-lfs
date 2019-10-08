# Basic Transfer API

The Basic transfer API is a simple, generic API for directly uploading and
downloading LFS objects. Git LFS servers can offload object storage to cloud
services like S3, or implement this API natively.

This is the original transfer adapter. All Git LFS clients and servers SHOULD
support it, and default to it if the [Batch API](./batch.md) request or response
do not specify a `transfer` property.

## Downloads

Downloading an object requires a download `action` object in the Batch API
response that looks like this:

```json
{
  "transfer": "basic",
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "authenticated": true,
      "actions": {
        "download": {
          "href": "https://some-download.com/1111111",
          "header": {
            "Authorization": "Basic ..."
          },
          "expires_in": 86400,
        }
      }
    }
  ]
}
```

The Basic transfer adapter will make a GET request on the `href`, expecting the
raw bytes returned in the HTTP response.

```
> GET https://some-download.com/1111111
> Authorization: Basic ...
<
< HTTP/1.1 200 OK
< Content-Type: application/octet-stream
< Content-Length: 123
<
< {contents}
```

## Uploads

The client uploads objects through individual PUT requests. The URL and headers
are provided by an upload `action` object.

```json
{
  "transfer": "basic",
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "authenticated": true,
      "actions": {
        "upload": {
          "href": "https://some-upload.com/1111111",
          "header": {
            "Authorization": "Basic ..."
          },
          "expires_in": 86400
        }
      }
    }
  ]
}
```

The Basic transfer adapter will make a PUT request on the `href`, sending the
raw bytes returned in the HTTP request.

```
> PUT https://some-upload.com/1111111
> Authorization: Basic ...
> Content-Type: application/octet-stream
> Content-Length: 123
>
> {contents}
>
< HTTP/1.1 200 OK
```

## Verification

The Batch API can optionally return a verify `action` object in addition to an
upload `action` object. If given, The Batch API expects a POST to the href
after a successful upload.

```json
{
  "transfer": "basic",
  "objects": [
    {
      "oid": "1111111",
      "size": 123,
      "authenticated": true,
      "actions": {
        "upload": {
          "href": "https://some-upload.com/1111111",
          "header": {
            "Authorization": "Basic ..."
          },
          "expires_in": 86400
        },
        "verify": {
          "href": "https://some-verify-callback.com",
          "header": {
            "Authorization": "Basic ..."
          },
          "expires_in": 86400
        }
      }
    }
  ]
}
```

Git LFS clients send:

* `oid` - The String OID of the Git LFS object.
* `size` - The integer size of the Git LFS object, in bytes.

```
> POST https://some-verify-callback.com
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Content-Length: 123
>
> {"oid": "{oid}", "size": 10000}
>
< HTTP/1.1 200 OK
```

A 200 response means that the object exists on the server.
