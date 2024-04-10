package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func CertificateConnectorResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the resource. Equal to target connection id.",
			},
			"connector_id": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the target connection within the Fivetran system.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateDestinationResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the resource. Equal to target destination id.",
			},
			"destination_id": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the target destination within the Fivetran system.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateConnectorDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the resource. Equal to target connection id.",
			},
			"connector_id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the target connection within the Fivetran system.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"certificates": certificateDatasourceItem(),
		},
	}
}

func CertificateDestinationDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the resource. Equal to target destination id.",
			},
			"destination_id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the target destination within the Fivetran system.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"certificates": certificateDatasourceItem(),
		},
	}
}

func certificateItemSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"hash": {
				IsId:        true,
				Required:    true,
				Readonly:    false,
				ValueType:   core.String,
				Description: "Hash of the certificate.",
			},
			"encoded_cert": {
				Required:     true,
				ResourceOnly: true,
				Sensitive:    true,
				Readonly:     false,
				ValueType:    core.String,
				Description:  "Base64 encoded certificate.",
			},
			"public_key": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The SSH public key.",
			},
			"name": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Certificate name.",
			},
			"type": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Type of the certificate.",
			},
			"sha1": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Certificate sha1.",
			},
			"sha256": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Certificate sha256.",
			},
			"validated_by": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "User name who validated the certificate.",
			},
			"validated_date": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The date when certificate was approved.",
			},
		},
	}
}

func certificateDatasourceItem() datasourceSchema.SetNestedBlock {
	return datasourceSchema.SetNestedBlock{
		NestedObject: datasourceSchema.NestedBlockObject{
			Attributes: certificateItemSchema().GetDatasourceSchema(),
		},
	}
}

func certificateResourceItem() resourceSchema.SetNestedBlock {
	return resourceSchema.SetNestedBlock{
		NestedObject: resourceSchema.NestedBlockObject{
			Attributes: certificateItemSchema().GetResourceSchema(),
		},
	}
}
