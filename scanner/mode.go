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
	nameMap          map[string]string
	mutex            *sync.Mutex
}

func (o *ScanRefsOptions) GetName(sha string) (string, bool) {
	o.mutex.Lock()
	name, ok := o.nameMap[sha]
	o.mutex.Unlock()
	return name, ok
}

func (o *ScanRefsOptions) SetName(sha, name string) {
	o.mutex.Lock()
	o.nameMap[sha] = name
	o.mutex.Unlock()
}

func NewScanRefsOptions() *ScanRefsOptions {
	return &ScanRefsOptions{
		nameMap: make(map[string]string, 0),
		mutex:   &sync.Mutex{},
	}
}
