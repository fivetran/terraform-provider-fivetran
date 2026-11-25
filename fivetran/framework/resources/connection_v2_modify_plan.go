package resources

import (
	"context"
	"errors"
	"fmt"

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
	warnUndeclaredImportedConfigFields(ctx, req, resp, "connection")

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

	secrets, secretsUnknown := connectionV2SecretsFromConfig(ctx, req.Config, configSecretsPath, true, &resp.Diagnostics)
	// auth is not computed, so a field moved to auth_secrets is absent from the plan already and
	// needs no state cleanup here.
	authSecrets, authSecretsUnknown := connectionV2SecretsFromConfig(ctx, req.Config, authSecretsPath, true, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if !isCreate && len(secrets) > 0 {
		var configuredConfig types.Dynamic
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("config"), &configuredConfig)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Omitted optional/computed config inherits its prior state. Remove fields now
		// managed through config_secrets so migration produces an update even without
		// a version change. Keep other inherited fields in the state mask.
		if configuredConfig.IsNull() || configuredConfig.IsUnderlyingValueNull() {
			inheritedConfig, diags := core.DynamicToMap(ctx, state.Config)
			resp.Diagnostics.Append(diags...)
			moved := false
			for name := range secrets {
				if _, ok := inheritedConfig[name]; ok {
					delete(inheritedConfig, name)
					moved = true
				}
			}
			if moved {
				plan.Config, diags = core.MapToDynamic(ctx, inheritedConfig)
				resp.Diagnostics.Append(diags...)
				if !resp.Diagnostics.HasError() {
					resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("config"), plan.Config)...)
				}
			}
			if resp.Diagnostics.HasError() {
				return
			}
		}
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

	if !isCreate && !isRootReplacement && len(planConfig) == 0 && len(planAuth) == 0 && len(secrets) == 0 && len(authSecrets) == 0 {
		return
	}

	meta, err := r.connectorMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Modify Connection V2 Plan",
			fmt.Sprintf("Unable to fetch metadata for service %q. Terraform cannot safely evaluate required or immutable dynamic config/auth fields without metadata. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", plan.Service.ValueString(), err),
		)
		return
	}

	// A required field can also be set in config_secrets or auth_secrets, which are only in the configuration.
	requiredConfig := mergeConnectionV2Secrets(planConfig, secrets)
	requiredAuth := mergeConnectionV2Secrets(planAuth, authSecrets)

	// A wholly unknown secrets attribute may hold any required field, so defer that slot to the API.
	validateRequired := func() {
		if !secretsUnknown {
			validateRequiredDynamicFields(requiredConfig, &meta.Config, path.Root("config"), &resp.Diagnostics, connectionWording)
		}
		if !authSecretsUnknown {
			validateRequiredDynamicFields(requiredAuth, &meta.Auth, path.Root("auth"), &resp.Diagnostics, connectionWording)
		}
	}

	if isCreate {
		validateRequired()
		return
	}

	if isRootReplacement {
		validateRequired()
		return
	}

	// Until the first update after import, immutable fields the API didn't return are missing
	// from state. Compare them against the filled-in state so they don't force a replacement.
	compareConfig, compareAuth := stateConfig, stateAuth
	imported, diags := isConnectionV2Imported(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if imported {
		compareConfig = fillImportedImmutableFields(planConfig, stateConfig, &meta.Config, path.Root("config"), true, state.Name, &resp.Diagnostics)
		compareAuth = fillImportedImmutableFields(planAuth, stateAuth, &meta.Auth, path.Root("auth"), false, state.Name, &resp.Diagnostics)
	}

	service := plan.Service.ValueString()
	hasImmutableReplacement := appendImmutableRequiresReplace(planConfig, compareConfig, &meta.Config, path.Root("config"), resp, service, "", connectionWording)
	hasImmutableReplacement = appendMovedSecretsRequiresReplace(secrets, stateConfig, &meta.Config, configSecretsPath, resp, service) || hasImmutableReplacement
	hasImmutableReplacement = appendImmutableRequiresReplace(planAuth, compareAuth, &meta.Auth, path.Root("auth"), resp, service, "", connectionWording) || hasImmutableReplacement
	hasImmutableReplacement = appendMovedSecretsRequiresReplace(authSecrets, stateAuth, &meta.Auth, authSecretsPath, resp, service) || hasImmutableReplacement
	if hasImmutableReplacement {
		validateRequired()
	}
}

// appendMovedSecretsRequiresReplace plans a replacement when an immutable field moved from config
// or auth to its secrets attribute with a different value.
func appendMovedSecretsRequiresReplace(secrets, stateValues map[string]interface{}, slot *metadata.Property, secretsPath path.Path, resp *resource.ModifyPlanResponse, service string) bool {
	var secretsReplacement resource.ModifyPlanResponse
	replace := appendImmutableRequiresReplace(movedConnectionV2Secrets(secrets, stateValues), stateValues, slot, secretsPath, &secretsReplacement, service, "", connectionWording)
	if replace {
		// Write-only values are null in the plan, so nested replacement paths do not
		// exist there. Keep field-specific warnings, but replace at the attribute root.
		resp.RequiresReplace.Append(secretsPath)
	}
	resp.Diagnostics.Append(secretsReplacement.Diagnostics...)
	return replace
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

func validateRequiredDynamicFields(values map[string]interface{}, slot *metadata.Property, root path.Path, diags *diag.Diagnostics, wording dynamicWording) {
	if slot == nil {
		return
	}

	for _, name := range slot.Required {
		if _, ok := values[name]; !ok {
			diags.AddAttributeError(
				root.AtName(name),
				"Missing Required Dynamic Field",
				wording.missingRequiredField(name),
			)
		}
	}

	for name, value := range values {
		prop := core.SlotProp(slot, name)
		if prop == nil || core.IsDynamicUnknownValue(value) {
			continue
		}
		fieldPath := root.AtName(name)
		validateNestedRequiredDynamicFields(value, prop, fieldPath, diags, wording)
	}
}

func validateNestedRequiredDynamicFields(value interface{}, prop *metadata.Property, valuePath path.Path, diags *diag.Diagnostics, wording dynamicWording) {
	switch v := value.(type) {
	case map[string]interface{}:
		validateRequiredDynamicFields(v, prop, valuePath, diags, wording)
	case []interface{}:
		if prop.Items == nil {
			return
		}
		for i, item := range v {
			if core.IsDynamicUnknownValue(item) {
				continue
			}
			validateNestedRequiredDynamicFields(item, prop.Items, valuePath.AtListIndex(i), diags, wording)
		}
	}
}

// basePath is the dot-separated path of planValues within the dynamic field, without list indexes.
// It is used to recognize set-typed fields, see core.IsSetTypedField.
func appendImmutableRequiresReplace(planValues, stateValues map[string]interface{}, slot *metadata.Property, root path.Path, resp *resource.ModifyPlanResponse, service, basePath string, wording dynamicWording) bool {
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
		valuePath := name
		if basePath != "" {
			valuePath = basePath + "." + name
		}
		stateValue, stateHasValue := stateValues[name]
		if prop.Immutable && core.IsMaskedValue(stateValue) {
			// Imported secrets are stored masked, so the real value can't be compared.
			if !core.IsMaskedValue(planValue) {
				addMaskedImmutableWarning(fieldPath, &resp.Diagnostics)
			}
			continue
		}
		if prop.Immutable && (!stateHasValue || !dynamicValuesEqualAtPath(planValue, stateValue, service, valuePath)) {
			resp.RequiresReplace.Append(fieldPath)
			wording.replacementWarning(fieldPath, &resp.Diagnostics)
			hasReplacement = true
			continue
		}

		hasReplacement = appendNestedImmutableRequiresReplace(planValue, stateValue, prop, fieldPath, resp, service, valuePath, wording) || hasReplacement
	}
	return hasReplacement
}

func appendNestedImmutableRequiresReplace(planValue, stateValue interface{}, prop *metadata.Property, fieldPath path.Path, resp *resource.ModifyPlanResponse, service, valuePath string, wording dynamicWording) bool {
	switch planNested := planValue.(type) {
	case map[string]interface{}:
		stateNested, _ := stateValue.(map[string]interface{})
		return appendImmutableRequiresReplace(planNested, stateNested, prop, fieldPath, resp, service, valuePath, wording)
	case []interface{}:
		if prop.Items == nil {
			return false
		}
		// Items of a set-typed list have no stable position, so pairing them with state by index
		// would report a reordered item as a changed one and force a replacement.
		if core.IsSetTypedField(service, valuePath) {
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
			hasReplacement = appendNestedImmutableRequiresReplace(item, stateItem, prop.Items, fieldPath.AtListIndex(i), resp, service, valuePath, wording) || hasReplacement
		}
		return hasReplacement
	}
	return false
}

// dynamicWording holds the resource-specific text of the plan-time diagnostics that the resources
// with metadata-driven dynamic config share.
type dynamicWording struct {
	// missingRequiredField is the detail of the error for a required field that is absent on create.
	missingRequiredField func(name string) string
	// replacementWarning is added for a change that forces the resource to be replaced.
	replacementWarning func(fieldPath path.Path, diags *diag.Diagnostics)
}

var connectionWording = dynamicWording{
	missingRequiredField: func(name string) string {
		return fmt.Sprintf("Connector metadata requires field %q when creating this connection.", name)
	},
	replacementWarning: addHistoricalResyncWarning,
}

func addHistoricalResyncWarning(fieldPath path.Path, diags *diag.Diagnostics) {
	diags.AddAttributeWarning(
		fieldPath,
		"Connection Replacement Triggers Historical Resync",
		fmt.Sprintf("Changing %s requires recreating this connection. Recreating a connection triggers a historical resync: all data re-syncs from scratch, can take days or weeks depending on data volume, and is not reversible.", fieldPath.String()),
	)
}

func addMaskedImmutableWarning(fieldPath path.Path, diags *diag.Diagnostics) {
	diags.AddAttributeWarning(
		fieldPath,
		"Immutable Field Imported Masked",
		fmt.Sprintf("The API returned %s masked when the resource was imported, so Terraform can't verify that the configured value matches. The field is immutable: a different value is not sent to the API and does not recreate the resource.", fieldPath.String()),
	)
}

func dynamicValuesEqualAtPath(left, right interface{}, service, valuePath string) bool {
	if core.IsDynamicUnknownValue(left) || core.IsDynamicUnknownValue(right) {
		return true
	}
	return core.DynamicValuesEqualAtPath(left, right, service, valuePath)
}
