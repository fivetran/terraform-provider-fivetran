package core

import (
	"sync"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

func TestLogServiceMetadataCacheDoesNotCollideWithAccountExternalLoggingKeys(t *testing.T) {
	cache := &sync.Map{}
	cache.Store("log-service:datadog_log", &metadata.ConnectorMetadata{})
	cache.Store("no-metadata:log-service:new_relic_log", true)

	for _, service := range []string{"datadog_log", "new_relic_log"} {
		meta, ok, err := LoadCachedLogServiceMetadata(cache, service)
		if err != nil || ok || meta != nil {
			t.Errorf("%s: LoadCachedLogServiceMetadata = (%v, %v, %v), want a cache miss", service, meta, ok, err)
		}
	}
}
