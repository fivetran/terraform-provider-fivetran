package resources

import (
	"context"
	"fmt"
	"strings"

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

// importedSchemaNameValue returns the part of the imported top-level schema (stored as name) that
// a schema name config field holds. The API returns these fields as the top-level schema instead
// of in config: schema_prefix is the whole name; for single-table connectors the name is
// "schema.table" (or "schema.table_group_name"), so schema is the part before the first dot and
// table or table_group_name the part after it. ok is false for any other field, and when the name
// can't hold the field's value.
func importedSchemaNameValue(field string, name types.String) (value string, isSchemaNameField, ok bool) {
	switch field {
	case "schema_prefix", "schema", "table", "table_group_name":
	default:
		return "", false, false
	}
	if name.IsNull() || name.IsUnknown() {
		return "", true, false
	}
	full := name.ValueString()
	schema, table, hasTable := strings.Cut(full, ".")
	switch field {
	case "schema_prefix":
		return full, true, true
	case "schema":
		return schema, true, true
	default:
		return table, true, hasTable
	}
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

// fillImportedImmutableFields returns a copy of stateValues with the immutable fields that are set
// in planValues but missing from the imported state, so the replacement check doesn't treat them as
// changed. It walks nested objects and indexed lists the same way appendImmutableRequiresReplace
// does, so an immutable field under a mutable parent (for example auth.client_access.client_id,
// since the API never returns auth) is filled too.
//
// A top-level schema name field of config (see importedSchemaNameValue) is filled only when it
// matches the imported name; a different value is a real change and still forces a replacement.
// Any other missing field can't be checked, so it is filled with a warning. isConfig is false for
// auth, which has no schema name fields. stateValues is never modified.
func fillImportedImmutableFields(planValues, stateValues map[string]interface{}, slot *metadata.Property, root path.Path, isConfig bool, name types.String, service string, diags *diag.Diagnostics) map[string]interface{} {
	filled, _ := fillMissingImmutable(planValues, stateValues, slot, root, isConfig, name, service, "", diags)
	return filled
}

// fillMissingImmutable fills one object level. It returns stateValues itself and false when
// nothing was filled. topLevelConfig is true only for the top level of config; basePath is the
// dot-separated path of planValues without list indexes, for core.IsSetTypedField.
func fillMissingImmutable(planValues, stateValues map[string]interface{}, slot *metadata.Property, root path.Path, topLevelConfig bool, name types.String, service, basePath string, diags *diag.Diagnostics) (map[string]interface{}, bool) {
	if slot == nil {
		return stateValues, false
	}

	var filled map[string]interface{}
	set := func(key string, value interface{}) {
		if filled == nil {
			filled = make(map[string]interface{}, len(stateValues)+1)
			for k, v := range stateValues {
				filled[k] = v
			}
		}
		filled[key] = value
	}

	for key, planValue := range planValues {
		prop := core.SlotProp(slot, key)
		if prop == nil || core.IsDynamicUnknownValue(planValue) {
			continue
		}
		fieldPath := root.AtName(key)
		valuePath := key
		if basePath != "" {
			valuePath = basePath + "." + key
		}
		stateValue, inState := stateValues[key]

		if prop.Immutable {
			if inState {
				continue
			}
			imported, isSchemaNameField, ok := importedSchemaNameValue(key, name)
			switch {
			case topLevelConfig && isSchemaNameField && (name.IsNull() || name.IsUnknown()):
				// Nothing to compare with: keep the replacement.
				continue
			case topLevelConfig && isSchemaNameField && ok:
				if planString, isString := planValue.(string); !isString || planString != imported {
					continue
				}
			default:
				addMissingImmutableWarning(fieldPath, basePath != "", diags)
			}
			set(key, planValue)
			continue
		}

		switch planNested := planValue.(type) {
		case map[string]interface{}:
			stateNested, _ := stateValue.(map[string]interface{})
			if nested, changed := fillMissingImmutable(planNested, stateNested, prop, fieldPath, false, name, service, valuePath, diags); changed {
				set(key, nested)
			}
		case []interface{}:
			// Like appendNestedImmutableRequiresReplace: set-typed lists aren't compared by position.
			if prop.Items == nil || core.IsSetTypedField(service, valuePath) {
				continue
			}
			stateItems, _ := stateValue.([]interface{})
			var items []interface{}
			for i, item := range planNested {
				planItem, isObject := item.(map[string]interface{})
				if !isObject {
					continue
				}
				var stateItem map[string]interface{}
				if i < len(stateItems) {
					stateItem, _ = stateItems[i].(map[string]interface{})
				}
				nested, changed := fillMissingImmutable(planItem, stateItem, prop.Items, fieldPath.AtListIndex(i), false, name, service, valuePath, diags)
				if !changed {
					continue
				}
				if items == nil {
					items = make([]interface{}, max(len(stateItems), len(planNested)))
					copy(items, stateItems)
				}
				items[i] = nested
			}
			if items != nil {
				set(key, items)
			}
		}
	}

	if filled == nil {
		return stateValues, false
	}
	return filled, true
}

func addMissingImmutableWarning(fieldPath path.Path, nested bool, diags *diag.Diagnostics) {
	sent := "the configured value is not sent to the API and does not recreate the resource."
	if nested {
		// The patch builder skips immutable fields only at the top level, so a nested one is sent
		// with its parent object.
		sent = "it does not recreate the resource, but it is sent to the API with its parent object, and the API may reject the update if the value differs."
	}
	diags.AddAttributeWarning(
		fieldPath,
		"Immutable Field Not Returned on Import",
		fmt.Sprintf("The API did not return %s when the resource was imported, so Terraform can't verify that the configured value matches. The field is immutable: %s Terraform stores it in state on the next apply and compares it as usual from then on.", fieldPath.String(), sent),
	)
}
