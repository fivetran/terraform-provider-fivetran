package model

import (
	"context"

	externallogging "github.com/fivetran/go-fivetran/external_logging"
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

var ExternalLoggingTFConfigType = map[string]attr.Type{
	"workspace_id":        types.StringType,
	"port":                types.Int64Type,
	"log_group_name":      types.StringType,
	"role_arn":            types.StringType,
	"external_id":         types.StringType,
	"region":              types.StringType,
	"sub_domain":          types.StringType,
	"host":                types.StringType,
	"hostname":            types.StringType,
	"enable_ssl":          types.BoolType,
	"channel":             types.StringType,
	"project_id":          types.StringType,
	"primary_key":         types.StringType,
	"api_key":             types.StringType,
	"token":               types.StringType,
	"access_key_id":       types.StringType,
	"service_account_key": types.StringType,
	"access_key_secret":   types.StringType,
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

	readStringValue(resp, &config, "workspace_id")
	readStringValue(resp, &config, "log_group_name")
	readStringValue(resp, &config, "role_arn")
	readStringValue(resp, &config, "external_id")
	readStringValue(resp, &config, "region")
	readStringValue(resp, &config, "sub_domain")
	readStringValue(resp, &config, "host")
	readStringValue(resp, &config, "hostname")
	readStringValue(resp, &config, "channel")
	readStringValue(resp, &config, "project_id")
	readStringValue(resp, &config, "access_key_id")
	readStringValue(resp, &config, "service_account_key")

	readBoolValue(resp, &config, "enable_ssl")

	readIntValue(resp, &config, "port")

	d.readSensitiveStringValue(resp, &config, planConfigForEmptySecretValuesAfterImport, "primary_key")
	d.readSensitiveStringValue(resp, &config, planConfigForEmptySecretValuesAfterImport, "api_key")
	d.readSensitiveStringValue(resp, &config, planConfigForEmptySecretValuesAfterImport, "token")
	d.readSensitiveStringValue(resp, &config, planConfigForEmptySecretValuesAfterImport, "access_key_secret")

	d.Config, _ = types.ObjectValue(ExternalLoggingTFConfigType, config)
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
