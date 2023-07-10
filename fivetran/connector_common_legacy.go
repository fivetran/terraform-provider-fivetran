package fivetran

import (
	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func connectorSchemaLegacy(readonly bool, version int) map[string]*schema.Schema {
	// Common for Resource and Datasource
	var result = map[string]*schema.Schema{
		// Id
		"id": {
			Type:        schema.TypeString,
			Computed:    !readonly,
			Required:    readonly,
			Description: "The unique identifier for the user within the account.",
		},

		// Computed
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The unique identifier for the team within the account",
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
			Description: "The unique identifier for the Group within the Fivetran system.",
		},
		"service": {
			Type:        schema.TypeString,
			Required:    !readonly,
			ForceNew:    !readonly,
			Computed:    readonly,
			Description: "The connector type name within the Fivetran system",
		},
		"destination_schema": getConnectorDestinationSchema(readonly),

		// Config
		"config": connectorSchemaConfig(readonly),
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
			Description: "The connector sync frequency in minutes",
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

func connectorRead(currentConfig *[]interface{}, resp fivetran.ConnectorCustomDetailsResponse, version int) map[string]interface{} {
	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)
	mapAddStr(msi, "service", resp.Data.Service)

	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", readDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())

	if version == 0 {
		mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
		mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
		mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
		mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
		mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
		mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
		mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
		mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))

		mapAddXInterface(msi, "status", connectorReadStatus(&resp))
	}
	upstreamConfig := connectorReadCustomConfig(&resp, currentConfig)

	if len(upstreamConfig) > 0 {
		mapAddXInterface(msi, "config", upstreamConfig)
	}

	return msi
}

// resourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func connectorReadStatus(resp *fivetran.ConnectorCustomDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", connectorReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", connectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

func connectorReadStatusFlattenTasks(resp *fivetran.ConnectorCustomDetailsResponse) []interface{} {
	if len(resp.Data.Status.Tasks) < 1 {
		return make([]interface{}, 0)
	}

	tasks := make([]interface{}, len(resp.Data.Status.Tasks))
	for i, v := range resp.Data.Status.Tasks {
		task := make(map[string]interface{})
		mapAddStr(task, "code", v.Code)
		mapAddStr(task, "message", v.Message)

		tasks[i] = task
	}

	return tasks
}

func connectorReadStatusFlattenWarnings(resp *fivetran.ConnectorCustomDetailsResponse) []interface{} {
	if len(resp.Data.Status.Warnings) < 1 {
		return make([]interface{}, 0)
	}

	warnings := make([]interface{}, len(resp.Data.Status.Warnings))
	for i, v := range resp.Data.Status.Warnings {
		warning := make(map[string]interface{})
		mapAddStr(warning, "code", v.Code)
		mapAddStr(warning, "message", v.Message)

		warnings[i] = warning
	}

	return warnings
}
