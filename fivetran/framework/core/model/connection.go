package model

import (
    "fmt"
    //"strings"

    //gfcommon "github.com/fivetran/go-fivetran/common"
    "github.com/fivetran/go-fivetran/connections"
    //"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
    //"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    //"github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/types"
    //"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectionDatasourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    ConnectedBy types.String `tfsdk:"connected_by"`
    CreatedAt   types.String `tfsdk:"created_at"`
    GroupId     types.String `tfsdk:"group_id"`
    Service     types.String `tfsdk:"service"`

    DestinationSchema types.Object `tfsdk:"destination_schema"`

    SucceededAt     types.String `tfsdk:"succeeded_at"`
    FailedAt        types.String `tfsdk:"failed_at"`
    ServiceVersion  types.String `tfsdk:"service_version"`
    SyncFrequency   types.Int64  `tfsdk:"sync_frequency"`
    ScheduleType    types.String `tfsdk:"schedule_type"`
    Paused          types.Bool   `tfsdk:"paused"`
    PauseAfterTrial types.Bool   `tfsdk:"pause_after_trial"`
    DailySyncTime   types.String `tfsdk:"daily_sync_time"`
    
    DataDelaySensitivity    types.String `tfsdk:"data_delay_sensitivity"`
    DataDelayThreshold      types.Int64  `tfsdk:"data_delay_threshold"`

    ProxyAgentId             types.String `tfsdk:"proxy_agent_id"`
    NetworkingMethod         types.String `tfsdk:"networking_method"`
    HybridDeploymentAgentId  types.String `tfsdk:"hybrid_deployment_agent_id"`
    PrivateLinkId            types.String `tfsdk:"private_link_id"`
    Status types.Object `tfsdk:"status"`
}

func (d *ConnectionDatasourceModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse) {
    d.Id = types.StringValue(resp.Data.DetailsResponseDataCommon.ID)
    d.Name = types.StringValue(resp.Data.DetailsResponseDataCommon.Schema)
    d.ConnectedBy = types.StringValue(resp.Data.DetailsResponseDataCommon.ConnectedBy)
    d.CreatedAt = types.StringValue(resp.Data.DetailsResponseDataCommon.CreatedAt.String())
    d.GroupId = types.StringValue(resp.Data.DetailsResponseDataCommon.GroupID)
    d.Service = types.StringValue(resp.Data.DetailsResponseDataCommon.Service)
    d.SucceededAt = types.StringValue(resp.Data.SucceededAt.String())
    d.FailedAt = types.StringValue(resp.Data.FailedAt.String())
    d.ServiceVersion = types.StringValue(fmt.Sprintf("%v", *resp.Data.ServiceVersion))
    d.SyncFrequency = types.Int64Value(int64(*resp.Data.SyncFrequency))
    d.ScheduleType = types.StringValue(resp.Data.ScheduleType)
    d.Paused = types.BoolValue(*resp.Data.Paused)
    d.PauseAfterTrial = types.BoolValue(*resp.Data.PauseAfterTrial)
    
    d.DataDelaySensitivity = types.StringValue(resp.Data.DataDelaySensitivity)

    d.DestinationSchema = getDestinationSchemaValue(resp.Data.DetailsResponseDataCommon.Service, resp.Data.DetailsResponseDataCommon.Schema, d.DestinationSchema)

    if resp.Data.DetailsResponseDataCommon.ProxyAgentId != "" {
        d.ProxyAgentId = types.StringValue(resp.Data.DetailsResponseDataCommon.ProxyAgentId)
    }

    if resp.Data.DetailsResponseDataCommon.NetworkingMethod != "" {
        d.NetworkingMethod = types.StringValue(resp.Data.DetailsResponseDataCommon.NetworkingMethod)
    }

    if resp.Data.DetailsResponseDataCommon.PrivateLinkId != "" {
        d.PrivateLinkId = types.StringValue(resp.Data.DetailsResponseDataCommon.PrivateLinkId)
    }

    if resp.Data.DetailsResponseDataCommon.HybridDeploymentAgentId != "" {
        d.HybridDeploymentAgentId = types.StringValue(resp.Data.DetailsResponseDataCommon.HybridDeploymentAgentId)
    }

    if resp.Data.DataDelayThreshold != nil {
        d.DataDelayThreshold = types.Int64Value(int64(*resp.Data.DataDelayThreshold))
    } else {
        d.DataDelayThreshold = types.Int64Null()
    }

    if resp.Data.DailySyncTime != "" {
        d.DailySyncTime = types.StringValue(resp.Data.DailySyncTime)
    } else {
        d.DailySyncTime = types.StringNull()
    }

    codeMessageAttrType := types.ObjectType{
        AttrTypes: codeMessageAttrTypes,
    }

    warns := []attr.Value{}
    for _, w := range resp.Data.Status.Warnings {
        warns = append(warns, readCommonResponse(w))
    }
    tasks := []attr.Value{}
    for _, t := range resp.Data.Status.Tasks {
        tasks = append(tasks, readCommonResponse(t))
    }

    wsV, _ := types.SetValue(codeMessageAttrType, warns)
    tsV, _ := types.SetValue(codeMessageAttrType, tasks)

    status, _ := types.ObjectValue(
        map[string]attr.Type{
            "setup_state":        types.StringType,
            "is_historical_sync": types.BoolType,
            "sync_state":         types.StringType,
            "update_state":       types.StringType,
            "tasks":              types.SetType{ElemType: codeMessageAttrType},
            "warnings":           types.SetType{ElemType: codeMessageAttrType},
        },
        map[string]attr.Value{
            "setup_state":        types.StringValue(resp.Data.Status.SetupState),
            "is_historical_sync": types.BoolPointerValue(resp.Data.Status.IsHistoricalSync),
            "sync_state":         types.StringValue(resp.Data.Status.SyncState),
            "update_state":       types.StringValue(resp.Data.Status.UpdateState),
            "warnings":           wsV,
            "tasks":              tsV,
        },
    )
    d.Status = status
}