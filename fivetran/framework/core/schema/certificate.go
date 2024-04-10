package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func CertificateConnectorResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: fingerprintCertificateConnectorSchema().GetResourceSchema(),
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateDestinationResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: fingerprintCertificateDestinationSchema().GetResourceSchema(),
		Blocks: map[string]resourceSchema.Block{
			"certificate": certificateResourceItem(),
		},
	}
}

func CertificateConnectorDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: fingerprintCertificateConnectorSchema().GetDatasourceSchema(),
		Blocks: map[string]datasourceSchema.Block{
			"certificates": certificateDatasourceItem(),
		},
	}
}

func CertificateDestinationDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: fingerprintCertificateDestinationSchema().GetDatasourceSchema(),
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
