package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	externallogging "github.com/fivetran/go-fivetran/external_logging"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const externalLoggingV2UserAgentSuffix = "fivetran_external_logging_v2"

var externalLoggingV2Wording = dynamicWording{
	missingRequiredField: func(name string) string {
		return fmt.Sprintf("Log service metadata requires field %q when creating the external logging service.", name)
	},
	replacementWarning: func(fieldPath path.Path, diags *diag.Diagnostics) {
		diags.AddAttributeWarning(
			fieldPath,
			"Replacement Deletes the External Logging Service",
			fmt.Sprintf("Changing %s requires recreating the external logging service of this destination. A destination can have only one, so Terraform deletes the existing service before creating the new one, and log delivery stops until the new service exists.", fieldPath.String()),
		)
	},
}

func ExternalLoggingV2() resource.Resource {
	return &externalLoggingV2{}
}

type externalLoggingV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &externalLoggingV2{}
var _ resource.ResourceWithImportState = &externalLoggingV2{}
var _ resource.ResourceWithValidateConfig = &externalLoggingV2{}
var _ resource.ResourceWithModifyPlan = &externalLoggingV2{}

func (r *externalLoggingV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_logging_v2"
}

func (r *externalLoggingV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ExternalLoggingV2ResourceSchema()
}

func (r *externalLoggingV2) addUnconfiguredClientError(diags *diag.Diagnostics) {
	diags.AddError(
		"Unconfigured Fivetran Client",
		"Please report this issue to the provider developers.",
	)
}

// ImportState reads the log service by its ID, which is the ID of the group (destination) it belongs to.
func (r *externalLoggingV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	response, err := r.readExternalLogging(ctx, req.ID)
	if err != nil || response.Data.Id == "" {
		resp.Diagnostics.AddError(
			"Unable to Import External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	meta, err := r.logServiceMetadata(ctx, response.Data.Service)
	if err != nil {
		addLogServiceMetadataError(response.Data.Service, err, &resp.Diagnostics)
		return
	}

	data := model.ExternalLoggingV2ResourceModel{
		RunSetupTests:        types.BoolValue(false),
		Config:               types.DynamicNull(),
		ConfigSecrets:        types.DynamicNull(),
		ConfigSecretsVersion: types.Int64Null(),
	}
	mask := model.ExternalLoggingV2ImportConfigMask(response.Data.Config, meta)
	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, mask, types.DynamicNull())...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"fivetran_external_logging_v2 import requires HCL review",
		"Terraform imported the log service using API-visible config fields. Add the matching fivetran_external_logging_v2 resource block to your configuration and add secret values to config_secrets (or config), because the API masks them.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *externalLoggingV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var data model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.logServiceMetadata(ctx, data.Service.ValueString())
	if err != nil {
		addLogServiceMetadataError(data.Service.ValueString(), err, &resp.Diagnostics)
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
	secrets, diags := externalLoggingV2ConfigSecrets(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// configMap stays the projection mask, so secrets are never tracked in state.
	requestConfig := mergeConfigSecrets(configMap, secrets)

	enabled := core.GetBoolOrDefault(data.Enabled, true)
	runSetupTests := core.GetBoolOrDefault(data.RunSetupTests, false)

	// The group-level create endpoint doesn't take run_setup_tests, so tests are run through the
	// dedicated setup-tests endpoint once the service exists.
	response, err := r.GetClient().NewExternalLoggingCreate().
		GroupId(data.GroupId.ValueString()).
		Service(data.Service.ValueString()).
		Enabled(enabled).
		ConfigCustom(&requestConfig).
		DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMap, data.Config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RunSetupTests = types.BoolValue(runSetupTests)

	// Save the created service before running the tests, so it stays tracked even if they fail.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if runSetupTests {
		r.runSetupTests(ctx, data.Id.ValueString(), &resp.Diagnostics)
	}
}

func (r *externalLoggingV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var data model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.readExternalLogging(ctx, data.Id.ValueString())
	if err != nil {
		if strings.HasPrefix(response.Code, "NotFound") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}
	if response.Data.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	meta, err := r.logServiceMetadata(ctx, response.Data.Service)
	if err != nil {
		addLogServiceMetadataError(response.Data.Service, err, &resp.Diagnostics)
		return
	}

	configMask, diags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	runSetupTests := data.RunSetupTests

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMask, data.Config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RunSetupTests = runSetupTests

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *externalLoggingV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan, state model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.logServiceMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		addLogServiceMetadataError(plan.Service.ValueString(), err, &resp.Diagnostics)
		return
	}

	planConfig, diags := core.DynamicToMap(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	stateConfig, diags := core.DynamicToMap(ctx, state.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var slot *metadata.Property
	if meta != nil {
		slot = &meta.Config
	}
	secrets, diags := externalLoggingV2ConfigSecrets(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configPatch := core.PrepareConfigPatchDynamic(planConfig, stateConfig, slot)
	applyConfigSecretsToPatch(configPatch, secrets, fieldChanged(plan.ConfigSecretsVersion, state.ConfigSecretsVersion))

	enabledPlan := core.GetBoolOrDefault(plan.Enabled, true)
	enabledChanged := enabledPlan != core.GetBoolOrDefault(state.Enabled, true)
	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	runSetupTestsRequested := runSetupTestsPlan && !core.GetBoolOrDefault(state.RunSetupTests, false)

	hasChanges := len(configPatch) > 0 || enabledChanged
	id := state.Id.ValueString()

	switch {
	case hasChanges:
		updateSvc := r.GetClient().NewExternalLoggingUpdate().
			ExternalLoggingId(id).
			RunSetupTests(runSetupTestsPlan)
		if enabledChanged {
			updateSvc.Enabled(enabledPlan)
		}
		if len(configPatch) > 0 {
			updateSvc.ConfigCustom(&configPatch)
		}

		response, err := updateSvc.DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update External Logging Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}

		resp.Diagnostics.Append(state.ReadFromResponse(ctx, response, meta, planConfig, plan.Config)...)
		if resp.Diagnostics.HasError() {
			return
		}
		warnFailedLogSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	case runSetupTestsRequested:
		// The API rejects a PATCH without update parameters, so tests are triggered through
		// the dedicated setup-tests endpoint.
		r.runSetupTests(ctx, id, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !hasChanges {
		// No PATCH was sent, so state still holds the old config. Reproject it onto the planned
		// config: this drops keys removed from the configuration (a removal sends no patch when
		// metadata is unavailable or the field is not nullable). The planned config is kept as is when it
		// holds the same values, so switching between an object and an equivalent map keeps the planned type.
		config, diags := core.MapToDynamicWithTemplate(ctx, core.ProjectDynamic(stateConfig, planConfig, slot), plan.Config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Config = config
	}

	state.RunSetupTests = types.BoolValue(runSetupTestsPlan)
	state.ConfigSecretsVersion = plan.ConfigSecretsVersion
	// config_secrets is write-only and must never be stored.
	state.ConfigSecrets = types.DynamicNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *externalLoggingV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var data model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewExternalLoggingDelete().
		ExternalLoggingId(data.Id.ValueString()).
		DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	if err != nil && !strings.HasPrefix(deleteResponse.Code, "NotFound") {
		resp.Diagnostics.AddError(
			"Unable to Delete External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
	}
}

func (r *externalLoggingV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.GetSkipPlanTimeValidation() {
		return
	}

	var data model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Service.IsNull() || data.Service.IsUnknown() || data.Service.ValueString() == "" {
		return
	}

	configMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	secretsMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.ConfigSecrets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || (len(configMap) == 0 && len(secretsMap) == 0) {
		return
	}

	meta, err := r.logServiceMetadata(ctx, data.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Validate External Logging Configuration",
			fmt.Sprintf("Unable to fetch metadata for service %q. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", data.Service.ValueString(), err),
		)
		return
	}
	if meta == nil {
		// The metadata API does not describe this log service; the API validates the config.
		return
	}

	validateDynamicObject(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
	rejectReadonlyDynamicFields(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
	validateDynamicObject(secretsMap, &meta.Config, path.Root("config_secrets"), &resp.Diagnostics)
	validateConfigSecrets(secretsMap, configMap, &meta.Config, &resp.Diagnostics)
}

func (r *externalLoggingV2) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.GetSkipPlanTimeValidation() {
		return
	}

	var plan model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Service.IsNull() || plan.Service.IsUnknown() || plan.Service.ValueString() == "" {
		return
	}

	// A config that is not known yet (for example jsondecode of another resource's computed value)
	// cannot be validated; it would look like an empty object and report every field as missing.
	if dynamicIsUnknown(plan.Config) {
		return
	}

	planConfig, diags := core.DynamicToMapPreserveUnknown(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isCreate := req.State.Raw.IsNull()
	// Changing service or group_id replaces the log service, so the new one must be validated as a create.
	replaced := false
	var state model.ExternalLoggingV2ResourceModel
	stateConfig := map[string]interface{}(nil)
	if !isCreate {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		stateConfig, diags = core.DynamicToMapPreserveUnknown(ctx, state.Config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		serviceChanged := knownAndChanged(plan.Service, state.Service)
		if serviceChanged {
			// A destination can have only one log service: Terraform deletes the old one first,
			// and a replacement that fails validation would leave the destination without one.
			externalLoggingV2Wording.replacementWarning(path.Root("service"), &resp.Diagnostics)
		}
		replaced = serviceChanged || knownAndChanged(plan.GroupId, state.GroupId)
		if len(planConfig) == 0 && !replaced {
			return
		}
	}

	meta, err := r.logServiceMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Modify External Logging Plan",
			fmt.Sprintf("Unable to fetch metadata for service %q. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", plan.Service.ValueString(), err),
		)
		return
	}
	if meta == nil {
		return
	}

	if isCreate || replaced {
		validateExternalLoggingV2RequiredFields(ctx, req.Config, planConfig, meta, &resp.Diagnostics)
		return
	}

	comparable := withoutUntrackedImmutableSecrets(planConfig, stateConfig, &meta.Config, &resp.Diagnostics)
	if appendImmutableRequiresReplace(comparable, stateConfig, &meta.Config, path.Root("config"), resp, plan.Service.ValueString(), "", externalLoggingV2Wording) {
		validateExternalLoggingV2RequiredFields(ctx, req.Config, planConfig, meta, &resp.Diagnostics)
	}
}

func (r *externalLoggingV2) readExternalLogging(ctx context.Context, id string) (externallogging.ExternalLoggingCustomResponse, error) {
	return r.GetClient().NewExternalLoggingDetails().
		ExternalLoggingId(id).
		DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
}

func (r *externalLoggingV2) runSetupTests(ctx context.Context, id string, diags *diag.Diagnostics) {
	response, err := r.GetClient().NewExternalLoggingSetupTests().
		ExternalLoggingId(id).
		DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	if err != nil {
		diags.AddError(
			"Unable to Run External Logging Setup Tests.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}
	warnFailedLogSetupTests(response.Data.SetupTests, diags)
}

// logServiceMetadata returns the metadata describing a log service, or nil when the metadata API has
// no description for it. Callers must treat nil as "no plan-time validation available".
func (r *externalLoggingV2) logServiceMetadata(ctx context.Context, service string) (*metadata.LogServiceMetadata, error) {
	cache := r.GetMetadataCache()
	if cache == nil {
		cache = &sync.Map{}
	}
	if r.GetClient() == nil {
		if meta, ok, err := core.LoadCachedLogServiceMetadata(cache, service); ok || err != nil {
			return meta, err
		}
		return nil, errUnconfiguredClient
	}
	return core.GetCachedLogServiceMetadataWithUserAgentSuffix(ctx, r.GetClient(), cache, service, externalLoggingV2UserAgentSuffix)
}

func addLogServiceMetadataError(service string, err error, diags *diag.Diagnostics) {
	diags.AddError(
		"Unable to Fetch Log Service Metadata.",
		fmt.Sprintf("Unable to fetch metadata for service %q: %v", service, err),
	)
}

func knownAndChanged(plan, state types.String) bool {
	return !state.IsNull() && !state.IsUnknown() && !plan.IsUnknown() && !plan.Equal(state)
}

// withoutUntrackedImmutableSecrets returns planConfig without immutable secret fields (format
// "password") that state doesn't track. Import leaves secrets out of state because the API only
// returns them masked, so such a field can't be compared and must not plan a replacement: that
// would delete the log service, for example when migrating from fivetran_external_logging with
// import blocks. Immutable fields are never sent in an update, so the value stays as it is.
func withoutUntrackedImmutableSecrets(planConfig, stateConfig map[string]interface{}, slot *metadata.Property, diags *diag.Diagnostics) map[string]interface{} {
	result := make(map[string]interface{}, len(planConfig))
	for name, value := range planConfig {
		prop := core.SlotProp(slot, name)
		if _, tracked := stateConfig[name]; !tracked && prop != nil && prop.Immutable && prop.Format == "password" && !core.IsDynamicUnknownValue(value) {
			diags.AddAttributeWarning(
				path.Root("config").AtName(name),
				"Immutable Secret Not Compared",
				fmt.Sprintf("State doesn't hold %s, for example because the log service was imported and the API only returns secrets masked, so Terraform can't check that the configured value matches. The field is immutable: the value is not sent to the API and doesn't recreate the external logging service.", path.Root("config").AtName(name).String()),
			)
			continue
		}
		result[name] = value
	}
	return result
}
