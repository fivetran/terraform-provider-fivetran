package resources

import (
	"context"
	"fmt"
	"sync"

	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConnectionV2() resource.Resource {
	return &connectionV2{}
}

type connectionV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionV2{}
var _ resource.ResourceWithImportState = &connectionV2{}

func (r *connectionV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_v2"
}

func (r *connectionV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ConnectionV2ResourceSchema()
}

func (r *connectionV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	details, err := r.GetClient().NewConnectionDetails().ConnectionID(req.ID).DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Connection V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	meta, err := r.connectorMetadata(ctx, details.Data.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Connection Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", details.Data.Service, err),
		)
		return
	}

	data := model.ConnectionV2ResourceModel{
		Auth:              types.DynamicNull(),
		RunSetupTests:     types.BoolValue(false),
		TrustCertificates: types.BoolValue(false),
		TrustFingerprints: types.BoolValue(false),
	}

	resp.Diagnostics.Append(data.ReadFromResponseForImport(ctx, details, meta)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"fivetran_connection_v2 import requires HCL review",
		"Terraform imported the connection using API-visible config fields. Add the matching fivetran_connection_v2 resource block to your configuration and restore any sensitive auth values because the API does not return auth secrets.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.connectorMetadata(ctx, data.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Connection Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", data.Service.ValueString(), err),
		)
		return
	}

	configMap, authMap := r.dynamicPlanMaps(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, false)

	svc := r.GetClient().NewConnectionCreate().
		Paused(true).
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan)

	if configMap != nil {
		svc.ConfigCustom(&configMap)
	}
	if authMap != nil {
		svc.AuthCustom(&authMap)
	}

	r.applyCreateRootFields(svc, data)

	response, err := svc.DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	resp.Diagnostics.Append(data.ReadFromCreateResponse(ctx, response, meta, configMap)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Auth = preserveDynamic(data.Auth)
	data.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	data.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	data.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	r.warnFailedSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.GetClient().NewConnectionDetails().ConnectionID(data.Id.ValueString()).DoCustom(ctx)
	if err != nil {
		if response.Code == "NotFound_Connector" || response.Code == "NotFound_Connection" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Connection V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	meta, err := r.connectorMetadata(ctx, response.Data.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Connection Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", response.Data.Service, err),
		)
		return
	}

	configMask, diags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	auth := data.Auth

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMask)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Auth = preserveDynamic(auth)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.connectorMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Connection Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", plan.Service.ValueString(), err),
		)
		return
	}

	planConfig, planAuth := r.dynamicPlanMaps(ctx, plan, &resp.Diagnostics)
	stateConfig, stateAuth := r.dynamicStateMaps(ctx, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	configPatch := core.PrepareConfigPatchDynamic(planConfig, stateConfig, &meta.Config)
	authPatch := core.PrepareConfigPatchDynamic(planAuth, stateAuth, &meta.Auth)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, false)

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(state.Id.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan)

	if len(configPatch) > 0 {
		svc.ConfigCustom(&configPatch)
	}
	if len(authPatch) > 0 {
		svc.AuthCustom(&authPatch)
	}

	r.applyUpdateRootFields(svc, plan, state)

	response, err := svc.DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	r.warnFailedSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	details, err := r.GetClient().NewConnectionDetails().ConnectionID(state.Id.ValueString()).DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connection V2 Resource After Update.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(plan.ReadFromResponse(ctx, details, meta, planConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Auth = preserveDynamic(plan.Auth)
	plan.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	plan.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	plan.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewConnectionDelete().ConnectionID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connection V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}

func (r *connectionV2) connectorMetadata(ctx context.Context, service string) (*metadata.ConnectorMetadata, error) {
	cache := r.GetMetadataCache()
	if cache == nil {
		cache = &sync.Map{}
	}
	if r.GetClient() == nil {
		if meta, ok, err := core.LoadCachedConnectorMetadata(cache, service); ok || err != nil {
			return meta, err
		}
		return nil, fmt.Errorf("unconfigured Fivetran client")
	}
	return core.GetCachedConnectorMetadata(ctx, r.GetClient(), cache, service)
}

func (r *connectionV2) dynamicPlanMaps(ctx context.Context, data model.ConnectionV2ResourceModel, diags *diag.Diagnostics) (map[string]interface{}, map[string]interface{}) {
	configMap, configDiags := core.DynamicToMap(ctx, data.Config)
	diags.Append(configDiags...)

	authMap, authDiags := core.DynamicToMap(ctx, data.Auth)
	diags.Append(authDiags...)

	return configMap, authMap
}

func (r *connectionV2) dynamicStateMaps(ctx context.Context, data model.ConnectionV2ResourceModel, diags *diag.Diagnostics) (map[string]interface{}, map[string]interface{}) {
	return r.dynamicPlanMaps(ctx, data, diags)
}

func (r *connectionV2) applyCreateRootFields(svc *connections.ConnectionCreateService, data model.ConnectionV2ResourceModel) {
	if !data.SyncFrequency.IsNull() && !data.SyncFrequency.IsUnknown() {
		value := int(data.SyncFrequency.ValueInt64())
		svc.SyncFrequency(&value)
	}
	if !data.DailySyncTime.IsNull() && !data.DailySyncTime.IsUnknown() {
		svc.DailySyncTime(data.DailySyncTime.ValueString())
	}
	if !data.PauseAfterTrial.IsNull() && !data.PauseAfterTrial.IsUnknown() {
		svc.PauseAfterTrial(data.PauseAfterTrial.ValueBool())
	}
	if !data.ProxyAgentId.IsNull() && !data.ProxyAgentId.IsUnknown() {
		svc.ProxyAgentId(data.ProxyAgentId.ValueString())
	}
	if !data.NetworkingMethod.IsNull() && !data.NetworkingMethod.IsUnknown() {
		svc.NetworkingMethod(data.NetworkingMethod.ValueString())
	}
	if !data.PrivateLinkId.IsNull() && !data.PrivateLinkId.IsUnknown() {
		svc.PrivateLinkId(data.PrivateLinkId.ValueString())
	}
	if !data.HybridDeploymentAgentId.IsNull() && !data.HybridDeploymentAgentId.IsUnknown() {
		svc.HybridDeploymentAgentId(data.HybridDeploymentAgentId.ValueString())
	}
	if !data.DataDelaySensitivity.IsNull() && !data.DataDelaySensitivity.IsUnknown() {
		svc.DataDelaySensitivity(data.DataDelaySensitivity.ValueString())
	}
	if !data.DataDelayThreshold.IsNull() && !data.DataDelayThreshold.IsUnknown() {
		value := int(data.DataDelayThreshold.ValueInt64())
		svc.DataDelayThreshold(&value)
	}
}

func (r *connectionV2) applyUpdateRootFields(svc *connections.ConnectionUpdateService, plan, state model.ConnectionV2ResourceModel) {
	if !plan.SyncFrequency.Equal(state.SyncFrequency) && !plan.SyncFrequency.IsNull() && !plan.SyncFrequency.IsUnknown() {
		value := int(plan.SyncFrequency.ValueInt64())
		svc.SyncFrequency(&value)
	}
	if !plan.ScheduleType.Equal(state.ScheduleType) && !plan.ScheduleType.IsNull() && !plan.ScheduleType.IsUnknown() {
		svc.ScheduleType(plan.ScheduleType.ValueString())
	}
	if !plan.DailySyncTime.Equal(state.DailySyncTime) && !plan.DailySyncTime.IsNull() && !plan.DailySyncTime.IsUnknown() {
		svc.DailySyncTime(plan.DailySyncTime.ValueString())
	}
	if !plan.PauseAfterTrial.Equal(state.PauseAfterTrial) && !plan.PauseAfterTrial.IsNull() && !plan.PauseAfterTrial.IsUnknown() {
		svc.PauseAfterTrial(plan.PauseAfterTrial.ValueBool())
	}
	if !plan.ProxyAgentId.Equal(state.ProxyAgentId) && !plan.ProxyAgentId.IsNull() && !plan.ProxyAgentId.IsUnknown() {
		svc.ProxyAgentId(plan.ProxyAgentId.ValueString())
	}
	if !plan.NetworkingMethod.Equal(state.NetworkingMethod) && !plan.NetworkingMethod.IsNull() && !plan.NetworkingMethod.IsUnknown() {
		svc.NetworkingMethod(plan.NetworkingMethod.ValueString())
	}
	if !plan.PrivateLinkId.Equal(state.PrivateLinkId) && !plan.PrivateLinkId.IsNull() && !plan.PrivateLinkId.IsUnknown() {
		svc.PrivateLinkId(plan.PrivateLinkId.ValueString())
	}
	if !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) && !plan.HybridDeploymentAgentId.IsNull() && !plan.HybridDeploymentAgentId.IsUnknown() {
		svc.HybridDeploymentAgentId(plan.HybridDeploymentAgentId.ValueString())
	}
	if !plan.DataDelaySensitivity.Equal(state.DataDelaySensitivity) && !plan.DataDelaySensitivity.IsNull() && !plan.DataDelaySensitivity.IsUnknown() {
		svc.DataDelaySensitivity(plan.DataDelaySensitivity.ValueString())
	}
	if !plan.DataDelayThreshold.Equal(state.DataDelayThreshold) && !plan.DataDelayThreshold.IsNull() && !plan.DataDelayThreshold.IsUnknown() {
		value := int(plan.DataDelayThreshold.ValueInt64())
		svc.DataDelayThreshold(&value)
	}
}

func (r *connectionV2) warnFailedSetupTests(setupTests []common.SetupTestResponse, diags *diag.Diagnostics) {
	for _, tr := range setupTests {
		if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
			diags.AddWarning(
				fmt.Sprintf("Connection setup test `%v` has status `%v`", tr.Title, tr.Status),
				tr.Message,
			)
		}
	}
}

func preserveDynamic(value types.Dynamic) types.Dynamic {
	if value.IsUnknown() {
		return types.DynamicNull()
	}
	return value
}
