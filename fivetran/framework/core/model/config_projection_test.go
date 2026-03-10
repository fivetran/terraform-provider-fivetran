package model

import (
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
)

// makeMeta builds a minimal ConnectorMetadata with the given properties in Config.
func makeMeta(props map[string]*metadata.Property) *metadata.ConnectorMetadata {
	return &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Properties: props,
		},
	}
}

// --- project() tests ---

func TestProject_NormalCase(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket":  {Nullable: false},
		"pattern": {Nullable: true},
	})
	remote := map[string]any{"bucket": "my-bucket", "pattern": "*.csv", "extra": "ignored"}
	mask := map[string]any{"bucket": "my-bucket", "pattern": "*.csv"}

	result := project(remote, mask, meta)

	if result["bucket"] != "my-bucket" {
		t.Errorf("bucket: got %v, want my-bucket", result["bucket"])
	}
	if result["pattern"] != "*.csv" {
		t.Errorf("pattern: got %v, want *.csv", result["pattern"])
	}
	if _, ok := result["extra"]; ok {
		t.Error("extra field should not be projected (not in mask)")
	}
}

func TestProject_PasswordFieldKeptFromLocal(t *testing.T) {
	// Sensitive fields (format=password) must keep local value, never read from remote masked value.
	meta := makeMeta(map[string]*metadata.Property{
		"secret_key": {Format: "password"},
	})
	remote := map[string]any{"secret_key": "******"}
	mask := map[string]any{"secret_key": "my-real-secret"}

	result := project(remote, mask, meta)

	if result["secret_key"] != "my-real-secret" {
		t.Errorf("secret_key: got %v, want my-real-secret (local value must be preserved)", result["secret_key"])
	}
}

func TestProject_ReadonlyFieldSkipped(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"computed_id": {Readonly: true},
		"bucket":      {},
	})
	remote := map[string]any{"computed_id": "server-generated", "bucket": "b"}
	mask := map[string]any{"computed_id": "whatever", "bucket": "b"}

	result := project(remote, mask, meta)

	if _, ok := result["computed_id"]; ok {
		t.Error("readonly field computed_id should be excluded from projected result")
	}
	if result["bucket"] != "b" {
		t.Errorf("bucket: got %v, want b", result["bucket"])
	}
}

// Metadata change scenario: a new field is added to the connector service
// but the user hasn't added it to their HCL yet.
// Expected: new field is ignored (not in mask), no drift.
func TestProject_NewFieldAddedToService_NotInMask(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket":    {},
		"new_field": {}, // newly added to service
	})
	remote := map[string]any{"bucket": "b", "new_field": "some-value"}
	mask := map[string]any{"bucket": "b"} // user hasn't added new_field to HCL

	result := project(remote, mask, meta)

	if _, ok := result["new_field"]; ok {
		t.Error("new_field not in mask should not appear in projected result — no drift expected")
	}
	if result["bucket"] != "b" {
		t.Errorf("bucket: got %v, want b", result["bucket"])
	}
}

// Metadata change scenario: a field is removed from the service (no longer returned in API response).
// User still has it in HCL/state.
// Expected: result[key] = nil, which surfaces drift so user knows the field is gone.
func TestProject_FieldRemovedFromService_StillInMask(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {},
		// "deprecated_field" removed from service, no longer in metadata or remote response
	})
	remote := map[string]any{"bucket": "b"} // deprecated_field no longer returned
	mask := map[string]any{"bucket": "b", "deprecated_field": "old-value"}

	result := project(remote, mask, meta)

	if v, ok := result["deprecated_field"]; !ok || v != nil {
		t.Errorf("deprecated_field absent in remote: want nil in result to surface drift, got %v (ok=%v)", v, ok)
	}
}

// Metadata change scenario: no metadata available (e.g. metadata fetch failed).
// Expected: project returns values from remote for all mask keys present in remote,
// nil for absent keys. No crash.
func TestProject_NilMetadata(t *testing.T) {
	remote := map[string]any{"bucket": "b"}
	mask := map[string]any{"bucket": "b", "pattern": "*.csv"}

	result := project(remote, mask, nil)

	if result["bucket"] != "b" {
		t.Errorf("bucket: got %v, want b", result["bucket"])
	}
	if v, ok := result["pattern"]; !ok || v != nil {
		t.Errorf("pattern absent in remote: want nil, got %v (ok=%v)", v, ok)
	}
}

// --- PrepareConfigPatchDynamic() tests ---

func TestPrepareConfigPatchDynamic_ChangedFieldIncluded(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {},
	})
	state := map[string]any{"bucket": "old-bucket"}
	plan := map[string]any{"bucket": "new-bucket"}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if patch["bucket"] != "new-bucket" {
		t.Errorf("changed field: got %v, want new-bucket", patch["bucket"])
	}
}

func TestPrepareConfigPatchDynamic_UnchangedFieldOmitted(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {},
	})
	state := map[string]any{"bucket": "same"}
	plan := map[string]any{"bucket": "same"}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if _, ok := patch["bucket"]; ok {
		t.Error("unchanged field should be omitted from patch")
	}
}

// Empty string is a valid value — must NOT be treated as a clear operation.
func TestPrepareConfigPatchDynamic_EmptyStringIsValidValue(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"pattern": {Nullable: true},
	})
	state := map[string]any{"pattern": "*.csv"}
	plan := map[string]any{"pattern": ""}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	v, ok := patch["pattern"]
	if !ok {
		t.Fatal("pattern should be in patch (value changed)")
	}
	if v != "" {
		t.Errorf("pattern: got %v, want empty string (not nil)", v)
	}
}

// Field removed from HCL, nullable → send null to clear it.
func TestPrepareConfigPatchDynamic_RemovedNullableFieldSendsNull(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"pattern": {Nullable: true},
	})
	state := map[string]any{"pattern": "*.csv"}
	plan := map[string]any{} // pattern removed from HCL

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	v, ok := patch["pattern"]
	if !ok {
		t.Fatal("nullable field removed from plan should be present in patch")
	}
	if v != nil {
		t.Errorf("nullable field removed from plan: got %v, want nil", v)
	}
}

// Field removed from HCL, non-nullable → omit from patch entirely.
func TestPrepareConfigPatchDynamic_RemovedNonNullableFieldOmitted(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {Nullable: false},
	})
	state := map[string]any{"bucket": "my-bucket"}
	plan := map[string]any{} // bucket removed from HCL

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if _, ok := patch["bucket"]; ok {
		t.Error("non-nullable field removed from plan should be omitted from patch, not sent as null")
	}
}

// Field removed from HCL, readonly → must not be sent even if nullable.
func TestPrepareConfigPatchDynamic_RemovedReadonlyFieldOmitted(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"computed_id": {Nullable: true, Readonly: true},
	})
	state := map[string]any{"computed_id": "abc"}
	plan := map[string]any{}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if _, ok := patch["computed_id"]; ok {
		t.Error("readonly field should never be sent in patch")
	}
}

// Metadata change scenario: a new field is added to the service and shows up in plan
// but was not in state (first time user adds it to HCL).
// Expected: new field included in patch.
func TestPrepareConfigPatchDynamic_NewFieldAddedToHCL(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket":    {},
		"new_field": {},
	})
	state := map[string]any{"bucket": "b"}
	plan := map[string]any{"bucket": "b", "new_field": "value"}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if patch["new_field"] != "value" {
		t.Errorf("new_field added to HCL: got %v, want value", patch["new_field"])
	}
	if _, ok := patch["bucket"]; ok {
		t.Error("unchanged bucket should be omitted from patch")
	}
}

// Metadata change scenario: field removed from metadata (no longer in metadata Properties),
// but it's still in state and plan (user hasn't removed it from HCL yet).
// Expected: field is still sent in patch since it's in plan and changed.
// No crash from missing metadata entry.
func TestPrepareConfigPatchDynamic_FieldRemovedFromMetadata_StillInPlan(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {}, // deprecated_field intentionally absent from metadata
	})
	state := map[string]any{"bucket": "b", "deprecated_field": "old"}
	plan := map[string]any{"bucket": "b", "deprecated_field": "new"}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if patch["deprecated_field"] != "new" {
		t.Errorf("field changed in plan should still be patched even if absent from metadata: got %v", patch["deprecated_field"])
	}
}

// Metadata change scenario: field removed from metadata AND removed from HCL (user cleaned up).
// Expected: field omitted from patch (no metadata → treated as non-nullable → skip).
func TestPrepareConfigPatchDynamic_FieldRemovedFromMetadataAndHCL(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"bucket": {},
	})
	state := map[string]any{"bucket": "b", "deprecated_field": "old"}
	plan := map[string]any{"bucket": "b"} // deprecated_field removed from HCL

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if _, ok := patch["deprecated_field"]; ok {
		t.Error("field absent from metadata and removed from HCL should be omitted from patch")
	}
}

// Nil metadata + field removed from plan: no metadata means we can't know nullability,
// so the safe default is to omit the field (don't send null blindly).
func TestPrepareConfigPatchDynamic_NilMetadata_RemovedFieldOmitted(t *testing.T) {
	state := map[string]any{"bucket": "b", "pattern": "*.csv"}
	plan := map[string]any{"bucket": "b"} // pattern removed

	patch := PrepareConfigPatchDynamic(state, plan, nil)

	if _, ok := patch["pattern"]; ok {
		t.Error("with nil metadata removed field should be omitted — cannot determine nullability")
	}
}

// Both state and plan empty → patch should be empty, no crash.
func TestPrepareConfigPatchDynamic_BothEmpty(t *testing.T) {
	patch := PrepareConfigPatchDynamic(map[string]any{}, map[string]any{}, nil)
	if len(patch) != 0 {
		t.Errorf("expected empty patch, got %v", patch)
	}
}

// Readonly field changed in plan — should still be sent (project skips readonly on read,
// but patch only skips readonly for *removed* nullable fields, not changed ones).
func TestPrepareConfigPatchDynamic_ReadonlyChangedInPlan_Included(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"computed_id": {Readonly: true},
	})
	state := map[string]any{"computed_id": "old"}
	plan := map[string]any{"computed_id": "new"}

	patch := PrepareConfigPatchDynamic(state, plan, meta)

	if patch["computed_id"] != "new" {
		t.Errorf("readonly field that changed in plan should still be included: got %v", patch["computed_id"])
	}
}

// Nested object: project should recurse and apply mask/password logic at each level.
func TestProject_NestedObject(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{
		"connection": {},
	})
	remote := map[string]any{
		"connection": map[string]any{"host": "db.example.com", "port": "5432", "extra": "ignored"},
	}
	mask := map[string]any{
		"connection": map[string]any{"host": "db.example.com", "port": "5432"},
	}

	result := project(remote, mask, meta)

	nested, ok := result["connection"].(map[string]any)
	if !ok {
		t.Fatalf("connection should be a nested map, got %T", result["connection"])
	}
	if nested["host"] != "db.example.com" {
		t.Errorf("nested host: got %v", nested["host"])
	}
	if nested["port"] != "5432" {
		t.Errorf("nested port: got %v", nested["port"])
	}
	if _, ok := nested["extra"]; ok {
		t.Error("extra key not in nested mask should be excluded")
	}
}

// Nested object: remote returns scalar where mask expects object — should fall through
// and set the remote scalar value rather than crashing.
func TestProject_NestedObject_TypeMismatch(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{"connection": {}})
	remote := map[string]any{"connection": "unexpected-scalar"}
	mask := map[string]any{"connection": map[string]any{"host": "h"}}

	result := project(remote, mask, meta)

	if result["connection"] != "unexpected-scalar" {
		t.Errorf("type mismatch: expected remote scalar to be passed through, got %v", result["connection"])
	}
}

// Nested object: remote key is absent entirely — outer key should become nil (drift surfaced).
func TestProject_NestedObject_AbsentInRemote(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{"connection": {}})
	remote := map[string]any{}
	mask := map[string]any{"connection": map[string]any{"host": "h"}}

	result := project(remote, mask, meta)

	if v, ok := result["connection"]; !ok || v != nil {
		t.Errorf("nested key absent in remote should be nil to surface drift, got %v (ok=%v)", v, ok)
	}
}

// Remote returns nil for an existing key (e.g. field explicitly nulled server-side).
// Should be passed through as nil.
func TestProject_RemoteValueIsNil(t *testing.T) {
	meta := makeMeta(map[string]*metadata.Property{"pattern": {Nullable: true}})
	remote := map[string]any{"pattern": nil}
	mask := map[string]any{"pattern": "*.csv"}

	result := project(remote, mask, meta)

	v, ok := result["pattern"]
	if !ok {
		t.Fatal("pattern key should be present in result")
	}
	if v != nil {
		t.Errorf("remote nil should be projected as nil, got %v", v)
	}
}
