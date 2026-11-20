package model

import (
	"context"

	externallogging "github.com/fivetran/go-fivetran/external_logging"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExternalLoggingV2ResourceModel struct {
	Id                   types.String  `tfsdk:"id"`
	GroupId              types.String  `tfsdk:"group_id"`
	Service              types.String  `tfsdk:"service"`
	Enabled              types.Bool    `tfsdk:"enabled"`
	RunSetupTests        types.Bool    `tfsdk:"run_setup_tests"`
	Config               types.Dynamic `tfsdk:"config"`
	ConfigSecrets        types.Dynamic `tfsdk:"config_secrets"`
	ConfigSecretsVersion types.Int64   `tfsdk:"config_secrets_version"`
}

// ReadFromResponse maps an API response into the model. configMask holds the config the user manages
// (plan or prior state): only its keys are tracked, and masked secrets keep their local value.
func (d *ExternalLoggingV2ResourceModel) ReadFromResponse(ctx context.Context, resp externallogging.ExternalLoggingCustomResponse, meta *metadata.LogServiceMetadata, configMask map[string]interface{}, configTemplate types.Dynamic) diag.Diagnostics {
	var diags diag.Diagnostics

	// A destination has at most one log service, and its ID is the group (destination) ID.
	d.Id = types.StringValue(resp.Data.Id)
	d.GroupId = types.StringValue(resp.Data.Id)
	d.Service = types.StringValue(resp.Data.Service)
	d.Enabled = types.BoolValue(resp.Data.Enabled)

	var slot *metadata.Property
	if meta != nil {
		slot = &meta.Config
	}
	remote := keepMaskedLocally(resp.Data.Config, configMask)
	projected := core.ProjectDynamic(remote, configMask, slot)

	config, configDiags := core.MapToDynamicWithTemplate(ctx, projected, configTemplate)
	diags.Append(configDiags...)
	if !diags.HasError() {
		d.Config = config
	}

	return diags
}

// ExternalLoggingV2ImportConfigMask selects the config keys tracked in state after import: every
// field the API returns except masked secrets and password fields, which the API only returns
// masked. Readonly fields are dropped too: config rejects them, so tracking them would make users
// remove them from the generated block.
func ExternalLoggingV2ImportConfigMask(config map[string]interface{}, meta *metadata.LogServiceMetadata) map[string]interface{} {
	mask := WithoutMaskedValues(config)
	if meta == nil {
		return mask
	}
	for key := range mask {
		if prop := core.SlotProp(&meta.Config, key); prop != nil && (prop.Format == "password" || prop.Readonly) {
			delete(mask, key)
		}
	}
	return mask
}
