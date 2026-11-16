package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// warnUndeclaredImportedConfigFields warns at plan time that imported config fields missing from
// the configuration will only stop being tracked, not be cleared. objectName names the managed
// object in the message, for example "destination".
func warnUndeclaredImportedConfigFields(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, objectName string) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	imported, diags := loadImportedConfigFields(ctx, req.Private)
	if diags.HasError() || len(imported) == 0 {
		return
	}

	var planConfig, stateConfig types.Dynamic
	if d := req.Plan.GetAttribute(ctx, path.Root("config"), &planConfig); d.HasError() || planConfig.IsUnknown() || planConfig.IsUnderlyingValueUnknown() {
		return
	}
	if d := req.State.GetAttribute(ctx, path.Root("config"), &stateConfig); d.HasError() {
		return
	}
	planMap, d := core.DynamicToMapPreserveUnknown(ctx, planConfig)
	if d.HasError() {
		return
	}
	stateMap, d := core.DynamicToMap(ctx, stateConfig)
	if d.HasError() {
		return
	}

	fields := undeclaredImportedConfigFields(imported, planMap, stateMap)
	if len(fields) == 0 {
		return
	}
	resp.Diagnostics.AddAttributeWarning(
		path.Root("config"),
		"Imported Config Fields Not in Configuration",
		fmt.Sprintf("These config fields were imported but are not set in your configuration: %s. "+
			"Terraform will stop tracking them and leave their values on the %s unchanged, although the plan shows them as removed. "+
			"To keep managing a field, add it to config. To clear a field on the %s, set it to null in config.",
			strings.Join(fields, ", "), objectName, objectName),
	)
}
