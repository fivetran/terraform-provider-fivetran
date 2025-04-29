package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ExternalLogging() resource.Resource {
	return &externalLogging{}
}

type externalLogging struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &externalLogging{}
var _ resource.ResourceWithImportState = &externalLogging{}

func (r *externalLogging) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_logging"
}

func (r *externalLogging) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ExternalLoggingResource()
}

func (r *externalLogging) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *externalLogging) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ExternalLogging

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewExternalLoggingCreate()
	svc.GroupId(data.GroupId.ValueString())
	svc.Service(data.Service.ValueString())
	svc.Enabled(core.GetBoolOrDefault(data.Enabled, true))

	config := data.GetConfig()
	svc.ConfigCustom(&config)

	createResponse, err := svc.DoCustom(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromCustomResponse(ctx, createResponse)

	runTests := core.GetBoolOrDefault(data.RunTests, false)
	if runTests {
		testsSvc := r.GetClient().NewExternalLoggingSetupTests().ExternalLoggingId(data.Id.ValueString())
		response, err := testsSvc.Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Start External Logging Tests.",
				fmt.Sprintf("%v; code: %v", err, response.Code),
			)
		}

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

		// nothing to read
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *externalLogging) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ExternalLogging

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	readResponse, err := r.GetClient().NewExternalLoggingDetails().ExternalLoggingId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read External Logging Resource.",
			fmt.Sprintf("%v; code: %v", err, readResponse.Code),
		)
		return
	}

	data.ReadFromResponse(ctx, readResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *externalLogging) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ExternalLogging
	hasChanges := false

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	runTests := core.GetBoolOrDefault(plan.RunTests, false)
	runTestsState := core.GetBoolOrDefault(state.RunTests, false)
	enabledPlan := core.GetBoolOrDefault(plan.Enabled, true)
	enabledState := core.GetBoolOrDefault(state.Enabled, true)

	svc := r.GetClient().NewExternalLoggingUpdate().ExternalLoggingId(state.Id.ValueString())

	if enabledPlan != enabledState {
		svc.Enabled(core.GetBoolOrDefault(plan.Enabled, true))
		hasChanges = true
	}

	if !plan.Config.Equal(state.Config) {
		config := plan.GetConfig()
		svc.ConfigCustom(&config)
		hasChanges = true
	}

	if hasChanges {
		updateResponse, err := svc.DoCustom(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update External Logging Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
			)
			return
		}

		state.ReadFromCustomResponse(ctx, updateResponse)
	}

	if runTests && runTests != runTestsState {
		testsSvc := r.GetClient().NewExternalLoggingSetupTests().ExternalLoggingId(state.Id.ValueString())
		response, err := testsSvc.Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Start External Logging Tests.",
				fmt.Sprintf("%v; code: %v", err, response.Code),
			)
		}

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
		// nothing to read
		state.RunTests = plan.RunTests
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *externalLogging) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ExternalLogging

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewExternalLoggingDelete().ExternalLoggingId(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
