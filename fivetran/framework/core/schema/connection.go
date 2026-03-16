package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ConnectionAttributesSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the connection within the Fivetran system.",
			},
			"name": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination.",
			},
			"connected_by": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The unique identifier of the user who has created the connection in your account.",
			},
			"created_at": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The timestamp of the time the connection was created in your account.",
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
				Description: "The connection service type (e.g., `postgres`, `mysql`, `s3`, `snowflake`). See [Fivetran connection types documentation](https://fivetran.com/docs/connectors) for available services.",
			},
			"config": {
				ValueType:    core.String,
				ResourceOnly: true,
				Description:  "Optional connection configuration as a JSON-encoded string. This config is merged with destination_schema fields and sent to the API during creation. The connection resource does not read this field back, allowing it to be managed separately by the `fivetran_connection_config` resource. Use this to provide service-specific required fields (e.g., `update_method` for Postgres/MySQL) or full connection configuration.",
			},
			"run_setup_tests": {
				ValueType:    core.Boolean,
				Description:  "Whether to run setup tests when creating the connection. Default: `false`. **Note:** This is a plan-only attribute and will not be stored in state.",
				ResourceOnly: true,
			},
			"trust_certificates": {
				ValueType:    core.Boolean,
				Description:  "Whether to automatically trust SSL certificates. Default: `false`. **Note:** This is a plan-only attribute.",
				ResourceOnly: true,
			},
			"trust_fingerprints": {
				ValueType:    core.Boolean,
				Description:  "Whether to automatically trust SSH fingerprints. Default: `false`. **Note:** This is a plan-only attribute.",
				ResourceOnly: true,
			},
			"succeeded_at": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The timestamp of the time the connection sync succeeded last time.",
			},
			"failed_at": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The timestamp of the time the connection sync failed last time.",
			},
			"service_version": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The connection type version within the Fivetran system.",
			},
			"sync_frequency": {
				DatasourceOnly: true,
				ValueType:      core.Integer,
				Description:    "The connection sync frequency in minutes.",
			},
			"schedule_type": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The connection schedule configuration type. Supported values: auto, manual.",
			},
			"paused": {
				DatasourceOnly: true,
				ValueType:      core.Boolean,
				Description:    "Specifies whether the connection is paused.",
			},
			"pause_after_trial": {
				DatasourceOnly: true,
				ValueType:      core.Boolean,
				Description:    "Specifies whether the connection should be paused after the free trial period has ended.",
			},
			"daily_sync_time": {
				DatasourceOnly: true,
				ValueType:      core.String,
				Description:    "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time.",
			},
			"proxy_agent_id": {
				ValueType:   core.String,
				Description: "The ID of the proxy agent to use. Required when `networking_method` is `ProxyAgent`.",
			},
			"networking_method": {
				ValueType:   core.StringEnum,
				Description: "The networking method for the connection. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.",
			},
			"hybrid_deployment_agent_id": {
				ValueType:   core.String,
				Description: "The hybrid deployment agent ID that refers to the controller created for the group the connection belongs to. If the value is specified, the system will try to associate the connection with an existing agent.",
			},
			"private_link_id": {
				ValueType:   core.String,
				Description: "The private link ID. Required when `networking_method` is `PrivateLink`.",
			},
			"data_delay_sensitivity": {
				ValueType:   core.String,
				Description: "The level of data delay notification threshold. Possible values: LOW, NORMAL, HIGH, CUSTOM, SYNC_FREQUENCY. The default value NORMAL. CUSTOM is only available for customers using the [Enterprise plan](https://fivetran.com/docs/getting-started/pricing#fivetranplans) or above.",
			},
			"data_delay_threshold": {
				ValueType:   core.Integer,
				Description: "Custom sync delay notification threshold in minutes. The default value is 0. This parameter is only used when data_delay_sensitivity set to CUSTOM.",
			},
		},
	}
}

func ConnectionResourceBlocks() map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"destination_schema": resourceSchema.SingleNestedBlock{
			Attributes: connectionDestinationSchema().GetResourceSchema(),
		},
	}
}

func ConnectionDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"destination_schema": datasourceSchema.SingleNestedBlock{
			Attributes: connectionDestinationSchema().GetDatasourceSchema(),
		},
		"status": connectionStatusBlock(),
	}
}

func connectionDestinationSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"name": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The connection schema name in destination. Has to be unique within the group (destination). Required for connection creation.",
			},
			"table": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The table name unique within the schema to which connection will sync the data. Required for connection creation.",
			},
			"prefix": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "The connection schema prefix has to be unique within the group (destination). Each replicated schema is prefixed with the provided value. Required for connection creation.",
			},
			"table_group_name": {
				ForceNew:    true,
				Required:    false,
				ValueType:   core.String,
				Description: "Table group name.",
			},
		},
	}
}

func connectionStatusBlock() datasourceSchema.SingleNestedBlock {
	return datasourceSchema.SingleNestedBlock{
		Attributes: map[string]datasourceSchema.Attribute{
			"setup_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current setup state of the connection. The available values are: <br /> - incomplete - the setup config is incomplete, the setup tests never succeeded  `connected` - the connection is properly set up, `broken` - the connection setup config is broken.",
			},
			"is_historical_sync": datasourceSchema.BoolAttribute{
				Computed:    true,
				Description: "The boolean specifying whether the connection should be triggered to re-sync all historical data. If you set this parameter to TRUE, the next scheduled sync will be historical. If the value is FALSE or not specified, the connection will not re-sync historical data. NOTE: When the value is TRUE, only the next scheduled sync will be historical, all subsequent ones will be incremental. This parameter is set to FALSE once the historical sync is completed.",
			},
			"sync_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current sync state of the connection. The available values are: `scheduled` - the sync is waiting to be run, `syncing` - the sync is currently running, `paused` - the sync is currently paused, `rescheduled` - the sync is waiting until more API calls are available in the source service.",
			},
			"update_state": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current data update state of the connection. The available values are: `on_schedule` - the sync is running smoothly, no delays, `delayed` - the data is delayed for a longer time than expected for the update.",
			},
			"tasks": datasourceSchema.SetNestedAttribute{
				Computed:    true,
				Description: "The collection of tasks for the connection.",
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
				Description: "The collection of warnings for the connection.",
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

func ConnectionsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
			"group_id": datasourceSchema.StringAttribute{
				Optional:    true,
				Description: "The ID of the group (destination) to filter connections by.",
			},
			"schema_name": datasourceSchema.StringAttribute{
				Optional:    true,
				Description: "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"connections": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: ConnectionAttributesSchema().GetDatasourceListSchema(),
				},
			},
		},
	}
}
