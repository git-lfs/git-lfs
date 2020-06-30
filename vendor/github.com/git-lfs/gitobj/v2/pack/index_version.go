package pack

type IndexVersion interface {
	// Name returns the name of the object located at the given offset "at",
	// in the Index file "idx".
	//
	// It returns an error if the object at that location could not be
	// parsed.
	Name(idx *Index, at int64) ([]byte, error)

	// Entry parses and returns the full *IndexEntry located at the offset
	// "at" in the Index file "idx".
	//
	// If there was an error parsing the IndexEntry at that location, it
	// will be returned.
	Entry(idx *Index, at int64) (*IndexEntry, error)

	// Width returns the number of bytes occupied by the header of a
	// particular index version.
	Width() int64
}
