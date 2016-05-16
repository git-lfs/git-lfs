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
