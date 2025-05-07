package model

import (
    "context"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type GroupConnections struct {
    Id            types.String `tfsdk:"id"`
    Schema        types.String `tfsdk:"schema"`
    Connections   types.Set    `tfsdk:"connections"`
}

var (
    elementConnectionType = map[string]attr.Type{
        "id":                  types.StringType,
        "group_id":            types.StringType,
        "service":             types.StringType,
        "service_version":     types.Int64Type,
        "schema":              types.StringType,
        "connected_by":        types.StringType,
        "created_at":          types.StringType,
        "succeeded_at":        types.StringType,
        "failed_at":           types.StringType,
        "sync_frequency":      types.Int64Type,
        "schedule_type":       types.StringType,
        "daily_sync_time":     types.StringType,
        "status":              types.ObjectType{AttrTypes: elementConnectionStatusType},
    }

    elementConnectionStatusType = map[string]attr.Type{
        "setup_state":         types.StringType,
        "is_historical_sync":  types.BoolType,
        "sync_state":          types.StringType,
        "update_state":        types.StringType,
        "warnings":            types.SetType{ElemType: types.ObjectType{AttrTypes: codeMessageAttrTypes}},
        "tasks":               types.SetType{ElemType: types.ObjectType{AttrTypes: codeMessageAttrTypes}},
    }
)

func (d *GroupConnections) ReadFromResponse(ctx context.Context, resp groups.GroupListConnectionsResponse) {
    if resp.Data.Items == nil {
        d.Connections = types.SetNull(types.ObjectType{AttrTypes: elementConnectionType})
    }

    connections := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        connection := map[string]attr.Value{}
        connection["id"] = types.StringValue(v.ID)
        connection["group_id"] = types.StringValue(v.GroupID)
        connection["service"] = types.StringValue(v.Service)
        connection["service_version"] = types.Int64Value(int64(*v.ServiceVersion))
        connection["schema"] = types.StringValue(v.Schema)
        connection["connected_by"] = types.StringValue(v.ConnectedBy)
        connection["created_at"] = types.StringValue(v.CreatedAt.String())
        connection["succeeded_at"] = types.StringValue(v.SucceededAt.String())
        connection["failed_at"] = types.StringValue(v.FailedAt.String())
        connection["sync_frequency"] = types.Int64Value(int64(*v.SyncFrequency))
        connection["schedule_type"] = types.StringValue(v.ScheduleType)

        if v.DailySyncTime != "" {
            connection["daily_sync_time"] = types.StringValue(v.DailySyncTime)
        } else {
            connection["daily_sync_time"] = types.StringNull()
        }

        warns := []attr.Value{}
        for _, w := range v.Status.Warnings {
            warns = append(warns, readCommonResponse(w))
        }

        tasks := []attr.Value{}
        for _, t := range v.Status.Tasks {
            tasks = append(tasks, readCommonResponse(t))
        }

        wsV, _ := types.SetValue(types.ObjectType{AttrTypes: codeMessageAttrTypes}, warns)
        tsV, _ := types.SetValue(types.ObjectType{AttrTypes: codeMessageAttrTypes}, tasks)

        status, _ := types.ObjectValue(
            elementConnectionStatusType,
            map[string]attr.Value{
                "setup_state":        types.StringValue(v.Status.SetupState),
                "is_historical_sync": types.BoolPointerValue(v.Status.IsHistoricalSync),
                "sync_state":         types.StringValue(v.Status.SyncState),
                "update_state":       types.StringValue(v.Status.UpdateState),
                "warnings":           wsV,
                "tasks":              tsV,
            },
        )

        connection["status"] = status

        objectConnection, _ := types.ObjectValue(elementConnectionType, connection)
        connections = append(connections, objectConnection)
    }

    if d.Id.IsUnknown() || d.Id.IsNull() {
        d.Id = types.StringValue("0")        
    }

    d.Connections, _ = types.SetValue(types.ObjectType{AttrTypes: elementConnectionType}, connections)
}