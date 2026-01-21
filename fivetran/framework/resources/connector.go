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
		Version:    4,
	}
}

func (r *connector) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connector) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	v0ConfigTfTypes := model.GetTfTypes(common.GetConfigFieldsMap(), 1)

	v0ConfigTfTypes["servers"] = tftypes.String

	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 4 (Schema.Version)
		0: {
			// Optionally, the PriorSchema field can be defined.
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 0)
			},
		},
		// State upgrade implementation from 1 (prior state version) to 4 (Schema.Version)
		1: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 1)
			},
		},
		// State upgrade implementation from 2 (prior state version) to 4 (Schema.Version)
		2: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 2)
			},
		},
		// State upgrade implementation from 3 (prior state version) to 4 (Schema.Version)
		3: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 3)
			},
		},
		// State upgrade implementation from 4 (prior state version) to 5 (Schema.Version)
		4: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				upgradeConnectorState(ctx, req, resp, 4)
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

	svc := r.GetClient().NewConnectionCreate().
		Paused(true). // on creation we always create paused connector
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		ConfigCustom(&configMap) // on creation we have config always with schema params

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

	// The only case when state of existing resource doesn't contain required fields - import operation
	isImportOperation := data.GroupId.IsNull() || data.GroupId.IsUnknown() || data.Service.IsNull() || data.Service.IsUnknown()

	id := data.Id.ValueString()

	// Recovery from 1.1.13 bug
	if data.Id.IsNull() || data.Id.IsUnknown() {
		recoveredId, log := r.recoverId(ctx, data)
		if recoveredId == "" {
			resp.Diagnostics.AddError(
				"Read error.",
				"Unable to recover resource id from state.\n"+log,
			)
			return
		}
		id = recoveredId
	}

	response, err := r.GetClient().NewConnectionDetails().ConnectionID(id).DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	data.ReadFromResponse(response, isImportOperation)
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

	planOnlyAttributesChanged := (runSetupTestsPlan && runSetupTestsPlan != runSetupTestsState) ||
		(trustCertificatesPlan && trustCertificatesPlan != trustCertificatesState) ||
		(trustFingerprintsPlan && trustFingerprintsPlan != trustFingerprintsState)

	hasUpdates, patch, authPatch, err := plan.HasUpdates(plan, state)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Update Connector Resource.",
            fmt.Sprintf("%v; ", err),
        )
    }

	updatePerformed := false

	if planOnlyAttributesChanged {
		response, err := r.GetClient().NewConnectionSetupTests().ConnectionID(state.Id.ValueString()).DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		if response.Data.SetupTests != nil && len(response.Data.SetupTests) > 0 {
			for _, tr := range response.Data.SetupTests {
				if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
					resp.Diagnostics.AddWarning(
						fmt.Sprintf("Connector setup test `%v` has status `%v`", tr.Title, tr.Status),
						tr.Message,
					)
				}
			}
		}
		if !hasUpdates {
			runSetupTestsToPreserve := plan.RunSetupTests
			if runSetupTestsToPreserve.IsNull() || runSetupTestsToPreserve.IsUnknown() {
				runSetupTestsToPreserve = state.RunSetupTests
			}
			trustCertificatesToPreserve := plan.TrustCertificates
			if trustCertificatesToPreserve.IsNull() || trustCertificatesToPreserve.IsUnknown() {
				trustCertificatesToPreserve = state.TrustCertificates
			}
			trustFingerprintsToPreserve := plan.TrustFingerprints
			if trustFingerprintsToPreserve.IsNull() || trustFingerprintsToPreserve.IsUnknown() {
				trustFingerprintsToPreserve = state.TrustFingerprints
			}

			plan.ReadFromCreateResponse(response)

			plan.RunSetupTests = runSetupTestsToPreserve
			plan.TrustCertificates = trustCertificatesToPreserve
			plan.TrustFingerprints = trustFingerprintsToPreserve
		}
		updatePerformed = true
	}

	if hasUpdates {
		runSetupTestsToPreserve := plan.RunSetupTests
		if runSetupTestsToPreserve.IsNull() || runSetupTestsToPreserve.IsUnknown() {
			runSetupTestsToPreserve = state.RunSetupTests
		}
		trustCertificatesToPreserve := plan.TrustCertificates
		if trustCertificatesToPreserve.IsNull() || trustCertificatesToPreserve.IsUnknown() {
			trustCertificatesToPreserve = state.TrustCertificates
		}
		trustFingerprintsToPreserve := plan.TrustFingerprints
		if trustFingerprintsToPreserve.IsNull() || trustFingerprintsToPreserve.IsUnknown() {
			trustFingerprintsToPreserve = state.TrustFingerprints
		}

		svc := r.GetClient().NewConnectionUpdate().
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

		if !plan.ProxyAgentId.Equal(state.ProxyAgentId) {
			if !plan.ProxyAgentId.IsNull() {
				svc.ProxyAgentId(plan.ProxyAgentId.ValueString())
		}
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
				"Unable to Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		plan.ReadFromCreateResponse(response)

		plan.RunSetupTests = runSetupTestsToPreserve
		plan.TrustCertificates = trustCertificatesToPreserve
		plan.TrustFingerprints = trustFingerprintsToPreserve

		updatePerformed = true
	}

	if !updatePerformed {
		runSetupTestsToPreserve := plan.RunSetupTests
		if runSetupTestsToPreserve.IsNull() || runSetupTestsToPreserve.IsUnknown() {
			runSetupTestsToPreserve = state.RunSetupTests
		}
		trustCertificatesToPreserve := plan.TrustCertificates
		if trustCertificatesToPreserve.IsNull() || trustCertificatesToPreserve.IsUnknown() {
			trustCertificatesToPreserve = state.TrustCertificates
		}
		trustFingerprintsToPreserve := plan.TrustFingerprints
		if trustFingerprintsToPreserve.IsNull() || trustFingerprintsToPreserve.IsUnknown() {
			trustFingerprintsToPreserve = state.TrustFingerprints
		}

		// re-read connector upstream with an additional request after update
		response, err := r.GetClient().NewConnectionDetails().ConnectionID(state.Id.ValueString()).DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read after Update Connector Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		plan.ReadFromResponse(response, false)

		plan.RunSetupTests = runSetupTestsToPreserve
		plan.TrustCertificates = trustCertificatesToPreserve
		plan.TrustFingerprints = trustFingerprintsToPreserve
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

	// Save state
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

	deleteResponse, err := r.GetClient().NewConnectionDelete().ConnectionID(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connector Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}

// in case if state was corrupted and computable values wasn't saved we could recover resource id using group_id and schema
func (r *connector) recoverId(ctx context.Context, data model.ConnectorResourceModel) (string, string) {
	id := ""
	log := ""
	if !(data.GroupId.IsNull() || data.GroupId.IsUnknown()) {
		groupId := data.GroupId.ValueString()
		schemaName := ""
		if !data.DestinationSchema.IsNull() && !data.DestinationSchema.IsUnknown() {
			destinationSchema, err := data.GetDestinatonSchemaForConfig()
			if err == nil {
				log = log + "\n" + fmt.Sprintf("Destination schema: \n %v", destinationSchema)
				if prefix, ok := destinationSchema["schema_prefix"]; ok && prefix != "" {
					schemaName = prefix.(string)
				} else {
					if name, ok := destinationSchema["schema"]; ok && name != "" {
						schemaName = name.(string)
						if table, ok := destinationSchema["table"]; ok && table != "" {
							schemaName = schemaName + "." + table.(string)
						}
					}
				}
			} else {
				log = log + "\n" + err.Error()
			}
		}
		log = log + "\n" + fmt.Sprintf("Schema `%s`, group `%s", schemaName, groupId)
		if schemaName != "" && groupId != "" {
			connectorsList, err := r.
				GetClient().
				NewGroupListConnections().
				GroupID(groupId).
				Limit(1000).
				Do(ctx)
			found := false
			if err == nil {
				for _, c := range connectorsList.Data.Items {
					if c.Schema == schemaName {
						id = c.ID
						found = true
						break
					}
				}
				for !found && connectorsList.Data.NextCursor != "" {
					connectorsList, err = r.GetClient().
						NewGroupListConnections().
						GroupID(groupId).
						Limit(1000).
						Cursor(connectorsList.Data.NextCursor).
						Do(ctx)
					if err != nil {
						log = log + "\n" + err.Error()
						break
					} else {
						for _, c := range connectorsList.Data.Items {
							if c.Schema == schemaName {
								id = c.ID
								found = true
								break
							}
						}
					}
				}
			} else {
				log = log + "\n" + err.Error()
			}
			if !found {
				log = log + "\n" + fmt.Sprintf("Can't find connector with schema = `%s` in group with id = `%s", schemaName, groupId)
			}
		} else {
			log = log + "\n" + " not enough data in state for recovery: " + fmt.Sprintf("schema:`%s`, group:`%s", schemaName, groupId)
		}
	}
	return id, log
}
