package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer func() {
		reader.Close()
	}()

	os.Stdout = writer
	fn()
	writer.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	return buf.String()
}

func TestPrintHelpTreatsShortHelpFlagAsRootCommand(t *testing.T) {
	original, hadOriginal := ManPages["git-lfs"]
	ManPages["git-lfs"] = "root help output"
	t.Cleanup(func() {
		if hadOriginal {
			ManPages["git-lfs"] = original
		} else {
			delete(ManPages, "git-lfs")
		}
	})

	output := captureStdout(t, func() {
		printHelp("-h")
	})

	if !strings.Contains(output, "root help output") {
		t.Fatalf("expected root help output, got %q", output)
	}
}
