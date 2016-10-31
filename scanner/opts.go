package scanner

import "sync"

type ScanningMode int

const (
	ScanRefsMode         = ScanningMode(iota) // 0 - or default scan mode
	ScanAllMode          = ScanningMode(iota)
	ScanLeftToRemoteMode = ScanningMode(iota)
)

type ScanRefsOptions struct {
	ScanMode         ScanningMode
	RemoteName       string
	SkipDeletedBlobs bool

	nameMap map[string]string
	mutex   *sync.Mutex
}

func (o *ScanRefsOptions) GetName(sha string) (string, bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	name, ok := o.nameMap[sha]
	return name, ok
}

func (o *ScanRefsOptions) SetName(sha, name string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.nameMap[sha] = name
}

func NewScanRefsOptions() *ScanRefsOptions {
	return &ScanRefsOptions{
		nameMap: make(map[string]string),
		mutex:   &sync.Mutex{},
	}
}
