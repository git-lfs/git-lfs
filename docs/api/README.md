# Git LFS API

The Git LFS client uses an HTTPS server to coordinate fetching and storing
large binary objects separately from a Git server. The basic process the client
goes through looks like this:

1. [Discover the LFS Server to use](./server-discovery.md).
2. [Apply Authentication](./authentication.md).
3. [Request the Batch API](./batch.md) to upload or download objects.
4. The Batch API's response dictates how the client will transfer the objects.

Current transfer adapters include:
  * [Basic](./basic-transfers.md)

Experimental transfer adapters include:
  * Tus.io (upload only)
  * [Custom](../custom-transfers.md)
