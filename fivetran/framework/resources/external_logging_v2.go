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

func externalLoggingV2WordingFor(accountLevel bool) dynamicWording {
	if accountLevel {
		return accountExternalLoggingWording
	}
	return externalLoggingV2Wording
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

// ImportState reads the log service by its ID: the ID of the group (destination) it belongs to, or
// the account ID for the account-level log service. The import ID "account" is an alias for the
// account-level log service; state stores its real ID, the account ID.
func (r *externalLoggingV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		r.addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	response, accountLevel, err := r.readForImport(ctx, req.ID)
	if err != nil {
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
	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, mask, types.DynamicNull(), accountLevel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(storeImportedConfigFields(ctx, resp.Private, data.Config)...)
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
	var slot *metadata.Property
	if meta != nil {
		slot = &meta.Config
	}
	requestConfig := core.CoerceMapToMetadata(mergeConfigSecrets(configMap, secrets), slot)

	enabled := core.GetBoolOrDefault(data.Enabled, true)
	runSetupTests := core.GetBoolOrDefault(data.RunSetupTests, false)

	response, err := r.createExternalLogging(ctx, externalLoggingV2TargetOf(data), data.Service.ValueString(), enabled, requestConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMap, data.Config, data.IsAccountLevel())...)
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
		r.runSetupTests(ctx, externalLoggingV2TargetOf(data), &resp.Diagnostics)
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

	response, err := r.readExternalLogging(ctx, externalLoggingV2TargetOf(data))
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

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMask, data.Config, data.IsAccountLevel())...)
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
	// Imported fields left out of config stop being tracked without being cleared upstream.
	omitUndeclaredImportedConfigFields(configPatch, importedConfigPatchFilter(ctx, req.Private, planConfig, stateConfig, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}
	// Secrets are sent when config_secrets_version changes or when they move from config, without
	// immutable and readonly fields, which ordinary config updates never send either.
	configPatch = core.CoerceMapToMetadata(applyConnectionV2SecretsToPatch(configPatch, secrets, stateConfig, slot, fieldChanged(plan.ConfigSecretsVersion, state.ConfigSecretsVersion)), slot)

	enabledPlan := core.GetBoolOrDefault(plan.Enabled, true)
	enabledChanged := enabledPlan != core.GetBoolOrDefault(state.Enabled, true)
	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	runSetupTestsRequested := runSetupTestsPlan && !core.GetBoolOrDefault(state.RunSetupTests, false)

	hasChanges := len(configPatch) > 0 || enabledChanged
	// group_id can't change in place: it forces a replacement.
	target := externalLoggingV2TargetOf(state)

	switch {
	case hasChanges:
		var enabledUpdate *bool
		if enabledChanged {
			enabledUpdate = &enabledPlan
		}
		var configUpdate map[string]interface{}
		if len(configPatch) > 0 {
			configUpdate = configPatch
		}

		response, err := r.updateExternalLogging(ctx, target, enabledUpdate, runSetupTestsPlan, configUpdate)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update External Logging Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}

		resp.Diagnostics.Append(state.ReadFromResponse(ctx, response, meta, planConfig, plan.Config, target.accountLevel())...)
		if resp.Diagnostics.HasError() {
			return
		}
		warnFailedLogSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	case runSetupTestsRequested:
		// The API rejects a PATCH without update parameters, so tests are triggered through
		// the dedicated setup-tests endpoint.
		r.runSetupTests(ctx, target, &resp.Diagnostics)
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
	resp.Diagnostics.Append(updateImportedConfigFields(ctx, req, resp)...)
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

	deleteResponse, err := r.deleteExternalLogging(ctx, externalLoggingV2TargetOf(data))
	if err != nil && !strings.HasPrefix(deleteResponse.Code, "NotFound") {
		resp.Diagnostics.AddError(
			"Unable to Delete External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
	}
}

func (r *externalLoggingV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data model.ExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Overlap is a local configuration error, independent of service metadata: it is checked even
	// when metadata is unavailable or plan-time validation is skipped.
	configMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	secretsMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.ConfigSecrets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	validateConnectionV2SecretsOverlap(configMap, secretsMap, path.Root("config"), configSecretsPath, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || r.GetSkipPlanTimeValidation() {
		return
	}

	if data.Service.IsNull() || data.Service.IsUnknown() || data.Service.ValueString() == "" {
		return
	}
	if len(configMap) == 0 && len(secretsMap) == 0 {
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
	validateDynamicObject(secretsMap, &meta.Config, configSecretsPath, &resp.Diagnostics)
	// Overlap with config was already checked above, so only the secret-field check is left.
	validateConfigSecrets(secretsMap, nil, &meta.Config, &resp.Diagnostics)
}

func (r *externalLoggingV2) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	warnUndeclaredImportedConfigFields(ctx, req, resp, "log service")
	if r.GetSkipPlanTimeValidation() {
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
	wording := externalLoggingV2WordingFor(plan.IsAccountLevel())
	// Changing service or group_id (including between a destination and the account level) replaces
	// the log service, so the new one must be validated as a create.
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
			// A destination, and the account, can have only one log service: Terraform deletes the
			// old one first, and a replacement that fails validation would leave it without one.
			wording.replacementWarning(path.Root("service"), &resp.Diagnostics)
		}
		replaced = serviceChanged || groupIdChanged(plan.GroupId, state.GroupId)
		secrets, _ := connectionV2SecretsFromConfig(ctx, req.Config, configSecretsPath, true, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(planConfig) == 0 && len(secrets) == 0 && !replaced {
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
		validateExternalLoggingV2RequiredFields(ctx, req.Config, planConfig, meta, wording, &resp.Diagnostics)
		return
	}

	service := plan.Service.ValueString()
	comparable := withoutUntrackedImmutableSecrets(planConfig, stateConfig, &meta.Config, &resp.Diagnostics)
	hasImmutableReplacement := appendImmutableRequiresReplace(comparable, stateConfig, &meta.Config, path.Root("config"), resp, service, "", wording)

	// A secret moved from config to config_secrets can still be compared with its tracked value,
	// so changing an immutable one while moving it recreates the log service.
	secrets, _ := connectionV2SecretsFromConfig(ctx, req.Config, configSecretsPath, true, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	var secretsReplacement resource.ModifyPlanResponse
	if appendImmutableRequiresReplace(movedConnectionV2Secrets(secrets, stateConfig), stateConfig, &meta.Config, configSecretsPath, &secretsReplacement, service, "", wording) {
		// Write-only values are null in the plan, so nested replacement paths do not
		// exist there. Keep field-specific warnings, but replace at the attribute root.
		resp.RequiresReplace.Append(configSecretsPath)
		hasImmutableReplacement = true
	}
	resp.Diagnostics.Append(secretsReplacement.Diagnostics...)

	if hasImmutableReplacement {
		validateExternalLoggingV2RequiredFields(ctx, req.Config, planConfig, meta, wording, &resp.Diagnostics)
	}
}

// externalLoggingV2AccountImportId is the import ID alias for the account-level log service, for users
// who don't know their account ID. Group IDs are generated words joined by "_", so no group has it.
const externalLoggingV2AccountImportId = "account"

// readForImport finds the log service an import ID addresses. The alias "account" addresses the
// account-level log service. A group ID addresses the log service of that destination. Otherwise the
// ID must be the account ID, which is the ID of the account-level log service. Group and account IDs
// never collide, so the group lookup goes first and the account one is only tried when it fails; the
// group lookup's error is reported when neither matches.
//
// The alias is handled before the group lookup, and the group lookup only counts when it returns the
// import ID: GET /external-logging/account is the account-level endpoint, so a group lookup for
// "account" would otherwise import the account-level log service as the log service of a destination
// with the account ID as group_id.
func (r *externalLoggingV2) readForImport(ctx context.Context, id string) (externallogging.ExternalLoggingCustomResponse, bool, error) {
	if id == externalLoggingV2AccountImportId {
		response, err := r.readExternalLogging(ctx, externalLoggingV2Target{})
		if err == nil && response.Data.Id == "" {
			err = fmt.Errorf("no account-level external logging service found")
		}
		return response, true, err
	}

	groupResponse, groupErr := r.readExternalLogging(ctx, externalLoggingV2Target{groupId: id})
	if groupErr == nil && groupResponse.Data.Id != "" && groupResponse.Data.Id == id {
		return groupResponse, false, nil
	}

	accountResponse, accountErr := r.readExternalLogging(ctx, externalLoggingV2Target{})
	if accountErr == nil && accountResponse.Data.Id != "" && accountResponse.Data.Id == id {
		return accountResponse, true, nil
	}

	if groupErr == nil {
		groupErr = fmt.Errorf("no external logging service found for destination or account ID %q", id)
	}
	return groupResponse, false, groupErr
}

// groupIdChanged reports a group_id change, including switching between a destination log service
// (group_id set) and the account-level one (group_id null). Both replace the log service.
func groupIdChanged(plan, state types.String) bool {
	return !plan.IsUnknown() && !state.IsUnknown() && !plan.Equal(state)
}

// logServiceMetadata returns the metadata describing a log service, or nil when the metadata API has
// no description for it. Callers must treat nil as "no plan-time validation available".
func (r *externalLoggingV2) logServiceMetadata(ctx context.Context, service string) (*metadata.LogServiceMetadata, error) {
	return readLogServiceMetadata(ctx, &r.ProviderResource, service, externalLoggingV2UserAgentSuffix)
}

// readLogServiceMetadata reads the metadata of a log service from /metadata/log-service-types/{service}
// through the provider's metadata cache, or returns nil when the metadata API has no description for
// it (for example a server that does not serve the endpoint yet). fivetran_external_logging_v2 and the
// deprecated fivetran_account_external_logging_v2 share it and the cache; only the User-Agent suffix
// differs. Without a configured client, only a cached answer is returned.
func readLogServiceMetadata(ctx context.Context, p *core.ProviderResource, service, userAgentSuffix string) (*metadata.LogServiceMetadata, error) {
	cache := p.GetMetadataCache()
	if cache == nil {
		cache = &sync.Map{}
	}
	if p.GetClient() == nil {
		if meta, ok, err := core.LoadCachedLogServiceMetadata(cache, service); ok || err != nil {
			return meta, err
		}
		return nil, errUnconfiguredClient
	}
	return core.GetCachedLogServiceMetadataWithUserAgentSuffix(ctx, p.GetClient(), cache, service, userAgentSuffix)
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
