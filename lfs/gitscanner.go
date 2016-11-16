package lfs

type GitScanner struct {
}

func NewGitScanner() *GitScanner {
	return &GitScanner{}
}

func (s *GitScanner) ScanRef(ref string) (*PointerChannelWrapper, error) {
	opts := NewScanRefsOptions()
	opts.ScanMode = ScanRefsMode
	opts.SkipDeletedBlobs = true
	return ScanRefsToChan(ref, "", opts)
}

func (s *GitScanner) ScanAll() (*PointerChannelWrapper, error) {
	opts := NewScanRefsOptions()
	opts.ScanMode = ScanAllMode
	opts.SkipDeletedBlobs = false
	return ScanRefsToChan("", "", opts)
}
