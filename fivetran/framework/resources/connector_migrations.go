package resources

import (
	"bytes"
	"context"
	"encoding/json"
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
	rawStateValue, err := unmarshalRawStateLenient(req.RawState, getConnectorStateModel(fromVersion))

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

	diags := &resp.Diagnostics

	lpaValue := tftypes.NewValue(tftypes.String, nil)
	if fromVersion == 3 {
		if !rawState["hybrid_deployment_agent_id"].IsNull() {
			lpaValue = rawState["hybrid_deployment_agent_id"]
		} else if !rawState["local_processing_agent_id"].IsNull() {
			lpaValue = rawState["local_processing_agent_id"]
		}
	} else if fromVersion > 3 {
		lpaValue = rawState["hybrid_deployment_agent_id"]
	}

	destinationSchemaFields := map[string]tftypes.Type{
		"name":   tftypes.String,
		"table":  tftypes.String,
		"prefix": tftypes.String,
	}

	config := rawState["config"]
	auth := rawState["auth"]
	destination_schema := rawState["destination_schema"]

	// Versions below 3 (providers up to 1.1.5) store config, auth and destination_schema as
	// single-element sets with primitives as strings, and unset strings as "". Version 3 and
	// later store them as blocks, and unset values as null.
	if fromVersion < 3 {
		config = emptyStringsToNull("config", config, diags)
		auth = emptyStringsToNull("auth", auth, diags)
		destination_schema = emptyStringsToNull("destination_schema", destination_schema, diags)
		config = convertSetToBlock("config", config, model.GetTfTypes(common.GetConfigFieldsMap(), 3), model.GetTfTypes(common.GetConfigFieldsMap(), fromVersion), diags)
		auth = convertSetToBlock("auth", auth, model.GetTfTypes(common.GetAuthFieldsMap(), 3), model.GetTfTypes(common.GetAuthFieldsMap(), fromVersion), diags)
		destination_schema = convertSetToBlock("destination_schema", destination_schema, destinationSchemaFields, destinationSchemaFields, diags)
	}

	if fromVersion < 5 {
		destination_schema = transformObjectValue(destination_schema,
			tftypes.Object{AttributeTypes: destinationSchemaFields},
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":             tftypes.String,
					"table":            tftypes.String,
					"prefix":           tftypes.String,
					"table_group_name": tftypes.String,
				},
			},
			diags)
	}

	if diags.HasError() {
		return
	}

	// Attributes that exist only from version 3 on are kept when the prior state has them.
	fromV3 := func(name string, t tftypes.Type) tftypes.Value {
		if fromVersion >= 3 {
			if v, ok := rawState[name]; ok {
				return v
			}
		}
		return tftypes.NewValue(t, nil)
	}

	runSetupTests := rawState["run_setup_tests"]
	trustFingerprints := rawState["trust_fingerprints"]
	trustCertificates := rawState["trust_certificates"]
	if fromVersion < 3 {
		runSetupTests = convertStringStateValueToBool("run_setup_tests", runSetupTests, diags)
		trustFingerprints = convertStringStateValueToBool("trust_fingerprints", trustFingerprints, diags)
		trustCertificates = convertStringStateValueToBool("trust_certificates", trustCertificates, diags)
	}

	dynamicValue, err := tfprotov6.NewDynamicValue(
		getConnectorStateModel(5),
		tftypes.NewValue(getConnectorStateModel(5), map[string]tftypes.Value{
			"id":                         rawState["id"],
			"name":                       rawState["name"],
			"connected_by":               rawState["connected_by"],
			"created_at":                 rawState["created_at"],
			"group_id":                   rawState["group_id"],
			"service":                    rawState["service"],
			"timeouts":                   rawState["timeouts"],
			"networking_method":          fromV3("networking_method", tftypes.String),
			"proxy_agent_id":             fromV3("proxy_agent_id", tftypes.String),
			"private_link_id":            fromV3("private_link_id", tftypes.String),
			"data_delay_sensitivity":     fromV3("data_delay_sensitivity", tftypes.String),
			"data_delay_threshold":       fromV3("data_delay_threshold", tftypes.Number),
			"data_checks":                fromV3("data_checks", tftypes.Bool),
			"hybrid_deployment_agent_id": lpaValue,
			"run_setup_tests":            runSetupTests,
			"trust_fingerprints":         trustFingerprints,
			"trust_certificates":         trustCertificates,

			"config":             config,
			"auth":               auth,
			"destination_schema": destination_schema,
		}),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Upgraded State",
			err.Error(),
		)
		return
	}

	resp.DynamicValue = &dynamicValue
}

func getConnectorStateModel(version int) tftypes.Type {
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
	if version >= 3 && version <= 5 {
		if version == 5 {
			base["destination_schema"] = tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":             tftypes.String,
					"table":            tftypes.String,
					"prefix":           tftypes.String,
					"table_group_name": tftypes.String,
				},
			}
			base["data_checks"] = tftypes.Bool
		} else {
			base["destination_schema"] = tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":   tftypes.String,
					"table":  tftypes.String,
					"prefix": tftypes.String,
				},
			}
		}

		base["run_setup_tests"] = tftypes.Bool
		base["trust_certificates"] = tftypes.Bool
		base["trust_fingerprints"] = tftypes.Bool
		base["proxy_agent_id"] = tftypes.String
		base["networking_method"] = tftypes.String
		base["hybrid_deployment_agent_id"] = tftypes.String
		if version == 3 {
			base["local_processing_agent_id"] = tftypes.String
		}
		base["private_link_id"] = tftypes.String
		base["data_delay_sensitivity"] = tftypes.String
		base["data_delay_threshold"] = tftypes.Number
		base["config"] = tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetConfigFieldsMap(), 3)}
		base["auth"] = tftypes.Object{AttributeTypes: model.GetTfTypes(common.GetAuthFieldsMap(), 3)}
	} else {
		base["destination_schema"] = tftypes.Set{ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":   tftypes.String,
				"table":  tftypes.String,
				"prefix": tftypes.String,
			},
		},
		}
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

// unmarshalRawStateLenient unmarshals prior state into stateType, dropping attributes that
// stateType does not have. config and auth state holds one attribute per connector config field
// the writing release knew; fields removed from the provider since then can't be unmarshalled.
func unmarshalRawStateLenient(rawState *tfprotov6.RawState, stateType tftypes.Type) (tftypes.Value, error) {
	if rawState == nil || rawState.JSON == nil {
		return rawState.Unmarshal(stateType)
	}

	decoder := json.NewDecoder(bytes.NewReader(rawState.JSON))
	decoder.UseNumber()
	var parsed interface{}
	if err := decoder.Decode(&parsed); err != nil {
		return tftypes.Value{}, err
	}

	pruned, err := json.Marshal(pruneJSONToType(parsed, stateType))
	if err != nil {
		return tftypes.Value{}, err
	}

	return (&tfprotov6.RawState{JSON: pruned}).Unmarshal(stateType)
}

func pruneJSONToType(value interface{}, t tftypes.Type) interface{} {
	switch typ := t.(type) {
	case tftypes.Object:
		object, ok := value.(map[string]interface{})
		if !ok {
			return value
		}
		pruned := make(map[string]interface{}, len(object))
		for name, attrValue := range object {
			if attrType, ok := typ.AttributeTypes[name]; ok {
				pruned[name] = pruneJSONToType(attrValue, attrType)
			}
		}
		return pruned
	case tftypes.Set:
		return pruneJSONListToType(value, typ.ElementType)
	case tftypes.List:
		return pruneJSONListToType(value, typ.ElementType)
	case tftypes.Map:
		object, ok := value.(map[string]interface{})
		if !ok {
			return value
		}
		pruned := make(map[string]interface{}, len(object))
		for key, elem := range object {
			pruned[key] = pruneJSONToType(elem, typ.ElementType)
		}
		return pruned
	default:
		return value
	}
}

func pruneJSONListToType(value interface{}, elemType tftypes.Type) interface{} {
	list, ok := value.([]interface{})
	if !ok {
		return value
	}
	pruned := make([]interface{}, len(list))
	for i, elem := range list {
		pruned[i] = pruneJSONToType(elem, elemType)
	}
	return pruned
}

// emptyStringsToNull replaces every "" in value with null.
func emptyStringsToNull(field string, value tftypes.Value, diags *diag.Diagnostics) tftypes.Value {
	result, err := tftypes.Transform(value, func(_ *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if !v.Type().Equal(tftypes.String) || !v.IsKnown() || v.IsNull() {
			return v, nil
		}
		var str string
		if err := v.As(&str); err != nil {
			return v, err
		}
		if str == "" {
			return tftypes.NewValue(tftypes.String, nil), nil
		}
		return v, nil
	})
	if err != nil {
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			err.Error(),
		)
		return value
	}
	return result
}

func convertSetToBlock(field string, value tftypes.Value, attrTypesNew, attrTypesOld map[string]tftypes.Type, diags *diag.Diagnostics) tftypes.Value {
	nullBlock := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, nil)
	if value.IsNull() {
		return nullBlock
	}

	var valueList []tftypes.Value
	if err := value.As(&valueList); err != nil {
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			fmt.Sprintf("%v (type %v)", err, value.Type()),
		)
		return nullBlock
	}

	switch len(valueList) {
	case 0:
		return nullBlock
	case 1:
	default:
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			fmt.Sprintf("Expected the prior state set to hold at most 1 element, got %d.", len(valueList)),
		)
		return nullBlock
	}

	var rawOldValue map[string]tftypes.Value
	if err := valueList[0].As(&rawOldValue); err != nil {
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			err.Error(),
		)
		return nullBlock
	}

	rawNewValue := make(map[string]tftypes.Value, len(attrTypesNew))
	for k, newType := range attrTypesNew {
		oldValue, ok := rawOldValue[k]
		if !ok || attrTypesOld[k] == nil {
			rawNewValue[k] = tftypes.NewValue(newType, nil)
			continue
		}
		rawNewValue[k] = transformValue(k, oldValue, attrTypesOld[k], newType, diags)
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, rawNewValue)
}

func removeUnsupportedAttributes(field string, value tftypes.Value, attrTypesNew map[string]tftypes.Type, diags *diag.Diagnostics) tftypes.Value {
	nullBlock := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, nil)
	if value.IsNull() {
		return nullBlock
	}

	var oldValues map[string]tftypes.Value
	if err := value.As(&oldValues); err != nil {
		diags.AddAttributeError(
			path.Root(field),
			"Unable to Convert Prior State",
			fmt.Sprintf("%v (type %v)", err, value.Type()),
		)
		return nullBlock
	}

	rawNewValue := make(map[string]tftypes.Value, len(attrTypesNew))
	for fieldName, fieldType := range attrTypesNew {
		if oldValue, ok := oldValues[fieldName]; ok {
			rawNewValue[fieldName] = oldValue
		} else {
			rawNewValue[fieldName] = tftypes.NewValue(fieldType, nil)
		}
	}

	return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesNew}, rawNewValue)
}

// transformValue converts a value stored with oldType in prior state into newType: strings to
// bools and numbers, single-element sets of objects to blocks, and sets of objects element-wise.
func transformValue(fieldName string, value tftypes.Value, oldType, newType tftypes.Type, diags *diag.Diagnostics) tftypes.Value {
	if newType.Equal(oldType) {
		return value
	}
	if newType.Equal(tftypes.Bool) && oldType.Equal(tftypes.String) {
		return convertStringStateValueToBool(fieldName, value, diags)
	}
	if newType.Equal(tftypes.Number) && oldType.Equal(tftypes.String) {
		return convertStringStateValueToNumber(fieldName, value, diags)
	}
	if value.IsNull() {
		return tftypes.NewValue(newType, nil)
	}

	oldSetType, oldIsSet := oldType.(tftypes.Set)

	if newSetType, ok := newType.(tftypes.Set); ok {
		if !oldIsSet {
			// e.g. "servers", which versions below 2 store as a string.
			return tftypes.NewValue(newType, nil)
		}
		var valueList []tftypes.Value
		if err := value.As(&valueList); err != nil {
			diags.AddAttributeError(
				path.Root(fieldName),
				"Unable to Convert Prior State",
				fmt.Sprintf("%v (type %v)", err, value.Type()),
			)
			return tftypes.NewValue(newType, nil)
		}
		newValueList := make([]tftypes.Value, 0, len(valueList))
		for _, rawValue := range valueList {
			newValueList = append(newValueList, transformObjectValue(rawValue, oldSetType.ElementType, newSetType.ElementType, diags))
		}
		return tftypes.NewValue(newType, newValueList)
	}

	newObjectType, newIsObject := newType.(tftypes.Object)
	if !newIsObject || !oldIsSet {
		diags.AddAttributeError(
			path.Root(fieldName),
			"Unable to Convert Prior State",
			fmt.Sprintf("Unsupported conversion from %v to %v.", oldType, newType),
		)
		return tftypes.NewValue(newType, nil)
	}
	oldElemType, ok := oldSetType.ElementType.(tftypes.Object)
	if !ok {
		diags.AddAttributeError(
			path.Root(fieldName),
			"Unable to Convert Prior State",
			fmt.Sprintf("Unsupported conversion from %v to %v.", oldType, newType),
		)
		return tftypes.NewValue(newType, nil)
	}
	return convertSetToBlock(fieldName, value, newObjectType.AttributeTypes, oldElemType.AttributeTypes, diags)
}

func transformObjectValue(value tftypes.Value, oldType, newType tftypes.Type, diags *diag.Diagnostics) tftypes.Value {
	if value.IsNull() {
		return tftypes.NewValue(newType, nil)
	}

	elemOldType, okOld := oldType.(tftypes.Object)
	elemNewType, okNew := newType.(tftypes.Object)
	if !okOld || !okNew {
		if newType.Equal(oldType) {
			return value
		}
		diags.AddError(
			"Unable to Convert Prior State",
			fmt.Sprintf("Unsupported conversion from %v to %v.", oldType, newType),
		)
		return tftypes.NewValue(newType, nil)
	}

	itemOld := map[string]tftypes.Value{}
	if err := value.As(&itemOld); err != nil {
		diags.AddError(
			"Unable to Convert Prior State",
			err.Error(),
		)
		return tftypes.NewValue(newType, nil)
	}

	itemNew := make(map[string]tftypes.Value, len(elemNewType.AttributeTypes))
	for k, newFieldType := range elemNewType.AttributeTypes {
		ev, ok := itemOld[k]
		oldFieldType := elemOldType.AttributeTypes[k]
		if ok && oldFieldType != nil {
			itemNew[k] = transformValue(k, ev, oldFieldType, newFieldType, diags)
		} else {
			itemNew[k] = tftypes.NewValue(newFieldType, nil)
		}
	}

	return tftypes.NewValue(newType, itemNew)
}

func convertStringStateValueToBool(field string, value tftypes.Value, diags *diag.Diagnostics) tftypes.Value {
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

func convertStringStateValueToNumber(field string, value tftypes.Value, diags *diag.Diagnostics) tftypes.Value {
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
