package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbtTransformation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDbtTransformationRead,
		Schema: map[string]*schema.Schema{
			"id":                {Type: schema.TypeString, Computed: true},
			"dbt_model_id":      {Type: schema.TypeString, Computed: true},
			"output_model_name": {Type: schema.TypeString, Computed: true},
			"dbt_project_id":    {Type: schema.TypeString, Computed: true},
			"last_run":          {Type: schema.TypeString, Computed: true},
			"next_run":          {Type: schema.TypeString, Computed: true},
			"status":            {Type: schema.TypeString, Computed: true},
			"schedule":          dataSourceDbtTransformationSchemaSchedule(),
			"run_tests":         {Type: schema.TypeString, Computed: true},
			"connector_ids":     {Type: schema.TypeString, Computed: true},
			"model_ids":         {Type: schema.TypeString, Optional: true},
		},
	}
}

func dataSourceDbtTransformationSchemaSchedule() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"schedule_type": {Type: schema.TypeString, Computed: true},
				"days_of_week":  {Type: schema.TypeString, Computed: true},
				"time_of_day":   {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

func dataSourceDbtTransformationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtTransformationDetailsService().DbtTransformationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for map string interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "dbt_model_id", resp.Data.DbtModelID)
	mapAddStr(msi, "output_model_name", resp.Data.OutputModelName)
	mapAddStr(msi, "dbt_project_id", resp.Data.DbtProjectId)
	mapAddStr(msi, "last_run", resp.Data.LastRun.String())
	mapAddStr(msi, "next_run", resp.Data.NextRun.String())
	mapAddStr(msi, "status", resp.Data.Status)
	mapAddXInterface(msi, "schedule", dataSourceDbtTransformationReadSchedule(&resp))
	mapAddStr(msi, "run_tests", boolToStr(resp.Data.RunTests))
	mapAddStr(msi, "connector_ids", resp.Data.ConnectorIds)
	mapAddStr(msi, "model_ids", resp.Data.ModelIds)

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

// dataSourceDbtTransformationReadSchedule receives a *fivetran.DbtTransformationDetailsResponse
// and returns a []interface{} containing the data type accepted by the "status" list.
func dataSourceDbtTransformationReadSchedule(resp *fivetran.DbtTransformationDetailsResponse) []interface{} {
	schedule := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "schedule_type", resp.Data.Schedule.ScheduleType)
	mapAddStr(s, "days_of_weel", resp.Data.Schedule.DaysOfWeek)
	mapAddStr(s, "interval", resp.Data.Schedule.Interval)
	mapAddStr(s, "time_of_day", resp.Data.Schedule.TimeOfDay)

	schedule[0] = s

	return schedule
}
