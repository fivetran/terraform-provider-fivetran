package fivetran

import (
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ID           = "id"
	CONNECTOR_ID = "connector_id"
)

func getConnectorSchemaStatus() *schema.Schema {
	var result = map[string]*schema.Schema{
		"setup_state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The current setup state of the connector. The available values are: <br /> - incomplete - the setup config is incomplete, the setup tests never succeeded <br /> - connected - the connector is properly set up <br /> - broken - the connector setup config is broken.",
		},
		"is_historical_sync": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The boolean specifying whether the connector should be triggered to re-sync all historical data. If you set this parameter to TRUE, the next scheduled sync will be historical. If the value is FALSE or not specified, the connector will not re-sync historical data. NOTE: When the value is TRUE, only the next scheduled sync will be historical, all subsequent ones will be incremental. This parameter is set to FALSE once the historical sync is completed.",
		},
		"sync_state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The current sync state of the connector. The available values are: <br /> - scheduled - the sync is waiting to be run <br /> - syncing - the sync is currently running <br /> - paused - the sync is currently paused <br /> - rescheduled - the sync is waiting until more API calls are available in the source service.",
		},
		"update_state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The current data update state of the connector. The available values are: <br /> - on_schedule - the sync is running smoothly, no delays <br /> - delayed - the data is delayed for a longer time than expected for the update.",
		},
		"tasks": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The collection of tasks for the connector",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"code": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Response status code",
					},
					"message": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Response status text",
					},
				},
			},
		},
		"warnings": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The collection of warnings for the connector",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"code": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Response status code",
					},
					"message": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Response status text",
					},
				},
			},
		},
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: result,
		},
	}
}

func getConnectorSchema(readonly bool, version int) map[string]*schema.Schema {
	// Common for Resource and Datasource
	var result = map[string]*schema.Schema{
		// Id
		"id": {
			Type:        schema.TypeString,
			Computed:    !readonly,
			Required:    readonly,
			Description: "The unique identifier for the connector within the Fivetran system.",
		},

		// Computed
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The name used both as the connector's name within the Fivetran system and as the source schema's name within your destination.",
		},
		"connected_by": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The unique identifier of the user who has created the connector in your account",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The timestamp of the time the connector was created in your account",
		},

		// Required
		"group_id": {
			Type:        schema.TypeString,
			Required:    !readonly,
			ForceNew:    !readonly,
			Computed:    readonly,
			Description: "The unique identifier for the Group (Destination) within the Fivetran system.",
		},
		"service": {
			Type:        schema.TypeString,
			Required:    !readonly,
			ForceNew:    !readonly,
			Computed:    readonly,
			Description: "The connector type name within the Fivetran system.",
		},
		"destination_schema": getConnectorDestinationSchema(readonly),

		// Config
		"config": getConnectorSchemaConfig(readonly),
	}

	if version == 0 {
		// Computed
		result["succeeded_at"] = &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The timestamp of the time the connector sync succeeded last time",
		}
		result["failed_at"] = &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The timestamp of the time the connector sync failed last time",
		}
		result["service_version"] = &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The connector type version within the Fivetran system.",
		}

		// Optional with default values in upstream
		result["sync_frequency"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    !readonly,
			Computed:    true,
			Description: "The connector sync frequency in minutes.",
		} // Default: 360
		result["schedule_type"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    !readonly,
			Computed:    true,
			Description: "The connector schedule configuration type. Supported values: auto, manual",
		} // Default: AUTO
		result["paused"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    !readonly,
			Computed:    true,
			Description: "Specifies whether the connector is paused",
		} // Default: false
		result["pause_after_trial"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    !readonly,
			Computed:    true,
			Description: "Specifies whether the connector should be paused after the free trial period has ende",
		} // Default: false
		// Optional nullable in upstream
		result["daily_sync_time"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    !readonly,
			Computed:    readonly,
			Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time",
		}

		result["status"] = getConnectorSchemaStatus()
	}

	// Resource specific
	if !readonly {
		result["auth"] = getConnectorSchemaAuth()
		result["trust_certificates"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specifies whether we should trust the certificate automatically. The default value is FALSE. If a certificate is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination certificate](https://fivetran.com/docs/rest-api/certificates#approveadestinationcertificate).",
		}
		result["trust_fingerprints"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specifies whether we should trust the SSH fingerprint automatically. The default value is FALSE. If a fingerprint is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination fingerprint](https://fivetran.com/docs/rest-api/certificates#approveadestinationfingerprint).",
		}
		result["run_setup_tests"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
		}

		// Internal resource attribute (no upstream value)
		result["last_updated"] = &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "",
		}
	}
	return result
}

func getConnectorSchemaAuth() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_access": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Your application client access fields.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "`Client ID` of your client application.",
							},
							"client_secret": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "`Client secret` of your client application.",
							},
							"user_agent": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Your company's name in your client application.",
							},
							"developer_token": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Your approved `Developer token` to connect to the API.",
							},
						},
					},
				},
				"refresh_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The long-lived `Refresh token` along with the `client_id` and `client_secret` parameters carry the information necessary to get a new access token for API resources.",
				},
				"access_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The `Access Token` carries the information necessary for API resources to fetch data.",
				},
				"realm_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "`Realm ID` of your application.",
				},
			},
		},
	}
}

func getConnectorDestinationSchema(readonly bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList, MaxItems: getMaxItems(readonly),
		Required: !readonly, Computed: readonly,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Optional:    !readonly,
					ForceNew:    !readonly,
					Computed:    readonly,
					Description: "The connector schema name in destination. Has to be unique within the group (destination). Required for connector creation.",
				},
				"table": {
					Type:        schema.TypeString,
					Optional:    !readonly,
					ForceNew:    !readonly,
					Computed:    readonly,
					Description: "The table name unique within the schema to which connector will sync the data. Required for connector creation.",
				},
				"prefix": {
					Type:        schema.TypeString,
					Optional:    !readonly,
					ForceNew:    !readonly,
					Computed:    readonly,
					Description: "The connector schema prefix has to be unique within the group (destination). Each replicated schema is prefixed with the provided value. Required for connector creation.",
				},
			},
		},
	}
}

func connectorRead(currentConfig *[]interface{}, resp connectors.DetailsWithCustomConfigNoTestsResponse, version int) (map[string]interface{}, error) {
	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	helpers.MapAddStr(msi, "id", resp.Data.ID)
	helpers.MapAddStr(msi, "group_id", resp.Data.GroupID)
	helpers.MapAddStr(msi, "service", resp.Data.Service)

	helpers.MapAddStr(msi, "name", resp.Data.Schema)

	ds, err := readDestinationSchema(resp.Data.Schema, resp.Data.Service)

	if err != nil {
		return nil, err
	}

	helpers.MapAddXInterface(msi, "destination_schema", ds)
	helpers.MapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	helpers.MapAddStr(msi, "created_at", resp.Data.CreatedAt.String())

	if version == 0 {
		helpers.MapAddStr(msi, "service_version", helpers.IntPointerToStr(resp.Data.ServiceVersion))
		helpers.MapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
		helpers.MapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
		helpers.MapAddStr(msi, "sync_frequency", helpers.IntPointerToStr(resp.Data.SyncFrequency))
		helpers.MapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
		helpers.MapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
		helpers.MapAddStr(msi, "paused", helpers.BoolPointerToStr(resp.Data.Paused))
		helpers.MapAddStr(msi, "pause_after_trial", helpers.BoolPointerToStr(resp.Data.PauseAfterTrial))

		helpers.MapAddXInterface(msi, "status", connectorReadStatus(&resp))
	}
	upstreamConfig := connectorReadCustomConfig(&resp, currentConfig, resp.Data.Service)

	if len(upstreamConfig) > 0 {
		helpers.MapAddXInterface(msi, "config", upstreamConfig)
	}

	return msi, nil
}

func readDestinationSchema(schema string, service string) ([]interface{}, error) {
	destination_schema := make([]interface{}, 1)

	ds := make(map[string]interface{})

	if _, ok := destinationSchemaFields[service]; !ok {
		return nil, fmt.Errorf("unknown service: `%v`", service)
	}

	if destinationSchemaFields[service]["schema_prefix"] {
		helpers.MapAddStr(ds, "prefix", schema)
	} else {
		s := strings.Split(schema, ".")
		helpers.MapAddStr(ds, "name", s[0])
		if len(s) > 1 && destinationSchemaFields[service]["table"] {
			helpers.MapAddStr(ds, "table", s[1])
		}
	}

	destination_schema[0] = ds
	return destination_schema, nil
}

// resourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func connectorReadStatus(resp *connectors.DetailsWithCustomConfigNoTestsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	helpers.MapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	helpers.MapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	helpers.MapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	helpers.MapAddStr(s, "is_historical_sync", helpers.BoolPointerToStr(resp.Data.Status.IsHistoricalSync))
	helpers.MapAddXInterface(s, "tasks", connectorReadStatusFlattenTasks(resp))
	helpers.MapAddXInterface(s, "warnings", connectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

func connectorReadStatusFlattenTasks(resp *connectors.DetailsWithCustomConfigNoTestsResponse) []interface{} {
	if len(resp.Data.Status.Tasks) < 1 {
		return make([]interface{}, 0)
	}

	tasks := make([]interface{}, len(resp.Data.Status.Tasks))
	for i, v := range resp.Data.Status.Tasks {
		task := make(map[string]interface{})
		helpers.MapAddStr(task, "code", v.Code)
		helpers.MapAddStr(task, "message", v.Message)

		tasks[i] = task
	}

	return tasks
}

func connectorReadStatusFlattenWarnings(resp *connectors.DetailsWithCustomConfigNoTestsResponse) []interface{} {
	if len(resp.Data.Status.Warnings) < 1 {
		return make([]interface{}, 0)
	}

	warnings := make([]interface{}, len(resp.Data.Status.Warnings))
	for i, v := range resp.Data.Status.Warnings {
		warning := make(map[string]interface{})
		helpers.MapAddStr(warning, "code", v.Code)
		helpers.MapAddStr(warning, "message", v.Message)

		warnings[i] = warning
	}

	return warnings
}
