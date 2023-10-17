package tq

import (
	"strings"
	"sync"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/fs"
	"github.com/git-lfs/git-lfs/v3/lfsapi"
	"github.com/git-lfs/git-lfs/v3/ssh"
	"github.com/rubyist/tracerx"
)

const (
	defaultMaxRetries          = 8
	defaultMaxRetryDelay       = 10
	defaultConcurrentTransfers = 8
)

type Manifest interface {
	APIClient() *lfsapi.Client
	MaxRetries() int
	MaxRetryDelay() int
	ConcurrentTransfers() int
	IsStandaloneTransfer() bool
	batchClient() BatchClient
	GetAdapterNames(dir Direction) []string
	GetDownloadAdapterNames() []string
	GetUploadAdapterNames() []string
	getAdapterNames(adapters map[string]NewAdapterFunc) []string
	RegisterNewAdapterFunc(name string, dir Direction, f NewAdapterFunc)
	NewAdapterOrDefault(name string, dir Direction) Adapter
	NewAdapter(name string, dir Direction) Adapter
	NewDownloadAdapter(name string) Adapter
	NewUploadAdapter(name string) Adapter
	Upgrade() *concreteManifest
	Upgraded() bool
}

type lazyManifest struct {
	f         *fs.Filesystem
	apiClient *lfsapi.Client
	operation string
	remote    string
	m         *concreteManifest
}

func newLazyManifest(f *fs.Filesystem, apiClient *lfsapi.Client, operation, remote string) *lazyManifest {
	return &lazyManifest{
		f:         f,
		apiClient: apiClient,
		operation: operation,
		remote:    remote,
		m:         nil,
	}
}

func (m *lazyManifest) APIClient() *lfsapi.Client {
	return m.Upgrade().APIClient()
}

func (m *lazyManifest) MaxRetries() int {
	return m.Upgrade().MaxRetries()
}

func (m *lazyManifest) MaxRetryDelay() int {
	return m.Upgrade().MaxRetryDelay()
}

func (m *lazyManifest) ConcurrentTransfers() int {
	return m.Upgrade().ConcurrentTransfers()
}

func (m *lazyManifest) IsStandaloneTransfer() bool {
	return m.Upgrade().IsStandaloneTransfer()
}

func (m *lazyManifest) batchClient() BatchClient {
	return m.Upgrade().batchClient()
}

func (m *lazyManifest) GetAdapterNames(dir Direction) []string {
	return m.Upgrade().GetAdapterNames(dir)
}

func (m *lazyManifest) GetDownloadAdapterNames() []string {
	return m.Upgrade().GetDownloadAdapterNames()
}

func (m *lazyManifest) GetUploadAdapterNames() []string {
	return m.Upgrade().GetUploadAdapterNames()
}

func (m *lazyManifest) getAdapterNames(adapters map[string]NewAdapterFunc) []string {
	return m.Upgrade().getAdapterNames(adapters)
}

func (m *lazyManifest) RegisterNewAdapterFunc(name string, dir Direction, f NewAdapterFunc) {
	m.Upgrade().RegisterNewAdapterFunc(name, dir, f)
}

func (m *lazyManifest) NewAdapterOrDefault(name string, dir Direction) Adapter {
	return m.Upgrade().NewAdapterOrDefault(name, dir)
}

func (m *lazyManifest) NewAdapter(name string, dir Direction) Adapter {
	return m.Upgrade().NewAdapter(name, dir)
}

func (m *lazyManifest) NewDownloadAdapter(name string) Adapter {
	return m.Upgrade().NewDownloadAdapter(name)
}

func (m *lazyManifest) NewUploadAdapter(name string) Adapter {
	return m.Upgrade().NewUploadAdapter(name)
}

func (m *lazyManifest) Upgrade() *concreteManifest {
	if m.m == nil {
		m.m = newConcreteManifest(m.f, m.apiClient, m.operation, m.remote)
	}
	return m.m
}

func (m *lazyManifest) Upgraded() bool {
	return m.m != nil
}

type concreteManifest struct {
	// maxRetries is the maximum number of retries a single object can
	// attempt to make before it will be dropped. maxRetryDelay is the maximum
	// time in seconds to wait between retry attempts when using backoff.
	maxRetries              int
	maxRetryDelay           int
	concurrentTransfers     int
	basicTransfersOnly      bool
	standaloneTransferAgent string
	tusTransfersAllowed     bool
	downloadAdapterFuncs    map[string]NewAdapterFunc
	uploadAdapterFuncs      map[string]NewAdapterFunc
	fs                      *fs.Filesystem
	apiClient               *lfsapi.Client
	sshTransfer             *ssh.SSHTransfer
	batchClientAdapter      BatchClient
	mu                      sync.Mutex
}

func (m *concreteManifest) APIClient() *lfsapi.Client {
	return m.apiClient
}

func (m *concreteManifest) MaxRetries() int {
	return m.maxRetries
}

func (m *concreteManifest) MaxRetryDelay() int {
	return m.maxRetryDelay
}

func (m *concreteManifest) ConcurrentTransfers() int {
	return m.concurrentTransfers
}

func (m *concreteManifest) IsStandaloneTransfer() bool {
	return m.standaloneTransferAgent != ""
}

func (m *concreteManifest) batchClient() BatchClient {
	if r := m.MaxRetries(); r > 0 {
		m.batchClientAdapter.SetMaxRetries(r)
	}
	return m.batchClientAdapter
}

func (m *concreteManifest) Upgrade() *concreteManifest {
	return m
}

func (m *concreteManifest) Upgraded() bool {
	return true
}

func NewManifest(f *fs.Filesystem, apiClient *lfsapi.Client, operation, remote string) Manifest {
	return newLazyManifest(f, apiClient, operation, remote)
}

func newConcreteManifest(f *fs.Filesystem, apiClient *lfsapi.Client, operation, remote string) *concreteManifest {
	if apiClient == nil {
		cli, err := lfsapi.NewClient(nil)
		if err != nil {
			tracerx.Printf("unable to init tq.Manifest: %s", err)
			return nil
		}
		apiClient = cli
	}

	sshTransfer := apiClient.SSHTransfer(operation, remote)
	useSSHMultiplexing := false
	if sshTransfer != nil {
		useSSHMultiplexing = sshTransfer.IsMultiplexingEnabled()
	}

	m := &concreteManifest{
		fs:                   f,
		apiClient:            apiClient,
		batchClientAdapter:   &tqClient{Client: apiClient},
		downloadAdapterFuncs: make(map[string]NewAdapterFunc),
		uploadAdapterFuncs:   make(map[string]NewAdapterFunc),
		sshTransfer:          sshTransfer,
	}

	var tusAllowed bool
	if git := apiClient.GitEnv(); git != nil {
		if v := git.Int("lfs.transfer.maxretries", 0); v > 0 {
			m.maxRetries = v
		}
		if v := git.Int("lfs.transfer.maxretrydelay", -1); v > -1 {
			m.maxRetryDelay = v
		}
		if v := git.Int("lfs.concurrenttransfers", 0); v > 0 {
			m.concurrentTransfers = v
		}
		m.basicTransfersOnly = git.Bool("lfs.basictransfersonly", false)
		m.standaloneTransferAgent = findStandaloneTransfer(
			apiClient, operation, remote,
		)
		tusAllowed = git.Bool("lfs.tustransfers", false)
		configureCustomAdapters(git, m)
	}

	if m.maxRetries < 1 {
		m.maxRetries = defaultMaxRetries
	}
	if m.maxRetryDelay < 1 {
		m.maxRetryDelay = defaultMaxRetryDelay
	}

	if m.concurrentTransfers < 1 {
		m.concurrentTransfers = defaultConcurrentTransfers
	}

	if sshTransfer != nil {
		if !useSSHMultiplexing {
			m.concurrentTransfers = 1
		}

		// Multiple concurrent transfers are not yet supported.
		m.batchClientAdapter = &SSHBatchClient{
			maxRetries: m.maxRetries,
			transfer:   sshTransfer,
		}
	}

	configureBasicDownloadAdapter(m)
	configureBasicUploadAdapter(m)
	if tusAllowed {
		configureTusAdapter(m)
	}
	configureSSHAdapter(m)
	return m
}

func findDefaultStandaloneTransfer(url string) string {
	if strings.HasPrefix(url, "file://") {
		return standaloneFileName
	}
	return ""
}

func findStandaloneTransfer(client *lfsapi.Client, operation, remote string) string {
	if operation == "" || remote == "" {
		v, _ := client.GitEnv().Get("lfs.standalonetransferagent")
		return v
	}

	ep := client.Endpoints.Endpoint(operation, remote)
	uc := config.NewURLConfig(client.GitEnv())
	v, ok := uc.Get("lfs", ep.Url, "standalonetransferagent")
	if !ok {
		return findDefaultStandaloneTransfer(ep.Url)
	}

	return v
}

// GetAdapterNames returns a list of the names of adapters available to be created
func (m *concreteManifest) GetAdapterNames(dir Direction) []string {
	switch dir {
	case Upload:
		return m.GetUploadAdapterNames()
	case Download:
		return m.GetDownloadAdapterNames()
	}
	return nil
}

// GetDownloadAdapterNames returns a list of the names of download adapters available to be created
func (m *concreteManifest) GetDownloadAdapterNames() []string {
	return m.getAdapterNames(m.downloadAdapterFuncs)
}

// GetUploadAdapterNames returns a list of the names of upload adapters available to be created
func (m *concreteManifest) GetUploadAdapterNames() []string {
	return m.getAdapterNames(m.uploadAdapterFuncs)
}

// getAdapterNames returns a list of the names of adapters available to be created
func (m *concreteManifest) getAdapterNames(adapters map[string]NewAdapterFunc) []string {
	if m.basicTransfersOnly {
		return []string{BasicAdapterName}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ret := make([]string, 0, len(adapters))
	for n, _ := range adapters {
		ret = append(ret, n)
	}
	return ret
}

// RegisterNewTransferAdapterFunc registers a new function for creating upload
// or download adapters. If a function with that name & direction is already
// registered, it is overridden
func (m *concreteManifest) RegisterNewAdapterFunc(name string, dir Direction, f NewAdapterFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch dir {
	case Upload:
		m.uploadAdapterFuncs[name] = f
	case Download:
		m.downloadAdapterFuncs[name] = f
	}
}

// Create a new adapter by name and direction; default to BasicAdapterName if doesn't exist
func (m *concreteManifest) NewAdapterOrDefault(name string, dir Direction) Adapter {
	if len(name) == 0 {
		name = BasicAdapterName
	}

	a := m.NewAdapter(name, dir)
	if a == nil {
		tracerx.Printf("Defaulting to basic transfer adapter since %q did not exist", name)
		a = m.NewAdapter(BasicAdapterName, dir)
	}
	return a
}

// Create a new adapter by name and direction, or nil if doesn't exist
func (m *concreteManifest) NewAdapter(name string, dir Direction) Adapter {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch dir {
	case Upload:
		if u, ok := m.uploadAdapterFuncs[name]; ok {
			return u(name, dir)
		}
	case Download:
		if d, ok := m.downloadAdapterFuncs[name]; ok {
			return d(name, dir)
		}
	}
	return nil
}

// Create a new download adapter by name, or BasicAdapterName if doesn't exist
func (m *concreteManifest) NewDownloadAdapter(name string) Adapter {
	return m.NewAdapterOrDefault(name, Download)
}

// Create a new upload adapter by name, or BasicAdapterName if doesn't exist
func (m *concreteManifest) NewUploadAdapter(name string) Adapter {
	return m.NewAdapterOrDefault(name, Upload)
}

// Env is any object with a config.Environment interface.
type Env interface {
	Get(key string) (val string, ok bool)
	GetAll(key string) []string
	Bool(key string, def bool) (val bool)
	Int(key string, def int) (val int)
	All() map[string][]string
}
