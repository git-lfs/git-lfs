package ssh

import (
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/pktline"
)

type SSHTransfer struct {
	lock         *sync.RWMutex
	conn         []*PktlineConnection
	osEnv        config.Environment
	gitEnv       config.Environment
	meta         *SSHMetadata
	operation    string
	multiplexing bool
}

func NewSSHTransfer(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation string) (*SSHTransfer, error) {
	conn, multiplexing, err := startConnection(0, osEnv, gitEnv, meta, operation)
	if err != nil {
		return nil, err
	}
	return &SSHTransfer{
		lock:         &sync.RWMutex{},
		osEnv:        osEnv,
		gitEnv:       gitEnv,
		meta:         meta,
		operation:    operation,
		multiplexing: multiplexing,
		conn:         []*PktlineConnection{conn},
	}, nil
}

func startConnection(id int, osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation string) (*PktlineConnection, bool, error) {
	exe, args, multiplexing := GetLFSExeAndArgs(osEnv, gitEnv, meta, "git-lfs-transfer", operation, true)
	cmd, err := subprocess.ExecCommand(exe, args...)
	if err != nil {
		return nil, false, err
	}
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, false, err
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, false, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, false, err
	}

	var pl Pktline
	if osEnv.Bool("GIT_TRACE_PACKET", false) {
		pl = &TraceablePktline{id: id, pl: pktline.NewPktline(r, w)}
	} else {
		pl = pktline.NewPktline(r, w)
	}
	conn := &PktlineConnection{
		cmd: cmd,
		pl:  pl,
		r:   r,
		w:   w,
	}
	err = conn.Start()
	if err != nil {
		r.Close()
		w.Close()
		cmd.Wait()
	}
	return conn, multiplexing, err
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance or nil if there is no such item.
func (tr *SSHTransfer) IsMultiplexingEnabled() bool {
	return tr.multiplexing
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance or nil if there is no such item.
func (tr *SSHTransfer) Connection(n int) *PktlineConnection {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	if n >= len(tr.conn) {
		return nil
	}
	return tr.conn[n]
}

// ConnectionCount returns the number of connections this object has.
func (tr *SSHTransfer) ConnectionCount() int {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return len(tr.conn)
}

// SetConnectionCount sets the number of connections to the specified number.
func (tr *SSHTransfer) SetConnectionCount(n int) error {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	return tr.setConnectionCount(n)
}

// SetConnectionCountAtLeast sets the number of connections to be not less than
// the specified number.
func (tr *SSHTransfer) SetConnectionCountAtLeast(n int) error {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	count := len(tr.conn)
	if n <= count {
		return nil
	}
	return tr.setConnectionCount(n)
}

func (tr *SSHTransfer) setConnectionCount(n int) error {
	count := len(tr.conn)
	if n < count {
		for _, item := range tr.conn[n:count] {
			if err := item.End(); err != nil {
				return err
			}
		}
		tr.conn = tr.conn[0:n]
	} else if n > count {
		for i := count; i < n; i++ {
			conn, _, err := startConnection(i, tr.osEnv, tr.gitEnv, tr.meta, tr.operation)
			if err != nil {
				return err
			}
			tr.conn = append(tr.conn, conn)
		}
	}
	return nil
}

func (tr *SSHTransfer) Shutdown() error {
	return tr.SetConnectionCount(0)
}
