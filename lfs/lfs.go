// Package lfs brings together the core LFS functionality
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package lfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/localstorage"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
)

var (
	LargeSizeThreshold = 5 * 1024 * 1024
)

// LocalMediaDir returns the root of lfs objects
func LocalMediaDir() string {
	if localstorage.Objects() != nil {
		return localstorage.Objects().RootDir
	}
	return ""
}

func LocalObjectTempDir() string {
	if localstorage.Objects() != nil {
		return localstorage.Objects().TempDir
	}
	return ""
}

func TempDir() string {
	return localstorage.TempDir
}

func TempFile(prefix string) (*os.File, error) {
	return localstorage.TempFile(prefix)
}

func LocalMediaPath(oid string) (string, error) {
	return localstorage.Objects().BuildObjectPath(oid)
}

func LocalMediaPathReadOnly(oid string) string {
	return localstorage.Objects().ObjectPath(oid)
}

func LocalReferencePath(sha string) string {
	if config.LocalReferenceDir == "" {
		return ""
	}
	return filepath.Join(config.LocalReferenceDir, sha[0:2], sha[2:4], sha)
}

func ObjectExistsOfSize(oid string, size int64) bool {
	path := localstorage.Objects().ObjectPath(oid)
	return tools.FileExistsOfSize(path, size)
}

func Environ(cfg *config.Configuration, manifest *tq.Manifest) []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron)+7)

	dltransfers := manifest.GetDownloadAdapterNames()
	sort.Strings(dltransfers)
	ultransfers := manifest.GetUploadAdapterNames()
	sort.Strings(ultransfers)

	fetchPruneConfig := cfg.FetchPruneConfig()

	env = append(env,
		fmt.Sprintf("LocalWorkingDir=%s", config.LocalWorkingDir),
		fmt.Sprintf("LocalGitDir=%s", config.LocalGitDir),
		fmt.Sprintf("LocalGitStorageDir=%s", config.LocalGitStorageDir),
		fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir()),
		fmt.Sprintf("LocalReferenceDir=%s", config.LocalReferenceDir),
		fmt.Sprintf("TempDir=%s", TempDir()),
		fmt.Sprintf("ConcurrentTransfers=%d", cfg.ConcurrentTransfers()),
		fmt.Sprintf("TusTransfers=%v", cfg.TusTransfersAllowed()),
		fmt.Sprintf("BasicTransfersOnly=%v", cfg.BasicTransfersOnly()),
		fmt.Sprintf("BatchTransfer=%v", cfg.BatchTransfer()),
		fmt.Sprintf("SkipDownloadErrors=%v", cfg.SkipDownloadErrors()),
		fmt.Sprintf("FetchRecentAlways=%v", fetchPruneConfig.FetchRecentAlways),
		fmt.Sprintf("FetchRecentRefsDays=%d", fetchPruneConfig.FetchRecentRefsDays),
		fmt.Sprintf("FetchRecentCommitsDays=%d", fetchPruneConfig.FetchRecentCommitsDays),
		fmt.Sprintf("FetchRecentRefsIncludeRemotes=%v", fetchPruneConfig.FetchRecentRefsIncludeRemotes),
		fmt.Sprintf("PruneOffsetDays=%d", fetchPruneConfig.PruneOffsetDays),
		fmt.Sprintf("PruneVerifyRemoteAlways=%v", fetchPruneConfig.PruneVerifyRemoteAlways),
		fmt.Sprintf("PruneRemoteName=%s", fetchPruneConfig.PruneRemoteName),
		fmt.Sprintf("AccessDownload=%s", cfg.Access("download")),
		fmt.Sprintf("AccessUpload=%s", cfg.Access("upload")),
		fmt.Sprintf("DownloadTransfers=%s", strings.Join(dltransfers, ",")),
		fmt.Sprintf("UploadTransfers=%s", strings.Join(ultransfers, ",")),
	)
	if len(cfg.FetchExcludePaths()) > 0 {
		env = append(env, fmt.Sprintf("FetchExclude=%s", strings.Join(cfg.FetchExcludePaths(), ", ")))
	}
	if len(cfg.FetchIncludePaths()) > 0 {
		env = append(env, fmt.Sprintf("FetchInclude=%s", strings.Join(cfg.FetchIncludePaths(), ", ")))
	}
	for _, ext := range cfg.Extensions() {
		env = append(env, fmt.Sprintf("Extension[%d]=%s", ext.Priority, ext.Name))
	}

	for _, e := range osEnviron {
		if !strings.Contains(strings.SplitN(e, "=", 2)[0], "GIT_") {
			continue
		}
		env = append(env, e)
	}

	return env
}

func InRepo() bool {
	return config.LocalGitDir != ""
}

func ClearTempObjects() error {
	if localstorage.Objects() == nil {
		return nil
	}
	return localstorage.Objects().ClearTempObjects()
}

func ScanObjectsChan() <-chan localstorage.Object {
	return localstorage.Objects().ScanObjectsChan()
}

func init() {
	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "
}

const (
	gitExt       = ".git"
	gitPtrPrefix = "gitdir: "
)

// only used in tests
func AllObjects() []localstorage.Object {
	return localstorage.Objects().AllObjects()
}

func LinkOrCopyFromReference(oid string, size int64) error {
	if ObjectExistsOfSize(oid, size) {
		return nil
	}
	altMediafile := LocalReferencePath(oid)
	mediafile, err := LocalMediaPath(oid)
	if err != nil {
		return err
	}
	if altMediafile != "" && tools.FileExistsOfSize(altMediafile, size) {
		return LinkOrCopy(altMediafile, mediafile)
	}
	return nil
}
