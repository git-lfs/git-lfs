# SSH protocol proposal

We'd like to implement a protocol for Git LFS that uses SSH protocol
exclusively, avoiding the need to use HTTPS altogether.  This will make
deployment and use easier in a variety of situations where access to certain
ports is limited.

This is merely a proposal, not a commitment to implement for either the client
or server side.  Implementers who prefer to use HTTP can continue to do so.

## What not to do

There are several possible approaches that could be adopted.  SSH provides a
native capability for the SFTP protocol, which can be used to transfer files.
However, in order to implement this on the server side, each access (upload or
download) must have an access control check instead of one at the beginning of
the operation.  This would be inefficient in some server-side implementations,
and nearly impossible to implement securely for implementations that use the
system OpenSSH for implementation.

## A more usable approach

Git already has some places we can look for inspiration.  Its SSH protocol is
based on the Git native protocol, which is based on the pkt-line scheme.
Recently, Git has learned about protocol version 2, which provides better
support for expressing and negotiating capabilities.

Ideally, we would allow multiple operations to occur on a single connection for
efficiency's sake, especially on high-latency connections, where the cost of SSH
connection setup may be high due to multiple round trips.  In addition, a
protocol which maps well onto HTTP may be beneficial for those server-side
implementations which would like to proxy connections to an HTTP-based backend.

## Preliminary design

This design assumes a reference to Git's pkt-line and protocol v2 documentation.
pkt-line headers for this document may contain values up to 65519 decimal.

To initiate a connection, Git LFS should run the following command:

    $ ssh [{user}@]{server} git-lfs-transfer {path} {operation}

If authentication fails, or some other connection error occurs, errors will be
read from standard error and displayed to the user.  The operation may be
`upload`, `download`, or `lock`.  Other operations may be implemented in the
future.

Once the connection is established, the server should send a capability
advertisement:

```
capability-advertisement = capability-list flush-pkt
capability-list = *capability
capability = PKT-LINE(key[=value] LF)

key = 1*(ALPHA | DIGIT | "-_")
value = 1*(ALPHA | DIGIT | "-_.,?\/{}[]()<>!@#$%^&*+=:;")
```

Unlike the Git protocol, but like IMAP, the protocol version is specified as a
capability.  This document defines protocol version 1, which is specified as
`version=1`.  If the server supports other protocol versions, it may enumerate
them here as well.

If the server supports locking, the `locking` capability should be advertised,
and the client may then use the `lock`, `unlock`, and `list-lock` commands.

No capabilities other than the base functionality specified here are enabled
without the client explicitly enabling them.  Note that the `value` production
here, unlike in Git, does not include the space character, since it is used as a
delimiter in parts of the protocol.

The client will then issue an appropriate version command:

```
version-request = PKT-LINE("version " number LF) flush-pkt
number = 1*DIGIT
```

The response from the server will look like the following:

```
version-response = status-command
                   delim-pkt
                   error-message
                   flush-pkt
status-command = PKT-LINE("status " http-status-code LF)
http-status-code = 3DIGIT
error-message = *PKT-LINE(data LF)
```

The `http-status-code` portion of the response is an HTTP status code, identical
to those used if the request is made over HTTP.

The response code should be 200 if the version is accepted or 400 if it is not.
Other values are possible if other errors occur.

### Requests to transfer objects

These commands may be used if the operation was `upload` or `download`.

The `batch` command is used to specify a JSON command identical those used at
the `info/lfs/object/batch` endpoint:

```
batch-request = batch-command
                *argument
                delim-pkt
                *oid-line
                flush-pkt
batch-command = PKT-LINE("batch" LF)
argument = PKT-LINE(key=[data] LF)
oid-line = PKT-LINE(oid size *(key=[value]) LF)
oid = 1*("a-f0-9")
size = 1*DIGIT
```

The `transfer` argument is equivalent to the corresponding value in the HTTP
JSON API.  The `refname` argument is equivalent to the `name` argument of the
`ref` object in the HTTP JSON API.

Unknown arguments should be ignored, as should unknown key-value pairs in the
`oid-line` production.

The response from the server will look like the following:

```
batch-response = status-command
                 *argument
                 delim-pkt
                 (*batch-oid-line | error-message)
                 flush-pkt
batch-oid-line = PKT-LINE(oid size action *(key=[value]) LF)
error-message = PKT-LINE("message " data LF)
```

If the status command is successful (that is, the status is not 200-series
response), the data provided matches the `*batch-oid-line` production;
otherwise, the data provided represents a user-visible error message.

The server response should contain one pkt-line per oid-size-action tuple.  That
is, the same oid and size may be repeated if there are multiple actions.  If the
server has no actions that are valid for an object, it should be listed once in
the response with the `noop` action.

The response for an oid may include a string, `id`, which is an opaque
identifier relevant only to the server to help it identify the object, and
another string, `token`, which is an opaque identifier relevant only to the
server to help it manage authentication.  These strings must meet the syntax for
the `value` production above; if arbitrary bytes are needed, Base64 encoding is
recommended.

The `expires-in` and `expires-at` key-value pairs have the same meaning as their
corresponding items from the HTTP JSON API.

These values, if specified, must be passed as arguments to the `get-object` and
`put-object` commands.

### Downloads

These commands may be used if the operation was `download`.

If the operation was `download`, the command `get-object` may be used to retrieve an object:

```
get-object-request = get-object-command
                     *arguments
                     flush-pkt
get-object-command = PKT-LINE("get-object " oid LF)

```

If the `id` or `token` responses were provided in the response to `batch`, they
must be specified as arguments here.  The server may choose to ignore the `oid`
field specified in favor of looking up the data using the `id` field.

The response looks like the following:

```
status-data-response = status-success-response | status-error-response
status-success-response = status-success-command
                          *argument
                          delim-pkt
                          binary-data
                          flush-pkt
status-success-command = PKT-LINE("status 200" LF)
binary-data = *PKT-LINE(data)
status-error-response = status-error-response
                        *argument
                        delim-pkt
                        error-message
                        flush-pkt
status-error-command = PKT-LINE("status " http-error-code LF)
http-error-code = ("4" | "5") 2DIGIT
```

The `size` argument is required on `status` responses to `get-object` commands.

### Uploads

These commands may be used if the operation was `upload`.

If the operation was `upload`, the commands `put-object` and `verify-object`
may be specified.  `put-object` is used to upload an object to the server:

```
put-object-request = put-object-command
                     *argument
                     delim-pkt
                     binary-data
                     flush-pkt
put-object-command = PKT-LINE("put-object " oid LF)
```

As above, the `size` command is required and `id` and `token` are required if
provided by the server.  The response matches the `status-data-response`
production.  The `binary-data` returned on success is not meaningful and should
be empty.

The `verify-object` command is used to verify an object:

```
verify-object-request = verify-object-command
                        *argument
                        flush-pkt
verify-object-command = PKT-LINE("verify-object " oid " " size LF)
```

The `size` production represents the size of the object represented by `oid` as
a decimal integer in bytes.  The `id` and `token` items from the batch request
must be passed as arguments here, if specified.

The response matches the following:

```
generic-status-response = generic-success-response | status-error-response
generic-success-response = generic-success-command
                           *argument
                           flush-pkt
generic-success-command = PKT-LINE("status 200" LF)
```

### Locks

These commands may be used if the operation was `lock`.

The `lock` command may be used to lock a file on a ref:
```
lock-request = lock-command
               *argument
               delim-pkt
               flush-pkt
lock-command = PKT-LINE("lock" LF)
```

The `path` and `refname` arguments correspond to the `path` component and the
`name` component of the `ref` object in the HTTP JSON API.

The response is as follows:


```
lock-response = lock-success-response | status-error-response
lock-success-response = lock-success-command
                        *argument
                        flush-pkt
lock-success-command = PKT-LINE("status 201" LF)
```

If the response is either successful or a 409 response, the arguments `id`,
`path`, `locked-at`, and `ownername` are provided.  In case of a successful
response, these attributes represent the created lock; if the response is a 409,
then the attributes represent the conflicting lock.

The `list-lock` command may be used to list and verify locks:

```
list-lock-request = list-lock-command
                    *argument
                    flush-pkt
list-lock-command = PKT-LINE("list-lock" LF)
```

The `path`, `id`, `cursor`, `limit`, and `refspec` correspond to the items in
the HTTP JSON API.

```
list-lock-response = list-lock-success-response | status-error-response
list-lock-success-response = list-lock-success-command
                             *argument
                             delim-pkt
                             *lock-spec
                             flush-pkt
list-lock-success-command = PKT-LINE("status 200" LF)
lock-spec = lock-decl
            path-id
            locked-at
            ownername-id
            owner-id
lock-decl = PKT-LINE("lock " lock-id LF)
lock-id = value
path-id = PKT-LINE("path " lock-id path LF)
path = data
locked-at = PKT-LINE("locked-at " lock-id timestamp LF)
ownername-id = PKT-LINE("ownername " lock-id ownername LF)
ownername = data
owner-id = PKT-LINE("owner " lock-id who LF)
who = ("ours" | "theirs")
```

The `lock-decl` production declares a new lock.  The `lock-id` production refers
to the ID provided by the server.  The same ID is repeated in each line to allow
for easier parsing.

The `next-cursor` argument indicates the next value of the `cursor` argument to
be passed to the `list-lock` command.  If there is no `next-cursor` argument,
this is the final response.

```
unlock-request = unlock-command
                 *argument
                 flush-pkt
unlock-command = PKT-LINE("unlock " lock-id LF)
```

The `force` and `refname` arguments have the same meaning as their corresponding
values in the HTTP JSON API.  The response matches the `generic-status-response`
production.
