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
			"id": {Type: schema.TypeString, Computed: true},

			"dbt_model_id": {Type: schema.TypeString, Required: true, ForceNew: true},

			"run_tests": {Type: schema.TypeBool, Required: true},
			"paused":    {Type: schema.TypeBool, Required: true},

			"schedule": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_type": {Type: schema.TypeString, Required: true},
						"days_of_week": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							Computed: true,
						},
						"interval":    {Type: schema.TypeInt, Computed: true, Optional: true},
						"time_of_day": {Type: schema.TypeString, Computed: true, Optional: true},
					},
				},
			},

			"dbt_project_id":    {Type: schema.TypeString, Computed: true},
			"output_model_name": {Type: schema.TypeString, Computed: true},
			"created_at":        {Type: schema.TypeString, Computed: true},
			"connector_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"model_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func resourceDbtTransformationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationCreateService()

	dbtModelId := d.Get("dbt_model_id").(string)

	svc.DbtModelId(dbtModelId)
	svc.RunTests(d.Get("run_tests").(bool))

	schedule := d.Get("schedule").([]interface{})[0].(map[string]interface{})
	svc.Schedule(createDbtTransformationSchedule(schedule))

	svc.Paused(d.Get("paused").(bool))

	resp, err := svc.Do(ctx)

	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	return resourceDbtTransformationRead(ctx, d, m)
}

func createDbtTransformationSchedule(s map[string]interface{}) *fivetran.DbtTransformationSchedule {
	result := fivetran.NewDbtTransformationSchedule()

	// schedule_type is required for schedule
	result.ScheduleType(s["schedule_type"].(string))

	if v, ok := s["days_of_week"]; ok {
		result.DaysOfWeek(xInterfaceStrXStr(v.(*schema.Set).List()))
	}

	if v, ok := s["interval"].(int); ok && v > 0 {
		result.Interval(v)
	}

	if v, ok := s["time_of_day"].(string); ok {
		result.TimeOfDay(v)
	}
	return result
}

func resourceDbtTransformationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtTransformationDetailsService().TransformationId(d.Get("id").(string)).Do(ctx)
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
	mapAddStr(mapStringInterface, "dbt_model_id", resp.Data.DbtModelId)
	mapAddStr(mapStringInterface, "output_model_name", resp.Data.OutputModelName)
	mapAddStr(mapStringInterface, "dbt_project_id", resp.Data.DbtProjectId)
	mapStringInterface["run_tests"] = resp.Data.RunTests
	mapStringInterface["paused"] = resp.Data.Paused
	mapAddXString(mapStringInterface, "connector_ids", resp.Data.ConnectorIds)
	mapAddXString(mapStringInterface, "model_ids", resp.Data.ModelIds)

	upstreamSchedule := make(map[string]interface{})
	upstreamSchedule["schedule_type"] = resp.Data.Schedule.ScheduleType
	upstreamSchedule["interval"] = resp.Data.Schedule.Interval
	upstreamSchedule["time_of_day"] = resp.Data.Schedule.TimeOfDay
	upstreamSchedule["days_of_week"] = resp.Data.Schedule.DaysOfWeek

	schedule := make([]interface{}, 0)
	mapStringInterface["schedule"] = append(schedule, upstreamSchedule)

	for k, v := range mapStringInterface {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceDbtTransformationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationModifyService()

	svc.DbtTransformationId(d.Get("id").(string))

	if d.HasChange("run_tests") {
		svc.RunTests(d.Get("run_tests").(bool))
	}

	if d.HasChange("paused") {
		svc.Paused(d.Get("paused").(bool))
	}

	if d.HasChanges("schedule.0.schedule_type", "schedule.0.days_of_week", "schedule.0.interval", "schedule.0.time_of_day") {
		schedule := fivetran.NewDbtTransformationSchedule()
		scheduleType := d.Get("schedule.0.schedule_type").(string)
		if d.HasChange("schedule.0.schedule_type") {
			schedule.ScheduleType(scheduleType)
		}
		if d.HasChange("schedule.0.days_of_week") {
			days := make([]string, 0)
			for _, day := range d.Get("schedule.0.days_of_week").(*schema.Set).List() {
				day = append(days, day.(string))
			}
			schedule.DaysOfWeek(days)
		}
		if d.HasChange("schedule.0.interval") {
			schedule.Interval(d.Get("schedule.0.interval").(int))
		}
		if d.HasChange("schedule.0.time_of_day") {
			schedule.TimeOfDay(d.Get("schedule.0.time_of_day").(string))
		}
		svc.Schedule(schedule)
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		// make sure the state is updated after a newDbtTransformationModify error.
		diags = resourceDbtTransformationRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	return resourceDbtTransformationRead(ctx, d, m)
}

func resourceDbtTransformationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationDeleteService()

	resp, err := svc.TransformationId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
