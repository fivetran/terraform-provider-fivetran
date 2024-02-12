package model

import (
    "context"

    externallogging "github.com/fivetran/go-fivetran/external_logging"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type ExternalLogging struct {
    Id         types.String `tfsdk:"id"`
    GroupId    types.String `tfsdk:"group_id"`
    Service    types.String `tfsdk:"service"`
    Enabled    types.Bool   `tfsdk:"enabled"`
    RunTests   types.Bool   `tfsdk:"run_setup_tests"`
    Config     types.Object `tfsdk:"config"`
}

var ExternalLoggingTFConfigType = map[string]attr.Type{
        "workspace_id":     types.StringType,
        "port":             types.Int64Type,
        "log_group_name":   types.StringType,
        "role_arn":         types.StringType,
        "external_id":      types.StringType,
        "region":           types.StringType,
        "sub_domain":       types.StringType,
        "host":             types.StringType,
        "hostname":         types.StringType,
        "enable_ssl":       types.BoolType,
        "channel":          types.StringType,
        "project_id":       types.StringType,
        "primary_key":      types.StringType,
        "api_key":          types.StringType,
        "token":            types.StringType,
}

func (d *ExternalLogging) ReadFromResponse(ctx context.Context, resp externallogging.ExternalLoggingResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.GroupId = types.StringValue(resp.Data.Id)
    d.Service = types.StringValue(resp.Data.Service)
    d.Enabled = types.BoolValue(resp.Data.Enabled)

    config := map[string]attr.Value{}
    config["workspace_id"] = types.StringValue(resp.Data.Config.WorkspaceId)
    config["port"] = types.Int64Value(int64(resp.Data.Config.Port))
    config["log_group_name"] = types.StringValue(resp.Data.Config.LogGroupName)
    config["role_arn"] = types.StringValue(resp.Data.Config.RoleArn)
    config["external_id"] = types.StringValue(resp.Data.Config.ExternalId)
    config["region"] = types.StringValue(resp.Data.Config.Region)
    config["sub_domain"] = types.StringValue(resp.Data.Config.SubDomain)
    config["host"] = types.StringValue(resp.Data.Config.Host)
    config["hostname"] = types.StringValue(resp.Data.Config.Hostname)
    config["enable_ssl"] = types.BoolValue(resp.Data.Config.EnableSsl)
    config["channel"] = types.StringValue(resp.Data.Config.Channel)
    config["project_id"] = types.StringValue(resp.Data.Config.ProjectId)

    if resp.Data.Config.PrimaryKey != "******" {
        config["primary_key"] = types.StringValue(resp.Data.Config.PrimaryKey)
    } else {
        config["primary_key"] = d.Config.Attributes()["primary_key"]
    }
     
    if resp.Data.Config.ApiKey != "******" {
        config["api_key"] = types.StringValue(resp.Data.Config.ApiKey)
    } else {
        config["api_key"] = d.Config.Attributes()["api_key"]
    }

    if resp.Data.Config.Token != "******" {
        config["token"] = types.StringValue(resp.Data.Config.Token)
    } else {
        config["token"] = d.Config.Attributes()["token"]
    }

    d.Config, _ = types.ObjectValue(ExternalLoggingTFConfigType, config)
}

func (d *ExternalLogging) ReadFromCreateResponse(ctx context.Context, resp externallogging.ExternalLoggingCustomResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.GroupId = types.StringValue(resp.Data.Id)
    d.Service = types.StringValue(resp.Data.Service)
    d.Enabled = types.BoolValue(resp.Data.Enabled)

    config := map[string]attr.Value{}
    config["workspace_id"] = types.StringValue(resp.Data.Config["workspace_id"].(string))
    config["port"] = types.Int64Value(int64(resp.Data.Config["port"].(float64)))
    config["log_group_name"] = types.StringValue(resp.Data.Config["log_group_name"].(string))
    config["role_arn"] = types.StringValue(resp.Data.Config["role_arn"].(string))
    config["external_id"] = types.StringValue(resp.Data.Config["external_id"].(string))
    config["region"] = types.StringValue(resp.Data.Config["region"].(string))
    config["sub_domain"] = types.StringValue(resp.Data.Config["sub_domain"].(string))
    config["host"] = types.StringValue(resp.Data.Config["host"].(string))
    config["hostname"] = types.StringValue(resp.Data.Config["hostname"].(string))
    config["enable_ssl"] = types.BoolValue(resp.Data.Config["enable_ssl"].(bool))
    config["channel"] = types.StringValue(resp.Data.Config["channel"].(string))
    config["project_id"] = types.StringValue(resp.Data.Config["project_id"].(string))

    if resp.Data.Config["primary_key"] != "******" {
        config["primary_key"] = types.StringValue(resp.Data.Config["primary_key"].(string))
    } else {
        config["primary_key"] = d.Config.Attributes()["primary_key"]
    }
     
    if resp.Data.Config["api_key"] != "******" {
        config["api_key"] = types.StringValue(resp.Data.Config["api_key"].(string))
    } else {
        config["api_key"] = d.Config.Attributes()["api_key"]
    }

    if resp.Data.Config["token"] != "******" {
        config["token"] = types.StringValue(resp.Data.Config["token"].(string))
    } else {
        config["token"] = d.Config.Attributes()["token"]
    }

    d.Config, _ = types.ObjectValue(ExternalLoggingTFConfigType, config)
}

func (d *ExternalLogging) GetConfig() map[string]interface{} {
    attr := d.Config.Attributes()

    config := make(map[string]interface{})
    for k, v := range attr{
        if !v.IsUnknown() {
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