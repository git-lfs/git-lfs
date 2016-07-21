# Git LFS v1.3 Batch API

The Git LFS Batch API extends the [batch v1 API](../v1/http-v1-batch.md), adding 
optional fields to the request and response to negotiate transfer methods. 

Only the differences from the v1 API will be listed here, everything else is
unchanged.

## POST /objects/batch
### Request changes

The v1.3 request adds an additional optional top-level field, `transfers`, 
which is an array of strings naming the transfer methods this client supports.

The default transfer method which simply uploads and downloads using simple HTTP 
`PUT` and `GET`, named "basic", is always supported and is implied.

Example request:

```
> POST https://git-lfs-server.com/objects/batch HTTP/1.1
> Accept: application/vnd.git-lfs+json
> Content-Type: application/vnd.git-lfs+json
> Authorization: Basic ... (if authentication is needed)
>
> {
>   "operation": "upload",
>   "transfers": [ "tus.io", "basic" ],
>   "objects": [
>     {
>       "oid": "1111111",
>       "size": 123
>     }
>   ]
> }
>
```

In the example above `"basic"` is included for illustration but is actually
unnecessary since it is always the fallback. The client is indicating that it is
able to upload using the resumable `"tus.io"` method, should the server support
that. The server may include a chosen method in the response, which must be
one of those listed, or `"basic"`.

### Response changes

If the server understands the new optional `transfers` field in the request, it
should determine which of the named transfer methods it also supports, and
include the chosen one in the response in the new `transfer` field. If only
`"basic"` is supported, the field is optional since that is the default.

If the server supports more than one of the named transfer methods, it should
pick the best one it supports.

Example response to the previous request if the server also supports `tus.io`:

```
< HTTP/1.1 200 Ok
< Content-Type: application/vnd.git-lfs+json
<
< {
<   "transfer": "tus.io",
<   "objects": [
<     {
<       "oid": "1111111",
<       "size": 123,
<       "actions": {
<         "upload": {
<           "href": "https://some-tus-io-upload.com",
<           "header": {
<             "Key": "value"
<           }
<         },
<         "verify": {
<           "href": "https://some-callback.com",
<           "header": {
<             "Key": "value"
<           }
<         }
<       }
>     }
<   ]
< }
```

Apart from naming the chosen transfer method in `transfer`, the server should
also return upload / download links in the `href` field which are compatible 
with the method chosen. If the server supports more than one method (and it's
advisable that the server implement at least `"basic` in all cases in addition 
to more sophisticated methods, to support older clients), the `href` is likely 
to be different for each. 

## Updated schemas

* [Batch request](./http-v1.3-batch-request-schema.json)
* [Batch response](./http-v1.3-batch-response-schema.json)
