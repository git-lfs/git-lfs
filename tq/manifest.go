package tq

import (
	"sync"

	"github.com/git-lfs/git-lfs/config"
	"github.com/rubyist/tracerx"
)

type Manifest struct {
	basicTransfersOnly   bool
	downloadAdapterFuncs map[string]NewAdapterFunc
	uploadAdapterFuncs   map[string]NewAdapterFunc
	mu                   sync.Mutex
}

func NewManifest() *Manifest {
	return &Manifest{
		downloadAdapterFuncs: make(map[string]NewAdapterFunc),
		uploadAdapterFuncs:   make(map[string]NewAdapterFunc),
	}
}

func ConfigureManifest(m *Manifest, gitEnv config.Environment) *Manifest {
	m.basicTransfersOnly = gitEnv.Bool("lfs.basictransfersonly", false)
	configureBasicDownloadAdapter(m)
	configureBasicUploadAdapter(m)
	if gitEnv.Bool("lfs.tustransfers", false) {
		configureTusAdapter(m)
	}
	configureCustomAdapters(gitEnv, m)
	return m
}

// GetAdapterNames returns a list of the names of adapters available to be created
func (m *Manifest) GetAdapterNames(dir Direction) []string {
	switch dir {
	case Upload:
		return m.GetUploadAdapterNames()
	case Download:
		return m.GetDownloadAdapterNames()
	}
	return nil
}

// GetDownloadAdapterNames returns a list of the names of download adapters available to be created
func (m *Manifest) GetDownloadAdapterNames() []string {
	return m.getAdapterNames(m.downloadAdapterFuncs)
}

// GetUploadAdapterNames returns a list of the names of upload adapters available to be created
func (m *Manifest) GetUploadAdapterNames() []string {
	return m.getAdapterNames(m.uploadAdapterFuncs)
}

// getAdapterNames returns a list of the names of adapters available to be created
func (m *Manifest) getAdapterNames(adapters map[string]NewAdapterFunc) []string {
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
func (m *Manifest) RegisterNewAdapterFunc(name string, dir Direction, f NewAdapterFunc) {
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
func (m *Manifest) NewAdapterOrDefault(name string, dir Direction) Adapter {
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
func (m *Manifest) NewAdapter(name string, dir Direction) Adapter {
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
func (m *Manifest) NewDownloadAdapter(name string) Adapter {
	return m.NewAdapterOrDefault(name, Download)
}

// Create a new upload adapter by name, or BasicAdapterName if doesn't exist
func (m *Manifest) NewUploadAdapter(name string) Adapter {
	return m.NewAdapterOrDefault(name, Upload)
}
