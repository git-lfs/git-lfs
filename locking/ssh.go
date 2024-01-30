package locking

import (
	"fmt"
	"strings"
	"time"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/git-lfs/git-lfs/v3/tr"
)

type sshLockClient struct {
	transfer *ssh.SSHTransfer
	*lfsapi.Client
}

func (c *sshLockClient) connection() (*ssh.PktlineConnection, error) {
	return c.transfer.Connection(0)
}

func (c *sshLockClient) parseLockResponse(status int, args []string, lines []string) (*Lock, string, error) {
	var lock *Lock
	var message string
	var err error
	seen := make(map[string]struct{})
	if status >= 200 && status <= 299 || status == 409 {
		lock = &Lock{}
		for _, entry := range args {
			if strings.HasPrefix(entry, "id=") {
				lock.Id = entry[3:]
				seen["id"] = struct{}{}
			} else if strings.HasPrefix(entry, "path=") {
				lock.Path = entry[5:]
				seen["path"] = struct{}{}
			} else if strings.HasPrefix(entry, "ownername=") {
				lock.Owner = &User{}
				lock.Owner.Name = entry[10:]
				seen["ownername"] = struct{}{}
			} else if strings.HasPrefix(entry, "locked-at=") {
				lock.LockedAt, err = time.Parse(time.RFC3339, entry[10:])
				if err != nil {
					return lock, "", errors.New(tr.Tr.Get("lock response: invalid locked-at: %s", entry))
				}
				seen["locked-at"] = struct{}{}
			}
		}
		if len(seen) != 4 {
			return nil, "", errors.New(tr.Tr.Get("incomplete fields for lock"))
		}
	}
	if status > 299 && len(lines) > 0 {
		message = lines[0]
	}
	return lock, message, nil
}

type owner string

const (
	ownerOurs    = owner("ours")
	ownerTheirs  = owner("theirs")
	ownerUnknown = owner("")
)

type lockData struct {
	lock Lock
	who  owner
}

func (c *sshLockClient) lockDataIsIncomplete(data *lockData) bool {
	return data.lock.Path == "" || data.lock.Owner == nil || data.lock.LockedAt.IsZero()
}

func (c *sshLockClient) parseListLockResponse(status int, args []string, lines []string) (all []Lock, ours []Lock, theirs []Lock, nextCursor string, message string, err error) {
	locks := make(map[string]*lockData)
	var last *lockData
	if status >= 200 && status <= 299 {
		for _, entry := range args {
			if strings.HasPrefix(entry, "next-cursor=") {
				if len(nextCursor) > 0 {
					return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: multiple next-cursor responses"))
				}
				nextCursor = entry[12:]
			}
		}
		for _, entry := range lines {
			values := strings.SplitN(entry, " ", 3)
			var cmd string
			if len(values) > 0 {
				cmd = values[0]
			}
			if cmd == "lock" {
				if len(values) != 2 {
					return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: invalid response: %q", entry))
				} else if last != nil && c.lockDataIsIncomplete(last) {
					return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: incomplete lock data"))
				}
				id := values[1]
				last = &lockData{who: ownerUnknown}
				last.lock.Id = id
				locks[id] = last
			} else if len(values) != 3 {
				return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: invalid response: %q", entry))
			} else if last == nil || last.lock.Id != values[1] {
				return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: interspersed response: %q", entry))
			} else {
				switch cmd {
				case "path":
					last.lock.Path = values[2]
				case "owner":
					last.who = owner(values[2])
				case "ownername":
					last.lock.Owner = &User{}
					last.lock.Owner.Name = values[2]
				case "locked-at":
					last.lock.LockedAt, err = time.Parse(time.RFC3339, values[2])
					if err != nil {
						return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: invalid locked-at: %s", entry))
					}
				}
			}
		}
		if last != nil && c.lockDataIsIncomplete(last) {
			return nil, nil, nil, "", "", errors.New(tr.Tr.Get("lock response: incomplete lock data"))
		}
		for _, lock := range locks {
			all = append(all, lock.lock)
			if lock.who == ownerOurs {
				ours = append(ours, lock.lock)
			} else if lock.who == ownerTheirs {
				theirs = append(theirs, lock.lock)
			}
		}
	} else if status > 299 && len(lines) > 0 {
		message = lines[0]
	}
	return all, ours, theirs, nextCursor, message, nil
}

func (c *sshLockClient) Lock(remote string, lockReq *lockRequest) (*lockResponse, int, error) {
	args := make([]string, 0, 3)
	args = append(args, fmt.Sprintf("path=%s", lockReq.Path))
	if lockReq.Ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", lockReq.Ref.Name))
	}
	conn, err := c.connection()
	if err != nil {
		return nil, 0, err
	}
	conn.Lock()
	defer conn.Unlock()
	err = conn.SendMessage("lock", args)
	if err != nil {
		return nil, 0, err
	}
	status, args, lines, err := conn.ReadStatusWithLines()
	if err != nil {
		return nil, status, err

	}
	var lock lockResponse
	lock.Lock, lock.Message, err = c.parseLockResponse(status, args, lines)
	return &lock, status, err
}

func (c *sshLockClient) Unlock(ref *git.Ref, remote, id string, force bool) (*unlockResponse, int, error) {
	args := make([]string, 0, 3)
	if ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", ref.Name))
	}
	conn, err := c.connection()
	if err != nil {
		return nil, 0, err
	}
	conn.Lock()
	defer conn.Unlock()
	err = conn.SendMessage(fmt.Sprintf("unlock %s", id), args)
	if err != nil {
		return nil, 0, err
	}
	status, args, lines, err := conn.ReadStatusWithLines()
	if err != nil {
		return nil, status, err

	}
	var lock unlockResponse
	lock.Lock, lock.Message, err = c.parseLockResponse(status, args, lines)
	return &lock, status, err
}

func (c *sshLockClient) Search(remote string, searchReq *lockSearchRequest) (*lockList, int, error) {
	values := searchReq.QueryValues()
	args := make([]string, 0, len(values))
	for key, value := range values {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}
	conn, err := c.connection()
	if err != nil {
		return nil, 0, err
	}
	conn.Lock()
	defer conn.Unlock()
	err = conn.SendMessage("list-lock", args)
	if err != nil {
		return nil, 0, err
	}
	status, args, lines, err := conn.ReadStatusWithLines()
	if err != nil {
		return nil, status, err
	}
	locks, _, _, nextCursor, message, err := c.parseListLockResponse(status, args, lines)
	if err != nil {
		return nil, status, err
	}
	list := &lockList{
		Locks:      locks,
		NextCursor: nextCursor,
		Message:    message,
	}
	return list, status, nil
}

func (c *sshLockClient) SearchVerifiable(remote string, vreq *lockVerifiableRequest) (*lockVerifiableList, int, error) {
	args := make([]string, 0, 3)
	if vreq.Ref != nil {
		args = append(args, fmt.Sprintf("refname=%s", vreq.Ref.Name))
	}
	if len(vreq.Cursor) > 0 {
		args = append(args, fmt.Sprintf("cursor=%s", vreq.Cursor))
	}
	if vreq.Limit > 0 {
		args = append(args, fmt.Sprintf("limit=%d", vreq.Limit))
	}
	conn, err := c.connection()
	if err != nil {
		return nil, 0, err
	}
	conn.Lock()
	defer conn.Unlock()
	err = conn.SendMessage("list-lock", args)
	if err != nil {
		return nil, 0, err
	}
	status, args, lines, err := conn.ReadStatusWithLines()
	if err != nil {
		return nil, status, err
	}
	_, ours, theirs, nextCursor, message, err := c.parseListLockResponse(status, args, lines)
	if err != nil {
		return nil, status, err
	}
	list := &lockVerifiableList{
		Ours:       ours,
		Theirs:     theirs,
		NextCursor: nextCursor,
		Message:    message,
	}
	return list, status, nil
}
