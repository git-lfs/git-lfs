package odb

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/errors"
)

// Signature represents a commit signature, which can represent either
// committership or authorship of the commit that this signature belongs to. It
// specifies a name, email, and time that the signature was created.
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
	parseTimeZoneOnly  = "2006 -0700"
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

var (
	signatureNameRe  = regexp.MustCompile("^[^<]+")
	signatureEmailRe = regexp.MustCompile("<(.*)>")
	signatureTimeRe  = regexp.MustCompile("[-]?\\d+")
)

// ParseSignature parses a given string into a signature instance, returning any
// error that it encounters along the way.
//
// ParseSignature expects the signature encoded in the given string to be
// formatted correctly, and reproduce-able by the Signature.String() function
// above.
func ParseSignature(str string) (*Signature, error) {
	name := signatureNameRe.FindString(str)
	emailParts := signatureEmailRe.FindStringSubmatch(str)
	timeParts := signatureTimeRe.FindAllStringSubmatch(str, 2)

	if len(emailParts) < 2 {
		return nil, errors.Errorf("git/odb: expected email in signature: %q", str)
	}
	email := emailParts[1]

	if len(timeParts) < 1 {
		return nil, errors.Errorf("git/odb: expected time in signature: %q", str)
	}

	epoch, err := strconv.ParseInt(timeParts[0][0], 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(epoch, 0).In(time.UTC)
	if len(timeParts) > 1 && timeParts[1][0] != "0000" {
		loc, err := parseTimeZone(timeParts[1][0])
		if err != nil {
			return nil, err
		}

		t = t.In(loc)
	}

	return &Signature{
		Name:  strings.TrimSpace(name),
		Email: email,
		When:  t,
	}, nil
}

// parseTimeZone returns the *time.Location corresponding to a Git-formatted
// string offset. For instance, the string "-0700" would format into a
// *time.Location that is 7 hours east of UTC.
func parseTimeZone(zone string) (*time.Location, error) {
	loc, err := time.Parse(parseTimeZoneOnly, fmt.Sprintf("1970 %s", zone))
	if err != nil {
		return nil, err
	}
	return loc.Location(), nil
}

// Commit encapsulates a Git commit entry.
type Commit struct {
	// Author is the Author this commit, or the original writer of the
	// contents.
	Author *Signature
	// Committer is the individual or entity that added this commit to the
	// history.
	Committer *Signature
	// ParentIDs are the IDs of all parents for which this commit is a
	// linear child.
	ParentIDs [][]byte
	// TreeID is the root Tree associated with this commit.
	TreeID []byte
	// Message is the commit message, including any signing information
	// associated with this commit.
	Message string
}

var _ Object = (*Commit)(nil)

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
	s := bufio.NewScanner(from)
	for s.Scan() {
		text := s.Text()
		if len(s.Text()) == 0 {
			continue
		}

		fields := strings.Fields(text)
		if len(fields) > 0 {
			switch fields[0] {
			case "tree":
				id, err := hex.DecodeString(fields[1])
				if err != nil {
					panic(1)
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
				author, err := ParseSignature(strings.Join(fields[1:], " "))
				if err != nil {
					return n, err
				}

				c.Author = author
			case "committer":
				committer, err := ParseSignature(strings.Join(fields[1:], " "))
				if err != nil {
					return n, err
				}

				c.Committer = committer
			default:
				if len(c.Message) == 0 {
					c.Message = s.Text()
				} else {
					c.Message = strings.Join([]string{c.Message, s.Text()}, "\n")
				}
			}
		}

		n = n + len(text+"\n")
	}

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

	n2, err := fmt.Fprintf(to, "author %s\ncommitter %s\n\n%s\n",
		c.Author, c.Committer, c.Message)
	if err != nil {
		return n, err
	}

	return n + n2, err
}
