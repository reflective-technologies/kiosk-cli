// Package prefetch provides background prefetching of API data for the TUI.
// It fetches data in the background when the app starts so that views can
// display content instantly when the user navigates to them.
package prefetch

import (
	"sync"
	"time"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
)

// DefaultPageSize is the number of apps to fetch per page
const DefaultPageSize = 10

// Cache holds prefetched data for the TUI views.
// It is safe for concurrent access.
type Cache struct {
	mu sync.RWMutex

	// Browse apps data (first page)
	browseApps       []api.App
	browseNextCursor *string // cursor for next page, nil if no more pages
	browseAppsErr    error
	browseLoaded     bool
}

// global cache instance
var globalCache = &Cache{}

// GetCache returns the global prefetch cache instance.
func GetCache() *Cache {
	return globalCache
}

// StartBrowseAppsPrefetch begins fetching the first page of browse apps in the background.
// This should be called early in the TUI lifecycle (e.g., during Init).
func (c *Cache) StartBrowseAppsPrefetch() {
	go c.fetchBrowseApps()
}

// fetchBrowseApps fetches the first page of browse apps from the API.
func (c *Cache) fetchBrowseApps() {
	cfg, err := config.Load()
	if err != nil {
		c.mu.Lock()
		c.browseAppsErr = err
		c.browseLoaded = true
		c.mu.Unlock()
		return
	}

	client := api.NewClient(cfg.APIUrl)
	result, err := client.ListAppsPaginated(DefaultPageSize, "")

	c.mu.Lock()
	if err != nil {
		c.browseAppsErr = err
	} else {
		c.browseApps = result.Apps
		c.browseNextCursor = result.NextCursor
	}
	c.browseLoaded = true
	c.mu.Unlock()
}

// BrowseAppsResult contains the result of the browse apps prefetch.
type BrowseAppsResult struct {
	Apps       []api.App
	NextCursor *string // cursor for next page, nil if no more pages
	Err        error
	Loaded     bool
}

// GetBrowseApps returns the prefetched browse apps if available.
// If the data hasn't been fetched yet, Loaded will be false.
func (c *Cache) GetBrowseApps() BrowseAppsResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return BrowseAppsResult{
		Apps:       c.browseApps,
		NextCursor: c.browseNextCursor,
		Err:        c.browseAppsErr,
		Loaded:     c.browseLoaded,
	}
}

// WaitForBrowseApps blocks until the browse apps are loaded and returns the result.
// This is useful when the view needs to wait for the data.
func (c *Cache) WaitForBrowseApps() BrowseAppsResult {
	// Simple polling wait - in practice the fetch is fast enough
	// that this won't be noticeable
	for {
		result := c.GetBrowseApps()
		if result.Loaded {
			return result
		}
		// Small sleep to avoid busy waiting
		time.Sleep(10 * time.Millisecond)
	}
}

// Reset clears all cached data. Useful for testing or when data needs to be refreshed.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.browseApps = nil
	c.browseNextCursor = nil
	c.browseAppsErr = nil
	c.browseLoaded = false
}
