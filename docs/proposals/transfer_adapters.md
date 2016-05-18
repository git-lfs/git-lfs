# Transfer adapters for resumable upload / download

## Concept

To allow the uploading and downloading of LFS content to be implemented in more
ways than the current simple HTTP GET/PUT approach. Features that could be
supported by opening this up to other protocols might include:

  - Resumable transfers
  - Block-level de-duplication
  - Delegation to 3rd party services like Dropbox / Google Drive / OneDrive
  - Non-HTTP services

## API extensions

See the [API documentation](../http-v1-batch.md) for specifics. All changes
are optional extras so there are no breaking changes to the API.

The current HTTP GET/PUT system will remain the default. When a version of the
git-lfs client supports alternative transfer mechanisms, it notifies the server
in the API request using the `accept-transfers` field. 

If the server also supports one of the mechanisms the client advertised, it may 
select one and alter the upload / download URLs to point at resources 
compatible with this transfer mechanism. It must also indicate the chosen 
transfer mechanism in the response using the `transfer` field. 

The URLs provided in this case may not be HTTP, they may be custom protocols.
It is up to each individual transfer mechanism to define how URLs are used.

## Client extensions

### Phase 1: refactoring & abstraction

1. Introduce a new concept of 'transfer adapter'. 
2. Adapters can provide either upload or download support, or both. This is 
   necessary because some mechanisms are unidirectional, e.g. HTTP Content-Range
   is download only, tus.io is upload only.
3. Refactor our current HTTP GET/PUT mechanism to be the default implementation 
   for both upload & download
4. The LFS core will pass oids to transfer to this adapter in bulk, and receive 
   events back from the adapter for transfer progress, and file completion.
5. Each adapter is responsible for its own parallelism, but should respect the
   `lfs.concurrenttransfers` setting. For example the default (current) approach
   will parallelise on files (oids), but others may parallelise in other ways
   e.g. downloading multiple parts of the same file at once
6. Each adapter should store its own temporary files. On file completion it must
   notify the core which in the case of a download is then responsible for 
   moving a completed file into permanent storage.
7. Update the core to have a registry of available transfer mechanisms which it
   passes to the API, and can recognise a chosen one in the response. Default
   to our refactored original.

### Phase 2: basic resumable downloads

1. Add a client transfer adapter for [HTTP Range headers](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.35)
2. Add a range request reference implementation to our integration test server

### Phase 3: basic resumable uploads

1. Add a client transfer adapter for [tus.io](http://tus.io) (upload only)
2. Add a tus.io reference implementation to our integration test server

### Phase 4: external transfer adapters

Ideally we should allow people to add other transfer implementations so that
we don't have to implement everything, or bloat the git-lfs binary with every
custom system possible.

Because Go is statically linked it's not possible to extend client functionality
at runtime through loading libaries, so instead I propose allowing an external
process to be invoked, and communicated with via a defined stream protocol. This
protocol will be logically identical to the internal adapters; the core passing
oids and receiving back progress and completion notifications; just that the 
implementation will be in an external process and the messages will be 
serialised over streams.

Only one process will be launched and will remain for the entire period of all
transfers. Like internal adapters, the external process will be responsible for
its own parallelism and temporary storage, so internally they can (should) do
multiple transfers at once.

1. Build a generic 'external' adapter which can invoke a named process and
   communicate with it using the standard stream protocol (probably just over
   stdout / stdin)
2. Establish a configuration for external adapters; minimum is an identifier 
   (client and server must agree on what that is) and a path to invoke
3. Implement a small test process in Go which simply wraps the default HTTP
   mechanism in an external process, to prove the approach (not in release)


