# Multipart HTTP transfer mode proposal

This is a proposal for a new transfer mode, designed to support multi-part HTTP uploads. This is a protocol extension to
Git LFS, defining a new transfer mode to be implemented by Git LFS clients and servers in addition to the current `basic`
transfer mode.

This proposal is based on the experimental `multipart-basic` transfor mode originally
[implemented by datopian/giftless](https://giftless.datopian.com/en/latest/multipart-spec.html).

## Reasoning
Many storage vendors and cloud vendors today offer an API to upload files in "parts" or "chunks", using multiple HTTP
requests, allowing improved stability and performance. This is especially handy when files are multiple gigabytes in
size, and a failure during the upload of a file would require re-uploading it, which could be extremely time consuming.

The purpose of the `multipart` transfer mode is to allow Git LFS servers and client facilitate direct-to-storage
uploads for backends supporting multipart or chunked uploads.

As the APIs offered by storage vendors differ greatly, `multipart` transfer mode will offer abstraction over most
of these complexities in hope of supporting as many storage vendors as possible.

## Terminology
Throughout this document, the following terms are in use:
* *LFS Server* - The HTTP server to which the LFS `batch` request is sent
* *Client* or *LFS Client* - a client using the Git LFS protocol to push large files to storage via an LFS server
* *Storage Backend* - The HTTP server handling actual storage; This may or may not be the same server as the LFS
server, and for the purpose of this document, typically it is not. A typical implementation of this protocol would have
the Storage Backend be a cloud storage service such as *Amazon S3* or *Google Cloud Storage*.

## Design Goals
* Abstract vendor specific API and flow into a generic protocol
* Remain as close as possible to the `basic` transfer API
* Work at least with the multi-part APIs of
  [Amazon S3](https://aws.amazon.com/s3/),
  [Google Cloud Storage](https://cloud.google.com/storage) and
  [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/),
* Define how uploads can be resumed by re-doing parts and not-redoing parts that were uploaded successfully
  (this may be vendor specific and not always supported)
* Do not require any state to be maintained in the server side

## High Level Protocol Specs
* The name of the transfer is `multipart`
* Batch requests are the same as `basic` requests except that `{"transfers": ["multipart", "basic"]}` is the
  expected transfers value. Clients MUST retain `basic` as the fallback transfer mode to ensure compatiblity with
  servers not implementing this extension.
* `{"operation": "download"}` replies work exactly like `basic` download request with no change
* `{"operation": "upload"}` replies will break the upload into several `actions`:
  * `parts` (optional), a list of zero or more part upload actions
  * `verify` (optional), an action to verify the file is in storage, similar to `basic` upload verify actions
  * `abort` (optional), an action to abort the upload and clean up all unfinished chunks and state
* Just like `basic` transfers, if the file fully exists and is committed to storage, no `actions` will be provided
  in the reply and the upload can simply be skipped
* If a `verify` action is provided, calling it is required and not optional. In some cases, this endpoint may be used
  to finalize the upload.
* An empty or missing list of `parts` with a `verify` action may mean all parts have been uploaded but `verify` still
  needs to be called by the client.
* Authentication and authorization behave just like with the `basic` protocol.

## Action Objects
Each one of the `parts` and the `verify` and `abort` actions contain instructions for the client on how to send a
request performing a the particular action. These are similar to `basic` transfer adapter `actions` but may include
some common, as well as action specific additional parameters.

All actions allow `href`, `header` and `expires_in` parameters just like `basic` transfer actions.

### `parts` actions
Each `parts` action should include the `pos` and `size` attributes, in addition to the attributes specified
above:
* `pos` indicate the position in bytes within the file in which the part should begin. If not specified, `0` (that is the
beginning of the file) is assumed.
* `size` is the size of the part in bytes. If `size` is omitted, default to read until the end of file.
* If both `pos` and `size` are omitted, the action is expected to be a single-part upload of the entire file

In addition, `parts` actions may include the following parameters:
* `method`, with `PUT` as the default method if none is specified. This allows customizing the HTTP method used when uploading
object parts.
* `want_digest` to specify an expected HTTP `Digest` header, as described below.

### `verify` action
The `verify` action is similar to `basic` transfer mode `verify`, with the following additional parameters:
* `params` - an object with additional parameters to send to the server when sending the `verify` request.
These parameters are to be sent to the server exactly as provided, as the value of the `params` JSON attribute.

### `abort` action
The `abort` action may include the `method` attribute as specified for `parts` actions above.

## Batch Request / Response Examples

### Upload Batch Request
The following is a ~10mb file upload request:
```json
{
  "transfers": ["multipart", "basic"],
  "operation": "upload",
  "objects": [
    {
      "oid": "20492a4d0d84f8beb1767f6616229f85d44c2827b64bdbfb260ee12fa1109e0e",
      "size": 10000000
    }
  ]
}
```

### Upload Batch Response
The following is a response for the same request, given an imaginary storage backend:

```json
{
  "transfer": "multipart",
  "objects": [
    {
      "oid": "20492a4d0d84f8beb1767f6616229f85d44c2827b64bdbfb260ee12fa1109e0e",
      "size": 10000000,
      "actions": {
        "parts": [
          {
            "href": "https://storage.cloud.example/storage/upload/20492a4d0d84?part=0",
            "header": {
              "Authorization": "Bearer someauthorizationtokenwillbesethere"
            },
            "pos": 0,
            "size": 2500000,
            "expires_in": 86400
          },
          {
            "href": "https://storage.cloud.example/storage/upload/20492a4d0d84?part=1",
            "header": {
              "Authorization": "Bearer someauthorizationtokenwillbesethere"
            },
            "pos": 2500000,
            "size": 2500000,
            "expires_in": 86400
          },
          {
            "href": "https://storage.cloud.example/storage/upload/20492a4d0d84?part=2",
            "header": {
              "Authorization": "Bearer someauthorizationtokenwillbesethere"
            },
            "pos": 5000000,
            "size": 2500000,
            "expires_in": 86400
          },
          {
            "href": "https://storage.cloud.example/storage/upload/20492a4d0d84?part=3",
            "header": {
              "Authorization": "Bearer someauthorizationtokenwillbesethere"
            },
            "pos": 7500000,
            "expires_in": 86400
          }
        ],
        "verify": {
          "href": "https://lfs.mycompany.example/myorg/myrepo/multipart/verify",
          "authenticated": true,
          "header": {
            "Authorization": "Basic 123abc123abc123abc123abc123="
          },
          "expires_in": 86400,
          "params": {
            "uploadId": "20492a4d0d84",
            "partIds": [0, 1, 2, 3]
          }
        },
        "abort": {
          "href": "https://storage.cloud.example/storage/upload/20492a4d0d84",
          "authenticated": true,
          "header": {
            "Authorization": "Basic 123abc123abc123abc123abc123="
          },
          "method": "DELETE",
          "expires_in": 86400
        }
      }
    }
  ]
}
```

### `verify` request example
Given the `batch` response above, after all parts have been uploaded the client should send the following `verify`
request to `https://lfs.mycompany.example/myorg/myrepo/multipart/verify`:

```
POST /myorg/myrepo/multipart/verify
Host: lfs.mycompany.example
Authorization: Basic 123abc123abc123abc123abc123=
Content-type: application/json

{
  "oid": "20492a4d0d84f8beb1767f6616229f85d44c2827b64bdbfb260ee12fa1109e0e",
  "size": 10000000,
  "params": {
    "uploadId": "20492a4d0d84",
    "partIds": [0, 1, 2, 3]
  }
}
```

Assuming that all parts were uploaded successfully, the server should respond with a `200 OK` response.

### `abort` request example
Given the `batch` response above, the client may choose to cancel the upload by sending the following
`abort` request to `https://storage.cloud.example/storage/upload/20492a4d0d84`:

```
> DELETE /storage/upload/20492a4d0d84
> Host: storage.cloud.example
> Content-length: 0
```

## Uploaded Part Digest
Some storage backends will support, or even require, uploading clients to send a digest of the uploaded part when
uploading the part. This is a useful capability even if not required, as it allows backends to validate each part
separately as it is uploaded.

To support this, `parts` request objects may include a `want_digest` value, which is expected to be a list of digest
algorithms in the same format of the `Want-Digest` HTTP header specified by [RFC-3230](https://tools.ietf.org/html/rfc3230).

Any cryptographically secure digest algorithm [registered with IANA](https://www.iana.org/assignments/http-dig-alg/http-dig-alg.xhtml)
via the process outlined in [RFC-3230](https://tools.ietf.org/html/rfc3230) may be specified in `want_digest`. Algorithms
considered cryptographically insecure, including `MD5` and `SHA-1`, should not be accepted. Namely, `contentMD5` is
**not** an accepted value of `want_digest`.

If one or more digest algorithms with non-zero q-value is specified in `want_digest`, clients *should* select a favored
supported algorithm, calculate the part digest using that algorithm, and send it when uploading the part using the `Digest` HTTP
header as specified by [RFC-3230 section 4.3.1](https://tools.ietf.org/html/rfc3230#section-4.3.1).

While clients may include the part digest calculated using more than one algorithm, this is typically not required and
should be avoided.

Note that if `want_digest` is specified but the client cannot support any of the requested algorithms, the client may
still choose to continue uploading parts without sending a `Digest` header. However, the storage server may choose
to reject the request in such cases.

### Uploaded Part Digest Example

#### Examples of a batch response with `want_digest` in the reply

With SHA-512 as a preferred algorithm, and SHA-256 as a less preferred option if SHA-512 is not possible:

```json
{
  "actions": {
    "parts": [
      {
        "href": "https://storage.cloud.example/storage/upload/20492a4d0d84?part=3",
        "header": {
          "Authorization": "Bearer someauthorizationtokenwillbesethere"
        },
        "pos": 7500001,
        "want_digest": "sha-512;q=1.0, sha-256;q=0.5"
      }
    ]
  }
}
```

#### Example of part upload request send to the storage server
Following on the `want_digest` value specified in the last example, the client should now send the following headers
to the storage server when uploading the part, assuming `SHA-512` is supported:

```
HTTP/1.1 PUT /storage/upload/20492a4d0d84?part=3
Authorization: Bearer someauthorizationtokenwillbesethere
Digest: SHA-512=thvDyvhfIqlvFe+A9MYgxAfm1q5thvDyvhfIqlvFe+A9MYgxAfm1q5=
```

## Expected HTTP Responses

For each one of the `parts`, as well as `abort` and `verify` requests sent by the client, the following responses are
to be expected:

* Any response with a `20x` status code is to be considered by clients as successful. This ambiguity is by design, to
support variances between vendors (which may use `200` or `201` to indicate a successful upload, for example).

* Any other response is to be considered as an error, and it is up to the client to decide whether the request should
be retried or not. Implementors are encouraged to follow standard HTTP error status code guidelines.

### `batch` replies for partially uploaded content
When content was already partially uploaded, the server is expected to return a normal reply but omit request and parts
which do not need to be repeated. If the entire file has been uploaded, it is expected that no `actions` value will be
returned, in which case clients should simply skip the upload.

However, if parts of the file were successfully uploaded while others weren't, it is expected that a normal reply would
be returned, but with less `parts` to send.

### `verify` HTTP 409 errors
An `HTTP 409` error on `verify` requests typically indicates that the file could not be fully committed or verified.
In this  case, clients should follow the following process to try and recover from the error:
* retry the `batch` request to see if any parts of the file were not uploaded yet. If there are still `parts` to upload
  (i.e. `parts` is not empty), proceed to upload them and re-do `verify`
* If `parts` is empty, it is possible that the file exists in storage but is corrupt / has wrong size. In this case
  it is recommended to issue an `abort` and re-attempt the same upload again
* It is recommended to take special note of the number of retries, to avoid infinite recovery attempt loops

## Additional Considerations

### Chunk sizing
It is up to the LFS server to decide the size of each file chunk.

### Action lifetime considerations
As multipart uploads tend to require much more time than simple uploads, it is recommended to allow for longer `expires_in`
values than one would consider for `basic` uploads. It is possible that the process of uploading a single object in multiple
parts may take several hours from `batch` to `verify`.

### Falling back to `basic` transfer for small files
Using multipart upload APIs has some complexity and speed overhead. For this reason, if a client specifies support for
both `multipart` and `basic` transfer modes in a batch request, and the object(s) uploaded are small enough to fit in a
single part upload, servers *may* choose to respond with a `basic` transfer mode even if `multipart` is supported:

For example a small (2mb) upload batch request:
```
{
  "transfers": ["multipart", "basic"],
  "operation": "upload",
  "objects": [
    {
      "oid": "13aea96040f2133033d103008d5d96cfe98b3361f7202d77bea97b2424a7a6cd",
      "size": 2000000
    }
  ]
}
```

May be responded with:
```
{
  "transfer": "basic",
  "objects": [
    ...
  ]
}
```
Even if the server does support `multipart`, as `basic` can be preferrable in this case.

## Implementation Notes

### Hiding initialization / commit complexities from clients
While `part` requests are typically quite similar between vendors, the specifics of multipart upload initialization and
commit procedures are very specific to vendors. For this reason, in many cases, it will be up to the LFS server to
take care of initialization and commit code. This is fine, as long as actual uploaded data is sent directly to the
storage backend.

For example, in the case of Amazon S3:
* All requests need to have an "upload ID" token which is obtained in an initial request
* When finalizing the upload, a special "commit" request need to be sent, listing all uploaded part IDs.

These are very hard to abstract in a way that would allow clients to send them directly to the server. In addition, as
we do not want to maintain any state in the server, there is a need to make two requests when finalizing the upload:
one to fetch a list of uploaded chunks, and another to send this list to the S3 finalization endpoint.

For this reason, it is expected that any initialization actions will be handled by the Git LFS server during the
`batch` request handling. In most cases, the `verify` action will also be responsible for any finalization / commit actions.
The `params` attribute of the `verify` action is designed specifically to transfer some vendor-specific "state" between
initialization and finalization of the upload process.

### Implementing Complex `abort` actions
Some storage backends will accept a simple `DELETE` or `POST` request to a URL, with no request body, in order to abort
the upload. In such cases, `abort` may refer directly to the storage backend. However, in cases where aborting the upload
requires more complex logic or some payload in the request body, `abort` actions should point to an endpoint of the LFS
server, and it should be up to the LFS server to abort the upload and clean up any partially uploaded parts.

As `abort` requests do not have a body, any parameters required by the LFS server in order to complete the request should
be passed as part of the URL in the `href` parameter.

It should be noted that clients will not always be able to `abort` partial uploads cleanly. Implementors are expected to
ensure proper cleanup of partially uploaded files via other means, such as a periodcal cron job that locates uncommitted
uploaded parts and deletes them.
