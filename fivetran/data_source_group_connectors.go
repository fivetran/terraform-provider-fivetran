package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupConnectors() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupConnectorsRead,
		Schema: map[string]*schema.Schema{
			"id":     {Type: schema.TypeString, Required: true},
			"schema": {Type: schema.TypeString, Optional: true},
			"connectors": {Type: schema.TypeSet, Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":              {Type: schema.TypeString, Computed: true},
						"group_id":        {Type: schema.TypeString, Computed: true},
						"service":         {Type: schema.TypeString, Computed: true},
						"service_version": {Type: schema.TypeInt, Computed: true},
						"schema":          {Type: schema.TypeString, Computed: true},
						"connected_by":    {Type: schema.TypeString, Computed: true},
						"created_at":      {Type: schema.TypeString, Computed: true},
						"succeeded_at":    {Type: schema.TypeString, Computed: true},
						"failed_at":       {Type: schema.TypeString, Computed: true},
						"sync_frequency":  {Type: schema.TypeInt, Computed: true},
						"schedule_type":   {Type: schema.TypeString, Computed: true},
						"status": {Type: schema.TypeSet, Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"setup_state":        {Type: schema.TypeString, Computed: true},
									"sync_state":         {Type: schema.TypeString, Computed: true},
									"update_state":       {Type: schema.TypeString, Computed: true},
									"is_historical_sync": {Type: schema.TypeBool, Computed: true},
									"tasks": {Type: schema.TypeSet, Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"code":    {Type: schema.TypeString, Computed: true},
												"message": {Type: schema.TypeString, Computed: true},
											},
										},
									},
									"warnings": {Type: schema.TypeSet, Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"code":    {Type: schema.TypeString, Computed: true},
												"message": {Type: schema.TypeString, Computed: true},
											},
										},
									},
								},
							},
						},
						"daily_sync_time": {Type: schema.TypeString, Computed: true},
					},
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
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("connectors", dataSourceGroupConnectorsFlattenConnectors(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	d.SetId(id)

	return diags
}

// dataSourceGroupConnectorsFlattenConnectors receives a *fivetran.GroupListConnectorsResponse and returns a []interface{}
// containing the data type accepted by the "connectors" set.
func dataSourceGroupConnectorsFlattenConnectors(resp *fivetran.GroupListConnectorsResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	connectors := make([]interface{}, len(resp.Data.Items), len(resp.Data.Items))
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
func dataSourceGroupConnectorsGetConnectors(client *fivetran.Client, id, schema string, ctx context.Context) (fivetran.GroupListConnectorsResponse, error) {
	var resp fivetran.GroupListConnectorsResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.GroupListConnectorsResponse
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
			return fivetran.GroupListConnectorsResponse{}, err
		}

		for _, item := range respInner.Data.Items {
			resp.Data.Items = append(resp.Data.Items, item)
		}

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
