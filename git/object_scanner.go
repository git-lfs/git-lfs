package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/rubyist/tracerx"
)

// object represents a generic Git object of any type.
type object struct {
	// Contents reads Git's internal object representation.
	Contents *io.LimitedReader
	// Oid is the ID of the object.
	Oid string
	// Size is the size in bytes of the object.
	Size int64
	// Type is the type of the object being held.
	Type string
}

// ObjectScanner is a scanner type that scans for Git objects reference-able in
// Git's object database by their unique OID.
type ObjectScanner struct {
	// object is the object that the ObjectScanner last scanned, or nil.
	object *object
	// err is the error (if any) that the ObjectScanner encountered during
	// its last scan, or nil.
	err error

	// from is the buffered source of input to the *ObjectScanner. It
	// expects input in the form described by
	// https://git-scm.com/docs/git-cat-file.
	from *bufio.Reader
	// to is a writer which accepts the object's OID to be scanned.
	to io.Writer
	// closeFn is an optional function that is run before the ObjectScanner
	// is closed. It is designated to clean up and close any resources held
	// by the ObjectScanner during runtime.
	closeFn func() error
}

// NewObjectScanner constructs a new instance of the `*ObjectScanner` type and
// returns it. It backs the ObjectScanner with an invocation of the `git
// cat-file --batch` command. If any errors were encountered while starting that
// command, they will be returned immediately.
//
// Otherwise, an `*ObjectScanner` is returned with no error.
func NewObjectScanner() (*ObjectScanner, error) {
	cmd := gitNoLFS("cat-file", "--batch")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open stdout")
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open stdin")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open stderr")
	}

	closeFn := func() error {
		if err := stdin.Close(); err != nil {
			return err
		}

		msg, _ := ioutil.ReadAll(stderr)
		if err = cmd.Wait(); err != nil {
			return errors.Errorf("Error in git cat-file --batch: %v %v", err, string(msg))
		}

		return nil
	}

	tracerx.Printf("run_command: git cat-file --batch")
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &ObjectScanner{
		from: bufio.NewReaderSize(stdout, 16384),
		to:   stdin,

		closeFn: closeFn,
	}, nil
}

// NewObjectScannerFrom returns a new `*ObjectScanner` populated with data from
// the given `io.Reader`, "r". It supplies no close function, and discards any
// input given to the Scan() function.
func NewObjectScannerFrom(r io.Reader) *ObjectScanner {
	return &ObjectScanner{
		from: bufio.NewReader(r),
		to:   ioutil.Discard,
	}
}

// Scan scans for a particular object given by the "oid" parameter. Once the
// scan is complete, the Contents(), Sha1(), Size() and Type() functions may be
// called and will return data corresponding to the given OID.
//
// Scan() returns whether the scan was successful, or in other words, whether or
// not the scanner can continue to progress.
func (s *ObjectScanner) Scan(oid string) bool {
	if err := s.reset(); err != nil {
		s.err = err
		return false
	}

	obj, err := s.scan(oid)
	s.object = obj

	if err != nil {
		if err != io.EOF {
			s.err = err
		}
		return false
	}
	return true
}

// Close closes and frees any resources owned by the *ObjectScanner that it is
// called upon. If there were any errors in freeing that (those) resource(s), it
// it will be returned, otherwise nil.
func (s *ObjectScanner) Close() error {
	if s == nil {
		return nil
	}

	if s.closeFn != nil {
		return s.closeFn()
	}
	return nil
}

// Contents returns an io.Reader which reads Git's representation of the object
// that was last scanned for.
func (s *ObjectScanner) Contents() io.Reader {
	return s.object.Contents
}

// Sha1 returns the SHA1 object ID of the object that was last scanned for.
func (s *ObjectScanner) Sha1() string {
	return s.object.Oid
}

// Size returns the size in bytes of the object that was last scanned for.
func (s *ObjectScanner) Size() int64 {
	return s.object.Size
}

// Type returns the type of the object that was last scanned for.
func (s *ObjectScanner) Type() string {
	return s.object.Type
}

// Err returns the error (if any) that was encountered during the last Scan()
// operation.
func (s *ObjectScanner) Err() error { return s.err }

// reset resets the `*ObjectScanner` to scan again by advancing the reader (if
// necessary) and clearing both the object and error fields on the
// `*ObjectScanner` instance.
func (s *ObjectScanner) reset() error {
	if s.object != nil {
		if s.object.Contents != nil {
			remaining := s.object.Contents.N
			if _, err := io.CopyN(ioutil.Discard, s.object.Contents, remaining); err != nil {
				return errors.Wrap(err, "unwind contents")
			}
		}

		// Consume extra LF inserted by cat-file
		if _, err := s.from.ReadByte(); err != nil {
			return err
		}
	}

	s.object, s.err = nil, nil

	return nil
}

type missingErr struct {
	oid string
}

func (m *missingErr) Error() string {
	return fmt.Sprintf("missing object: %s", m.oid)
}

func IsMissingObject(err error) bool {
	_, ok := err.(*missingErr)
	return ok
}

// scan scans for and populates a new Git object given an OID.
func (s *ObjectScanner) scan(oid string) (*object, error) {
	if _, err := fmt.Fprintln(s.to, oid); err != nil {
		return nil, err
	}

	l, err := s.from.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	fields := bytes.Fields(l)
	switch len(fields) {
	case 2:
		if string(fields[1]) == "missing" {
			return nil, &missingErr{oid: oid}
		}
		break
	case 3:
		oid = string(fields[0])
		typ := string(fields[1])
		size, _ := strconv.Atoi(string(fields[2]))
		contents := io.LimitReader(s.from, int64(size))

		return &object{
			Contents: contents.(*io.LimitedReader),
			Oid:      oid,
			Size:     int64(size),
			Type:     typ,
		}, nil
	}
	return nil, errors.Errorf("invalid line: %q", l)
}
