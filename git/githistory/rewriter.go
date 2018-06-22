package githistory

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/lfs"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/git-lfs/tq"
)

// Rewriter allows rewriting topologically equivalent Git histories
// between two revisions.
type Rewriter struct {
	// mu guards entries and commits (see below)
	mu *sync.Mutex
	// entries is a mapping of old tree entries to new (rewritten) ones.
	// Since TreeEntry contains a []byte (and is therefore not a key-able
	// type), a unique TreeEntry -> string function is used for map keys.
	entries map[string]*odb.TreeEntry
	// commits is a mapping of old commit SHAs to new ones, where the ASCII
	// hex encoding of the SHA1 values are used as map keys.
	commits map[string][]byte
	// filter is an optional value used to specify which tree entries
	// (blobs, subtrees) are modifiable given a BlobFn. If non-nil, this
	// filter will cull out any unmodifiable subtrees and blobs.
	filter *filepathfilter.Filter
	// db is the *ObjectDatabase from which blobs, commits, and trees are
	// loaded from.
	db *odb.ObjectDatabase
	// l is the *tasklog.Logger to which updates are written.
	l *tasklog.Logger
}

// RewriteOptions is an options type given to the Rewrite() function.
type RewriteOptions struct {
	// Include is the list of refs of which commits reachable by that ref
	// will be included.
	Include []string
	// Exclude is the list of refs of which commits reachable by that ref
	// will be excluded.
	Exclude []string

	// UpdateRefs specifies whether the Rewriter should move refs from the
	// original graph onto the migrated one. If true, the refs will be
	// moved, and a reflog entry will be created.
	UpdateRefs bool

	// Verbose mode prints migrated objects.
	Verbose bool

	// ObjectMapFilePath is the path to the map of old sha1 to new sha1
	// commits
	ObjectMapFilePath string

	// BlobFn specifies a function to rewrite blobs.
	//
	// It is called once per unique, unchanged path. That is to say, if
	// /a/foo and /a/bar contain identical contents, the BlobFn will be
	// called twice: once for /a/foo and once for /a/bar, but no more on
	// each blob for subsequent revisions, so long as each entry remains
	// unchanged.
	BlobFn BlobRewriteFn
	// TreeCallbackFn specifies a function to rewrite trees after they have
	// been reassembled by calling the above BlobFn on all existing tree
	// entries.
	TreeCallbackFn TreeCallbackFn
}

// blobFn returns a useable BlobRewriteFn, either the one that was given in the
// *RewriteOptions, or a noopBlobFn.
func (r *RewriteOptions) blobFn() BlobRewriteFn {
	if r.BlobFn == nil {
		return noopBlobFn
	}
	return r.BlobFn
}

// treeFn returns a useable TreeRewriteFn, either the one that was given in the
// *RewriteOptions, or a noopTreeFn.
func (r *RewriteOptions) treeFn() TreeCallbackFn {
	if r.TreeCallbackFn == nil {
		return noopTreeFn
	}
	return r.TreeCallbackFn
}

// BlobRewriteFn is a mapping function that takes a given blob and returns a
// new, modified blob. If it returns an error, the new blob will not be written
// and instead the error will be returned from the Rewrite() function.
//
// Invocations of an instance of BlobRewriteFn are not expected to store the
// returned blobs in the *git/odb.ObjectDatabase.
//
// The path argument is given to be an absolute path to the tree entry being
// rewritten, where the repository root is the root of the path given. For
// instance, a file "b.txt" in directory "dir" would be given as "/dir/b.txt",
// where as a file "a.txt" in the root would be given as "/a.txt".
//
// As above, the path separators are OS specific, and equivalent to the result
// of filepath.Join(...) or os.PathSeparator.
type BlobRewriteFn func(path string, b *odb.Blob) (*odb.Blob, error)

// TreeCallbackFn specifies a function to call before writing a re-written tree
// to the object database. The TreeCallbackFn can return a modified tree to be
// written to the object database instead of one generated from calling BlobFn
// on all of the tree entries.
//
// Trees returned from a TreeCallbackFn MUST have all objects referenced in the
// entryset already written to the object database.
//
// TreeCallbackFn can be nil, and will therefore exhibit behavior equivalent to
// only calling the BlobFn on existing tree entries.
//
// If the TreeCallbackFn returns an error, it will be returned from the
// Rewrite() invocation.
type TreeCallbackFn func(path string, t *odb.Tree) (*odb.Tree, error)

type rewriterOption func(*Rewriter)

var (
	// WithFilter is an optional argument given to the NewRewriter
	// constructor function to limit invocations of the BlobRewriteFn to
	// only pathspecs that match the given *filepathfilter.Filter.
	WithFilter = func(filter *filepathfilter.Filter) rewriterOption {
		return func(r *Rewriter) {
			r.filter = filter
		}
	}

	// WithLoggerto logs updates caused by the *git/githistory.Rewriter to
	// the given io.Writer "sink".
	WithLoggerTo = func(sink io.Writer) rewriterOption {
		return WithLogger(tasklog.NewLogger(sink))
	}

	// WithLogger logs updates caused by the *git/githistory.Rewriter to the
	// be given to the provided logger, "l".
	WithLogger = func(l *tasklog.Logger) rewriterOption {
		return func(r *Rewriter) {
			r.l = l
		}
	}

	// noopBlobFn is a no-op implementation of the BlobRewriteFn. It returns
	// the blob that it was given, and returns no error.
	noopBlobFn = func(path string, b *odb.Blob) (*odb.Blob, error) { return b, nil }
	// noopTreeFn is a no-op implementation of the TreeRewriteFn. It returns
	// the tree that it was given, and returns no error.
	noopTreeFn = func(path string, t *odb.Tree) (*odb.Tree, error) { return t, nil }
)

// NewRewriter constructs a *Rewriter from the given *ObjectDatabase instance.
func NewRewriter(db *odb.ObjectDatabase, opts ...rewriterOption) *Rewriter {
	rewriter := &Rewriter{
		mu:      new(sync.Mutex),
		entries: make(map[string]*odb.TreeEntry),
		commits: make(map[string][]byte),

		db: db,
	}

	for _, opt := range opts {
		opt(rewriter)
	}
	return rewriter
}

// ScanForPointers scans through the range of commits given by
// *RewriteOptions.{Left,Right} and adds any pointers matching the rewrite
// filter to the transfer queue to be downloaded
func (r *Rewriter) ScanForPointers(q *tq.TransferQueue, opt *RewriteOptions, gf *lfs.GitFilter) error {
	// Obtain a list of commits to scan
	commits, err := r.commitsToMigrate(opt)
	if err != nil {
		return err
	}

	waiter := r.l.Waiter("migrate: Scanning commits")
	defer waiter.Complete()

	for _, oid := range commits {
		commit, err := r.db.Commit(oid)
		if err != nil {
			return err
		}

		if err := r.scanTree(q, gf, commit.TreeID, ""); err != nil {
			return err
		}
	}

	return nil
}

// Rewrite rewrites the range of commits given by *RewriteOptions.{Left,Right}
// using the BlobRewriteFn to rewrite the individual blobs.
func (r *Rewriter) Rewrite(opt *RewriteOptions) ([]byte, error) {
	// First, obtain a list of commits to rewrite.
	commits, err := r.commitsToMigrate(opt)
	if err != nil {
		return nil, err
	}

	var perc *tasklog.PercentageTask
	if opt.UpdateRefs {
		perc = r.l.Percentage("migrate: Rewriting commits", uint64(len(commits)))
	} else {
		perc = r.l.Percentage("migrate: Examining commits", uint64(len(commits)))
	}

	var vPerc *tasklog.PercentageTask
	if opt.Verbose {
		vPerc = perc
	}

	var objectMapFile *os.File
	if len(opt.ObjectMapFilePath) > 0 {
		objectMapFile, err = os.OpenFile(opt.ObjectMapFilePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			return nil, fmt.Errorf("Could not create object map file: %v", err)
		}
		defer objectMapFile.Close()
	}

	// Keep track of the last commit that we rewrote. Callers often want
	// this so that they can perform a git-update-ref(1).
	var tip []byte
	for _, oid := range commits {
		// Load the original commit to access the data necessary in
		// order to rewrite it.
		original, err := r.db.Commit(oid)
		if err != nil {
			return nil, err
		}

		// Rewrite the tree given at that commit.
		rewrittenTree, err := r.rewriteTree(oid, original.TreeID, "", opt.blobFn(), opt.treeFn(), vPerc)
		if err != nil {
			return nil, err
		}

		// Create a new list of parents from the original commit to
		// point at the rewritten parents in order to create a
		// topologically equivalent DAG.
		//
		// This operation is safe since we are visiting the commits in
		// reverse topological order and therefore have seen all parents
		// before children (in other words, r.uncacheCommit(...) will
		// always return a value, if the prospective parent is a part of
		// the migration).
		rewrittenParents := make([][]byte, 0, len(original.ParentIDs))
		for _, originalParent := range original.ParentIDs {
			rewrittenParent, ok := r.uncacheCommit(originalParent)
			if !ok {
				// If we haven't seen the parent before, this
				// means that we're doing a partial migration
				// and the parent that we're looking for isn't
				// included.
				//
				// Use the original parent to properly link
				// history across the migration boundary.
				rewrittenParent = originalParent
			}

			rewrittenParents = append(rewrittenParents, rewrittenParent)
		}

		// Construct a new commit using the original header information,
		// but the rewritten set of parents as well as root tree.
		rewrittenCommit := &odb.Commit{
			Author:       original.Author,
			Committer:    original.Committer,
			ExtraHeaders: original.ExtraHeaders,
			Message:      original.Message,

			ParentIDs: rewrittenParents,
			TreeID:    rewrittenTree,
		}

		var newSha []byte

		if original.Equal(rewrittenCommit) {
			newSha = make([]byte, len(oid))
			copy(newSha, oid)
		} else {
			newSha, err = r.db.WriteCommit(rewrittenCommit)
			if err != nil {
				return nil, err
			}
			if objectMapFile != nil {
				if _, err := fmt.Fprintf(objectMapFile, "%x,%x\n", oid, newSha); err != nil {
					return nil, err
				}
			}
		}

		// Cache that commit so that we can reassign children of this
		// commit.
		r.cacheCommit(oid, newSha)

		// Increment the percentage displayed in the terminal.
		perc.Count(1)

		// Move the tip forward.
		tip = newSha
	}

	if opt.UpdateRefs {
		refs, err := r.refsToMigrate()
		if err != nil {
			return nil, errors.Wrap(err, "could not find refs to update")
		}

		root, _ := r.db.Root()

		updater := &refUpdater{
			CacheFn: r.uncacheCommit,
			Logger:  r.l,
			Refs:    refs,
			Root:    root,

			db: r.db,
		}

		if err := updater.UpdateRefs(); err != nil {
			return nil, errors.Wrap(err, "could not update refs")
		}
	}

	return tip, err
}

// scanTree recursively scans through a tree and adds any pointers matching the
// rewrite filter to the transfer queue to be downloaded
func (r *Rewriter) scanTree(q *tq.TransferQueue, gf *lfs.GitFilter, treeOID []byte, path string) error {
	tree, err := r.db.Tree(treeOID)
	if err != nil {
		return err
	}

	for _, entry := range tree.Entries {
		var fullpath string
		if len(path) > 0 {
			fullpath = strings.Join([]string{path, entry.Name}, "/")
		} else {
			fullpath = entry.Name
		}

		if !r.allows(entry.Type(), fullpath) {
			continue
		}

		// If this is a symlink, skip it
		if entry.Filemode == 0120000 {
			continue
		}

		switch entry.Type() {
		case odb.BlobObjectType:
			// Check if the blob is a pointer, and if so,
			// add it to the transfer queue
			blob, err := r.db.Blob(entry.Oid)
			if err != nil {
				return err
			}

			ptr, err := lfs.DecodePointer(blob.Contents)
			if errors.IsNotAPointerError(err) {
				continue
			}
			if err != nil {
				return err
			}

			downloadPath, err := gf.ObjectPath(ptr.Oid)
			if err != nil {
				return err
			}

			// Only add files to the transfer queue that aren't already cached
			if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
				q.Add(entry.Name, downloadPath, ptr.Oid, ptr.Size)
			}

		case odb.TreeObjectType:
			// Scan all subtrees
			err = r.scanTree(q, gf, entry.Oid, fullpath)

		}
		if err != nil {
			return err
		}
	}

	return nil
}

// rewriteTree is a recursive function which rewrites a tree given by the ID
// "sha" and path "path". It uses the given BlobRewriteFn to rewrite all blobs
// within the tree, either calling that function or recurring down into subtrees
// by re-assigning the SHA.
//
// Once it is done assembling the entries in a given subtree, it then calls the
// TreeCallbackFn, "tfn" to perform a final traversal of the subtree before
// saving it to the object database.
//
// It returns the new SHA of the rewritten tree, or an error if the tree was
// unable to be rewritten.
func (r *Rewriter) rewriteTree(commitOID []byte, treeOID []byte, path string, fn BlobRewriteFn, tfn TreeCallbackFn, perc *tasklog.PercentageTask) ([]byte, error) {
	tree, err := r.db.Tree(treeOID)
	if err != nil {
		return nil, err
	}

	entries := make([]*odb.TreeEntry, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		var fullpath string
		if len(path) > 0 {
			fullpath = strings.Join([]string{path, entry.Name}, "/")
		} else {
			fullpath = entry.Name
		}

		if !r.allows(entry.Type(), fullpath) {
			entries = append(entries, copyEntry(entry))
			continue
		}

		// If this is a symlink, skip it
		if entry.Filemode == 0120000 {
			entries = append(entries, copyEntry(entry))
			continue
		}

		if cached := r.uncacheEntry(entry); cached != nil {
			entries = append(entries, copyEntry(cached))
			continue
		}

		var oid []byte

		switch entry.Type() {
		case odb.BlobObjectType:
			oid, err = r.rewriteBlob(commitOID, entry.Oid, fullpath, fn, perc)
		case odb.TreeObjectType:
			oid, err = r.rewriteTree(commitOID, entry.Oid, fullpath, fn, tfn, perc)
		default:
			oid = entry.Oid

		}
		if err != nil {
			return nil, err
		}

		entries = append(entries, r.cacheEntry(entry, &odb.TreeEntry{
			Filemode: entry.Filemode,
			Name:     entry.Name,
			Oid:      oid,
		}))
	}

	rewritten, err := tfn("/"+path, &odb.Tree{Entries: entries})
	if err != nil {
		return nil, err
	}

	if tree.Equal(rewritten) {
		return treeOID, nil
	}
	return r.db.WriteTree(rewritten)
}

func copyEntry(e *odb.TreeEntry) *odb.TreeEntry {
	if e == nil {
		return nil
	}

	oid := make([]byte, len(e.Oid))
	copy(oid, e.Oid)

	return &odb.TreeEntry{
		Filemode: e.Filemode,
		Name:     e.Name,
		Oid:      oid,
	}
}

func (r *Rewriter) allows(typ odb.ObjectType, abs string) bool {
	switch typ {
	case odb.BlobObjectType:
		return r.Filter().Allows(strings.TrimPrefix(abs, "/"))
	case odb.CommitObjectType, odb.TreeObjectType:
		return true
	default:
		panic(fmt.Sprintf("git/githistory: unknown entry type: %s", typ))
	}
}

// rewriteBlob calls the given BlobRewriteFn "fn" on a blob given in the object
// database by the SHA1 "from" []byte. It writes and returns the new blob SHA,
// or an error if either the BlobRewriteFn returned one, or if the object could
// not be loaded/saved.
func (r *Rewriter) rewriteBlob(commitOID, from []byte, path string, fn BlobRewriteFn, perc *tasklog.PercentageTask) ([]byte, error) {
	blob, err := r.db.Blob(from)
	if err != nil {
		return nil, err
	}

	b, err := fn(path, blob)
	if err != nil {
		return nil, err
	}

	if !blob.Equal(b) {
		sha, err := r.db.WriteBlob(b)
		if err != nil {
			return nil, err
		}

		// Close the source blob, so long as it is not equal to the
		// rewritten blob. If the two are equal, as in the check above
		// this comment, calling r.db.WriteBlob(b) will have already
		// closed both "b" and "blob" since they are the same.
		//
		// Closing an *os.File twice causes an `os.ErrInvalid` to be
		// returned.
		if err = blob.Close(); err != nil {
			return nil, err
		}

		if perc != nil {
			perc.Entry(fmt.Sprintf("migrate: commit %s: %s", hex.EncodeToString(commitOID), path))
		}

		return sha, nil
	}

	// Close the source blob, since it is identical to the rewritten blob,
	// but neither were written.
	if err := blob.Close(); err != nil {
		return nil, err
	}
	return from, nil
}

// commitsToMigrate returns an in-memory copy of a list of commits according to
// the output of git-rev-list(1) (given the *RewriteOptions), where each
// outputted commit is 20 bytes of raw SHA1.
//
// If any error was encountered, it will be returned.
func (r *Rewriter) commitsToMigrate(opt *RewriteOptions) ([][]byte, error) {
	waiter := r.l.Waiter("migrate: Sorting commits")
	defer waiter.Complete()

	scanner, err := git.NewRevListScanner(
		opt.Include, opt.Exclude, r.scannerOpts())
	if err != nil {
		return nil, err
	}

	var commits [][]byte
	for scanner.Scan() {
		commits = append(commits, scanner.OID())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if err = scanner.Close(); err != nil {
		return nil, err
	}
	return commits, nil
}

// refsToMigrate returns a list of references to migrate, or an error if loading
// those references failed.
func (r *Rewriter) refsToMigrate() ([]*git.Ref, error) {
	var refs []*git.Ref
	var err error

	if root, ok := r.db.Root(); ok {
		refs, err = git.AllRefsIn(root)
	} else {
		refs, err = git.AllRefs()
	}

	if err != nil {
		return nil, err
	}

	var local []*git.Ref
	for _, ref := range refs {
		if ref.Type == git.RefTypeRemoteBranch || ref.Type == git.RefTypeRemoteTag {
			continue
		}

		local = append(local, ref)
	}

	return local, nil
}

// scannerOpts returns a *git.ScanRefsOptions instance to be given to the
// *git.RevListScanner.
//
// If the database this *Rewriter is operating in a given root (not in memory)
// it re-assigns the working directory to be there.
func (r *Rewriter) scannerOpts() *git.ScanRefsOptions {
	opts := &git.ScanRefsOptions{
		Mode:        git.ScanRefsMode,
		Order:       git.TopoRevListOrder,
		Reverse:     true,
		CommitsOnly: true,

		SkippedRefs: make([]string, 0),
		Mutex:       new(sync.Mutex),
		Names:       make(map[string]string),
	}

	if root, ok := r.db.Root(); ok {
		opts.WorkingDir = root
	}
	return opts
}

// Filter returns the filter used by this *Rewriter to filter subtrees, blobs
// (see above).
func (r *Rewriter) Filter() *filepathfilter.Filter {
	return r.filter
}

// cacheEntry caches then given "from" entry so that it is always rewritten as
// a *TreeEntry equivalent to "to".
func (r *Rewriter) cacheEntry(from, to *odb.TreeEntry) *odb.TreeEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[r.entryKey(from)] = to

	return to
}

// uncacheEntry returns a *TreeEntry that is cached from the given *TreeEntry
// "from". That is to say, it returns the *TreeEntry that "from" should be
// rewritten to, or nil if none could be found.
func (r *Rewriter) uncacheEntry(from *odb.TreeEntry) *odb.TreeEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.entries[r.entryKey(from)]
}

// entryKey returns a unique key for a given *TreeEntry "e".
func (r *Rewriter) entryKey(e *odb.TreeEntry) string {
	return fmt.Sprintf("%s:%x", e.Name, e.Oid)
}

// cacheEntry caches then given "from" commit so that it is always rewritten as
// a *git/odb.Commit equivalent to "to".
func (r *Rewriter) cacheCommit(from, to []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commits[hex.EncodeToString(from)] = to
}

// uncacheCommit returns a *git/odb.Commit that is cached from the given
// *git/odb.Commit "from". That is to say, it returns the *git/odb.Commit that
// "from" should be rewritten to and true, or nil and false if none could be
// found.
func (r *Rewriter) uncacheCommit(from []byte) ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.commits[hex.EncodeToString(from)]
	return c, ok
}
