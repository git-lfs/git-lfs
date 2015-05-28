package contentaddressable

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"reflect"
	"runtime"
)

var supOid = "a2b71d6ee8997eb87b25ab42d566c44f6a32871752c7c73eb5578cb1182f7be0"

func TestFile(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)

	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	by, err := ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, 0, len(by))

	assertEqual(t, nil, aw.Accept())

	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, "SUP", string(by))

	assertEqual(t, nil, aw.Close())
}

func TestFileMismatch(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, "b2b71d6ee8997eb87b25ab42d566c44f6a32871752c7c73eb5578cb1182f7be0")
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)

	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	by, err := ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, 0, len(by))

	err = aw.Accept()
	if err == nil || !strings.Contains(err.Error(), "Content mismatch") {
		t.Errorf("Expected mismatch error: %s", err)
	}

	by, err = ioutil.ReadFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, "", string(by))

	assertEqual(t, nil, aw.Close())

	_, err = ioutil.ReadFile(filename)
	assertEqual(t, true, os.IsNotExist(err))
}

func TestFileCancel(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)

	n, err := aw.Write([]byte("SUP"))
	assertEqual(t, nil, err)
	assertEqual(t, 3, n)

	assertEqual(t, nil, aw.Close())

	for _, name := range []string{aw.filename, aw.tempFilename} {
		if _, err := os.Stat(name); err == nil {
			t.Errorf("%s exists?", name)
		}
	}
}

func TestFileLocks(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)
	assertEqual(t, filename, aw.filename)
	assertEqual(t, filename+"-temp", aw.tempFilename)

	files := []string{aw.filename, aw.tempFilename}

	for _, name := range files {
		if _, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0665); err == nil {
			t.Errorf("Able to open %s!", name)
		}
	}

	assertEqual(t, nil, aw.Close())

	for _, name := range files {
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0665)
		assertEqualf(t, nil, err, "unable to open %s: %s", name, err)
		cleanupFile(f)
	}
}

func TestFileDuel(t *testing.T) {
	test := SetupFile(t)
	defer test.Teardown()

	filename := filepath.Join(test.Path, supOid)
	aw, err := NewFile(filename)
	assertEqual(t, nil, err)
	defer aw.Close()

	if _, err := NewFile(filename); err == nil {
		t.Errorf("Expected a file open conflict!")
	}
}

func SetupFile(t *testing.T) *FileTest {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting wd: %s", err)
	}

	return &FileTest{filepath.Join(wd, "File"), t}
}

type FileTest struct {
	Path string
	*testing.T
}

func (t *FileTest) Teardown() {
	if err := os.RemoveAll(t.Path); err != nil {
		t.Fatalf("Error removing %s: %s", t.Path, err)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	checkAssertion(t, expected, actual, "")
}

func assertEqualf(t *testing.T, expected, actual interface{}, format string, args ...interface{}) {
	checkAssertion(t, expected, actual, format, args...)
}

func checkAssertion(t *testing.T, expected, actual interface{}, format string, args ...interface{}) {
	if expected == nil {
		if actual == nil {
			return
		}
	} else if reflect.DeepEqual(expected, actual) {
		return
	}

	_, file, line, _ := runtime.Caller(2) // assertEqual + checkAssertion
	t.Logf("%s:%d\nExpected: %v\nActual:   %v", file, line, expected, actual)
	if len(args) > 0 {
		t.Logf("! - "+format, args...)
	}
	t.FailNow()
}
