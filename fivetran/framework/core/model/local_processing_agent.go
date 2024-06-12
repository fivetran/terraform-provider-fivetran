package model

import (
    "context"

    "github.com/fivetran/go-fivetran/local_processing_agent"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
    elementType = map[string]attr.Type{
        "connection_id":    types.StringType,
        "schema":           types.StringType,
        "service":          types.StringType,
    }
)

type LocalProcessingAgent struct {
    Id                  types.String `tfsdk:"id"`
    DisplayName         types.String `tfsdk:"display_name"`
    GroupId             types.String `tfsdk:"group_id"`
    RegisteredAt        types.String `tfsdk:"registered_at"`
    ConfigJson          types.String `tfsdk:"config_json"`
    AuthJson            types.String `tfsdk:"auth_json"`
    DockerComposeYaml   types.String `tfsdk:"docker_compose_yaml"`
    ReAuth              types.Bool   `tfsdk:"re_auth"`
    Usage               types.Set    `tfsdk:"usage"`
}

func (d *LocalProcessingAgent) ReadFromResponse(ctx context.Context, resp localprocessingagent.LocalProcessingAgentDetailsResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.DisplayName = types.StringValue(resp.Data.DisplayName)
    d.GroupId = types.StringValue(resp.Data.GroupId)
    d.RegisteredAt = types.StringValue(resp.Data.RegisteredAt)

    if resp.Data.Usage == nil {
        d.Usage = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    for _, v := range resp.Data.Usage {
        item := map[string]attr.Value{}
        item["connection_id"] = types.StringValue(v.ConnectionId)
        item["schema"] = types.StringValue(v.Schema)
        item["service"] = types.StringValue(v.Service)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Usage, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *LocalProcessingAgent) ReadFromCreateResponse(ctx context.Context, resp localprocessingagent.LocalProcessingAgentCreateResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.DisplayName = types.StringValue(resp.Data.DisplayName)
    d.GroupId = types.StringValue(resp.Data.GroupId)
    d.RegisteredAt = types.StringValue(resp.Data.RegisteredAt)
    d.ConfigJson = types.StringValue(resp.Data.Files.ConfigJson)
    d.AuthJson = types.StringValue(resp.Data.Files.AuthJson)
    d.DockerComposeYaml = types.StringValue(resp.Data.Files.DockerComposeYaml)
}