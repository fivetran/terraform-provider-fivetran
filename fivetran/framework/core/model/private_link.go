package model

import (
	"context"

	privatelink "github.com/fivetran/go-fivetran/private_link"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type PrivateLink struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Region        types.String `tfsdk:"region"`
	Service       types.String `tfsdk:"service"`
	CloudProvider types.String `tfsdk:"cloud_provider"`
	State         types.String `tfsdk:"state"`
	StateSummary  types.String `tfsdk:"state_summary"`
	CreatedAt     types.String `tfsdk:"created_at"`
	CreatedBy     types.String `tfsdk:"created_by"`
	Host          types.String `tfsdk:"host"`
	ConfigMap     types.Map    `tfsdk:"config_map"`
}

var PrivateLinkConfigType = map[string]attr.Type{
	"connection_service_name":       types.StringType,
	"account_url":                   types.StringType,
	"vpce_id":                       types.StringType,
	"aws_account_id":                types.StringType,
	"cluster_identifier":            types.StringType,
	"connection_service_id":         types.StringType,
	"workspace_url":                 types.StringType,
	"pls_id":                        types.StringType,
	"sub_resource_name":             types.StringType,
	"private_dns_records":           types.StringType,
	"private_connection_service_id": types.StringType,
}

// ReadFromResponse is used by the data source: every documented config key is present
// (null when the API doesn't return it), plus any other keys the API returns.
func (d *PrivateLink) ReadFromResponse(ctx context.Context, resp privatelink.PrivateLinkCustomResponse) {
	d.ReadFromCustomResponse(ctx, resp)

	config := map[string]attr.Value{}
	for k := range PrivateLinkConfigType {
		config[k] = types.StringNull()
	}
	for k, v := range d.ConfigMap.Elements() {
		config[k] = v
	}
	d.ConfigMap, _ = types.MapValue(types.StringType, config)
}

func (d *PrivateLink) ReadFromCustomResponse(ctx context.Context, resp privatelink.PrivateLinkCustomResponse) {
	d.Id = types.StringValue(resp.Data.Id)
	d.Name = types.StringValue(resp.Data.Name)
	d.Service = types.StringValue(resp.Data.Service)
	d.Region = types.StringValue(resp.Data.Region)
	d.CloudProvider = types.StringValue(resp.Data.CloudProvider)
	d.State = types.StringValue(resp.Data.State)
	d.StateSummary = types.StringValue(resp.Data.StateSummary)
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
	d.CreatedBy = types.StringValue(resp.Data.CreatedBy)
	d.Host = types.StringValue(resp.Data.Host)

	config := map[string]attr.Value{}
	for k, v := range resp.Data.Config {
		if v != nil {
			if str, ok := v.(string); ok && str != "" {
				config[k] = types.StringValue(str)
			}
			// Skip nil values and empty strings - don't add them to the config map
		}
	}
	d.ConfigMap, _ = types.MapValue(types.StringType, config)
}

func (d *PrivateLink) GetConfigMap() map[string]interface{} {
	config := make(map[string]interface{})
	for k, v := range d.ConfigMap.Elements() {
		if !v.IsUnknown() && !v.IsNull() {
			if t, ok := v.(basetypes.Int64Value); ok {
				config[k] = t.ValueInt64()
			}

			if t, ok := v.(basetypes.BoolValue); ok {
				config[k] = t.ValueBool()
			}

			if t, ok := v.(basetypes.StringValue); ok {
				config[k] = t.ValueString()
			}
		}
	}

	return config
}
