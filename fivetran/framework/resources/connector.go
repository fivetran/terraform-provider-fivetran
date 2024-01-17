package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func Connector() resource.Resource {
	return &connector{}
}

type connector struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connector{}
var _ resource.ResourceWithUpgradeState = &connector{}
var _ resource.ResourceWithImportState = &connector{}

func (r *connector) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *connector) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.ConnectorAttributesSchema().GetResourceSchema(),
		Blocks:     fivetranSchema.ConnectorResourceBlocks(ctx),
		Version:    3,
	}
}

func (r *connector) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connector) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {

	v0ConfigTfTypes := model.GetTfTypes(common.GetConfigFieldsMap(), 1)

	v0ConfigTfTypes["servers"] = tftypes.String

	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 3 (Schema.Version)
		0: {
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 0)
			},
		},
		// State upgrade implementation from 1 (prior state version) to 3 (Schema.Version)
		1: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 1)
			},
		},
		// State upgrade implementation from 2 (prior state version) to 3 (Schema.Version)
		2: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 2)
			},
		},
	}
}

func (r *connector) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configMap, err := data.GetConfigMap(true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Resource.",
			fmt.Sprintf("%v;", err),
		)

		return
	}
	noConfig := configMap == nil
	authMap, err := data.GetAuthMap(true)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Resource.",
			fmt.Sprintf("%v;", err),
		)

		return
	}
	noAuth := authMap == nil

	destinationSchema, err := data.GetDestinatonSchemaForConfig()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Resource.",
			fmt.Sprintf("%v;", err),
		)

		return
	}

	if noConfig {
		configMap = make(map[string]interface{})
	}
	for k, v := range destinationSchema {
		configMap[k] = v
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, false)

	svc := r.GetClient().NewConnectorCreate().
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		ConfigCustom(&configMap) // on creation we have config always with schema params

	if !noAuth {
		svc.AuthCustom(&authMap)
	}

	response, err := svc.
		DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)

		return
	}

	data.ReadFromCreateResponse(response)

	data.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	data.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	data.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connector) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.GetClient().NewConnectorDetails().ConnectorID(data.Id.ValueString()).DoCustom(ctx)

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

func (r *connector) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, false)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, false)

	runSetupTestsState := core.GetBoolOrDefault(state.RunSetupTests, false)
	trustCertificatesState := core.GetBoolOrDefault(state.TrustCertificates, false)
	trustFingerprintsState := core.GetBoolOrDefault(state.TrustFingerprints, false)

	stateConfigMap, err := state.GetConfigMap(false)
	// this is not expected - state should contain only known fields relative to service
	// but we have to check error just in case
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Resource.",
			fmt.Sprintf("%v; ", err),
		)
	}

	stateAuthMap, err := state.GetAuthMap(false)

	// this is not expected - state should contain only known fields relative to service
	// but we have to check error just in case
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector Resource.",
			fmt.Sprintf("%v; ", err),
		)
	}

	planConfigMap, err := plan.GetConfigMap(false)

	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; ", err),
			)
		}
	}

	planAuthMap, err := plan.GetAuthMap(false)

	if err != nil {
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; ", err),
			)
		}
	}

	patch := model.PrepareConfigAuthPatch(stateConfigMap, planConfigMap, plan.Service.ValueString(), common.GetConfigFieldsMap())
	authPatch := model.PrepareConfigAuthPatch(stateAuthMap, planAuthMap, plan.Service.ValueString(), common.GetAuthFieldsMap())

	if len(patch) > 0 || len(authPatch) > 0 {
		svc := r.GetClient().NewConnectorModify().
			RunSetupTests(runSetupTestsPlan).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan).
			ConnectorID(state.Id.ValueString())

		if len(patch) > 0 {
			svc.ConfigCustom(&patch)
		}
		if len(authPatch) > 0 {
			svc.AuthCustom(&authPatch)
		}

		response, err := svc.DoCustom(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		plan.ReadFromCreateResponse(response)
	} else {
		// If values of testing fields changed we should run tests
		if runSetupTestsPlan && runSetupTestsPlan != runSetupTestsState ||
			trustCertificatesPlan && trustCertificatesPlan != trustCertificatesState ||
			trustFingerprintsPlan && trustFingerprintsPlan != trustFingerprintsState {

			response, err := r.GetClient().NewConnectorSetupTests().ConnectorID(state.Id.ValueString()).DoCustom(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
				)
				return
			}
			// nothing to read
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connector) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewConnectorDelete().ConnectorID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connector Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
