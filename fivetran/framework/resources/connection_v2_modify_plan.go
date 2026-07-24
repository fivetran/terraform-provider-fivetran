package resources

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *connectionV2) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Service.IsNull() || plan.Service.IsUnknown() || plan.Service.ValueString() == "" {
		return
	}

	isCreate := req.State.Raw.IsNull()
	isRootReplacement := false
	var state model.ConnectionV2ResourceModel
	if !isCreate {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		isRootReplacement = warnRootReplacementIfChanged("service", plan.Service, state.Service, &resp.Diagnostics)
		isRootReplacement = warnRootReplacementIfChanged("group_id", plan.GroupId, state.GroupId, &resp.Diagnostics) || isRootReplacement
	}

	if r.GetSkipPlanTimeValidation() {
		return
	}

	planConfig, planAuth := r.dynamicValidationMaps(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	stateConfig := map[string]interface{}(nil)
	stateAuth := map[string]interface{}(nil)
	if !isCreate {
		stateConfig, stateAuth = r.dynamicValidationMaps(ctx, state, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !isCreate && !isRootReplacement && len(planConfig) == 0 && len(planAuth) == 0 {
		return
	}

	meta, err := r.connectorMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Modify Connection V2 Plan",
			fmt.Sprintf("Unable to fetch metadata for service %q. Terraform cannot safely evaluate required or immutable dynamic config/auth fields without metadata. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", plan.Service.ValueString(), err),
		)
		return
	}

	if isCreate {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics)
		validateRequiredDynamicFields(planAuth, &meta.Auth, path.Root("auth"), &resp.Diagnostics)
		return
	}

	if isRootReplacement {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics)
		validateRequiredDynamicFields(planAuth, &meta.Auth, path.Root("auth"), &resp.Diagnostics)
		return
	}

	hasImmutableReplacement := appendImmutableRequiresReplace(planConfig, stateConfig, &meta.Config, path.Root("config"), resp)
	hasImmutableReplacement = appendImmutableRequiresReplace(planAuth, stateAuth, &meta.Auth, path.Root("auth"), resp) || hasImmutableReplacement
	if hasImmutableReplacement {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics)
		validateRequiredDynamicFields(planAuth, &meta.Auth, path.Root("auth"), &resp.Diagnostics)
	}
}

func warnRootReplacementIfChanged(name string, planValue, stateValue types.String, diags *diag.Diagnostics) bool {
	if planValue.IsNull() || planValue.IsUnknown() || stateValue.IsNull() || stateValue.IsUnknown() {
		return false
	}
	if !planValue.Equal(stateValue) {
		addHistoricalResyncWarning(path.Root(name), diags)
		return true
	}
	return false
}

func validateRequiredDynamicFields(values map[string]interface{}, slot *metadata.Property, root path.Path, diags *diag.Diagnostics) {
	if slot == nil {
		return
	}

	for _, name := range slot.Required {
		if _, ok := values[name]; !ok {
			diags.AddAttributeError(
				root.AtName(name),
				"Missing Required Dynamic Field",
				fmt.Sprintf("Connector metadata requires field %q when creating this connection.", name),
			)
		}
	}

	for name, value := range values {
		prop := core.SlotProp(slot, name)
		if prop == nil || core.IsDynamicUnknownValue(value) {
			continue
		}
		fieldPath := root.AtName(name)
		validateNestedRequiredDynamicFields(value, prop, fieldPath, diags)
	}
}

func validateNestedRequiredDynamicFields(value interface{}, prop *metadata.Property, valuePath path.Path, diags *diag.Diagnostics) {
	switch v := value.(type) {
	case map[string]interface{}:
		validateRequiredDynamicFields(v, prop, valuePath, diags)
	case []interface{}:
		if prop.Items == nil {
			return
		}
		for i, item := range v {
			if core.IsDynamicUnknownValue(item) {
				continue
			}
			validateNestedRequiredDynamicFields(item, prop.Items, valuePath.AtListIndex(i), diags)
		}
	}
}

func appendImmutableRequiresReplace(planValues, stateValues map[string]interface{}, slot *metadata.Property, root path.Path, resp *resource.ModifyPlanResponse) bool {
	if slot == nil {
		return false
	}

	hasReplacement := false
	for name, planValue := range planValues {
		prop := core.SlotProp(slot, name)
		if prop == nil || core.IsDynamicUnknownValue(planValue) {
			continue
		}

		fieldPath := root.AtName(name)
		stateValue, stateHasValue := stateValues[name]
		if prop.Immutable && (!stateHasValue || !dynamicValuesEqual(planValue, stateValue)) {
			resp.RequiresReplace.Append(fieldPath)
			addHistoricalResyncWarning(fieldPath, &resp.Diagnostics)
			hasReplacement = true
			continue
		}

		hasReplacement = appendNestedImmutableRequiresReplace(planValue, stateValue, prop, fieldPath, resp) || hasReplacement
	}
	return hasReplacement
}

func appendNestedImmutableRequiresReplace(planValue, stateValue interface{}, prop *metadata.Property, fieldPath path.Path, resp *resource.ModifyPlanResponse) bool {
	switch planNested := planValue.(type) {
	case map[string]interface{}:
		stateNested, _ := stateValue.(map[string]interface{})
		return appendImmutableRequiresReplace(planNested, stateNested, prop, fieldPath, resp)
	case []interface{}:
		if prop.Items == nil {
			return false
		}
		stateItems, _ := stateValue.([]interface{})
		hasReplacement := false
		for i, item := range planNested {
			if core.IsDynamicUnknownValue(item) {
				continue
			}
			var stateItem interface{}
			if i < len(stateItems) {
				stateItem = stateItems[i]
			}
			hasReplacement = appendNestedImmutableRequiresReplace(item, stateItem, prop.Items, fieldPath.AtListIndex(i), resp) || hasReplacement
		}
		return hasReplacement
	}
	return false
}

func addHistoricalResyncWarning(fieldPath path.Path, diags *diag.Diagnostics) {
	diags.AddAttributeWarning(
		fieldPath,
		"Connection Replacement Triggers Historical Resync",
		fmt.Sprintf("Changing %s requires recreating this connection. Recreating a connection triggers a historical resync: all data re-syncs from scratch, can take days or weeks depending on data volume, and is not reversible.", fieldPath.String()),
	)
}

func dynamicValuesEqual(left, right interface{}) bool {
	if core.IsDynamicUnknownValue(left) || core.IsDynamicUnknownValue(right) {
		return true
	}
	return reflect.DeepEqual(left, right)
}
