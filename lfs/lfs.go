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

func Environ(cfg *config.Configuration, manifest *tq.Manifest, envOverrides map[string]string) []string {
	osEnviron := os.Environ()
	env := make([]string, 0, len(osEnviron)+7)

	api, err := lfsapi.NewClient(cfg)
	if err != nil {
		// TODO(@ttaylorr): don't panic
		panic(err.Error())
	}

	if envOverrides == nil {
		envOverrides = make(map[string]string, 0)
	}

	download := api.Endpoints.AccessFor(api.Endpoints.Endpoint("download", cfg.Remote()).Url)
	upload := api.Endpoints.AccessFor(api.Endpoints.Endpoint("upload", cfg.PushRemote()).Url)

	dltransfers := manifest.GetDownloadAdapterNames()
	sort.Strings(dltransfers)
	ultransfers := manifest.GetUploadAdapterNames()
	sort.Strings(ultransfers)

	fetchPruneConfig := NewFetchPruneConfig(cfg.Git)

	references := strings.Join(cfg.LocalReferenceDirs(), ", ")

	env = append(env,
		fmt.Sprintf("LocalWorkingDir=%s", cfg.LocalWorkingDir()),
		fmt.Sprintf("LocalGitDir=%s", cfg.LocalGitDir()),
		fmt.Sprintf("LocalGitStorageDir=%s", cfg.LocalGitStorageDir()),
		fmt.Sprintf("LocalMediaDir=%s", cfg.LFSObjectDir()),
		fmt.Sprintf("LocalReferenceDirs=%s", references),
		fmt.Sprintf("TempDir=%s", cfg.TempDir()),
		fmt.Sprintf("ConcurrentTransfers=%d", api.ConcurrentTransfers()),
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
		fmt.Sprintf("AccessDownload=%s", download.Mode()),
		fmt.Sprintf("AccessUpload=%s", upload.Mode()),
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
		key := strings.SplitN(e, "=", 2)[0]
		if !strings.HasPrefix(key, "GIT_") {
			continue
		}
		if val, ok := envOverrides[key]; ok {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		} else {
			env = append(env, e)
		}
	}

	return env
}

func init() {
	tracerx.DefaultKey = "GIT"
	tracerx.Prefix = "trace git-lfs: "
	if len(os.Getenv("GIT_TRACE")) < 1 {
		if tt := os.Getenv("GIT_TRANSFER_TRACE"); len(tt) > 0 {
			os.Setenv("GIT_TRACE", tt)
		} else if cv := os.Getenv("GIT_CURL_VERBOSE"); len(cv) > 0 {
			os.Setenv("GIT_TRACE", cv)
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
	altMediafiles := cfg.Filesystem().ObjectReferencePaths(oid)
	mediafile, err := cfg.Filesystem().ObjectPath(oid)
	if err != nil {
		return err
	}
	for _, altMediafile := range altMediafiles {
		tracerx.Printf("altMediafile: %s", altMediafile)
		if altMediafile != "" && tools.FileExistsOfSize(altMediafile, size) {
			err = LinkOrCopy(cfg, altMediafile, mediafile)
			if err == nil {
				break
			}
		}
	}
	return err
}
