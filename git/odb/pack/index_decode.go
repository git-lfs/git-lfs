package pack

const (
	// MagicWidth is the width of the magic header of packfiles version 2
	// and newer.
	MagicWidth = 4
	// VersionWidth is the width of the version following the magic header.
	VersionWidth = 4
	// V1Width is the total width of the header in V1.
	V1Width = 0
	// V2Width is the total width of the header in V2.
	V2Width = MagicWidth + VersionWidth
)
