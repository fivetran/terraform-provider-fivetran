package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Status         struct {
// 	SetupState       string `json:"setup_state"`
// 	SyncState        string `json:"sync_state"`
// 	UpdateState      string `json:"update_state"`
// 	IsHistoricalSync bool   `json:"is_historical_sync"`
// 	Tasks            []struct {
// 		Code    string `json:"code"`
// 		Message string `json:"message"`
// 	} `json:"tasks"`
// 	Warnings []struct {
// 		Code    string `json:"code"`
// 		Message string `json:"message"`
// 	} `json:"warnings"`
// } `json:"status"`

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id":              {Type: schema.TypeString, Required: true},
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
			"status": {Type: schema.TypeList, Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"setup_state":        {Type: schema.TypeString, Computed: true},
						"sync_state":         {Type: schema.TypeString, Computed: true},
						"update_state":       {Type: schema.TypeString, Computed: true},
						"is_historical_sync": {Type: schema.TypeString, Computed: true}, // REST API/go-fivetran bool
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
					},
				},
			},
			"config": {Type: schema.TypeList, Computed: true, // CONTINUE HERE
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// "setup_state":  {Type: schema.TypeString, Computed: true},
						// "sync_state":   {Type: schema.TypeString, Computed: true},
						// "update_state": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["group_id"] = resp.Data.GroupID
	msi["service"] = resp.Data.Service
	msi["service_version"] = resp.Data.ServiceVersion
	msi["schema"] = resp.Data.Schema
	msi["connected_by"] = resp.Data.ConnectedBy
	msi["created_at"] = resp.Data.CreatedAt.String()
	msi["succeeded_at"] = resp.Data.SucceededAt.String()
	msi["failed_at"] = resp.Data.FailedAt.String()
	msi["sync_frequency"] = resp.Data.SyncFrequency
	msi["schedule_type"] = resp.Data.ScheduleType
	msi["status"] = dataSourceConnectorReadStatus(&resp)
	// msi["config"] = ...(&resp) // CONTINUE HERE
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

// dataSourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func dataSourceConnectorReadStatus(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	var status []interface{}

	s := make(map[string]interface{})
	s["setup_state"] = resp.Data.Status.SetupState
	s["sync_state"] = resp.Data.Status.SyncState
	s["update_state"] = resp.Data.Status.UpdateState
	if resp.Data.Status.IsHistoricalSync != nil {
		s["is_historical_sync"] = boolToStr(*resp.Data.Status.IsHistoricalSync)
	}
	s["tasks"] = dataSourceConnectorReadStatusFlattenTasks(resp)
	s["warnings"] = dataSourceConnectorReadStatusFlattenWarnings(resp)
	status = append(status, s)

	return status
}

func dataSourceConnectorReadStatusFlattenTasks(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Status.Tasks) < 1 {
		return make([]interface{}, 0)
	}

	tasks := make([]interface{}, len(resp.Data.Status.Tasks), len(resp.Data.Status.Tasks))
	for i, v := range resp.Data.Status.Tasks {
		task := make(map[string]interface{})
		task["code"] = v.Code
		task["message"] = v.Message

		tasks[i] = task
	}

	return tasks
}

func dataSourceConnectorReadStatusFlattenWarnings(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Status.Warnings) < 1 {
		return make([]interface{}, 0)
	}

	warnings := make([]interface{}, len(resp.Data.Status.Warnings), len(resp.Data.Status.Warnings))
	for i, v := range resp.Data.Status.Warnings {
		warning := make(map[string]interface{})
		warning["code"] = v.Code
		warning["message"] = v.Message

		warnings[i] = warning
	}

	return warnings
}

// // dataSourceConnectorsMetadataFlattenMetadata receives a *fivetran.ConnectorsSourceMetadataResponse and returns a []interface{}
// // containing the data type accepted by the "sources" set.
// func dataSourceConnectorsMetadataFlattenMetadata1(resp *fivetran.ConnectorsSourceMetadataResponse) []interface{} {
// 	if resp.Data.Items == nil {
// 		return make([]interface{}, 0)
// 	}

// 	sources := make([]interface{}, len(resp.Data.Items), len(resp.Data.Items))
// 	for i, v := range resp.Data.Items {
// 		source := make(map[string]interface{})
// 		source["id"] = v.ID
// 		source["name"] = v.Name
// 		source["type"] = v.Type
// 		source["description"] = v.Description
// 		source["icon_url"] = v.IconURL
// 		source["link_to_docs"] = v.LinkToDocs
// 		source["link_to_erd"] = v.LinkToErd

// 		sources[i] = source
// 	}

// 	return sources
// }
