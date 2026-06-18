package model

import (
	"context"
	"fmt"
	"time"

	gfcommon "github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectionV2ResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ConnectedBy types.String `tfsdk:"connected_by"`
	CreatedAt   types.String `tfsdk:"created_at"`
	GroupId     types.String `tfsdk:"group_id"`
	Service     types.String `tfsdk:"service"`

	Config types.Dynamic `tfsdk:"config"`
	Auth   types.Dynamic `tfsdk:"auth"`

	SucceededAt     types.String `tfsdk:"succeeded_at"`
	FailedAt        types.String `tfsdk:"failed_at"`
	ServiceVersion  types.String `tfsdk:"service_version"`
	SyncFrequency   types.Int64  `tfsdk:"sync_frequency"`
	ScheduleType    types.String `tfsdk:"schedule_type"`
	PauseAfterTrial types.Bool   `tfsdk:"pause_after_trial"`
	DailySyncTime   types.String `tfsdk:"daily_sync_time"`

	ProxyAgentId            types.String `tfsdk:"proxy_agent_id"`
	NetworkingMethod        types.String `tfsdk:"networking_method"`
	HybridDeploymentAgentId types.String `tfsdk:"hybrid_deployment_agent_id"`
	PrivateLinkId           types.String `tfsdk:"private_link_id"`

	DataDelaySensitivity types.String `tfsdk:"data_delay_sensitivity"`
	DataDelayThreshold   types.Int64  `tfsdk:"data_delay_threshold"`

	RunSetupTests     types.Bool `tfsdk:"run_setup_tests"`
	TrustCertificates types.Bool `tfsdk:"trust_certificates"`
	TrustFingerprints types.Bool `tfsdk:"trust_fingerprints"`

	Status types.Object `tfsdk:"status"`
}

func ConnectionV2CodeMessageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"code":    types.StringType,
		"message": types.StringType,
	}
}

func ConnectionV2StatusAttrTypes() map[string]attr.Type {
	codeMessageType := types.ObjectType{
		AttrTypes: ConnectionV2CodeMessageAttrTypes(),
	}

	return map[string]attr.Type{
		"setup_state":        types.StringType,
		"is_historical_sync": types.BoolType,
		"sync_state":         types.StringType,
		"update_state":       types.StringType,
		"tasks":              types.SetType{ElemType: codeMessageType},
		"warnings":           types.SetType{ElemType: codeMessageType},
	}
}

func ConnectionV2ResourceModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                         types.StringType,
		"name":                       types.StringType,
		"connected_by":               types.StringType,
		"created_at":                 types.StringType,
		"group_id":                   types.StringType,
		"service":                    types.StringType,
		"config":                     types.DynamicType,
		"auth":                       types.DynamicType,
		"succeeded_at":               types.StringType,
		"failed_at":                  types.StringType,
		"service_version":            types.StringType,
		"sync_frequency":             types.Int64Type,
		"schedule_type":              types.StringType,
		"pause_after_trial":          types.BoolType,
		"daily_sync_time":            types.StringType,
		"proxy_agent_id":             types.StringType,
		"networking_method":          types.StringType,
		"hybrid_deployment_agent_id": types.StringType,
		"private_link_id":            types.StringType,
		"data_delay_sensitivity":     types.StringType,
		"data_delay_threshold":       types.Int64Type,
		"run_setup_tests":            types.BoolType,
		"trust_certificates":         types.BoolType,
		"trust_fingerprints":         types.BoolType,
		"status":                     types.ObjectType{AttrTypes: ConnectionV2StatusAttrTypes()},
	}
}

func (d *ConnectionV2ResourceModel) ReadFromCreateResponse(ctx context.Context, resp connections.DetailsWithCustomConfigResponse, meta *metadata.ConnectorMetadata, configMask map[string]interface{}) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DetailsResponseDataCommon, resp.Data.Config, meta, configMask)
}

func (d *ConnectionV2ResourceModel) ReadFromResponse(ctx context.Context, resp connections.DetailsWithCustomConfigNoTestsResponse, meta *metadata.ConnectorMetadata, configMask map[string]interface{}) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DetailsResponseDataCommon, resp.Data.Config, meta, configMask)
}

func (d *ConnectionV2ResourceModel) ReadFromResponseForImport(ctx context.Context, resp connections.DetailsWithCustomConfigNoTestsResponse, meta *metadata.ConnectorMetadata) diag.Diagnostics {
	return d.readFromResponseData(ctx, resp.Data.DetailsResponseDataCommon, resp.Data.Config, meta, resp.Data.Config)
}

func (d *ConnectionV2ResourceModel) readFromResponseData(ctx context.Context, data connections.DetailsResponseDataCommon, config map[string]interface{}, meta *metadata.ConnectorMetadata, configMask map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Id = types.StringValue(data.ID)
	d.Name = types.StringValue(data.Schema)
	d.ConnectedBy = stringValueOrNull(data.ConnectedBy)
	d.CreatedAt = timeValueOrNull(data.CreatedAt)
	d.GroupId = types.StringValue(data.GroupID)
	d.Service = types.StringValue(data.Service)

	d.SucceededAt = timeValueOrNull(data.SucceededAt)
	d.FailedAt = timeValueOrNull(data.FailedAt)
	d.ServiceVersion = intPointerStringValue(data.ServiceVersion)
	d.SyncFrequency = intPointerInt64Value(data.SyncFrequency)
	d.ScheduleType = stringValueOrNull(data.ScheduleType)
	d.PauseAfterTrial = boolPointerValue(data.PauseAfterTrial)
	d.DailySyncTime = stringValueOrNull(data.DailySyncTime)

	d.ProxyAgentId = stringValueOrNull(data.ProxyAgentId)
	d.NetworkingMethod = stringValueOrNull(data.NetworkingMethod)
	d.HybridDeploymentAgentId = stringValueOrNull(data.HybridDeploymentAgentId)
	d.PrivateLinkId = stringValueOrNull(data.PrivateLinkId)

	d.DataDelaySensitivity = stringValueOrNull(data.DataDelaySensitivity)
	d.DataDelayThreshold = intPointerInt64Value(data.DataDelayThreshold)
	d.Status = connectionV2StatusValue(data.Status)

	configSlot := (*metadata.Property)(nil)
	if meta != nil {
		configSlot = &meta.Config
	}
	projectedConfig := core.ProjectDynamic(config, configMask, configSlot)
	dynamicConfig, dynamicDiags := core.MapToDynamic(ctx, projectedConfig)
	diags.Append(dynamicDiags...)
	if !diags.HasError() {
		d.Config = dynamicConfig
	}

	return diags
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func timeValueOrNull(value time.Time) types.String {
	if value.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(value.String())
}

func intPointerStringValue(value *int) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(fmt.Sprintf("%v", *value))
}

func intPointerInt64Value(value *int) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}

func boolPointerValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func connectionV2ReadCommonResponse(r gfcommon.CommonResponse) attr.Value {
	result, _ := types.ObjectValue(ConnectionV2CodeMessageAttrTypes(),
		map[string]attr.Value{
			"code":    types.StringValue(r.Code),
			"message": types.StringValue(r.Message),
		})
	return result
}

func connectionV2StatusValue(status connections.StatusResponse) types.Object {
	codeMessageType := types.ObjectType{
		AttrTypes: ConnectionV2CodeMessageAttrTypes(),
	}

	warnings := make([]attr.Value, 0, len(status.Warnings))
	for _, w := range status.Warnings {
		warnings = append(warnings, connectionV2ReadCommonResponse(w))
	}

	tasks := make([]attr.Value, 0, len(status.Tasks))
	for _, t := range status.Tasks {
		tasks = append(tasks, connectionV2ReadCommonResponse(t))
	}

	warningsValue, _ := types.SetValue(codeMessageType, warnings)
	tasksValue, _ := types.SetValue(codeMessageType, tasks)

	result, _ := types.ObjectValue(
		ConnectionV2StatusAttrTypes(),
		map[string]attr.Value{
			"setup_state":        stringValueOrNull(status.SetupState),
			"is_historical_sync": boolPointerValue(status.IsHistoricalSync),
			"sync_state":         stringValueOrNull(status.SyncState),
			"update_state":       stringValueOrNull(status.UpdateState),
			"warnings":           warningsValue,
			"tasks":              tasksValue,
		},
	)

	return result
}
