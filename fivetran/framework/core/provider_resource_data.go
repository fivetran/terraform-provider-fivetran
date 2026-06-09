package core

import (
	"sync"

	fivetran "github.com/fivetran/go-fivetran"
)

// ProviderResourceData is passed as ResourceData to all resources.
// It carries the Fivetran client and the per-provider-instance metadata cache.
type ProviderResourceData struct {
	Client        *fivetran.Client
	MetadataCache *sync.Map
}
