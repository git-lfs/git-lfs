package wildmatch

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// opt is an option type for configuring a new Wildmatch instance.
type opt func(w *Wildmatch)

var (
	// CaseFold allows the receiving Wildmatch to match paths with
	// different case structuring as in the pattern.
	CaseFold opt = func(w *Wildmatch) {
		w.caseFold = true
	}

	// SystemCase either folds or does not fold filepaths and patterns,
	// according to whether or not the operating system on which Wildmatch
	// runs supports case sensitive files or not.
	SystemCase opt
)

const (
	sep byte = '/'
)

// Wildmatch implements pattern matching against filepaths using the format
// described in the package documentation.
//
// For more, see documentation for package 'wildmatch'.
type Wildmatch struct {
	// ts are the token set used to match the given pattern.
	ts []token
	// p is the raw pattern used to derive the token set.
	p string

	// caseFold allows the instance Wildmatch to match patterns with the
	// same character but different case structures.
	caseFold bool
}

// NewWildmatch constructs a new Wildmatch instance which matches filepaths
// according to the given pattern and the rules for matching above.
//
// If the pattern is malformed, for instance, it has an unclosed character
// group, escape sequence, or character class, NewWildmatch will panic().
func NewWildmatch(p string, opts ...opt) *Wildmatch {
	w := &Wildmatch{p: slashEscape(p)}

	for _, opt := range opts {
		opt(w)
	}

	if w.caseFold {
		// Before parsing the pattern, convert it to lower-case.
		w.p = strings.ToLower(w.p)
	}

	w.ts = parseTokens(strings.Split(w.p, string(sep)))

	return w
}

const (
	// escapes is a constant string containing all escapable characters
	escapes = "\\[]*?"
)

// slashEscape converts paths "p" to POSIX-compliant path, independent of which
// escape character the host machine uses.
//
// slashEscape resepcts escapable sequences, and thus will not transform
// `foo\*bar` to `foo/*bar` on non-Windows operating systems.
func slashEscape(p string) string {
	var pp string

	for i := 0; i < len(p); {
		c := p[i]

		switch c {
		case '\\':
			if i+1 < len(p) && escapable(p[i+1]) {
				pp += `\`
				pp += string(p[i+1])

				i += 2
			} else {
				pp += `/`
				i += 1
			}
		default:
			pp += string(c)
			i += 1
		}
	}

	return pp
}

// escapable returns whether the given "c" is escapable.
func escapable(c byte) bool {
	return strings.IndexByte(escapes, c) > -1
}

// parseTokens parses a separated list of patterns into a sequence of
// representative Tokens that will compose the pattern when applied in sequence.
func parseTokens(dirs []string) []token {
	if len(dirs) == 0 {
		return make([]token, 0)
	}

	switch dirs[0] {
	case "":
		return parseTokens(dirs[1:])
	case "**":
		rest := parseTokens(dirs[1:])
		if len(rest) == 0 {
			// If there are no remaining tokens, return a lone
			// doubleStar token.
			return []token{&doubleStar{
				Until: nil,
			}}
		}

		// Otherwise, return a doubleStar token that will match greedily
		// until the first component in the remainder of the pattern,
		// and then the remainder of the pattern.
		return append([]token{&doubleStar{
			Until: rest[0],
		}}, rest[1:]...)
	default:
		// Ordinarily, simply return the appropriate component, and
		// continue on.
		return append([]token{&component{
			fns: parseComponent(dirs[0]),
		}}, parseTokens(dirs[1:])...)
	}
}

// nonEmpty returns the non-empty strings in "all".
func nonEmpty(all []string) (ne []string) {
	for _, x := range all {
		if len(x) > 0 {
			ne = append(ne, x)
		}
	}
	return ne
}

// Match returns true if and only if the pattern matched by the receiving
// Wildmatch matches the entire filepath "t".
func (w *Wildmatch) Match(t string) bool {
	dirs, ok := w.consume(t)
	if !ok {
		return false
	}
	return len(dirs) == 0
}

// consume performs the inner match of "t" against the receiver's pattern, and
// returns a slice of remaining directory paths, and whether or not there was a
// disagreement while matching.
func (w *Wildmatch) consume(t string) ([]string, bool) {
	if w.caseFold {
		// If the receiving Wildmatch is case insensitive, the pattern
		// "w.p" will be lower-case.
		//
		// To preserve insensitivity, lower the given path "t", as well.
		t = strings.ToLower(t)
	}

	dirs := strings.Split(t, string(sep))
	isDir := strings.HasSuffix(t, string(sep))

	// Match each directory token-wise, allowing each token to consume more
	// than one directory in the case of the '**' pattern.
	for _, tok := range w.ts {
		var ok bool

		dirs, ok = tok.Consume(dirs, isDir)
		if !ok {
			// If a pattern could not match the remainder of the
			// filepath, return so immediately, along with the paths
			// that we did successfully manage to match.
			return dirs, false
		}
	}
	return dirs, true
}

// String implements fmt.Stringer and returns the receiver's pattern in the format
// specified above.
func (w *Wildmatch) String() string {
	return w.p
}

// token matches zero, one, or more directory components.
type token interface {
	// Consume matches zero, one, or more directory components.
	//
	// Consider the following examples:
	//
	//   (["foo", "bar", "baz"]) -> (["oo", "bar", baz"], true)
	//   (["foo", "bar", "baz"]) -> (["bar", baz"], true)
	//   (["foo", "bar", "baz"]) -> (["baz"], true)
	//   (["foo", "bar", "baz"]) -> ([], true)
	//   (["foo", "bar", "baz"]) -> (["foo", "bar", "baz"], false)
	//   (["foo", "bar", "baz"]) -> (["oo", "bar", "baz"], false)
	//   (["foo", "bar", "baz"]) -> (["bar", "baz"], false)
	//
	// The Consume operation can reduce the size of a single entry in the
	// slice (see: example (1) above), or remove it entirely, (see: examples
	// (2), (3), and (4) above). It can also refuse to match forward after
	// making any amount of progress (see: examples (5), (6), and (7)
	// above).
	//
	// Consume accepts a slice representing a path-delimited filepath on
	// disk, and a bool indicating whether the given path is a directory
	// (i.e., "foo/bar/" is, but "foo/bar" isn't).
	Consume(path []string, isDir bool) ([]string, bool)

	// String returns the string representation this component of the
	// pattern; i.e., a string that, when parsed, would form the same token.
	String() string
}

// doubleStar is an implementation of the Token interface which greedily matches
// one-or-more path components until a successor token.
type doubleStar struct {
	Until token
}

// Consume implements token.Consume as above.
func (d *doubleStar) Consume(path []string, isDir bool) ([]string, bool) {
	// If there are no remaining tokens to match, allow matching the entire
	// path.
	if d.Until == nil {
		return nil, true
	}

	for i := len(path); i > 0; i-- {
		rest, ok := d.Until.Consume(path[i:], false)
		if ok {
			return rest, ok
		}
	}

	// If no match has been found, we assume that the '**' token matches the
	// empty string, and defer pattern matching to the rest of the path.
	return d.Until.Consume(path, isDir)
}

// String implements Component.String.
func (d *doubleStar) String() string {
	if d.Until == nil {
		return "**"
	}
	return fmt.Sprintf("**/%s", d.Until.String())
}

// componentFn is a functional type designed to match a single component of a
// directory structure by reducing the unmatched part, and returning whether or
// not a match was successful.
type componentFn interface {
	Apply(s string) (rest string, ok bool)
	String() string
}

// cfn is a wrapper type for the Component interface that includes an applicable
// function, and a string that represents it.
type cfn struct {
	fn  func(s string) (rest string, ok bool)
	str string
}

// Apply executes the component function as described above.
func (c *cfn) Apply(s string) (rest string, ok bool) {
	return c.fn(s)
}

// String returns the string representation of this component.
func (c *cfn) String() string {
	return c.str
}

// component is an implementation of the Token interface, which matches a single
// component at the front of a tree structure by successively applying
// implementations of the componentFn type.
type component struct {
	// fns is the list of componentFn implementations to be successively
	// applied.
	fns []componentFn
}

// parseComponent parses a single component from its string representation,
// including wildcards, character classes, string literals, and escape
// sequences.
func parseComponent(s string) []componentFn {
	if len(s) == 0 {
		// The empty string represents the absence of componentFn's.
		return make([]componentFn, 0)
	}

	switch s[0] {
	case '\\':
		// If the first character is a '\', the following character is a
		// part of an escape sequence, or it is unclosed.
		if len(s) < 2 {
			panic("wildmatch: unclosed escape sequence")
		}

		literal := substring(string(s[1]))

		var rest []componentFn
		if len(s) > 2 {
			// If there is more to follow, i.e., "\*foo", then parse
			// the remainder.
			rest = parseComponent(s[2:])
		}
		return cons(literal, rest)
	case '[':
		var (
			// i will denote the currently-inspected index of the character
			// group.
			i int = 1
			// include will denote the list of included runeFn's
			// composing the character group.
			include []runeFn
			// exclude will denote the list of excluded runeFn's
			// composing the character group.
			exclude []runeFn
			// run is the current run of strings (to either compose
			// a range, or select "any")
			run string
			// neg is whether we have seen a negation marker.
			neg bool
		)

		for i < len(s) {
			if s[i] == '^' || s[i] == '!' {
				// Once a '^' or '!' character has been seen,
				// anything following it will be negated.
				neg = !neg
				i = i + 1
			} else if strings.HasPrefix(s[i:], "[:") {
				close := strings.Index(s[i:], ":]")
				if close < 0 {
					panic("unclosed character class")
				}

				if close == 1 {
					// The case "[:]" has a prefix "[:", and
					// a suffix ":]", but the atom refers to
					// a character group including the
					// literal ":", not an ill-formed
					// character class.
					//
					// Parse it as such; increment one
					// _less_ than expected, to terminate
					// the group.
					run += "[:]"
					i = i + 2
					continue
				}

				// Find the associated character class.
				name := strings.TrimPrefix(
					strings.ToLower(s[i:i+close]), "[:")
				fn, ok := classes[name]
				if !ok {
					panic(fmt.Sprintf("wildmatch: unknown class: %q", name))
				}

				include, exclude = appendMaybe(!neg, include, exclude, fn)
				// Advance to the first index beyond the closing
				// ":]".
				i = i + close + 2
			} else if s[i] == '-' {
				if i < len(s) {
					// If there is a range marker at the
					// non-final position, construct a range
					// and an optional "any" match:
					var start, end byte
					if len(run) > 0 {
						// If there is at least one
						// character in the run, use it
						// as the starting point of the
						// range, and remove it from the
						// run.
						start = run[len(run)-1]
						run = run[:len(run)-1]
					}
					end = s[i+1]

					if len(run) > 0 {
						// If there is still information
						// in the run, construct a rune
						// function matching any
						// characters in the run.
						cfn := anyRune(run)

						include, exclude = appendMaybe(!neg, include, exclude, cfn)
						run = ""
					}

					// Finally, construct the rune range and
					// add it appropriately.
					bfn := between(rune(start), rune(end))
					include, exclude = appendMaybe(!neg,
						include, exclude, bfn)

					i = i + 2
				} else {
					// If this is in the final position, add
					// it to the run and exit the loop.
					run = run + "-"
					i = i + 2
				}
			} else if s[i] == '\\' {
				// If we encounter an escape sequence in the
				// group, check its bounds and add it to the
				// run.
				if i+1 >= len(s) {
					panic("wildmatch: unclosed escape")
				}
				run = run + string(s[i+1])
				i = i + 2
			} else if s[i] == ']' {
				// If we encounter a closing ']', then stop
				// parsing the group.
				break
			} else {
				// Otherwise, add the character to the run and
				// advance forward.
				run = run + string(s[i])
				i = i + 1
			}
		}

		if len(run) > 0 {
			fn := anyRune(run)
			include, exclude = appendMaybe(!neg, include, exclude, fn)
		}

		var rest string
		if i+1 < len(s) {
			rest = s[i+1:]
		}
		// Assemble a character class, and cons it in front of the
		// remainder of the component pattern.
		return cons(charClass(include, exclude), parseComponent(rest))
	case '?':
		return []componentFn{wildcard(1, parseComponent(s[1:]))}
	case '*':
		return []componentFn{wildcard(-1, parseComponent(s[1:]))}
	default:
		// Advance forward until we encounter a special character
		// (either '*', '[', '*', or '?') and parse across the divider.
		var i int
		for ; i < len(s); i++ {
			if s[i] == '[' ||
				s[i] == '*' ||
				s[i] == '?' ||
				s[i] == '\\' {
				break
			}
		}

		return cons(substring(s[:i]), parseComponent(s[i:]))
	}
}

// appendMaybe appends the value "x" to either "a" or "b" depending on "yes".
func appendMaybe(yes bool, a, b []runeFn, x runeFn) (ax, bx []runeFn) {
	if yes {
		return append(a, x), b
	}
	return a, append(b, x)
}

// cons prepends the "head" componentFn to the "tail" of componentFn's.
func cons(head componentFn, tail []componentFn) []componentFn {
	return append([]componentFn{head}, tail...)
}

// Consume implements token.Consume as above by applying the above set of
// componentFn's in succession to the first element of the path tree.
func (c *component) Consume(path []string, isDir bool) ([]string, bool) {
	if len(path) == 0 {
		return path, false
	}

	head := path[0]
	for _, fn := range c.fns {
		var ok bool

		// Apply successively the component functions to make progress
		// matching the head.
		if head, ok = fn.Apply(head); !ok {
			// If any of the functions failed to match, there are
			// no other paths to match success, so return a failure
			// immediately.
			return path, false
		}
	}

	if len(head) > 0 {
		return append([]string{head}, path[1:]...), false
	}

	if len(path) == 1 {
		// Components can not match directories. If we were matching the
		// last path in a tree structure, we can only match if it
		// _wasn't_ a directory.
		return path[1:], !isDir
	}

	return path[1:], true
}

// String implements token.String.
func (c *component) String() string {
	var str string

	for _, fn := range c.fns {
		str += fn.String()
	}
	return str
}

// substring returns a componentFn that matches a prefix of "sub".
func substring(sub string) componentFn {
	return &cfn{
		fn: func(s string) (rest string, ok bool) {
			if !strings.HasPrefix(s, sub) {
				return s, false
			}
			return s[len(sub):], true
		},
		str: sub,
	}
}

// wildcard returns a componentFn that greedily matches until a set of other
// component functions no longer matches.
func wildcard(n int, fns []componentFn) componentFn {
	until := func(s string) (string, bool) {
		head := s
		for _, fn := range fns {
			var ok bool

			if head, ok = fn.Apply(head); !ok {
				return s, false
			}
		}

		if len(head) > 0 {
			return s, false
		}
		return "", true
	}

	var str string = "*"
	for _, fn := range fns {
		str += fn.String()
	}

	return &cfn{
		fn: func(s string) (rest string, ok bool) {
			if n > -1 {
				if n > len(s) {
					return "", false
				}
				return until(s[n:])
			}

			for i := len(s); i > 0; i-- {
				rest, ok = until(s[i:])
				if ok {
					return rest, ok
				}
			}
			return until(s)
		},
		str: str,
	}
}

// charClass returns a component function emulating a character class, i.e.,
// that a single character can match if and only if it is included in one of the
// includes (or true if there were no includes) and none of the excludes.
func charClass(include, exclude []runeFn) componentFn {
	return &cfn{
		fn: func(s string) (rest string, ok bool) {
			if len(s) == 0 {
				return s, false
			}

			// Find "r", the first rune in the string "s".
			r, l := utf8.DecodeRuneInString(s)

			var match bool
			for _, ifn := range include {
				// Attempt to find a match on "r" with "ifn".
				if ifn(r) {
					match = true
					break
				}
			}

			// If there wasn't a match and there were some including
			// patterns, return a failure to match. Otherwise, continue on
			// to make sure that no patterns exclude the rune "r".
			if !match && len(include) != 0 {
				return s, false
			}

			for _, efn := range exclude {
				// Attempt to find a negative match on "r" with "efn".
				if efn(r) {
					return s, false
				}
			}

			// If we progressed this far, return the remainder of the
			// string.
			return s[l:], true
		},
		str: "<charclass>",
	}
}

// runeFn matches a single rune.
type runeFn func(rune) bool

var (
	// classes is a mapping from character class name to a rune function
	// that implements its behavior.
	classes = map[string]runeFn{
		"alnum": func(r rune) bool {
			return unicode.In(r, unicode.Number, unicode.Letter)
		},
		"alpha": unicode.IsLetter,
		"blank": func(r rune) bool {
			return r == ' ' || r == '\t'
		},
		"cntrl": unicode.IsControl,
		"digit": unicode.IsDigit,
		"graph": unicode.IsGraphic,
		"lower": unicode.IsLower,
		"print": unicode.IsPrint,
		"punct": unicode.IsPunct,
		"space": unicode.IsSpace,
		"upper": unicode.IsUpper,
		"xdigit": func(r rune) bool {
			return unicode.IsDigit(r) ||
				('a' <= r && r <= 'f') ||
				('A' <= r && r <= 'F')
		},
	}
)

// anyRune returns true so long as the rune "r" appears in the string "s".
func anyRune(s string) runeFn {
	return func(r rune) bool {
		return strings.IndexRune(s, r) > -1
	}
}

// between returns true so long as the rune "r" appears between "a" and "b".
func between(a, b rune) runeFn {
	if b < a {
		a, b = b, a
	}

	return func(r rune) bool {
		return a <= r && r <= b
	}
}
