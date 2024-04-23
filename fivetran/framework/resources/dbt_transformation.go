package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/datasources"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func DbtTransformation() resource.Resource {
	return &dbtTransformation{}
}

type dbtTransformation struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &dbtTransformation{}
var _ resource.ResourceWithImportState = &dbtTransformation{}

func (r *dbtTransformation) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_transformation"
}

func (r *dbtTransformation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtTransformationResourceSchema(ctx)
}

func (r *dbtTransformation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dbtTransformation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtTransformationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectId := data.DbtProjectId.ValueString()

	if ok := ensureProjectIsReady(resp, ctx, r.GetClient(), projectId); !ok {
		resp.Diagnostics.AddError(
			"Dbt Project is not Ready",
			fmt.Sprintf("Dbt project `%v` is not in READY status.", projectId),
		)
		return
	}

	var filteredModelId interface{} = nil
	modelName := data.DbtModelName.ValueString()

	for filteredModelId == nil {
		modelsResp, err := datasources.GetAllDbtModelsForProject(r.GetClient(), ctx, projectId, 1000)
		if err != nil {
			resp.Diagnostics.AddError(
				"DbtProject Models Read Error.",
				fmt.Sprintf("%v; code: %v; message: %v", err, modelsResp.Code, modelsResp.Message),
			)
			return
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
			resp.Diagnostics.AddError(
				"Unable to create dbt Transformation.",
				fmt.Sprintf("Unable to fetch models from specified project. Timed out: model with name %v not found in project %v.", modelName, projectId),
			)
			return
		}
		helpers.ContextDelay(ctx, 10*time.Second)
	}

	dbtModelId := filteredModelId.(string)

	svc := r.GetClient().NewDbtTransformationCreateService()
	svc.DbtModelId(dbtModelId)
	svc.RunTests(data.RunTests.ValueBool())
	svc.Paused(data.Paused.ValueBool())

	scheduleRequest := fivetran.NewDbtTransformationSchedule()

	scheduleTypeAttr := data.Schedule.Attributes()["schedule_type"].(basetypes.StringValue)

	if scheduleTypeAttr.IsUnknown() || scheduleTypeAttr.IsNull() {
		resp.Diagnostics.AddError(
			"Unable to create dbt Transformation.",
			"Field `schedule.schedule_type` is required.",
		)
		return
	}

	scheduleRequest.ScheduleType(scheduleTypeAttr.ValueString())

	dofAttr := data.Schedule.Attributes()["days_of_week"]
	if !dofAttr.IsUnknown() && !dofAttr.IsNull() {
		dof := []string{}
		for _, d := range dofAttr.(basetypes.SetValue).Elements() {
			dof = append(dof, d.(basetypes.StringValue).ValueString())
		}
		scheduleRequest.DaysOfWeek(dof)
	}
	intervalAttr := data.Schedule.Attributes()["interval"].(basetypes.Int64Value)
	if !intervalAttr.IsNull() && !intervalAttr.IsUnknown() {
		scheduleRequest.Interval(int(intervalAttr.ValueInt64()))
	}

	todAttr := data.Schedule.Attributes()["time_of_day"].(basetypes.StringValue)
	if !todAttr.IsNull() && !todAttr.IsUnknown() {
		scheduleRequest.TimeOfDay(todAttr.ValueString())
	}
	svc.Schedule(scheduleRequest)

	transformationResp, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create dbt Transformation resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, transformationResp.Code, transformationResp.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, transformationResp, nil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		deleteResponse, err := r.GetClient().NewDbtTransformationDeleteService().TransformationId(transformationResp.Data.ID).Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Cleanup dbt Transformation Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
			)
		}
	}
}

func (r *dbtTransformation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtTransformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	transformationResponse, err := r.GetClient().NewDbtTransformationDetailsService().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read dbt Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, transformationResponse.Code, transformationResponse.Message),
		)
	}

	data.ReadFromResponse(ctx, transformationResponse, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dbtTransformation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var state model.DbtTransformationResourceModel
	var plan model.DbtTransformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewDbtTransformationModifyService()

	svc.DbtTransformationId(state.Id.ValueString())

	if !plan.Paused.Equal(state.Paused) {
		svc.Paused(plan.Paused.ValueBool())
	}

	if !plan.RunTests.Equal(state.RunTests) {
		svc.RunTests(plan.RunTests.ValueBool())
	}

	if !plan.Schedule.Equal(state.Schedule) {
		planScheduleAttrs := plan.Schedule.Attributes()
		stateScheduleAttrs := state.Schedule.Attributes()
		schedule := fivetran.NewDbtTransformationSchedule()
		if !planScheduleAttrs["schedule_type"].Equal(stateScheduleAttrs["schedule_type"]) {
			schedule.ScheduleType(planScheduleAttrs["schedule_type"].(basetypes.StringValue).ValueString())
		}
		if !planScheduleAttrs["interval"].Equal(stateScheduleAttrs["interval"]) {
			schedule.Interval(int(planScheduleAttrs["interval"].(basetypes.Int64Value).ValueInt64()))
		}
		if !planScheduleAttrs["time_of_day"].Equal(stateScheduleAttrs["time_of_day"]) {
			schedule.TimeOfDay(planScheduleAttrs["time_of_day"].(basetypes.StringValue).ValueString())
		}
		if !planScheduleAttrs["days_of_week"].Equal(stateScheduleAttrs["days_of_week"]) {
			dof := []string{}
			for _, d := range planScheduleAttrs["days_of_week"].(basetypes.SetValue).Elements() {
				dof = append(dof, d.(basetypes.StringValue).ValueString())
			}
			schedule.DaysOfWeek(dof)
		}
		svc.Schedule(schedule)
	}

	updatedResp, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update dbt Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, updatedResp.Code, updatedResp.Message),
		)
		return
	}

	plan.ReadFromResponse(ctx, updatedResp, nil)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dbtTransformation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtTransformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	deleteResponse, err := r.GetClient().NewDbtTransformationDeleteService().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete dbt Transformation Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
	}
}
