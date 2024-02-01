package schema

import (
	"sync"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	destinationDatasourceSchemaAttrs  map[string]datasourceSchema.Attribute
	destinationDatasourceSchemaBlocks map[string]datasourceSchema.Block

	destinationResourceSchemaAttrs  map[string]resourceSchema.Attribute
	destinationResourceSchemaBlocks map[string]resourceSchema.Block

	destinationDatasourceSchemaAttrsMutex  sync.RWMutex = sync.RWMutex{}
	destinationResourceSchemaAttrsMutex    sync.RWMutex = sync.RWMutex{}
	destinationResourceSchemaBlocksMutex   sync.RWMutex = sync.RWMutex{}
	destinationDatasourceSchemaBlocksMutex sync.RWMutex = sync.RWMutex{}
)

func GetResourceDestinationConfigSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(destinationResourceSchemaAttrs) == 0 {
		result := make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetDestinationFieldsMap() {
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			}
		}
		if destinationResourceSchemaAttrsMutex.TryLock() {
			destinationResourceSchemaAttrs = result
		}
		return result
	}
	return destinationResourceSchemaAttrs
}

func GetDatasourceDestinationConfigSchemaAttributes() map[string]datasourceSchema.Attribute {
	if len(destinationDatasourceSchemaAttrs) == 0 {
		result := make(map[string]datasourceSchema.Attribute)
		for fn, f := range common.GetDestinationFieldsMap() {
			result[fn] = schemaAttributeFromConfigField(f, true).(datasourceSchema.Attribute)
		}
		if destinationDatasourceSchemaAttrsMutex.TryLock() {
			destinationDatasourceSchemaAttrs = result
		}
		return result
	}
	return destinationDatasourceSchemaAttrs
}

func GetResourceDestinationConfigSchemaBlocks() map[string]resourceSchema.Block {
	if len(destinationResourceSchemaBlocks) == 0 {
		result := make(map[string]resourceSchema.Block)
		for fn, f := range common.GetDestinationFieldsMap() {
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				result[fn] = schemaBlockFromConfigField(f)
			}
		}
		if destinationResourceSchemaBlocksMutex.TryLock() {
			destinationResourceSchemaBlocks = result
		}
		return result
	}
	return destinationResourceSchemaBlocks
}

func GetDatasourceDestinationConfigSchemaBlocks() map[string]datasourceSchema.Block {
	if len(destinationDatasourceSchemaBlocks) == 0 {
		result := make(map[string]datasourceSchema.Block)
		for fn, f := range common.GetDestinationFieldsMap() {
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				result[fn] = schemaBlockFromConfigField(f)
			}
		}
		if destinationDatasourceSchemaBlocksMutex.TryLock() {
			destinationDatasourceSchemaBlocks = result
		}
		return result
	}
	return destinationDatasourceSchemaBlocks
}
