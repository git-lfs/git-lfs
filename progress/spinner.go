package progress

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/olekukonko/ts"
)

// Indeterminate progress indicator 'spinner'
type Spinner struct {
	stage int
	msg   string
}

var spinnerChars = []byte{'|', '/', '-', '\\'}

// Print a spinner (stage) to out followed by msg (no linefeed)
func (s *Spinner) Print(out io.Writer, msg string) {
	s.msg = msg
	s.Spin(out)
}

// Just spin the spinner one more notch & use the last message
func (s *Spinner) Spin(out io.Writer) {
	s.stage = (s.stage + 1) % len(spinnerChars)
	s.update(out, string(spinnerChars[s.stage]), s.msg)
}

// Finish the spinner with a completion message & newline
func (s *Spinner) Finish(out io.Writer, finishMsg string) {
	s.msg = finishMsg
	s.stage = 0
	var sym string
	if runtime.GOOS == "windows" {
		// Windows console sucks, can't do nice check mark except in ConEmu (not cmd or git bash)
		// So play it safe & boring
		sym = "*"
	} else {
		sym = fmt.Sprintf("%c", '\u2714')
	}
	s.update(out, sym, finishMsg)
	out.Write([]byte{'\n'})
}

func (s *Spinner) update(out io.Writer, prefix, msg string) {

	str := fmt.Sprintf("%v %v", prefix, msg)

	width := 80 // default to 80 chars wide if ts.GetSize() fails
	size, err := ts.GetSize()
	if err == nil {
		width = size.Col()
	}
	padding := strings.Repeat(" ", maxInt(0, width-len(str)))

	fmt.Fprintf(out, "\r%v%v", str, padding)

}

func NewSpinner() *Spinner {
	return &Spinner{}
}

// maxInt returns the greater of two `int`s, "a", or "b". This function
// originally comes from `github.com/git-lfs/git-lfs/tools#MaxInt`, but would
// introduce an import cycle if depended on directly.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
