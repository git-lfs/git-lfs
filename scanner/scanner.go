package scanner

import (
	"bytes"
	"github.com/github/git-media/git"
	"github.com/github/git-media/pointer"
	"github.com/rubyist/tracerx"
	"strconv"
	"time"
)

type ScannedPointer struct {
	Name string
	*pointer.Pointer
}

func Scan(ref string) ([]*ScannedPointer, error) {
	fileNameMap := make(map[string]string, 0)

	// Gets all objects git knows about
	var buf bytes.Buffer
	t := time.Now()
	objects, _ := git.RevListObjects(ref, "", ref == "")
	tracerx.PerformanceSince("rev-list --objects", t)

	for _, o := range objects {
		fileNameMap[o.Sha1] = o.Name
		buf.WriteString(o.Sha1 + "\n")
	}

	// Get type and size info for all objects
	t = time.Now()
	objects, _ = git.CatFileBatchCheck(&buf)
	tracerx.PerformanceSince("cat-file --batch-check", t)

	// Pull out git objects that are type blob and size < 200 bytes.
	// These are the likely git media pointer files
	var mediaObjects bytes.Buffer
	for _, o := range objects {
		if o.Type == "blob" && o.Size < 200 {
			mediaObjects.WriteString(o.Sha1 + "\n")
		}
	}

	// Take all of the git media shas and pull out the pointer file contents
	// It comes out of here in the format:
	// <sha1> <type> <size><LF>
	// <contents><LF>
	// This string contains all the data, so we parse it out below
	t = time.Now()
	data, _ := git.CatFileBatch(&mediaObjects)
	tracerx.PerformanceSince("cat-file --batch", t)

	r := bytes.NewBufferString(data)

	pointers := make([]*ScannedPointer, 0)
	t = time.Now()
	for {
		l, err := r.ReadBytes('\n')
		if err != nil { // Probably check for EOF
			break
		}

		fields := bytes.Fields(l)
		s, _ := strconv.Atoi(string(fields[2]))

		nbuf := make([]byte, s)
		_, err = r.Read(nbuf)
		if err != nil {
			return nil, err // Legit errors
		}

		sha1 := string(fields[0])
		name := fileNameMap[sha1]

		p, err := pointer.Decode(bytes.NewBuffer(nbuf))
		if err == nil {
			pointers = append(pointers, &ScannedPointer{name, p})
		}

		_, err = r.ReadBytes('\n') // Extra \n inserted by cat-file
		if err != nil {            // Probably check for EOF
			break
		}
	}
	tracerx.PerformanceSince("decode pointers", t)
	return pointers, nil
}
