package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Destination() resource.Resource {
	return &destination{}
}

type destination struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &destination{}
var _ resource.ResourceWithImportState = &destination{}
var _ resource.ResourceWithUpgradeState = &destination{}

func (r *destination) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *destination) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.DestinationAttributesSchema().GetResourceSchema(),
		Blocks:     fivetranSchema.DestinationResourceBlocks(ctx),
		Version:    2,
	}
}

func (r *destination) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 2 (Schema.Version)
		0: {
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeDestinationState(ctx, req, resp, 0)
			},
		},
		1: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeDestinationState(ctx, req, resp, 1)
			},
		},
	}
}

func (r *destination) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *destination) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}
	var data model.DestinationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configMap, err := data.GetConfigMap(true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Destination Resource.",
			fmt.Sprintf("%v;", err),
		)

		return
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, true)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, false)
	daylightSavingTimeEnabledPlan := core.GetBoolOrDefault(data.DaylightSavingTimeEnabled, false)

	svc := r.GetClient().NewDestinationCreate().
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		Region(data.Region.ValueString()).
		TimeZoneOffset(data.TimeZoneOffset.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		DaylightSavingTimeEnabled(daylightSavingTimeEnabledPlan).
		ConfigCustom(&configMap)

	if data.HybridDeploymentAgentId.ValueString() != "" {
		svc.HybridDeploymentAgentId(data.HybridDeploymentAgentId.ValueString())
	}

	if data.NetworkingMethod.ValueString() != "" {
		svc.NetworkingMethod(data.NetworkingMethod.ValueString())
	}

	if data.PrivateLinkId.ValueString() != "" {
		svc.PrivateLinkId(data.PrivateLinkId.ValueString())
	}

	response, err := svc.
		DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Destination Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)

		return
	}

	// For some reason tests may fail on first run, but succeed on second
	if runSetupTestsPlan && strings.ToLower(response.Data.SetupStatus) != "connected" {
		resp.Diagnostics.AddWarning(
			"Setup Tests for destination failed on creation. Running post-creation attempt.",
			fmt.Sprintf("%v", response.Data.SetupTests),
		)

		rsts := r.GetClient().NewDestinationSetupTests().
			DestinationID(response.Data.ID).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan)

		stResponse, err := rsts.DoCustom(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Destination Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, stResponse.Code, stResponse.Message),
			)

			return
		}

		if strings.ToLower(stResponse.Data.SetupStatus) != "connected" {
			if stResponse.Data.SetupTests != nil && len(stResponse.Data.SetupTests) > 0 {
				for _, tr := range stResponse.Data.SetupTests {
					if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
						resp.Diagnostics.AddWarning(
							fmt.Sprintf("Destination setup test `%v` has status `%v`", tr.Title, tr.Status),
							tr.Message,
						)
					}
				}
			}
		}

		detailsResponse, err := r.GetClient().NewDestinationDetails().DestinationID(response.Data.ID).DoCustom(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Create Destination Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, detailsResponse.Code, detailsResponse.Message),
			)

			return
		}

		// re-read destination details after setup-tests finished
		data.ReadFromResponse(detailsResponse)
	} else {
		data.ReadFromResponseWithTests(response)
	}
	data.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	data.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	data.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destination) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DestinationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	id := data.Id.ValueString()

	// Recovery from 1.1.13 bug
	if data.Id.IsUnknown() || data.Id.IsNull() {
		// Currently group_id -> 1:1 <- destination_id
		id = data.GroupId.ValueString()
	}

	response, err := r.GetClient().
		NewDestinationDetails().
		DestinationID(id).
		DoCustom(ctx)

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

func (r *destination) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.DestinationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, true)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, false)
	daylightSavingTimeEnabledPlan := core.GetBoolOrDefault(plan.DaylightSavingTimeEnabled, false)

	runSetupTestsState := core.GetBoolOrDefault(state.RunSetupTests, false)
	trustCertificatesState := core.GetBoolOrDefault(state.TrustCertificates, false)
	trustFingerprintsState := core.GetBoolOrDefault(state.TrustFingerprints, false)
	daylightSavingTimeEnabledState := core.GetBoolOrDefault(state.DaylightSavingTimeEnabled, false)

	hasUpdates, patch, err := plan.HasUpdates(plan, state)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Update Destination Resource.",
            fmt.Sprintf("%v; ", err),
        )
    }

	updatePerformed := false
	if hasUpdates {
		svc := r.GetClient().NewDestinationUpdate().
				RunSetupTests(runSetupTestsPlan).
				TrustCertificates(trustCertificatesPlan).
				TrustFingerprints(trustFingerprintsPlan).
				Region(plan.Region.ValueString()).
				DestinationID(state.Id.ValueString())

		if (!plan.TimeZoneOffset.Equal(state.TimeZoneOffset)) {
			svc.TimeZoneOffset(plan.TimeZoneOffset.ValueString())
		}

		if (daylightSavingTimeEnabledPlan != daylightSavingTimeEnabledState) {
			svc.DaylightSavingTimeEnabled(daylightSavingTimeEnabledPlan)
		}

		if (!plan.HybridDeploymentAgentId.IsUnknown() && !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId)) {
			svc.HybridDeploymentAgentId(plan.HybridDeploymentAgentId.ValueString())
		}

		if (!plan.NetworkingMethod.IsUnknown() && !plan.NetworkingMethod.Equal(state.NetworkingMethod)) {
			svc.NetworkingMethod(plan.NetworkingMethod.ValueString())
		}

		if (!plan.PrivateLinkId.IsUnknown() && !plan.PrivateLinkId.Equal(state.PrivateLinkId)) {
			svc.PrivateLinkId(plan.PrivateLinkId.ValueString())
		}

		if len(patch) > 0 {
			svc.ConfigCustom(&patch)
		}

		response, err := svc.DoCustom(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Destination Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		
		updatePerformed = true
		plan.ReadFromResponseWithTests(response)

		if runSetupTestsPlan && response.Data.SetupTests != nil && len(response.Data.SetupTests) > 0 {
			for _, tr := range response.Data.SetupTests {
				if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
					resp.Diagnostics.AddWarning(
						fmt.Sprintf("Destination setup test `%v` has status `%v`", tr.Title, tr.Status),
						tr.Message,
					)
				}
			}
		}
	} else {
		// If values of testing fields changed we should run tests
		if runSetupTestsPlan && runSetupTestsPlan != runSetupTestsState ||
			trustCertificatesPlan && trustCertificatesPlan != trustCertificatesState ||
			trustFingerprintsPlan && trustFingerprintsPlan != trustFingerprintsState {

			response, err := r.GetClient().NewDestinationSetupTests().DestinationID(state.Id.ValueString()).Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Destination Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
				)
				return
			}

			plan.ReadFromLegacyResponse(response)
			if response.Data.SetupTests != nil && len(response.Data.SetupTests) > 0 {
				for _, tr := range response.Data.SetupTests {
					if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
						resp.Diagnostics.AddWarning(
							fmt.Sprintf("Destination setup test `%v` has status `%v`", tr.Title, tr.Status),
							tr.Message,
						)
					}
				}
			}

			// there were no changes in config so we can just copy it from state
			plan.Config = state.Config
			updatePerformed = true
		}
	}

	if !updatePerformed {
		// re-read connector upstream with an additional request after update
		response, err := r.GetClient().NewDestinationDetails().DestinationID(state.Id.ValueString()).DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read after Update Destination Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		plan.ReadFromResponse(response)
	}

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *destination) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DestinationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewDestinationDelete().DestinationID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Destination Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
