package framework

import (
	"context"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// makeSlot builds a metadata.Property slot with the given child properties.
func makeSlot(props map[string]*metadata.Property) *metadata.Property {
	return &metadata.Property{Properties: props}
}

// --- DynamicToMap ---

func TestDynamicToMap_Null(t *testing.T) {
	t.Parallel()
	m, diags := DynamicToMap(context.Background(), types.DynamicNull())
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if m != nil {
		t.Errorf("expected nil for null dynamic, got %v", m)
	}
}

func TestDynamicToMap_SimpleObject(t *testing.T) {
	t.Parallel()
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"bucket": types.StringType, "enabled": types.BoolType},
		map[string]attr.Value{"bucket": types.StringValue("my-bucket"), "enabled": types.BoolValue(true)},
	)
	m, diags := DynamicToMap(context.Background(), types.DynamicValue(obj))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if m["bucket"] != "my-bucket" {
		t.Errorf("bucket: got %v, want my-bucket", m["bucket"])
	}
	if m["enabled"] != true {
		t.Errorf("enabled: got %v, want true", m["enabled"])
	}
}

func TestDynamicToMap_NestedObject(t *testing.T) {
	t.Parallel()
	inner, _ := types.ObjectValue(
		map[string]attr.Type{"key": types.StringType},
		map[string]attr.Value{"key": types.StringValue("val")},
	)
	outer, _ := types.ObjectValue(
		map[string]attr.Type{"nested": types.ObjectType{AttrTypes: map[string]attr.Type{"key": types.StringType}}},
		map[string]attr.Value{"nested": inner},
	)
	m, diags := DynamicToMap(context.Background(), types.DynamicValue(outer))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	nested, ok := m["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested should be map[string]interface{}, got %T", m["nested"])
	}
	if nested["key"] != "val" {
		t.Errorf("nested.key: got %v, want val", nested["key"])
	}
}

func TestDynamicToMap_EmptyStringPreserved(t *testing.T) {
	t.Parallel()
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"pattern": types.StringType},
		map[string]attr.Value{"pattern": types.StringValue("")},
	)
	m, diags := DynamicToMap(context.Background(), types.DynamicValue(obj))
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	v, ok := m["pattern"]
	if !ok || v != "" {
		t.Errorf("empty string should be preserved, got %v (ok=%v)", v, ok)
	}
}

// --- project ---

func TestProject_NormalFieldTakesRemote(t *testing.T) {
	t.Parallel()
	t.Run("normal_field_takes_remote_value", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"bucket": {}})
		remote := map[string]interface{}{"bucket": "remote-bucket", "extra": "ignored"}
		mask := map[string]interface{}{"bucket": "local-bucket"}

		result := project(remote, mask, slot)

		if result["bucket"] != "remote-bucket" {
			t.Errorf("got %v, want remote-bucket", result["bucket"])
		}
		if _, ok := result["extra"]; ok {
			t.Error("extra not in mask should not appear in result")
		}
	})
}

func TestProject_SensitiveKeepsLocal(t *testing.T) {
	t.Parallel()
	t.Run("sensitive_keeps_local_value", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"secret_key": {Format: "password"}})
		remote := map[string]interface{}{"secret_key": "****"}
		mask := map[string]interface{}{"secret_key": "my-real-secret"}

		result := project(remote, mask, slot)

		if result["secret_key"] != "my-real-secret" {
			t.Errorf("sensitive: got %v, want my-real-secret (local must be preserved)", result["secret_key"])
		}
	})
}

func TestProject_ReadonlyTakesRemote(t *testing.T) {
	t.Parallel()
	t.Run("readonly_takes_remote_value", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{
			"public_key": {Readonly: true},
			"bucket":     {},
		})
		remote := map[string]interface{}{"public_key": "server-key", "bucket": "b"}
		mask := map[string]interface{}{"public_key": "local-ignored", "bucket": "b"}

		result := project(remote, mask, slot)

		if result["public_key"] != "server-key" {
			t.Errorf("readonly: got %v, want server-key (remote stored for reference)", result["public_key"])
		}
	})
}

func TestProject_MissingFromAPI_SetNil(t *testing.T) {
	t.Parallel()
	t.Run("missing_from_api_set_to_nil", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"bucket": {}, "gone": {}})
		remote := map[string]interface{}{"bucket": "b"}
		mask := map[string]interface{}{"bucket": "b", "gone": "old-val"}

		result := project(remote, mask, slot)

		v, ok := result["gone"]
		if !ok || v != nil {
			t.Errorf("field absent from API: want nil to surface drift, got %v (ok=%v)", v, ok)
		}
	})
}

func TestProject_NestedObject(t *testing.T) {
	t.Parallel()
	slot := makeSlot(map[string]*metadata.Property{
		"opts": {Properties: map[string]*metadata.Property{"key": {}}},
	})
	remote := map[string]interface{}{"opts": map[string]interface{}{"key": "remote-val", "extra": "nope"}}
	mask := map[string]interface{}{"opts": map[string]interface{}{"key": "local-val"}}

	result := project(remote, mask, slot)

	nested, ok := result["opts"].(map[string]interface{})
	if !ok {
		t.Fatalf("nested should be map, got %T", result["opts"])
	}
	if nested["key"] != "remote-val" {
		t.Errorf("nested.key: got %v, want remote-val", nested["key"])
	}
	if _, ok := nested["extra"]; ok {
		t.Error("extra not in mask should be excluded from nested result")
	}
}

func TestProject_NilSlot_Passthrough(t *testing.T) {
	t.Parallel()
	remote := map[string]interface{}{"bucket": "b", "pattern": "*.csv"}
	mask := map[string]interface{}{"bucket": "b", "pattern": "*.csv"}

	result := project(remote, mask, nil)

	if result["bucket"] != "b" || result["pattern"] != "*.csv" {
		t.Errorf("nil slot: expected passthrough, got %v", result)
	}
}

// --- PrepareConfigPatchDynamic ---

func TestPatch_ChangedFieldIncluded(t *testing.T) {
	t.Parallel()
	slot := makeSlot(map[string]*metadata.Property{"bucket": {}})
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"bucket": "new"},
		map[string]interface{}{"bucket": "old"},
		slot,
	)
	if patch["bucket"] != "new" {
		t.Errorf("changed field: got %v, want new", patch["bucket"])
	}
}

func TestPatch_UnchangedFieldOmitted(t *testing.T) {
	t.Parallel()
	slot := makeSlot(map[string]*metadata.Property{"bucket": {}})
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"bucket": "same"},
		map[string]interface{}{"bucket": "same"},
		slot,
	)
	if _, ok := patch["bucket"]; ok {
		t.Error("unchanged field should be omitted from patch")
	}
}

func TestPatch_EmptyStringSentVerbatim(t *testing.T) {
	t.Parallel()
	t.Run("empty_string_not_coerced_to_null", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"pattern": {Nullable: true}})
		patch := PrepareConfigPatchDynamic(
			map[string]interface{}{"pattern": ""},
			map[string]interface{}{"pattern": "*.csv"},
			slot,
		)
		v, ok := patch["pattern"]
		if !ok {
			t.Fatal("empty string change should be in patch")
		}
		if v != "" {
			t.Errorf("empty string must not be coerced to nil, got %v", v)
		}
	})
}

func TestPatch_NullableFieldRemovedSendsNull(t *testing.T) {
	t.Parallel()
	t.Run("nullable_field_removed_sends_null", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"pattern": {Nullable: true}})
		patch := PrepareConfigPatchDynamic(
			map[string]interface{}{},
			map[string]interface{}{"pattern": "*.csv"},
			slot,
		)
		v, ok := patch["pattern"]
		if !ok {
			t.Fatal("nullable removal should be present in patch")
		}
		if v != nil {
			t.Errorf("nullable removal: want nil (JSON null), got %v", v)
		}
	})
}

func TestPatch_NonNullableFieldRemovedOmitted(t *testing.T) {
	t.Parallel()
	t.Run("non_nullable_field_removed_omitted", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"bucket": {Nullable: false}})
		patch := PrepareConfigPatchDynamic(
			map[string]interface{}{},
			map[string]interface{}{"bucket": "b"},
			slot,
		)
		if _, ok := patch["bucket"]; ok {
			t.Error("non-nullable removal should be omitted, not sent as null")
		}
	})
}

func TestPatch_ReadonlyFieldNeverSent(t *testing.T) {
	t.Parallel()
	t.Run("readonly_field_never_sent_in_patch", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"public_key": {Readonly: true}})
		patch := PrepareConfigPatchDynamic(
			map[string]interface{}{"public_key": "new"},
			map[string]interface{}{"public_key": "old"},
			slot,
		)
		if _, ok := patch["public_key"]; ok {
			t.Error("readonly field must never be sent in patch")
		}
	})
}

func TestPatch_ImmutableFieldNeverSent(t *testing.T) {
	t.Parallel()
	t.Run("immutable_field_never_sent_in_patch", func(t *testing.T) {
		slot := makeSlot(map[string]*metadata.Property{"account_name": {Immutable: true}})
		patch := PrepareConfigPatchDynamic(
			map[string]interface{}{"account_name": "new"},
			map[string]interface{}{"account_name": "old"},
			slot,
		)
		if _, ok := patch["account_name"]; ok {
			t.Error("immutable field must never be sent in patch")
		}
	})
}

func TestPatch_SensitiveFieldSentIfChanged(t *testing.T) {
	t.Parallel()
	slot := makeSlot(map[string]*metadata.Property{"secret_key": {Format: "password"}})
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"secret_key": "new-secret"},
		map[string]interface{}{"secret_key": "old-secret"},
		slot,
	)
	if patch["secret_key"] != "new-secret" {
		t.Errorf("sensitive changed: got %v, want new-secret", patch["secret_key"])
	}
}

func TestPatch_NilSlot_ChangedFieldIncluded(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"bucket": "new", "pattern": "same"},
		map[string]interface{}{"bucket": "old", "pattern": "same"},
		nil,
	)
	if patch["bucket"] != "new" {
		t.Errorf("nil slot changed field: got %v, want new", patch["bucket"])
	}
	if _, ok := patch["pattern"]; ok {
		t.Error("nil slot unchanged field should be omitted")
	}
}

func TestPatch_NilSlot_RemovedFieldOmitted(t *testing.T) {
	t.Parallel()
	patch := PrepareConfigPatchDynamic(
		map[string]interface{}{"bucket": "b"},
		map[string]interface{}{"bucket": "b", "pattern": "*.csv"},
		nil,
	)
	if _, ok := patch["pattern"]; ok {
		t.Error("nil slot removed field should be omitted — nullability unknown")
	}
}
