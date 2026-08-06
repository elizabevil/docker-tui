package config

import "sync"

// CachedOperations wraps LoadOperations in a process-wide sync.Once so
// callers (the action bar, the keyboard dispatcher, future Move
// entry-points) can share the same parsed registry without re-reading
// the embedded JSONC on every call.
//
// When user-level overrides are added in a future iteration this is the
// single seam that needs to learn about them: the loader signature
// gains a user-dir argument and the cache key incorporates the path.
var (
	cachedOperations *Operations
	cachedOperationsErr error
	cachedOperationsOnce sync.Once
)

// CachedLoadOperations returns the shared Operations handle, loading
// from the embedded JSONC on first use. The returned pointer is stable
// across calls and safe to read concurrently.
func CachedLoadOperations() (*Operations, error) {
	cachedOperationsOnce.Do(func() {
		cachedOperations, cachedOperationsErr = LoadOperations()
	})
	return cachedOperations, cachedOperationsErr
}

// ResetCachedOperationsForTest clears the shared cache. Tests that
// mutate the embedded files (none today) call it between cases.
func ResetCachedOperationsForTest() {
	cachedOperations = nil
	cachedOperationsErr = nil
	cachedOperationsOnce = sync.Once{}
}