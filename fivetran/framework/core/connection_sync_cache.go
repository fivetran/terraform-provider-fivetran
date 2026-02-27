package core

import (
	"context"
	"sync"
	"time"

	"github.com/fivetran/go-fivetran"
)

// ConnectionSyncCache caches whether a connection has synced data to avoid
// redundant GET /connections/{id} calls during plan. Multiple schema resources
// for the same connection_id share a single cached result.
type ConnectionSyncCache struct {
	mu      sync.Mutex
	entries map[string]syncCacheEntry
}

type syncCacheEntry struct {
	hasSynced bool
	fetchedAt time.Time
}

const syncCacheTTL = 1 * time.Minute

// ConnectionSyncStatus is the singleton cache shared across all schema resources.
var ConnectionSyncStatus = &ConnectionSyncCache{
	entries: make(map[string]syncCacheEntry),
}

// Reset clears all cached entries (for testing).
func (c *ConnectionSyncCache) Reset() {
	c.mu.Lock()
	c.entries = make(map[string]syncCacheEntry)
	c.mu.Unlock()
}

// HasSynced returns whether the connection has synced data (succeeded_at or
// failed_at is set). Results are cached per connection_id for the duration of
// syncCacheTTL. Returns false on API errors (don't block plan on a failed check).
func (c *ConnectionSyncCache) HasSynced(ctx context.Context, client *fivetran.Client, connectionId string) bool {
	c.mu.Lock()
	if entry, ok := c.entries[connectionId]; ok && time.Since(entry.fetchedAt) < syncCacheTTL {
		c.mu.Unlock()
		return entry.hasSynced
	}
	c.mu.Unlock()

	// Fetch outside the lock to avoid holding it during API call
	connResp, err := client.NewConnectionDetails().ConnectionID(connectionId).Do(ctx)

	hasSynced := false
	if err == nil {
		hasSynced = !connResp.Data.SucceededAt.IsZero() || !connResp.Data.FailedAt.IsZero()
	}

	c.mu.Lock()
	c.entries[connectionId] = syncCacheEntry{hasSynced: hasSynced, fetchedAt: time.Now()}
	c.mu.Unlock()

	return hasSynced
}
