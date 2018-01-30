// package Wildmatch is an implementation of Git's wildmatch.c-style pattern
// matching.
//
// Wildmatch patterns are comprised of any combination of the following three
// components:
//
//   - String literals. A string literal is "foo", or "foo\*" (matching "foo",
//     and "foo\", respectively). In general, string literals match their exact
//     contents in a filepath, and cannot match over directories unless they
//     include the operating system-specific path separator.
//
//   - Wildcards. There are three types of wildcards:
//
//      - Single-asterisk ('*'): matches any combination of characters, any
//        number of times. Does not match path separators.
//
//      - Single-question mark ('?'): matches any single character, but not a
//        path separator.
//
//      - Double-asterisk ('**'): greedily matches any number of directories.
//        For example, '**/foo' matches '/foo', 'bar/baz/woot/foot', but not
//        'foo/bar'. Double-asterisks must be separated by filepath separators
//        on either side.
//
//   - Character groups. A character group is composed of a set of included and
//     excluded character types. The set of included character types begins the
//     character group, and a '^' or '!' separates it from the set of excluded
//     character types.
//
//     A character type can be one of the following:
//
//       - Character literal: a single character, i.e., 'c'.
//
//       - Character group: a group of characters, i.e., '[:alnum:]', etc.
//
//       - Character range: a range of characters, i.e., 'a-z'.
//
// A Wildmatch pattern can be any combination of the above components, in any
// ordering, and repeated any number of times.
package wildmatch
