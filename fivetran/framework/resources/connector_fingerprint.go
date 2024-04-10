package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ConnectorFingerprint() resource.Resource {
	return &connectorFingerprint{}
}

type connectorFingerprint struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectorFingerprint{}
var _ resource.ResourceWithImportState = &connectorFingerprint{}

func (r *connectorFingerprint) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_fingerprints"
}

func (r *connectorFingerprint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.FingerprintConnectorResource()
}

func (r *connectorFingerprint) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectorFingerprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.FingerprintConnector

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	for _, item := range data.Fingerprint.Elements() {
		if element, ok := item.(basetypes.ObjectValue); ok {
			svc := r.GetClient().NewCertificateConnectorFingerprintApprove()
			svc.ConnectorID(data.ConnectorId.ValueString())
			svc.Hash(element.Attributes()["hash"].(basetypes.StringValue).ValueString())
			svc.PublicKey(element.Attributes()["public_key"].(basetypes.StringValue).ValueString())
			if response, err := svc.Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Connector Fingerprint Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
				)

				return
			}
		}
	}

	listResponse, err := core.ReadFromSourceFingerprintCommon(ctx, r.GetClient(), data.ConnectorId.ValueString(), "connector")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Fingerprint Resource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorFingerprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.FingerprintConnector
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	listResponse, err := core.ReadFromSourceFingerprintCommon(ctx, r.GetClient(), data.ConnectorId.ValueString(), "connector")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connector Fingerprint Resource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorFingerprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.FingerprintConnector

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	planMap := make(map[string]string)
	for _, item := range plan.Fingerprint.Elements() {
		if element, ok := item.(basetypes.ObjectValue); ok {
			planMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["public_key"].(basetypes.StringValue).ValueString()
		}
	}

	stateMap := make(map[string]string)
	for _, connector := range state.Fingerprint.Elements() {
		if element, ok := connector.(basetypes.ObjectValue); ok {
			stateMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["public_key"].(basetypes.StringValue).ValueString()
		}
	}

	/* sync */
	for stateKey, _ := range stateMap {
		_, found := planMap[stateKey]

		if !found {
			if updateResponse, err := r.GetClient().NewConnectorFingerprintRevoke().ConnectorID(plan.ConnectorId.ValueString()).Hash(stateKey).Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Fingerprint Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
				)
				return
			}
		}
	}

	for planKey, planValue := range planMap {
		_, exists := stateMap[planKey]

		if !exists {
			if updateResponse, err := r.GetClient().NewCertificateConnectorFingerprintApprove().ConnectorID(plan.ConnectorId.ValueString()).Hash(planKey).PublicKey(planValue).Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Connector Fingerprint Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
				)
				return
			}
		}
	}

	listResponse, err := core.ReadFromSourceFingerprintCommon(ctx, r.GetClient(), plan.ConnectorId.ValueString(), "connector")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector Fingerprint Resource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	plan.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorFingerprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.FingerprintConnector

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	hashes := []string{}
	for _, items := range data.Fingerprint.Elements() {
		if element, ok := items.(basetypes.ObjectValue); ok {
			hashes = append(hashes, element.Attributes()["hash"].(basetypes.StringValue).ValueString())
		}
	}

	response, err := core.RevokeFingerptints(ctx, r.GetClient(), data.Id.ValueString(), "connector", hashes)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Connector Fingerprints Resource.",
			fmt.Sprintf("%v; code: %v", err, response.Code),
		)
	}
}
