package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

func buildDescription(fieldDescription map[string]string) string {
	var result []string

	keys := make([]string, 0, len(fieldDescription))
	for k := range fieldDescription {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, service := range keys {
		if fieldDescription[service] != "" {
			result = append(result, fmt.Sprintf("\t- Service `%v`: %v", service, fieldDescription[service]))
		}
	}
	if len(result) > 0 {
		return "Field usage depends on `service` value: \n" + strings.Join(result, "\n")
	} else {
		return ""
	}
}

func schemaAttributeFromConfigField(cf common.ConfigField, datasource bool) interface{} {
	switch cf.FieldValueType {
	case common.Boolean:
		if datasource {
			return datasourceSchema.BoolAttribute{Computed: true, Description: buildDescription(cf.Description)}
		} else {
			return resourceSchema.BoolAttribute{Optional: !cf.Readonly, Computed: true, Description: buildDescription(cf.Description)}
		}
	case common.Integer:
		if datasource {
			return datasourceSchema.Int64Attribute{Computed: true, Description: buildDescription(cf.Description)}
		} else {
			return resourceSchema.Int64Attribute{Optional: !cf.Readonly, Computed: true, Description: buildDescription(cf.Description)}
		}
	case common.String:
		if datasource {
			return datasourceSchema.StringAttribute{Computed: true, Sensitive: cf.Sensitive, Description: buildDescription(cf.Description)}
		} else {
			return resourceSchema.StringAttribute{
				Optional:    !cf.Readonly,
				Computed:    cf.Readonly || !cf.Nullable,
				Sensitive:   cf.Sensitive,
				Description: buildDescription(cf.Description),
			}
		}
	case common.StringList:
		elemType := types.StringType
		if datasource {
			return datasourceSchema.SetAttribute{
				ElementType: elemType,
				Computed:    true,
				Description: buildDescription(cf.Description),
			}
		} else {
			return resourceSchema.SetAttribute{
				ElementType: elemType,
				Optional:    !cf.Readonly,
				Computed:    cf.Readonly,
				Sensitive:   cf.Sensitive,
				Description: buildDescription(cf.Description),
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
				Description: buildDescription(cf.Description),
			}
		} else {
			return resourceSchema.SetNestedAttribute{
				Optional: !cf.Readonly,
				Computed: cf.Readonly,
				NestedObject: resourceSchema.NestedAttributeObject{
					Attributes: toResourceAttr(subFields),
				},
				Description: buildDescription(cf.Description),
			}
		}
	case common.Object:
		subFields := make(map[string]interface{})
		for fn, f := range cf.ItemFields {
			subFields[fn] = schemaAttributeFromConfigField(f, datasource)
		}
		if datasource {
			return datasourceSchema.SingleNestedAttribute{
				Computed:    true,
				Attributes:  toDatasourceAttr(subFields),
				Description: buildDescription(cf.Description),
			}
		} else {
			return resourceSchema.SingleNestedAttribute{
				Optional:    !cf.Readonly,
				Computed:    cf.Readonly,
				Attributes:  toResourceAttr(subFields),
				Description: buildDescription(cf.Description),
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
