package schema

import (
	"sync"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
)

var (
	externalLoggingResourceSchemaAttrs   map[string]resourceSchema.Attribute
	externalLoggingDatasourceSchemaAttrs map[string]datasourceSchema.Attribute

	externalLoggingResourceSchemaAttrsMutex   sync.RWMutex = sync.RWMutex{}
	externalLoggingDatasourceSchemaAttrsMutex sync.RWMutex = sync.RWMutex{}
)

func GetResourceExternalLoggingConfigSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(externalLoggingResourceSchemaAttrs) == 0 {
		result := make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetExternalLoggingFieldsMap() {
			if fn == "port" {
				// special case for port to add default value for backward compatibility
				result[fn] = resourceSchema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Default:     int64default.StaticInt64(0),
					Description: "Port",
				}
			} else if fn == "enable_ssl" {
				// special case for enable_ssl to add default value for backward compatibility
				result[fn] = resourceSchema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     booldefault.StaticBool(false),
					Description: "Enable SSL",
				}
			} else if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			}
		}
		if externalLoggingResourceSchemaAttrsMutex.TryLock() {
			externalLoggingResourceSchemaAttrs = result
			externalLoggingResourceSchemaAttrsMutex.Unlock()
		}
		return result
	}
	return externalLoggingResourceSchemaAttrs
}

func GetDatasourceExternalLoggingConfigSchemaAttributes() map[string]datasourceSchema.Attribute {
	if len(externalLoggingDatasourceSchemaAttrs) == 0 {
		result := make(map[string]datasourceSchema.Attribute)
		for fn, f := range common.GetExternalLoggingFieldsMap() {
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = schemaAttributeFromConfigField(f, true).(datasourceSchema.Attribute)
			}
		}
		if externalLoggingDatasourceSchemaAttrsMutex.TryLock() {
			externalLoggingDatasourceSchemaAttrs = result
			externalLoggingDatasourceSchemaAttrsMutex.Unlock()
		}
		return result
	}
	return externalLoggingDatasourceSchemaAttrs
}
