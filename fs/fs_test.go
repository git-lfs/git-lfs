package fs

import "testing"

func TestDecodeNone(t *testing.T) {
	evaluate(t, "A:\\some\\regular\\windows\\path", "A:\\some\\regular\\windows\\path")
}

func TestDecodeSingle(t *testing.T) {
	evaluate(t, "A:\\bl\\303\\204\\file.txt", "A:\\blÄ\\file.txt")
}

func TestDecodeMultiple(t *testing.T) {
	evaluate(t, "A:\\fo\\130\\file\\303\\261.txt", "A:\\fo\130\\file\303\261.txt")
}

func evaluate(t *testing.T, input string, expected string) {

	fs := Filesystem{}
	output := fs.DecodePathname(input)

	if output != expected {
		t.Errorf("Expecting same path, got: %s, want: %s.", output, expected)
	}

}
