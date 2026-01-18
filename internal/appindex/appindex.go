package appindex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/reflective-technologies/kiosk-cli/internal/config"
)

// AppEntry represents a single installed app
type AppEntry struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	GitUrl      string    `json:"gitUrl"`
	InstalledAt time.Time `json:"installedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Index holds all installed apps
type Index struct {
	Apps map[string]*AppEntry `json:"apps"` // keyed by "org/repo"
}

// indexFileName is the name of the index file
const indexFileName = "apps.json"

// IndexPath returns the path to ~/.kiosk/apps.json
func IndexPath() string {
	return filepath.Join(config.KioskDir(), indexFileName)
}

// Load reads the app index from disk
func Load() (*Index, error) {
	idx := &Index{
		Apps: make(map[string]*AppEntry),
	}

	data, err := os.ReadFile(IndexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return idx, nil // Return empty index if file doesn't exist
		}
		return nil, err
	}

	if err := json.Unmarshal(data, idx); err != nil {
		return nil, err
	}

	// Ensure Apps map is initialized
	if idx.Apps == nil {
		idx.Apps = make(map[string]*AppEntry)
	}

	return idx, nil
}

// Save writes the app index to disk
func Save(idx *Index) error {
	if err := config.EnsureInitialized(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(IndexPath(), data, 0644)
}

// Get returns an app entry by key (org/repo), or nil if not found
func (idx *Index) Get(key string) *AppEntry {
	return idx.Apps[key]
}

// Add adds or updates an app in the index
func (idx *Index) Add(key string, entry *AppEntry) {
	if entry.InstalledAt.IsZero() {
		entry.InstalledAt = time.Now()
	}
	entry.UpdatedAt = time.Now()
	idx.Apps[key] = entry
}

// Remove removes an app from the index
func (idx *Index) Remove(key string) {
	delete(idx.Apps, key)
}

// Has checks if an app is in the index
func (idx *Index) Has(key string) bool {
	_, ok := idx.Apps[key]
	return ok
}

// List returns all app keys
func (idx *Index) List() []string {
	keys := make([]string, 0, len(idx.Apps))
	for k := range idx.Apps {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of apps in the index
func (idx *Index) Count() int {
	return len(idx.Apps)
}

// ValidateFilesystem checks if each app's directory exists
// Returns a map of key -> exists
func (idx *Index) ValidateFilesystem() map[string]bool {
	result := make(map[string]bool)
	for key := range idx.Apps {
		path := filepath.Join(config.AppsDir(), key)
		_, err := os.Stat(path)
		result[key] = err == nil
	}
	return result
}
