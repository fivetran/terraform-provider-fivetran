package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// importedConfigFieldsKey is the private state key that lists the top-level config fields
// stored in state by import. Import tracks every field the API returns, including defaults the
// user never wrote. State alone can't tell those apart from fields the user removed from the
// configuration, and an update would send null for a removed nullable field, clearing a value
// the user never managed. Each successful update drops the fields that are now declared or no
// longer tracked from the list (see updateImportedConfigFields).
const importedConfigFieldsKey = "imported_config_fields"

type privateStateGetter interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

type privateStateSetter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// storeImportedConfigFields records the top-level keys of the imported config in private state.
func storeImportedConfigFields(ctx context.Context, private privateStateSetter, config types.Dynamic) diag.Diagnostics {
	configMap, diags := core.DynamicToMap(ctx, config)
	if diags.HasError() {
		return diags
	}
	keys := make([]string, 0, len(configMap))
	for k := range configMap {
		keys = append(keys, k)
	}
	diags.Append(setImportedConfigFields(ctx, private, keys)...)
	return diags
}

// setImportedConfigFields writes the import record, or removes it when keys is empty.
func setImportedConfigFields(ctx context.Context, private privateStateSetter, keys []string) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(keys) == 0 {
		return private.SetKey(ctx, importedConfigFieldsKey, nil)
	}
	sort.Strings(keys)
	value, err := json.Marshal(keys)
	if err != nil {
		diags.AddError("Unable to Store Imported Config Fields", err.Error())
		return diags
	}
	diags.Append(private.SetKey(ctx, importedConfigFieldsKey, value)...)
	return diags
}

// loadImportedConfigFields returns the config keys recorded by import, or nil when the resource
// was not imported or has been updated since.
func loadImportedConfigFields(ctx context.Context, private privateStateGetter) (map[string]bool, diag.Diagnostics) {
	if private == nil {
		return nil, nil
	}
	value, diags := private.GetKey(ctx, importedConfigFieldsKey)
	if diags.HasError() || len(value) == 0 {
		return nil, diags
	}
	var keys []string
	if err := json.Unmarshal(value, &keys); err != nil {
		diags.AddError("Unable to Read Imported Config Fields", err.Error())
		return nil, diags
	}
	result := make(map[string]bool, len(keys))
	for _, k := range keys {
		result[k] = true
	}
	return result, diags
}

// updateImportedConfigFields updates the import record after a successful update. It keeps the
// imported fields that are still in the new state but not set in the configuration. That happens
// when config is omitted and inherits its prior state (fivetran_connection_v2), so those fields are
// still undeclared and a later partial config must not clear them. Every other imported field is
// either declared now or no longer tracked, so it is dropped from the record. When config is
// required, the new state holds only configured fields and the record is removed.
func updateImportedConfigFields(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) diag.Diagnostics {
	imported, diags := loadImportedConfigFields(ctx, req.Private)
	if diags.HasError() || len(imported) == 0 {
		return diags
	}

	var configured, newState types.Dynamic
	diags.Append(req.Config.GetAttribute(ctx, path.Root("config"), &configured)...)
	diags.Append(resp.State.GetAttribute(ctx, path.Root("config"), &newState)...)
	if diags.HasError() {
		return diags
	}
	configuredMap, d := core.DynamicToMapPreserveUnknown(ctx, configured)
	diags.Append(d...)
	newStateMap, d := core.DynamicToMap(ctx, newState)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	var keep []string
	for k := range newStateMap {
		if _, declared := configuredMap[k]; imported[k] && !declared {
			keep = append(keep, k)
		}
	}
	diags.Append(setImportedConfigFields(ctx, resp.Private, keep)...)
	return diags
}

// undeclaredImportedConfigFields returns, sorted, the imported fields that are still in state but
// not in the planned config: fields the user never declared after import.
func undeclaredImportedConfigFields(imported map[string]bool, plan, state map[string]interface{}) []string {
	var fields []string
	for k := range state {
		if !imported[k] {
			continue
		}
		if _, inPlan := plan[k]; inPlan {
			continue
		}
		fields = append(fields, k)
	}
	sort.Strings(fields)
	return fields
}

// omitUndeclaredImportedConfigFields drops from an update patch the nulls that
// PrepareConfigPatchDynamic emits for imported fields missing from the configuration, so
// Terraform stops tracking them without clearing them upstream. An explicit null in the
// configuration keeps the field in the plan, so it is still sent.
func omitUndeclaredImportedConfigFields(patch map[string]interface{}, fields []string) {
	for _, k := range fields {
		if v, ok := patch[k]; ok && v == nil {
			delete(patch, k)
		}
	}
}

// importedConfigPatchFilter loads the import record and returns the imported fields that are
// in state but not in the planned config.
func importedConfigPatchFilter(ctx context.Context, private privateStateGetter, plan, state map[string]interface{}, diags *diag.Diagnostics) []string {
	imported, d := loadImportedConfigFields(ctx, private)
	diags.Append(d...)
	if len(imported) == 0 {
		return nil
	}
	return undeclaredImportedConfigFields(imported, plan, state)
}

// planUndeclaredImportedConfig keeps the prior config in the plan when the configuration differs
// from it only by imported fields it doesn't declare, so those fields don't show as removed.
// Terraform accepts a planned value that differs from the configuration when it is exactly the
// prior value. The fields stay in state, are refreshed by Read, and are never sent to the API.
// It returns true when it planned the prior config; the caller then defers
// restoreUnchangedImportedPlan so that an import with no other change plans nothing.
//
// When config has other changes too (another field changes, a field moves to config_secrets, or a
// value is unknown until apply), the plan must follow the configuration: the undeclared fields show
// as removed, the update stops tracking them without clearing them, and a warning says so.
// objectName names the managed object in the warning, for example "destination".
func planUndeclaredImportedConfig(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, objectName string) bool {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return false
	}
	imported, diags := loadImportedConfigFields(ctx, req.Private)
	if diags.HasError() || len(imported) == 0 {
		return false
	}

	var planConfig, stateConfig types.Dynamic
	if d := req.Plan.GetAttribute(ctx, path.Root("config"), &planConfig); d.HasError() || planConfig.IsUnknown() || planConfig.IsUnderlyingValueUnknown() {
		return false
	}
	if d := req.State.GetAttribute(ctx, path.Root("config"), &stateConfig); d.HasError() {
		return false
	}
	planMap, d := core.DynamicToMapPreserveUnknown(ctx, planConfig)
	if d.HasError() {
		return false
	}
	stateMap, d := core.DynamicToMapPreserveUnknown(ctx, stateConfig)
	if d.HasError() {
		return false
	}

	// A field set in config_secrets is declared: it moves out of config, which is a real change.
	var secretsDiags diag.Diagnostics
	secrets, secretsUnknown := connectionV2SecretsFromConfig(ctx, req.Config, configSecretsPath, true, &secretsDiags)
	if secretsDiags.HasError() {
		return false
	}
	var fields []string
	for _, k := range undeclaredImportedConfigFields(imported, planMap, stateMap) {
		if _, inSecrets := secrets[k]; !inSecrets {
			fields = append(fields, k)
		}
	}
	if len(fields) == 0 {
		return false
	}

	// A wholly unknown config_secrets may move any field out of config.
	if !secretsUnknown && configWithFieldsEqualsState(planMap, stateMap, fields) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("config"), stateConfig)...)
		return !resp.Diagnostics.HasError()
	}

	resp.Diagnostics.AddAttributeWarning(
		path.Root("config"),
		"Imported Config Fields Not in Configuration",
		fmt.Sprintf("These config fields were imported but are not set in your configuration: %s. "+
			"Because config has other changes in this plan, the plan shows them as removed, but Terraform will only stop tracking them and leave their values on the %s unchanged. "+
			"To keep managing a field, add it to config. To clear a field on the %s, set it to null in config.",
			strings.Join(fields, ", "), objectName, objectName),
	)
	return false
}

// configWithFieldsEqualsState reports whether the planned config, with the given fields copied
// from state, holds exactly the state values. A planned value that is still unknown never matches.
func configWithFieldsEqualsState(plan, state map[string]interface{}, fields []string) bool {
	if len(plan)+len(fields) != len(state) {
		return false
	}
	for k, v := range plan {
		stateValue, ok := state[k]
		if !ok || containsDynamicUnknown(v) || !reflect.DeepEqual(v, stateValue) {
			return false
		}
	}
	return true
}

func containsDynamicUnknown(value interface{}) bool {
	switch v := value.(type) {
	case core.DynamicUnknownValue:
		return true
	case map[string]interface{}:
		for _, item := range v {
			if containsDynamicUnknown(item) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if containsDynamicUnknown(item) {
				return true
			}
		}
	}
	return false
}

// restoreUnchangedImportedPlan plans the prior state when, after planUndeclaredImportedConfig kept
// the prior config, nothing else changes. Terraform's proposed plan differed from state in config,
// so the framework already marked every computed attribute that is null in the configuration as
// unknown, which would still plan an in-place update. It does nothing when the plan has errors, a
// replacement, or any other change.
func restoreUnchangedImportedPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if resp.Diagnostics.HasError() || len(resp.RequiresReplace) > 0 {
		return
	}
	if planUnchangedExceptComputedUnknowns(resp.Plan.Raw, req.Config.Raw, req.State.Raw) {
		resp.Plan.Raw = req.State.Raw.Copy()
	}
}

// planUnchangedExceptComputedUnknowns reports whether plan equals state, except for unknown values
// at paths that are null in the configuration (computed attributes the framework marked unknown).
// An unknown value that comes from the configuration is a real change.
func planUnchangedExceptComputedUnknowns(plan, config, state tftypes.Value) bool {
	if plan.Equal(state) {
		return true
	}
	if !plan.IsKnown() {
		return config.IsNull()
	}
	if plan.IsNull() || state.IsNull() || !state.IsKnown() || !plan.Type().Is(tftypes.Object{}) || !state.Type().Is(tftypes.Object{}) {
		return false
	}

	var planAttrs, stateAttrs, configAttrs map[string]tftypes.Value
	if err := plan.As(&planAttrs); err != nil {
		return false
	}
	if err := state.As(&stateAttrs); err != nil {
		return false
	}
	if !config.IsNull() {
		if !config.IsKnown() || !config.Type().Is(tftypes.Object{}) {
			return false
		}
		if err := config.As(&configAttrs); err != nil {
			return false
		}
	}
	if len(planAttrs) != len(stateAttrs) {
		return false
	}
	for name, planValue := range planAttrs {
		stateValue, ok := stateAttrs[name]
		if !ok {
			return false
		}
		configValue, ok := configAttrs[name]
		if !ok {
			configValue = tftypes.NewValue(planValue.Type(), nil)
		}
		if !planUnchangedExceptComputedUnknowns(planValue, configValue, stateValue) {
			return false
		}
	}
	return true
}
