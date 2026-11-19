package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	configSecretsPath = path.Root("config_secrets")
	authSecretsPath   = path.Root("auth_secrets")
)

// connectionV2SecretsFromConfig reads a write-only secrets attribute (config_secrets or
// auth_secrets) from the configuration. Write-only attributes are always null in the plan and
// the state. Unknown values are kept when preserveUnknown is set, for plan-time checks.
func connectionV2SecretsFromConfig(ctx context.Context, config tfsdk.Config, secretsPath path.Path, preserveUnknown bool, diags *diag.Diagnostics) map[string]interface{} {
	if config.Raw.IsNull() {
		return nil
	}

	var value types.Dynamic
	diags.Append(config.GetAttribute(ctx, secretsPath, &value)...)
	if diags.HasError() {
		return nil
	}

	return configSecretsMap(ctx, value, preserveUnknown, diags)
}

func configSecretsMap(ctx context.Context, value types.Dynamic, preserveUnknown bool, diags *diag.Diagnostics) map[string]interface{} {
	var secrets map[string]interface{}
	var secretsDiags diag.Diagnostics
	if preserveUnknown {
		secrets, secretsDiags = core.DynamicToMapPreserveUnknown(ctx, value)
	} else {
		secrets, secretsDiags = core.DynamicToMap(ctx, value)
	}
	diags.Append(secretsDiags...)
	return secrets
}

// mergeConnectionV2Secrets returns values (config or auth) with the matching secrets fields
// added. Neither map is modified.
func mergeConnectionV2Secrets(config, secrets map[string]interface{}) map[string]interface{} {
	if len(secrets) == 0 {
		return config
	}

	merged := make(map[string]interface{}, len(config)+len(secrets))
	for k, v := range config {
		merged[k] = v
	}
	for k, v := range secrets {
		merged[k] = v
	}
	return merged
}

// validateConnectionV2SecretsOverlap reports every top-level field set in both values and secrets,
// e.g. config and config_secrets, or auth and auth_secrets.
func validateConnectionV2SecretsOverlap(values, secrets map[string]interface{}, valuesPath, secretsPath path.Path, diags *diag.Diagnostics) {
	for name := range secrets {
		if _, ok := values[name]; ok {
			diags.AddAttributeError(
				secretsPath.AtName(name),
				"Conflicting Dynamic Field",
				fmt.Sprintf("Field %q is set in both %s and %s. Set it in only one of them.", name, valuesPath, secretsPath),
			)
		}
	}
}

// applyConnectionV2SecretsToPatch adds mutable secrets fields (config_secrets or auth_secrets) when
// their version changed or they moved from config or auth. This prevents removed nullable fields from being cleared on migration.
// Immutable and read-only fields are excluded at every nesting level; moving them only
// removes their stored values, without sending them to the API again.
func applyConnectionV2SecretsToPatch(patch, secrets, stateValues map[string]interface{}, slot *metadata.Property, versionChanged bool) map[string]interface{} {
	for name, value := range secrets {
		if _, moved := stateValues[name]; !versionChanged && !moved {
			continue
		}
		value, mutable := connectionV2SecretPatchValue(value, core.SlotProp(slot, name))
		if !mutable {
			// A migrated nullable object may have a clearing PATCH. Remove it when
			// all its descendants are excluded instead of clearing the object.
			delete(patch, name)
			continue
		}
		if patch == nil {
			patch = make(map[string]interface{}, len(secrets))
		}
		patch[name] = value
	}
	return patch
}

// connectionV2SecretPatchValue copies a value without immutable or read-only fields.
// The bool reports whether anything remains to send. Explicit empty values are preserved,
// and array items keep their positions even when all fields in an item are excluded.
func connectionV2SecretPatchValue(value interface{}, prop *metadata.Property) (interface{}, bool) {
	if prop == nil {
		return value, true
	}
	if prop.Immutable || prop.Readonly {
		return nil, false
	}

	switch v := value.(type) {
	case map[string]interface{}:
		filtered := make(map[string]interface{}, len(v))
		for name, child := range v {
			if child, mutable := connectionV2SecretPatchValue(child, core.SlotProp(prop, name)); mutable {
				filtered[name] = child
			}
		}
		return filtered, len(v) == 0 || len(filtered) > 0
	case []interface{}:
		if prop.Items == nil {
			return value, true
		}
		filtered := make([]interface{}, len(v))
		mutable := len(v) == 0
		for i, child := range v {
			var childMutable bool
			filtered[i], childMutable = connectionV2SecretPatchValue(child, prop.Items)
			mutable = mutable || childMutable
		}
		return filtered, mutable
	default:
		return value, true
	}
}

// movedConnectionV2Secrets returns fields whose prior values can still be compared during migration.
// Once a field is write-only, its previous value is no longer available for change detection.
func movedConnectionV2Secrets(secrets, stateValues map[string]interface{}) map[string]interface{} {
	moved := make(map[string]interface{})
	for name, value := range secrets {
		if _, ok := stateValues[name]; ok {
			moved[name] = value
		}
	}
	return moved
}
