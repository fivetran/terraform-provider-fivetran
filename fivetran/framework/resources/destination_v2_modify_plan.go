package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *destinationV2) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Service.IsNull() || plan.Service.IsUnknown() || plan.Service.ValueString() == "" {
		return
	}

	if r.GetSkipPlanTimeValidation() {
		return
	}

	isCreate := req.State.Raw.IsNull()
	var state model.DestinationV2ResourceModel
	if !isCreate {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	// service and group_id carry schema-level RequiresReplace; metadata only
	// needs to be consulted for required fields when a new destination is created.
	isRootReplacement := !isCreate && (fieldChanged(plan.Service, state.Service) || fieldChanged(plan.GroupId, state.GroupId))

	planConfig, diags := core.DynamicToMapPreserveUnknown(ctx, plan.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateConfig := map[string]interface{}(nil)
	if !isCreate {
		stateConfig, diags = core.DynamicToMapPreserveUnknown(ctx, state.Config)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !isCreate && !isRootReplacement && len(planConfig) == 0 {
		return
	}

	meta, err := r.destinationMetadata(ctx, plan.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Modify Destination V2 Plan",
			fmt.Sprintf("Unable to fetch metadata for service %q. Terraform cannot safely evaluate required or immutable dynamic config fields without metadata. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", plan.Service.ValueString(), err),
		)
		return
	}

	if isCreate || isRootReplacement {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics, destinationWording)
		return
	}

	if appendImmutableRequiresReplace(planConfig, stateConfig, &meta.Config, path.Root("config"), resp, plan.Service.ValueString(), "", destinationWording) {
		validateRequiredDynamicFields(planConfig, &meta.Config, path.Root("config"), &resp.Diagnostics, destinationWording)
	}
}

var destinationWording = dynamicWording{
	missingRequiredField: func(name string) string {
		return fmt.Sprintf("Destination metadata requires field %q when creating this destination.", name)
	},
	replacementWarning: func(fieldPath path.Path, diags *diag.Diagnostics) {
		diags.AddAttributeWarning(
			fieldPath,
			"Destination Replacement",
			fmt.Sprintf("Changing %s requires recreating this destination: Terraform will delete the existing destination and create a new one with the planned configuration.", fieldPath.String()),
		)
	},
}
