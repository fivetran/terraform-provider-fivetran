package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDbtTransformation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDbtTransformationCreate,
		ReadContext:   resourceDbtTransformationRead,
		UpdateContext: resourceDbtTransformationUpdate,
		DeleteContext: resourceDbtTransformationDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":                {Type: schema.TypeString, Computed: true},
			"dbt_model_id":      {Type: schema.TypeString, Computed: true},
			"output_model_name": {Type: schema.TypeString, Computed: true},
			"dbt_project_id":    {Type: schema.TypeString, Computed: true},
			"last_run":          {Type: schema.TypeString, Computed: true},
			"next_run":          {Type: schema.TypeString, Computed: true},
			"status":            {Type: schema.TypeString, Computed: true},
			"schedule":          resourceDbtTransformationSchemaSchedule(),
			"run_tests":         {Type: schema.TypeString, Computed: true},
			"connector_ids":     {Type: schema.TypeString, Computed: true},
			"model_ids":         {Type: schema.TypeString, Optional: true},
		},
	}
}

func resourceDbtTransformationSchemaSchedule() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"schedule_type": {Type: schema.TypeString, Computed: true},
				"days_of_week":  {Type: schema.TypeString, Computed: true},
				"interval":      {Type: schema.TypeString, Computed: true},
				"time_of_day":   {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

func resourceDbtTransformationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationCreate()

	svc.DbtModelID(d.Get("resourceDbtTransformationCreate"))
	svc.RunTests(strToBool(d.Get("run_tests").(string)))

	transformationSchedule := resourceDbtTransformationUpdateSchedule(d)

	svc.Schedule(transformationSchedule)

	d.SetId(resp.Data.ID)
	resourceDbtTransformationRead(ctx, d, m)

	return diags
}

func resourceDbtTransformationUpdateSchedule(d *schema.ResourceData) *DbtTransformationSchedule {
	fivetranSchedule := NewDbtTransformationSchedule()
	var schedule = d.Get("schedule").([]interface{})

	if len(schedule) < 1 {
		return fivetranSchedule
	}

	if schedule[0] == nil {
		return fivetranSchedule
	}

	c := schedule[0].(map[string]interface{})

	if v := c["schedule_type"].(string); v != "" {
		fivetranSchedule.ScheduleType(v)
	}

	if v := c["days_of_week"].([]interface{}); len(v) > 0 {
		fivetranSchedule.DaysOfWeek(xInterfaceStrXStr(v))
	}

	if v := c["time_of_day"].(string); v != "" {
		fivetranSchedule.TimeOfDay()
	}

	return fivetranSchedule
}

func resourceDbtTransformationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtTransformationDetails().DbtTransformationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	mapStringInterface := make(map[string]interface{})
	mapAddStr(mapStringInterface, "id", resp.Data.ID)
	mapAddStr(mapStringInterface, "dbt_model_id", resp.Data.DbtModelID)
	mapAddStr(mapStringInterface, "output_model_name", resp.Data.DbtModelID)
	mapAddStr(mapStringInterface, "dbt_project_id", resp.Data.DbtProjectId)
	mapAddStr(mapStringInterface, "last_run", resp.Data.LastRun.String())
	mapAddStr(mapStringInterface, "next_run", resp.Data.NextRun.String())
	mapAddStr(mapStringInterface, "status", resp.Data.Status)
	mapAddStr(mapStringInterface, "run_tests", resp.Data.RunTests)
	mapAddXInterface(mapStringInterface, "connector_ids", resp.Data.ConnectorsIds.([]interface{}))
	mapAddXInterface(mapStringInterface, "models_ids", resp.Data.ModelIds.([]interface{}))

	currentSchedule := d.Get("schedule").([]interface{})
	upstreamSchedule := resourceDbtTransformationReadSchedule(&resp, currentSchedule)

	if len(upstreamSchedule) > 0 {
		mapAddXInterface(mapStringInterface, "schedule", upstreamSchedule)
	}

	for k, v := range mapStringInterface {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

// resourceDbtTransformationReadSchedule receives a *fivetran.DbtTransformationDetailsResponse
// and return a []interface{} containing the data type accepted by the "config" list.
func resourceDbtTransformationReadSchedule(resp *DbtTransformationDetailsResponse, currentSchedule []interface{}) []interface{} {
	schedule := make([]interface{}, 1)

	c := make(map[string]interface{})

	mapAddStr(c, "schedule_type", resp.Data.Schedule.ScheduleType)
	mapAddXInterface(c, "days_of_week", xStrXInterface(resp.Data.Schedule.DaysOfWeek))
	mapAddStr(c, "interval", resp.Data.Schedule.Interval)
	mapAddStr(c, "time_of_day", resp.Data.Schedule.TimeOfDay)

	schedule[0] = c

	return schedule
}

func resourceDbtTransformationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationModifyService()

	svc.DbtTransformationID(d.Get("id").(string))

	if d.HasChange("run_tests") {
		svc.RunTests(strToBool(d.Get("run_tests").(string)))
	}

	svc.Schedule(resourceDbtTransformationUpdateSchedule(d))

	resp, err := svc.Do(ctx)
	if err != nil {
		// make sure the state is updated after a newDbtTransformationModify error.
		diags = resourceDbtTransformationRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	return resourceDbtTransformationRead(ctx, d, m)
}

func resourceDbtTransformationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationDeleteService()

	resp, err := svc.DbtTransformationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
