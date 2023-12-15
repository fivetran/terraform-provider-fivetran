package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	authResourceSchemaAttrs  map[string]resourceSchema.Attribute
	authResourceSchemaBlocks map[string]resourceSchema.Block
)

func GetResourceAuthSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(authResourceSchemaAttrs) == 0 {
		authResourceSchemaAttrs = make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetAuthFieldsMap() {
			attr := schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				authResourceSchemaAttrs[fn] = attr
			}
		}
	}
	return authResourceSchemaAttrs
}

func GetResourceAuthSchemaBlocks() map[string]resourceSchema.Block {
	if len(authResourceSchemaBlocks) == 0 {
		authResourceSchemaBlocks = make(map[string]resourceSchema.Block)
		for fn, f := range common.GetAuthFieldsMap() {
			attr := schemaBlockFromConfigField(f)
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				authResourceSchemaBlocks[fn] = attr
			}
		}
	}
	return authResourceSchemaBlocks
}
