package model

import (
	"context"

	externallogging "github.com/fivetran/go-fivetran/external_logging"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ExternalLogging struct {
	Id       types.String `tfsdk:"id"`
	GroupId  types.String `tfsdk:"group_id"`
	Service  types.String `tfsdk:"service"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	RunTests types.Bool   `tfsdk:"run_setup_tests"`
	Config   types.Object `tfsdk:"config"`
}

func isSensitiveExternalLoggingAttribute(attribute string) bool {
	if attr, ok := common.GetExternalLoggingFieldsMap()[attribute]; ok {
		return attr.Sensitive
	}
	return true
}

func getExternalLoggingTFConfigType() map[string]attr.Type {
	return getAttrTypes(common.GetExternalLoggingFieldsMap())
}

func CoalesceToStringNull(value attr.Value) attr.Value {
	if value == nil {
		return types.StringNull()
	}

	return value
}

func (d *ExternalLogging) ReadFromCustomResponse(
	ctx context.Context,
	resp externallogging.ExternalLoggingCustomResponse,
	planConfigForEmptySecretValuesAfterImport map[string]attr.Value) {
	d.Id = types.StringValue(resp.Data.Id)
	d.GroupId = types.StringValue(resp.Data.Id)
	d.Service = types.StringValue(resp.Data.Service)
	d.Enabled = types.BoolValue(resp.Data.Enabled)

	config := map[string]attr.Value{}

	for key, _ := range getExternalLoggingTFConfigType() {
		d.readValue(resp, &config, planConfigForEmptySecretValuesAfterImport, key)
	}
	d.Config, _ = types.ObjectValue(getExternalLoggingTFConfigType(), config)
}

func (d *ExternalLogging) readValue(
	resp externallogging.ExternalLoggingCustomResponse,
	config *map[string]attr.Value,
	planConfigForEmptySecretValuesAfterImport map[string]attr.Value,
	key string) {
	if isSensitiveExternalLoggingAttribute(key) {
		d.readSensitiveStringValue(resp, config, planConfigForEmptySecretValuesAfterImport, key)
	} else if t, ok := getExternalLoggingTFConfigType()[key]; ok {
		if t == types.StringType {
			readStringValue(resp, config, key)
		} else if t == types.Int64Type {
			readIntValue(resp, config, key)
		} else if t == types.BoolType {
			readBoolValue(resp, config, key)
		}
	}
}

func readIntValue(resp externallogging.ExternalLoggingCustomResponse, config *map[string]attr.Value, key string) {
	if resp.Data.Config[key] != nil {
		(*config)[key] = types.Int64Value(int64(resp.Data.Config[key].(float64)))
	} else {
		(*config)[key] = types.Int64Value(0)
	}
}

func readBoolValue(resp externallogging.ExternalLoggingCustomResponse, config *map[string]attr.Value, key string) {
	if resp.Data.Config[key] != nil {
		(*config)[key] = types.BoolValue(resp.Data.Config[key].(bool))
	} else {
		(*config)[key] = types.BoolValue(false)
	}
}

func (d *ExternalLogging) readSensitiveStringValue(
	resp externallogging.ExternalLoggingCustomResponse,
	config *map[string]attr.Value,
	planConfigForEmptySecretValuesAfterImport map[string]attr.Value,
	key string) {
	if resp.Data.Config[key] != nil && resp.Data.Config[key] != "" && resp.Data.Config[key] != "******" {
		(*config)[key] = types.StringValue(resp.Data.Config[key].(string))
	} else if mapHasValue(planConfigForEmptySecretValuesAfterImport, key) {
		(*config)[key] = planConfigForEmptySecretValuesAfterImport[key]
	} else if mapHasValue(d.Config.Attributes(), key) && !d.Config.Attributes()[key].IsNull() {
		(*config)[key] = d.Config.Attributes()[key]
	} else {
		(*config)[key] = types.StringNull()
	}
}

func readStringValue(resp externallogging.ExternalLoggingCustomResponse, config *map[string]attr.Value, key string) {
	if resp.Data.Config[key] != nil && resp.Data.Config[key] != "" {
		(*config)[key] = types.StringValue(resp.Data.Config[key].(string))
	} else {
		(*config)[key] = types.StringNull()
	}
}

func mapHasValue(valuesMap map[string]attr.Value, key string) bool {
	if valuesMap == nil {
		return false
	}

	value, exists := valuesMap[key]
	if !exists {
		return false
	}

	return !value.IsNull() && !value.IsUnknown()
}

func (d *ExternalLogging) GetConfig() map[string]interface{} {
	attr := d.Config.Attributes()

	config := make(map[string]interface{})
	for k, v := range attr {
		if !v.IsUnknown() && !v.IsNull() {
			if t, ok := v.(basetypes.Int64Value); ok {
				config[k] = t.ValueInt64()
			}

			if t, ok := v.(basetypes.BoolValue); ok {
				config[k] = t.ValueBool()
			}

			if t, ok := v.(basetypes.StringValue); ok {
				config[k] = t.ValueString()
			}
		}
	}

	return config
}
