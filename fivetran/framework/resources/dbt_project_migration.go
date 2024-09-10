package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var projectConfigAttr = map[string]tftypes.Type{
	"git_remote_url":   tftypes.String,
	"git_branch":  		tftypes.String,
	"folder_path": 		tftypes.String,
}

func upgradeDbtProjectState(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse, fromVersion int) {
	rawStateValue, err := req.RawState.Unmarshal(getDbtProjectStateModel(fromVersion))

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Unmarshal Prior State",
			err.Error(),
		)
		return
	}

	var rawState map[string]tftypes.Value

	if err := rawStateValue.As(&rawState); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Convert Prior State",
			err.Error(),
		)
		return
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		getDbtProjectStateModel(1),
		tftypes.NewValue(getDbtProjectStateModel(1), map[string]tftypes.Value{
			"id": 				rawState["id"],
			"group_id":			rawState["group_id"],
			"default_schema":	rawState["default_schema"],
			"dbt_version":		rawState["dbt_version"],
			"target_name":		rawState["target_name"],
			"threads":			rawState["threads"],
			"type":				rawState["type"],
			"status":			rawState["status"],
			"created_at":		rawState["created_at"],
			"created_by_id":	rawState["created_by_id"],
			"public_key":		rawState["public_key"],
			"environment_vars":	rawState["environment_vars"],
			"ensure_readiness": rawState["ensure_readiness"],
			"timeouts":                  rawState["timeouts"],
			"models":			rawState["models"],
			"project_config": 	convertSetToBlock("project_config", 
									rawState["project_config"], 
									projectConfigAttr,
									projectConfigAttr, 
									resp.Diagnostics),
		}),
	)

	resp.DynamicValue = &dynamicValue
}

func getDbtProjectStateModel(version int) tftypes.Type {
	base := map[string]tftypes.Type{
		"id":           	tftypes.String,
		"group_id":         tftypes.String,
		"default_schema":	tftypes.String,
		"dbt_version":   	tftypes.String,
		"target_name":     	tftypes.String,
		"threads":      	tftypes.Number,
		"type":     		tftypes.String,
		"status":     		tftypes.String,
		"created_at":     	tftypes.String,
		"created_by_id":    tftypes.String,
		"public_key":     	tftypes.String,
		"ensure_readiness": tftypes.Bool,
		"environment_vars": tftypes.Set{ElementType:tftypes.String},
		"timeouts": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"create": tftypes.String,
			},
		},
		"models":  			tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"id":        tftypes.String,
						"model_name": tftypes.String,
						"scheduled":         tftypes.Bool,
					},
				},
			},
		}

	if version == 0 {
		base["project_config"] = tftypes.Set{ElementType: tftypes.Object{AttributeTypes: projectConfigAttr}}
	} else if version == 1 {
		base["project_config"] = tftypes.Object{AttributeTypes: projectConfigAttr}
	}


	return tftypes.Object{AttributeTypes: base}
}