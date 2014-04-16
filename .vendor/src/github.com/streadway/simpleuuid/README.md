# Simple UUIDs

[![Build Status][1]][2]

[1]: https://secure.travis-ci.org/streadway/simpleuuid.png
[2]: http://www.travis-ci.org/streadway/simpleuuid

This implements a variant of Format 1 from [RFC 4122][rfc4122] that is intended
to be roughly sortable and play nicely as Cassandra TimeUUID keys.  Much of the
inspiration comes form having used [ryanking's simple\_uuid
gem](https://github.com/ryanking/simple_uuid).

As the package implies, this is a simple UUID generator.  It doesn't offer
total sorting, monotonic increments or prevent timing collisions.

This package does offer a simple combination of random and time based
uniqueness that will play nicely if you want a unique key from a Time object.
If your time objects sort with a granularity of 100 nanoseconds then the UUIDs
generated will have the same order.  UUIDs with the same time have undefined
order.

# Other UUIDs

The other formats described in [RFC 4122][rfc4122] should be parsable either in
text or byte form, though will not be sortable or likely have a meaningful time
componenet.

# Contributions

Send a pull request with a tested `go fmt` change in a branch other than
`master`.  Do try to organize your commits to be atomic to the changes
introduced.

# License

Copyright (C) 2012 by Sean Treadway ([streadway](http://github.com/streadway))

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

[rfc4122]: http://www.ietf.org/rfc/rfc4122.txt
