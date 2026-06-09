package framework

import (
	"context"
	"sync"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/metadata"
)

// GetCachedConnectorMetadata returns metadata for the given service, fetching once per provider instance.
// The cache is passed explicitly so each provider instance is hermetic (no package-level state).
// Errors are not cached — transient failures are retried on the next call.
func GetCachedConnectorMetadata(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string) (*metadata.ConnectorMetadata, error) {
	if v, ok := cache.Load(service); ok {
		return v.(*metadata.ConnectorMetadata), nil
	}

	resp, err := client.NewMetadataDetails().Service(service).Do(ctx)
	if err != nil {
		return nil, err
	}

	meta := resp.Data.ConnectorMetadata
	// LoadOrStore: concurrent fetches are idempotent; first stored value wins.
	actual, _ := cache.LoadOrStore(service, &meta)
	return actual.(*metadata.ConnectorMetadata), nil
}
