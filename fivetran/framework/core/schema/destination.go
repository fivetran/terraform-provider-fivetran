package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DestinationAttributesSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        	true,
				ValueType:   	core.String,
				Description: 	"The unique identifier for the destination within the Fivetran system.",
			},
			"group_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the Group within the Fivetran system.",
			},
			"service": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The destination type id within the Fivetran system.",
			},
			"region": {
				Required:    true,
				ValueType:   core.String,
				Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
			},
			"time_zone_offset": {
				Required:    true,
				ValueType:   core.String,
				Description: "Determines the time zone for the Fivetran sync schedule.",
			},
			"trust_certificates": {
				ValueType:    core.Boolean,
				Description:  "Specifies whether we should trust the certificate automatically. The default value is FALSE. If a certificate is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination certificate](https://fivetran.com/docs/rest-api/certificates#approveadestinationcertificate).",
				ResourceOnly: true,
			},
			"trust_fingerprints": {
				ValueType:    core.Boolean,
				Description:  "Specifies whether we should trust the SSH fingerprint automatically. The default value is FALSE. If a fingerprint is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination fingerprint](https://fivetran.com/docs/rest-api/certificates#approveadestinationfingerprint).",
				ResourceOnly: true,
			},
			"run_setup_tests": {
				ValueType:    core.Boolean,
				Description:  "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
				ResourceOnly: true,
			},
			"setup_status": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "Destination setup status.",
			},
			"daylight_saving_time_enabled": {
				ValueType:   core.Boolean,
				Description: "Shift my UTC offset with daylight savings time (US Only)",
			},
			"private_link_id": {
				ValueType:   core.String,
				Description: "The private link ID.",
			},
			"hybrid_deployment_agent_id": {
				ValueType:   core.String,
				Description: "The hybrid deployment agent ID that refers to the controller created for the group the connection belongs to. If the value is specified, the system will try to associate the connection with an existing agent.",
			},
			"networking_method": {
				ValueType:   	core.StringEnum,
				Description: 	"Possible values: Directly, SshTunnel, ProxyAgent, PrivateLink.",
			},
		},
	}
}

func DestinationResourceBlocks(ctx context.Context) map[string]resourceSchema.Block {

	config := resourceSchema.SingleNestedBlock{
		Attributes: GetResourceDestinationConfigSchemaAttributes(),
	}

	blocks := GetResourceDestinationConfigSchemaBlocks()
	if len(blocks) > 0 {
		config.Blocks = blocks
	}

	return map[string]resourceSchema.Block{
		"config": config,
		"timeouts": timeouts.Block(ctx, timeouts.Opts{
			Create: true,
			Update: true,
		}),
	}
}

func DestinationDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"config": datasourceSchema.SingleNestedBlock{
			Attributes: GetDatasourceDestinationConfigSchemaAttributes(),
		},
	}
}

func DestinationsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"destinations": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: DestinationAttributesSchema().GetDatasourceListSchema(),
				},
			},
		},
	}
}
