package lfs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rubyist/tracerx"
)

type sshAuthResponse struct {
	Message   string            `json:"-"`
	Href      string            `json:"href"`
	Header    map[string]string `json:"header"`
	ExpiresAt string            `json:"expires_at"`
}

func sshAuthenticate(endpoint Endpoint, operation, oid string) (sshAuthResponse, error) {

	// This is only used as a fallback where the Git URL is SSH but server doesn't support a full SSH binary protocol
	// and therefore we derive a HTTPS endpoint for binaries instead; but check authentication here via SSH

	res := sshAuthResponse{}
	if len(endpoint.SshUserAndHost) == 0 {
		return res, nil
	}

	tracerx.Printf("ssh: %s git-lfs-authenticate %s %s %s",
		endpoint.SshUserAndHost, endpoint.SshPath, operation, oid)

	exe, args := sshGetExeAndArgs(endpoint)
	args = append(args,
		"git-lfs-authenticate",
		endpoint.SshPath,
		operation, oid)

	cmd := exec.Command(exe, args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		res.Message = string(out)
	} else {
		err = json.Unmarshal(out, &res)
	}

	return res, err
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func sshGetExeAndArgs(endpoint Endpoint) (exe string, baseargs []string) {
	if len(endpoint.SshUserAndHost) == 0 {
		return "", nil
	}

	ssh := os.Getenv("GIT_SSH")
	isPlink := strings.EqualFold(filepath.Base(ssh), "plink")
	isTortoise := strings.EqualFold(filepath.Base(ssh), "tortoiseplink")
	if ssh == "" {
		ssh = "ssh"
	}

	args := make([]string, 0, 4)
	if isTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(endpoint.SshPort) > 0 {
		if isPlink {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, endpoint.SshPort)
	}
	args = append(args, endpoint.SshUserAndHost)

	return ssh, args
}

// Below here is the pure-SSH API interface
// The API is basically the same except there's no need for hypermedia links
func NewSshApiContext(endpoint Endpoint) ApiContext {
	ctx := &SshApiContext{endpoint: endpoint}

	err := ctx.connect()
	if err != nil {
		// TODO - any way to log this? Seems only by returning errors & logging in commands package
		// not usable, discard
		ctx = nil
	}
	return ctx
}

type SshApiContext struct {
	// Endpoint which was used to open this connection
	endpoint Endpoint

	// The command which is running ssh
	cmd *exec.Cmd
	// Streams for communicating
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (self *SshApiContext) Endpoint() Endpoint {
	return self.endpoint
}

func (self *SshApiContext) Close() error {
	// Docs say "It is incorrect to call Wait before all writes to the pipe have completed."
	// But that actually means in parallel https://github.com/golang/go/issues/9307 so we're ok here
	errbytes, readerr := ioutil.ReadAll(self.stderr)
	if readerr == nil && len(errbytes) > 0 {
		// Copy to our stderr for info
		fmt.Fprintf(os.Stderr, "Messages from SSH server:\n%v", string(errbytes))
	}
	err := self.cmd.Wait()
	if err != nil {
		return fmt.Errorf("Error closing ssh connection: %v\nstderr: %v", err.Error(), string(errbytes))
	}
	self.stdin.Close()
	self.stdout.Close()
	self.stderr.Close()
	self.cmd = nil

	return nil
}

func (self *SshApiContext) connect() error {
	ssh, args := sshGetExeAndArgs(self.endpoint)

	// Now add remote program and path
	serverCommand := "git-lfs-serve"
	if c, ok := Config.GitConfig("lfs.sshserver"); ok {
		serverCommand = c
	}
	args = append(args, serverCommand)
	args = append(args, self.endpoint.SshPath)

	cmd := exec.Command(ssh, args...)

	outp, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stdout: %v", err.Error())
	}
	errp, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stderr: %v", err.Error())
	}
	inp, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Unable to connect to ssh stdin: %v", err.Error())
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Unable to start ssh command: %v", err.Error())
	}

	self.cmd = cmd
	self.stdin = inp
	self.stdout = outp
	self.stderr = errp

	return nil

}

func (self *SshApiContext) Download(oid string) (io.ReadCloser, int64, *WrappedError) {
	// TODO
	return nil, 0, nil
}
func (self *SshApiContext) Upload(oid string, sz int64, content io.Reader, cb CopyCallback) *WrappedError {
	// TODO
	return nil
}
