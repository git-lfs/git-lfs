# Locking feature proposal

We need the ability to lock files to discourage (we can never prevent) parallel
editing of binary files which will result in an unmergeable situation. This is
not a common theme in git (for obvious reasons, it conflicts with its
distributed, parallel nature), but is a requirement of any binary management
system, since files are very often completely unmergeable, and no-one likes
having to throw their work away & do it again.

## What not to do: single branch model

The simplest way to organise locking is to require that if binary files are only
ever edited on a single branch, and therefore editing this file can follow a
simple sequence:

1. File starts out read-only locally
2. User locks the file, user is required to have the latest version locally from
   the 'main' branch
3. User edits file & commits 1 or more times
4. User pushes these commits to the main branch
5. File is unlocked (and made read only locally again)

## A more usable approach: multi-branch model

In practice teams need to work on more than one branch, and sometimes that work
will have corresponding binary edits.

It's important to remember that the core requirement is to prevent *unintended
parallel edits of an unmergeable file*.

One way to address this would be to say that locking a file locks it across all
branches, and that lock is only released when the branch where the edit is is
merged back into a 'primary' branch. The problem is that although that allows
branching and also prevents merge conflicts, it forces merging of feature
branches before a further edit can be made by someone else.

An alternative is that locking a file locks it across all branches, but when the
lock is released, further locks on that file can only be taken on a descendant
of the latest edit that has been made, whichever branch it is on. That means
a change to the rules of the lock sequence, namely:

1. File starts out read-only locally
2. User tries to lock a file. This is only allowed if:
   * The file is not already locked by anyone else, AND
   * One of the following are true:
      * The user has, or agrees to check out, a descendant of the latest commit
        that was made for that file, whatever branch that was on, OR
      * The user stays on their current commit but resets the locked file to the
        state of the latest commit (making it modified locally, and
        also cherry-picking changes for that file in practice).
3. User edits file & commits 1 or more times, on any branch they like
4. User pushes the commits
5. File is unlocked if:
   * the latest commit to that file has been pushed (on any branch), and
   * the file is not locally edited

This means that long-running branches can be maintained but that editing of a
binary file must always incorporate the latest binary edits. This means that if
this system is always respected, there is only ever one linear stream of
development for this binary file, even though that 'thread' may wind its way
across many different branches in the process.

This does mean that no-one's changes are accidentally lost, but it does mean
that we are either making new branches dependent on others, OR we're
cherry-picking changes to individual files across branches. This does change
the traditional git workflow, but importantly it achieves the core requirement
of never *accidentally* losing anyone's changes. How changes are threaded
across branches is always under the user's control.

## Breaking the rules
We must allow the user to break the rules if they know what they are doing.
Locking is there to prevent unintended binary merge conflicts, but sometimes you
might want to intentionally create one, with the full knowledge that you're
going to have to manually merge the result (or more likely, pick one side and
discard the other) later down the line. There are 2 cases of rule breaking to
support:

1. **Break someone else's lock**
  People lock files and forget they've locked them, then go on holiday, or
  worse, leave the company. You can't be stuck not being able to edit that file
  so must be able to forcibly break someone else's lock. Ideally this should
  result in some kind of notification to the original locker (might need to be a
  special value-add on BB/Stash). This effectively removes the other person's
  lock and is likely to cause them problems if they had edited and try to push
  next time.

2. **Allow a parallel lock**
  Actually similar to breaking someone else's lock, except it lets you take
  another lock on a file in parallel, leaving their lock in place too, and
  knowing that you're going to have to resolve the merge problem later.  You
  could handle this just by manually making files read/write, then using 'force
  push' to override hooks that prevent pushing when not locked. However by
  explicitly registering a parallel lock (possible form: 'git lfs lock
  --force') this could be recorded and communicated to anyone else with a lock,
  letting them know about possible merge issues down the line.

## Detailed feature points
|No | Feature | Notes
|---|---------|------------------
|1  |Lock server must be available at same API URL|
|2  |Identify unmergeable files as subset of lfs files|`git lfs track -b` ?
|3  |Make unmergeable files read-only on checkout|Perform in smudge filter
|4  |Lock a file<ul><li>Check with server which must atomically check/set</li><li>Check person requesting the lock is checked out on a commit which is a descendent of the last edit of that file (locally or on server, although last lock shouldn't have been released until push anyway), or allow --force to break rule</li><li>Record lock on server</li><li>Make file read/write locally if success</li></ul>|`git lfs lock <file>`?
|5  |Release a lock<ul><li>Check if locally modified, if so must discard</li><li>Check if user has more recent commit of this file than server, if so must push first</li><li>Release lock on server atomically</li><li>Make local file read-only</li></ul>|`git lfs unlock <file>`?
|6  |Break a lock, ie override someone else's lock and take it yourself.<ul><li>Release lock on server atomically</li><li>Proceed as per 'Lock a file'</li><li>Notify original lock holder HOW?</li></ul>|`git lfs lock -break <file>`?
|7  |Release lock on reset (maybe). Configurable option / prompt? May be resetting just to start editing again|
|8  |Release lock on push (maybe, if unmodified). See above|
|9  |Cater for read-only binary files when merging locally<ul><li>Because files are read-only this might prevent merge from working when actually it's valid.</li><li>Always fine to merge the latest version of a binary file to anywhere else</li><li>Fine to merge the non-latest version if user is aware that this may cause merge problems (see Breaking the rules)</li><li>Therefore this feature is about dealing with the read-only flag and issuing a warning if not the latest</li></ul>|
|10 |List current locks<ul><li>That the current user has</li><li>That anyone has</li><li>Potentially scoped to folder</li></ul>|`git lfs lock --list [paths...]`
|11 |Reject a push containing a binary file currently locked by someone else|pre-receive hook on server, allow --force to override (i.e. existing parameter to git push)

## Locking challenges

### Making files read-only

This is useful because it means it provides a reminder that the user should be
locking the file before they start to edit it, to avoid the case of an unexpected
merge later on. 

I've done some tests with chmod and discovered:

* Removing the write bit doesn't cause the file to be marked modified (good)
* In most editors it either prevents saving or (in Apple tools) prompts to
  'unlock'. The latter is slightly unhelpful
* In terms of marking files that need locking, adding custom flags to
  .gitattributes (like 'lock') seems to work; `git check-attr -a <file>`
  correctly lists the custom attribute
* Once a file is marked read-only however, `git checkout` replaces it without
  prompting, with the write bit set
* We can use the `post-checkout` hook to make files read-only, but we don't get
  any file information, on refs. This means we'd have to scan the whole working
  copy to figure out what we needed to mark read-only. To do this we'd have to
  have the attribute information and all the current lock information. This
  could be time consuming.
  * A way to speed up the `post-checkout` would be to diff the pre- and post-ref
    information that's provided and only check the files that changed. In the case
    of single-file checkouts I'm not sure this is possible though.
  * We could also feed either the diff or a file scan into `git check-attr --stdin`
    in order to share the exe, or do our own attribute matching
* It's not entirely clear yet how merge & rebase might operate. May also need
  the `post-merge` hook
* See contrib/hooks/setgitperms.perl for an example; so this isn't unprecedented

#### Test cases for post-checkout

* Checkout a branch
  * Calls `post-checkout` with pre/post SHA and branch=1
* Checkout a tag
  * Calls `post-checkout` with pre/post SHA and branch=1 (even though it's a tag)
* Checkout by commit SHA
  * Calls `post-checkout` with pre/post SHA and branch=1 (even though it's a plain SHA)
* Checkout named files (e.g. discard changes)
  * Calls `post-checkout` with identical pre/post SHA (HEAD) and branch=0
* Reset all files (discard all changes ie git reset --hard HEAD) 
  * Doesn't call `post-checkout` - could restore write bit, but must have been
    set anyway for file to be edited, so not a problem?
* Reset a branch to a previous commit
  * Doesn't call `post-checkout` - PROBLEM because can restore write bit & file
    was not modified. BUT: rare & maybe liveable
* Merge a branch with lockable file changes (non-conflicting)
* Rebase a branch with lockable files (non-conflicting)
* Merge conflicts - fix then commit
* Rebase conflicts - fix then continue
* 


## Implementation details (Initial simple API-only pass)
### Types
To make the implementing locking on the lfs-test-server as well as other servers
in the future easier, it makes sense to create a `lock` package that can be
depended upon from any server. This will go along with Steve's refactor which
touches the `lfs` package quite a bit.

Below are enumerated some of the types that will presumably land in this
sub-package.

```go
// Lock represents a single lock that against a particular path.
//
// Locks returned from the API may or may not be currently active, according to
// the Expired flag.
type Lock struct {
        // Id is the unique identifier corresponding to this particular Lock. It
        // must be consistent with the local copy, and the server's copy.
        Id string `json:"id"`
        // Path is an absolute path to the file that is locked as a part of this
        // lock.
        Path string `json:"path"`
        // Committer is the author who initiated this lock.
        Committer struct {
               Name  string `json:"name"`
               Email string `json:"email"`
        } `json:"creator"`
        // CommitSHA is the commit that this Lock was created against. It is
        // strictly equal to the SHA of the minimum commit negotiated in order
        // to create this lock.
        CommitSHA string `json:"commit_sha"
        // LockedAt is a required parameter that represents the instant in time
        // that this lock was created. For most server implementations, this
        // should be set to the instant at which the lock was initially
        // received.
        LockedAt time.Time `json:"locked_at"`
        // ExpiresAt is an optional parameter that represents the instant in
        // time that the lock stopped being active. If the lock is still active,
        // the server can either a) not send this field, or b) send the
        // zero-value of time.Time.
        UnlockedAt time.Time `json:"unlocked_at,omitempty"`
}

// Active returns whether or not the given lock is still active against the file
// that it is protecting.
func (l *Lock) Active() bool {
        return time.IsZero(l.UnlockedAt)
}
```

### Proposed Commands

#### `git lfs lock <path>`

The `lock` command will be used in accordance with the multi-branch flow as
proposed above to request that lock be granted to the specific path passed an
argument to the command.

```go
// LockRequest encapsulates the payload sent across the API when a client would
// like to obtain a lock against a particular path on a given remote.
type LockRequest struct {
        // Path is the path that the client would like to obtain a lock against.
        Path      string `json:"path"`
        // LatestRemoteCommit is the SHA of the last known commit from the
        // remote that we are trying to create the lock against, as found in
        // `.git/refs/origin/<name>`.
        LatestRemoteCommit string `json:"latest_remote_commit"`
        // Committer is the individual that wishes to obtain the lock.
        Committer struct {
              // Name is the name of the individual who would like to obtain the
              // lock, for instance: "Rick Olson".
              Name string `json:"name"`
              // Email is the email assopsicated with the individual who would
              // like to obtain the lock, for instance: "rick@github.com".
              Email string `json:"email"`
        } `json:"committer"`
}
```

```go
// LockResponse encapsulates the information sent over the API in response to
// a `LockRequest`.
type LockResponse struct {
        // Lock is the Lock that was optionally created in response to the
        // payload that was sent (see above). If the lock already exists, then
        // the existing lock is sent in this field instead, and the author of
        // that lock remains the same, meaning that the client failed to obtain
        // that lock. An HTTP status of "409 - Conflict" is used here.
        //
        // If the lock was unable to be created, this field will hold the
        // zero-value of Lock and the Err field will provide a more detailed set
        // of information.
        //
        // If an error was experienced in creating this lock, then the
        // zero-value of Lock should be sent here instead.
        Lock Lock `json:"lock"`
        // CommitNeeded holds the minimum commit SHA that client must have to
        // obtain the lock.
        CommitNeeded string `json:"commit_needed"`
        // Err is the optional error that was encountered while trying to create
        // the above lock.
        Err error `json:"error,omitempty"`
}
```


#### `git lfs unlock <path>`

The `unlock` command is responsible for releasing the lock against a particular
file. The command takes a `<path>` argument which the LFS client will have to
internally resolve into a Id to unlock.

The API associated with this command can also be used on the server to remove
existing locks after a push.

```go
// An UnlockRequest is sent by the client over the API when they wish to remove
// a lock associated with the given Id.
type UnlockRequest struct {
        // Id is the identifier of the lock that the client wishes to remove.
        Id string `json:"id"`
}
```

```go
// UnlockResult is the result sent back from the API when asked to remove a
// lock.
type UnlockResult struct {
        // Lock is the lock corresponding to the asked-about lock in the
        // `UnlockPayload` (see above). If no matching lock was found, this
        // field will take the zero-value of Lock, and Err will be non-nil.
        Lock Lock `json:"lock"`
        // Err is an optional field which holds any error that was experienced
        // while removing the lock.
        Err error `json:"error,omitempty"`
}
```

Clients can determine whether or not their lock was removed by calling the
`Active()` method on the returned Lock, if `UnlockResult.Err` is nil.

#### `git lfs locks (-r <remote>|-b <branch|-p <path>)|(-i id)`

For many operations, the LFS client will need to have knowledge of existing
locks on the server. Additionally, the client should not have to self-sort/index
this (potentially) large set. To remove this need, both the `locks` command and
corresponding API method take several filters.

Clients should turn the flag-values that were passed during the command
invocation into `Filter`s as described below, and batched up into the `Filters`
field in the `LockListRequest`.

```go
// Property is a constant-type that narrows fields pertaining to the server's
// Locks.
type Property string

const (
        Branch Property = "branch"
        Id     Property = "id"
        // (etc) ...
)

// LockListRequest encapsulates the request sent to the server when the client
// would like a list of locks that match the given criteria.
type LockListRequest struct {
        // Filters is the set of filters to query against. If the client wishes
        // to obtain a list of all locks, an empty array should be passed here.
        Filters []{
                // Prop is the property to search against.
                Prop Property `json:"prop"`
                // Value is the value that the property must take.
                Value string `json:"value"`
        } `json:"filters"`
        // Cursor is an optional field used to tell the server which lock was
        // seen last, if scanning through multiple pages of results.
        //
        // Servers must return a list of locks sorted in reverse chronological
        // order, so the Cursor provides a consistent method of viewing all
        // locks, even if more were created between two requests.
        Cursor string `json:"cursor,omitempty"`
        // Limit is the maximum number of locks to return in a single page.
        Limit int `json:"limit"`
}
```

```go
// LockList encapsulates a set of Locks.
type LockList struct {
        // Locks is the set of locks returned back, typically matching the query
        // parameters sent in the LockListRequest call. If no locks were matched
        // from a given query, then `Locks` will be represented as an empty
        // array.
        Locks []Lock `json:"locks"`
        // NextCursor returns the Id of the Lock the client should update its
        // cursor to, if there are multiple pages of results for a particular
        // `LockListRequest`.
        NextCursor string `json:"next_cursor,omitempty"`
        // Err populates any error that was encountered during the search. If no
        // error was encountered and the operation was succesful, then a value
        // of nil will be passed here.
        Err error `json:"error,omitempty"`
}
