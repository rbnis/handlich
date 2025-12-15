package memory

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("short code not found")

// Backend implements an in-memory URL shortening backend
type Backend struct {
	mu    sync.RWMutex
	store map[string]string
}

// New creates a new in-memory backend
func New() *Backend {
	return &Backend{
		store: make(map[string]string),
	}
}

// Get retrieves the long URL for a given short code
func (b *Backend) Get(shortCode string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	longURL, exists := b.store[shortCode]
	if !exists {
		return "", ErrNotFound
	}

	return longURL, nil
}

// Set stores a mapping from short code to long URL
func (b *Backend) Set(shortCode string, longURL string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.store[shortCode] = longURL
	return nil
}

// Close releases any resources held by the backend
func (b *Backend) Close() error {
	return nil
}
