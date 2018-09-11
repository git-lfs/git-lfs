package storage

// Backend is an encapsulation of a set of read-only and read-write interfaces
// for reading and writing objects.
type Backend interface {
	// Storage returns a read source and optionally a write source.
	// Generally, the write location, if present, should also be a read
	// location.
	Storage() (Storage, WritableStorage)
}
