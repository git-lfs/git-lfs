# Bulk transfers proposal 

I'd like to extend the custom transfers capabilities in Git Lfs.

Lets extend the way Git Lfs allows custom transfer agents to handle the 
artifact transfers from single artifact per process to group of artifacts
per process leaving all the implementation details to the custom transfer agent.

## Motivation

Currently custom transfer agent is responsible for transferring single artifact
between client and server, Git Lfs is responsible for artifact set discovery
and scheduling the transfers.
Transfers are scheduled one by one in a linear manner and are processed in a
concurrency model "one process per one artifact transfer".
Although this implementation is simple and performs well for a whole lot of scenarios,
it has some limitations:
- Minimal file size limitation.
  Transferring one artifact per process makes little sense for files smaller than
  some magic number due to the overhead of process creation and management, network overheads
  and natural process count limits per system. The exact mileage may vary, but each
  concrete setup will have a magic number that defines the lower bound of file size
  applicable effectively. Going below this magic number will lead to poor overall
  operation times: download/upload times per artifact will stop lowering along with
  artifact sizes and total pull/push operation times will go up.
- Single file data limitation.
  Transferring artifacts one by one limits the opportunity to optimize network traffic
  size by compressing the artifacts or using some binary protocols that are more efficient
  when transferring multiple artifacts in a single session. This is especially true for
  large text files that can be compressed well and the compression ratio improves with the
  size of the data. This might seem as a counterintuitive statement in terms of git, but
  a lot of modern game engines for example use text based assets (json, yaml, xml, etc)
  that can reach hundreds of MBs in size when uncompressed. For big project the amount of
  such files can be in thousands and they typically have no distinct patterns that allow
  managing them on the configuration level.

### How does the proposal adress the limitations above?

- Minimal file size limitation.
  By allowing custom transfer agents to handle multiple artifacts per process
  we allow them to implement their own logic of grouping artifacts into batches
  that make sense for their specific use cases. This way the custom transfer agent
  can decide how many artifacts it wants to handle per process based on the artifact
  sizes and other factors that are relevant for its specific implementation.
- Single file data limitation.
  By allowing custom transfer agents to handle multiple artifacts per process
  we enable them to optimize the transfer of related artifacts together, potentially
  using more efficient data formats or compression techniques that take advantage
  of the similarities between the artifacts. This can lead to better overall
  transfer speeds and reduced bandwidth usage, especially for large text files
  that can be compressed effectively.

## Terms

The best term to describe the proposed feature would be "batching".
Though this term is already taken in Git Lfs context to describe the way
Git Lfs client and server communicate about the set of artifacts to be transferred.
So to keep the terminology consistent and avoid confusion, we will use the term
"bulking" to describe the proposed feature of transferring multiple artifacts
together in a single process.

## Changes to the API

To support this proposal, we need to make some changes to the Git Lfs API.

1. Extend the custom transfers protocol.
   This includes defining new request and response formats that support 
   multiple artifacts per transfer operation and handling the results of
   transferring of single artifacts and the whole bulks.
2. Update the Git Lfs client to support the new bulk transfer capabilities.
   This includes implementing the logic for planning bulk transfers, scheduling
   them, and processing the results.
3. Error handling, retries and protocol fallbacks.
   This includes defining how the client and server should handle errors that
   occur during bulk transfers, and how to fall back to single artifact transfers
   in case of failures. This also includes implementing retry logic for failed
   artifact or bulk transfers.
4. Documentation and examples.
   This includes updating the Git Lfs documentation to include information about
   the new bulk transfer capabilities and how to use them effectively.

## Further reading

To demonstrate the proposed changes I modified the existing docs/custom-transfers.md
document to include the bulk transfer capabilities.

The important parts are the new request and response formats for the custom transfer
operations and the retry and fallback mechanisms described.

The main changes are:
* introducing the `mode` config opption to enable/disable bulking
* bulk protocol flow and request/response formats
* retry support for single artifact transfers and bulks, which is also included
  into the basic custom transfers protocol
* fallback support from bulking to single artifact transfers