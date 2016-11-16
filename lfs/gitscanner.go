package lfs

type GitScanner struct {
	remote      string
	skippedRefs []string
}

func NewGitScanner() *GitScanner {
	return &GitScanner{}
}

func (s *GitScanner) Remote(r string) {
	s.remote = r
	s.skippedRefs = calcSkippedRefs(r)
}

func (s *GitScanner) ScanLeftToRemote(left string) (*PointerChannelWrapper, error) {
	return scanRefsToChan(left, "", s.opts(ScanLeftToRemoteMode))
}

func (s *GitScanner) ScanRefRange(left, right string) (*PointerChannelWrapper, error) {
	opts := s.opts(ScanRefsMode)
	opts.SkipDeletedBlobs = false
	return scanRefsToChan(left, right, opts)
}

func (s *GitScanner) ScanRefWithDeleted(ref string) (*PointerChannelWrapper, error) {
	return s.ScanRefRange(ref, "")
}

func (s *GitScanner) ScanRef(ref string) (*PointerChannelWrapper, error) {
	opts := s.opts(ScanRefsMode)
	opts.SkipDeletedBlobs = true
	return scanRefsToChan(ref, "", opts)
}

func (s *GitScanner) ScanAll() (*PointerChannelWrapper, error) {
	opts := s.opts(ScanAllMode)
	opts.SkipDeletedBlobs = false
	return scanRefsToChan("", "", opts)
}

func (s *GitScanner) opts(mode ScanningMode) *ScanRefsOptions {
	opts := newScanRefsOptions()
	opts.ScanMode = mode
	opts.RemoteName = s.remote
	opts.skippedRefs = s.skippedRefs
	return opts
}
