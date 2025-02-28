package ssh

import (
	"bytes"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
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
	tracerx.Printf("spawning pure SSH connection (#%d)", id)
	var errbuf bytes.Buffer
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
	cmd.Stderr = &errbuf
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
		err = errors.Join(err, errors.New(tr.Tr.Get("Failed to connect to remote SSH server: %s", cmd.Stderr)))
		tracerx.Printf("pure SSH connection unsuccessful (#%d)", id)
	} else {
		tracerx.Printf("pure SSH connection successful (#%d)", id)
	}
	return conn, multiplexing, controlPath, err
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance or nil if there is no such item.
func (st *SSHTransfer) IsMultiplexingEnabled() bool {
	return st.multiplexing
}

// Connection returns the nth connection (starting from 0) in this transfer
// instance if it is initialized and otherwise initializes a new connection and
// saves it in the nth position.  In all cases, nil is returned with an error
// if n is greater than the maximum number of connections, including when
// the connection array itself is nil.
func (st *SSHTransfer) Connection(n int) (*PktlineConnection, error) {
	st.lock.RLock()
	if n >= len(st.conn) {
		st.lock.RUnlock()
		return nil, errors.New(tr.Tr.Get("pure SSH connection unavailable (#%d)", n))
	}
	if st.conn[n] != nil {
		defer st.lock.RUnlock()
		return st.conn[n], nil
	}
	st.lock.RUnlock()

	st.lock.Lock()
	defer st.lock.Unlock()
	if st.conn[n] != nil {
		return st.conn[n], nil
	}
	conn, _, err := st.spawnConnection(n)
	if err != nil {
		return nil, err
	}
	st.conn[n] = conn
	return conn, nil
}

// ConnectionCount returns the number of connections this object has.
func (st *SSHTransfer) ConnectionCount() int {
	st.lock.RLock()
	defer st.lock.RUnlock()
	return len(st.conn)
}

// SetConnectionCount sets the number of connections to the specified number.
func (st *SSHTransfer) SetConnectionCount(n int) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.setConnectionCount(n)
}

// SetConnectionCountAtLeast sets the number of connections to be not less than
// the specified number.
func (st *SSHTransfer) SetConnectionCountAtLeast(n int) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	count := len(st.conn)
	if n <= count {
		return nil
	}
	return st.setConnectionCount(n)
}

func (st *SSHTransfer) spawnConnection(n int) (*PktlineConnection, string, error) {
	conn, _, controlPath, err := startConnection(n, st.osEnv, st.gitEnv, st.meta, st.operation, st.controlPath)
	if err != nil {
		tracerx.Printf("failed to spawn pure SSH connection (#%d): %s", n, err)
		return nil, "", err
	}
	return conn, controlPath, err
}

func (st *SSHTransfer) setConnectionCount(n int) error {
	count := len(st.conn)
	if n < count {
		tn := n
		if tn == 0 {
			tn = 1
		}
		for i, item := range st.conn[tn:count] {
			if item == nil {
				tracerx.Printf("skipping uninitialized lazy pure SSH connection (#%d) (resetting total from %d to %d)", i, count, n)
				continue
			}
			tracerx.Printf("terminating pure SSH connection (#%d) (resetting total from %d to %d)", tn+i, count, n)
			if err := item.End(); err != nil {
				return err
			}
		}
		st.conn = st.conn[0:tn]
	} else if n > count {
		for i := count; i < n; i++ {
			if i == 0 {
				conn, controlPath, err := st.spawnConnection(i)
				if err != nil {
					return err
				}
				st.conn = append(st.conn, conn)
				st.controlPath = controlPath
			} else {
				st.conn = append(st.conn, nil)
			}
		}
	}
	if n == 0 && count > 0 {
		tracerx.Printf("terminating pure SSH connection (#0) (resetting total from %d to %d)", count, n)
		if err := st.conn[0].End(); err != nil {
			return err
		}
		st.conn = nil
		st.controlPath = ""
	}
	return nil
}

func (st *SSHTransfer) Shutdown() error {
	tracerx.Printf("shutting down pure SSH connections")
	return st.SetConnectionCount(0)
}
