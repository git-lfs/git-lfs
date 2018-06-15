package commands

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/locking"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
)

// Populate man pages
//go:generate go run ../docs/man/mangen.go

var (
	Debugging    = false
	ErrorBuffer  = &bytes.Buffer{}
	ErrorWriter  = io.MultiWriter(os.Stderr, ErrorBuffer)
	OutputWriter = io.MultiWriter(os.Stdout, ErrorBuffer)
	ManPages     = make(map[string]string, 20)
	tqManifest   = make(map[string]*tq.Manifest)

	cfg       *config.Configuration
	apiClient *lfsapi.Client
	global    sync.Mutex

	includeArg string
	excludeArg string
)

// getTransferManifest builds a tq.Manifest from the global os and git
// environments.
func getTransferManifest() *tq.Manifest {
	return getTransferManifestOperationRemote("", "")
}

// getTransferManifestOperationRemote builds a tq.Manifest from the global os
// and git environments and operation-specific and remote-specific settings.
// Operation must be "download", "upload", or the empty string.
func getTransferManifestOperationRemote(operation, remote string) *tq.Manifest {
	c := getAPIClient()

	global.Lock()
	defer global.Unlock()

	k := fmt.Sprintf("%s.%s", operation, remote)
	if tqManifest[k] == nil {
		tqManifest[k] = tq.NewManifest(cfg.Filesystem(), c, operation, remote)
	}

	return tqManifest[k]
}

func getAPIClient() *lfsapi.Client {
	global.Lock()
	defer global.Unlock()

	if apiClient == nil {
		c, err := lfsapi.NewClient(cfg)
		if err != nil {
			ExitWithError(err)
		}
		apiClient = c
	}
	return apiClient
}

func closeAPIClient() error {
	global.Lock()
	defer global.Unlock()
	if apiClient == nil {
		return nil
	}
	return apiClient.Close()
}

func newLockClient() *locking.Client {
	lockClient, err := locking.NewClient(cfg.PushRemote(), getAPIClient())
	if err == nil {
		os.MkdirAll(cfg.LFSStorageDir(), 0755)
		err = lockClient.SetupFileCache(cfg.LFSStorageDir())
	}

	if err != nil {
		Exit("Unable to create lock system: %v", err.Error())
	}

	// Configure dirs
	lockClient.LocalWorkingDir = cfg.LocalWorkingDir()
	lockClient.LocalGitDir = cfg.LocalGitDir()
	lockClient.SetLockableFilesReadOnly = cfg.SetLockableFilesReadOnly()

	return lockClient
}

// newDownloadCheckQueue builds a checking queue, checks that objects are there but doesn't download
func newDownloadCheckQueue(manifest *tq.Manifest, remote string, options ...tq.Option) *tq.TransferQueue {
	return newDownloadQueue(manifest, remote, append(options,
		tq.DryRun(true),
	)...)
}

// newDownloadQueue builds a DownloadQueue, allowing concurrent downloads.
func newDownloadQueue(manifest *tq.Manifest, remote string, options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Download, manifest, remote, append(options,
		tq.RemoteRef(currentRemoteRef()),
	)...)
}

func currentRemoteRef() *git.Ref {
	return git.NewRefUpdate(cfg.Git, cfg.PushRemote(), cfg.CurrentRef(), nil).Right()
}

func buildFilepathFilter(config *config.Configuration, includeArg, excludeArg *string) *filepathfilter.Filter {
	inc, exc := determineIncludeExcludePaths(config, includeArg, excludeArg)
	return filepathfilter.New(inc, exc)
}

func downloadTransfer(p *lfs.WrappedPointer) (name, path, oid string, size int64) {
	path, _ = cfg.Filesystem().ObjectPath(p.Oid)
	return p.Name, path, p.Oid, p.Size
}

// Get user-readable manual install steps for hooks
func getHookInstallSteps() string {
	hooks := lfs.LoadHooks(cfg.HookDir())
	steps := make([]string, 0, len(hooks))
	for _, h := range hooks {
		steps = append(steps, fmt.Sprintf(
			"Add the following to .git/hooks/%s:\n\n%s",
			h.Type, tools.Indent(h.Contents)))
	}

	return strings.Join(steps, "\n\n")
}

func installHooks(force bool) error {
	hooks := lfs.LoadHooks(cfg.HookDir())
	for _, h := range hooks {
		if err := h.Install(force); err != nil {
			return err
		}
	}

	return nil
}

// uninstallHooks removes all hooks in range of the `hooks` var.
func uninstallHooks() error {
	if !cfg.InRepo() {
		return errors.New("Not in a git repository")
	}

	hooks := lfs.LoadHooks(cfg.HookDir())
	for _, h := range hooks {
		if err := h.Uninstall(); err != nil {
			return err
		}
	}

	return nil
}

// Error prints a formatted message to Stderr.  It also gets printed to the
// panic log if one is created for this command.
func Error(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprintln(ErrorWriter, format)
		return
	}
	fmt.Fprintf(ErrorWriter, format+"\n", args...)
}

// Print prints a formatted message to Stdout.  It also gets printed to the
// panic log if one is created for this command.
func Print(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprintln(OutputWriter, format)
		return
	}
	fmt.Fprintf(OutputWriter, format+"\n", args...)
}

// Exit prints a formatted message and exits.
func Exit(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(2)
}

// ExitWithError either panics with a full stack trace for fatal errors, or
// simply prints the error message and exits immediately.
func ExitWithError(err error) {
	errorWith(err, Panic, Exit)
}

// FullError prints either a full stack trace for fatal errors, or just the
// error message.
func FullError(err error) {
	errorWith(err, LoggedError, Error)
}

func errorWith(err error, fatalErrFn func(error, string, ...interface{}), errFn func(string, ...interface{})) {
	if Debugging || errors.IsFatalError(err) {
		fatalErrFn(err, "%s", err)
		return
	}

	errFn("%s", err)
}

// Debug prints a formatted message if debugging is enabled.  The formatted
// message also shows up in the panic log, if created.
func Debug(format string, args ...interface{}) {
	if !Debugging {
		return
	}
	log.Printf(format, args...)
}

// LoggedError prints the given message formatted with its arguments (if any) to
// Stderr. If an empty string is passed as the "format" argument, only the
// standard error logging message will be printed, and the error's body will be
// omitted.
//
// It also writes a stack trace for the error to a log file without exiting.
func LoggedError(err error, format string, args ...interface{}) {
	if len(format) > 0 {
		Error(format, args...)
	}
	file := handlePanic(err)

	if len(file) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors logged to %s\nUse `git lfs logs last` to view the log.\n", file)
	}
}

// Panic prints a formatted message, and writes a stack trace for the error to
// a log file before exiting.
func Panic(err error, format string, args ...interface{}) {
	LoggedError(err, format, args...)
	os.Exit(2)
}

func Cleanup() {
	if err := cfg.Cleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing old temp files: %s\n", err)
	}
}

func PipeMediaCommand(name string, args ...string) error {
	return PipeCommand("bin/"+name, args...)
}

func PipeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func requireStdin(msg string) {
	var out string

	stat, err := os.Stdin.Stat()
	if err != nil {
		out = fmt.Sprintf("Cannot read from STDIN. %s (%s)", msg, err)
	} else if (stat.Mode() & os.ModeCharDevice) != 0 {
		out = fmt.Sprintf("Cannot read from STDIN. %s", msg)
	}

	if len(out) > 0 {
		Error(out)
		os.Exit(1)
	}
}

func requireInRepo() {
	if !cfg.InRepo() {
		Print("Not in a git repository.")
		os.Exit(128)
	}
}

func handlePanic(err error) string {
	if err == nil {
		return ""
	}

	return logPanic(err)
}

func logPanic(loggedError error) string {
	var (
		fmtWriter  io.Writer = os.Stderr
		lineEnding string    = "\n"
	)

	now := time.Now()
	name := now.Format("20060102T150405.999999999")
	full := filepath.Join(cfg.LocalLogDir(), name+".log")

	if err := os.MkdirAll(cfg.LocalLogDir(), 0755); err != nil {
		full = ""
		fmt.Fprintf(fmtWriter, "Unable to log panic to %s: %s\n\n", cfg.LocalLogDir(), err.Error())
	} else if file, err := os.Create(full); err != nil {
		filename := full
		full = ""
		defer func() {
			fmt.Fprintf(fmtWriter, "Unable to log panic to %s\n\n", filename)
			logPanicToWriter(fmtWriter, err, lineEnding)
		}()
	} else {
		fmtWriter = file
		lineEnding = gitLineEnding(cfg.Git)
		defer file.Close()
	}

	logPanicToWriter(fmtWriter, loggedError, lineEnding)

	return full
}

func ipAddresses() []string {
	ips := make([]string, 0, 1)
	ifaces, err := net.Interfaces()
	if err != nil {
		ips = append(ips, "Error getting network interface: "+err.Error())
		return ips
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, _ := i.Addrs()
		l := make([]string, 0, 1)
		if err != nil {
			ips = append(ips, "Error getting IP address: "+err.Error())
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			l = append(l, ip.String())
		}
		if len(l) > 0 {
			ips = append(ips, strings.Join(l, " "))
		}
	}
	return ips
}

func logPanicToWriter(w io.Writer, loggedError error, le string) {
	// log the version
	gitV, err := git.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	fmt.Fprint(w, config.VersionDesc+le)
	fmt.Fprint(w, gitV+le)

	// log the command that was run
	fmt.Fprint(w, le)
	fmt.Fprintf(w, "$ %s", filepath.Base(os.Args[0]))
	if len(os.Args) > 0 {
		fmt.Fprintf(w, " %s", strings.Join(os.Args[1:], " "))
	}
	fmt.Fprint(w, le)

	// log the error message and stack trace
	w.Write(ErrorBuffer.Bytes())
	fmt.Fprint(w, le)

	fmt.Fprintf(w, "%+v"+le, loggedError)

	for key, val := range errors.Context(err) {
		fmt.Fprintf(w, "%s=%v"+le, key, val)
	}

	fmt.Fprint(w, le+"Current time in UTC: "+le)
	fmt.Fprint(w, time.Now().UTC().Format("2006-01-02 15:04:05")+le)

	fmt.Fprint(w, le+"ENV:"+le)

	// log the environment
	for _, env := range lfs.Environ(cfg, getTransferManifest()) {
		fmt.Fprint(w, env+le)
	}

	fmt.Fprint(w, le+"Client IP addresses:"+le)

	for _, ip := range ipAddresses() {
		fmt.Fprint(w, ip+le)
	}
}

func determineIncludeExcludePaths(config *config.Configuration, includeArg, excludeArg *string) (include, exclude []string) {
	if includeArg == nil {
		include = config.FetchIncludePaths()
	} else {
		include = tools.CleanPaths(*includeArg, ",")
	}
	if excludeArg == nil {
		exclude = config.FetchExcludePaths()
	} else {
		exclude = tools.CleanPaths(*excludeArg, ",")
	}
	return
}

func buildProgressMeter(dryRun bool, d tq.Direction) *tq.Meter {
	m := tq.NewMeter()
	m.Logger = m.LoggerFromEnv(cfg.Os)
	m.DryRun = dryRun
	m.Direction = d
	return m
}

func requireGitVersion() {
	minimumGit := "1.8.2"

	if !git.IsGitVersionAtLeast(minimumGit) {
		gitver, err := git.Version()
		if err != nil {
			Exit("Error getting git version: %s", err)
		}
		Exit("git version >= %s is required for Git LFS, your version: %s", minimumGit, gitver)
	}
}
