package model

import (
	"context"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccountExternalLoggingMaskedValue is what the API returns in place of secret config values.
const AccountExternalLoggingMaskedValue = "******"

type AccountExternalLoggingV2ResourceModel struct {
	Id            types.String  `tfsdk:"id"`
	Service       types.String  `tfsdk:"service"`
	Enabled       types.Bool    `tfsdk:"enabled"`
	RunSetupTests types.Bool    `tfsdk:"run_setup_tests"`
	Config        types.Dynamic `tfsdk:"config"`
}

// ReadFromResponse maps an API response into the model. configMask holds the config the user manages
// (plan or prior state): only its keys are tracked, and masked secrets keep their local value.
// meta may be nil when the metadata API does not describe the log service.
func (d *AccountExternalLoggingV2ResourceModel) ReadFromResponse(ctx context.Context, resp core.AccountExternalLoggingResponse, meta *metadata.ConnectorMetadata, configMask map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Id = types.StringValue(resp.Data.Id)
	d.Service = types.StringValue(resp.Data.Service)
	d.Enabled = types.BoolValue(resp.Data.Enabled)

	var slot *metadata.Property
	if meta != nil {
		slot = &meta.Config
	}
	remote := keepMaskedLocally(resp.Data.Config, configMask)
	projected := core.ProjectDynamic(remote, configMask, slot)

	config, configDiags := core.MapToDynamic(ctx, projected)
	diags.Append(configDiags...)
	if !diags.HasError() {
		d.Config = config
	}

	return diags
}

// WithoutMaskedValues returns a copy of config without the secrets the API masked, so that an
// imported resource does not store the placeholder as if it were the real value.
func WithoutMaskedValues(config map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(config))
	for key, value := range config {
		switch v := value.(type) {
		case string:
			if v == AccountExternalLoggingMaskedValue {
				continue
			}
		case map[string]interface{}:
			value = WithoutMaskedValues(v)
		}
		result[key] = value
	}
	return result
}

// keepMaskedLocally replaces masked placeholders in remote with the local (mask) value, so that
// secrets never overwrite the configured value even when metadata does not flag them as passwords.
func keepMaskedLocally(remote, mask map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(remote))
	for key, value := range remote {
		maskValue, inMask := mask[key]
		switch v := value.(type) {
		case string:
			if v == AccountExternalLoggingMaskedValue && inMask {
				value = maskValue
			}
		case map[string]interface{}:
			if nestedMask, ok := maskValue.(map[string]interface{}); ok {
				value = keepMaskedLocally(v, nestedMask)
			}
		}
		result[key] = value
	}
	return result
}
