package lfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/localstorage"
	"github.com/github/git-lfs/tools"
	"github.com/github/git-lfs/vendor/_nuts/github.com/rubyist/tracerx"
)

const (
	tempDirPerms       = 0755
	localMediaDirPerms = 0755
	localLogDirPerms   = 0755
)

var (
	LargeSizeThreshold = 5 * 1024 * 1024
	objects            *localstorage.LocalStorage
	LocalMediaDir      string // root of lfs objects
	LocalObjectTempDir string // where temporarily downloading objects are stored
	TempDir            = filepath.Join(os.TempDir(), "git-lfs")
	checkedTempDir     string
)

func ResolveDirs() {

	config.ResolveGitBasicDirs()
	TempDir = filepath.Join(config.LocalGitDir, "lfs", "tmp") // temp files per worktree

	objs, err := localstorage.New(
		filepath.Join(config.LocalGitStorageDir, "lfs", "objects"),
		filepath.Join(TempDir, "objects"),
	)

	if err != nil {
		panic(fmt.Sprintf("Error trying to init LocalStorage: %s", err))
	}

	objects = objs
	LocalMediaDir = objs.RootDir
	LocalObjectTempDir = objs.TempDir
	config.LocalLogDir = filepath.Join(objs.RootDir, "logs")
	if err := os.MkdirAll(config.LocalLogDir, localLogDirPerms); err != nil {
		panic(fmt.Errorf("Error trying to create log directory in '%s': %s", config.LocalLogDir, err))
	}
}

func TempFile(prefix string) (*os.File, error) {
	if checkedTempDir != TempDir {
		if err := os.MkdirAll(TempDir, tempDirPerms); err != nil {
			return nil, err
		}
		checkedTempDir = TempDir
	}

	return ioutil.TempFile(TempDir, prefix)
}

func ResetTempDir() error {
	checkedTempDir = ""
	return os.RemoveAll(TempDir)
}

func LocalMediaPath(oid string) (string, error) {
	return objects.BuildObjectPath(oid)
}

func LocalMediaPathReadOnly(oid string) string {
	return objects.ObjectPath(oid)
}

func LocalReferencePath(sha string) string {
	if config.LocalReferenceDir == "" {
		return ""
	}
	return filepath.Join(config.LocalReferenceDir, sha[0:2], sha[2:4], sha)
}

func ObjectExistsOfSize(oid string, size int64) bool {
	path := objects.ObjectPath(oid)
	return tools.FileExistsOfSize(path, size)
}

func Environ() []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron)+7)
	env = append(env,
		fmt.Sprintf("LocalWorkingDir=%s", config.LocalWorkingDir),
		fmt.Sprintf("LocalGitDir=%s", config.LocalGitDir),
		fmt.Sprintf("LocalGitStorageDir=%s", config.LocalGitStorageDir),
		fmt.Sprintf("LocalMediaDir=%s", LocalMediaDir),
		fmt.Sprintf("LocalReferenceDir=%s", config.LocalReferenceDir),
		fmt.Sprintf("TempDir=%s", TempDir),
		fmt.Sprintf("ConcurrentTransfers=%d", config.Config.ConcurrentTransfers()),
		fmt.Sprintf("BatchTransfer=%v", config.Config.BatchTransfer()),
		fmt.Sprintf("SkipDownloadErrors=%v", config.Config.SkipDownloadErrors()),
		fmt.Sprintf("FetchRecentAlways=%v", config.Config.FetchPruneConfig().FetchRecentAlways),
		fmt.Sprintf("FetchRecentRefsDays=%d", config.Config.FetchPruneConfig().FetchRecentRefsDays),
		fmt.Sprintf("FetchRecentCommitsDays=%d", config.Config.FetchPruneConfig().FetchRecentCommitsDays),
		fmt.Sprintf("FetchRecentRefsIncludeRemotes=%v", config.Config.FetchPruneConfig().FetchRecentRefsIncludeRemotes),
		fmt.Sprintf("PruneOffsetDays=%d", config.Config.FetchPruneConfig().PruneOffsetDays),
		fmt.Sprintf("PruneVerifyRemoteAlways=%v", config.Config.FetchPruneConfig().PruneVerifyRemoteAlways),
		fmt.Sprintf("PruneRemoteName=%s", config.Config.FetchPruneConfig().PruneRemoteName),
		fmt.Sprintf("AccessDownload=%s", config.Config.Access("download")),
		fmt.Sprintf("AccessUpload=%s", config.Config.Access("upload")),
	)
	if len(config.Config.FetchExcludePaths()) > 0 {
		env = append(env, fmt.Sprintf("FetchExclude=%s", strings.Join(config.Config.FetchExcludePaths(), ", ")))
	}
	if len(config.Config.FetchIncludePaths()) > 0 {
		env = append(env, fmt.Sprintf("FetchInclude=%s", strings.Join(config.Config.FetchIncludePaths(), ", ")))
	}
	for _, ext := range config.Config.Extensions() {
		env = append(env, fmt.Sprintf("Extension[%d]=%s", ext.Priority, ext.Name))
	}

	for _, e := range osEnviron {
		if !strings.Contains(e, "GIT_") {
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
	if objects == nil {
		return nil
	}
	return objects.ClearTempObjects()
}

func ScanObjectsChan() <-chan localstorage.Object {
	return objects.ScanObjectsChan()
}

func init() {
	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "

	ResolveDirs()
}

const (
	gitExt       = ".git"
	gitPtrPrefix = "gitdir: "
)

// only used in tests
func AllObjects() []localstorage.Object {
	return objects.AllObjects()
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
