package core

import (
	"context"
	"fmt"
	"strings"
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

// destinationMetadataCachePrefix keeps destination metadata in its own key space inside the
// shared provider cache so destination and connector service ids can never collide.
const destinationMetadataCachePrefix = "destination:"

// GetCachedDestinationMetadataWithUserAgentSuffix returns destination metadata for the given service,
// fetching once per provider instance. Errors are not cached — transient failures are retried on the next call.
func GetCachedDestinationMetadataWithUserAgentSuffix(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string, userAgentSuffix string) (*metadata.DestinationMetadata, error) {
	return getCachedDestinationMetadata(ctx, client, cache, service, userAgentSuffix)
}

func getCachedDestinationMetadata(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string, userAgentSuffix string) (*metadata.DestinationMetadata, error) {
	if cache != nil {
		if meta, ok, err := LoadCachedDestinationMetadata(cache, service); ok || err != nil {
			return meta, err
		}
	}
	if client == nil {
		return nil, fmt.Errorf("unconfigured Fivetran client")
	}

	detailsService := client.NewDestinationMetadataDetails().Service(service)
	var (
		resp metadata.DestinationMetadataResponse
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

	meta := resp.Data.DestinationMetadata
	if cache == nil {
		return &meta, nil
	}

	// LoadOrStore: concurrent fetches are idempotent; first stored value wins.
	actual, _ := cache.LoadOrStore(DestinationMetadataCacheKey(service), &meta)
	return destinationMetadataCacheValue(service, actual)
}

// DestinationMetadataCacheKey returns the cache key under which destination metadata for service is stored.
func DestinationMetadataCacheKey(service string) string {
	return destinationMetadataCachePrefix + service
}

func LoadCachedDestinationMetadata(cache *sync.Map, service string) (*metadata.DestinationMetadata, bool, error) {
	if cache == nil {
		return nil, false, nil
	}
	v, ok := cache.Load(DestinationMetadataCacheKey(service))
	if !ok {
		return nil, false, nil
	}
	meta, err := destinationMetadataCacheValue(service, v)
	return meta, true, err
}

func destinationMetadataCacheValue(service string, value interface{}) (*metadata.DestinationMetadata, error) {
	meta, ok := value.(*metadata.DestinationMetadata)
	if !ok {
		return nil, fmt.Errorf("destination metadata cache contained unexpected type %T for service %q", value, service)
	}
	return meta, nil
}

// logServiceMetadataCachePrefix keeps log service metadata in its own key space inside the
// shared provider cache so log and connector service ids can never collide.
const logServiceMetadataCachePrefix = "log-service-metadata:"

// noLogServiceMetadataCachePrefix marks log services for which the metadata API has no description,
// so the negative answer is not requested again on every plan.
const noLogServiceMetadataCachePrefix = "no-metadata:" + logServiceMetadataCachePrefix

// GetCachedLogServiceMetadataWithUserAgentSuffix returns log service metadata for the given service,
// fetching once per provider instance. It returns nil without an error when the metadata API has no
// description for the service (for example a server that does not serve the endpoint yet); that answer
// is cached too. Other errors are not cached — transient failures are retried on the next call.
func GetCachedLogServiceMetadataWithUserAgentSuffix(ctx context.Context, client *fivetran.Client, cache *sync.Map, service string, userAgentSuffix string) (*metadata.LogServiceMetadata, error) {
	if cache != nil {
		if meta, ok, err := LoadCachedLogServiceMetadata(cache, service); ok || err != nil {
			return meta, err
		}
	}
	if client == nil {
		return nil, fmt.Errorf("unconfigured Fivetran client")
	}

	resp, err := client.NewLogServiceMetadataDetails().
		Service(service).
		DoWithUserAgentSuffix(ctx, userAgentSuffix)
	if err != nil {
		if strings.HasPrefix(resp.Code, "NotFound") {
			if cache != nil {
				cache.Store(noLogServiceMetadataCachePrefix+service, true)
			}
			return nil, nil
		}
		return nil, err
	}

	meta := resp.Data.LogServiceMetadata
	if cache == nil {
		return &meta, nil
	}

	// LoadOrStore: concurrent fetches are idempotent; first stored value wins.
	actual, _ := cache.LoadOrStore(LogServiceMetadataCacheKey(service), &meta)
	return logServiceMetadataCacheValue(service, actual)
}

// LogServiceMetadataCacheKey returns the cache key under which log service metadata for service is stored.
func LogServiceMetadataCacheKey(service string) string {
	return logServiceMetadataCachePrefix + service
}

// LoadCachedLogServiceMetadata reports a cached "no metadata" answer as found with a nil value.
func LoadCachedLogServiceMetadata(cache *sync.Map, service string) (*metadata.LogServiceMetadata, bool, error) {
	if cache == nil {
		return nil, false, nil
	}
	if _, ok := cache.Load(noLogServiceMetadataCachePrefix + service); ok {
		return nil, true, nil
	}
	v, ok := cache.Load(LogServiceMetadataCacheKey(service))
	if !ok {
		return nil, false, nil
	}
	meta, err := logServiceMetadataCacheValue(service, v)
	return meta, true, err
}

func logServiceMetadataCacheValue(service string, value interface{}) (*metadata.LogServiceMetadata, error) {
	meta, ok := value.(*metadata.LogServiceMetadata)
	if !ok {
		return nil, fmt.Errorf("log service metadata cache contained unexpected type %T for service %q", value, service)
	}
	return meta, nil
}
