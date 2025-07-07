package fivetrantypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*JsonConfigValue)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*JsonConfigValue)(nil)
)

type JsonConfigValue struct {
	basetypes.StringValue
}

func (v JsonConfigValue) Type(_ context.Context) attr.Type {
	return JsonConfigType{}
}

func (v JsonConfigValue) Equal(o attr.Value) bool {
	other, ok := o.(JsonConfigValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v JsonConfigValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(JsonConfigValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	result, err := jsonEqual(newValue.ValueString(), v.ValueString())
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return false, diags
	}

	return result, diags
}

func (v JsonConfigValue) Unmarshal(target any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("JSON Config Unmarshal Error", "json string value is null"))
		return diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("JSON Config Unmarshal Error", "json string value is unknown"))
		return diags
	}

	err := json.Unmarshal([]byte(v.ValueString()), target)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("JSON Config Unmarshal Error", err.Error()))
	}

	return diags
}

func NewJsonConfigNull() JsonConfigValue {
	return JsonConfigValue{
		StringValue: basetypes.NewStringNull(),
	}
}

func NewJsonConfigUnknown() JsonConfigValue {
	return JsonConfigValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}

func NewJsonConfigValue(value string) JsonConfigValue {
	return JsonConfigValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

func NewJsonConfigPointerValue(value *string) JsonConfigValue {
	return JsonConfigValue{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}