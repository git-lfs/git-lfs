package tlog

import (
	"fmt"
	"runtime"
	"time"
)

// Indeterminate progress indicator 'spinner'
type Spinner struct {
	stage int
	msg   string

	updates chan *Update
}

var spinnerChars = []byte{'|', '/', '-', '\\'}

func NewSpinner() *Spinner {
	return &Spinner{
		updates: make(chan *Update),
	}
}

func (s *Spinner) Updates() <-chan *Update {
	return s.updates
}

func (s *Spinner) Throttled() bool {
	return false
}

func (s *Spinner) Spinf(fstr string, vs ...interface{}) {
	s.msg = fmt.Sprintf(fstr, vs...)
	s.spin(s.msg)
}

func (s *Spinner) Spin() {
	s.spin(s.msg)
}

// Finish the spinner with a completion message & newline
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

func (s *Spinner) spin(msg string) {
	s.stage = (s.stage + 1) % len(spinnerChars)
	s.update(string(spinnerChars[s.stage]), msg)
}

func (s *Spinner) update(sym, msg string) {
	s.updates <- &Update{
		S:  fmt.Sprintf("%s %s", sym, msg),
		At: time.Now(),
	}
}
