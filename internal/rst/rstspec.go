// Package rst provides utilities for parsing reStructuredText files and related configuration.
package rst

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// RstspecURL is the URL to the canonical rstspec.toml file in the snooty-parser repository.
const RstspecURL = "https://raw.githubusercontent.com/mongodb/snooty-parser/refs/heads/main/snooty/rstspec.toml"

// RstspecCacheTTL is the time-to-live for the cached rstspec.toml (24 hours).
const RstspecCacheTTL = 24 * time.Hour

// RstspecCacheDir is the directory for storing cache files.
const RstspecCacheDir = ".audit-cli"

// RstspecCacheFileName is the name of the rstspec cache file.
const RstspecCacheFileName = "rstspec-cache.json"

// RstspecComposable represents a composable definition from rstspec.toml.
type RstspecComposable struct {
	ID           string                       `toml:"id"`
	Title        string                       `toml:"title"`
	Default      string                       `toml:"default"`
	Dependencies []map[string]string          `toml:"dependencies"`
	Options      []RstspecComposableOption    `toml:"options"`
}

// RstspecComposableOption represents an option within a composable.
type RstspecComposableOption struct {
	ID    string `toml:"id"`
	Title string `toml:"title"`
}

// RstspecTabOption represents a tab option within a tabset.
type RstspecTabOption struct {
	ID    string `toml:"id"`
	Title string `toml:"title"`
}

// RstspecConfig represents the structure of the rstspec.toml file.
// This includes all sections, though most commands will only use specific parts.
type RstspecConfig struct {
	Composables []RstspecComposable `toml:"composables"`
	// Tabs contains tabset definitions (e.g., drivers, platforms, cloud-providers)
	Tabs map[string][]RstspecTabOption `toml:"tabs"`
	// Additional sections can be added here as needed:
	// Directives  map[string]interface{} `toml:"directive"`
	// Roles       map[string]interface{} `toml:"role"`
	// etc.
}

// GetComposableOptionTitle returns the human-readable title for a composable option.
// For example, GetComposableOptionTitle("language", "nodejs") returns "Node.js".
func (c *RstspecConfig) GetComposableOptionTitle(composableID, optionID string) (string, bool) {
	for _, comp := range c.Composables {
		if comp.ID == composableID {
			for _, opt := range comp.Options {
				if opt.ID == optionID {
					return opt.Title, true
				}
			}
		}
	}
	return "", false
}

// GetTabOptionTitle returns the human-readable title for a tab option.
// For example, GetTabOptionTitle("drivers", "nodejs") returns "Node.js".
func (c *RstspecConfig) GetTabOptionTitle(tabsetID, optionID string) (string, bool) {
	if tabset, ok := c.Tabs[tabsetID]; ok {
		for _, opt := range tabset {
			if opt.ID == optionID {
				return opt.Title, true
			}
		}
	}
	return "", false
}

// BuildComposableIDToTitleMap builds a map from option ID to title for a specific composable.
// For example, BuildComposableIDToTitleMap("language") returns {"nodejs": "Node.js", "python": "Python", ...}.
func (c *RstspecConfig) BuildComposableIDToTitleMap(composableID string) map[string]string {
	result := make(map[string]string)
	for _, comp := range c.Composables {
		if comp.ID == composableID {
			for _, opt := range comp.Options {
				result[opt.ID] = opt.Title
			}
			break
		}
	}
	return result
}

// BuildTabIDToTitleMap builds a map from tab ID to title for a specific tabset.
// For example, BuildTabIDToTitleMap("drivers") returns {"nodejs": "Node.js", "python": "Python", ...}.
func (c *RstspecConfig) BuildTabIDToTitleMap(tabsetID string) map[string]string {
	result := make(map[string]string)
	if tabset, ok := c.Tabs[tabsetID]; ok {
		for _, opt := range tabset {
			result[opt.ID] = opt.Title
		}
	}
	return result
}

// RstspecCache represents the cached rstspec.toml data.
type RstspecCache struct {
	Timestamp   time.Time         `json:"timestamp"`
	Composables []RstspecComposable `json:"composables"`
	Tabs        map[string][]RstspecTabOption `json:"tabs"`
}

// getRstspecCachePath returns the path to the rstspec cache file.
func getRstspecCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, RstspecCacheDir, RstspecCacheFileName), nil
}

// loadRstspecCache loads the rstspec from the cache file.
func loadRstspecCache() (*RstspecConfig, error) {
	cachePath, err := getRstspecCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache RstspecCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse rstspec cache: %w", err)
	}

	// Check if cache is expired
	if time.Since(cache.Timestamp) > RstspecCacheTTL {
		return nil, fmt.Errorf("rstspec cache expired")
	}

	return &RstspecConfig{
		Composables: cache.Composables,
		Tabs:        cache.Tabs,
	}, nil
}

// saveRstspecCache saves the rstspec to the cache file.
func saveRstspecCache(config *RstspecConfig) error {
	cachePath, err := getRstspecCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := RstspecCache{
		Timestamp:   time.Now(),
		Composables: config.Composables,
		Tabs:        config.Tabs,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rstspec cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rstspec cache: %w", err)
	}

	return nil
}

// fetchRstspecFromURL fetches and parses rstspec.toml from the remote URL.
func fetchRstspecFromURL() (*RstspecConfig, error) {
	resp, err := http.Get(RstspecURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch rstspec.toml: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read rstspec.toml: %w", err)
	}

	var config RstspecConfig
	if err := toml.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rstspec.toml: %w", err)
	}

	return &config, nil
}

// FetchRstspec fetches and parses the canonical rstspec.toml file.
//
// This function uses a local cache (stored in ~/.audit-cli/rstspec-cache.json)
// to avoid repeated network requests. The cache has a 24-hour TTL.
// If the cache is missing or expired, it fetches from the snooty-parser repository.
// If the network request fails and a cached version exists (even if expired),
// it falls back to the cached version for offline support.
//
// Returns:
//   - *RstspecConfig: The parsed rstspec configuration
//   - error: Any error encountered during fetch or parse
//
// Example:
//
//	config, err := rst.FetchRstspec()
//	if err != nil {
//	    return fmt.Errorf("failed to fetch rstspec: %w", err)
//	}
//	fmt.Printf("Found %d composables\n", len(config.Composables))
func FetchRstspec() (*RstspecConfig, error) {
	// Try to load from cache first
	config, err := loadRstspecCache()
	if err == nil {
		return config, nil
	}

	// Cache miss or expired, try to fetch from URL
	config, fetchErr := fetchRstspecFromURL()
	if fetchErr != nil {
		// Network failed - try to use expired cache as fallback for offline support
		cachePath, pathErr := getRstspecCachePath()
		if pathErr == nil {
			if data, readErr := os.ReadFile(cachePath); readErr == nil {
				var cache RstspecCache
				if jsonErr := json.Unmarshal(data, &cache); jsonErr == nil {
					// Return expired cache with a warning
					fmt.Fprintf(os.Stderr, "Warning: Could not fetch rstspec.toml (%v), using expired cache\n", fetchErr)
					return &RstspecConfig{
						Composables: cache.Composables,
						Tabs:        cache.Tabs,
					}, nil
				}
			}
		}
		return nil, fetchErr
	}

	// Save to cache for next time
	if saveErr := saveRstspecCache(config); saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not save rstspec cache: %v\n", saveErr)
	}

	return config, nil
}
