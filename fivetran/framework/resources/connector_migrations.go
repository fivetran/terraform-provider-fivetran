package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func upgradeConnectorState(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse, fromVersion int) {
	rawStateValue, err := req.RawState.Unmarshal(getConnectorStateModel(fromVersion))

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
		getConnectorStateModel(3),
		tftypes.NewValue(getConnectorStateModel(3), map[string]tftypes.Value{
			"id":           rawState["id"],
			"name":         rawState["name"],
			"connected_by": rawState["connected_by"],
			"created_at":   rawState["created_at"],
			"group_id":     rawState["group_id"],
			"service":      rawState["service"],
			"timeouts":     rawState["timeouts"],

			"run_setup_tests":    convertStringStateValueToBool("run_setup_tests", rawState["run_setup_tests"], resp.Diagnostics),
			"trust_fingerprints": convertStringStateValueToBool("trust_fingerprints", rawState["trust_fingerprints"], resp.Diagnostics),
			"trust_certificates": convertStringStateValueToBool("trust_certificates", rawState["trust_certificates"], resp.Diagnostics),

			"config": convertSetToBlock("config", rawState["config"], model.GetTfTypes(common.GetConfigFieldsMap(), 3), model.GetTfTypes(common.GetConfigFieldsMap(), fromVersion), resp.Diagnostics),
			"auth":   convertSetToBlock("auth", rawState["auth"], model.GetTfTypes(common.GetAuthFieldsMap(), 3), model.GetTfTypes(common.GetAuthFieldsMap(), fromVersion), resp.Diagnostics),
			"destination_schema": convertSetToBlock("destination_schema", rawState["destination_schema"],
				map[string]tftypes.Type{
					"name":   tftypes.String,
					"table":  tftypes.String,
					"prefix": tftypes.String,
				},
				map[string]tftypes.Type{
					"name":   tftypes.String,
					"table":  tftypes.String,
					"prefix": tftypes.String,
				},
				resp.Diagnostics),
		}),
	)

	resp.DynamicValue = &dynamicValue
}

func getConnectorStateModel(version int) tftypes.Type {
	dsObj := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":   tftypes.String,
			"table":  tftypes.String,
			"prefix": tftypes.String,
		},
	}
	base := map[string]tftypes.Type{
		"id":           tftypes.String,
		"name":         tftypes.String,
		"connected_by": tftypes.String,
		"created_at":   tftypes.String,
		"group_id":     tftypes.String,
		"service":      tftypes.String,

		"timeouts": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"create": tftypes.String,
				"update": tftypes.String,
			},
		},
	}
	if version == 3 {
		base["destination_schema"] = dsObj
		base["run_setup_tests"] = tftypes.Bool
		base["trust_certificates"] = tftypes.Bool
		base["trust_fingerprints"] = tftypes.Bool

		base["local_processing_agent_id"] = tftypes.String
		base["proxy_agent_id"] = tftypes.String
		base["private_link_id"] = tftypes.String
		base["networking_method"] = tftypes.String

		base["config"] = tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetConfigFieldsMap(), 3)}
		base["auth"] = tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetAuthFieldsMap(), 3)}
	} else {
		base["destination_schema"] = tftypes.Set{ElementType: dsObj}
		base["run_setup_tests"] = tftypes.String
		base["trust_certificates"] = tftypes.String
		base["trust_fingerprints"] = tftypes.String
		base["last_updated"] = tftypes.String

		base["config"] = tftypes.Set{ElementType: tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetConfigFieldsMap(), version)}}
		base["auth"] = tftypes.Set{ElementType: tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetAuthFieldsMap(), version)}}

		if version == 0 {
			base["sync_frequency"] = tftypes.String
			base["schedule_type"] = tftypes.String
			base["paused"] = tftypes.String
			base["pause_after_trial"] = tftypes.String
			base["daily_sync_time"] = tftypes.String
			base["succeeded_at"] = tftypes.String
			base["failed_at"] = tftypes.String
			base["service_version"] = tftypes.String

			base["status"] = tftypes.Set{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"setup_state":        tftypes.String,
						"is_historical_sync": tftypes.String,
						"sync_state":         tftypes.String,
						"update_state":       tftypes.String,
						"tasks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"code":    tftypes.String,
									"message": tftypes.String,
								},
							},
						},
						"warnings": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"code":    tftypes.String,
									"message": tftypes.String,
								},
							},
						},
					},
				},
			}
		}
	}

	return tftypes.Object{AttributeTypes: base}
}

func convertSetToBlock(field string, value tftypes.Value, attrTypesNew, attrTypesOld map[string]tftypes.Type, diags diag.Diagnostics) tftypes.Value {
	if !value.IsNull() {
		var valueList []tftypes.Value

		if err := value.As(&valueList); err != nil {
			diags.AddAttributeError(
				path.Root(field),
				"Unable to Convert Prior State",
				err.Error(),
			)
			panic(fmt.Sprintf("%v \n %v", err, value.Type()))
		}
		if len(valueList) == 1 {
			oldValue := valueList[0]

			var rawOldValue map[string]tftypes.Value

			if err := oldValue.As(&rawOldValue); err != nil {
				diags.AddError(
					"Unable to Convert Prior State",
					err.Error(),
				)
				return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, nil)
			}

			rawNewValue := make(map[string]tftypes.Value)

			for k, v := range rawOldValue {
				rawNewValue[k] = transformValue(k, v, attrTypesOld[k], attrTypesNew[k], diags)
			}

			return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, rawNewValue)
		}
		if len(valueList) == 0 {
			return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, nil)
		}
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			"Expected set to have size of 1",
		)
		panic(fmt.Sprintf("Wrong block size %v %v", field, len(valueList)))
	}
	return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, nil)
}

func transformValue(fieldName string, value tftypes.Value, oldType, newType tftypes.Type, diags diag.Diagnostics) tftypes.Value {
	if newType.Equal(tftypes.Bool) && oldType.Equal(tftypes.String) {
		return convertStringStateValueToBool(fieldName, value, diags)
	} else if newType.Equal(tftypes.Number) && oldType.Equal(tftypes.String) {
		return convertStringStateValueToNumber(fieldName, value, diags)
	} else {
		if newType.Equal(oldType) {
			return value
		} else {
			if _, ok := newType.(tftypes.Set); ok {
				var valueList []tftypes.Value
				if oldSetType, ok := oldType.(tftypes.Set); ok {
					if err := value.As(&valueList); err != nil {
						diags.AddAttributeError(
							path.Root(fieldName),
							"Unable to Convert Prior State",
							err.Error(),
						)
						panic(fmt.Sprintf("%v \n %v", err, value.Type()))
					}

					newValueList := []tftypes.Value{}
					elemTypeOld := oldSetType.ElementType
					elemTypeNew := newType.(tftypes.Set).ElementType
					for _, rawValue := range valueList {
						newValueList = append(newValueList, transformObjectValue(rawValue, elemTypeOld, elemTypeNew, diags))
					}

					return tftypes.NewValue(newType, newValueList)
				} else {
					return tftypes.NewValue(newType, nil)
				}
			} else {
				elemOldFieldTypes := oldType.(tftypes.Set).ElementType.(tftypes.Object).AttributeTypes
				elemNewFieldTypes := newType.(tftypes.Object).AttributeTypes
				return convertSetToBlock(fieldName, value, elemNewFieldTypes, elemOldFieldTypes, diags)
				//return transformObjectValue(value, oldType, newType, diags)
			}
		}
	}
}

func transformObjectValue(value tftypes.Value, oldType, newType tftypes.Type, diags diag.Diagnostics) tftypes.Value {
	itemOld := map[string]tftypes.Value{}
	if err := value.As(&itemOld); err != nil {
		diags.AddError(
			"Unable to Convert Prior State",
			err.Error(),
		)
		return tftypes.NewValue(newType, nil)
	}

	elemOldFieldTypes := oldType.(tftypes.Object).AttributeTypes
	elemNewFieldTypes := newType.(tftypes.Object).AttributeTypes

	itemNew := map[string]tftypes.Value{}
	for k, v := range itemOld {
		itemNew[k] = transformValue(k, v, elemOldFieldTypes[k], elemNewFieldTypes[k], diags)
	}

	return tftypes.NewValue(newType, itemNew)
}

func convertStringStateValueToBool(field string, value tftypes.Value, diags diag.Diagnostics) tftypes.Value {
	if !value.IsNull() {
		var valueStr string
		if err := value.As(&valueStr); err != nil {
			diags.AddAttributeError(
				path.Root(field),
				"Unable to Convert Prior State",
				err.Error(),
			)
			return tftypes.NewValue(tftypes.Bool, nil)
		}
		return tftypes.NewValue(tftypes.Bool, helpers.StrToBool(valueStr))
	}
	return tftypes.NewValue(tftypes.Bool, nil)
}

func convertStringStateValueToNumber(field string, value tftypes.Value, diags diag.Diagnostics) tftypes.Value {
	if !value.IsNull() {
		var valueStr string
		if err := value.As(&valueStr); err != nil {
			diags.AddAttributeError(
				path.Root(field),
				"Unable to Convert Prior State",
				err.Error(),
			)
			return tftypes.NewValue(tftypes.Number, nil)
		}
		return tftypes.NewValue(tftypes.Number, helpers.StrToInt(valueStr))
	}
	return tftypes.NewValue(tftypes.Number, nil)
}
