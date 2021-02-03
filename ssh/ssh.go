package ssh

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

type SSHMetadata struct {
	UserAndHost string
	Port        string
	Path        string
}

func FormatArgs(cmd string, args []string, needShell bool) (string, []string) {
	if !needShell {
		return cmd, args
	}

	return subprocess.FormatForShellQuotedArgs(cmd, args)
}

func GetLFSExeAndArgs(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, operation, method string) (string, []string) {
	exe, args, needShell := GetExeAndArgs(osEnv, gitEnv, meta)
	args = append(args, fmt.Sprintf("git-lfs-authenticate %s %s", meta.Path, operation))
	exe, args = FormatArgs(exe, args, needShell)
	tracerx.Printf("run_command: %s %s", exe, strings.Join(args, " "))
	return exe, args
}

// Parse command, and if it looks like a valid command, return the ssh binary
// name, the command to run, and whether we need a shell.  If not, return
// existing as the ssh binary name.
func parseShellCommand(command string, existing string) (ssh string, cmd string, needShell bool) {
	ssh = existing
	if cmdArgs := tools.QuotedFields(command); len(cmdArgs) > 0 {
		needShell = true
		ssh = cmdArgs[0]
		cmd = command
	}
	return
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func GetExeAndArgs(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata) (exe string, baseargs []string, needShell bool) {
	var cmd string

	isPlink := false
	isTortoise := false

	ssh, _ := osEnv.Get("GIT_SSH")
	sshCmd, _ := osEnv.Get("GIT_SSH_COMMAND")
	ssh, cmd, needShell = parseShellCommand(sshCmd, ssh)

	if ssh == "" {
		sshCmd, _ := gitEnv.Get("core.sshcommand")
		ssh, cmd, needShell = parseShellCommand(sshCmd, defaultSSHCmd)
	}

	if cmd == "" {
		cmd = ssh
	}

	basessh := filepath.Base(ssh)

	if basessh != defaultSSHCmd {
		// Strip extension for easier comparison
		if ext := filepath.Ext(basessh); len(ext) > 0 {
			basessh = basessh[:len(basessh)-len(ext)]
		}
		isPlink = strings.EqualFold(basessh, "plink")
		isTortoise = strings.EqualFold(basessh, "tortoiseplink")
	}

	args := make([]string, 0, 7)

	if isTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(meta.Port) > 0 {
		if isPlink || isTortoise {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, meta.Port)
	}

	if sep, ok := sshSeparators[basessh]; ok {
		// inserts a separator between cli -options and host/cmd commands
		// example: $ ssh -p 12345 -- user@host.com git-lfs-authenticate ...
		args = append(args, sep, meta.UserAndHost)
	} else {
		// no prefix supported, strip leading - off host to prevent cmd like:
		// $ git config lfs.url ssh://-proxycmd=whatever
		// $ plink -P 12345 -proxycmd=foo git-lfs-authenticate ...
		//
		// Instead, it'll attempt this, and eventually return an error
		// $ plink -P 12345 proxycmd=foo git-lfs-authenticate ...
		args = append(args, sshOptPrefixRE.ReplaceAllString(meta.UserAndHost, ""))
	}

	return cmd, args, needShell
}

const defaultSSHCmd = "ssh"

var (
	sshOptPrefixRE = regexp.MustCompile(`\A\-+`)
	sshSeparators  = map[string]string{
		"ssh":          "--",
		"lfs-ssh-echo": "--", // used in lfs integration tests only
	}
)
