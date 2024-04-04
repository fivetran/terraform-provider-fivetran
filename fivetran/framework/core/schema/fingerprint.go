package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func FingerprintConnectorResource() resourceSchema.Schema {
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
			"fingerprint": fingerprintResourceItem(),
		},
	}
}

func FingerprintDestinationResource() resourceSchema.Schema {
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
			"fingerprint": fingerprintResourceItem(),
		},
	}
}

func FingerprintConnectorDatasource() datasourceSchema.Schema {
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
			"fingerprints": fingerprintDatasourceItem(),
		},
	}
}


func FingerprintDestinationDatasource() datasourceSchema.Schema {
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
			"fingerprints": fingerprintDatasourceItem(),
		},
	}
}

func fingerprintDatasourceItem() datasourceSchema.SetNestedBlock {
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

func fingerprintResourceItem() resourceSchema.SetNestedBlock {
	return resourceSchema.SetNestedBlock{
		NestedObject: resourceSchema.NestedBlockObject{
			Attributes: map[string]resourceSchema.Attribute{
				"hash": resourceSchema.StringAttribute{
					Required:    true,
					Description: "Hash of the fingerprint.",
				},
				"public_key": resourceSchema.StringAttribute{
					Required:    true,
					Description: "The SSH public key.",
				},
				"validated_by": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "User name who validated the fingerprint.",
				},
				"validated_date": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "The date when SSH fingerprint was approved.",
				},
			},
		},
	}
}