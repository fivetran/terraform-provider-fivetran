package resources

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithValidateConfig = &connectionV2{}

func (r *connectionV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.GetSkipPlanTimeValidation() {
		return
	}

	var data model.ConnectionV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Service.IsNull() || data.Service.IsUnknown() {
		return
	}

	configMap, configDiags := core.DynamicToMap(ctx, data.Config)
	resp.Diagnostics.Append(configDiags...)
	authMap, authDiags := core.DynamicToMap(ctx, data.Auth)
	resp.Diagnostics.Append(authDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if configMap == nil && authMap == nil {
		return
	}

	meta, err := r.connectorMetadata(ctx, data.Service.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Validate Connection V2 Configuration",
			fmt.Sprintf("Unable to fetch metadata for service %q. Terraform cannot safely validate dynamic config/auth fields without metadata. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", data.Service.ValueString(), err),
		)
		return
	}

	validateDynamicObject(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
	validateDynamicObject(authMap, &meta.Auth, path.Root("auth"), &resp.Diagnostics)
}

func validateDynamicObject(values map[string]interface{}, slot *metadata.Property, root path.Path, diags *diag.Diagnostics) {
	if values == nil {
		return
	}
	if slot == nil {
		diags.AddAttributeError(
			root,
			"Unsupported Dynamic Field",
			"Metadata does not define this dynamic object.",
		)
		return
	}

	for name, value := range values {
		prop := metadataSlotProp(slot, name)
		fieldPath := root.AtName(name)
		if prop == nil {
			diags.AddAttributeError(
				fieldPath,
				"Unsupported Dynamic Field",
				fmt.Sprintf("Metadata for this service does not define field %q.", name),
			)
			continue
		}

		if !core.IsKnownMetadataFieldStatus(prop.FieldStatus) {
			diags.AddAttributeWarning(
				fieldPath,
				"Unknown Metadata Field Status",
				fmt.Sprintf("Field %q has unknown metadata fieldStatus %q. Terraform will validate the field shape, but the field availability may need provider support in the future.", name, prop.FieldStatus),
			)
		} else if core.ShouldWarnForMetadataFieldStatus(prop) {
			diags.AddAttributeWarning(
				fieldPath,
				"Non-GA Dynamic Field",
				fmt.Sprintf("Field %q is marked as %q in connector metadata. It is accepted because the metadata endpoint returned it for this account, but availability may change before general availability.", name, prop.FieldStatus),
			)
		}

		validateDynamicValue(value, prop, fieldPath, diags)
	}
}

func validateDynamicValue(value interface{}, prop *metadata.Property, valuePath path.Path, diags *diag.Diagnostics) {
	if value == nil || prop == nil {
		return
	}

	switch prop.Type {
	case "string":
		validateStringValue(value, prop, valuePath, diags)
	case "integer":
		if !isIntegerValue(value) {
			addTypeError(valuePath, prop.Type, value, diags)
		}
	case "number":
		if !isNumberValue(value) {
			addTypeError(valuePath, prop.Type, value, diags)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			addTypeError(valuePath, prop.Type, value, diags)
		}
	case "array":
		items, ok := value.([]interface{})
		if !ok {
			addTypeError(valuePath, prop.Type, value, diags)
			return
		}
		if prop.Items == nil {
			return
		}
		for i, item := range items {
			validateDynamicValue(item, prop.Items, valuePath.AtListIndex(i), diags)
		}
	case "object":
		nested, ok := value.(map[string]interface{})
		if !ok {
			addTypeError(valuePath, prop.Type, value, diags)
			return
		}
		validateDynamicObject(nested, prop, valuePath, diags)
	case "":
		return
	default:
		diags.AddAttributeError(
			valuePath,
			"Unknown Metadata Field Type",
			fmt.Sprintf("Connector metadata returned unsupported type %q. Expected one of string, integer, number, boolean, array, or object.", prop.Type),
		)
	}
}

func validateStringValue(value interface{}, prop *metadata.Property, valuePath path.Path, diags *diag.Diagnostics) {
	stringValue, ok := value.(string)
	if !ok {
		addTypeError(valuePath, prop.Type, value, diags)
		return
	}
	if len(prop.Enum) == 0 {
		return
	}
	for _, allowed := range prop.Enum {
		if stringValue == allowed {
			return
		}
	}
	diags.AddAttributeError(
		valuePath,
		"Invalid Dynamic Field Value",
		fmt.Sprintf("Value %q is not allowed. Allowed values: %s.", stringValue, strings.Join(prop.Enum, ", ")),
	)
}

func addTypeError(valuePath path.Path, expected string, value interface{}, diags *diag.Diagnostics) {
	diags.AddAttributeError(
		valuePath,
		"Invalid Dynamic Field Type",
		fmt.Sprintf("Expected metadata type %q, got Terraform value type %T.", expected, value),
	)
}

func isIntegerValue(value interface{}) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return math.Trunc(float64(v)) == float64(v)
	case float64:
		return math.Trunc(v) == v
	default:
		return false
	}
}

func isNumberValue(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

func metadataSlotProp(slot *metadata.Property, key string) *metadata.Property {
	if slot == nil || slot.Properties == nil {
		return nil
	}
	return slot.Properties[key]
}
