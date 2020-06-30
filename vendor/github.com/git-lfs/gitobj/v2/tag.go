package gitobj

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"
)

type Tag struct {
	Object     []byte
	ObjectType ObjectType
	Name       string
	Tagger     string

	Message string
}

// Decode implements Object.Decode and decodes the uncompressed tag being
// read. It returns the number of uncompressed bytes being consumed off of the
// stream, which should be strictly equal to the size given.
//
// If any error was encountered along the way it will be returned, and the
// receiving *Tag is considered invalid.
func (t *Tag) Decode(hash hash.Hash, r io.Reader, size int64) (int, error) {
	scanner := bufio.NewScanner(io.LimitReader(r, size))

	var (
		finishedHeaders bool
		message         []string
	)

	for scanner.Scan() {
		if finishedHeaders {
			message = append(message, scanner.Text())
		} else {
			if len(scanner.Bytes()) == 0 {
				finishedHeaders = true
				continue
			}

			parts := strings.SplitN(scanner.Text(), " ", 2)
			if len(parts) < 2 {
				return 0, fmt.Errorf("gitobj: invalid tag header: %s", scanner.Text())
			}

			switch parts[0] {
			case "object":
				sha, err := hex.DecodeString(parts[1])
				if err != nil {
					return 0, fmt.Errorf("gitobj: unable to decode SHA-1: %s", err)
				}

				t.Object = sha
			case "type":
				t.ObjectType = ObjectTypeFromString(parts[1])
			case "tag":
				t.Name = parts[1]
			case "tagger":
				t.Tagger = parts[1]
			default:
				return 0, fmt.Errorf("gitobj: unknown tag header: %s", parts[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	t.Message = strings.Join(message, "\n")

	return int(size), nil
}

// Encode encodes the Tag's contents to the given io.Writer, "w". If there was
// any error copying the Tag's contents, that error will be returned.
//
// Otherwise, the number of bytes written will be returned.
func (t *Tag) Encode(w io.Writer) (int, error) {
	headers := []string{
		fmt.Sprintf("object %s", hex.EncodeToString(t.Object)),
		fmt.Sprintf("type %s", t.ObjectType),
		fmt.Sprintf("tag %s", t.Name),
		fmt.Sprintf("tagger %s", t.Tagger),
	}

	return fmt.Fprintf(w, "%s\n\n%s", strings.Join(headers, "\n"), t.Message)
}

// Equal returns whether the receiving and given Tags are equal, or in other
// words, whether they are represented by the same SHA-1 when saved to the
// object database.
func (t *Tag) Equal(other *Tag) bool {
	if (t == nil) != (other == nil) {
		return false
	}

	if t != nil {
		return bytes.Equal(t.Object, other.Object) &&
			t.ObjectType == other.ObjectType &&
			t.Name == other.Name &&
			t.Tagger == other.Tagger &&
			t.Message == other.Message
	}

	return true
}

// Type implements Object.ObjectType by returning the correct object type for
// Tags, TagObjectType.
func (t *Tag) Type() ObjectType {
	return TagObjectType
}
