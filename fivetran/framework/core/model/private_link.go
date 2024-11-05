package model

import (
    "context"

    "github.com/fivetran/go-fivetran/private_link"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type PrivateLink struct {
    Id              types.String `tfsdk:"id"`
    Name            types.String `tfsdk:"name"`
    Region          types.String `tfsdk:"region"`
    Service         types.String `tfsdk:"service"`
    CloudProvider   types.String `tfsdk:"cloud_provider"`
    State           types.String `tfsdk:"state"`
    StateSummary    types.String `tfsdk:"state_summary"`
    CreatedAt       types.String `tfsdk:"created_at"`
    CreatedBy       types.String `tfsdk:"created_by"`
    ConfigMap       types.Map    `tfsdk:"config_map"`
}

var PrivateLinkConfigType = map[string]attr.Type{
        "connection_service_name":        types.StringType,
        "account_url":                    types.StringType,
        "vpce_id":                        types.StringType,
        "aws_account_id":                 types.StringType,
        "cluster_identifier":             types.StringType,
        "connection_service_id":          types.StringType,
        "workspace_url":                  types.StringType,
        "pls_id":                         types.StringType,
        "sub_resource_name":              types.StringType,
        "private_dns_regions":            types.StringType,
        "private_connection_service_id":  types.StringType,
}

func (d *PrivateLink) ReadFromResponse(ctx context.Context, resp privatelink.PrivateLinkResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Service = types.StringValue(resp.Data.Service)
    d.Region = types.StringValue(resp.Data.Region)
    d.CloudProvider = types.StringValue(resp.Data.CloudProvider)
    d.State = types.StringValue(resp.Data.State)
    d.StateSummary = types.StringValue(resp.Data.StateSummary)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
    d.CreatedBy = types.StringValue(resp.Data.CreatedBy)

    config := map[string]attr.Value{}

    if resp.Data.Config.ConnectionServiceName != "" {
        config["connection_service_name"] = types.StringValue(resp.Data.Config.ConnectionServiceName)
    } else {
        config["connection_service_name"] = types.StringNull()
    }

    if resp.Data.Config.AccountUrl != "" {
        config["account_url"] = types.StringValue(resp.Data.Config.AccountUrl)
    } else {
        config["account_url"] = types.StringNull()
    }

    if resp.Data.Config.VpceId != "" {
        config["vpce_id"] = types.StringValue(resp.Data.Config.VpceId)
    } else {
        config["vpce_id"] = types.StringNull()
    }

    if resp.Data.Config.AwsAccountId != "" {
        config["aws_account_id"] = types.StringValue(resp.Data.Config.AwsAccountId)
    } else {
        config["aws_account_id"] = types.StringNull()
    }

    if resp.Data.Config.ClusterIdentifier != "" {
        config["cluster_identifier"] = types.StringValue(resp.Data.Config.ClusterIdentifier)  
    } else {
        config["cluster_identifier"] = types.StringNull()
    }

    if resp.Data.Config.ConnectionServiceId != "" {
        config["connection_service_id"] = types.StringValue(resp.Data.Config.ConnectionServiceId)
    } else {
        config["connection_service_id"] = types.StringNull()
    }

    if resp.Data.Config.WorkspaceUrl != "" {
        config["workspace_url"] = types.StringValue(resp.Data.Config.WorkspaceUrl)
    } else {
        config["workspace_url"] = types.StringNull()
    }

    if resp.Data.Config.PlsId != "" {
        config["pls_id"] = types.StringValue(resp.Data.Config.PlsId)
    } else {
        config["pls_id"] = types.StringNull()
    }

    if resp.Data.Config.SubResourceName != "" {
        config["sub_resource_name"] = types.StringValue(resp.Data.Config.SubResourceName)
    } else {
        config["sub_resource_name"] = types.StringNull()
    }

    if resp.Data.Config.PrivateDnsRegions != "" {
        config["private_dns_regions"] = types.StringValue(resp.Data.Config.PrivateDnsRegions)
    } else {
        config["private_dns_regions"] = types.StringNull()
    }

    if resp.Data.Config.PrivateConnectionServiceId != "" {
        config["private_connection_service_id"] = types.StringValue(resp.Data.Config.PrivateConnectionServiceId)
    } else {
        config["private_connection_service_id"] = types.StringNull()
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

    config := map[string]attr.Value{}
    for k, v := range resp.Data.Config {
        config[k] = types.StringValue(v.(string))
    }
    d.ConfigMap, _ = types.MapValue(types.StringType, config)
}

func (d *PrivateLink) GetConfigMap() map[string]interface{} {
    config := make(map[string]interface{})
    for k, v := range d.ConfigMap.Elements(){
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