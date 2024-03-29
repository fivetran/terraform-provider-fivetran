package fivetrantypes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*JsonSchemaType)(nil)
	_ xattr.TypeWithValidate  = (*JsonSchemaType)(nil)
)

type JsonSchemaType struct {
	basetypes.StringType
}

func (t JsonSchemaType) String() string {
	return "fivetrantypes.JsonSchemaType"
}

func (t JsonSchemaType) ValueType(ctx context.Context) attr.Value {
	return JsonSchemaValue{}
}

func (t JsonSchemaType) Equal(o attr.Type) bool {
	other, ok := o.(JsonSchemaType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t JsonSchemaType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Is(tftypes.String) {
		err := fmt.Errorf("expected String value, received %T with value: %v", in, in)
		diags.AddAttributeError(
			path,
			"JSON Normalized Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var valueString string

	if err := in.As(&valueString); err != nil {
		diags.AddAttributeError(
			path,
			"JSON Schema Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)

		return diags
	}

	if ok := json.Valid([]byte(valueString)); !ok {
		diags.AddAttributeError(
			path,
			"Invalid JSON Schema Value",
			"A string value was provided that is not valid JSON string format.\n\n"+
				"Given Value: "+valueString+"\n",
		)

		return diags
	}

	return diags
}

func (t JsonSchemaType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return JsonSchemaValue{
		StringValue: in,
	}, nil
}

func (t JsonSchemaType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
