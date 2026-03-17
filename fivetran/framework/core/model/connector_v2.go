package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConnectorV2ResourceModel is the model for fivetran_connector_v2.
// It is identical to ConnectorResourceModel except destination_schema is a plain
// string (e.g. "schema" or "schema.table") instead of a nested object block.
type ConnectorV2ResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ConnectedBy       types.String `tfsdk:"connected_by"`
	CreatedAt         types.String `tfsdk:"created_at"`
	GroupId           types.String `tfsdk:"group_id"`
	Service           types.String `tfsdk:"service"`
	DestinationSchema types.String `tfsdk:"destination_schema"`

	ProxyAgentId            types.String `tfsdk:"proxy_agent_id"`
	NetworkingMethod        types.String `tfsdk:"networking_method"`
	HybridDeploymentAgentId types.String `tfsdk:"hybrid_deployment_agent_id"`
	PrivateLinkId           types.String `tfsdk:"private_link_id"`

	DataDelaySensitivity types.String `tfsdk:"data_delay_sensitivity"`
	DataDelayThreshold   types.Int64  `tfsdk:"data_delay_threshold"`

	Config   types.Dynamic  `tfsdk:"config"`
	Auth     types.Object   `tfsdk:"auth"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`

	RunSetupTests     types.Bool `tfsdk:"run_setup_tests"`
	TrustCertificates types.Bool `tfsdk:"trust_certificates"`
	TrustFingerprints types.Bool `tfsdk:"trust_fingerprints"`
}

// ParseDestinationSchemaConfigs returns ordered API config candidates derived from
// the destination_schema string value.
//
// For "schema.table":  [{schema, table}, {schema, table_group_name}]
// For "schema":        [{schema}, {schema_prefix}]
//
// The provider tries each candidate in order on Create, stopping on first success.
func (d *ConnectorV2ResourceModel) ParseDestinationSchemaConfigs() ([]map[string]any, error) {
	val := d.DestinationSchema.ValueString()
	if val == "" {
		return nil, fmt.Errorf("destination_schema must not be empty")
	}
	if idx := strings.Index(val, "."); idx != -1 {
		schemaName := val[:idx]
		tableName := val[idx+1:]
		return []map[string]any{
			{"schema": schemaName, "table": tableName},
			{"schema": schemaName, "table_group_name": tableName},
		}, nil
	}
	return []map[string]any{
		{"schema": val},
		{"schema_prefix": val},
	}, nil
}

func (d *ConnectorV2ResourceModel) GetConfigMapFromDynamic(ctx context.Context) (map[string]any, error) {
	if d.Config.IsNull() || d.Config.IsUnknown() {
		return nil, nil
	}
	return dynamicToMap(ctx, d.Config), nil
}

func (d *ConnectorV2ResourceModel) ReadConfigFromResponse(ctx context.Context, remote map[string]any, meta *metadata.ConnectorMetadata, isImporting bool) {
	if d.Config.IsNull() && !isImporting {
		return
	}
	mask := dynamicToMap(ctx, d.Config)
	projected := project(remote, mask, meta)
	d.Config = mapToDynamic(ctx, projected)
}

func (d *ConnectorV2ResourceModel) GetAuthMap(nullOnNull bool) (map[string]any, error) {
	if d.Auth.IsNull() && nullOnNull {
		return nil, nil
	}
	serviceName := d.Service.ValueString()
	serviceFields := common.GetAuthFieldsForService(serviceName)
	allFields := common.GetAuthFieldsMap()
	result := getValueFromAttrValue(d.Auth, allFields, nil, serviceName).(map[string]any)
	err := patchServiceSpecificFields(result, serviceName, serviceFields, allFields)
	return result, err
}

func (d *ConnectorV2ResourceModel) ReadFromCreateResponse(resp connections.DetailsWithCustomConfigResponse) {
	c := ConnectorModelContainer{}
	c.ReadFromResponseData(resp.Data.DetailsResponseDataCommon, resp.Data.Config)
	d.readFromContainer(c)
}

func (d *ConnectorV2ResourceModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse, isImporting bool) {
	c := ConnectorModelContainer{}
	c.ReadFromResponseData(resp.Data.DetailsResponseDataCommon, resp.Data.Config)
	d.readFromContainer(c)
}

func (d *ConnectorV2ResourceModel) HasUpdates(ctx context.Context, plan ConnectorV2ResourceModel, state ConnectorV2ResourceModel, meta *metadata.ConnectorMetadata) (bool, map[string]any, map[string]any, error) {
	stateConfigMap, err := state.GetConfigMapFromDynamic(ctx)
	if err != nil {
		return false, nil, nil, err
	}
	planConfigMap, err := plan.GetConfigMapFromDynamic(ctx)
	if err != nil {
		return false, nil, nil, err
	}
	stateAuthMap, err := state.GetAuthMap(false)
	if err != nil {
		return false, nil, nil, err
	}
	planAuthMap, err := plan.GetAuthMap(false)
	if err != nil {
		return false, nil, nil, err
	}

	patch := PrepareConfigPatchDynamic(stateConfigMap, planConfigMap, meta)
	authPatch := PrepareConfigAuthPatch(stateAuthMap, planAuthMap, plan.Service.ValueString(), common.GetAuthFieldsMap())

	if len(patch) > 0 ||
		len(authPatch) > 0 ||
		!plan.ProxyAgentId.Equal(state.ProxyAgentId) ||
		!plan.PrivateLinkId.Equal(state.PrivateLinkId) ||
		!plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) ||
		!plan.DataDelaySensitivity.Equal(state.DataDelaySensitivity) ||
		!plan.DataDelayThreshold.Equal(state.DataDelayThreshold) ||
		!plan.NetworkingMethod.Equal(state.NetworkingMethod) {
		return true, patch, authPatch, nil
	}
	return false, nil, nil, nil
}

func (d *ConnectorV2ResourceModel) readFromContainer(c ConnectorModelContainer) {
	d.Id = types.StringValue(c.Id)
	d.Name = types.StringValue(c.Schema)
	d.ConnectedBy = types.StringValue(c.ConnectedBy)
	d.CreatedAt = types.StringValue(c.CreatedAt)
	d.GroupId = types.StringValue(c.GroupId)
	d.Service = types.StringValue(c.Service)

	if !d.DataDelaySensitivity.IsUnknown() && !d.DataDelaySensitivity.IsNull() {
		d.DataDelaySensitivity = types.StringValue(c.DataDelaySensitivity)
	}

	if c.DataDelayThreshold != nil {
		d.DataDelayThreshold = types.Int64Value(int64(*c.DataDelayThreshold))
	} else {
		d.DataDelayThreshold = types.Int64Null()
	}

	// destination_schema preserves the user's original plan value (e.g. "schema.table").
	// The API only returns the schema portion, so we do not overwrite to avoid permanent drift.
	// On import (unknown/null) we fall back to the API's schema string.
	if d.DestinationSchema.IsUnknown() || d.DestinationSchema.IsNull() {
		d.DestinationSchema = types.StringValue(c.Schema)
	}

	if c.HybridDeploymentAgentId != "" && !d.HybridDeploymentAgentId.IsUnknown() && !d.HybridDeploymentAgentId.IsNull() {
		d.HybridDeploymentAgentId = types.StringValue(c.HybridDeploymentAgentId)
	} else {
		d.HybridDeploymentAgentId = types.StringNull()
	}

	if c.PrivateLinkId != "" {
		d.PrivateLinkId = types.StringValue(c.PrivateLinkId)
	} else {
		d.PrivateLinkId = types.StringNull()
	}

	if c.NetworkingMethod != "" {
		d.NetworkingMethod = types.StringValue(c.NetworkingMethod)
	} else {
		d.NetworkingMethod = types.StringNull()
	}

	if c.ProxyAgentId != "" {
		d.ProxyAgentId = types.StringValue(c.ProxyAgentId)
	} else {
		d.ProxyAgentId = types.StringNull()
	}
}
