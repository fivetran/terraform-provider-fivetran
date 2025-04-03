package schema

import (
	"sync"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	authResourceSchemaAttrs  map[string]resourceSchema.Attribute
	authResourceSchemaBlocks map[string]resourceSchema.Block

	authResourceSchemaAttrsMutex  sync.RWMutex = sync.RWMutex{}
	authResourceSchemaBlocksMutex sync.RWMutex = sync.RWMutex{}
)

func GetResourceAuthSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(authResourceSchemaAttrs) == 0 {
		result := make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetAuthFieldsMap() {
			attr := schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				result[fn] = attr
			}
		}
		if authResourceSchemaAttrsMutex.TryLock() {
			authResourceSchemaAttrs = result
		}
		return result
	}
	return authResourceSchemaAttrs
}

func GetResourceAuthSchemaBlocks() map[string]resourceSchema.Block {
	if len(authResourceSchemaBlocks) == 0 {
		result := make(map[string]resourceSchema.Block)
		for fn, f := range common.GetAuthFieldsMap() {
			attr := schemaBlockFromConfigField(f)
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				result[fn] = attr
			}
		}
		if authResourceSchemaBlocksMutex.TryLock() {
			authResourceSchemaBlocks = result
		}
		return result
	}
	return authResourceSchemaBlocks
}
