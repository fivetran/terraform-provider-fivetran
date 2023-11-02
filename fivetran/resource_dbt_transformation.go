package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDbtTransformation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDbtTransformationCreate,
		ReadContext:          resourceDbtTransformationRead,
		UpdateContext:        resourceDbtTransformationUpdate,
		DeleteContext:        resourceDbtTransformationDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:               getDbtTransformationSchema(false),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func getDbtTransformationSchema(datasource bool) map[string]*schema.Schema {
	maxItems := 1
	if datasource {
		maxItems = 0
	}
	result := map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: !datasource,
			Required: datasource,
		},
		"dbt_project_id": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "The unique identifier for the dbt Project within the Fivetran system.",
		},
		"dbt_model_name": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "Target dbt Model name.",
		},

		"run_tests": {Type: schema.TypeBool, Required: !datasource, Computed: datasource, Description: "The field indicating whether the tests have been configured for dbt Transformation. By default, the value is false."},
		"paused":    {Type: schema.TypeBool, Required: !datasource, Computed: datasource, Description: "The field indicating whether the transformation will be created in paused state. By default, the value is false."},
		"schedule": {
			Type:        schema.TypeList,
			MaxItems:    maxItems,
			Required:    !datasource,
			Computed:    datasource,
			Description: "dbt Transformation schedule parameters.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"schedule_type": {Type: schema.TypeString, Required: !datasource, Computed: datasource, Description: "The type of the schedule to run the dbt Transformation on. The following values are supported: INTEGRATED, TIME_OF_DAY, INTERVAL. For INTEGRATED schedule type, interval and time_of_day values are ignored and only the days_of_week parameter values are taken into account (but may be empty or null). For TIME_OF_DAY schedule type, the interval parameter value is ignored and the time_of_day values is taken into account along with days_of_week value. For INTERVAL schedule type, time_of_day value is ignored and the interval parameter value is taken into account along with days_of_week value."},
					"days_of_week": {
						Type:        schema.TypeSet,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Optional:    !datasource,
						Computed:    true,
						Description: "The set of the days of the week the transformation should be launched on. The following values are supported: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.",
					},
					"interval":    {Type: schema.TypeInt, Computed: true, Optional: !datasource, Description: "The time interval in minutes between subsequent transformation runs."},
					"time_of_day": {Type: schema.TypeString, Computed: true, Optional: !datasource, Description: `The time of the day the transformation should be launched at. Supported values are: "00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00", "22:00", "23:00"`},
				},
			},
		},

		// resdonly fields
		"dbt_model_id":      {Type: schema.TypeString, Computed: true, Description: "The unique identifier for the dbt Model within the Fivetran system."},
		"output_model_name": {Type: schema.TypeString, Computed: true, Description: "The dbt Model name."},
		"created_at":        {Type: schema.TypeString, Computed: true, Description: "The timestamp of the dbt Transformation creation."},
		"connector_ids": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    true,
			Description: "Identifiers of related connectors.",
		},
		"model_ids": {
			Type:        schema.TypeSet,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    true,
			Description: "Identifiers of related models.",
		},
	}
	return result
}

func resourceDbtTransformationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	client := m.(*fivetran.Client)
	projectId := d.Get("dbt_project_id").(string)
	modelName := d.Get("dbt_model_name").(string)

	diags, ok := ensureProjectIsReady(ctx, client, projectId, nil, nil, nil)

	if !ok {
		return diags
	}

	var filteredModelId interface{} = nil

	for filteredModelId == nil {
		modelsResp, err := getAllDbtModelsForProject(client, ctx, projectId)
		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, modelsResp.Code, modelsResp.Message))
		}
		for _, model := range modelsResp.Data.Items {
			if model.ModelName == modelName {
				filteredModelId = model.ID
				break
			}
		}
		if filteredModelId != nil {
			break
		}
		if dl, ok := ctx.Deadline(); ok && time.Now().After(dl.Add(-20*time.Second)) {
			return helpers.NewDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("timed out: model with name %v not found in project %v.", modelName, projectId))
		}
		helpers.ContextDelay(ctx, 10*time.Second)
	}

	dbtModelId := filteredModelId.(string)

	svc := client.NewDbtTransformationCreateService()

	svc.DbtModelId(dbtModelId)
	svc.RunTests(d.Get("run_tests").(bool))

	schedule := d.Get("schedule").([]interface{})[0].(map[string]interface{})

	scheduleRequest := fivetran.NewDbtTransformationSchedule()

	// schedule_type is required for schedule
	scheduleRequest.ScheduleType(schedule["schedule_type"].(string))

	if v, ok := schedule["days_of_week"]; ok {
		scheduleRequest.DaysOfWeek(helpers.XInterfaceStrXStr(v.(*schema.Set).List()))
	}

	if v, ok := schedule["interval"].(int); ok && v > 0 {
		scheduleRequest.Interval(v)
	}

	if v, ok := schedule["time_of_day"].(string); ok {
		scheduleRequest.TimeOfDay(v)
	}

	svc.Schedule(scheduleRequest)

	svc.Paused(d.Get("paused").(bool))

	resp, err := svc.Do(ctx)

	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	return resourceDbtTransformationRead(ctx, d, m)
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
		return helpers.NewDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	modelResp, err := client.NewDbtModelDetails().ModelId(resp.Data.DbtModelId).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	mapStringInterface := make(map[string]interface{})
	helpers.MapAddStr(mapStringInterface, "dbt_model_name", modelResp.Data.ModelName)

	helpers.MapAddStr(mapStringInterface, "id", resp.Data.ID)
	helpers.MapAddStr(mapStringInterface, "dbt_model_id", resp.Data.DbtModelId)
	helpers.MapAddStr(mapStringInterface, "output_model_name", resp.Data.OutputModelName)
	helpers.MapAddStr(mapStringInterface, "dbt_project_id", resp.Data.DbtProjectId)
	helpers.MapAddStr(mapStringInterface, "created_at", resp.Data.CreatedAt)
	helpers.MapAddXString(mapStringInterface, "connector_ids", resp.Data.ConnectorIds)
	helpers.MapAddXString(mapStringInterface, "model_ids", resp.Data.ModelIds)

	mapStringInterface["run_tests"] = resp.Data.RunTests
	mapStringInterface["paused"] = resp.Data.Paused

	upstreamSchedule := make(map[string]interface{})
	upstreamSchedule["schedule_type"] = resp.Data.Schedule.ScheduleType
	upstreamSchedule["interval"] = resp.Data.Schedule.Interval
	upstreamSchedule["time_of_day"] = resp.Data.Schedule.TimeOfDay
	upstreamSchedule["days_of_week"] = resp.Data.Schedule.DaysOfWeek

	schedule := make([]interface{}, 0)
	mapStringInterface["schedule"] = append(schedule, upstreamSchedule)

	for k, v := range mapStringInterface {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
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
				days = append(days, day.(string))
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
		return helpers.NewDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	return resourceDbtTransformationRead(ctx, d, m)
}

func resourceDbtTransformationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtTransformationDeleteService()

	resp, err := svc.TransformationId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
