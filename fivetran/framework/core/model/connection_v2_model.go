package model

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Paused          types.Bool   `tfsdk:"paused"`
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
		"paused":                     types.BoolType,
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
