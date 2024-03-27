package commands

import (
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
)

// Handles the process of checking out a single file, and updating the git
// index.
func newSingleCheckout(gitEnv config.Environment, remote string) abstractCheckout {
	clean, ok := gitEnv.Get("filter.lfs.clean")
	if !ok || len(clean) == 0 {
		return &noOpCheckout{remote: remote}
	}

	// Get a converter from repo-relative to cwd-relative
	// Since writing data & calling git update-index must be relative to cwd
	pathConverter, err := lfs.NewRepoToCurrentPathConverter(cfg)
	if err != nil {
		Panic(err, tr.Tr.Get("Could not convert file paths"))
	}

	return &singleCheckout{
		gitIndexer:    &gitIndexer{},
		pathConverter: pathConverter,
		manifest:      nil,
		remote:        remote,
	}
}

type abstractCheckout interface {
	Manifest() tq.Manifest
	Skip() bool
	Run(*lfs.WrappedPointer)
	RunToPath(*lfs.WrappedPointer, string) error
	Close()
}

type singleCheckout struct {
	gitIndexer    *gitIndexer
	pathConverter lfs.PathConverter
	manifest      tq.Manifest
	remote        string
}

func (c *singleCheckout) Manifest() tq.Manifest {
	if c.manifest == nil {
		c.manifest = getTransferManifestOperationRemote("download", c.remote)
	}
	return c.manifest
}

func (c *singleCheckout) Skip() bool {
	return false
}

func (c *singleCheckout) Run(p *lfs.WrappedPointer) {
	cwdfilepath := c.pathConverter.Convert(p.Name)

	// Check the content - either missing or still this pointer (not exist is ok)
	filepointer, err := lfs.DecodePointerFromFile(cwdfilepath)
	if err != nil && !os.IsNotExist(err) {
		if errors.IsNotAPointerError(err) || errors.IsBadPointerKeyError(err) {
			// File has non-pointer content, leave it alone
			return
		}

		LoggedError(err, tr.Tr.Get("Checkout error: %s", err))
		return
	}

	if filepointer != nil && filepointer.Oid != p.Oid {
		// User has probably manually reset a file to another commit
		// while leaving it a pointer; don't mess with this
		return
	}

	if err := c.RunToPath(p, cwdfilepath); err != nil {
		if errors.IsDownloadDeclinedError(err) {
			// acceptable error, data not local (fetch not run or include/exclude)
			Error(tr.Tr.Get("Skipped checkout for %q, content not local. Use fetch to download.", p.Name))
		} else {
			FullError(errors.Wrap(err, tr.Tr.Get("could not check out %q", p.Name)))
		}
		return
	}

	// errors are only returned when the gitIndexer is starting a new cmd
	if err := c.gitIndexer.Add(cwdfilepath); err != nil {
		Panic(err, tr.Tr.Get("Could not update the index"))
	}
}

// RunToPath checks out the pointer specified by p to the given path.  It does
// not perform any sort of sanity checking or add the path to the index.
func (c *singleCheckout) RunToPath(p *lfs.WrappedPointer, path string) error {
	gitfilter := lfs.NewGitFilter(cfg)
	return gitfilter.SmudgeToFile(path, p.Pointer, false, c.manifest, nil)
}

func (c *singleCheckout) Close() {
	if err := c.gitIndexer.Close(); err != nil {
		LoggedError(err, "%s\n%s", tr.Tr.Get("Error updating the Git index:"), c.gitIndexer.Output())
	}
}

type noOpCheckout struct {
	manifest tq.Manifest
	remote   string
}

func (c *noOpCheckout) Manifest() tq.Manifest {
	if c.manifest == nil {
		c.manifest = getTransferManifestOperationRemote("download", c.remote)
	}
	return c.manifest
}

func (c *noOpCheckout) Skip() bool {
	return true
}

func (c *noOpCheckout) RunToPath(p *lfs.WrappedPointer, path string) error {
	return nil
}

func (c *noOpCheckout) Run(p *lfs.WrappedPointer) {}
func (c *noOpCheckout) Close()                    {}

// Don't fire up the update-index command until we have at least one file to
// give it. Otherwise git interprets the lack of arguments to mean param-less update-index
// which can trigger entire working copy to be re-examined, which triggers clean filters
// and which has unexpected side effects (e.g. downloading filtered-out files)
type gitIndexer struct {
	cmd    *subprocess.Cmd
	input  io.WriteCloser
	output bytes.Buffer
	mu     sync.Mutex
}

func (i *gitIndexer) Add(path string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cmd == nil {
		// Fire up the update-index command
		cmd, err := git.UpdateIndexFromStdin()
		if err != nil {
			return err
		}
		cmd.Stdout = &i.output
		cmd.Stderr = &i.output
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		err = cmd.Start()
		if err != nil {
			return err
		}
		i.cmd = cmd
		i.input = stdin
	}

	i.input.Write([]byte(path + "\n"))
	return nil
}

func (i *gitIndexer) Output() string {
	return i.output.String()
}

func (i *gitIndexer) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.input != nil {
		i.input.Close()
	}

	if i.cmd != nil {
		return i.cmd.Wait()
	}

	return nil
}
