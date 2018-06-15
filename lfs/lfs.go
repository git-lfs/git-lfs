// Package lfs brings together the core LFS functionality
// NOTE: Subject to change, do not rely on this package from outside git-lfs source
package lfs

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/git-lfs/git-lfs/config"
	"github.com/git-lfs/git-lfs/lfsapi"
	"github.com/git-lfs/git-lfs/tools"
	"github.com/git-lfs/git-lfs/tq"
	"github.com/rubyist/tracerx"
)

func Environ(cfg *config.Configuration, manifest *tq.Manifest) []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron)+7)

	api, err := lfsapi.NewClient(cfg)
	if err != nil {
		// TODO(@ttaylorr): don't panic
		panic(err.Error())
	}

	download := api.Endpoints.AccessFor(api.Endpoints.Endpoint("download", cfg.Remote()).Url)
	upload := api.Endpoints.AccessFor(api.Endpoints.Endpoint("upload", cfg.PushRemote()).Url)

	dltransfers := manifest.GetDownloadAdapterNames()
	sort.Strings(dltransfers)
	ultransfers := manifest.GetUploadAdapterNames()
	sort.Strings(ultransfers)

	fetchPruneConfig := NewFetchPruneConfig(cfg.Git)

	env = append(env,
		fmt.Sprintf("LocalWorkingDir=%s", cfg.LocalWorkingDir()),
		fmt.Sprintf("LocalGitDir=%s", cfg.LocalGitDir()),
		fmt.Sprintf("LocalGitStorageDir=%s", cfg.LocalGitStorageDir()),
		fmt.Sprintf("LocalMediaDir=%s", cfg.LFSObjectDir()),
		fmt.Sprintf("LocalReferenceDir=%s", cfg.LocalReferenceDir()),
		fmt.Sprintf("TempDir=%s", cfg.TempDir()),
		fmt.Sprintf("ConcurrentTransfers=%d", api.ConcurrentTransfers),
		fmt.Sprintf("TusTransfers=%v", cfg.TusTransfersAllowed()),
		fmt.Sprintf("BasicTransfersOnly=%v", cfg.BasicTransfersOnly()),
		fmt.Sprintf("SkipDownloadErrors=%v", cfg.SkipDownloadErrors()),
		fmt.Sprintf("FetchRecentAlways=%v", fetchPruneConfig.FetchRecentAlways),
		fmt.Sprintf("FetchRecentRefsDays=%d", fetchPruneConfig.FetchRecentRefsDays),
		fmt.Sprintf("FetchRecentCommitsDays=%d", fetchPruneConfig.FetchRecentCommitsDays),
		fmt.Sprintf("FetchRecentRefsIncludeRemotes=%v", fetchPruneConfig.FetchRecentRefsIncludeRemotes),
		fmt.Sprintf("PruneOffsetDays=%d", fetchPruneConfig.PruneOffsetDays),
		fmt.Sprintf("PruneVerifyRemoteAlways=%v", fetchPruneConfig.PruneVerifyRemoteAlways),
		fmt.Sprintf("PruneRemoteName=%s", fetchPruneConfig.PruneRemoteName),
		fmt.Sprintf("LfsStorageDir=%s", cfg.LFSStorageDir()),
		fmt.Sprintf("AccessDownload=%s", download),
		fmt.Sprintf("AccessUpload=%s", upload),
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

func init() {
	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "
	if len(os.Getenv("GIT_TRACE")) < 1 {
		if tt := os.Getenv("GIT_TRANSFER_TRACE"); len(tt) > 0 {
			os.Setenv("GIT_TRACE", tt)
		}
	}
}

const (
	gitExt       = ".git"
	gitPtrPrefix = "gitdir: "
)

func LinkOrCopyFromReference(cfg *config.Configuration, oid string, size int64) error {
	if cfg.LFSObjectExists(oid, size) {
		return nil
	}
	altMediafile := cfg.Filesystem().ObjectReferencePath(oid)
	mediafile, err := cfg.Filesystem().ObjectPath(oid)
	if err != nil {
		return err
	}
	if altMediafile != "" && tools.FileExistsOfSize(altMediafile, size) {
		return LinkOrCopy(cfg, altMediafile, mediafile)
	}
	return nil
}
