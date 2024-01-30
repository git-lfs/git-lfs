package ssh

import (
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/pktline"
	"github.com/rubyist/tracerx"
)

type SSHTransfer struct {
	lock         *sync.RWMutex
	conn         []*PktlineConnection
	osEnv        config.Environment
	gitEnv       config.Environment
	meta         *SSHMetadata
	operation    string
	multiplexing bool
	controlPath  string
}

func NewSSHTransfer(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation string) (*SSHTransfer, error) {
	conn, multiplexing, controlPath, err := startConnection(0, osEnv, gitEnv, meta, operation, "")
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
		controlPath:  controlPath,
		conn:         []*PktlineConnection{conn},
	}, nil
}

func startConnection(id int, osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation string, multiplexControlPath string) (conn *PktlineConnection, multiplexing bool, controlPath string, err error) {
	tracerx.Printf("spawning pure SSH connection")
	exe, args, multiplexing, controlPath := GetLFSExeAndArgs(osEnv, gitEnv, meta, "git-lfs-transfer", operation, true, multiplexControlPath)
	cmd, err := subprocess.ExecCommand(exe, args...)
	if err != nil {
		return nil, false, "", err
	}
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, false, "", err
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, false, "", err
	}
	err = cmd.Start()
	if err != nil {
		return nil, false, "", err
	}

	var pl Pktline
	if osEnv.Bool("GIT_TRACE_PACKET", false) {
		pl = &TraceablePktline{id: id, pl: pktline.NewPktline(r, w)}
	} else {
		pl = pktline.NewPktline(r, w)
	}
	conn = &PktlineConnection{
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
	tracerx.Printf("pure SSH connection successful")
	return conn, multiplexing, controlPath, err
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance or nil if there is no such item.
func (tr *SSHTransfer) IsMultiplexingEnabled() bool {
	return tr.multiplexing
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance if it is initialized and otherwise initializes a new connection and
// saves it in the nth position.  In all cases, nil is returned if n is greater
// than the maximum number of connections.
func (tr *SSHTransfer) Connection(n int) (*PktlineConnection, error) {
	tr.lock.RLock()
	if n >= len(tr.conn) {
		tr.lock.RUnlock()
		return nil, nil
	}
	if tr.conn[n] != nil {
		defer tr.lock.RUnlock()
		return tr.conn[n], nil
	}
	tr.lock.RUnlock()

	tr.lock.Lock()
	defer tr.lock.Unlock()
	if tr.conn[n] != nil {
		return tr.conn[n], nil
	}
	conn, _, err := tr.spawnConnection(n)
	return conn, err
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

func (tr *SSHTransfer) spawnConnection(n int) (*PktlineConnection, string, error) {
	conn, _, controlPath, err := startConnection(n, tr.osEnv, tr.gitEnv, tr.meta, tr.operation, tr.controlPath)
	if err != nil {
		tracerx.Printf("failed to spawn pure SSH connection: %s", err)
		return nil, "", err
	}
	return conn, controlPath, err
}

func (tr *SSHTransfer) setConnectionCount(n int) error {
	count := len(tr.conn)
	if n < count {
		tn := n
		if tn == 0 {
			tn = 1
		}
		for _, item := range tr.conn[tn:count] {
			if item == nil {
				tracerx.Printf("skipping uninitialized lazy pure SSH connection (%d -> %d)", count, n)
				continue
			}
			tracerx.Printf("terminating pure SSH connection (%d -> %d)", count, n)
			if err := item.End(); err != nil {
				return err
			}
		}
		tr.conn = tr.conn[0:tn]
	} else if n > count {
		for i := count; i < n; i++ {
			if i == 0 {
				conn, controlPath, err := tr.spawnConnection(i)
				if err != nil {
					return err
				}
				tr.conn = append(tr.conn, conn)
				tr.controlPath = controlPath
			} else {
				tr.conn = append(tr.conn, nil)
			}
		}
	}
	if n == 0 && count > 0 {
		tracerx.Printf("terminating pure SSH connection (%d -> %d)", count, n)
		if err := tr.conn[0].End(); err != nil {
			return err
		}
		tr.conn = nil
		tr.controlPath = ""
	}
	return nil
}

func (tr *SSHTransfer) Shutdown() error {
	tracerx.Printf("shutting down pure SSH connection")
	return tr.SetConnectionCount(0)
}
