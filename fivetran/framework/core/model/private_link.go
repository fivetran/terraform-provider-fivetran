package model

import (
    "context"

    "github.com/fivetran/go-fivetran/private_link"
    "github.com/hashicorp/terraform-plugin-framework/types"
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
    
    Config          types.Object `tfsdk:"config"`
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

    if resp.Data.Config.PrivateConnectionServiceId != "" {
        config["private_connection_service_id"] = types.StringValue(resp.Data.Config.PrivateConnectionServiceId)
    } else {
        config["private_connection_service_id"] = types.StringNull()
    }

    d.Config, _ = types.ObjectValue(PrivateLinkConfigType, config)
}

func (d *PrivateLink) GetConfig() privatelink.PrivateLinkConfig {
    var config privatelink.PrivateLinkConfig

    if !d.Config.Attributes()["connection_service_name"].IsNull() {
        config.ConnectionServiceName(d.Config.Attributes()["connection_service_name"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["account_url"].IsNull() {
        config.AccountUrl(d.Config.Attributes()["account_url"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["vpce_id"].IsNull() {
        config.VpceId(d.Config.Attributes()["vpce_id"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["aws_account_id"].IsNull() {
        config.AwsAccountId(d.Config.Attributes()["aws_account_id"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["cluster_identifier"].IsNull() {
        config.ClusterIdentifier(d.Config.Attributes()["cluster_identifier"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["workspace_url"].IsNull() {
        config.WorkspaceUrl(d.Config.Attributes()["workspace_url"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["pls_id"].IsNull() {
        config.PlsId(d.Config.Attributes()["pls_id"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["sub_resource_name"].IsNull() {
        config.SubResourceName(d.Config.Attributes()["sub_resource_name"].(types.String).ValueString())
    }

    if !d.Config.Attributes()["private_connection_service_id"].IsNull() {
        config.PrivateConnectionServiceId(d.Config.Attributes()["private_connection_service_id"].(types.String).ValueString())
    }

    return config
}