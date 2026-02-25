package core

import (
	"context"
	"sync"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/metadata"
)

var connectorMetadataCache sync.Map // key: service string, value: *metadata.ConnectorMetadata

func GetCachedConnectorMetadata(ctx context.Context, client *fivetran.Client, service string) (*metadata.ConnectorMetadata, error) {
	if v, ok := connectorMetadataCache.Load(service); ok {
		return v.(*metadata.ConnectorMetadata), nil
	}
	resp, err := client.NewMetadataDetails().Service(service).Do(ctx)
	if err != nil {
		return nil, err
	}
	meta := resp.Data.ConnectorMetadata
	connectorMetadataCache.Store(service, &meta)
	return &meta, nil
}
