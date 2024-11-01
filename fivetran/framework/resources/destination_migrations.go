package resources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func upgradeDestinationState(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse, fromVersion int) {
	rawStateValue, err := req.RawState.Unmarshal(getDestinationStateModel(fromVersion))

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
		getDestinationStateModel(2),
		tftypes.NewValue(getDestinationStateModel(2), map[string]tftypes.Value{
			"id":                           rawState["id"],
			"group_id":                     rawState["group_id"],
			"service":                      rawState["service"],
			"region":                       rawState["region"],
			"timeouts":                     rawState["timeouts"],
			"time_zone_offset":             rawState["time_zone_offset"],
			"setup_status":                 rawState["setup_status"],
			"daylight_saving_time_enabled": tftypes.NewValue(tftypes.Bool, nil),
			"local_processing_agent_id":    tftypes.NewValue(tftypes.String, nil),
			"networking_method":            tftypes.NewValue(tftypes.String, nil),
            "private_link_id":              tftypes.NewValue(tftypes.String, nil),

			"run_setup_tests":    convertStringStateValueToBool("run_setup_tests", rawState["run_setup_tests"], resp.Diagnostics),
			"trust_fingerprints": convertStringStateValueToBool("trust_fingerprints", rawState["trust_fingerprints"], resp.Diagnostics),
			"trust_certificates": convertStringStateValueToBool("trust_certificates", rawState["trust_certificates"], resp.Diagnostics),

			"config": convertSetToBlock(
				"config",
				rawState["config"],
				model.GetTfTypesDestination(common.GetDestinationFieldsMap(), 1),
				model.GetTfTypesDestination(common.GetDestinationFieldsMap(), fromVersion), resp.Diagnostics),
		}),
	)

	resp.DynamicValue = &dynamicValue
}
func getDestinationStateModel(version int) tftypes.Type {
	base := map[string]tftypes.Type{
		"id":               tftypes.String,
		"group_id":         tftypes.String,
		"service":          tftypes.String,
		"region":           tftypes.String,
		"time_zone_offset": tftypes.String,
		"setup_status":     tftypes.String,

		"timeouts": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"create": tftypes.String,
				"update": tftypes.String,
			},
		},
	}
	if version == 1 ||  version == 2 {
		base["run_setup_tests"] = tftypes.Bool
		base["trust_certificates"] = tftypes.Bool
		base["trust_fingerprints"] = tftypes.Bool
		base["daylight_saving_time_enabled"] = tftypes.Bool
		base["local_processing_agent_id"] = tftypes.String
		base["networking_method"] = tftypes.String

		base["config"] = tftypes.Object{AttributeTypes: model.GetTfTypesDestination(common.GetDestinationFieldsMap(), 1)}
	} else {
		base["run_setup_tests"] = tftypes.String
		base["trust_certificates"] = tftypes.String
		base["trust_fingerprints"] = tftypes.String
		base["last_updated"] = tftypes.String

		base["config"] = tftypes.Set{ElementType: tftypes.Object{AttributeTypes: model.GetTfTypesDestination(common.GetDestinationFieldsMap(), 0)}}
	}

	return tftypes.Object{AttributeTypes: base}
}
