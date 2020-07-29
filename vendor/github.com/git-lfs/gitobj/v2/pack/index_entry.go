package pack

// IndexEntry specifies data encoded into an entry in the pack index.
type IndexEntry struct {
	// PackOffset is the number of bytes before the associated object in a
	// packfile.
	PackOffset uint64
}
