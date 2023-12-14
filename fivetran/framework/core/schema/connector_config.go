package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	configDatasourceSchemaAttrs map[string]datasourceSchema.Attribute
	configResourceSchemaAttrs   map[string]resourceSchema.Attribute
	configResourceSchemaBlocks  map[string]resourceSchema.Block
)

func ConnectorDatasourceConfig() datasourceSchema.SingleNestedAttribute {
	return datasourceSchema.SingleNestedAttribute{
		Computed:   true,
		Attributes: GetDatasourceConfigSchemaAttributes(),
	}
}

func GetDatasourceConfigSchemaAttributes() map[string]datasourceSchema.Attribute {
	if len(configDatasourceSchemaAttrs) == 0 {
		configDatasourceSchemaAttrs = make(map[string]datasourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			configDatasourceSchemaAttrs[fn] = schemaAttributeFromConfigField(f, true).(datasourceSchema.Attribute)
		}
	}
	return configDatasourceSchemaAttrs
}

func GetResourceConfigSchemaAttributes() map[string]resourceSchema.Attribute {
	if len(configDatasourceSchemaAttrs) == 0 {
		configResourceSchemaAttrs = make(map[string]resourceSchema.Attribute)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType != common.ObjectList && f.FieldValueType != common.Object {
				configResourceSchemaAttrs[fn] = schemaAttributeFromConfigField(f, false).(resourceSchema.Attribute)
			}
		}
	}
	return configResourceSchemaAttrs
}

func GetResourceConfigSchemaBlocks() map[string]resourceSchema.Block {
	if len(configResourceSchemaBlocks) == 0 {
		configResourceSchemaBlocks = make(map[string]resourceSchema.Block)
		for fn, f := range common.GetConfigFieldsMap() {
			if f.FieldValueType == common.ObjectList || f.FieldValueType == common.Object {
				configResourceSchemaBlocks[fn] = schemaBlockFromConfigField(f)
			}
		}
	}
	return configResourceSchemaBlocks
}

func schemaBlockFromConfigField(cf common.ConfigField) resourceSchema.Block {
	subFields := make(map[string]interface{})
	subBlocks := make(map[string]resourceSchema.Block)

	for fn, f := range cf.ItemFields {
		if f.FieldValueType != common.ObjectList {
			subFields[fn] = schemaAttributeFromConfigField(f, false)
		} else {
			subBlocks[fn] = schemaBlockFromConfigField(f)
		}
	}
	if cf.FieldValueType == common.ObjectList {
		return resourceSchema.SetNestedBlock{
			NestedObject: resourceSchema.NestedBlockObject{
				Attributes: toResourceAttr(subFields),
				Blocks:     subBlocks,
			},
		}
	}
	if cf.FieldValueType == common.Object {
		return resourceSchema.SingleNestedBlock{
			Attributes: toResourceAttr(subFields),
			Blocks:     subBlocks,
		}
	}
	return nil
}

func schemaAttributeFromConfigField(cf common.ConfigField, datasource bool) interface{} {
	switch cf.FieldValueType {
	case common.Boolean:
		if datasource {
			return datasourceSchema.BoolAttribute{Computed: true}
		} else {
			return resourceSchema.BoolAttribute{Optional: !cf.Readonly, Computed: true}
		}
	case common.Integer:
		if datasource {
			return datasourceSchema.Int64Attribute{Computed: true}
		} else {
			return resourceSchema.Int64Attribute{Optional: !cf.Readonly, Computed: true}
		}
	case common.String:
		if datasource {
			return datasourceSchema.StringAttribute{Computed: true, Sensitive: cf.Sensitive}
		} else {
			return resourceSchema.StringAttribute{
				Optional:  !cf.Readonly,
				Computed:  cf.Readonly || !cf.Nullable,
				Sensitive: cf.Sensitive,
			}
		}
	case common.StringList:
		elemType := types.StringType
		if datasource {
			return datasourceSchema.SetAttribute{
				ElementType: elemType,
				Computed:    true,
			}
		} else {
			return resourceSchema.SetAttribute{
				ElementType: elemType,
				Optional:    !cf.Readonly,
				Computed:    cf.Readonly,
				Sensitive:   cf.Sensitive,
			}
		}
	case common.ObjectList:
		subFields := make(map[string]interface{})
		for fn, f := range cf.ItemFields {
			subFields[fn] = schemaAttributeFromConfigField(f, datasource)
		}
		if datasource {
			return datasourceSchema.SetNestedAttribute{
				Computed: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: toDatasourceAttr(subFields),
				},
			}
		} else {
			return resourceSchema.SetNestedAttribute{
				Optional: !cf.Readonly,
				Computed: cf.Readonly,
				NestedObject: resourceSchema.NestedAttributeObject{
					Attributes: toResourceAttr(subFields),
				},
			}
		}
	case common.Object:
		subFields := make(map[string]interface{})
		for fn, f := range cf.ItemFields {
			subFields[fn] = schemaAttributeFromConfigField(f, datasource)
		}
		if datasource {
			return datasourceSchema.SingleNestedAttribute{
				Computed:   true,
				Attributes: toDatasourceAttr(subFields),
			}
		} else {
			return resourceSchema.SingleNestedAttribute{
				Optional:   !cf.Readonly,
				Computed:   cf.Readonly,
				Attributes: toResourceAttr(subFields),
			}
		}
	}
	return nil
}

func toDatasourceAttr(in map[string]interface{}) map[string]datasourceSchema.Attribute {
	result := make(map[string]datasourceSchema.Attribute)
	for k, v := range in {
		result[k] = v.(datasourceSchema.Attribute)
	}
	return result
}
func toResourceAttr(in map[string]interface{}) map[string]resourceSchema.Attribute {
	result := make(map[string]resourceSchema.Attribute)
	for k, v := range in {
		result[k] = v.(resourceSchema.Attribute)
	}
	return result
}
