package model

import (
	"context"

	"github.com/fivetran/go-fivetran/destinations"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DestinationV2ResourceModel struct {
	Id                        types.String  `tfsdk:"id"`
	GroupId                   types.String  `tfsdk:"group_id"`
	Service                   types.String  `tfsdk:"service"`
	Config                    types.Dynamic `tfsdk:"config"`
	ConfigSecrets             types.Dynamic `tfsdk:"config_secrets"`
	ConfigSecretsVersion      types.Int64   `tfsdk:"config_secrets_version"`
	Region                    types.String  `tfsdk:"region"`
	TimeZoneOffset            types.String  `tfsdk:"time_zone_offset"`
	DaylightSavingTimeEnabled types.Bool    `tfsdk:"daylight_saving_time_enabled"`
	SetupStatus               types.String  `tfsdk:"setup_status"`

	NetworkingMethod        types.String `tfsdk:"networking_method"`
	PrivateLinkId           types.String `tfsdk:"private_link_id"`
	ProxyAgentId            types.String `tfsdk:"proxy_agent_id"`
	HybridDeploymentAgentId types.String `tfsdk:"hybrid_deployment_agent_id"`

	RunSetupTests     types.Bool `tfsdk:"run_setup_tests"`
	TrustCertificates types.Bool `tfsdk:"trust_certificates"`
	TrustFingerprints types.Bool `tfsdk:"trust_fingerprints"`
}

func DestinationV2ResourceModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"group_id":                     types.StringType,
		"service":                      types.StringType,
		"config":                       types.DynamicType,
		"config_secrets":               types.DynamicType,
		"config_secrets_version":       types.Int64Type,
		"region":                       types.StringType,
		"time_zone_offset":             types.StringType,
		"daylight_saving_time_enabled": types.BoolType,
		"setup_status":                 types.StringType,
		"networking_method":            types.StringType,
		"private_link_id":              types.StringType,
		"proxy_agent_id":               types.StringType,
		"hybrid_deployment_agent_id":   types.StringType,
		"run_setup_tests":              types.BoolType,
		"trust_certificates":           types.BoolType,
		"trust_fingerprints":           types.BoolType,
	}
}

func (d *DestinationV2ResourceModel) ReadFromCreateResponse(ctx context.Context, resp destinations.DestinationDetailsWithSetupTestsCustomResponse, meta *metadata.DestinationMetadata, configMask map[string]interface{}) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DestinationDetailsBase, resp.Data.Config, meta, configMask)
}

func (d *DestinationV2ResourceModel) ReadFromResponse(ctx context.Context, resp destinations.DestinationDetailsCustomResponse, meta *metadata.DestinationMetadata, configMask map[string]interface{}) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DestinationDetailsBase, resp.Data.Config, meta, configMask)
}

func (d *DestinationV2ResourceModel) ReadFromResponseForImport(ctx context.Context, resp destinations.DestinationDetailsCustomResponse, meta *metadata.DestinationMetadata) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DestinationDetailsBase, resp.Data.Config, meta, importConfigMask(resp.Data.Config, meta))
}

func (d *DestinationV2ResourceModel) readFromResponseData(ctx context.Context, data destinations.DestinationDetailsBase, config map[string]interface{}, meta *metadata.DestinationMetadata, configMask map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Id = types.StringValue(data.ID)
	d.GroupId = types.StringValue(data.GroupID)
	d.Service = types.StringValue(data.Service)
	d.Region = stringValueOrNull(data.Region)
	d.TimeZoneOffset = types.StringValue(data.TimeZoneOffset)
	d.DaylightSavingTimeEnabled = types.BoolValue(data.DaylightSavingTimeEnabled)
	d.SetupStatus = stringValueOrNull(data.SetupStatus)

	d.NetworkingMethod = stringValueOrNull(data.NetworkingMethod)
	d.PrivateLinkId = stringValueOrNull(data.PrivateLinkId)
	d.ProxyAgentId = stringValueOrNull(data.ProxyAgentId)
	d.HybridDeploymentAgentId = stringValueOrNull(data.HybridDeploymentAgentId)

	configSlot := (*metadata.Property)(nil)
	if meta != nil {
		configSlot = &meta.Config
	}
	projectedConfig := core.ProjectDynamic(config, configMask, configSlot)
	dynamicConfig, dynamicDiags := core.MapToDynamic(ctx, projectedConfig)
	diags.Append(dynamicDiags...)
	if !diags.HasError() {
		d.Config = dynamicConfig
	}

	return diags
}

// importConfigMask selects the config keys tracked in state after import: every
// field the API returns except password fields, which the API only returns masked.
// Readonly fields are dropped too: they can't be set in HCL, so tracking them would
// force users to remove them from the generated block.
func importConfigMask(config map[string]interface{}, meta *metadata.DestinationMetadata) map[string]interface{} {
	mask := make(map[string]interface{}, len(config))
	for key, value := range config {
		if meta != nil {
			if prop := core.SlotProp(&meta.Config, key); prop != nil && (prop.Format == "password" || prop.Readonly) {
				continue
			}
		}
		mask[key] = value
	}
	return mask
}
