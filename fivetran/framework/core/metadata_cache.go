package core

import (
	"context"
	"fmt"
	"sync"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/metadata"
)

// GetCachedConnectorMetadata returns metadata for the given service, fetching once per provider instance.
// The cache is passed explicitly so each provider instance is hermetic (no package-level state).
// Errors are not cached — transient failures are retried on the next call.
func GetCachedConnectorMetadata(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string) (*metadata.ConnectorMetadata, error) {
	return getCachedConnectorMetadata(ctx, client, cache, service, "")
}

func GetCachedConnectorMetadataWithUserAgentSuffix(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string, userAgentSuffix string) (*metadata.ConnectorMetadata, error) {
	return getCachedConnectorMetadata(ctx, client, cache, service, userAgentSuffix)
}

func getCachedConnectorMetadata(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string, userAgentSuffix string) (*metadata.ConnectorMetadata, error) {
	if cache != nil {
		if meta, ok, err := LoadCachedConnectorMetadata(cache, service); ok || err != nil {
			return meta, err
		}
	}
	if client == nil {
		return nil, fmt.Errorf("unconfigured Fivetran client")
	}

	detailsService := client.NewMetadataDetails().Service(service)
	var (
		resp metadata.ConnectorMetadataResponse
		err  error
	)
	if userAgentSuffix != "" {
		resp, err = detailsService.DoWithUserAgentSuffix(ctx, userAgentSuffix)
	} else {
		resp, err = detailsService.Do(ctx)
	}
	if err != nil {
		return nil, err
	}

	meta := resp.Data.ConnectorMetadata
	if cache == nil {
		return &meta, nil
	}

	// LoadOrStore: concurrent fetches are idempotent; first stored value wins.
	actual, _ := cache.LoadOrStore(service, &meta)
	return connectorMetadataCacheValue(service, actual)
}

func LoadCachedConnectorMetadata(cache *sync.Map, service string) (*metadata.ConnectorMetadata, bool, error) {
	if cache == nil {
		return nil, false, nil
	}
	v, ok := cache.Load(service)
	if !ok {
		return nil, false, nil
	}
	meta, err := connectorMetadataCacheValue(service, v)
	return meta, true, err
}

func connectorMetadataCacheValue(service string, value interface{}) (*metadata.ConnectorMetadata, error) {
	meta, ok := value.(*metadata.ConnectorMetadata)
	if !ok {
		return nil, fmt.Errorf("metadata cache contained unexpected type %T for service %q", value, service)
	}
	return meta, nil
}
