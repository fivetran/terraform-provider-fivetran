package resources

import (
	"context"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// externalLoggingV2ConfigSecretsValue reads config_secrets from the raw config. Write-only
// values are only available there: Terraform always sends them as null in plan and state.
func externalLoggingV2ConfigSecretsValue(ctx context.Context, config tfsdk.Config) (types.Dynamic, diag.Diagnostics) {
	secrets := types.DynamicNull()
	if config.Raw.IsNull() {
		return secrets, nil
	}
	diags := config.GetAttribute(ctx, path.Root("config_secrets"), &secrets)
	return secrets, diags
}

func externalLoggingV2ConfigSecrets(ctx context.Context, config tfsdk.Config) (map[string]interface{}, diag.Diagnostics) {
	secrets, diags := externalLoggingV2ConfigSecretsValue(ctx, config)
	if diags.HasError() {
		return nil, diags
	}
	secretsMap, convertDiags := core.DynamicToMap(ctx, secrets)
	diags.Append(convertDiags...)
	return secretsMap, diags
}

// validateExternalLoggingV2RequiredFields checks required config fields, counting secret
// fields provided through config_secrets. It skips the check while config_secrets is
// unknown, because the provided secret fields can't be known yet.
func validateExternalLoggingV2RequiredFields(ctx context.Context, config tfsdk.Config, planConfig map[string]interface{}, meta *metadata.LogServiceMetadata, diags *diag.Diagnostics) {
	secrets, secretsDiags := externalLoggingV2ConfigSecretsValue(ctx, config)
	diags.Append(secretsDiags...)
	if diags.HasError() {
		return
	}
	if secrets.IsUnknown() || secrets.IsUnderlyingValueUnknown() {
		return
	}
	secretsMap, secretsDiags := core.DynamicToMapPreserveUnknown(ctx, secrets)
	diags.Append(secretsDiags...)
	if diags.HasError() {
		return
	}
	validateRequiredDynamicFields(mergeConfigSecrets(planConfig, secretsMap), &meta.Config, path.Root("config"), diags, externalLoggingV2Wording)
}
