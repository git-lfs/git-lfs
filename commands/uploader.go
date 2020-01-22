package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
)

func uploadForRefUpdates(ctx *uploadContext, updates []*git.RefUpdate, pushAll bool) error {
	gitscanner, err := ctx.buildGitScanner()
	if err != nil {
		return err
	}

	defer func() {
		gitscanner.Close()
		ctx.ReportErrors()
	}()

	verifyLocksForUpdates(ctx.lockVerifier, updates)
	rightSides := make([]string, 0, len(updates))
	for _, update := range updates {
		right := update.Right().Sha
		if update.LeftCommitish() != right {
			rightSides = append(rightSides, right)
		}
	}
	for _, update := range updates {
		// initialized here to prevent looped defer
		q := ctx.NewQueue(
			tq.RemoteRef(update.Right()),
		)
		err := uploadLeftOrAll(gitscanner, ctx, q, rightSides, update, pushAll)
		ctx.CollectErrors(q)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("ref %s:", update.Left().Name))
		}
	}

	return nil
}

func uploadLeftOrAll(g *lfs.GitScanner, ctx *uploadContext, q *tq.TransferQueue, bases []string, update *git.RefUpdate, pushAll bool) error {
	cb := ctx.gitScannerCallback(q)
	if pushAll {
		if err := g.ScanRefWithDeleted(update.LeftCommitish(), cb); err != nil {
			return err
		}
	} else {
		left := update.LeftCommitish()
		right := update.Right().Sha
		if left == right {
			right = ""
		}
		if err := g.ScanMultiRangeToRemote(left, bases, cb); err != nil {
			return err
		}
	}
	return ctx.scannerError()
}

type uploadContext struct {
	Remote       string
	DryRun       bool
	Manifest     *tq.Manifest
	uploadedOids tools.StringSet
	gitfilter    *lfs.GitFilter

	logger *tasklog.Logger
	meter  *tq.Meter

	committerName  string
	committerEmail string

	lockVerifier *lockVerifier

	// allowMissing specifies whether pushes containing missing/corrupt
	// pointers should allow pushing Git blobs
	allowMissing bool

	// tracks errors from gitscanner callbacks
	scannerErr error
	errMu      sync.Mutex

	// filename => oid
	missing   map[string]string
	corrupt   map[string]string
	otherErrs []error
}

func newUploadContext(dryRun bool) *uploadContext {
	remote := cfg.PushRemote()
	manifest := getTransferManifestOperationRemote("upload", remote)
	ctx := &uploadContext{
		Remote:       remote,
		Manifest:     manifest,
		DryRun:       dryRun,
		uploadedOids: tools.NewStringSet(),
		gitfilter:    lfs.NewGitFilter(cfg),
		lockVerifier: newLockVerifier(manifest),
		allowMissing: cfg.Git.Bool("lfs.allowincompletepush", false),
		missing:      make(map[string]string),
		corrupt:      make(map[string]string),
		otherErrs:    make([]error, 0),
	}

	var sink io.Writer = os.Stdout
	if dryRun {
		sink = ioutil.Discard
	}

	ctx.logger = tasklog.NewLogger(sink,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	ctx.meter = buildProgressMeter(ctx.DryRun, tq.Upload)
	ctx.logger.Enqueue(ctx.meter)
	ctx.committerName, ctx.committerEmail = cfg.CurrentCommitter()
	return ctx
}

func (c *uploadContext) NewQueue(options ...tq.Option) *tq.TransferQueue {
	return tq.NewTransferQueue(tq.Upload, c.Manifest, c.Remote, append(options,
		tq.DryRun(c.DryRun),
		tq.WithProgress(c.meter),
	)...)
}

func (c *uploadContext) scannerError() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()

	return c.scannerErr
}

func (c *uploadContext) addScannerError(err error) {
	c.errMu.Lock()
	defer c.errMu.Unlock()

	if c.scannerErr != nil {
		c.scannerErr = fmt.Errorf("%v\n%v", c.scannerErr, err)
	} else {
		c.scannerErr = err
	}
}

func (c *uploadContext) buildGitScanner() (*lfs.GitScanner, error) {
	gitscanner := lfs.NewGitScanner(cfg, nil)
	gitscanner.FoundLockable = func(n string) { c.lockVerifier.LockedByThem(n) }
	gitscanner.PotentialLockables = c.lockVerifier
	return gitscanner, gitscanner.RemoteForPush(c.Remote)
}

func (c *uploadContext) gitScannerCallback(tqueue *tq.TransferQueue) func(*lfs.WrappedPointer, error) {
	return func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			c.addScannerError(err)
		} else {
			c.UploadPointers(tqueue, p)
		}
	}
}

// AddUpload adds the given oid to the set of oids that have been uploaded in
// the current process.
func (c *uploadContext) SetUploaded(oid string) {
	c.uploadedOids.Add(oid)
}

// HasUploaded determines if the given oid has already been uploaded in the
// current process.
func (c *uploadContext) HasUploaded(oid string) bool {
	return c.uploadedOids.Contains(oid)
}

func (c *uploadContext) prepareUpload(unfiltered ...*lfs.WrappedPointer) []*lfs.WrappedPointer {
	numUnfiltered := len(unfiltered)
	uploadables := make([]*lfs.WrappedPointer, 0, numUnfiltered)

	// XXX(taylor): temporary measure to fix duplicate (broken) results from
	// scanner
	uniqOids := tools.NewStringSet()

	// separate out objects that _should_ be uploaded, but don't exist in
	// .git/lfs/objects. Those will skipped if the server already has them.
	for _, p := range unfiltered {
		// object already uploaded in this process, or we've already
		// seen this OID (see above), skip!
		if uniqOids.Contains(p.Oid) || c.HasUploaded(p.Oid) {
			continue
		}
		uniqOids.Add(p.Oid)

		// canUpload determines whether the current pointer "p" can be
		// uploaded through the TransferQueue below. It is set to false
		// only when the file is locked by someone other than the
		// current committer.
		var canUpload bool = true

		if c.lockVerifier.LockedByThem(p.Name) {
			// If the verification state is enabled, this failed
			// locks verification means that the push should fail.
			//
			// If the state is disabled, the verification error is
			// silent and the user can upload.
			//
			// If the state is undefined, the verification error is
			// sent as a warning and the user can upload.
			canUpload = !c.lockVerifier.Enabled()
		}

		c.lockVerifier.LockedByUs(p.Name)

		if canUpload {
			// estimate in meter early (even if it's not going into
			// uploadables), since we will call Skip() based on the
			// results of the download check queue.
			c.meter.Add(p.Size)

			uploadables = append(uploadables, p)
		}
	}

	return uploadables
}

func (c *uploadContext) UploadPointers(q *tq.TransferQueue, unfiltered ...*lfs.WrappedPointer) {
	if c.DryRun {
		for _, p := range unfiltered {
			if c.HasUploaded(p.Oid) {
				continue
			}

			Print("push %s => %s", p.Oid, p.Name)
			c.SetUploaded(p.Oid)
		}

		return
	}

	pointers := c.prepareUpload(unfiltered...)
	for _, p := range pointers {
		t, err := c.uploadTransfer(p)
		if err != nil && !errors.IsCleanPointerError(err) {
			ExitWithError(err)
		}

		q.Add(t.Name, t.Path, t.Oid, t.Size, t.Missing)
		c.SetUploaded(p.Oid)
	}
}

func (c *uploadContext) CollectErrors(tqueue *tq.TransferQueue) {
	tqueue.Wait()

	for _, err := range tqueue.Errors() {
		if malformed, ok := err.(*tq.MalformedObjectError); ok {
			if malformed.Missing() {
				c.missing[malformed.Name] = malformed.Oid
			} else if malformed.Corrupt() {
				c.corrupt[malformed.Name] = malformed.Oid
			}
		} else {
			c.otherErrs = append(c.otherErrs, err)
		}
	}
}

func (c *uploadContext) ReportErrors() {
	c.meter.Finish()

	for _, err := range c.otherErrs {
		FullError(err)
	}

	if len(c.missing) > 0 || len(c.corrupt) > 0 {
		var action string
		if c.allowMissing {
			action = "missing objects"
		} else {
			action = "failed"
		}

		Print("LFS upload %s:", action)
		for name, oid := range c.missing {
			Print("  (missing) %s (%s)", name, oid)
		}
		for name, oid := range c.corrupt {
			Print("  (corrupt) %s (%s)", name, oid)
		}

		if !c.allowMissing {
			pushMissingHint := []string{
				"hint: Your push was rejected due to missing or corrupt local objects.",
				"hint: You can disable this check with: 'git config lfs.allowincompletepush true'",
			}
			Print(strings.Join(pushMissingHint, "\n"))
			os.Exit(2)
		}
	}

	if len(c.otherErrs) > 0 {
		os.Exit(2)
	}

	if c.lockVerifier.HasUnownedLocks() {
		Print("Unable to push locked files:")
		for _, unowned := range c.lockVerifier.UnownedLocks() {
			Print("* %s - %s", unowned.Path(), unowned.Owners())
		}

		if c.lockVerifier.Enabled() {
			Exit("ERROR: Cannot update locked files.")
		} else {
			Error("WARNING: The above files would have halted this push.")
		}
	} else if c.lockVerifier.HasOwnedLocks() {
		Print("Consider unlocking your own locked files: (`git lfs unlock <path>`)")
		for _, owned := range c.lockVerifier.OwnedLocks() {
			Print("* %s", owned.Path())
		}
	}
}

var (
	githubHttps, _ = url.Parse("https://github.com")
	githubSsh, _   = url.Parse("ssh://github.com")

	// hostsWithKnownLockingSupport is a list of scheme-less hostnames
	// (without port numbers) that are known to implement the LFS locking
	// API.
	//
	// Additions are welcome.
	hostsWithKnownLockingSupport = []*url.URL{
		githubHttps, githubSsh,
	}
)

func (c *uploadContext) uploadTransfer(p *lfs.WrappedPointer) (*tq.Transfer, error) {
	var missing bool

	filename := p.Name
	oid := p.Oid

	localMediaPath, err := c.gitfilter.ObjectPath(oid)
	if err != nil {
		return nil, errors.Wrapf(err, "Error uploading file %s (%s)", filename, oid)
	}

	if len(filename) > 0 {
		if missing, err = c.ensureFile(filename, localMediaPath, oid); err != nil && !errors.IsCleanPointerError(err) {
			return nil, err
		}
	}

	return &tq.Transfer{
		Name:    filename,
		Path:    localMediaPath,
		Oid:     oid,
		Size:    p.Size,
		Missing: missing,
	}, nil
}

// ensureFile makes sure that the cleanPath exists before pushing it.  If it
// does not exist, it attempts to clean it by reading the file at smudgePath.
func (c *uploadContext) ensureFile(smudgePath, cleanPath, oid string) (bool, error) {
	if _, err := os.Stat(cleanPath); err == nil {
		return false, nil
	}

	localPath := filepath.Join(cfg.LocalWorkingDir(), smudgePath)
	file, err := os.Open(localPath)
	if err != nil {
		return !c.allowMissing, nil
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return false, err
	}

	cleaned, err := c.gitfilter.Clean(file, file.Name(), stat.Size(), nil)
	if cleaned != nil {
		cleaned.Teardown()
	}

	if err != nil {
		return false, err
	}
	return false, nil
}

// supportsLockingAPI returns whether or not a given url is known to support
// the LFS locking API by whether or not its hostname is included in the list
// above.
func supportsLockingAPI(rawurl string) bool {
	u, err := url.Parse(rawurl)
	if err != nil {
		tracerx.Printf("commands: unable to parse %q to determine locking support: %v", rawurl, err)
		return false
	}

	for _, supported := range hostsWithKnownLockingSupport {
		if supported.Scheme == u.Scheme &&
			supported.Hostname() == u.Hostname() &&
			strings.HasPrefix(u.Path, supported.Path) {
			return true
		}
	}
	return false
}

// disableFor disables lock verification for the given lfsapi.Endpoint,
// "endpoint".
func disableFor(rawurl string) error {
	tracerx.Printf("commands: disabling lock verification for %q", rawurl)

	key := strings.Join([]string{"lfs", rawurl, "locksverify"}, ".")

	_, err := cfg.SetGitLocalKey(key, "false")
	return err
}
