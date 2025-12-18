package schema

import (
	"sync"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	configDatasourceSchemaAttrs map[string]datasourceSchema.Attribute
	configResourceSchemaAttrs   map[string]resourceSchema.Attribute
	configResourceSchemaBlocks  map[string]resourceSchema.Block

	configDatasourceSchemaAttrsMutex sync.RWMutex = sync.RWMutex{}
	configResourceSchemaAttrsMutex   sync.RWMutex = sync.RWMutex{}
	configResourceSchemaBlocksMutex  sync.RWMutex = sync.RWMutex{}
)

func GetDatasourceConnectorConfigSchemaAttributes() map[string]datasourceSchema.Attribute {
	if len(configDatasourceSchemaAttrs) == 0 {
		result := make(map[string]datasourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			result[fn] = schemaAttributeFromConfigField(f, true).(datasourceSchema.Attribute)
		}
		if configDatasourceSchemaAttrsMutex.TryLock() {
			configDatasourceSchemaAttrs = result
		}
		return result
	}
	return configDatasourceSchemaAttrs
}

func GetResourceConnectorConfigSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(configResourceSchemaAttrs) == 0 {
		result := make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			}
		}
		if configResourceSchemaAttrsMutex.TryLock() {
			configResourceSchemaAttrs = result
		}
		return result
	}
	return configResourceSchemaAttrs
}

func GetResourceConnectorConfigSchemaBlocks() map[string]resourceSchema.Block {
	if len(configResourceSchemaBlocks) == 0 {
		result := make(map[string]resourceSchema.Block)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				result[fn] = schemaBlockFromConfigField(f)
			}
		}
		if configResourceSchemaBlocksMutex.TryLock() {
			configResourceSchemaBlocks = result
		}
		return result
	}
	return configResourceSchemaBlocks
}
