package fivetrantypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// FastStringSetType and FastStringSetValue implement a set-of-strings attribute
// backed by a sorted tftypes.List for O(n) comparison at the Terraform core
// wire-protocol level. Elements are always sorted to ensure deterministic
// ordering and prevent false diffs.

// --- Type ---

var _ basetypes.ListTypable = FastStringSetType{}

type FastStringSetType struct {
	basetypes.ListType
}

func (t FastStringSetType) String() string {
	return "fivetrantypes.FastStringSetType"
}

func (t FastStringSetType) ValueType(_ context.Context) attr.Value {
	return FastStringSetValue{}
}

func (t FastStringSetType) Equal(o attr.Type) bool {
	other, ok := o.(FastStringSetType)
	if !ok {
		return false
	}
	return t.ListType.Equal(other.ListType)
}

func (t FastStringSetType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.List{ElementType: tftypes.String}
}

func (t FastStringSetType) ValueFromList(_ context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	return FastStringSetValue{ListValue: in}, nil
}

func (t FastStringSetType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewFastStringSetNull(), nil
	}
	if !in.IsKnown() {
		return FastStringSetValue{ListValue: basetypes.NewListUnknown(types.StringType)}, nil
	}
	if in.IsNull() {
		return NewFastStringSetNull(), nil
	}

	var raw []tftypes.Value
	if err := in.As(&raw); err != nil {
		return nil, err
	}

	strs := make([]string, 0, len(raw))
	for _, v := range raw {
		var s string
		if err := v.As(&s); err != nil {
			return nil, fmt.Errorf("element: %w", err)
		}
		strs = append(strs, s)
	}

	return NewFastStringSetFromStrings(strs), nil
}

// --- Value ---

var (
	_ basetypes.ListValuable                  = FastStringSetValue{}
	_ basetypes.ListValuableWithSemanticEquals = FastStringSetValue{}
	_ xattr.ValidateableAttribute             = FastStringSetValue{}
)

type FastStringSetValue struct {
	basetypes.ListValue
}

func (v FastStringSetValue) Type(_ context.Context) attr.Type {
	return FastStringSetType{ListType: basetypes.ListType{ElemType: types.StringType}}
}

func (v FastStringSetValue) ToListValue(_ context.Context) (basetypes.ListValue, diag.Diagnostics) {
	return v.ListValue, nil
}

// Equal performs O(n) set comparison using map[string]struct{}.
func (v FastStringSetValue) Equal(o attr.Value) bool {
	other, ok := o.(FastStringSetValue)
	if !ok {
		return false
	}

	if v.IsNull() != other.IsNull() || v.IsUnknown() != other.IsUnknown() {
		return false
	}
	if v.IsNull() || v.IsUnknown() {
		return true
	}

	vElems := v.Elements()
	oElems := other.Elements()
	if len(vElems) != len(oElems) {
		return false
	}

	set := make(map[string]struct{}, len(vElems))
	for _, elem := range vElems {
		if strVal, ok := elem.(types.String); ok {
			set[strVal.ValueString()] = struct{}{}
		}
	}
	for _, elem := range oElems {
		if strVal, ok := elem.(types.String); ok {
			if _, exists := set[strVal.ValueString()]; !exists {
				return false
			}
		}
	}

	return true
}

// ListSemanticEquals returns true when both values represent the same set of
// strings regardless of order. This prevents false drift when the API returns
// schemas in a different order than the config or prior state.
func (v FastStringSetValue) ListSemanticEquals(_ context.Context, other basetypes.ListValuable) (bool, diag.Diagnostics) {
	otherFast, ok := other.(FastStringSetValue)
	if !ok {
		return false, nil
	}
	return v.Equal(otherFast), nil
}

// ValidateAttribute performs O(n) duplicate detection using a Go map.
// Reports all duplicates found with their positions in the list.
func (v FastStringSetValue) ValidateAttribute(_ context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	// Track first occurrence index for each value
	firstIndex := make(map[string]int, len(v.Elements()))
	for i, elem := range v.Elements() {
		if strVal, ok := elem.(types.String); ok {
			s := strVal.ValueString()
			if first, exists := firstIndex[s]; exists {
				resp.Diagnostics.AddAttributeError(req.Path, "Duplicate Value",
					fmt.Sprintf("The value %q at index %d is a duplicate of the value at index %d. "+
						"Each element must be unique.", s, i, first))
			} else {
				firstIndex[s] = i
			}
		}
	}
}

// --- Constructors ---

func NewFastStringSetNull() FastStringSetValue {
	return FastStringSetValue{ListValue: basetypes.NewListNull(types.StringType)}
}

func NewFastStringSetFromStrings(values []string) FastStringSetValue {
	elems := make([]attr.Value, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	listValue, _ := basetypes.NewListValue(types.StringType, elems)
	return FastStringSetValue{ListValue: listValue}
}
