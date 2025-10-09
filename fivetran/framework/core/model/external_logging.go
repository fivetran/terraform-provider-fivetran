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

    if resp.Data.Config.WorkspaceId != "" {
        config["workspace_id"] = types.StringValue(resp.Data.Config.WorkspaceId)
    } else {
        config["workspace_id"] = types.StringNull()
    }

    if resp.Data.Config.LogGroupName != "" {
        config["log_group_name"] = types.StringValue(resp.Data.Config.LogGroupName)
    } else {
        config["log_group_name"] = types.StringNull()
    }

    if resp.Data.Config.RoleArn != "" {
        config["role_arn"] = types.StringValue(resp.Data.Config.RoleArn)
    } else {
        config["role_arn"] = types.StringNull()
    }

    if resp.Data.Config.ExternalId != "" {
        config["external_id"] = types.StringValue(resp.Data.Config.ExternalId)
    } else {
        config["external_id"] = types.StringNull()
    }

    if resp.Data.Config.Region != "" {
        config["region"] = types.StringValue(resp.Data.Config.Region)  
    } else {
        config["region"] = types.StringNull()
    }

    if resp.Data.Config.SubDomain != "" {
        config["sub_domain"] = types.StringValue(resp.Data.Config.SubDomain)
    } else {
        config["sub_domain"] = types.StringNull()
    }

    if resp.Data.Config.Host != "" {
        config["host"] = types.StringValue(resp.Data.Config.Host)
    } else {
        config["host"] = types.StringNull()
    }

    if resp.Data.Config.Hostname != "" {
        config["hostname"] = types.StringValue(resp.Data.Config.Hostname)
    } else {
        config["hostname"] = types.StringNull()
    }

    if resp.Data.Config.Channel != "" {
        config["channel"] = types.StringValue(resp.Data.Config.Channel)
    } else {
        config["channel"] = types.StringNull()
    }

    if resp.Data.Config.ProjectId != "" {
        config["project_id"] = types.StringValue(resp.Data.Config.ProjectId)
    } else {
        config["project_id"] = types.StringNull()
    }

    config["enable_ssl"] = types.BoolValue(resp.Data.Config.EnableSsl)    
    config["port"] = types.Int64Value(int64(resp.Data.Config.Port))

    if resp.Data.Config.PrimaryKey != "" {
        if resp.Data.Config.PrimaryKey != "******" {
            config["primary_key"] = types.StringValue(resp.Data.Config.PrimaryKey)
        } else {
            config["primary_key"] = CoalesceToStringNull(d.Config.Attributes()["primary_key"])
        }
    } else {
        config["primary_key"] = types.StringNull()
    }
    
    if resp.Data.Config.ApiKey != "" {
        if resp.Data.Config.ApiKey != "******" {
            config["api_key"] = types.StringValue(resp.Data.Config.ApiKey)
        } else {
            config["api_key"] = CoalesceToStringNull(d.Config.Attributes()["api_key"])
        }
    } else {
        config["api_key"] = types.StringNull()
    }

    if resp.Data.Config.Token != "" {
        if resp.Data.Config.Token != "******" {
            config["token"] = types.StringValue(resp.Data.Config.Token)
        } else {
            config["token"] = CoalesceToStringNull(d.Config.Attributes()["token"])
        }
    } else {
        config["token"] = types.StringNull()
    }

    d.Config, _ = types.ObjectValue(ExternalLoggingTFConfigType, config)
}

func CoalesceToStringNull(value attr.Value) attr.Value {
	if value == nil {
        return types.StringNull()
    }

    return value
}

func (d *ExternalLogging) ReadFromCustomResponse(ctx context.Context, resp externallogging.ExternalLoggingCustomResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.GroupId = types.StringValue(resp.Data.Id)
    d.Service = types.StringValue(resp.Data.Service)
    d.Enabled = types.BoolValue(resp.Data.Enabled)

    config := map[string]attr.Value{}
    if resp.Data.Config["workspace_id"] != nil && resp.Data.Config["workspace_id"] != "" {
        config["workspace_id"] = types.StringValue(resp.Data.Config["workspace_id"].(string))
    } else {
        config["workspace_id"] = types.StringNull()
    }
    
    if resp.Data.Config["log_group_name"] != nil && resp.Data.Config["log_group_name"] != "" {
        config["log_group_name"] = types.StringValue(resp.Data.Config["log_group_name"].(string))
    } else {
        config["log_group_name"] = types.StringNull()
    }
    
    if resp.Data.Config["role_arn"] != nil && resp.Data.Config["role_arn"] != "" {
        config["role_arn"] = types.StringValue(resp.Data.Config["role_arn"].(string))
    } else {
        config["role_arn"] = types.StringNull()
    }

    if resp.Data.Config["external_id"] != nil && resp.Data.Config["external_id"] != "" {
        config["external_id"] = types.StringValue(resp.Data.Config["external_id"].(string))
    } else {
        config["external_id"] = types.StringNull()
    }
    
    if resp.Data.Config["region"] != nil && resp.Data.Config["region"] != "" {
        config["region"] = types.StringValue(resp.Data.Config["region"].(string))
    } else {
        config["region"] = types.StringNull()
    }
    
    if resp.Data.Config["sub_domain"] != nil && resp.Data.Config["sub_domain"] != "" {
        config["sub_domain"] = types.StringValue(resp.Data.Config["sub_domain"].(string))
    } else {
        config["sub_domain"] = types.StringNull()
    }
    
    if resp.Data.Config["host"] != nil && resp.Data.Config["host"] != "" {
        config["host"] = types.StringValue(resp.Data.Config["host"].(string))
    } else {
        config["host"] = types.StringNull()
    }
    
    if resp.Data.Config["hostname"] != nil && resp.Data.Config["hostname"] != "" {
        config["hostname"] = types.StringValue(resp.Data.Config["hostname"].(string))
    } else {
        config["hostname"] = types.StringNull()
    }
    
    if resp.Data.Config["channel"] != nil && resp.Data.Config["channel"] != "" {
        config["channel"] = types.StringValue(resp.Data.Config["channel"].(string))
    } else {
        config["channel"] = types.StringNull()
    }
    
    if resp.Data.Config["project_id"] != nil && resp.Data.Config["project_id"] != "" {
        config["project_id"] = types.StringValue(resp.Data.Config["project_id"].(string))
    } else {
        config["project_id"] = types.StringNull()
    }

    if resp.Data.Config["enable_ssl"] != nil {
        config["enable_ssl"] = types.BoolValue(resp.Data.Config["enable_ssl"].(bool))    
    } else {
        config["enable_ssl"] = types.BoolValue(false)
    }

    if resp.Data.Config["port"] != nil {
        config["port"] = types.Int64Value(int64(resp.Data.Config["port"].(float64)))
    } else {
        config["port"] = types.Int64Value(0)
    }

    if resp.Data.Config["primary_key"] != nil && resp.Data.Config["primary_key"] != "" && resp.Data.Config["primary_key"] != "******" {
        config["primary_key"] = types.StringValue(resp.Data.Config["primary_key"].(string))
    } else if !d.Config.Attributes()["primary_key"].IsNull() {
        config["primary_key"] = d.Config.Attributes()["primary_key"]
    } else {
        config["primary_key"] = types.StringNull()
    }
     
    if resp.Data.Config["api_key"] != nil && resp.Data.Config["api_key"] != "" && resp.Data.Config["api_key"] != "******" {
        config["api_key"] = types.StringValue(resp.Data.Config["api_key"].(string))
    } else if !d.Config.Attributes()["api_key"].IsNull() {
        config["api_key"] = d.Config.Attributes()["api_key"]
    } else {
        config["api_key"] = types.StringNull()
    }

    if resp.Data.Config["token"] != nil && resp.Data.Config["token"] != "" && resp.Data.Config["token"] != "******" {
        config["token"] = types.StringValue(resp.Data.Config["token"].(string))
    } else if !d.Config.Attributes()["token"].IsNull() {
        config["token"] = d.Config.Attributes()["token"]
    } else {
        config["token"] = types.StringNull()
    }

    d.Config, _ = types.ObjectValue(ExternalLoggingTFConfigType, config)
}

func (d *ExternalLogging) GetConfig() map[string]interface{} {
    attr := d.Config.Attributes()

    config := make(map[string]interface{})
    for k, v := range attr{
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