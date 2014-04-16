/*
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
*/

/*
Implements a variant of Format 1 from RFC 4122 that is intended for stable
sorting and play nicely as Cassandra TimeUUID keys.

The other formats described in RFC 4122 should be parsable either in text or
byte form, though will not be sortable or likely have a meaningful time
componenet.
*/
package simpleuuid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

const (
	gregorianEpoch = 0x01B21DD213814000
	size           = 16
	variant8       = 8 // sec. 4.1.1
	version1       = 1 // sec. 4.1.3
)

var (
	errLength = errors.New("mismatched UUID length")
)

/*
Byte encoded sequence in the following form:

   0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                          time_low                             |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |       time_mid                |         time_hi_and_version   |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |clk_seq_hi_res |  clk_seq_low  |         node (0-1)            |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                         node (2-5)                            |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type UUID []byte

type uuidTime int64

// Makes a copy of the UUID. Assumes the provided UUID is valid
func Copy(uuid UUID) UUID {
	dup, _ := NewBytes(uuid)
	return dup
}

// Allocates a new UUID from the given time, up to 8 bytes clock sequence and
// node data supplied by the caller.  The high 4 bits of the first byte will be
// masked with the UUID variant.
//
// Byte slices shorter than 8 will be right aligned to the clock, and node
// fields.  For example, if you provide 4 byte slice of 0x0a0b0c0d", the last 8
// bytes of the new UUID will be 0x08000000a0b0c0d.
func NewTimeBytes(t time.Time, bytes []byte) (UUID, error) {
	if len(bytes) > size {
		return nil, errLength
	}

	me := make([]byte, size)
	ts := fromUnixNano(t.UTC().UnixNano())

	// time masked with version
	binary.BigEndian.PutUint32(me[0:4], uint32(ts&0xffffffff))
	binary.BigEndian.PutUint16(me[4:6], uint16((ts>>32)&0xffff))
	binary.BigEndian.PutUint16(me[6:8], uint16((ts>>48)&0x0fff)|version1<<12)

	// right aligned remaining 8 bytes masked with variant
	copy(me[8+8-len(bytes):size], bytes[:len(bytes)])
	me[8] = me[8]&0x0f | variant8<<4

	return UUID(me), nil
}

// Allocate a UUID from a 16 byte sequence.  This can take any version,
// although versions other than 1 will not have a meaningful time component.
func NewBytes(bytes []byte) (UUID, error) {
	if len(bytes) != size {
		return nil, errLength
	}

	// Copy out this slice so not to hold a reference to the container
	b := make([]byte, size)
	copy(b, bytes[0:size])

	return UUID(b), nil
}

// Allocate a new UUID from a time, encoding the timestamp from the UTC
// timezone and using a random value for the clock and node.
func NewTime(t time.Time) (UUID, error) {
	rnd := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, rnd)
	if n != len(rnd) {
		return nil, errLength
	}
	if err != nil {
		return nil, err
	}
	return NewTimeBytes(t, rnd)
}

// Parse and allocate from a string encoded UUID like:
// "6ba7b811-9dad-11d1-80b4-00c04fd430c8".  Does not validate the time, node
// or clock are reasonable values, though it is intended to round trip from a
// string to a string for all versions of UUIDs.
func NewString(s string) (UUID, error) {
	normalized := strings.Replace(s, "-", "", -1)

	if hex.DecodedLen(len(normalized)) != size {
		return nil, errLength
	}

	bytes, err := hex.DecodeString(normalized)

	if err != nil {
		return nil, err
	}

	return UUID(bytes), nil
}

// The time section of the UUID in the UTC timezone
func (me UUID) Time() time.Time {
	nsec := me.Nanoseconds()
	return time.Unix(nsec/1e9, nsec%1e9).UTC()
}

// Returns the time_low, time_mid and time_hi sections of the UUID in 100
// nanosecond resolution since the unix Epoch.  Negative values indicate
// time prior to the gregorian epoch (00:00:00.00, 15 October 1582).
func (me UUID) Nanoseconds() int64 {
	time_low := uuidTime(binary.BigEndian.Uint32(me[0:4]))
	time_mid := uuidTime(binary.BigEndian.Uint16(me[4:6]))
	time_hi := uuidTime((binary.BigEndian.Uint16(me[6:8]) & 0x0fff))

	return toUnixNano((time_low) + (time_mid << 32) + (time_hi << 48))
}

/*
 The 4 bit version of the underlying UUID.

   Version  Description
      1     The time-based version specified in RFC4122.

      2     DCE Security version, with embedded POSIX UIDs.

      3     The name-based version specified in RFC4122
            that uses MD5 hashing.

      4     The randomly or pseudo- randomly generated version
            specified in RFC4122.

      5     The name-based version specified in RFC4122
            that uses SHA-1 hashing.
*/
func (me UUID) Version() int8 {
	return int8((binary.BigEndian.Uint16(me[6:8]) & 0xf000) >> 12)
}

/*
The 3 bit variant of the underlying UUID.

  Msb0  Msb1  Msb2  Description
   0     x     x    Reserved, NCS backward compatibility.
   1     0     x    The variant specified in RFC4122.
   1     1     0    Reserved, Microsoft Corporation backward compatibility
   1     1     1    Reserved for future definition.
*/
func (me UUID) Variant() int8 {
	return int8((binary.BigEndian.Uint16(me[8:10]) & 0xe000) >> 13)
}

// The timestamp in hex encoded form.
func (me UUID) String() string {
	return hex.EncodeToString(me[0:4]) + "-" +
		hex.EncodeToString(me[4:6]) + "-" +
		hex.EncodeToString(me[6:8]) + "-" +
		hex.EncodeToString(me[8:10]) + "-" +
		hex.EncodeToString(me[10:16])
}

// Stable comparison, first of the times then of the node values.
func (me UUID) Compare(other UUID) int {
	nsMe := me.Nanoseconds()
	nsOther := other.Nanoseconds()
	if nsMe > nsOther {
		return 1
	} else if nsMe < nsOther {
		return -1
	}
	return bytes.Compare(me[8:], other[8:])
}

// The underlying byte slice.  Treat the slice returned as immutable.
func (me UUID) Bytes() []byte {
	return me
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (me *UUID) UnmarshalJSON(b []byte) error {
	var field string
	if err := json.Unmarshal(b, &field); err != nil {
		return err
	}

	uuid, err := NewString(field)
	if err != nil {
		return err
	}

	*me = uuid

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (me UUID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + me.String() + `"`), nil
}

// Utility functions

func fromUnixNano(ns int64) uuidTime {
	return uuidTime((ns / 100) + gregorianEpoch)
}

func toUnixNano(t uuidTime) int64 {
	return int64((t - gregorianEpoch) * 100)
}
