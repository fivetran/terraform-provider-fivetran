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

var configSecretsPath = path.Root("config_secrets")

// configSecretsFromConfig reads config_secrets from the configuration. It is a write-only
// attribute, so it is always null in the plan and the state. Unknown values are kept when
// preserveUnknown is set, for plan-time checks.
func configSecretsFromConfig(ctx context.Context, config tfsdk.Config, preserveUnknown bool, diags *diag.Diagnostics) map[string]interface{} {
	if config.Raw.IsNull() {
		return nil
	}

	var value types.Dynamic
	diags.Append(config.GetAttribute(ctx, configSecretsPath, &value)...)
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

// mergeConfigSecrets returns config with the config_secrets fields added. Neither map is modified.
func mergeConfigSecrets(config, secrets map[string]interface{}) map[string]interface{} {
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

// validateConfigSecretsOverlap reports every top-level field set in both config and config_secrets.
func validateConfigSecretsOverlap(config, secrets map[string]interface{}, diags *diag.Diagnostics) {
	for name := range secrets {
		if _, ok := config[name]; ok {
			diags.AddAttributeError(
				configSecretsPath.AtName(name),
				"Conflicting Dynamic Field",
				fmt.Sprintf("Field %q is set in both config and config_secrets. Set it in only one of them.", name),
			)
		}
	}
}

// applyConfigSecretsToPatch adds mutable config_secrets fields when their version changed or
// they moved from config. This prevents removed nullable fields from being cleared on migration.
// Immutable and read-only fields follow the same PATCH exclusions as config; moving them only
// removes their stored values, without sending them to the API again.
func applyConfigSecretsToPatch(patch, secrets, stateConfig map[string]interface{}, slot *metadata.Property, versionChanged bool) map[string]interface{} {
	for name, value := range secrets {
		if prop := core.SlotProp(slot, name); prop != nil && (prop.Immutable || prop.Readonly) {
			continue
		}
		if _, moved := stateConfig[name]; versionChanged || moved {
			if patch == nil {
				patch = make(map[string]interface{}, len(secrets))
			}
			patch[name] = value
		}
	}
	return patch
}

// movedConfigSecrets returns fields whose prior values can still be compared during migration.
// Once a field is write-only, its previous value is no longer available for change detection.
func movedConfigSecrets(secrets, stateConfig map[string]interface{}) map[string]interface{} {
	moved := make(map[string]interface{})
	for name, value := range secrets {
		if _, ok := stateConfig[name]; ok {
			moved[name] = value
		}
	}
	return moved
}
