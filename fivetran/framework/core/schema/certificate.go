package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func CertificateConnectorResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the resource. Equal to target connection id.",
			},
			"connector_id": resourceSchema.StringAttribute{
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The unique identifier for the target connection within the Fivetran system.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateDestinationResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the resource. Equal to target destination id.",
			},
			"destination_id": resourceSchema.StringAttribute{
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The unique identifier for the target destination within the Fivetran system.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateConnectorDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
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
	return datasourceSchema.Schema {
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

func certificateDatasourceItem() datasourceSchema.SetNestedBlock {
	return datasourceSchema.SetNestedBlock{
		NestedObject: datasourceSchema.NestedBlockObject{
			Attributes: map[string]datasourceSchema.Attribute{
				"hash": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Hash of the fingerprint.",
				},
				"public_key": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "The SSH public key.",
				},
				"name": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate name.",
				},
				"type": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate key.",
				},
				"sha1": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate sha1.",
				},
				"sha256": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate sha256.",
				},
				"validated_by": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "User name who validated the fingerprint.",
				},
				"validated_date": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "The date when SSH fingerprint was approved.",
				},
			},
		},
	}
}

func certificateResourceItem() resourceSchema.SetNestedBlock {
	return resourceSchema.SetNestedBlock{
		NestedObject: resourceSchema.NestedBlockObject{
			Attributes: map[string]resourceSchema.Attribute{
				"hash": resourceSchema.StringAttribute{
					Required:    true,
					Description: "Hash of the fingerprint.",
				},
				"public_key": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "The SSH public key.",
				},
				"name": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate name.",
				},
				"type": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate key.",
				},
				"sha1": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate sha1.",
				},
				"sha256": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Certificate sha256.",
				},
				"validated_by": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "User name who validated the fingerprint.",
				},
				"validated_date": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "The date when SSH fingerprint was approved.",
				},
				"encoded_cert": resourceSchema.StringAttribute{
					Required:    true,
					Sensitive:   true,
					Description: "Base64 encoded certificate.",
				},
			},
		},
	}
}