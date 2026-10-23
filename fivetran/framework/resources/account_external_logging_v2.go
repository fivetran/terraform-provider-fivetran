package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/fivetran/go-fivetran/common"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const accountExternalLoggingV2UserAgentSuffix = "fivetran_account_external_logging_v2"

const logServiceMetadataCachePrefix = "log-service:"

// noMetadataCachePrefix marks services for which the metadata API has no description, so the
// negative answer is not requested again on every plan.
const noMetadataCachePrefix = "no-metadata:"

func AccountExternalLoggingV2() resource.Resource {
	return &accountExternalLoggingV2{}
}

type accountExternalLoggingV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &accountExternalLoggingV2{}
var _ resource.ResourceWithImportState = &accountExternalLoggingV2{}
var _ resource.ResourceWithValidateConfig = &accountExternalLoggingV2{}
var _ resource.ResourceWithModifyPlan = &accountExternalLoggingV2{}

func (r *accountExternalLoggingV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_external_logging_v2"
}

func (r *accountExternalLoggingV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.AccountExternalLoggingV2ResourceSchema()
}

// ImportState reads the account-level log service. The account has at most one, so the import ID
// is not used for the lookup; the log service ID is the conventional value.
func (r *accountExternalLoggingV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	response, err := core.ReadAccountExternalLogging(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix)
	if err != nil || response.Data.Id == "" {
		resp.Diagnostics.AddError(
			"Unable to Import Account External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	meta, err := r.logServiceMetadata(ctx, response.Data.Service)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Log Service Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", response.Data.Service, err),
		)
		return
	}

	data := model.AccountExternalLoggingV2ResourceModel{
		RunSetupTests: types.BoolValue(false),
		Config:        types.DynamicNull(),
	}
	mask := model.WithoutMaskedValues(response.Data.Config)
	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, mask)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"fivetran_account_external_logging_v2 import requires HCL review",
		"Terraform imported the log service using API-visible config fields. Add the matching fivetran_account_external_logging_v2 resource block to your configuration and restore secret values, because the API masks them.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountExternalLoggingV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.AccountExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.logServiceMetadata(ctx, data.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Log Service Metadata.",
			fmt.Sprintf("Unable to fetch metadata for service %q: %v", data.Service.ValueString(), err),
		)
		return
	}

	configMap, diags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := core.GetBoolOrDefault(data.Enabled, true)
	runSetupTests := core.GetBoolOrDefault(data.RunSetupTests, false)

	response, err := core.CreateAccountExternalLogging(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix, core.AccountExternalLoggingCreateRequest{
		Service:       data.Service.ValueString(),
		Enabled:       &enabled,
		Config:        configMap,
		RunSetupTests: &runSetupTests,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Account External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMap)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RunSetupTests = types.BoolValue(runSetupTests)

	warnFailedLogSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountExternalLoggingV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.AccountExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := core.ReadAccountExternalLogging(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix)
	if err != nil {
		if strings.HasPrefix(response.Code, "NotFound") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Account External Logging Resource.",
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
		resp.Diagnostics.AddError(
			"Unable to Fetch Log Service Metadata.",
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

	resp.Diagnostics.Append(data.ReadFromResponse(ctx, response, meta, configMask)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RunSetupTests = runSetupTests

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *accountExternalLoggingV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state model.AccountExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.logServiceMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Fetch Log Service Metadata.",
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

	var slot *metadata.Property
	if meta != nil {
		slot = &meta.Config
	}
	configPatch := core.PrepareConfigPatchDynamic(planConfig, stateConfig, slot)

	enabledPlan := core.GetBoolOrDefault(plan.Enabled, true)
	enabledChanged := enabledPlan != core.GetBoolOrDefault(state.Enabled, true)
	runSetupTestsPlan := core.GetBoolOrDefault(plan.RunSetupTests, false)
	runSetupTestsRequested := runSetupTestsPlan && !core.GetBoolOrDefault(state.RunSetupTests, false)

	hasChanges := len(configPatch) > 0 || enabledChanged

	switch {
	case hasChanges:
		updateReq := core.AccountExternalLoggingUpdateRequest{
			RunSetupTests: &runSetupTestsPlan,
		}
		if enabledChanged {
			updateReq.Enabled = &enabledPlan
		}
		if len(configPatch) > 0 {
			updateReq.Config = configPatch
		}

		response, err := core.UpdateAccountExternalLogging(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix, updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Account External Logging Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}

		resp.Diagnostics.Append(state.ReadFromResponse(ctx, response, meta, planConfig)...)
		if resp.Diagnostics.HasError() {
			return
		}
		warnFailedLogSetupTests(response.Data.SetupTests, &resp.Diagnostics)

	case runSetupTestsRequested:
		// The API rejects a PATCH without update parameters, so tests are triggered through
		// the dedicated setup-tests endpoint.
		response, err := core.RunAccountExternalLoggingSetupTests(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Run Account External Logging Setup Tests.",
				fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
			)
			return
		}
		warnFailedLogSetupTests(response.Data.SetupTests, &resp.Diagnostics)
	}

	state.RunSetupTests = types.BoolValue(runSetupTestsPlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *accountExternalLoggingV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	deleteResponse, err := core.DeleteAccountExternalLogging(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix)
	if err != nil && !strings.HasPrefix(deleteResponse.Code, "NotFound") {
		resp.Diagnostics.AddError(
			"Unable to Delete Account External Logging Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
	}
}

func (r *accountExternalLoggingV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.GetSkipPlanTimeValidation() {
		return
	}

	var data model.AccountExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Service.IsNull() || data.Service.IsUnknown() || data.Service.ValueString() == "" {
		return
	}

	configMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || len(configMap) == 0 {
		return
	}

	meta, err := r.logServiceMetadata(ctx, data.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Validate Account External Logging Configuration",
			fmt.Sprintf("Unable to fetch metadata for service %q. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", data.Service.ValueString(), err),
		)
		return
	}
	if meta == nil {
		// The metadata API does not describe this log service; the API validates the config.
		return
	}

	validateDynamicObject(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
}

func (r *accountExternalLoggingV2) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || r.GetSkipPlanTimeValidation() {
		return
	}

	var plan model.AccountExternalLoggingV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Service.IsNull() || plan.Service.IsUnknown() || plan.Service.ValueString() == "" {
		return
	}

	planConfig, diags := core.DynamicToMapPreserveUnknown(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isCreate := req.State.Raw.IsNull()
	var state model.AccountExternalLoggingV2ResourceModel
	stateConfig := map[string]interface{}(nil)
	if !isCreate {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		stateConfig, diags = core.DynamicToMapPreserveUnknown(ctx, state.Config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(planConfig) == 0 {
			return
		}
	}

	meta, err := r.logServiceMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Modify Account External Logging Plan",
			fmt.Sprintf("Unable to fetch metadata for service %q. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", plan.Service.ValueString(), err),
		)
		return
	}
	if meta == nil {
		return
	}

	if isCreate {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics)
		return
	}

	if appendImmutableRequiresReplace(planConfig, stateConfig, &meta.Config, path.Root("config"), resp, plan.Service.ValueString(), "") {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics)
	}
}

// logServiceMetadata returns the metadata describing a log service from
// /metadata/log-service-types/{service}, or nil when the metadata API has no description for it
// (for example a server that does not serve the endpoint yet). Callers must treat nil as
// "no plan-time validation available".
func (r *accountExternalLoggingV2) logServiceMetadata(ctx context.Context, service string) (*metadata.ConnectorMetadata, error) {
	cache := r.GetMetadataCache()
	if cache == nil {
		cache = &sync.Map{}
	}

	// Log and connector metadata share the provider cache, so log entries get their own key space.
	cacheKey := logServiceMetadataCachePrefix + service
	if _, ok := cache.Load(noMetadataCachePrefix + cacheKey); ok {
		return nil, nil
	}
	if meta, ok, err := core.LoadCachedConnectorMetadata(cache, cacheKey); ok || err != nil {
		return meta, err
	}
	if r.GetClient() == nil {
		return nil, errUnconfiguredClient
	}

	resp, err := core.ReadLogServiceMetadata(ctx, r.GetClient(), accountExternalLoggingV2UserAgentSuffix, service)
	if err != nil {
		if strings.HasPrefix(resp.Code, "NotFound") {
			cache.Store(noMetadataCachePrefix+cacheKey, true)
			return nil, nil
		}
		return nil, err
	}

	meta := resp.Data.ConnectorMetadata
	actual, _ := cache.LoadOrStore(cacheKey, &meta)
	cached, ok := actual.(*metadata.ConnectorMetadata)
	if !ok {
		return nil, fmt.Errorf("metadata cache contained unexpected type %T for log service %q", actual, service)
	}
	return cached, nil
}

func warnFailedLogSetupTests(setupTests []common.SetupTestResponse, diags *diag.Diagnostics) {
	for _, tr := range setupTests {
		if tr.Status != "PASSED" && tr.Status != "SKIPPED" {
			diags.AddWarning(
				fmt.Sprintf("Log service setup test `%v` has status `%v`", tr.Title, tr.Status),
				tr.Message,
			)
		}
	}
}
