package pack

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// delayedObjectReader provides an interface for reading from an Object while
// loading object data into memory only on demand.  It implements io.ReadCloser.
type delayedObjectReader struct {
	obj *Object
	mr  io.Reader
}

// Read implements the io.Reader method by instantiating a new underlying reader
// only on demand.
func (d *delayedObjectReader) Read(b []byte) (int, error) {
	if d.mr == nil {
		data, err := d.obj.Unpack()
		if err != nil {
			return 0, err
		}
		d.mr = io.MultiReader(
			// Git object header:
			strings.NewReader(fmt.Sprintf("%s %d\x00",
				d.obj.Type(), len(data),
			)),

			// Git object (uncompressed) contents:
			bytes.NewReader(data),
		)
	}
	return d.mr.Read(b)
}

// Close implements the io.Closer interface.
func (d *delayedObjectReader) Close() error {
	return nil
}
