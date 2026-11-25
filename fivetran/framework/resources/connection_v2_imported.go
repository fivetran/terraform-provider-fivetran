package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectionV2ImportedKey is the private state key set by import and removed by the first
// successful update. Import builds config only from the fields the API returns, so an immutable
// field the API leaves out (for example schema_prefix, returned as the top-level schema) is
// missing from state, and the first plan would replace the connection. While the key is set,
// such a field is not compared with state (see fillImportedImmutableFields). The update writes
// the configured value to state, so later plans compare it as usual.
const connectionV2ImportedKey = "imported"

// connectionV2SchemaNameFields are the config fields the API returns as the connection's
// top-level schema (stored as name) instead of in config.
var connectionV2SchemaNameFields = map[string]bool{
	"schema":        true,
	"schema_prefix": true,
}

func setConnectionV2Imported(ctx context.Context, private privateStateSetter, imported bool) diag.Diagnostics {
	if imported {
		return private.SetKey(ctx, connectionV2ImportedKey, []byte("true"))
	}
	return private.SetKey(ctx, connectionV2ImportedKey, nil)
}

func isConnectionV2Imported(ctx context.Context, private privateStateGetter) (bool, diag.Diagnostics) {
	if private == nil {
		return false, nil
	}
	value, diags := private.GetKey(ctx, connectionV2ImportedKey)
	return len(value) > 0, diags
}

// fillImportedImmutableFields returns a copy of stateValues with the top-level immutable fields
// that are set in planValues but missing from the imported state, so the replacement check
// doesn't treat them as changed. A schema name field (see connectionV2SchemaNameFields) is filled
// only when it matches the imported name; a different value is a real change and still forces a
// replacement. Any other missing field can't be checked, so it is filled with a warning.
// isConfig is false for auth, which has no schema name fields.
func fillImportedImmutableFields(planValues, stateValues map[string]interface{}, slot *metadata.Property, root path.Path, isConfig bool, name types.String, diags *diag.Diagnostics) map[string]interface{} {
	if slot == nil {
		return stateValues
	}

	var filled map[string]interface{}
	for key, planValue := range planValues {
		prop := core.SlotProp(slot, key)
		if prop == nil || !prop.Immutable || core.IsDynamicUnknownValue(planValue) {
			continue
		}
		if _, inState := stateValues[key]; inState {
			continue
		}

		if isConfig && connectionV2SchemaNameFields[key] {
			planString, ok := planValue.(string)
			if !ok || name.IsNull() || name.IsUnknown() || planString != name.ValueString() {
				continue
			}
		} else {
			addMissingImmutableWarning(root.AtName(key), diags)
		}

		if filled == nil {
			filled = make(map[string]interface{}, len(stateValues)+1)
			for k, v := range stateValues {
				filled[k] = v
			}
		}
		filled[key] = planValue
	}

	if filled == nil {
		return stateValues
	}
	return filled
}

func addMissingImmutableWarning(fieldPath path.Path, diags *diag.Diagnostics) {
	diags.AddAttributeWarning(
		fieldPath,
		"Immutable Field Not Returned on Import",
		fmt.Sprintf("The API did not return %s when the resource was imported, so Terraform can't verify that the configured value matches. The field is immutable: the configured value is not sent to the API and does not recreate the resource. Terraform stores it in state on the next apply and compares it as usual from then on.", fieldPath.String()),
	)
}
