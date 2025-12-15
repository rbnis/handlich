package backends

import (
	"fmt"

	"github.com/rbnis/handlich/internal/backends/file"
	"github.com/rbnis/handlich/internal/backends/memory"
	"github.com/rbnis/handlich/internal/config"
)

// NewBackend creates a backend based on the provided configuration
func NewBackend(cfg *config.Config) (Backend, error) {
	switch cfg.Backend.Type {
	case config.BackendTypeMemory:
		return memory.New(), nil

	case config.BackendTypeFile:
		if cfg.Backend.File.Path == "" {
			return nil, fmt.Errorf("file backend requires a file path")
		}
		return file.New(cfg.Backend.File.Path)

	case config.BackendTypeRedis:
		return nil, fmt.Errorf("redis backend not yet implemented")

	case config.BackendTypeSqlite:
		return nil, fmt.Errorf("sqlite backend not yet implemented")

	default:
		return nil, fmt.Errorf("unknown backend type: %s", cfg.Backend.Type)
	}
}
