package core

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SchemaLockManager provides per-connection_id mutexes to serialize schema
// modification requests. The Fivetran API uses optimistic locking on the
// connection schema resource, so concurrent PATCH requests for the same
// connection_id produce 409 Conflict errors.
type SchemaLockManager struct {
	locks sync.Map // map[string]*sync.Mutex
}

// SchemaLocks is the singleton lock manager shared across all schema-related
// resources and actions in the provider.
var SchemaLocks = &SchemaLockManager{}

// Lock acquires the mutex for the given connection ID. If no mutex exists yet,
// one is created atomically.
func (m *SchemaLockManager) Lock(connectionId string) {
	val, _ := m.locks.LoadOrStore(connectionId, &sync.Mutex{})
	val.(*sync.Mutex).Lock()
}

// Unlock releases the mutex for the given connection ID.
func (m *SchemaLockManager) Unlock(connectionId string) {
	val, ok := m.locks.Load(connectionId)
	if ok {
		val.(*sync.Mutex).Unlock()
	}
}

const schemaConflictMaxRetries = 5

// schemaConflictBackoff is the base delay between retries (doubles each attempt).
var schemaConflictBackoff = 1 * time.Second

// SchemaConflictBackoff returns the current backoff duration (for test save/restore).
func SchemaConflictBackoff() time.Duration { return schemaConflictBackoff }

// SetSchemaConflictBackoff overrides the backoff duration (for fast tests).
func SetSchemaConflictBackoff(d time.Duration) { schemaConflictBackoff = d }

// IsConflictError returns true if the error indicates an HTTP 409 Conflict.
// The go-fivetran SDK returns errors like "status code: 409; expected: 200".
func IsConflictError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "status code: 409")
}

// RetryOnSchemaConflict executes the given operation, retrying with exponential
// backoff when the API returns a 409 Conflict (optimistic lock failure).
// The operation should perform the full read-modify-write cycle so that each
// retry re-reads the latest state. On non-conflict errors the error is added
// to diagnostics and the function returns immediately.
func RetryOnSchemaConflict(ctx context.Context, diagnostics *diag.Diagnostics, summary string, operation func() error) {
	for attempt := 0; attempt <= schemaConflictMaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return
		}
		if !IsConflictError(err) || attempt == schemaConflictMaxRetries {
			diagnostics.AddError(summary, err.Error())
			return
		}

		delay := schemaConflictBackoff * (1 << attempt)
		tflog.Warn(ctx, "Schema update conflict (409), retrying",
			map[string]any{
				"attempt": attempt + 1,
				"delay":   delay.String(),
			})

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			diagnostics.AddError(summary, "Operation cancelled during conflict retry.")
			return
		}
	}
}
