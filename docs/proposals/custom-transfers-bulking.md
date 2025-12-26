# Batch artifact transfers concurrency proposal 

I'd like to extend transfers concurrency capabilities in Git Lfs.

Lets extend the way Git Lfs allows transfer agents (including custom transfer agents) 
to handle the artifact transfers from single artifact per process to group of artifacts
per process.

## Motivation

Currently transfer agent is responsible for transferring single artifact between 
client and server, Git Lfs is responsible for artifact set discovery and scheduling the transfers.
Transfers are scheduled one by one in a linear manner and are processed in a
concurrency mode "one process per one artifact transfer". This is especially true
for custom transfer agents, where Git Lfs spawns a new process for each artifact
transfer request.
Although this implementation is simple and performs well for a whole lot of scenarios,
it has some limitations:
- Minimal file size limitation.
  Transferring one artifact per process makes little sense for files smaller than
  some magic number due to the overhead of process creation and management, network overheads
  and natural process count limits per system. The exact mileage may vary, but each
  concrete setup will have a magic number that defines the lower bound of file size
  applicable effectively. Going below this magic number will lead to poor overall
  operation times: download/upload times per artifact will stop decreasing along with
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
  By allowing transfer agents to handle multiple artifacts per process we allow implementing 
  any applicable logic of grouping artifacts into transfer batches that make sense for 
  specific use cases. This way the transfer agent can decide how many artifacts agent wants 
  to handle per process based on the artifact sizes and other factors that are relevant 
  for its specific implementation.
- Single file data limitation.
  By allowing transfer agents to handle multiple artifacts per process we enable them to 
  optimize the transfer of related artifacts together, potentially using more efficient data 
  formats or compression techniques that take advantage of the similarities between the 
  artifacts. This can lead to better overall transfer speeds and reduced bandwidth usage, 
  especially for large text files that can be compressed effectively.

## Terms

The best term to describe the proposed feature would be "artifact batching".
Though this term is already taken in Git Lfs context to describe the way
Git Lfs client and server communicate about the set of artifacts to be transferred.
So to keep the terminology consistent and avoid confusion, we will intoduce the term
"transfer concurrency mode" to describe the proposed feature of transferring multiple 
artifacts together in a single process.
We will define two transfer concurrency modes:
1. Single artifact per process mode - the existing mode where each transfer process
handles a single artifact transfer. This is the default mode and will be used
if no other mode is specified or supported. The short term for this mode will be
"basic" mode.
2. Multiple artifacts per process mode - the new mode where each transfer process
handles multiple artifact transfers. This mode allows transfer agents to optimize 
the transfer of related artifacts together. The short term for this mode will be
"batch" mode. Cause effectively we are transferring batches of artifacts in each
transfer process.

This naming schema shows the ties to the existing gil-lfs terminology and 
configuration parameters, while clearly defining the new feature and its purpose.

## Required changes

To implement the proposed feature, we need to make some changes to the Git Lfs client
and the Git Lfs API.

1. Client configuration for transfer concurrency mode.
We should introduce a new client configuration options to allow users to specify
the desired transfer concurrency mode. This option will allow users to choose the preferred mode
and will further be negotiated with the server during the batch api calls. The server will
confirm the mode to be used for the batch session based on the client preference and the 
server capabilities. User should see a clear warning message if the requested mode
is not supported by the server and the client falls back to the default one.
To maintain backward compatibility, the default mode will be "basic".
For maintaining backward compatibility with existing transfer implementations,
we should also introduce the protocol versioning for custom transfer agents
to allow negotiation of the supported concurrency modes during the initialization phase.
This will also require to introduce the protocol version config option to allow
users to specify the desired protocol version to be used with custom transfer agents.
2. Client and server should agree on the supported transfer concurrency modes. 
We should modify the batch api to include the negotiation about the used transfer concurrency mode.
Each batch request should include information about the supported transfer concurrency modes and define 
the preferred one from the client configuration. Server should confirm the transfer concurrency mode to 
be used for the current session in the batch response.
To maintain backward compatibility, the default mode will be "basic".
3. Custom transfer agents support.
We should modify the custom transfer agents implementation to support the new batch concurrency mode.
This requires defining the new protocol flow and request/response formats to support multiple artifacts 
per transfer operation. Also we should extend the protocol to allow individual and batch error retries.
This is essential in case of new concurrency batch mode to deduce the fallback cases.
To maintain backward compatibility, the protocol flow should include the explicit concurrency mode 
support confirmation for new mode and consider the existing flow as a "basic" mode.

## Implementation details
Some of the proposed changes have an implementation details that should be considered or dicussed further.
As a baseline for this discussion this pull request can be read to see a draft implementation of the proposal:
https://github.com/git-lfs/git-lfs/pull/6119

This pull request is mainly focused on the custom transfer agents support. 
I learned several important things during the implementation that should be considered
when implementing the rest of the proposal.

1. The current version of custom tranfers protocol does not support error retries and always fallbacks in case
of any error reported from the custom transfer agent. 
This effectively means that in case of any retriable error the transfer agent is bound to implement all the retry 
logic on its own. This makes sense in general, but introduce the question: 
when custom agent retries on its own - user is seeing a weird misleading data about the download/upload progress 
cause there is no clear way to report the progress of retries to Git Lfs client. Meanwhile reporting the errors 
means that Git Lfs client will fallback to the default transfer agent which is not desired in any general case.
2. Current transfer queue implementation keeps secret about the transfer queue state from the transfer agents.
This leads to the situation when transfer agent has no idea about the overall transfer queue content and to maintain
the operability it has to rely on assumption: when transfer queue has anything to tranfer - it will be sending the 
request immediately. If this is not the case - we wait for some time and then just flush ourself. 
This should be addresed during the implementation of new concurrency mode to allow transfer agents to explicitly 
know that the current queue is empty and they should flush their state.

## Implementation plan
To implement the proposed feature, we can follow these steps:
1. Describe the batch API changes required to support the concurrency modes negotiation and implement them.
2. Implement the client configuration options to allow users to specify the desired concurrency mode.
3. Describe and implement the changes required to support the new concurrency mode in custom transfer agents.
4. Update the tests and documentation to reflect the changes made to support the new concurrency mode.

## Batch API changes
To support the proposed concurrency mode, we need to make some changes to the batch API.
Git Lfs client should include the supported concurrency modes and the preferred one into
every batch api request. Server should confirm the concurrency mode to be used for the current batch api response.
It seems the most straightforward way to implement this is to add two new fields to the batch api request:
- "supported_transfer_concurrency_modes": array of strings - list of supported transfer concurrency modes by the client.
- "preferred_transfer_concurrency_mode": string - the transfer concurrency mode selected for the current batch session.
As for the response server should only include one new field:
- "transfer_concurrency_mode": string - the transfer concurrency mode selected for the current batch session.
This way we allow extensibility for future transfer concurrency modes and maintain backward compatibility
by considering the absence of these fields as a signal to use the default "basic" mode.
We should also introduce new error cases to indicate that the none of the requested transfer concurrency modes
consigured on the client are supported by the server.
This way client will try to use the configured mode and if server does not support it - client
will get a clear error message and will be able to fallback to the default mode or inform the user
about the misconfiguration.

## Custom transfer protocol changes
To support the proposed concurrency mode in custom transfer agents, we need to make some changes to the custom transfer protocol.

Effectively we should define two different protocol flows:
1. Single artifact per process flow - the existing protocol flow that supports transferring single artifact per process.
2. Batch concurrency mode flow - the new protocol flow that supports transferring multiple artifacts per process.
Also we need to include the error retry mechanism to allow transfer agents to report retriable errors
without forcing Git Lfs client to fallback to the default transfer agent.

Git Lfs and custom transfer agents should negotiate the supported concurrency strategies during the initialization phase of the 
custom transfer protocol. After the negotiation is comlete the protocol version should be fixed for the whole session and both 
sides should follow the agreed concurrency mode.

Basic flow is already defined and implemented. The only requred change here is to implement the optional error retry mechanism 
to allow transfer agents to report retriable errors without forcing Git Lfs client to fallback to the default transfer agent.

Batch concurrency mode flow requires more changes to the protocol.
We need to define the new transfer request format that defines a set of artifacts to be transferred.
Then we need to define the progress tracking mechanism that allows transfer agents to report the progress of each artifact in the batch.
Finally we need to implement the error retry mechanism to allow transfer agents to report retriable errors for individual artifacts in the batch.

The protocol proposition is described in more details in the draft implementation pull request: 
https://github.com/git-lfs/git-lfs/pull/6119

