package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/groups"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupConnectors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupConnectorsRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schema": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional filter. When defined, the data source will only contain information for the connector with the specified schema name.",
			},
			"connectors": dataSourceGroupConnectorsSchemaConnectors(),
		},
	}
}

func dataSourceGroupConnectorsSchemaConnectors() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Set: func(v interface{}) int {
			return helpers.StringInt32Hash(v.(map[string]interface{})["id"].(string))
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the Connector within the Fivetran system.",
				},
				"group_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the Group within the Fivetran system.",
				},
				"service": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The connector type name within the Fivetran system",
				},
				"service_version": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The connector type version within the Fivetran system",
				},
				"schema": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name used both as the connector's name within the Fivetran system and as the source schema's name within your destination",
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
				"succeeded_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The timestamp of the time the connector sync succeeded last time",
				},
				"failed_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The timestamp of the time the connector sync failed last time",
				},
				"sync_frequency": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The connector sync frequency in minutes",
				},
				"schedule_type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The connector schedule configuration type. Supported values: auto, manual",
				},
				"status": {
					Type:        schema.TypeSet,
					Optional:    true,
					Computed:    true,
					Description: "",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"setup_state": {
								Type:        schema.TypeString,
								Computed:    true,
								Description: "The current setup state of the connector. The available values are: <br /> - incomplete - the setup config is incomplete, the setup tests never succeeded <br /> - connected - the connector is properly set up <br /> - broken - the connector setup config is broken.",
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
							"is_historical_sync": {
								Type:        schema.TypeBool,
								Computed:    true,
								Description: "The boolean specifying whether the connector should be triggered to re-sync all historical data. If you set this parameter to TRUE, the next scheduled sync will be historical. If the value is FALSE or not specified, the connector will not re-sync historical data. NOTE: When the value is TRUE, only the next scheduled sync will be historical, all subsequent ones will be incremental. This parameter is set to FALSE once the historical sync is completed.",
							},
							"tasks": {
								Type:        schema.TypeSet,
								Computed:    true,
								Optional:    true,
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
								Type:     schema.TypeSet,
								Optional: true,
								Computed: true,
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
						},
					},
				},
				"daily_sync_time": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time",
				},
			},
		},
	}
}

func dataSourceGroupConnectorsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	id := d.Get("id").(string)
	schema := d.Get("schema").(string)

	resp, err := dataSourceGroupConnectorsGetConnectors(client, id, schema, ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	flatConnectorsList := dataSourceGroupConnectorsFlattenConnectors(&resp)

	if err := d.Set("connectors", flatConnectorsList); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	d.SetId(id)

	return diags
}

// dataSourceGroupConnectorsFlattenConnectors receives a *fivetran.GroupListConnectorsResponse and returns a []interface{}
// containing the data type accepted by the "connectors" set.
func dataSourceGroupConnectorsFlattenConnectors(resp *groups.GroupListConnectorsResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	connectors := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		connector := make(map[string]interface{})
		connector["id"] = v.ID
		connector["group_id"] = v.GroupID
		connector["service"] = v.Service
		connector["service_version"] = v.ServiceVersion
		connector["schema"] = v.Schema
		connector["connected_by"] = v.ConnectedBy
		connector["created_at"] = v.CreatedAt.String()
		connector["succeeded_at"] = v.SucceededAt.String()
		connector["failed_at"] = v.FailedAt.String()
		connector["sync_frequency"] = v.SyncFrequency
		connector["schedule_type"] = v.ScheduleType
		connector["daily_sync_time"] = v.DailySyncTime

		// Status
		var statusTasks []interface{}
		for _, v := range v.Status.Tasks {
			t := make(map[string]interface{})
			t["code"] = v.Code
			t["message"] = v.Message
			statusTasks = append(statusTasks, t)
		}
		var statusWarnings []interface{}
		for _, v := range v.Status.Warnings {
			w := make(map[string]interface{})
			w["code"] = v.Code
			w["message"] = v.Message
			statusWarnings = append(statusWarnings, w)
		}
		var status []interface{}
		s := make(map[string]interface{})
		s["setup_state"] = v.Status.SetupState
		s["sync_state"] = v.Status.SyncState
		s["update_state"] = v.Status.UpdateState
		s["is_historical_sync"] = v.Status.IsHistoricalSync
		s["tasks"] = statusTasks
		s["warnings"] = statusWarnings
		status = append(status, s)
		connector["status"] = status

		connectors[i] = connector
	}

	return connectors
}

// dataSourceGroupConnectorsGetConnectors gets the connectors list of a group. It handles limits and cursors.
func dataSourceGroupConnectorsGetConnectors(client *fivetran.Client, id, schema string, ctx context.Context) (groups.GroupListConnectorsResponse, error) {
	var resp groups.GroupListConnectorsResponse
	var respNextCursor string

	for {
		var err error
		var respInner groups.GroupListConnectorsResponse
		svc := client.NewGroupListConnectors()
		if schema != "" {
			svc.Schema(schema)
		}
		if respNextCursor == "" {
			respInner, err = svc.GroupID(id).Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.GroupID(id).Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return groups.GroupListConnectorsResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
