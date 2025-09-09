package model

import (
    "context"
    "fmt"

    "github.com/fivetran/go-fivetran/connections"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Connections struct {
    Id             types.String `tfsdk:"id"` 
    GroupId        types.String `tfsdk:"group_id"`
    ConnectionName types.String `tfsdk:"connection_name"`
    Connections    types.Set    `tfsdk:"connections"`
}

func (d *Connections) ReadFromResponse(ctx context.Context, resp connections.ConnectionsListResponse) {
    elementAttrType := map[string]attr.Type{
        "id":                           types.StringType,
        "name":                         types.StringType,
        "connected_by":                 types.StringType,
        "created_at":                   types.StringType,
        "group_id":                     types.StringType,
        "service":                      types.StringType,
        "succeeded_at":                 types.StringType,
        "failed_at":                    types.StringType,
        "service_version":              types.StringType,
        "sync_frequency":               types.Int64Type,
        "schedule_type":                types.StringType,
        "paused":                       types.BoolType,
        "pause_after_trial":            types.BoolType,
        "daily_sync_time":              types.StringType,
        "data_delay_sensitivity":       types.StringType,
        "data_delay_threshold":         types.Int64Type,
        "proxy_agent_id":               types.StringType,
        "networking_method":            types.StringType,
        "hybrid_deployment_agent_id":   types.StringType,
        "private_link_id":              types.StringType,
    }

    if resp.Data.Items == nil {
        d.Connections = types.SetNull(types.ObjectType{AttrTypes: elementAttrType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.ID)
        item["name"] = types.StringValue(v.Schema)
        item["connected_by"] = types.StringValue(v.ConnectedBy)
        item["created_at"] = types.StringValue(v.CreatedAt.String())
        item["group_id"] = types.StringValue(v.GroupID)
        item["service"] = types.StringValue(v.Service)
        item["succeeded_at"] = types.StringValue(v.SucceededAt.String())
        item["failed_at"] = types.StringValue(v.FailedAt.String())
        item["service_version"] = types.StringValue(fmt.Sprintf("%v", *v.ServiceVersion))
        item["sync_frequency"] = types.Int64Value(int64(*v.SyncFrequency))
        item["schedule_type"] = types.StringValue(v.ScheduleType)
        item["paused"] = types.BoolValue(*v.Paused)
        item["pause_after_trial"] = types.BoolValue(*v.PauseAfterTrial)
        item["daily_sync_time"] = types.StringValue(v.DailySyncTime)
        item["data_delay_sensitivity"] = types.StringValue(v.DataDelaySensitivity)
        item["data_delay_threshold"] = types.Int64Value(int64(*v.DataDelayThreshold))
        item["proxy_agent_id"] = types.StringValue(v.ProxyAgentId)
        item["networking_method"] = types.StringValue(v.NetworkingMethod)
        item["hybrid_deployment_agent_id"] = types.StringValue(v.HybridDeploymentAgentId)
        item["private_link_id"] = types.StringValue(v.PrivateLinkId)

        objectValue, _ := types.ObjectValue(elementAttrType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Connections, _ = types.SetValue(types.ObjectType{AttrTypes: elementAttrType}, items)
}