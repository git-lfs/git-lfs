# Git LFS & SSH

Git LFS now has 2 ways of handling using SSH git remote URLs:

1. redirect to a HTTPS API URL via a pre-call to `ssh user@host git-lfs-
authenticate` 
2. a pure SSH API implementation of the API (plus reference server)

Option 1 continues to work if you've chosen to use it, but for people who would
prefer LFS to consistently use SSH for the upload/download of large files too,
there is now an alternative.

## The basics

When asked to upload or download large file data with an SSH URL (if co-located, the git clone URL can be used without having to specify lfs.url in gitconfig), a connection is first established with the server and we attempt to invoke the full server implementation - by default this is called `git-lfs-serve` but you can call something else by setting `lfs.sshservercmd` in gitconfig.

If this full SSH server command is not available then this will fail, and we
will fall back on the old way of using HTTPS instead with `git-lfs-
authenticate`. If it succeeds, this connection is retained as a context and used for all subsequent uploads and downloads until termination of the current
command.

The only argument passed to the `git-lfs-serve` command is the relative path of
the repo, as extracted from the URL. For more information on installing the
reference server & configuring it, see [git-lfs-serve](git-lfs-serve.md)

Authentication is handled in the same way as opening any SSH connection,
primarily through SSH keys. Manual authentication may be supported via `git
credentials` in future.

## The protocol

As mentioned above a single SSH connection is used for multiple operations,
avoiding the need to incur the overhead of negotiating afresh on each request.
Communication simply occurs over the stdin/stdout of the processes at either end of the connection.

The requests and human-readable responses are in JSON format, roughly anagolous
to JSON-RPC except that we allow raw byte streams to be interleaved with the
data for transferring large file content. If we'd limited the protocol to
strictly JSON-RPC the binary data would have had to be text encoded (e.g.
base64) which would have made everything a lot slower.

Because of this, all JSON request and response structures are terminated with a
binary 0 in the stream to indicate termination of the JSON; this allows
efficient reading of variable-length data within a persistent re-usable stream.

## The request / response structures

Note that the use of 'id' is mostly for familiarity with JSON-RPC; at this point parallel requests aren't possible (within a single connection), but later on this might come in handy if the server back end uses a message queue or something to aggregate the work of many connections.

| **Request** || 
|---------|-----------------------------------------------| 
| id      | Numerical identifier of this request          | 
| method  | The method name to be called                | 
| params  | Nested structure of method-specific parameters|

| **Response** ||
|---------|---------------------------------------------------------| 
| id | Numerical identifier of the request we're responding to | 
| error   | Should only be present if an error occurred             | 
| result  | Nested structure of method-specific result data         |

Each method has different late-bound JSON nested structures underneath 'params'
and 'result' to afford maximum flexibility to the API.

## Currently implemented methods

In the specifications below we only specify the contents of the 'params' and
'result' nested structures, but the surrounding request / response structures
are always there.

Where a byte stream of data is either uploaded or downloaded, it is marked with
STREAM_UP and STREAM_DOWN. In some cases there may be a response to the stream,
which is marked with 'STREAM Response' and is the same format as other response
objects.

|**DownloadInfo**|| 
|-----------|-------------| 
|Purpose    | Get the size of a large file (if it exists), ready to download |
|Params     | oid (string): the SHA of the large file| 
|Result     | size (number): size in bytes (or zero and an error if it doesn't exist)|

|**Download**|| 
|-----------|-------------| 
|Purpose    | Retrieve the content of a large file as a binary stream | 
|Params     | oid (string): the SHA of the large file| 
|           | size (number): the expected size of the stream (server should report an error if this is wrong)| |STREAM_DOWN| Response is a binary stream of exactly 'size' bytes. Client must read all bytes.|

|**Upload**|| 
|-----------|-------------| 
|Purpose    | Upload the content of a large file | 
|Params     | oid (string): the SHA of the large file| 
|           | size (number): size in bytes of the data to upload| 
|Result     | okToSend (bool): True if the server is ready to receive on this basis| 
|STREAM_UP  | If server accepted, client should send a stream of exactly size bytes| 
|STREAM Result| receivedOk: Indicates server received & stored the data successfully|

|**Exit**|| 
|-----------|-------------| 
|Purpose    | Exit the server-side process gracefully| 
|Params     |None| 
|Result     |None|

## Methods to be added later 
(These methods are some of those I implemented in my LFS system and will be useful later)

|**PickFirst**|| 
|-----------|-------------| 
|Purpose |Out of a list of LOB SHAs in order of preference, return which one (if any) the server has a complete copy of already. This is used to probe for previous versions of a file to exchange a binary delta of. Note that in all cases (upload and download) the client is responsible for creating the list of possible ancestor candidates, whether sending or receiving. This means the server doesn't have to have the git repo available, and the client always has the git commits when downloading anyway (that's how it decides what to download)| 
|Params  |oids: array of strings identifying LOBs in order of preference (usually ancestors of a file)| 
|Result  |picked: first sha in the list that server has a complete file copy of, or blank string if none are present. The server should confirm that all data is present but does not need to check the sha integrity (done post delta application)|

|**UploadDelta**|| 
|-----------|-------------| 
|Purpose    | Ask to upload a binary patch between 2 lobs which the client has calculated so the server can apply it to its own store, without uploading the entire file content. This is only about the chunk content; metadata is uploaded with **Upload** as usual and should be done before calling this method.| 
|Params     | baseOid (string): the SHA of the binary file content to use as a base. Client should have already identified that server has this via **PickFirst**|
|           | targetOid (string): the SHA of the binary file content we want to reconstruct from base + delta| 
|           | size (Number): size in bytes of the binary delta| 
|Result     | okToSend: True if server is ready to receive delta on this basis| 
|STREAM_UP  | Immediately after okToSend, a binary stream of bytes will be sent by the client to the server of length 'size' above. The server must read all the bytes and then generate the final file from the delta + base (must check SHA integrity) and store it.| 
|STREAM_UP Result |receivedOk: True if server received all the bytes and stored the file successfully. On failure, return Error.|

|**DownloadDeltaInfo**|| 
|-----------|-------------| 
|Purpose    | Ask the server to generate a binary patch between 2 lobs which we know it has (or re-usean existing delta). This is only about the chunk content; metadata is downloaded the usual way.| 
|Params     | baseOid (string): the SHA of the binary file content to use as a base| 
|           | targetOid (string): the SHA of the binary file content we want to reconstruct from base + delta| 
|Result     | size (Number): size in bytes of delta if server has generated it ready to to send (Error otherwise). Server should keep this calculated delta for a while, at least 1 day (maybe longer to re-use for multiple clients). 0 if there was a problem (error identifies). The client should subsequently request the calculated delta if it wants it (may choose not to if borderline)| 
|           | Client should follow up with a call to **DownloadDeltaContent** to trigger the binary data send, which includes all the same params|

|**DownloadDeltaContent**|| 
|-----------|-------------| 
|Purpose    | Begin downloading a LOB delta file to apply locally against a base LOB to generate new content. Metadata is not included, that's downloaded the usual way| 
|Params     | baseOid (string): the SHA of the binary file content to use as a base| 
|           | targetOid (string): the SHA of the binary file content we want to reconstruct from base + delta| 
|           | size (Number): size in bytes of delta as reported from **DownloadDeltaInfo**.|  
|STREAM_DOWN| Response is a binary stream of data of exactly size bytes. Client must read all the bytes and use to apply to base LOB to create new content.|

