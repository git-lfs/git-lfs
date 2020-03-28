package standalone

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/subprocess"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/rubyist/tracerx"
)

// inputMessage represents a message from Git LFS to the standalone transfer
// agent. Not all fields will be filled in on all requests.
type inputMessage struct {
	Event     string `json:"event"`
	Operation string `json:"operation"`
	Remote    string `json:"remote"`
	Oid       string `json:"oid"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
}

// errorMessage represents an optional error message that may occur in a
// completion response.
type errorMessage struct {
	Message string `json:"message"`
}

// completeMessage represents a completion response.
type completeMessage struct {
	Event string        `json:"event"`
	Oid   string        `json:"oid"`
	Path  string        `json:"path,omitempty"`
	Error *errorMessage `json:"error,omitempty"`
}

type fileHandler struct {
	remotePath   string
	remoteConfig *config.Configuration
	output       *os.File
	config       *config.Configuration
}

// fileUrlFromRemote looks up the URL depending on the remote. The remote can be
// a literal URL or the name of a remote.
//
// In this situation, we only accept file URLs.
func fileUrlFromRemote(cfg *config.Configuration, name string, direction string) (*url.URL, error) {
	if strings.HasPrefix(name, "file://") {
		if url, err := url.Parse(name); err == nil {
			return url, nil
		}
	}

	apiClient, err := lfsapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	for _, remote := range cfg.Remotes() {
		if remote != name {
			continue
		}
		remoteEndpoint := apiClient.Endpoints.Endpoint(direction, remote)
		if !strings.HasPrefix(remoteEndpoint.Url, "file://") {
			return nil, nil
		}
		return url.Parse(remoteEndpoint.Url)
	}
	return nil, nil
}

// gitDirAtPath finds the .git directory corresponding to the given path, which
// may be the .git directory itself, the working tree, or the root of a bare
// repository.
//
// We filter out the GIT_DIR environment variable to ensure we get the expected
// result, and we change directories to ensure that we can make use of
// filepath.Abs. Using --absolute-git-dir instead of --git-dir is not an option
// because we support Git versions that don't have --absolute-git-dir.
func gitDirAtPath(path string) (string, error) {
	// Filter out all the GIT_* environment variables.
	env := os.Environ()
	n := 0
	for _, val := range env {
		if !strings.HasPrefix(val, "GIT_") {
			env[n] = val
			n++
		}
	}
	env = env[:n]

	// Trim any trailing .git path segment.
	if filepath.Base(path) == ".git" {
		path = filepath.Dir(path)
	}

	curdir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	err = os.Chdir(path)
	if err != nil {
		return "", err
	}

	cmd := subprocess.ExecCommand("git", "rev-parse", "--git-dir")
	cmd.Cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to call git rev-parse --git-dir")
	}

	gitdir, err := tools.TranslateCygwinPath(strings.TrimRight(string(out), "\n"))
	if err != nil {
		return "", errors.Wrap(err, "unable to translate path")
	}

	gitdir, err = filepath.Abs(gitdir)
	if err != nil {
		return "", errors.Wrap(err, "unable to canonicalize path")
	}

	err = os.Chdir(curdir)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(gitdir)
}

func fixUrlPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}

	// When parsing a file URL, Go produces a path starting with a slash. If
	// it looks like there's a Windows drive letter at the beginning, strip
	// off the beginning slash. If this is a Unix-style path from a
	// Cygwin-like environment, we'll canonicalize it later.
	re := regexp.MustCompile("/[A-Za-z]:/")
	if re.MatchString(path) {
		return path[1:]
	}
	return path
}

// newHandler creates a new handler for the protocol.
func newHandler(cfg *config.Configuration, output *os.File, msg *inputMessage) (*fileHandler, error) {
	url, err := fileUrlFromRemote(cfg, msg.Remote, msg.Operation)
	if err != nil {
		return nil, err
	}
	if url == nil {
		return nil, errors.New("no valid file:// URLs found")
	}

	path, err := tools.TranslateCygwinPath(fixUrlPath(url.Path))
	if err != nil {
		return nil, err
	}

	gitdir, err := gitDirAtPath(path)
	if err != nil {
		return nil, err
	}

	tracerx.Printf("using %q as remote git directory", gitdir)

	return &fileHandler{
		remotePath:   path,
		remoteConfig: config.NewIn(gitdir, gitdir),
		output:       output,
		config:       cfg,
	}, nil
}

// dispatch dispatches the event depending on the message type.
func (h *fileHandler) dispatch(msg *inputMessage) bool {
	switch msg.Event {
	case "init":
		fmt.Fprintln(h.output, "{}")
	case "upload":
		h.respond(h.upload(msg.Oid, msg.Size, msg.Path))
	case "download":
		h.respond(h.download(msg.Oid, msg.Size))
	case "terminate":
		return false
	default:
		standaloneFailure(fmt.Sprintf("unknown event %q", msg.Event), nil)
	}
	return true
}

// respond sends a response to an upload or download command, using the return
// values from those functions.
func (h *fileHandler) respond(oid string, path string, err error) {
	response := &completeMessage{
		Event: "complete",
		Oid:   oid,
		Path:  path,
	}
	if err != nil {
		response.Error = &errorMessage{Message: err.Error()}
	}
	json.NewEncoder(h.output).Encode(response)
}

// upload performs the upload action for the given OID, size, and path. It
// returns arguments suitable for the respond method.
func (h *fileHandler) upload(oid string, size int64, path string) (string, string, error) {
	if h.remoteConfig.LFSObjectExists(oid, size) {
		// Already there, nothing to do.
		return oid, "", nil
	}
	dest, err := h.remoteConfig.Filesystem().ObjectPath(oid)
	if err != nil {
		return oid, "", err
	}
	return oid, "", lfs.LinkOrCopy(h.remoteConfig, path, dest)
}

// download performs the download action for the given OID and size. It returns
// arguments suitable for the respond method.
func (h *fileHandler) download(oid string, size int64) (string, string, error) {
	if !h.remoteConfig.LFSObjectExists(oid, size) {
		tracerx.Printf("missing object in %q (%s)", h.remotePath, oid)
		return oid, "", errors.Errorf("remote missing object %s", oid)
	}

	src, err := h.remoteConfig.Filesystem().ObjectPath(oid)
	if err != nil {
		return oid, "", err
	}

	tmp, err := lfs.TempFile(h.config, "download")
	if err != nil {
		return oid, "", err
	}
	tmp.Close()
	os.Remove(tmp.Name())
	path := tmp.Name()
	return oid, path, lfs.LinkOrCopy(h.config, src, path)
}

// standaloneFailure reports a fatal error.
func standaloneFailure(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
	os.Exit(2)
}

// ProcessStandaloneData is the primary endpoint for processing data with a
// standalone transfer agent. It reads input from the specified input file and
// produces output to the specified output file.
func ProcessStandaloneData(cfg *config.Configuration, input *os.File, output *os.File) error {
	var handler *fileHandler

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		var msg inputMessage
		if err := json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&msg); err != nil {
			return errors.Wrapf(err, "error decoding json")
		}
		if handler == nil {
			var err error
			handler, err = newHandler(cfg, output, &msg)
			if err != nil {
				return errors.Wrapf(err, "error creating handler")
			}
		}
		if !handler.dispatch(&msg) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "error reading input")
	}
	return nil
}
