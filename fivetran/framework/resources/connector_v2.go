package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConnectorV2() resource.Resource {
	return &connectorV2{}
}

type connectorV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectorV2{}
var _ resource.ResourceWithImportState = &connectorV2{}
var _ resource.ResourceWithValidateConfig = &connectorV2{}

func (r *connectorV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_v2"
}

func (r *connectorV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := fivetranSchema.ConnectorAttributesSchema().GetResourceSchema()
	attrs["config"] = fivetranSchema.ConnectorConfigDynamicAttribute()
	attrs["destination_schema"] = fivetranSchema.ConnectorDestinationSchemaAttribute()
	resp.Schema = schema.Schema{
		Attributes: attrs,
		Blocks:     fivetranSchema.ConnectorV2ResourceBlocks(ctx),
		Version:    0,
	}
}

func (r *connectorV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data model.ConnectorV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := data.Service.ValueString()
	if service == "" || data.Config.IsNull() || data.Config.IsUnknown() {
		return
	}

	if r.GetClient() == nil {
		return
	}

	meta, err := core.GetCachedConnectorMetadata(ctx, r.GetClient(), service)
	if err != nil {
		return
	}

	configMap := model.DynamicToMapPublic(ctx, data.Config)
	for key := range configMap {
		if _, ok := meta.Config.Properties[key]; !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("config"),
				"Unknown config field",
				fmt.Sprintf("Field %q is not valid for service %q.", key, service),
			)
		}
	}
}

func (r *connectorV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectorV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data model.ConnectorV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseConfigMap, err := data.GetConfigMapFromDynamic(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Connector Resource.", fmt.Sprintf("%v;", err))
		return
	}
	if baseConfigMap == nil {
		baseConfigMap = make(map[string]any)
	}

	authMap, err := data.GetAuthMap(true)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Connector Resource.", fmt.Sprintf("%v;", err))
		return
	}
	noAuth := authMap == nil

	schemaConfigs, err := data.ParseDestinationSchemaConfigs()
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Connector Resource.", err.Error())
		return
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, false)

	// Try each destination schema config candidate; stop on first success.
	var lastErr error
	var response connections.DetailsWithCustomConfigResponse
	for _, schemaFields := range schemaConfigs {
		configMap := make(map[string]any)
		for k, v := range baseConfigMap {
			configMap[k] = v
		}
		for k, v := range schemaFields {
			configMap[k] = v
		}

		svc := r.GetClient().NewConnectionCreate().
			Paused(true).
			Service(data.Service.ValueString()).
			GroupID(data.GroupId.ValueString()).
			RunSetupTests(runSetupTestsPlan).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan).
			ConfigCustom(&configMap)

		if data.ProxyAgentId.ValueString() != "" {
			svc.ProxyAgentId(data.ProxyAgentId.ValueString())
		}
		if data.NetworkingMethod.ValueString() != "" {
			svc.NetworkingMethod(data.NetworkingMethod.ValueString())
		}
		if data.PrivateLinkId.ValueString() != "" {
			svc.PrivateLinkId(data.PrivateLinkId.ValueString())
		}
		if data.DataDelaySensitivity.ValueString() != "" {
			svc.DataDelaySensitivity(data.DataDelaySensitivity.ValueString())
		}
		if !data.DataDelayThreshold.IsNull() {
			value := int(data.DataDelayThreshold.ValueInt64())
			svc.DataDelayThreshold(&value)
		}
		if data.HybridDeploymentAgentId.ValueString() != "" {
			svc.HybridDeploymentAgentId(data.HybridDeploymentAgentId.ValueString())
		}
		if !noAuth {
			svc.AuthCustom(&authMap)
		}

		response, lastErr = svc.DoCustom(ctx)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Resource.",
			fmt.Sprintf("All destination_schema configurations failed. Last error: %v; code: %v; message: %v", lastErr, response.Code, response.Message),
		)
		return
	}

	data.ReadFromCreateResponse(response)

	data.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	data.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	data.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	if runSetupTestsPlan && response.Data.SetupTests != nil && len(response.Data.SetupTests) > 0 {
		for _, tr := range response.Data.SetupTests {
			if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
				resp.Diagnostics.AddWarning(
					fmt.Sprintf("Connector setup test `%v` has status `%v`", tr.Title, tr.Status),
					tr.Message,
				)
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data model.ConnectorV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	isImportOperation := data.GroupId.IsNull() || data.GroupId.IsUnknown() || data.Service.IsNull() || data.Service.IsUnknown()

	id := data.Id.ValueString()

	runSetupTests := data.RunSetupTests
	trustCertificates := data.TrustCertificates
	trustFingerprints := data.TrustFingerprints

	response, err := r.GetClient().NewConnectionDetails().ConnectionID(id).DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	data.ReadFromResponse(response, isImportOperation)

	meta, metaErr := core.GetCachedConnectorMetadata(ctx, r.GetClient(), data.Service.ValueString())
	if metaErr == nil {
		data.ReadConfigFromResponse(ctx, response.Data.Config, meta, isImportOperation)
	}

	data.RunSetupTests = runSetupTests
	data.TrustCertificates = trustCertificates
	data.TrustFingerprints = trustFingerprints

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state model.ConnectorV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, false)

	runSetupTestsState := core.GetBoolOrDefault(state.RunSetupTests, false)
	trustCertificatesState := core.GetBoolOrDefault(state.TrustCertificates, false)
	trustFingerprintsState := core.GetBoolOrDefault(state.TrustFingerprints, false)

	planOnlyAttributesChanged := (runSetupTestsPlan && runSetupTestsPlan != runSetupTestsState) ||
		(trustCertificatesPlan && trustCertificatesPlan != trustCertificatesState) ||
		(trustFingerprintsPlan && trustFingerprintsPlan != trustFingerprintsState)

	meta, err := core.GetCachedConnectorMetadata(ctx, r.GetClient(), plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to fetch connector metadata.", fmt.Sprintf("%v;", err))
		return
	}

	hasUpdates, patch, authPatch, err := plan.HasUpdates(ctx, plan, state, meta)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Connector Resource.", fmt.Sprintf("%v; ", err))
	}

	updatePerformed := false
	if hasUpdates {
		svc := r.GetClient().NewConnectionUpdate().
			RunSetupTests(false).
			ConnectionID(state.Id.ValueString())

		if !plan.PrivateLinkId.Equal(state.PrivateLinkId) {
			svc.PrivateLinkId(plan.PrivateLinkId.ValueString())
		}
		if !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) {
			svc.HybridDeploymentAgentId(plan.HybridDeploymentAgentId.ValueString())
		}
		if len(patch) > 0 {
			svc.ConfigCustom(&patch)
		}
		if len(authPatch) > 0 {
			svc.AuthCustom(&authPatch)
		}
		if !plan.ProxyAgentId.Equal(state.ProxyAgentId) && !plan.ProxyAgentId.IsNull() {
			svc.ProxyAgentId(plan.ProxyAgentId.ValueString())
		}
		if !plan.NetworkingMethod.Equal(state.NetworkingMethod) && plan.NetworkingMethod.ValueString() != "" {
			svc.NetworkingMethod(plan.NetworkingMethod.ValueString())
		}
		if !plan.DataDelaySensitivity.Equal(state.DataDelaySensitivity) {
			svc.DataDelaySensitivity(plan.DataDelaySensitivity.ValueString())
		}
		if !plan.DataDelayThreshold.IsNull() {
			value := int(plan.DataDelayThreshold.ValueInt64())
			svc.DataDelayThreshold(&value)
		}

		updateResponse, err := svc.DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
			)
			return
		}
		plan.ReadFromCreateResponse(updateResponse)
		plan.RunSetupTests = types.BoolValue(runSetupTestsPlan)
		plan.TrustCertificates = types.BoolValue(trustCertificatesPlan)
		plan.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)
		updatePerformed = true
	}

	if planOnlyAttributesChanged || (updatePerformed && runSetupTestsPlan) {
		testResponse, err := r.GetClient().NewConnectionSetupTests().
			ConnectionID(state.Id.ValueString()).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan).
			DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, testResponse.Code, testResponse.Message),
			)
			return
		}
		if testResponse.Data.SetupTests != nil {
			for _, tr := range testResponse.Data.SetupTests {
				if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
					resp.Diagnostics.AddWarning(
						fmt.Sprintf("Connector setup test `%v` has status `%v`", tr.Title, tr.Status),
						tr.Message,
					)
				}
			}
		}
		if !updatePerformed {
			plan.ReadFromCreateResponse(testResponse)
			plan.RunSetupTests = types.BoolValue(runSetupTestsPlan)
			plan.TrustCertificates = types.BoolValue(trustCertificatesPlan)
			plan.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)
		}
		updatePerformed = true
	}

	if !updatePerformed {
		readResponse, err := r.GetClient().NewConnectionDetails().ConnectionID(state.Id.ValueString()).DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read after Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, readResponse.Code, readResponse.Message),
			)
			return
		}
		plan.ReadFromResponse(readResponse, false)
		plan.RunSetupTests = types.BoolValue(runSetupTestsPlan)
		plan.TrustCertificates = types.BoolValue(trustCertificatesPlan)
		plan.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)
	}

	if plan.RunSetupTests.IsUnknown() {
		plan.RunSetupTests = state.RunSetupTests
	}
	if plan.TrustCertificates.IsUnknown() {
		plan.TrustCertificates = state.TrustCertificates
	}
	if plan.TrustFingerprints.IsUnknown() {
		plan.TrustFingerprints = state.TrustFingerprints
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data model.ConnectorV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewConnectionDelete().ConnectionID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connector Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
	}
}
