package odb

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

// Signature represents a commit signature, which can represent either
// committership or authorship of the commit that this signature belongs to. It
// specifies a name, email, and time that the signature was created.
//
// NOTE: this type is _not_ used by the `*Commit` instance, as it does not
// preserve cruft bytes. It is kept as a convenience type to test with.
type Signature struct {
	// Name is the first and last name of the individual holding this
	// signature.
	Name string
	// Email is the email address of the individual holding this signature.
	Email string
	// When is the instant in time when the signature was created.
	When time.Time
}

const (
	formatTimeZoneOnly = "-0700"
)

// String implements the fmt.Stringer interface and formats a Signature as
// expected in the Git commit internal object format. For instance:
//
//  Taylor Blau <ttaylorr@github.com> 1494258422 -0600
func (s *Signature) String() string {
	at := s.When.Unix()
	zone := s.When.Format(formatTimeZoneOnly)

	return fmt.Sprintf("%s <%s> %d %s", s.Name, s.Email, at, zone)
}

// ExtraHeader encapsulates a key-value pairing of header key to header value.
// It is stored as a struct{string, string} in memory as opposed to a
// map[string]string to maintain ordering in a byte-for-byte encode/decode round
// trip.
type ExtraHeader struct {
	// K is the header key, or the first run of bytes up until a ' ' (\x20)
	// character.
	K string
	// V is the header value, or the remaining run of bytes in the line,
	// stripping off the above "K" field as a prefix.
	V string
}

// Commit encapsulates a Git commit entry.
type Commit struct {
	// Author is the Author this commit, or the original writer of the
	// contents.
	//
	// NOTE: this field is stored as a string to ensure any extra "cruft"
	// bytes are preserved through migration.
	Author string
	// Committer is the individual or entity that added this commit to the
	// history.
	//
	// NOTE: this field is stored as a string to ensure any extra "cruft"
	// bytes are preserved through migration.
	Committer string
	// ParentIDs are the IDs of all parents for which this commit is a
	// linear child.
	ParentIDs [][]byte
	// TreeID is the root Tree associated with this commit.
	TreeID []byte
	// ExtraHeaders stores headers not listed above, for instance
	// "encoding", "gpgsig", or "mergetag" (among others).
	ExtraHeaders []*ExtraHeader
	// Message is the commit message, including any signing information
	// associated with this commit.
	Message string
}

// Type implements Object.ObjectType by returning the correct object type for
// Commits, CommitObjectType.
func (c *Commit) Type() ObjectType { return CommitObjectType }

// Decode implements Object.Decode and decodes the uncompressed commit being
// read. It returns the number of uncompressed bytes being consumed off of the
// stream, which should be strictly equal to the size given.
//
// If any error was encountered along the way, that will be returned, along with
// the number of bytes read up to that point.
func (c *Commit) Decode(from io.Reader, size int64) (n int, err error) {
	var finishedHeaders bool
	var messageParts []string

	s := bufio.NewScanner(from)
	for s.Scan() {
		text := s.Text()
		n = n + len(text+"\n")

		if len(s.Text()) == 0 {
			finishedHeaders = true
			continue
		}

		if fields := strings.Fields(text); len(fields) > 0 && !finishedHeaders {
			switch fields[0] {
			case "tree":
				id, err := hex.DecodeString(fields[1])
				if err != nil {
					return n, err
				}
				c.TreeID = id
			case "parent":
				id, err := hex.DecodeString(fields[1])
				if err != nil {
					return n, err
				}
				c.ParentIDs = append(c.ParentIDs, id)
			case "author":
				c.Author = strings.Join(fields[1:], " ")
			case "committer":
				c.Committer = strings.Join(fields[1:], " ")
			default:
				c.ExtraHeaders = append(c.ExtraHeaders, &ExtraHeader{
					K: fields[0],
					V: strings.Join(fields[1:], " "),
				})
			}
		} else {
			messageParts = append(messageParts, s.Text())
		}
	}

	c.Message = strings.Join(messageParts, "\n")

	if err = s.Err(); err != nil {
		return n, err
	}
	return n, err
}

// Encode encodes the commit's contents to the given io.Writer, "w". If there was
// any error copying the commit's contents, that error will be returned.
//
// Otherwise, the number of bytes written will be returned.
func (c *Commit) Encode(to io.Writer) (n int, err error) {
	n, err = fmt.Fprintf(to, "tree %s\n", hex.EncodeToString(c.TreeID))
	if err != nil {
		return n, err
	}

	for _, pid := range c.ParentIDs {
		n1, err := fmt.Fprintf(to, "parent %s\n", hex.EncodeToString(pid))
		if err != nil {
			return n, err
		}

		n = n + n1
	}

	n2, err := fmt.Fprintf(to, "author %s\ncommitter %s\n", c.Author, c.Committer)
	if err != nil {
		return n, err
	}

	n = n + n2

	for _, hdr := range c.ExtraHeaders {
		n3, err := fmt.Fprintf(to, "%s %s\n", hdr.K, hdr.V)
		if err != nil {
			return n, err
		}

		n = n + n3
	}

	n4, err := fmt.Fprintf(to, "\n%s\n", c.Message)
	if err != nil {
		return n, err
	}

	return n + n4, err
}
