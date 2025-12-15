package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/rbnis/handlich/internal/backends"
	"github.com/rbnis/handlich/internal/backends/file"
	"github.com/rbnis/handlich/internal/backends/memory"
	"github.com/rbnis/handlich/internal/config"
	"github.com/rbnis/handlich/internal/logger"
)

var backend backends.Backend

func main() {
	configPath := flag.String("config", "configs/handlich.yaml", "Path to the configuration file")
	flag.Parse()

	log := logger.Get()

	config, err := config.Load(*configPath)
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.SetLevel(config.LogLevel)

	// Initialize backend
	backend, err = backends.NewBackend(config)
	if err != nil {
		log.Error("Failed to initialize backend", "error", err)
		os.Exit(1)
	}
	defer backend.Close()

	log.Info("Backend initialized", "type", config.Backend.Type)

	//http.HandleFunc("/api/*", apiHandler)
	http.HandleFunc("/", redirectHandler)

	log.Info("Starting http server on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Get()

	// Extract short code from path (remove leading slash)
	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	// Handle root path
	if shortCode == "" {
		http.Error(w, "URL shortener service", http.StatusOK)
		return
	}

	// Get long URL from backend
	longURL, err := backend.Get(shortCode)
	if err != nil {
		if errors.Is(err, memory.ErrNotFound) || errors.Is(err, file.ErrNotFound) {
			log.Info("Short code not found", "shortCode", shortCode)
			http.NotFound(w, r)
			return
		}

		log.Error("Failed to get redirect", "shortCode", shortCode, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log and redirect
	log.Info("Redirecting", "shortCode", shortCode, "longURL", longURL)
	http.Redirect(w, r, longURL, http.StatusSeeOther)
}
