package backends

// Backend defines the interface for URL shortening backends
type Backend interface {
	// Get retrieves the long URL for a given short code
	Get(shortCode string) (string, error)

	// Set stores a mapping from short code to long URL
	// Returns an error if the backend is read-only or if the operation fails
	Set(shortCode string, longURL string) error

	// Close releases any resources held by the backend
	Close() error
}
