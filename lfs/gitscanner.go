package lfs

type GitScanner struct {
}

func NewGitScanner() *GitScanner {
	return &GitScanner{}
}

func (s *GitScanner) ScanAll() (pointers *PointerChannelWrapper, err error) {
	opts := NewScanRefsOptions()
	opts.ScanMode = ScanAllMode
	opts.SkipDeletedBlobs = false
	return ScanRefsToChan("", "", opts)
}
