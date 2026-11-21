package model

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConnectionV2ResourceModel is the Terraform state shape for fivetran_connection_v2.
type ConnectionV2ResourceModel struct {
	// ForceNew — changing any of these destroys and recreates the resource.
	Id                types.String `tfsdk:"id"`
	Service           types.String `tfsdk:"service"`
	GroupId           types.String `tfsdk:"group_id"`
	DestinationSchema types.String `tfsdk:"destination_schema"`

	// Dynamic config slots — shape resolved at plan time from connector metadata.
	Config types.Dynamic `tfsdk:"config"`
	Auth   types.Dynamic `tfsdk:"auth"`

	// Computed-only root attributes.
	Name           types.String `tfsdk:"name"`
	ConnectedBy    types.String `tfsdk:"connected_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	SucceededAt    types.String `tfsdk:"succeeded_at"`
	FailedAt       types.String `tfsdk:"failed_at"`
	ServiceVersion types.String `tfsdk:"service_version"`
	Status         types.Object `tfsdk:"status"`

	// Optional + Computed root attributes.
	Paused                  types.Bool   `tfsdk:"paused"`
	SyncFrequency           types.Int64  `tfsdk:"sync_frequency"`
	ScheduleType            types.String `tfsdk:"schedule_type"`
	DailySyncTime           types.String `tfsdk:"daily_sync_time"`
	PauseAfterTrial         types.Bool   `tfsdk:"pause_after_trial"`
	NetworkingMethod        types.String `tfsdk:"networking_method"`
	ProxyAgentId            types.String `tfsdk:"proxy_agent_id"`
	PrivateLinkId           types.String `tfsdk:"private_link_id"`
	HybridDeploymentAgentId types.String `tfsdk:"hybrid_deployment_agent_id"`
	DataDelaySensitivity    types.String `tfsdk:"data_delay_sensitivity"`
	DataDelayThreshold      types.Int64  `tfsdk:"data_delay_threshold"`

	// Plan-only — not round-tripped from the API.
	RunSetupTests     types.Bool `tfsdk:"run_setup_tests"`
	TrustCertificates types.Bool `tfsdk:"trust_certificates"`
	TrustFingerprints types.Bool `tfsdk:"trust_fingerprints"`
}

// StatusAttrTypes is the attr.Type map for the status nested object.
// Used when constructing null/unknown status values and reading from API responses.
var StatusAttrTypes = map[string]attr.Type{
	"setup_state":        types.StringType,
	"sync_state":         types.StringType,
	"update_state":       types.StringType,
	"is_historical_sync": types.BoolType,
	"tasks": types.SetType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
		"code":    types.StringType,
		"message": types.StringType,
	}}},
	"warnings": types.SetType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
		"code":    types.StringType,
		"message": types.StringType,
	}}},
}
