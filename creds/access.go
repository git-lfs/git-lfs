package creds

type AccessMode string

const (
	NoneAccess      AccessMode = "none"
	BasicAccess     AccessMode = "basic"
	PrivateAccess   AccessMode = "private"
	NegotiateAccess AccessMode = "negotiate"
	EmptyAccess     AccessMode = ""
)

type Access struct {
	mode AccessMode
	url  string
}

func NewAccess(mode AccessMode, url string) Access {
	return Access{url: url, mode: mode}
}

// Returns a copy of an AccessMode with the mode upgraded to newMode
func (a *Access) Upgrade(newMode AccessMode) Access {
	return Access{url: a.url, mode: newMode}
}

func (a *Access) Mode() AccessMode {
	return a.mode
}

func (a *Access) URL() string {
	return a.url
}
