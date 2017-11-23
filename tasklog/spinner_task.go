package tasklog

import (
	"fmt"
	"runtime"
	"time"
)

// Spinner implements tasklog.Task by logging messages next to a continuously
// updating spinner, alternating between the below characters to create a
// "spinning" effect.
type Spinner struct {
	// stage is zero-based index into the spinnerChars array below and
	// represents the next spinner characeter's index.
	stage int
	// msg is the last message that was sent by the spinner.
	msg string

	// updates is the channel of updates used to send spin messages to the
	// *tasklog.Logger.
	updates chan *Update
}

// spinnerChars is the list of characters that are used to simulate a "spinning"
// effect.
var spinnerChars = []byte{'|', '/', '-', '\\'}

// NewSpinner instantiates a new spinner insta.ce
func NewSpinner() *Spinner {
	return &Spinner{
		updates: make(chan *Update),
	}
}

// Updates implements tasklog.Task.Updates() and provides an unbuffered channel
// of updates to be logged.
func (s *Spinner) Updates() <-chan *Update {
	return s.updates
}

// Throttled implements tasklog.Task.Throttled() and returns false.
func (s *Spinner) Throttled() bool {
	return false
}

// Spinf sends a spin message advancing to the next character, and displaying
// the formatted (see: package 'fmt') message as given below.
func (s *Spinner) Spinf(fstr string, vs ...interface{}) {
	s.msg = fmt.Sprintf(fstr, vs...)

	s.Spin()
}

// Spin sends a spin message advancing to the next spin character, and using the
// last sent message.
func (s *Spinner) Spin() {
	s.spin(s.msg)
}

// Finish the spinner with a completion message.
func (s *Spinner) Finish(fstr string, vs ...interface{}) {
	var sym string
	if runtime.GOOS == "windows" {
		// Windows' console can't render UTF-8 check marks outside of
		// ConEmu, so fall-back to '*'.
		sym = "*"
	} else {
		sym = fmt.Sprintf("%c", '\u2714')
	}
	s.update(sym, fmt.Sprintf(fstr, vs...))

	close(s.updates)
}

// spin sends an update and advances the stage.
func (s *Spinner) spin(msg string) {
	s.update(string(spinnerChars[s.stage]), msg)
	s.stage = (s.stage + 1) % len(spinnerChars)
}

// update sends an update.
func (s *Spinner) update(sym, msg string) {
	s.updates <- &Update{
		S:  fmt.Sprintf("%s %s", sym, msg),
		At: time.Now(),
	}
}
