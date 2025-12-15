package file

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/rbnis/handlich/internal/logger"
	"go.yaml.in/yaml/v3"
)

var (
	ErrNotFound = errors.New("short code not found")
	ErrReadOnly = errors.New("file backend is read-only")
)

// Redirect represents a single redirect mapping
type Redirect struct {
	Short string `yaml:"short"`
	Long  string `yaml:"long"`
}

// RedirectsFile represents the structure of the redirects YAML file
type RedirectsFile struct {
	Redirects []Redirect `yaml:"redirects"`
}

// Backend implements a file-based URL shortening backend
type Backend struct {
	mu           sync.RWMutex
	store        map[string]string
	filePath     string
	lastModTime  time.Time
	stopCh       chan struct{}
	reloadTicker *time.Ticker
}

// New creates a new file-based backend that loads from the specified file
func New(filePath string) (*Backend, error) {
	b := &Backend{
		store:        make(map[string]string),
		filePath:     filePath,
		stopCh:       make(chan struct{}),
		reloadTicker: time.NewTicker(5 * time.Second),
	}

	// Initial load
	if err := b.reload(); err != nil {
		return nil, err
	}

	// Start watching for changes
	go b.watchFile()

	return b, nil
}

// reload reads the file and updates the in-memory store
func (b *Backend) reload() error {
	fileInfo, err := os.Stat(b.filePath)
	if err != nil {
		return err
	}

	// Check if file has been modified
	if !b.lastModTime.IsZero() && !fileInfo.ModTime().After(b.lastModTime) {
		return nil // No changes
	}

	data, err := os.ReadFile(b.filePath)
	if err != nil {
		return err
	}

	var redirectsFile RedirectsFile
	if err := yaml.Unmarshal(data, &redirectsFile); err != nil {
		return err
	}

	// Build new store
	newStore := make(map[string]string)
	for _, redirect := range redirectsFile.Redirects {
		newStore[redirect.Short] = redirect.Long
	}

	// Update store atomically
	b.mu.Lock()
	b.store = newStore
	b.lastModTime = fileInfo.ModTime()
	b.mu.Unlock()

	logger.Get().Info("Reloaded redirects from file",
		"file", b.filePath,
		"count", len(newStore),
	)

	return nil
}

// watchFile periodically checks for file changes and reloads
func (b *Backend) watchFile() {
	for {
		select {
		case <-b.stopCh:
			return
		case <-b.reloadTicker.C:
			if err := b.reload(); err != nil {
				logger.Get().Error("Failed to reload redirects",
					"error", err,
					"file", b.filePath,
				)
			}
		}
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

// Set is not supported for read-only file backend
func (b *Backend) Set(shortCode string, longURL string) error {
	return ErrReadOnly
}

// Close stops the file watcher and releases resources
func (b *Backend) Close() error {
	if b.reloadTicker != nil {
		b.reloadTicker.Stop()
	}
	close(b.stopCh)
	return nil
}
