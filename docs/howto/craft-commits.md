# Crafting Commits

The Git LFS project strives to be a self-documenting history of its own
development. That is to say, as contributors to Git LFS, we hope that the
information guiding our decision-making process during development is retained
in the history of the project.

The way we retain this kind of information is by creating thoughtful commits.
This document outlines answers to two critical questions:

  1. What makes a good commit message?

  2. What goes in a commit?

We answer these questions in an effort to make it as easy as possible to review
new code as it is being written, and to refer back to old code when we have a
question about it. In each case, a developer should be able to quickly parse
through information, and figure out what they need to know.

## What makes a good commit message?

A good commit message is descriptive but not verbose. It contains information
relevant in the decision making process, but is not redundant or repetitive. It
explains the detail of a change, why it was made, and evaluates alternatives.

Presented here is an example of a hypothetical commit message:

```
git/githistory/rewriter.go: don't revisit unchanged subtrees

Since [1], we have used package git/githistory to re-assemble blobs,
trees, and commits in order to rewrite history as when running
'git lfs migrate import', or similar.

Since package git/githistory assumes that a rewrite is a pure function
of the contents of a given tree, there is no need to run a migration on
the same tree twice. In other words, it is possible to cache the
result of such a migration and reuse it, without opening the tree up.

This saves on time proportional to the similarity of trees throughout
history in a repository. It has the potential to consume more memory,
but this cost is likely outweighed in the savings from avoiding opening
up many trees.

In the future, this can be improved by only caching a certain amount of
trees, i.e., as in a circular buffer.

[1]: <SHA-1>
```

The commit body follows the form below:

  1. A short, 72-character max description of the change, prefixed with the
     package or file under consideration.

  2. A brief historical overview that guides the reader's attention towards the
     change being made.

  3. An overview of the change itself, with an optional section describing why
     the change might benefit the surrounding code, or why it's a step in the
     right direction (e.g., preparing for a future commit, etc.)

  4. A possible future direction, or something that a reader might want to know
     when referring back to the commit later on.

The commit message is designed to prepare the reader to review or look at the
diff associated with the commit. This occurs in two situations: a reviewer is
looking at a pull request and wants to understand the development flow, or, an
engineer is curious why a program is behaving in a certain way, and runs `git
blame`, thereby uncovering the commit message.

In either case, the commit message's job is to guide the reader towards an
understanding of the code attached to it. That is, a reader should be able to
(with little prior context) be able to understand a sketch of the code attached
to the commit.

A commit message can additionally include niceties like:

  - Performance metrics, indicating a positive/negative/unchanged delta before
    and after the commit.

  - A (brief) script, showing how the change might be observed.

  - A motivating example, or

  - Anything else not directly related to Git LFS that motivates or illustrates
    the rationale behind a change.

## What goes in a commit?

Now that we have described what makes a good commit message, let's consider how
we might make it easy to write such a message. It is often difficult to write a
good commit message for a large change, and this difficulty is often an
indication to reduce the scope of a commit, though it is also sometimes the case
that this _isn't_ the thing to do.
