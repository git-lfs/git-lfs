package ssh

import (
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
)

type PktlineConnection struct {
	r   io.ReadCloser
	w   io.WriteCloser
	mu  sync.Mutex
	cmd *subprocess.Cmd
	pl  Pktline
}

func (conn *PktlineConnection) Lock() {
	conn.mu.Lock()
}

func (conn *PktlineConnection) Unlock() {
	conn.mu.Unlock()
}

func (conn *PktlineConnection) Start() error {
	conn.Lock()
	defer conn.Unlock()
	return conn.negotiateVersion()
}

func (conn *PktlineConnection) End() error {
	conn.Lock()
	defer conn.Unlock()
	err := conn.SendMessage("quit", nil)
	if err != nil {
		return err
	}
	_, err = conn.ReadStatus()
	conn.r.Close()
	conn.w.Close()
	conn.cmd.Wait()
	return err
}

func (conn *PktlineConnection) negotiateVersion() error {
	pkts, err := conn.pl.ReadPacketList()
	if err != nil {
		return errors.NewProtocolError(tr.Tr.Get("Unable to negotiate version with remote side (unable to read capabilities)"), err)
	}
	ok := false
	for _, line := range pkts {
		if line == "version=1" {
			ok = true
		}
	}
	if !ok {
		return errors.NewProtocolError(tr.Tr.Get("Unable to negotiate version with remote side (missing version=1)"), nil)
	}
	err = conn.SendMessage("version 1", nil)
	if err != nil {
		return errors.NewProtocolError(tr.Tr.Get("Unable to negotiate version with remote side (unable to send version)"), err)
	}
	status, args, _, err := conn.ReadStatusWithLines()
	if err != nil {
		return errors.NewProtocolError(tr.Tr.Get("Unable to negotiate version with remote side (unable to read status)"), err)
	}
	if status != 200 {
		text := tr.Tr.Get("no error provided")
		if len(args) > 0 {
			text = tr.Tr.Get("server said: %q", args[0])
		}
		return errors.NewProtocolError(tr.Tr.Get("Unable to negotiate version with remote side (unexpected status %d; %s)", status, text), nil)
	}
	return nil
}

func (conn *PktlineConnection) SendMessage(command string, args []string) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	for _, arg := range args {
		err = conn.pl.WritePacketText(arg)
		if err != nil {
			return err
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) SendMessageWithLines(command string, args []string, lines []string) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	for _, arg := range args {
		err = conn.pl.WritePacketText(arg)
		if err != nil {
			return err
		}
	}
	err = conn.pl.WriteDelim()
	if err != nil {
		return err
	}
	for _, line := range lines {
		err = conn.pl.WritePacketText(line)
		if err != nil {
			return err
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) SendMessageWithData(command string, args []string, data io.Reader) error {
	err := conn.pl.WritePacketText(command)
	if err != nil {
		return err
	}
	for _, arg := range args {
		err = conn.pl.WritePacketText(arg)
		if err != nil {
			return err
		}
	}
	err = conn.pl.WriteDelim()
	if err != nil {
		return err
	}
	buf := make([]byte, 32768)
	for {
		n, err := data.Read(buf)
		if n > 0 {
			err := conn.pl.WritePacket(buf[0:n])
			if err != nil {
				return err
			}
		}
		if err != nil {
			break
		}
	}
	return conn.pl.WriteFlush()
}

func (conn *PktlineConnection) ReadStatus() (int, error) {
	status := 0
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, errors.NewProtocolError(tr.Tr.Get("error reading packet"), err)
		}
		switch {
		case pktLen == 0:
			if !seenStatus {
				return 0, errors.NewProtocolError(tr.Tr.Get("no status seen"), nil)
			}
			return status, nil
		case !seenStatus:
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, errors.NewProtocolError(tr.Tr.Get("expected status line, got %q", s), err)
			}
			seenStatus = true
		default:
			return 0, errors.NewProtocolError(tr.Tr.Get("unexpected data, got %q", s), err)
		}
	}
}

// ReadStatusWithData reads a status, arguments, and any binary data.  Note that
// the reader must be fully exhausted before invoking any other read methods.
func (conn *PktlineConnection) ReadStatusWithData() (int, []string, io.Reader, error) {
	args := make([]string, 0, 100)
	status := 0
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("error reading packet"), err)
		}
		if pktLen == 0 {
			if !seenStatus {
				return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("no status seen"), nil)
			}
			return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("unexpected flush packet"), nil)
		} else if !seenStatus {
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("expected status line, got %q", s), err)
			}
			seenStatus = true
		} else if pktLen == 1 {
			break
		} else {
			args = append(args, s)
		}
	}

	return status, args, pktlineReader(conn.pl), nil
}

// ReadStatusWithLines reads a status, arguments, and a set of text lines.
func (conn *PktlineConnection) ReadStatusWithLines() (int, []string, []string, error) {
	args := make([]string, 0, 100)
	lines := make([]string, 0, 100)
	status := 0
	seenDelim := false
	seenStatus := false
	for {
		s, pktLen, err := conn.pl.ReadPacketTextWithLength()
		if err != nil {
			return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("error reading packet"), err)
		}
		switch {
		case pktLen == 0:
			if !seenStatus {
				return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("no status seen"), nil)
			}
			return status, args, lines, nil
		case seenDelim:
			lines = append(lines, s)
		case !seenStatus:
			ok := false
			if strings.HasPrefix(s, "status ") {
				status, err = strconv.Atoi(s[7:])
				ok = err == nil
			}
			if !ok {
				return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("expected status line, got %q", s), err)
			}
			seenStatus = true
		case pktLen == 1:
			if seenDelim {
				return 0, nil, nil, errors.NewProtocolError(tr.Tr.Get("unexpected delimiter packet"), nil)
			}
			seenDelim = true
		default:
			args = append(args, s)
		}
	}
}
