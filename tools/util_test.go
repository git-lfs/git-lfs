package tools

import (
	"io"
	"os"
	"testing"
)

func TestMethodExists(t *testing.T) {
	// testing following methods exist in all platform.
	_, _ = CheckCloneFileSupported(os.TempDir())
	_, _ = CloneFile(io.Writer(nil), io.Reader(nil))
	_, _ = CloneFileByPath("", "")
}
