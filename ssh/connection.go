package ssh

import (
	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/git-lfs/pktline"
)

type SSHTransfer struct {
	conn *PktlineConnection
}

func NewSSHTransfer(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation string) (*SSHTransfer, error) {
	exe, args := GetLFSExeAndArgs(osEnv, gitEnv, meta, "git-lfs-transfer", operation)
	cmd := subprocess.ExecCommand(exe, args...)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	w, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	var pl Pktline
	if osEnv.Bool("GIT_TRACE_PACKET", false) {
		pl = &TraceablePktline{pl: pktline.NewPktline(r, w)}
	} else {
		pl = pktline.NewPktline(r, w)
	}
	conn := &PktlineConnection{
		cmd: cmd,
		pl:  pl,
	}
	err = conn.Start()
	if err != nil {
		return nil, err
	}
	return &SSHTransfer{
		conn: conn,
	}, nil
}

func (tr *SSHTransfer) Connection() *PktlineConnection {
	return tr.conn
}
