package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Connection() resource.Resource {
	return &connection{}
}

type connection struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connection{}
var _ resource.ResourceWithImportState = &connection{}

func (r *connection) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *connection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.ConnectionAttributesSchema().GetResourceSchema(),
		Blocks:     fivetranSchema.ConnectionResourceBlocks(),
		Version:    1,
	}
}

func (r *connection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}


	destinationSchema, err := data.GetDestinatonSchemaForConfig()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Resource.",
			fmt.Sprintf("%v;", err),
		)

		return
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, false)

	svc := r.GetClient().NewConnectionCreate().
		Paused(true). // on creation we always create paused connection
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		ConfigCustom(&destinationSchema)

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

	response, err := svc.
		DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
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
					fmt.Sprintf("Connection setup test `%v` has status `%v`", tr.Title, tr.Status),
					tr.Message,
				)
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.GetClient().NewConnectionDetails().ConnectionID(data.Id.ValueString()).DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	data.ReadFromResponse(response)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, false)

	svc := r.GetClient().NewConnectionUpdate().
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		ConnectionID(state.Id.ValueString())

	if !plan.PrivateLinkId.Equal(state.PrivateLinkId) {
		svc.PrivateLinkId(plan.PrivateLinkId.ValueString())
	}

	if !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) {
		svc.HybridDeploymentAgentId(plan.HybridDeploymentAgentId.ValueString())
	}

	if !plan.ProxyAgentId.Equal(state.ProxyAgentId) {
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

	response, err := svc.DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}
	plan.ReadFromCreateResponse(response)

	if runSetupTestsPlan && response.Data.SetupTests != nil && len(response.Data.SetupTests) > 0 {
		for _, tr := range response.Data.SetupTests {
			if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
				resp.Diagnostics.AddWarning(
					fmt.Sprintf("Connection setup test `%v` has status `%v`", tr.Title, tr.Status),
					tr.Message,
				)
			}
		}
	}

	details, err := r.GetClient().NewConnectionDetails().ConnectionID(state.Id.ValueString()).DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read after Update Connection Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}
	plan.ReadFromResponse(details)

	// Set up synthetic values
	if plan.RunSetupTests.IsUnknown() {
		plan.RunSetupTests = state.RunSetupTests
	}
	if plan.TrustCertificates.IsUnknown() {
		plan.TrustCertificates = state.TrustCertificates
	}
	if plan.TrustFingerprints.IsUnknown() {
		plan.TrustFingerprints = state.TrustFingerprints
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewConnectionDelete().ConnectionID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connection Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}