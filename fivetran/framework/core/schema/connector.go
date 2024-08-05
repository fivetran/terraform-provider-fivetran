package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ConnectorAttributesSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the connector within the Fivetran system.",
			},
			"name": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The name used both as the connector's name within the Fivetran system and as the source schema's name within your destination.",
			},
			"connected_by": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The unique identifier of the user who has created the connector in your account.",
			},
			"created_at": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The timestamp of the time the connector was created in your account.",
			},
			"group_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the Group (Destination) within the Fivetran system.",
			},
			"service": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The connector type id within the Fivetran system.",
			},

			"run_setup_tests": {
				ValueType:    core.Boolean,
				Description:  "Specifies whether the setup tests should be run automatically. The default value is FALSE.",
				ResourceOnly: true,
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

			"succeeded_at": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The timestamp of the time the connector sync succeeded last time.",
			},
			"failed_at": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The timestamp of the time the connector sync failed last time.",
			},
			"service_version": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The connector type version within the Fivetran system.",
			},
			"sync_frequency": {
				DatasourceOnly: true,
				ValueType:      core.Integer,
				Description:    "The connector sync frequency in minutes.",
			},
			"schedule_type": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The connector schedule configuration type. Supported values: auto, manual.",
			},
			"paused": {
				DatasourceOnly: true,
				ValueType:      core.Boolean,
				Description:    "Specifies whether the connector is paused.",
			},
			"pause_after_trial": {
				DatasourceOnly: true,
				ValueType:      core.Boolean,
				Description:    "Specifies whether the connector should be paused after the free trial period has ended.",
			},
			"daily_sync_time": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time.",
			},
			"proxy_agent_id": {
				ValueType:   core.String,
				Description: "The proxy agent ID.",
			},
			"networking_method": {
				ValueType:   	core.StringEnum,
				Description: 	"Possible values: Directly, SshTunnel, ProxyAgent.",
			},
			"local_processing_agent_id": {
				ValueType:   core.String,
				Description: "The local processing agent ID that refers to the controller created for the group the connection belongs to. If the value is specified, the system will try to associate the connection with an existing agent.",
			},
		},
	}
}

func ConnectorResourceBlocks(ctx context.Context) map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"destination_schema": resourceSchema.SingleNestedBlock{
			Attributes: destinationSchemaAttributes().GetResourceSchema(),
		},
		"config": resourceSchema.SingleNestedBlock{
			Attributes: GetResourceConnectorConfigSchemaAttributes(),
			Blocks:     GetResourceConnectorConfigSchemaBlocks(),
		},
		"auth": resourceSchema.SingleNestedBlock{
			Attributes: GetResourceAuthSchemaAttributes(),
			Blocks:     GetResourceAuthSchemaBlocks(),
		},
		"timeouts": timeouts.Block(ctx, timeouts.Opts{
			Create: true,
			Update: true,
		}),
	}
}

func connectorStatusBlock() datasourceSchema.SingleNestedBlock {
	return datasourceSchema.SingleNestedBlock{
		Attributes: map[string]datasourceSchema.Attribute{
			"setup_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current setup state of the connector. The available values are: <br /> - incomplete - the setup config is incomplete, the setup tests never succeeded  `connected` - the connector is properly set up, `broken` - the connector setup config is broken.",
			},
			"is_historical_sync": datasourceSchema.BoolAttribute{
				Computed:    true,
				Description: "The boolean specifying whether the connector should be triggered to re-sync all historical data. If you set this parameter to TRUE, the next scheduled sync will be historical. If the value is FALSE or not specified, the connector will not re-sync historical data. NOTE: When the value is TRUE, only the next scheduled sync will be historical, all subsequent ones will be incremental. This parameter is set to FALSE once the historical sync is completed.",
			},
			"sync_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current sync state of the connector. The available values are: `scheduled` - the sync is waiting to be run, `syncing` - the sync is currently running, `paused` - the sync is currently paused, `rescheduled` - the sync is waiting until more API calls are available in the source service.",
			},
			"update_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current data update state of the connector. The available values are: `on_schedule` - the sync is running smoothly, no delays, `delayed` - the data is delayed for a longer time than expected for the update.",
			},
			"tasks": datasourceSchema.SetNestedAttribute{
				Computed:    true,
				Description: "The collection of tasks for the connector.",
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"code": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "Task code.",
						},
						"message": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "Task message.",
						},
					},
				},
			},
			"warnings": datasourceSchema.SetNestedAttribute{
				Computed:    true,
				Description: "The collection of warnings for the connector.",
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"code": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "Warning code.",
						},
						"message": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "Warning message.",
						},
					},
				},
			},
		},
	}
}

func ConnectorDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"destination_schema": datasourceSchema.SingleNestedBlock{
			Attributes: destinationSchemaAttributes().GetDatasourceSchema(),
		},
		"config": datasourceSchema.SingleNestedBlock{
			Attributes: GetDatasourceConnectorConfigSchemaAttributes(),
		},
		"status": connectorStatusBlock(),
	}
}

func destinationSchemaAttributes() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"name": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The connector schema name in destination. Has to be unique within the group (destination). Required for connector creation.",
			},
			"table": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The table name unique within the schema to which connector will sync the data. Required for connector creation.",
			},
			"prefix": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The connector schema prefix has to be unique within the group (destination). Each replicated schema is prefixed with the provided value. Required for connector creation.",
			},
		},
	}
}
