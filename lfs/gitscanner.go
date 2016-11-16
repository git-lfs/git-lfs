package lfs

type GitScanner struct {
}

func NewGitScanner() *GitScanner {
	return &GitScanner{}
}

func (s *GitScanner) ScanRefRange(left, right string) (*PointerChannelWrapper, error) {
	opts := NewScanRefsOptions()
	opts.ScanMode = ScanRefsMode
	opts.SkipDeletedBlobs = false
	return ScanRefsToChan(left, right, opts)
}

func (s *GitScanner) ScanRefWithDeleted(ref string) (*PointerChannelWrapper, error) {
	return s.ScanRefRange(ref, "")
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
