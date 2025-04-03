package schema

import (
	"sync"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	configOldDatasourceSchemaAttrs map[string]datasourceSchema.Attribute
	configOldResourceSchemaAttrs   map[string]resourceSchema.Attribute
	configOldResourceSchemaBlocks  map[string]resourceSchema.Block

	configOldDatasourceSchemaAttrsMutex sync.RWMutex = sync.RWMutex{}
	configOldResourceSchemaAttrsMutex   sync.RWMutex = sync.RWMutex{}
	configOldResourceSchemaBlocksMutex  sync.RWMutex = sync.RWMutex{}
)

func GetDatasourceConnectorConfigSchemaAttributes() map[string]datasourceSchema.Attribute {
	if len(configDatasourceSchemaAttrs) == 0 {
		result := make(map[string]datasourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			result[fn] = schemaAttributeFromConfigField(f, true).(datasourceSchema.Attribute)
		}
		if configOldDatasourceSchemaAttrsMutex.TryLock() {
			configOldDatasourceSchemaAttrs = result
		}
		return result
	}
	return configOldDatasourceSchemaAttrs
}

func GetResourceConnectorConfigSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(configResourceSchemaAttrs) == 0 {
		result := make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			}
		}
		if configOldResourceSchemaAttrsMutex.TryLock() {
			configOldResourceSchemaAttrs = result
		}
		return result
	}
	return configOldResourceSchemaAttrs
}

func GetResourceConnectorConfigSchemaBlocks() map[string]resourceSchema.Block {
	if len(configResourceSchemaBlocks) == 0 {
		result := make(map[string]resourceSchema.Block)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				result[fn] = schemaBlockFromConfigField(f)
			}
		}
		if configOldResourceSchemaBlocksMutex.TryLock() {
			configOldResourceSchemaBlocks = result
		}
		return result
	}
	return configOldResourceSchemaBlocks
}
