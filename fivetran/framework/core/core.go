package core

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type clientContainer struct {
	client *fivetran.Client
}

type ProviderDatasource struct {
	clientContainer
}

type ProviderResource struct {
	clientContainer
}

func (d *clientContainer) GetClient() *fivetran.Client {
	return d.client
}

func (d *ProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	d.getClient(resp.Diagnostics, req.ProviderData)
}

func (d *ProviderDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.getClient(resp.Diagnostics, req.ProviderData)
}

func (d *clientContainer) getClient(diag diag.Diagnostics, data any) {
	// Prevent panic if the provider has not been configured.
	if data == nil {
		return
	}

	client, ok := data.(*fivetran.Client)

	if !ok {
		diag.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *fivetran.Client, got: %T. Please report this issue to the provider developers.", data),
		)
		return
	}

	d.client = client
}

type FieldValueType int64

const (
	String FieldValueType = iota
	Integer
	Boolean
	StringEnum
	StringsList
	StringsSet
)

type SchemaField struct {
	ValueType FieldValueType

	IsId           bool
	Required       bool
	ForceNew       bool
	DatasourceOnly bool
	ResourceOnly   bool
	Sensitive      bool

	DefaultString string

	Readonly    bool
	Description string
}

type Schema struct {
	Fields map[string]SchemaField
}

func (s Schema) GetDatasourceSchema() map[string]datasourceSchema.Attribute {
	result := map[string]datasourceSchema.Attribute{}
	for k, v := range s.Fields {
		if !v.ResourceOnly {
			result[k] = v.getDatasourceSchemaAttribute()
		}
	}
	return result
}

func (s Schema) GetResourceSchema() map[string]resourceSchema.Attribute {
	result := make(map[string]resourceSchema.Attribute)
	for k, v := range s.Fields {
		if !v.DatasourceOnly {
			result[k] = v.getResourceSchemaAttribute()
		}
	}
	return result
}

func (s SchemaField) getDatasourceSchemaAttribute() datasourceSchema.Attribute {
	var result datasourceSchema.Attribute
	switch s.ValueType {
	case StringEnum:
		result = datasourceSchema.StringAttribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Sensitive:   s.Sensitive,
			Description: s.Description,
		}
	case String:
		result = datasourceSchema.StringAttribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Sensitive:   s.Sensitive,
			Description: s.Description,
		}
	case Boolean:
		result = datasourceSchema.BoolAttribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Description: s.Description,
		}
	case Integer:
		result = datasourceSchema.Int64Attribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Description: s.Description,
		}
	case StringsList:
		result = datasourceSchema.ListAttribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Sensitive:   s.Sensitive,
			Description: s.Description,
			ElementType: basetypes.StringType{},
		}
	case StringsSet:
		result = datasourceSchema.SetAttribute{
			Required:    s.IsId,
			Computed:    !s.IsId,
			Sensitive:   s.Sensitive,
			Description: s.Description,
			ElementType: basetypes.StringType{},
		}
	}

	return result
}

func (s SchemaField) getResourceSchemaAttribute() resourceSchema.Attribute {
	var result resourceSchema.Attribute
	switch s.ValueType {
	case StringEnum:
		var stringAttribute = resourceSchema.StringAttribute{
			Required:    s.Required,
			Computed:    !s.Required || s.Readonly || (s.IsId && !s.Required),
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
			Sensitive:   s.Sensitive,
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			}
		}
		if s.DefaultString != "" {
			stringAttribute.Default = stringdefault.StaticString(s.DefaultString)
		}
		result = stringAttribute
	case String:
		var stringAttribute = resourceSchema.StringAttribute{
			Required:    s.Required,
			Computed:    s.Readonly || (s.IsId && !s.Required),
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
			Sensitive:   s.Sensitive,
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			}
		}
		if s.DefaultString != "" {
			stringAttribute.Default = stringdefault.StaticString(s.DefaultString)
		}
		result = stringAttribute
	case Boolean:
		var stringAttribute = resourceSchema.BoolAttribute{
			Required:    s.Required,
			Computed:    !s.Required || s.Readonly,
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			}
		}
		result = stringAttribute
	case Integer:
		var stringAttribute = resourceSchema.Int64Attribute{
			Required:    s.Required,
			Computed:    !s.Required || s.Readonly,
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			}
		}
		result = stringAttribute
	case StringsList:
		var stringAttribute = resourceSchema.ListAttribute{
			Required:    s.Required,
			Computed:    (s.ValueType == StringEnum && !s.Required) || s.Readonly || (s.IsId && !s.Required),
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
			ElementType: basetypes.StringType{},
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.List{
				listplanmodifier.RequiresReplace(),
			}
		}
		result = stringAttribute
	case StringsSet:
		var stringAttribute = resourceSchema.SetAttribute{
			Required:    s.Required,
			Computed:    (s.ValueType == StringEnum && !s.Required) || s.Readonly || (s.IsId && !s.Required),
			Optional:    !s.Required && !s.Readonly && !s.IsId,
			Description: s.Description,
			ElementType: basetypes.StringType{},
		}
		if s.ForceNew {
			stringAttribute.PlanModifiers = []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			}
		}
		result = stringAttribute
	}
	return result
}

func GetBoolOrDefault(value basetypes.BoolValue, fallback bool) bool {
	if value.IsNull() || value.IsUnknown() {
		return fallback
	}
	return value.ValueBool()
}
