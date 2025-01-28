package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/transformations"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func Transformation() resource.Resource {
	return &transformation{}
}

type transformation struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &transformation{}
var _ resource.ResourceWithImportState = &transformation{}

func (r *transformation) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fivetran_transformation"
}

func (r *transformation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.TransformationResource()
}

func (r *transformation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *transformation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Transformation
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	transformationType := data.ProjectType.ValueString()
	client := r.GetClient()
	svc := client.NewTransformationCreate()
	svc.ProjectType(transformationType)
	svc.Paused(data.Paused.ValueBool())

	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		config := fivetran.NewTransformationConfig()
		configAttributes := data.Config.Attributes()
		/* DBT_CORE */
		if !configAttributes["project_id"].(basetypes.StringValue).IsNull() && !configAttributes["project_id"].(basetypes.StringValue).IsUnknown() {
			if transformationType != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "transformation_config.project_id"),
				)
				return
			}

			config.ProjectId(configAttributes["project_id"].(basetypes.StringValue).ValueString())
		}

		if !configAttributes["name"].(basetypes.StringValue).IsNull() && !configAttributes["name"].(basetypes.StringValue).IsUnknown() {
			if transformationType != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "transformation_config.name"),
				)
				return
			}

			config.Name(configAttributes["name"].(basetypes.StringValue).ValueString())			
		} 

		if !configAttributes["steps"].IsUnknown() && !configAttributes["steps"].IsNull() {
			if transformationType != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "transformation_config.steps"),
				)
				return
			}

			evars := []transformations.TransformationStep{}

			for _, ev := range configAttributes["steps"].(basetypes.ListValue).Elements() {
				if element, ok := ev.(basetypes.ObjectValue); ok {
					step := transformations.TransformationStep{}
					step.Name = element.Attributes()["name"].(basetypes.StringValue).ValueString()
					step.Command = element.Attributes()["command"].(basetypes.StringValue).ValueString()
					evars = append(evars, step)
				}
			}

			config.Steps(evars)
		}

		/* QUICKSTART */
		packageName := ""
		if !configAttributes["package_name"].(basetypes.StringValue).IsNull() && !configAttributes["package_name"].(basetypes.StringValue).IsUnknown() {
			if transformationType != "QUICKSTART" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for QUICKSTART type transformation", "transformation_config.package_name"),
				)
				return
			}

			packageName = configAttributes["package_name"].(basetypes.StringValue).ValueString()
			
			config.PackageName(packageName)
		}

		connectionIds := []string{}
		if !configAttributes["connection_ids"].IsUnknown() && !configAttributes["connection_ids"].IsNull() {
			if transformationType != "QUICKSTART" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for QUICKSTART type transformation", "transformation_config.connection_ids"),
				)
				return
			}

			for _, ev := range configAttributes["connection_ids"].(basetypes.SetValue).Elements() {
				connectionIds = append(connectionIds, ev.(basetypes.StringValue).ValueString())
			}


			config.ConnectionIds(connectionIds)
		}

		if len(connectionIds) == 0 && packageName == "" && transformationType == "QUICKSTART" {
			resp.Diagnostics.AddError(
				"Unable to Create Transformation Resource.",
				fmt.Sprintf("For a QUICKSTART type transformation, at least one of the `%v` or `%v` parameters must be set.", "transformation_config.package_name", "transformation_config.connection_ids"),
			)
			return
		}

		if !configAttributes["excluded_models"].IsUnknown() && !configAttributes["excluded_models"].IsNull() {
			if transformationType != "QUICKSTART" {
				resp.Diagnostics.AddError(
					"Unable to Create Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for QUICKSTART type transformation", "transformation_config.excluded_models"),
				)
				return
			}

			evars := []string{}
			for _, ev := range configAttributes["excluded_models"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			config.ExcludedModels(evars)
		}

		svc.TransformationConfig(config)
	}

	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		schedule := fivetran.NewTransformationSchedule()
		scheduleAttributes := data.Schedule.Attributes()

		if !scheduleAttributes["time_of_day"].IsNull() && !scheduleAttributes["time_of_day"].IsUnknown() {
			schedule.TimeOfDay(scheduleAttributes["time_of_day"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["schedule_type"].IsNull() && !scheduleAttributes["schedule_type"].IsUnknown() {
			schedule.ScheduleType(scheduleAttributes["schedule_type"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["interval"].IsNull() && !scheduleAttributes["interval"].IsUnknown() {
			schedule.Interval(int(scheduleAttributes["interval"].(basetypes.Int64Value).ValueInt64()))
		}
		if !scheduleAttributes["smart_syncing"].IsNull() && !scheduleAttributes["smart_syncing"].IsUnknown() {
			schedule.SmartSyncing(scheduleAttributes["smart_syncing"].(basetypes.BoolValue).ValueBool())			
		}

		if !scheduleAttributes["connection_ids"].IsUnknown() && !scheduleAttributes["connection_ids"].IsNull() {
			if transformationType != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Update Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "schedule.connection_ids"),
				)
				return
			}

			evars := []string{}
			for _, ev := range scheduleAttributes["connection_ids"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			schedule.ConnectionIds(evars)
		}

		if !scheduleAttributes["days_of_week"].IsUnknown() && !scheduleAttributes["days_of_week"].IsNull() {
			evars := []string{}
			for _, ev := range scheduleAttributes["days_of_week"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			schedule.DaysOfWeek(evars)
		}

		if !scheduleAttributes["cron"].IsUnknown() && !scheduleAttributes["cron"].IsNull() {
			evars := []string{}
			for _, ev := range scheduleAttributes["cron"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			schedule.Cron(evars)
		}

		svc.TransformationSchedule(schedule)
	}

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromResponse(ctx, createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		// Do cleanup on error
		deleteResponse, err := client.NewTransformationDelete().TransformationId(createResponse.Data.Id).Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Cleanup Transformation Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
			)
		}
	}
}

func (r *transformation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Transformation

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readResponse, err := r.GetClient().NewTransformationDetails().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, readResponse.Code, readResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, readResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *transformation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var state model.Transformation
	var plan model.Transformation

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewTransformationUpdate().TransformationId(state.Id.ValueString())

	hasChanges := false
	pausedPlan := core.GetBoolOrDefault(plan.Paused, true)
	pausedState := core.GetBoolOrDefault(state.Paused, true)

	if pausedPlan != pausedState {
		svc.Paused(pausedPlan)
		hasChanges = true
	}

	if !plan.Config.IsNull() && !plan.Config.IsUnknown() && !plan.Config.Equal(state.Config) {
		config := fivetran.NewTransformationConfig()
		configPlanAttributes := plan.Config.Attributes()
		configStateAttributes := state.Config.Attributes()

		if !configPlanAttributes["name"].IsNull() &&
		!configPlanAttributes["name"].IsUnknown() && 
		!configStateAttributes["name"].(basetypes.StringValue).Equal(configPlanAttributes["name"].(basetypes.StringValue)) {
			if state.ProjectType.ValueString() != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Update Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "transformation_config.name"),
				)
				return
			}

			hasChanges = true
			config.Name(configPlanAttributes["name"].(basetypes.StringValue).ValueString())			
		}

		if !configPlanAttributes["steps"].IsUnknown() &&
		!configPlanAttributes["steps"].IsNull() &&
		!configStateAttributes["steps"].(basetypes.ListValue).Equal(configPlanAttributes["steps"].(basetypes.ListValue)) {
			if state.ProjectType.ValueString() != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Update Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "transformation_config.steps"),
				)
				return
			}

			evars := []transformations.TransformationStep{}
			for _, ev := range configPlanAttributes["steps"].(basetypes.ListValue).Elements() {
				if element, ok := ev.(basetypes.ObjectValue); ok {
					var step transformations.TransformationStep
					step.Name = element.Attributes()["name"].(basetypes.StringValue).ValueString()
					step.Command = element.Attributes()["command"].(basetypes.StringValue).ValueString()
					evars = append(evars, step)
				}
			}

			hasChanges = true
			config.Steps(evars)
		}

		if !configPlanAttributes["excluded_models"].IsUnknown() && 
		!configPlanAttributes["excluded_models"].IsNull() &&
		!configStateAttributes["excluded_models"].(basetypes.SetValue).Equal(configPlanAttributes["excluded_models"].(basetypes.SetValue)) {
			if state.ProjectType.ValueString() != "QUICKSTART" {
				resp.Diagnostics.AddError(
					"Unable to Update Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for QUICKSTART type transformation", "transformation_config.excluded_models"),
				)
				return
			}

			evars := []string{}
			for _, ev := range configPlanAttributes["excluded_models"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}

			hasChanges = true
			config.ExcludedModels(evars)
		}

		if hasChanges {
			svc.TransformationConfig(config)
		}
	}

	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() && !plan.Schedule.Equal(state.Schedule) {
		schedule := fivetran.NewTransformationSchedule()
		schedulePlanAttributes := plan.Schedule.Attributes()
		scheduleStateAttributes := state.Schedule.Attributes()

		if !schedulePlanAttributes["time_of_day"].IsNull() && 
		!schedulePlanAttributes["time_of_day"].IsUnknown() &&
		!scheduleStateAttributes["time_of_day"].(basetypes.StringValue).Equal(schedulePlanAttributes["time_of_day"].(basetypes.StringValue)) {
			hasChanges = true
			schedule.TimeOfDay(schedulePlanAttributes["time_of_day"].(basetypes.StringValue).ValueString())
		}

		if !schedulePlanAttributes["schedule_type"].IsNull() &&
		!schedulePlanAttributes["schedule_type"].IsUnknown() &&
		!scheduleStateAttributes["schedule_type"].(basetypes.StringValue).Equal(schedulePlanAttributes["schedule_type"].(basetypes.StringValue)) {
			hasChanges = true
			schedule.ScheduleType(schedulePlanAttributes["schedule_type"].(basetypes.StringValue).ValueString())			
		}

		if !schedulePlanAttributes["interval"].IsNull() &&
		!schedulePlanAttributes["interval"].IsUnknown() &&
		!scheduleStateAttributes["interval"].(basetypes.Int64Value).Equal(schedulePlanAttributes["interval"].(basetypes.Int64Value)) {
			hasChanges = true
			schedule.Interval(int(schedulePlanAttributes["interval"].(basetypes.Int64Value).ValueInt64()))
		}

		if !schedulePlanAttributes["smart_syncing"].IsNull() &&
		!schedulePlanAttributes["smart_syncing"].IsUnknown() &&
		!scheduleStateAttributes["smart_syncing"].(basetypes.BoolValue).Equal(schedulePlanAttributes["smart_syncing"].(basetypes.BoolValue)) {
			hasChanges = true
			schedule.SmartSyncing(schedulePlanAttributes["smart_syncing"].(basetypes.BoolValue).ValueBool())			
		}

		if !schedulePlanAttributes["connection_ids"].IsUnknown() &&
		!schedulePlanAttributes["connection_ids"].IsNull() &&
		!scheduleStateAttributes["connection_ids"].(basetypes.SetValue).Equal(schedulePlanAttributes["connection_ids"].(basetypes.SetValue)) {
			if plan.ProjectType.ValueString() != "DBT_CORE" {
				resp.Diagnostics.AddError(
					"Unable to Update Transformation Resource.",
					fmt.Sprintf("The parameter `%v` can be set only for DBT_CORE type transformation", "schedule.connection_ids"),
				)
				return
			}

			evars := []string{}
			for _, ev := range schedulePlanAttributes["connection_ids"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			hasChanges = true
			schedule.ConnectionIds(evars)
		}

		if !schedulePlanAttributes["days_of_week"].IsUnknown() &&
		!schedulePlanAttributes["days_of_week"].IsNull() &&
		!scheduleStateAttributes["days_of_week"].(basetypes.SetValue).Equal(schedulePlanAttributes["days_of_week"].(basetypes.SetValue)) {
			evars := []string{}
			for _, ev := range schedulePlanAttributes["days_of_week"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			hasChanges = true
			schedule.DaysOfWeek(evars)
		}

		if !schedulePlanAttributes["cron"].IsUnknown() &&
		!schedulePlanAttributes["cron"].IsNull() &&
		!scheduleStateAttributes["cron"].(basetypes.SetValue).Equal(schedulePlanAttributes["cron"].(basetypes.SetValue)) {
			evars := []string{}
			for _, ev := range schedulePlanAttributes["cron"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			hasChanges = true
			schedule.Cron(evars)
		}

		if hasChanges {
			svc.TransformationSchedule(schedule)
		}
	}

	if hasChanges {
		updateResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Transformation Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
			)
			return
		}

		plan.ReadFromResponse(ctx, updateResponse)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *transformation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Transformation

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewTransformationDelete().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
