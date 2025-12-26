# Adding Custom Transfer Agents to LFS

## Introduction

Git LFS supports multiple ways to transfer (upload and download) files. In the
core client, the basic way to do this is via a one-off HTTP request via the URL
returned from the LFS API for a given object. The core client also supports
extensions to allow resuming of downloads (via `Range` headers) and uploads (via
the [tus.io](http://tus.io) protocol).

Some people might want to be able to transfer content in other ways, however.
To enable this, git-lfs allows configuring Custom Transfers, which are
simply processes which must adhere to the protocol defined later in this
document. git-lfs will invoke the process at the start of all transfers,
and will communicate with the process via stdin/stdout for each transfer.

## Custom Transfer Type Selection

In the LFS API request, the client includes a list of transfer types it
supports. When replying, the API server will pick one of these and make any
necessary adjustments to the returned object actions, in case the picked
transfer type needs custom details about how to do each transfer.

## Using a Custom Transfer Type without the API server

In some cases the transfer agent can figure out by itself how and where
the transfers should be made, without having to query the API server.
In this case it's possible to use the custom transfer agent directly,
without querying the server, by using the following config option:

* `lfs.standalonetransferagent`, `lfs.<url>.standalonetransferagent`

  Specifies a custom transfer agent to be used if the API server URL matches as
  in `git config --get-urlmatch lfs.standalonetransferagent <apiurl>`.
  `git-lfs` will not contact the API server.  It instead sets stage 2 transfer
  actions to `null`.  `lfs.<url>.standalonetransferagent` can be used to
  configure a custom transfer agent for individual remotes.
  `lfs.standalonetransferagent` unconditionally configures a custom transfer
  agent for all remotes.  The custom transfer agent must be specified in
  a `lfs.customtransfer.<name>` settings group.

## Defining a Custom Transfer Type

A custom transfer process is defined under a settings group called
`lfs.customtransfer.<name>`, where `<name>` is an identifier (see
[Naming](#naming) below).

* `lfs.customtransfer.<name>.path`

  `path` should point to the process you wish to invoke. This will be invoked at
  the start of all transfers (possibly many times, see the `concurrent` option
  below) and the protocol over stdin/stdout is defined below in the
  [Protocol](#protocol) section.

* `lfs.customtransfer.<name>.args`

  If the custom transfer process requires any arguments, these can be provided
  here. Typically you would only need this if your process was multi-purpose or
  particularly flexible, most of the time you won't need it.  Note that this
  string will be expanded by the shell.

* `lfs.customtransfer.<name>.concurrent`

  If true (the default), git-lfs will invoke the custom transfer process
  multiple times in parallel, according to `lfs.concurrenttransfers`, splitting
  the transfer workload between the processes.

  If you would prefer that only one instance of the transfer process is invoked,
  maybe because you want to do your own parallelism internally (e.g. slicing
  files into parts), set this to false.

* `lfs.customtransfer.<name>.direction`

  Specifies which direction the custom transfer process supports, either
  `download`, `upload`, or `both`. The default if unspecified is `both`.

* `lfs.customtransfer.<name>.protocol`

  Specifies which protocol version the custom transfer process supports. The
  current protocol version is `2`. If unspecified, protocol version `1` is used.

* `lfs.customtransfer.<name>.concurrencyMode`

  Specifies the tranfer concurreny mode. This setting can be set either to
  `basic`, `batch` or `any` for using one of the transfer concurrency modes. 
  These transfer concurrency modes are explained below. The default if 
  unspecified is `basic`.
  If set to `any` the tranfer should specify the concurrency mode explicitly on the initation stage, the `basic` assumed otherwise. `any` allows falling back 
  from `batch` to `basic` in case of batch transfer failure that cannot be 
  retried. `

## Naming

Each custom transfer must have a name which is unique to the underlying
mechanism, and the client and the server must agree on that name. The client
will advertise this name to the server as a supported transfer approach, and if
the server supports it, it will return relevant object action links. Because
these may be very different from standard HTTP URLs it's important that the
client and server agree on the name.

For example, let's say I've implemented a custom transfer process which uses
NFS. I could call this transfer type `nfs` - although it's not specific to my
configuration exactly, it is specific to the way NFS works, and the server will
need to give me different URLs. Assuming I define my transfer like this, and the
server supports it, I might start getting object action links back like
`nfs://<host>/path/to/object`

## Transfer concurrency modes

Transfer concurrency mode defines the way Git Lfs schedules operations to 
custom transfer process and the protocol version to be used for communicating 
between custom transfer and the Git Lfs client. 
There are two concurrency modes available for custom transfers: basic and batch.

### Basic

Basic concurrency mode uses single artifact as unit of work for every transfer 
operation.
Git Lfs will request every process to transfer a single artifact per time and 
expect the progress/result report for this single artifact before moving to the 
next artifact.
Concurrency in this mode applies per artifact - every single concurrent process
gets a single unique atrifact to work with per time.

### Batch

Batch concurrency mode uses group (batch) of artifacts as unit of work for every 
transfer operation. Git Lfs will request every process to transfer a batch of 
artifacts per time and expect the progress/result report for every atrifact in 
batch before moving to next batch.
Concurrency in this mode applies per batch - every concurrent process gets a 
batch of unique artifacts to work with per time.

### Transfer concurrency mode selection and behaviour

The effective transfer concurrency mode is determined as follows:
1. If `lfs.customtransfer.<name>.concurrencyMode` is set to `basic` or is not 
specified, Git Lfs will use basic mode and the basic protocol for all transfers.
2. If `lfs.customtransfer.<name>.concurrencyMode` is set to `batch`, Git Lfs will 
attempt to initiate the batch transfer protocol. The custom transfer process must 
confirm the batch mode support on the initiation stage, otherwise the transfer will 
fail and no fallback will be attempted.
3. If `lfs.customtransfer.<name>.concurrencyMode` is set to `any`, Git Lfs will 
attempt to initiate the batch transfer protocol. If the custom process confirms 
the batch mode support it will be used.
Otherwise Git Lfs will fallback to basic mode and continue the transfers using 
basic protocol. Initiation stage will not be repeated.
This mode also allows to fallback from batch to basic mode due to errors without 
losing any progress.
To trigger the fallback batch transfer process should report a batch failure with 
a non-retriable error on the batch transfer stages.

While using batch mode partial batch failures are allowed. This means that batch 
transfer can be reported as successful though some items from the batch are 
failed with retriable errors. For such batches the successfully transferred items 
will be used and the rest will be included into the next batch transfer for 
retry.
Failing the batch with a non-retryable error will trigger the fallback to `basic` 
mode if allowed by mode configurations. Retrying the whole batch will discard any 
successfully transferred items and retry the whole batch again.

## Protocol

The git-lfs client communicates with the custom transfer process via the stdin
and stdout streams. No file content is communicated on these streams, only
request / response metadata. The metadata exchanged is always in JSON format.
External files will be referenced when actual content is exchanged.

### Line Delimited JSON

Because multiple JSON messages will be exchanged on the same stream it's useful
to delimit them explicitly rather than have the parser find the closing `}` in
an arbitrary stream, therefore each JSON structure will be sent and received on
a **single line** as per [Line Delimited
JSON](https://en.wikipedia.org/wiki/JSON_Streaming#Line_delimited_JSON_2).

In other words when git-lfs sends a JSON message to the custom transfer it will
be on a single line, with a line feed at the end. The transfer process must
respond the same way by writing a JSON structure back to stdout with a single
line feed at the end (and flush the output).

### Versioning
Since this protocol is likely to evolve over time, each protocol version is tied to a
specific version number. The current protocol version is 2. 
To maintain backward compatibility, git-lfs will negotiate the protocol version with the
custom transfer first and only behave according to the supported protocol version protocol 
and features features inside a single custom transfer session. 
The effective protocol version can differ from current latest version and is determined 
during the negotiation process.

Section below describes each protocol version in detail to allow implementing the correct 
version.

## Protocol Specification

This section contains the detailed specification of the protocol stages and versions.

Stage 1 and Stage 3 are common between all protocol versions, while Stage 2 is defined 
per protocol version.

### Stage 1: Initiation and version negotiation

Immediately after invoking a custom transfer process, git-lfs sends initiation
data to the process over stdin. This tells the process useful information about
the configuration and allows to perform effective protocol version negotiation. 

Effective protocol version negotiation process is arranged according to the following rules:
1. Git Lfs always starts the negotiation by sending the initiation message with
    the current configured protocol version. The protocol version is configured 
    via `lfs.customtransfer.<name>.protocol` setting. If this setting is not
    specified Git Lfs assumes protocol version `1`.
2. The custom transfer process should respond with the supported protocol version
    and features to be used during the session.
3. If the custom transfer process responds with the same protocol version as
    the one sent by Git Lfs, Git Lfs will continue the session using the
    current protocol version features.    
4. If the custom transfer process responds with a lower protocol version than
    the one sent by Git Lfs, Git Lfs will switch to the lower protocol version
    and continue the session using the lower protocol version features.
5. If the custom transfer process responds with unrecognized protocol version, 
    Git Lfs will terminate the session with an error.
6. If the custom transfer process responds with no explicit protocol version,
    Git Lfs will continue the session using the protocol version 1.
    This allows backward compatibility with existing custom transfer implementations
    that don't support the protocol version negotiation.

#### Default behaviour

In case the protocol version is configured to `1` or not specified the initiation message 
will look like this:

```json
{ "event": "init", "operation": "download", "remote": "origin", "concurrent": true, "concurrenttransfers": 3 }
```

* `event`: always `init` to identify this message
* `operation`: will be `upload` or `download` depending on transfer direction
* `remote`: The Git remote.  It can be a remote name like `origin` or an URL
  like `ssh://git.example.com//path/to/repo`.  A standalone transfer agent can
  use it to determine the location of remote files.
* `concurrent`: reflects the value of `lfs.customtransfer.<name>.concurrent`, in
  case the process needs to know
* `concurrenttransfers`: reflects the value of `lfs.concurrenttransfers`, for if
  the transfer process wants to implement its own concurrency and wants to
  respect this setting.

No version negotiation is performed by default to maintain backward compatibility.

The transfer process should use the information it needs from the initiation
structure, and also perform any one-off setup tasks it needs to do. It should
then respond on stdout with a simple empty confirmation structure, as follows:


```json
{ }
```

Or if there was an error:

```json
{ "error": { "code": 32, "message": "Some init failure message" } }
```

In case of any error the initiation step will not be retried.

#### Versioned behaviour

In case the protocol version is configured to 2 or above the initiation 
message look like this:

```json
{ "event": "init", "operation": "download", "remote": "origin", "concurrent": true, "concurrenttransfers": 3, "protocol": 2, "concurrencyMode": "batch" }
```

* `event`: always `init` to identify this message
* `operation`: will be `upload` or `download` depending on transfer direction
* `remote`: the Git remote.  It can be a remote name like `origin` or an URL
  like `ssh://git.example.com//path/to/repo`.  A standalone transfer agent can
  use it to determine the location of remote files.
* `concurrent`: reflects the value of `lfs.customtransfer.<name>.concurrent`, in
  case the process needs to know
* `concurrenttransfers`: reflects the value of `lfs.concurrenttransfers`, for if
  the transfer process wants to implement its own concurrency and wants to
  respect this setting.
* `protocol`: the current configured protocol version. This is configured via
  `lfs.customtransfer.<name>.protocol` setting. 
* `concurrencyMode`: the current configured transfer concurrency mode. This is 
  configured via `lfs.customtransfer.<name>.concurrencyMode` setting.

The version negotiation is performed according to the rules defined above.
The behaviour on this stage is affects the effective protocol version and features.

The transfer process should use the information it needs from the initiation
structure, and also perform any one-off setup tasks it needs to do. It should
then respond on stdout with a simple empty confirmation structure, as follows:

```json
{ "protocol": 2, "concurrencyMode": "batch" }
```

* `protocol`: the protocol version supported by the custom transfer process.
  If this version is lower than the one sent by Git Lfs, Git Lfs will switch
  to this lower protocol version for the session. If it's higher, Git Lfs
  will terminate the session with an error.
* `concurrencyMode`: optional value to specify the concurrency mode to use.
  If this value is set to `batch` - this will confirm the batch concurrency mode
  support and enable it. All further communications will be performed according to
  batch protocol. If this value is not specified or set to `basic` - this will 
  switch the transfer to basic concurrency mode and all further communications will 
  be performed according to basic protocol.

Or if the transfer is aware about the newer protocol version but does not support it:

```json
{ "protocol": 1 }
```

* `protocol`: the protocol version supported by the custom transfer process.
  In this case Git Lfs will switch to protocol version 1 for the session.

In this case Git Lfs will continue the session using the protocol version `1`.

Or if the transfer process doesn't specify the protocol version:

```json
{ }
```

In this case Git Lfs will continue the session using the protocol version `1`.

Or if there was an error:

```json
{ "error": { "code": 32, "message": "Some init failure message" } }
```

In case of any error during the initiation step will not be retried and process terminated.

### Stage 2: Transfers (Version 2)

For protocol version 2 transfer stages are defined according to the selected transfer concurrency mode.
There are two concurrency modes available for protocol version 2: basic and batch.

#### Batch Protocol

After the initiation exchange, git-lfs will send any number of batch definitions
to the stdin of the transfer process. Multiple batch transfers can be processed
concurrently when the `concurrent` setting is enabled, allowing for efficient
parallel processing of different batches across multiple worker processes.

The batch definition consists of a header, items, and a footer. Each batch
definition is sent as a series of JSON messages, each on a single line.

**Important**: If the number of remaining files is less than the configured
`lfs.transfer.batchSize`, git-lfs will automatically flush the pending transfers after a
100ms timeout to ensure timely processing of incomplete batches.


##### 2.1 Batch Transfer Header

The first message in a batch transfer is always the batch header, which defines the
batch transfer general batch data. The header defines the batch ID and the number of items
in following batch transfer. 

It looks like this:

```json
{ "event": "batch-header", "bid": "batch-1", "totalSize": 1024, "objectsCount": 5 }
```

Fields:
* `event`: always `batch-header` to identify this message
* `bid`: unique identifier for this batch (persistent throughout the batch transfer)
* `size`: total size in bytes of all files in the batch
* `objectsCount`: number of objects that will follow in this batch

##### 2.2 Batch Transfer Item

Next git-lfs will send the individual items that are part of the batch transfer.
The number of items is defined in the batch header, and each item is sent as a
separate JSON message. Each item can be an upload or a download, depending on
the operation defined in the initiation message.

###### Uploads

For uploads the message sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "upload", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "bid": "batch-1", "size": 1024, "path": "/path/to/file.png", "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `upload` to identify this message
* `oid`: the identifier of the LFS object
* `bid`: the identifier of the batch as defined in the `batch-header`
* `size`: the size of the LFS object
* `path`: the file which the transfer process should read the upload data from
* `action`: the `upload` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

###### Downloads

For downloads the message sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "download", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "size": 1024, "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `download` to identify this message
* `oid`: the identifier of the LFS object
* `bid`: the identifier of the batch as defined in the `batch-header`
* `size`: the size of the LFS object
* `action`: the `download` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

Note there is no file path included in the download request; the transfer
process should create a file itself and return the path in the final response
after completion (see below).

##### 2.3 Batch Footer Message

Last git-lfs sends a batch footer message to indicate the end of the batch
transfer definition. This message signals that no more items will be sent for
this batch, and it provides the total size of all items in the batch.
It looks like this:

```json
{ "event": "batch-footer", "bid": "batch-1", "totalSize": 1024, "objectsCount": 5 }
```

Fields:
* `event`: always `batch-footer` to identify this message
* `bid`: same batch ID as used in `batch-header` to maintain context
* `size`: total size in bytes of all files in the batch (should match the `size` in `batch-header`)
* `objectsCount`: total number of objects in the batch (should match the `objectsCount` in `batch-header`)

##### 2.4 Processing Strategies

Once a batch transfer definition is complete (after receiving the batch-footer), 
the transfer adapter has complete freedom in how it processes the items within 
the batch. Common strategies include:

1. **Sequential Processing**: Process items one by one in the order received
2. **Parallel Processing**: Process multiple items simultaneously using threads/workers
3. **Batch Operations**: Download/upload all items as a single archive or package
4. **Hybrid Approaches**: Combine strategies based on file sizes, types, or other criteria

The protocol places no restrictions on processing order - completion messages
can be sent in any order as items finish processing. 

##### Progress Events

After the batch transfer definition is complete, the transfer process can start
processing the items. 

The transfer process should post one or more progress events to indicate the
batch transfer progress.
It looks like this:

```json
{ "event": "progress", "bid": "batch-1", "bytesSoFar": 256, "bytesSinceLast": 122 }
```

Fields:
* `event`: always `progress` to identify this message
* `bid`: the batch transfer identifier
* `bytesSoFar`: total bytes transferred for this batch so far
* `bytesSinceLast`: bytes transferred since the last progress event

##### Response Messages

###### Individual Item Completion

When an individual item within a batch transfer completes, the transfer
process should send a completion message for that item. 

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "path": "/tmp/downloaded/file.bin" }
```

Fields:
* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `bid`: the identifier of the batch as defined in the `batch-header`
* `path`: for downloads, the temporary path where the file was downloaded

Or, if there was a failure transferring this item:

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "error": { "code": 2, "message": "Explain what happened to this transfer", "retry": true } }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `bid`: the identifier of the batch as defined in the `batch-header`
* `error`: should contain a `code` and `message` explaining the error and 
  optional `retry` flag indicating whether this object should be retried. 
  If `retry` is not specified it is assumed to be false.

These messages have to be sent for each item in the batch transfer, and they can 
be sent in any order. The transfer adapter has complete freedom in how it 
processes the items within a batch - they can be processed sequentially, in 
parallel, or using any other strategy (e.g., downloading all files as a single 
archive and then extracting them). Git-LFS will track completion of individual 
items regardless of the order in which completion messages arrive.

Errors for a single transfer request should not terminate the process or batch 
transfer. The error should be returned in the response structure instead.

`retry` flag indicates whether this batch item can be retried later and does not
have any effect on the batch result. This flag is optional and if not specified
it is assumed to be false.

The custom transfer adapter does not need to check the SHA of the file content
it has downloaded, git-lfs will do that before moving the final content into
the LFS store.

###### Batch Completion

After all the item completion messages are sent, the entire batch transfer completion
message follows. This message indicates that the entire batch transfer is completed.

For successful completion, it looks like this:

```json
{ "event": "batch-complete", "bid": "batch-1" }
```

Fields:
* `event`: always `batch-complete` to identify this message
* `bid`: the identifier of the batch as defined in the `batch-header`

If there was an error during the batch transfer, it can be sent as follows:

```json
{ "event": "batch-complete", "bid": "batch-1", "error": { "code": 500, "message": "Batch transfer failed due to network error", "retry": true } }
```

Fields:
* `event`: always `batch-complete` to identify this message
* `bid`: the identifier of the batch as defined in the `batch-header`
* `error`: should contain a `code` and `message` explaining the error and 
  optional `retry` flag indicating whether this batch should be retried. 
  If `retry` is not specified it is assumed to be false.

`retry` flag in case of batches corresponds to retrying the whole batch regardless of any previous 
successful results for individual artifacts. A batch can be reported as successful though
some artifacts can be retried later via the retry flag in the batch item completion messages. 

The batch transfer completion error never precedes the item completion messages, and it
is sent only once after all items have been processed.

The batch transfer failure is not considered a fatal error, and the process can continue 
to handle other batches or items.


#### Basic Protocol

After the initiation exchange, git-lfs will send any number of transfer
requests to the stdin of the transfer process, in a serial sequence. Once a
transfer request is sent to the process, it awaits a completion response before
sending the next request.

##### Uploads

For uploads the request sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "upload", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "size": 346232, "path": "/path/to/file.png", "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `upload` to identify this message
* `oid`: the identifier of the LFS object
* `size`: the size of the LFS object
* `path`: the file which the transfer process should read the upload data from
* `action`: the `upload` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

The transfer process should post one or more [progress messages](#progress) and
then a final completion message as follows:

```json
{ "event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a" }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object

Or if there was an error in the transfer:

```json
{ "event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "error": { "code": 2, "message": "Explain what happened to this transfer", "retry": true } }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `error`: should contain a `code` and `message` explaining the error and 
  optional `retry` flag indicating whether this object should be retried. 
  If `retry` is not specified it is assumed to be false.

In case of an error the retry flag indicates whether this transfer should be retried or abort any further transfers.

##### Downloads

For downloads the request sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "download", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "size": 21245, "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `download` to identify this message
* `oid`: the identifier of the LFS object
* `size`: the size of the LFS object
* `action`: the `download` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

Note there is no file path included in the download request; the transfer
process should create a file itself and return the path in the final response
after completion (see below).

The transfer process should post one or more [progress messages](#progress) and
then a final completion message as follows:

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "path": "/path/to/file.png" }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `path`: the path to a file containing the downloaded data, which the transfer
  process relinquishes control of to git-lfs. git-lfs will move the file into
  LFS storage.

Or, if there was a failure transferring this item:

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "error": { "code": 2, "message": "Explain what happened to this transfer", "retry": true } }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `error`: should contain a `code` and `message` explaining the error and 
  optional `retry` flag indicating whether this object should be retried. 
  If `retry` is not specified it is assumed to be false.

In case of an error the retry flag indicates whether this transfer should be retried 
or abort any further transfers.

Errors for a single transfer request should not terminate the process. The error
should be returned in the response structure instead.

The custom transfer adapter does not need to check the SHA of the file content
it has downloaded, git-lfs will do that before moving the final content into
the LFS store.

##### Progress

In order to support progress reporting while data is uploading / downloading,
the transfer process should post messages to stdout as follows before sending
the final completion message:

```json
{ "event": "progress", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bytesSoFar": 1234, "bytesSinceLast": 64 }
```

* `event`: always `progress` to identify this message
* `oid`: the identifier of the LFS object
* `bytesSoFar`: the total number of bytes transferred so far
* `bytesSinceLast`: the number of bytes transferred since the last progress
  message

The transfer process should post these messages such that the last one sent
has `bytesSoFar` equal to the file size on success.

##### Response Messages

### Stage 2: Transfers (Version 1)

After the initiation exchange, git-lfs will send any number of transfer
requests to the stdin of the transfer process, in a serial sequence. Once a
transfer request is sent to the process, it awaits a completion response before
sending the next request.

##### Uploads

For uploads the request sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "upload", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "size": 346232, "path": "/path/to/file.png", "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `upload` to identify this message
* `oid`: the identifier of the LFS object
* `size`: the size of the LFS object
* `path`: the file which the transfer process should read the upload data from
* `action`: the `upload` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

The transfer process should post one or more [progress messages](#progress) and
then a final completion message as follows:

```json
{ "event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a" }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object

Or if there was an error in the transfer:

```json
{ "event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "error": { "code": 2, "message": "Explain what happened to this transfer" } }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `error`: should contain a `code` and `message` explaining the error

##### Downloads

For downloads the request sent from git-lfs to the transfer process will look
like this:

```json
{ "event": "download", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "size": 21245, "action": { "href": "nfs://server/path", "header": { "key": "value" } } }
```

* `event`: always `download` to identify this message
* `oid`: the identifier of the LFS object
* `size`: the size of the LFS object
* `action`: the `download` action copied from the response from the batch API.
  This contains `href` and `header` contents, which are named per HTTP
  conventions, but can be interpreted however the custom transfer agent wishes
  (this is an NFS example, but it doesn't even have to be an URL). Generally,
  `href` will give the primary connection details, with `header` containing any
  miscellaneous information needed.  `action` is `null` for standalone transfer
  agents.

Note there is no file path included in the download request; the transfer
process should create a file itself and return the path in the final response
after completion (see below).

The transfer process should post one or more [progress messages](#progress) and
then a final completion message as follows:

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "path": "/path/to/file.png" }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `path`: the path to a file containing the downloaded data, which the transfer
  process relinquishes control of to git-lfs. git-lfs will move the file into
  LFS storage.

Or, if there was a failure transferring this item:

```json
{ "event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "error": { "code": 2, "message": "Explain what happened to this transfer" } }
```

* `event`: always `complete` to identify this message
* `oid`: the identifier of the LFS object
* `error`: should contain a `code` and `message` explaining the error

Errors for a single transfer request should not terminate the process. The error
should be returned in the response structure instead.

The custom transfer adapter does not need to check the SHA of the file content
it has downloaded, git-lfs will do that before moving the final content into
the LFS store.

##### Progress

In order to support progress reporting while data is uploading / downloading,
the transfer process should post messages to stdout as follows before sending
the final completion message:

```json
{ "event": "progress", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bytesSoFar": 1234, "bytesSinceLast": 64 }
```

* `event`: always `progress` to identify this message
* `oid`: the identifier of the LFS object
* `bytesSoFar`: the total number of bytes transferred so far
* `bytesSinceLast`: the number of bytes transferred since the last progress
  message

The transfer process should post these messages such that the last one sent
has `bytesSoFar` equal to the file size on success.

### Stage 3: Finish & Cleanup

When all transfers have been processed, git-lfs will send the following message
to the stdin of the transfer process:

```json
{ "event": "terminate" }
```

On receiving this message the transfer process should clean up and terminate.
No response is expected.

## Protocol Flow Examples

Here's an example flow for downloading 2 files in a batch:

### Client to Process:
```json
{"event": "init", "operation": "download", "remote": "origin", "concurrent": true, "concurrenttransfers": 3}
{"event": "batch-header", "bid": "batch-1", "totalSize": 3072, "objectsCount": 2}
{"event": "download", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "size": 1024, "path": "", "action": {"href": "https://example.com/file1"}}
{"event": "download", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "bid": "batch-1", "size": 2048, "path": "", "action": {"href": "https://example.com/file2"}}
{"event": "batch-footer", "bid": "batch-1", "totalSize": 3072, "objectsCount": 2}
```

### Process to Client (Sequential Processing):
```json
{ }
{"event": "progress", "bid": "batch-1", "bytesSoFar": 0, "bytesSinceLast": 0}
{"event": "progress", "bid": "batch-1", "bytesSoFar": 1024, "bytesSinceLast": 1024}
{"event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "path": "/tmp/file1" }
{"event": "progress", "bid": "batch-1", "bytesSoFar": 3072, "bytesSinceLast": 2048}
{"event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "bid": "batch-1", "path": "/tmp/file2" }
{"event": "batch-complete", "bid": "batch-1" }
```

### Process to Client (Parallel Processing):
```json
{ }
{"event": "progress", "bid": "batch-1", "bytesSoFar": 0, "bytesSinceLast": 0}
{"event": "progress", "bid": "batch-1", "bytesSoFar": 2048, "bytesSinceLast": 2048}
{"event": "complete", "oid": "bf3e3e2af9366a3b704ae0c31de5afa64193ebabffde2091936ad2e7510bc03a", "bid": "batch-1", "path": "/tmp/file2" }
{"event": "progress", "bid": "batch-1", "bytesSoFar": 3072, "bytesSinceLast": 1024}
{"event": "complete", "oid": "22ab5f63670800cc7be06dbed816012b0dc411e774754c7579467d2536a9cf3e", "bid": "batch-1", "path": "/tmp/file1" }
{"event": "batch-complete", "bid": "batch-1" }
```

Note: In the parallel example, file2 completes before file1, demonstrating that 
completion order is independent of the order in which items were sent.

### Cleanup:
```json
{"event": "terminate"}
```


## Error handling

Any unexpected fatal errors in the transfer process (not errors specific to a
transfer request) should set the exit code to non-zero and print information to
stderr. Otherwise the exit code should be 0 even if some transfers failed.

## A Note On Verify Actions

You may have noticed that that only the `upload` and `download` actions are
passed to the custom transfer agent for processing, what about the `verify`
action, if the API returns one?

Custom transfer agents do not handle the verification process, only the
upload and download of content. The verify link is typically used to notify
a system *other* than the actual content store after an upload was completed,
therefore it makes more sense for that to be handled via the normal API process.
