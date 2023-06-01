package fivetran

import (
	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ID           = "id"
	CONNECTOR_ID = "connector_id"
)

func getConnectorSchema(readonly bool, version int) map[string]*schema.Schema {
	// Common for Resource and Datasource
	var result = map[string]*schema.Schema{
		// Id
		"id": {Type: schema.TypeString, Computed: !readonly, Required: readonly},

		// Computed
		"name":         {Type: schema.TypeString, Computed: true},
		"connected_by": {Type: schema.TypeString, Computed: true},
		"created_at":   {Type: schema.TypeString, Computed: true},

		// Required
		"group_id":           {Type: schema.TypeString, Required: !readonly, ForceNew: !readonly, Computed: readonly},
		"service":            {Type: schema.TypeString, Required: !readonly, ForceNew: !readonly, Computed: readonly},
		"destination_schema": connectorDestinationSchemaSchema(readonly),

		// Config
		"config": getConnectorSchemaConfig(),
	}

	if version == 0 {
		// Computed
		result["succeeded_at"] = &schema.Schema{Type: schema.TypeString, Computed: true}
		result["failed_at"] = &schema.Schema{Type: schema.TypeString, Computed: true}
		result["service_version"] = &schema.Schema{Type: schema.TypeString, Computed: true}

		// Optional with default values in upstream
		result["sync_frequency"] = &schema.Schema{Type: schema.TypeString, Optional: !readonly, Computed: true}    // Default: 360
		result["schedule_type"] = &schema.Schema{Type: schema.TypeString, Optional: !readonly, Computed: true}     // Default: AUTO
		result["paused"] = &schema.Schema{Type: schema.TypeString, Optional: !readonly, Computed: true}            // Default: false
		result["pause_after_trial"] = &schema.Schema{Type: schema.TypeString, Optional: !readonly, Computed: true} // Default: false
		// Optional nullable in upstream
		result["daily_sync_time"] = &schema.Schema{Type: schema.TypeString, Optional: !readonly, Computed: readonly}
		result["status"] = connectorSchemaStatus()
	}

	// Resource specific
	if !readonly {
		result["auth"] = connectorSchemaAuth()
		result["trust_certificates"] = &schema.Schema{Type: schema.TypeString, Optional: true}
		result["trust_fingerprints"] = &schema.Schema{Type: schema.TypeString, Optional: true}
		result["run_setup_tests"] = &schema.Schema{Type: schema.TypeString, Optional: true}

		// Internal resource attribute (no upstream value)
		result["last_updated"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	}
	return result
}

func connectorSchemaStatus() *schema.Schema {
	var result = map[string]*schema.Schema{
		"setup_state":        {Type: schema.TypeString, Computed: true},
		"is_historical_sync": {Type: schema.TypeString, Computed: true},
		"sync_state":         {Type: schema.TypeString, Computed: true},
		"update_state":       {Type: schema.TypeString, Computed: true},
		"tasks": {Type: schema.TypeList, Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"code":    {Type: schema.TypeString, Computed: true},
					"message": {Type: schema.TypeString, Computed: true},
				},
			},
		},
		"warnings": {Type: schema.TypeList, Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"code":    {Type: schema.TypeString, Computed: true},
					"message": {Type: schema.TypeString, Computed: true},
				},
			},
		},
	}

	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: result,
		},
	}
}

func getMaxItems(readonly bool) int {
	if readonly {
		return 0
	}
	return 1
}

func connectorDestinationSchemaSchema(readonly bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList, MaxItems: getMaxItems(readonly),
		Required: !readonly, Computed: readonly,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
				"table":  {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
				"prefix": {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
			},
		},
	}
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
	upstreamConfig := getConnectorReadCustomConfig(&resp, currentConfig)

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
