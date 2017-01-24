package commands

import (
	"bytes"
	"fmt"
	"io"
	"log"
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
	"github.com/git-lfs/git-lfs/progress"
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
	cfg          = config.Config

	tqManifest *tq.Manifest
	apiClient  *lfsapi.Client
	global     sync.Mutex

	includeArg string
	excludeArg string
)

// getTransferManifest builds a tq.Manifest from the global os and git
// environments.
func getTransferManifest() *tq.Manifest {
	c := getAPIClient()

	global.Lock()
	defer global.Unlock()

	if tqManifest == nil {
		tqManifest = tq.NewManifestWithClient(c)
	}

	return tqManifest
}

func getAPIClient() *lfsapi.Client {
	global.Lock()
	defer global.Unlock()

	if apiClient == nil {
		c, err := lfsapi.NewClient(cfg.Os, cfg.Git)
		if err != nil {
			ExitWithError(err)
		}
		apiClient = c
	}
	return apiClient
}

func newLockClient(remote string) *locking.Client {
	lockClient, err := locking.NewClient(remote, getAPIClient())
	if err == nil {
		err = lockClient.SetupFileCache(filepath.Join(config.LocalGitStorageDir, "lfs"))
	}

	if err != nil {
		Exit("Unable to create lock system: %v", err.Error())
	}

	// Configure dirs
	lockClient.LocalWorkingDir = config.LocalWorkingDir
	lockClient.LocalGitDir = config.LocalGitDir

	return lockClient
}

// newDownloadCheckQueue builds a checking queue, checks that objects are there but doesn't download
func newDownloadCheckQueue(manifest *tq.Manifest, remote string, options ...tq.Option) *tq.TransferQueue {
	allOptions := make([]tq.Option, 0, len(options)+1)
	allOptions = append(allOptions, options...)
	allOptions = append(allOptions, tq.DryRun(true))
	return newDownloadQueue(manifest, remote, allOptions...)
}

// newDownloadQueue builds a DownloadQueue, allowing concurrent downloads.
func newDownloadQueue(manifest *tq.Manifest, remote string, options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Download, manifest, remote, options...)
}

// newUploadQueue builds an UploadQueue, allowing `workers` concurrent uploads.
func newUploadQueue(manifest *tq.Manifest, remote string, options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Upload, manifest, remote, options...)
}

func buildFilepathFilter(config *config.Configuration, includeArg, excludeArg *string) *filepathfilter.Filter {
	inc, exc := determineIncludeExcludePaths(config, includeArg, excludeArg)
	return filepathfilter.New(inc, exc)
}

func downloadTransfer(p *lfs.WrappedPointer) (name, path, oid string, size int64) {
	path, _ = lfs.LocalMediaPath(p.Oid)

	return p.Name, path, p.Oid, p.Size
}

func uploadTransfer(p *lfs.WrappedPointer) (*tq.Transfer, error) {
	filename := p.Name
	oid := p.Oid

	localMediaPath, err := lfs.LocalMediaPath(oid)
	if err != nil {
		return nil, errors.Wrapf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if len(filename) > 0 {
		if err = ensureFile(filename, localMediaPath); err != nil && !errors.IsCleanPointerError(err) {
			return nil, err
		}
	}

	return &tq.Transfer{
		Name: filename,
		Path: localMediaPath,
		Oid:  oid,
		Size: p.Size,
	}, nil
}

// ensureFile makes sure that the cleanPath exists before pushing it.  If it
// does not exist, it attempts to clean it by reading the file at smudgePath.
func ensureFile(smudgePath, cleanPath string) error {
	if _, err := os.Stat(cleanPath); err == nil {
		return nil
	}

	localPath := filepath.Join(config.LocalWorkingDir, smudgePath)
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	cleaned, err := lfs.PointerClean(file, file.Name(), stat.Size(), nil)
	if cleaned != nil {
		cleaned.Teardown()
	}

	if err != nil {
		return err
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
// Stderr. If an empty string is passed as the "format" arguemnt, only the
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
	if err := lfs.ClearTempObjects(); err != nil {
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
	if !lfs.InRepo() {
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
	var fmtWriter io.Writer = os.Stderr

	now := time.Now()
	name := now.Format("20060102T150405.999999999")
	full := filepath.Join(config.LocalLogDir, name+".log")

	if err := os.MkdirAll(config.LocalLogDir, 0755); err != nil {
		full = ""
		fmt.Fprintf(fmtWriter, "Unable to log panic to %s: %s\n\n", config.LocalLogDir, err.Error())
	} else if file, err := os.Create(full); err != nil {
		filename := full
		full = ""
		defer func() {
			fmt.Fprintf(fmtWriter, "Unable to log panic to %s\n\n", filename)
			logPanicToWriter(fmtWriter, err)
		}()
	} else {
		fmtWriter = file
		defer file.Close()
	}

	logPanicToWriter(fmtWriter, loggedError)

	return full
}

func logPanicToWriter(w io.Writer, loggedError error) {
	// log the version
	gitV, err := git.Config.Version()
	if err != nil {
		gitV = "Error getting git version: " + err.Error()
	}

	fmt.Fprintln(w, config.VersionDesc)
	fmt.Fprintln(w, gitV)

	// log the command that was run
	fmt.Fprintln(w)
	fmt.Fprintf(w, "$ %s", filepath.Base(os.Args[0]))
	if len(os.Args) > 0 {
		fmt.Fprintf(w, " %s", strings.Join(os.Args[1:], " "))
	}
	fmt.Fprintln(w)

	// log the error message and stack trace
	w.Write(ErrorBuffer.Bytes())
	fmt.Fprintln(w)

	fmt.Fprintf(w, "%s\n", loggedError)
	for _, stackline := range errors.StackTrace(loggedError) {
		fmt.Fprintln(w, stackline)
	}

	for key, val := range errors.Context(err) {
		fmt.Fprintf(w, "%s=%v\n", key, val)
	}

	fmt.Fprintln(w, "\nENV:")

	// log the environment
	for _, env := range lfs.Environ(cfg, getTransferManifest()) {
		fmt.Fprintln(w, env)
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

func buildProgressMeter(dryRun bool) *progress.ProgressMeter {
	return progress.NewMeter(
		progress.WithOSEnv(cfg.Os),
		progress.DryRun(dryRun),
	)
}

// isCommandEnabled returns whether the environment variable GITLFS<CMD>ENABLED
// is "truthy" according to config.Os.Bool (see
// github.com/git-lfs/git-lfs/config#Configuration.Env.Os), returning false
// by default if the enviornment variable is not specified.
//
// This function call should only guard commands that do not yet have stable
// APIs or solid server implementations.
func isCommandEnabled(cfg *config.Configuration, cmd string) bool {
	return cfg.Os.Bool(fmt.Sprintf("GITLFS%sENABLED", strings.ToUpper(cmd)), false)
}

func requireGitVersion() {
	minimumGit := "1.8.2"

	if !git.Config.IsGitVersionAtLeast(minimumGit) {
		gitver, err := git.Config.Version()
		if err != nil {
			Exit("Error getting git version: %s", err)
		}
		Exit("git version >= %s is required for Git LFS, your version: %s", minimumGit, gitver)
	}
}

func init() {
	log.SetOutput(ErrorWriter)
}
