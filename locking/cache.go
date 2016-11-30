package locking

// This file caches active locks locally so that we can more easily retrieve
// a list of locally locked files without consulting the server
// This only includes locks which the local committer has taken, not all locks

// Cache a successful lock for faster local lookup later
func cacheLock(filePath, id string) error {
	// TODO
	return nil
}

// Remove a cached lock by path becuase it's been relinquished
func cacheUnlock(filePath string) error {
	// TODO
	return nil
}

// Remove a cached lock by id becuase it's been relinquished
func cacheUnlockById(id string) error {
	// TODO
	return nil
}

// Get the list of cached locked files
func cachedLocks() []string {
	// TODO
	return nil
}

// Fetch locked files for the current committer and cache them locally
// This can be used to sync up locked files when moving machines
func fetchLocksToCache(remoteName string) error {
	// TODO
	return nil
}
