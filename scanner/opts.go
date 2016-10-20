package scanner

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

	nameCache *nameCache
}

func (o *ScanRefsOptions) GetName(sha string) (string, bool) {
	return o.nameCache.GetName(sha)
}

func (o *ScanRefsOptions) SetName(sha, name string) {
	o.nameCache.Cache(sha, name)
}

func NewScanRefsOptions() *ScanRefsOptions {
	return &ScanRefsOptions{
		nameCache: newNameCache(),
	}
}
