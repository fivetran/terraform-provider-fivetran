package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func fingerprintCertificateConnectorSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the resource. Equal to target connection id.",
			},
			"connector_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the target connection within the Fivetran system.",
			},
		},
	}
}

func fingerprintCertificateConnectionSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the target connection within the Fivetran system.",
			},
		},
	}
}

func fingerprintCertificateDestinationSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the resource. Equal to target destination id.",
			},
			"destination_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the target destination within the Fivetran system.",
			},
		},
	}
}

func fingerprintItemSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"hash": {
				IsId:        true,
				Required:    true,
				ValueType:   core.String,
				Description: "Hash of the fingerprint.",
			},
			"public_key": {
				Required:    true,
				ValueType:   core.String,
				Description: "The SSH public key.",
			},
			"validated_by": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "User name who validated the fingerprint.",
			},
			"validated_date": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The date when fingerprint was approved.",
			},
		},
	}
}

func FingerprintConnectorResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: fingerprintCertificateConnectorSchema().GetResourceSchema(),
		Blocks: map[string]resourceSchema.Block{
			"fingerprint": fingerprintResourceItem(),
		},
	}
}

func FingerprintConnectorDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		DeprecationMessage: "This datasource is Deprecated, please migrate to actual resource",
		Attributes: fingerprintCertificateConnectorSchema().GetDatasourceSchema(),
		Blocks: map[string]datasourceSchema.Block{
			"fingerprints": fingerprintDatasourceItem(),
		},
	}
}

func FingerprintConnectionDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: fingerprintCertificateConnectorSchema().GetDatasourceSchema(),
		Blocks: map[string]datasourceSchema.Block{
			"fingerprints": fingerprintDatasourceItem(),
		},
	}
}

func FingerprintDestinationResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: fingerprintCertificateDestinationSchema().GetResourceSchema(),
		Blocks: map[string]resourceSchema.Block{
			"fingerprint": fingerprintResourceItem(),
		},
	}
}

func FingerprintDestinationDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: fingerprintCertificateDestinationSchema().GetDatasourceSchema(),
		Blocks: map[string]datasourceSchema.Block{
			"fingerprints": fingerprintDatasourceItem(),
		},
	}
}

func fingerprintDatasourceItem() datasourceSchema.SetNestedBlock {
	return datasourceSchema.SetNestedBlock{
		NestedObject: datasourceSchema.NestedBlockObject{
			Attributes: fingerprintItemSchema().GetDatasourceSchema(),
		},
	}
}

func fingerprintResourceItem() resourceSchema.SetNestedBlock {
	return resourceSchema.SetNestedBlock{
		NestedObject: resourceSchema.NestedBlockObject{
			Attributes: fingerprintItemSchema().GetResourceSchema(),
		},
	}
}
