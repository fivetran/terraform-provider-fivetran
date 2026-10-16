package model

import (
	"context"
	"sort"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ReconcileConfigAfterWrite is used after a create or update call.
// The response-derived config is kept as is, except for configured subfields that it resolved to null
// (for example, because the response omits them). Those keep their planned value, so Terraform accepts
// the apply result. Values that the default read logic already keeps (sensitive fields, empty strings,
// empty collections) are not changed.
//
// sent is the config sent in the request, keyed by API field name. Only kept subfields whose top-level
// field was sent are returned as unconfirmed paths, for the caller to report.
func ReconcileConfigAfterWrite(ctx context.Context, planned, applied types.Object, sent map[string]interface{}) (types.Object, []string, diag.Diagnostics) {
	return reconcileConfigAfterWriteWithFields(ctx, planned, applied, sent, common.GetDestinationFieldsMap())
}

func reconcileConfigAfterWriteWithFields(
	ctx context.Context,
	planned, applied types.Object,
	sent map[string]interface{},
	fieldsMap map[string]common.ConfigField,
) (types.Object, []string, diag.Diagnostics) {
	if planned.IsNull() || planned.IsUnknown() || applied.IsNull() || applied.IsUnknown() {
		return applied, nil, nil
	}

	var diags diag.Diagnostics
	var unconfirmed []string

	elements := make(map[string]attr.Value, len(applied.Attributes()))
	for name, value := range applied.Attributes() {
		elements[name] = value
	}

	for name, plannedValue := range planned.Attributes() {
		apiName := name
		if cf, ok := fieldsMap[name]; ok && cf.ApiField != "" {
			apiName = cf.ApiField
		}
		_, wasSent := sent[apiName]

		value, paths, d := keepPlannedWhereNull(ctx, plannedValue, elements[name], "config."+name)
		diags.Append(d...)
		elements[name] = value
		if wasSent {
			unconfirmed = append(unconfirmed, paths...)
		}
	}

	result, d := types.ObjectValue(applied.AttributeTypes(ctx), elements)
	diags.Append(d...)
	if diags.HasError() {
		return applied, nil, diags
	}

	sort.Strings(unconfirmed)
	return result, unconfirmed, diags
}

// keepPlannedWhereNull returns the planned value where the applied value is null, and recurses into
// nested objects. The returned paths are the subfields that kept their planned value.
func keepPlannedWhereNull(ctx context.Context, planned, applied attr.Value, path string) (attr.Value, []string, diag.Diagnostics) {
	if planned == nil || planned.IsNull() || planned.IsUnknown() || applied == nil {
		return applied, nil, nil
	}
	if applied.IsNull() {
		return planned, []string{path}, nil
	}

	plannedObject, isPlannedObject := planned.(types.Object)
	appliedObject, isAppliedObject := applied.(types.Object)
	if !isPlannedObject || !isAppliedObject {
		return applied, nil, nil
	}

	var diags diag.Diagnostics
	var paths []string

	elements := make(map[string]attr.Value, len(appliedObject.Attributes()))
	for name, value := range appliedObject.Attributes() {
		elements[name] = value
	}
	for name, plannedChild := range plannedObject.Attributes() {
		value, childPaths, d := keepPlannedWhereNull(ctx, plannedChild, elements[name], path+"."+name)
		diags.Append(d...)
		elements[name] = value
		paths = append(paths, childPaths...)
	}

	result, d := types.ObjectValue(appliedObject.AttributeTypes(ctx), elements)
	diags.Append(d...)
	if diags.HasError() {
		return applied, nil, diags
	}
	return result, paths, diags
}
