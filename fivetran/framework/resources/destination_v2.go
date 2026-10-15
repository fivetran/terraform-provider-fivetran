package resources

import (
	"context"
	"fmt"
	"sync"

	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/destinations"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DestinationV2() resource.Resource {
	return &destinationV2{}
}

type destinationV2 struct {
	core.ProviderResource
}

const destinationV2UserAgentSuffix = "fivetran_destination_v2"

// The API defaults run_setup_tests to true and both trust flags to false.
const (
	destinationV2RunSetupTestsDefault     = true
	destinationV2TrustCertificatesDefault = false
	destinationV2TrustFingerprintsDefault = false
)

var _ resource.ResourceWithConfigure = &destinationV2{}
var _ resource.ResourceWithImportState = &destinationV2{}
var _ resource.ResourceWithModifyPlan = &destinationV2{}
var _ resource.ResourceWithValidateConfig = &destinationV2{}

func (r *destinationV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination_v2"
}

func (r *destinationV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.DestinationV2ResourceSchema()
}

func (r *destinationV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	details, err := r.GetClient().NewDestinationDetails().DestinationID(req.ID).DoCustomWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Destination V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	meta, err := r.destinationMetadata(ctx, details.Data.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Destination Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", details.Data.Service, err),
		)
		return
	}

	data := model.DestinationV2ResourceModel{
		RunSetupTests:     types.BoolValue(destinationV2RunSetupTestsDefault),
		TrustCertificates: types.BoolValue(destinationV2TrustCertificatesDefault),
		TrustFingerprints: types.BoolValue(destinationV2TrustFingerprintsDefault),
	}

	resp.Diagnostics.Append(data.ReadFromResponseForImport(ctx, details, meta)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"fivetran_destination_v2 import requires HCL review",
		"Terraform imported the destination using API-visible config fields. Add the matching fivetran_destination_v2 resource block to your configuration and add any sensitive config values (for example passwords or private keys) because the API does not return secrets.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destinationV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.destinationMetadata(ctx, data.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Destination Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", data.Service.ValueString(), err),
		)
		return
	}

	configMap, diags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if configMap == nil {
		configMap = map[string]interface{}{}
	}

	runSetupTestsPlan := core.GetBoolOrDefault(data.RunSetupTests, destinationV2RunSetupTestsDefault)
	trustCertificatesPlan := core.GetBoolOrDefault(data.TrustCertificates, destinationV2TrustCertificatesDefault)
	trustFingerprintsPlan := core.GetBoolOrDefault(data.TrustFingerprints, destinationV2TrustFingerprintsDefault)
	configPlan := data.Config

	svc := r.GetClient().NewDestinationCreate().
		Service(data.Service.ValueString()).
		GroupID(data.GroupId.ValueString()).
		TimeZoneOffset(data.TimeZoneOffset.ValueString()).
		RunSetupTests(runSetupTestsPlan).
		TrustCertificates(trustCertificatesPlan).
		TrustFingerprints(trustFingerprintsPlan).
		ConfigCustom(&configMap)

	r.applyCreateRootFields(svc, data)

	response, err := svc.DoCustomWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Destination V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	resp.Diagnostics.Append(data.ReadFromCreateResponse(ctx, response, meta, configMap)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Terraform requires the applied config to equal the planned config. Drift
	// introduced by the API (normalization, defaults) is surfaced by the next Read.
	data.Config = configPlan
	data.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	data.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	data.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	warnFailedDestinationSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destinationV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.GetClient().NewDestinationDetails().DestinationID(data.Id.ValueString()).DoCustomWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
	if err != nil {
		if response.Code == "NotFound_Destination" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Destination V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	meta, err := r.destinationMetadata(ctx, response.Data.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Destination Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", response.Data.Service, err),
		)
		return
	}

	configMask, diags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	runSetupTests := data.RunSetupTests
	trustCertificates := data.TrustCertificates
	trustFingerprints := data.TrustFingerprints

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMask)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RunSetupTests = runSetupTests
	data.TrustCertificates = trustCertificates
	data.TrustFingerprints = trustFingerprints

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destinationV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.destinationMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Destination Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", plan.Service.ValueString(), err),
		)
		return
	}

	planConfig, diags := core.DynamicToMap(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	stateConfig, diags := core.DynamicToMap(ctx, state.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configPatch := core.PrepareConfigPatchDynamic(planConfig, stateConfig, &meta.Config)

	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, destinationV2RunSetupTestsDefault)
	trustCertificatesPlan := core.GetBoolOrDefault(plan.TrustCertificates, destinationV2TrustCertificatesDefault)
	trustFingerprintsPlan := core.GetBoolOrDefault(plan.TrustFingerprints, destinationV2TrustFingerprintsDefault)
	configPlan := plan.Config

	hasOtherChanges := len(configPatch) > 0 || r.hasRootFieldChanges(plan, state)

	if hasOtherChanges {
		// The API defaults run_setup_tests to true when omitted from the PATCH, so
		// the setup-test flags are always sent explicitly to reflect the TF config.
		svc := r.GetClient().NewDestinationUpdate().
			DestinationID(state.Id.ValueString()).
			RunSetupTests(runSetupTestsPlan).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan)

		if len(configPatch) > 0 {
			svc.ConfigCustom(&configPatch)
		}
		r.applyUpdateRootFields(svc, plan, state)

		response, err := svc.DoCustomWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Destination V2 Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		if runSetupTestsPlan {
			warnFailedDestinationSetupTests(response.Data.SetupTests, &resp.Diagnostics)
		}
	} else if runSetupTestsPlan && r.hasSetupTestFieldChanges(plan, state) {
		// The Fivetran API rejects an update PATCH containing only
		// run_setup_tests/trust_certificates/trust_fingerprints ("No update
		// parameters specified"). Use the dedicated setup-tests endpoint instead.
		response, err := r.GetClient().NewDestinationSetupTests().
			DestinationID(state.Id.ValueString()).
			TrustCertificates(trustCertificatesPlan).
			TrustFingerprints(trustFingerprintsPlan).
			DoCustom(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Run Destination V2 Setup Tests.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		warnFailedDestinationSetupTests(response.Data.SetupTests, &resp.Diagnostics)
	}

	details, err := r.GetClient().NewDestinationDetails().DestinationID(state.Id.ValueString()).DoCustomWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Destination V2 Resource After Update.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(plan.ReadFromResponse(ctx, details, meta, planConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Config = configPlan
	plan.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	plan.TrustCertificates = types.BoolValue(trustCertificatesPlan)
	plan.TrustFingerprints = types.BoolValue(trustFingerprintsPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *destinationV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewDestinationDelete().DestinationID(data.Id.ValueString()).DoWithUserAgentSuffix(ctx, destinationV2UserAgentSuffix)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Destination V2 Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}

func (r *destinationV2) destinationMetadata(ctx context.Context, service string) (*metadata.DestinationMetadata, error) {
	cache := r.GetMetadataCache()
	if cache == nil {
		cache = &sync.Map{}
	}
	if r.GetClient() == nil {
		if meta, ok, err := core.LoadCachedDestinationMetadata(cache, service); ok || err != nil {
			return meta, err
		}
		return nil, errUnconfiguredClient
	}
	return core.GetCachedDestinationMetadataWithUserAgentSuffix(ctx, r.GetClient(), cache, service, destinationV2UserAgentSuffix)
}

func (r *destinationV2) applyCreateRootFields(svc *destinations.DestinationCreateService, data model.DestinationV2ResourceModel) {
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		svc.Region(data.Region.ValueString())
	}
	if !data.DaylightSavingTimeEnabled.IsNull() && !data.DaylightSavingTimeEnabled.IsUnknown() {
		svc.DaylightSavingTimeEnabled(data.DaylightSavingTimeEnabled.ValueBool())
	}
	if !data.NetworkingMethod.IsNull() && !data.NetworkingMethod.IsUnknown() {
		svc.NetworkingMethod(data.NetworkingMethod.ValueString())
	}
	if !data.PrivateLinkId.IsNull() && !data.PrivateLinkId.IsUnknown() {
		svc.PrivateLinkId(data.PrivateLinkId.ValueString())
	}
	if !data.ProxyAgentId.IsNull() && !data.ProxyAgentId.IsUnknown() {
		svc.ProxyAgentId(data.ProxyAgentId.ValueString())
	}
	if !data.HybridDeploymentAgentId.IsNull() && !data.HybridDeploymentAgentId.IsUnknown() {
		svc.HybridDeploymentAgentId(data.HybridDeploymentAgentId.ValueString())
	}
}

// hasRootFieldChanges reports whether any root-level field applyUpdateRootFields
// would send has actually changed between plan and state.
func (r *destinationV2) hasRootFieldChanges(plan, state model.DestinationV2ResourceModel) bool {
	fields := []struct{ plan, state attr.Value }{
		{plan.Region, state.Region},
		{plan.TimeZoneOffset, state.TimeZoneOffset},
		{plan.DaylightSavingTimeEnabled, state.DaylightSavingTimeEnabled},
		{plan.NetworkingMethod, state.NetworkingMethod},
		{plan.PrivateLinkId, state.PrivateLinkId},
		{plan.ProxyAgentId, state.ProxyAgentId},
		{plan.HybridDeploymentAgentId, state.HybridDeploymentAgentId},
	}
	for _, f := range fields {
		if fieldChanged(f.plan, f.state) {
			return true
		}
	}
	return false
}

func (r *destinationV2) hasSetupTestFieldChanges(plan, state model.DestinationV2ResourceModel) bool {
	return fieldChanged(plan.RunSetupTests, state.RunSetupTests) ||
		fieldChanged(plan.TrustCertificates, state.TrustCertificates) ||
		fieldChanged(plan.TrustFingerprints, state.TrustFingerprints)
}

func (r *destinationV2) applyUpdateRootFields(svc *destinations.DestinationUpdateService, plan, state model.DestinationV2ResourceModel) {
	if fieldChanged(plan.Region, state.Region) {
		svc.Region(plan.Region.ValueString())
	}
	if fieldChanged(plan.TimeZoneOffset, state.TimeZoneOffset) {
		svc.TimeZoneOffset(plan.TimeZoneOffset.ValueString())
	}
	if fieldChanged(plan.DaylightSavingTimeEnabled, state.DaylightSavingTimeEnabled) {
		svc.DaylightSavingTimeEnabled(plan.DaylightSavingTimeEnabled.ValueBool())
	}
	if fieldChanged(plan.NetworkingMethod, state.NetworkingMethod) {
		svc.NetworkingMethod(plan.NetworkingMethod.ValueString())
	}
	if fieldChanged(plan.PrivateLinkId, state.PrivateLinkId) {
		svc.PrivateLinkId(plan.PrivateLinkId.ValueString())
	}
	if fieldChanged(plan.ProxyAgentId, state.ProxyAgentId) {
		svc.ProxyAgentId(plan.ProxyAgentId.ValueString())
	}
	if fieldChanged(plan.HybridDeploymentAgentId, state.HybridDeploymentAgentId) {
		svc.HybridDeploymentAgentId(plan.HybridDeploymentAgentId.ValueString())
	}
}

func warnFailedDestinationSetupTests(setupTests []common.SetupTestResponse, diags *diag.Diagnostics) {
	for _, tr := range setupTests {
		if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
			diags.AddWarning(
				fmt.Sprintf("Destination setup test `%v` has status `%v`", tr.Title, tr.Status),
				tr.Message,
			)
		}
	}
}
