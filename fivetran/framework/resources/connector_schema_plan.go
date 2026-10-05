package resources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithModifyPlan = &connectorSchema{}

// ModifyPlan resolves the connector by group_id + connector_name on update: connector_id is computed and the plan
// keeps the value from state, which is outdated when the connector was recreated.
func (r *connectorSchema) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// nothing to resolve on create (Create resolves it) or destroy
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var config, plan model.ConnectorSchemaResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.ConnectorId.IsNull() ||
		config.GroupId.IsNull() || config.GroupId.IsUnknown() ||
		config.ConnectorName.IsNull() || config.ConnectorName.IsUnknown() {
		return
	}

	client := r.GetClient()
	if client == nil {
		return
	}

	// lookup failures are reported by Update
	connectorID, err := findConnectorIdByGroupAndSchemaName(ctx, client, &config)
	if err != nil || connectorID == plan.ConnectorId.ValueString() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("connector_id"), connectorID)...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), connectorID)...)
}
