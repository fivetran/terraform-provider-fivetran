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

// destinationV2ConfigSecretsValue reads config_secrets from the raw config. Write-only
// values are only available there: Terraform always sends them as null in plan and state.
func destinationV2ConfigSecretsValue(ctx context.Context, config tfsdk.Config) (types.Dynamic, diag.Diagnostics) {
	secrets := types.DynamicNull()
	if config.Raw.IsNull() {
		return secrets, nil
	}
	diags := config.GetAttribute(ctx, path.Root("config_secrets"), &secrets)
	return secrets, diags
}

func destinationV2ConfigSecrets(ctx context.Context, config tfsdk.Config) (map[string]interface{}, diag.Diagnostics) {
	secrets, diags := destinationV2ConfigSecretsValue(ctx, config)
	if diags.HasError() {
		return nil, diags
	}
	secretsMap, convertDiags := core.DynamicToMap(ctx, secrets)
	diags.Append(convertDiags...)
	return secretsMap, diags
}

// mergeConfigSecrets returns a copy of config with the secret fields added, for the create request.
func mergeConfigSecrets(config, secrets map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{}, len(config)+len(secrets))
	for k, v := range config {
		merged[k] = v
	}
	for k, v := range secrets {
		merged[k] = v
	}
	return merged
}

// applyConfigSecretsToPatch adds the secret fields to an update patch when send is true
// (config_secrets_version changed). Otherwise it only drops the null that
// PrepareConfigPatchDynamic emits for a field moved from config to config_secrets,
// so the move doesn't clear the secret on the server.
func applyConfigSecretsToPatch(patch, secrets map[string]interface{}, send bool) {
	for k, v := range secrets {
		if send {
			patch[k] = v
			continue
		}
		if patchVal, ok := patch[k]; ok && patchVal == nil {
			delete(patch, k)
		}
	}
}

// validateConfigSecrets checks that config_secrets holds only secret fields (format
// "password" in metadata) and that none of them is also set in config. Fields unknown to
// metadata are reported by validateDynamicObject.
func validateConfigSecrets(secrets, config map[string]interface{}, slot *metadata.Property, diags *diag.Diagnostics) {
	for name := range secrets {
		fieldPath := path.Root("config_secrets").AtName(name)
		if _, ok := config[name]; ok {
			diags.AddAttributeError(
				fieldPath,
				"Duplicate Secret Field",
				fmt.Sprintf("Field %q is set in both config and config_secrets. Set it in only one of them.", name),
			)
			continue
		}
		if prop := core.SlotProp(slot, name); prop != nil && prop.Format != "password" {
			diags.AddAttributeError(
				fieldPath,
				"Non-Secret Field in config_secrets",
				fmt.Sprintf("Field %q is not a secret field. Set it in config instead.", name),
			)
		}
	}
}
