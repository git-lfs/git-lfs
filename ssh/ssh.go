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

type sshVariant string

const (
	variantSSH      = sshVariant("ssh")
	variantSimple   = sshVariant("simple")
	variantPutty    = sshVariant("putty")
	variantTortoise = sshVariant("tortoiseplink")
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

func GetLFSExeAndArgs(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata, command, operation string) (string, []string) {
	exe, args, needShell := GetExeAndArgs(osEnv, gitEnv, meta)
	args = append(args, fmt.Sprintf("%s %s %s", command, meta.Path, operation))
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

func findVariant(variant string) (bool, sshVariant) {
	switch variant {
	case "ssh", "simple", "putty", "tortoiseplink":
		return false, sshVariant(variant)
	case "plink":
		return false, variantPutty
	case "auto":
		return true, ""
	default:
		return false, variantSSH
	}
}

func autodetectVariant(osEnv config.Environment, gitEnv config.Environment, basessh string) sshVariant {
	if basessh != defaultSSHCmd {
		// Strip extension for easier comparison
		if ext := filepath.Ext(basessh); len(ext) > 0 {
			basessh = basessh[:len(basessh)-len(ext)]
		}
		if strings.EqualFold(basessh, "plink") {
			return variantPutty
		}
		if strings.EqualFold(basessh, "tortoiseplink") {
			return variantTortoise
		}
	}
	return "ssh"
}

func getVariant(osEnv config.Environment, gitEnv config.Environment, basessh string) sshVariant {
	variant, ok := osEnv.Get("GIT_SSH_VARIANT")
	if !ok {
		variant, ok = gitEnv.Get("ssh.variant")
	}
	autodetect, val := findVariant(variant)
	if ok && !autodetect {
		return val
	}
	return autodetectVariant(osEnv, gitEnv, basessh)
}

// Return the executable name for ssh on this machine and the base args
// Base args includes port settings, user/host, everything pre the command to execute
func GetExeAndArgs(osEnv config.Environment, gitEnv config.Environment, meta *SSHMetadata) (exe string, baseargs []string, needShell bool) {
	var cmd string

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
	variant := getVariant(osEnv, gitEnv, basessh)

	args := make([]string, 0, 7)

	if variant == variantTortoise {
		// TortoisePlink requires the -batch argument to behave like ssh/plink
		args = append(args, "-batch")
	}

	if len(meta.Port) > 0 {
		if variant == variantPutty || variant == variantTortoise {
			args = append(args, "-P")
		} else {
			args = append(args, "-p")
		}
		args = append(args, meta.Port)
	}

	if variant == variantSSH {
		// inserts a separator between cli -options and host/cmd commands
		// example: $ ssh -p 12345 -- user@host.com git-lfs-authenticate ...
		args = append(args, "--", meta.UserAndHost)
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
)
