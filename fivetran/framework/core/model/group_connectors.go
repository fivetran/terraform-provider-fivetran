package model

import (
    "context"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type GroupConnectors struct {
    Id           types.String `tfsdk:"id"`
    Schema       types.String `tfsdk:"schema"`
    Connectors   types.Set    `tfsdk:"connectors"`
}

var (
    elementConnectorType = map[string]attr.Type{
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
        "status":              types.ObjectType{AttrTypes: elementConnectorStatusType},
    }

    elementConnectorStatusType = map[string]attr.Type{
        "setup_state":         types.StringType,
        "is_historical_sync":  types.BoolType,
        "sync_state":          types.StringType,
        "update_state":        types.StringType,
        "warnings":            types.SetType{ElemType: types.ObjectType{AttrTypes: codeMessageAttrTypes}},
        "tasks":               types.SetType{ElemType: types.ObjectType{AttrTypes: codeMessageAttrTypes}},
    }
)

func (d *GroupConnectors) ReadFromResponse(ctx context.Context, resp groups.GroupListConnectionsResponse) {
    if resp.Data.Items == nil {
        d.Connectors = types.SetNull(types.ObjectType{AttrTypes: elementConnectorType})
    }

    connectors := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        connector := map[string]attr.Value{}
        connector["id"] = types.StringValue(v.ID)
        connector["group_id"] = types.StringValue(v.GroupID)
        connector["service"] = types.StringValue(v.Service)
        connector["service_version"] = types.Int64Value(int64(*v.ServiceVersion))
        connector["schema"] = types.StringValue(v.Schema)
        connector["connected_by"] = types.StringValue(v.ConnectedBy)
        connector["created_at"] = types.StringValue(v.CreatedAt.String())
        connector["succeeded_at"] = types.StringValue(v.SucceededAt.String())
        connector["failed_at"] = types.StringValue(v.FailedAt.String())
        connector["sync_frequency"] = types.Int64Value(int64(*v.SyncFrequency))
        connector["schedule_type"] = types.StringValue(v.ScheduleType)

        if v.DailySyncTime != "" {
            connector["daily_sync_time"] = types.StringValue(v.DailySyncTime)
        } else {
            connector["daily_sync_time"] = types.StringNull()
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
            elementConnectorStatusType,
            map[string]attr.Value{
                "setup_state":        types.StringValue(v.Status.SetupState),
                "is_historical_sync": types.BoolPointerValue(v.Status.IsHistoricalSync),
                "sync_state":         types.StringValue(v.Status.SyncState),
                "update_state":       types.StringValue(v.Status.UpdateState),
                "warnings":           wsV,
                "tasks":              tsV,
            },
        )

        connector["status"] = status

        objectConnector, _ := types.ObjectValue(elementConnectorType, connector)
        connectors = append(connectors, objectConnector)
    }

    if d.Id.IsUnknown() || d.Id.IsNull() {
        d.Id = types.StringValue("0")        
    }

    d.Connectors, _ = types.SetValue(types.ObjectType{AttrTypes: elementConnectorType}, connectors)
}