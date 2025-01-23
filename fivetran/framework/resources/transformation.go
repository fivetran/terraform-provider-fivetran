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

	client := r.GetClient()
	svc := client.NewTransformationCreate()
	svc.ProjectType(data.ProjectType.ValueString())
	svc.Paused(data.Paused.ValueBool())

	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		config := fivetran.NewTransformationConfig()
		configAttributes := data.Config.Attributes()
		if !configAttributes["project_id"].(basetypes.StringValue).IsNull() && !configAttributes["project_id"].(basetypes.StringValue).IsUnknown() {
			config.ProjectId(configAttributes["project_id"].(basetypes.StringValue).ValueString())			
		}
		if !configAttributes["name"].(basetypes.StringValue).IsNull() && !configAttributes["name"].(basetypes.StringValue).IsUnknown() {
			config.Name(configAttributes["name"].(basetypes.StringValue).ValueString())			
		}
		if !configAttributes["package_name"].(basetypes.StringValue).IsNull() && !configAttributes["package_name"].(basetypes.StringValue).IsUnknown() {
			config.PackageName(configAttributes["package_name"].(basetypes.StringValue).ValueString())			
		}

		if !configAttributes["connection_ids"].IsUnknown() && !configAttributes["connection_ids"].IsNull() {
			evars := []string{}
			for _, ev := range configAttributes["connection_ids"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			config.ConnectionIds(evars)
		}

		if !configAttributes["excluded_models"].IsUnknown() && !configAttributes["excluded_models"].IsNull() {
			evars := []string{}
			for _, ev := range configAttributes["excluded_models"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			config.ExcludedModels(evars)
		}

		if !configAttributes["steps"].IsUnknown() && !configAttributes["steps"].IsNull() {
			evars := []transformations.TransformationStep{}
			for _, ev := range configAttributes["steps"].(basetypes.SetValue).Elements() {
				if element, ok := ev.(basetypes.ObjectValue); ok {
					step := transformations.TransformationStep{}
					step.Name = element.Attributes()["name"].(basetypes.StringValue).ValueString()
					step.Command = element.Attributes()["command"].(basetypes.StringValue).ValueString()
					evars = append(evars, step)
				}
			}
			config.Steps(evars)
		}

		svc.TransformationConfig(config)
	}

	if !data.Schedule.IsNull() && !data.Schedule.IsUnknown() {
		schedule := fivetran.NewTransformationSchedule()
		scheduleAttributes := data.Schedule.Attributes()

		if !scheduleAttributes["time_of_day"].(basetypes.StringValue).IsNull() && !scheduleAttributes["time_of_day"].(basetypes.StringValue).IsUnknown() {
			schedule.TimeOfDay(scheduleAttributes["time_of_day"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["schedule_type"].(basetypes.StringValue).IsNull() && !scheduleAttributes["schedule_type"].(basetypes.StringValue).IsUnknown() {
			schedule.ScheduleType(scheduleAttributes["schedule_type"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["interval"].(basetypes.Int64Value).IsNull() && !scheduleAttributes["interval"].(basetypes.Int64Value).IsUnknown() {
			schedule.Interval(int(scheduleAttributes["interval"].(basetypes.Int64Value).ValueInt64()))
		}
		if !scheduleAttributes["smart_syncing"].(basetypes.BoolValue).IsNull() && !scheduleAttributes["smart_syncing"].(basetypes.BoolValue).IsUnknown() {
			schedule.SmartSyncing(scheduleAttributes["smart_syncing"].(basetypes.BoolValue).ValueBool())			
		}

		if !scheduleAttributes["connection_ids"].IsUnknown() && !scheduleAttributes["connection_ids"].IsNull() {
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

	svc := r.GetClient().NewTransformationUpdate()
	svc.Paused(plan.Paused.ValueBool())

	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		config := fivetran.NewTransformationConfig()
		configAttributes := plan.Config.Attributes()
		if !configAttributes["project_id"].(basetypes.StringValue).IsNull() && !configAttributes["project_id"].(basetypes.StringValue).IsUnknown() {
			config.ProjectId(configAttributes["project_id"].(basetypes.StringValue).ValueString())			
		}
		if !configAttributes["name"].(basetypes.StringValue).IsNull() && !configAttributes["name"].(basetypes.StringValue).IsUnknown() {
			config.Name(configAttributes["name"].(basetypes.StringValue).ValueString())			
		}
		if !configAttributes["package_name"].(basetypes.StringValue).IsNull() && !configAttributes["package_name"].(basetypes.StringValue).IsUnknown() {
			config.PackageName(configAttributes["package_name"].(basetypes.StringValue).ValueString())			
		}

		if !configAttributes["connection_ids"].IsUnknown() && !configAttributes["connection_ids"].IsNull() {
			evars := []string{}
			for _, ev := range configAttributes["connection_ids"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			config.ConnectionIds(evars)
		}

		if !configAttributes["excluded_models"].IsUnknown() && !configAttributes["excluded_models"].IsNull() {
			evars := []string{}
			for _, ev := range configAttributes["excluded_models"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			config.ExcludedModels(evars)
		}

		if !configAttributes["steps"].IsUnknown() && !configAttributes["steps"].IsNull() {
			evars := []transformations.TransformationStep{}
			for _, ev := range configAttributes["steps"].(basetypes.SetValue).Elements() {
				if element, ok := ev.(basetypes.ObjectValue); ok {
					var step transformations.TransformationStep
					step.Name = element.Attributes()["name"].(basetypes.StringValue).ValueString()
					step.Command = element.Attributes()["command"].(basetypes.StringValue).ValueString()
					evars = append(evars, step)
				}
			}
			config.Steps(evars)
		}

		svc.TransformationConfig(config)
	}

	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() {
		schedule := fivetran.NewTransformationSchedule()
		scheduleAttributes := plan.Schedule.Attributes()

		if !scheduleAttributes["time_of_day"].(basetypes.StringValue).IsNull() && !scheduleAttributes["time_of_day"].(basetypes.StringValue).IsUnknown() {
			schedule.TimeOfDay(scheduleAttributes["time_of_day"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["schedule_type"].(basetypes.StringValue).IsNull() && !scheduleAttributes["schedule_type"].(basetypes.StringValue).IsUnknown() {
			schedule.ScheduleType(scheduleAttributes["schedule_type"].(basetypes.StringValue).ValueString())			
		}
		if !scheduleAttributes["interval"].(basetypes.Int64Value).IsNull() && !scheduleAttributes["interval"].(basetypes.Int64Value).IsUnknown() {
			schedule.Interval(int(scheduleAttributes["interval"].(basetypes.Int64Value).ValueInt64()))
		}
		if !scheduleAttributes["smart_syncing"].(basetypes.BoolValue).IsNull() && !scheduleAttributes["smart_syncing"].(basetypes.BoolValue).IsUnknown() {
			schedule.SmartSyncing(scheduleAttributes["smart_syncing"].(basetypes.BoolValue).ValueBool())			
		}

		if !scheduleAttributes["connection_ids"].IsUnknown() && !scheduleAttributes["connection_ids"].IsNull() {
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

	updateResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(ctx, updateResponse)

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
