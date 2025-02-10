package model

import (
    "context"

    sdk "github.com/fivetran/go-fivetran/transformations"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type Transformation struct {
    Id                  types.String `tfsdk:"id"`
    Status              types.String `tfsdk:"status"`
    ProjectType         types.String `tfsdk:"type"`
    Paused              types.Bool   `tfsdk:"paused"`
    CreatedAt           types.String `tfsdk:"created_at"`
    CreatedById         types.String `tfsdk:"created_by_id"`
    OutputModelNames    types.Set    `tfsdk:"output_model_names"`
    Schedule            types.Object `tfsdk:"schedule"`
    Config              types.Object `tfsdk:"transformation_config"`
}

var (
    stepAttrTypes = map[string]attr.Type{
        "name":    types.StringType,
        "command": types.StringType,
    }

    stepSetAttrType = types.ObjectType{
        AttrTypes: stepAttrTypes,
    }

    scheduleAttrs = map[string]attr.Type{
        "schedule_type":    types.StringType,
        "days_of_week":     types.SetType{ElemType: types.StringType},
        "cron":             types.SetType{ElemType: types.StringType},
        "connection_ids":   types.SetType{ElemType: types.StringType},
        "interval":         types.Int64Type,
        "time_of_day":      types.StringType,
        "smart_syncing":    types.BoolType,
    }

    configAttrs = map[string]attr.Type{
        "project_id":           types.StringType,
        "package_name":         types.StringType,
        "name":                 types.StringType,
        "excluded_models":      types.SetType{ElemType: types.StringType},
        "connection_ids":       types.SetType{ElemType: types.StringType},
        "steps":                types.ListType{ElemType: types.ObjectType{AttrTypes: stepAttrTypes}},
        "upgrade_available":    types.BoolType,
    }
)

func (d *Transformation) ReadFromResponse(ctx context.Context, resp sdk.TransformationResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Status = types.StringValue(resp.Data.Status)
    d.ProjectType = types.StringValue(resp.Data.ProjectType)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
    d.CreatedById = types.StringValue(resp.Data.CreatedById)
    d.Paused = types.BoolValue(resp.Data.Paused)

    if resp.Data.OutputModelNames != nil {
        d.OutputModelNames = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.OutputModelNames))
    } else {
        d.OutputModelNames = types.SetNull(types.StringType)
    }

    scheduleAttrValues := map[string]attr.Value{}
    scheduleAttrValues["smart_syncing"] = types.BoolValue(resp.Data.TransformationSchedule.SmartSyncing)

    if resp.Data.TransformationSchedule.ScheduleType == "INTERVAL" || resp.Data.TransformationSchedule.Interval > 0 {
        scheduleAttrValues["interval"] = types.Int64Value(int64(resp.Data.TransformationSchedule.Interval))
    } else {
        if !d.Schedule.Attributes()["interval"].IsUnknown() {
            scheduleAttrValues["interval"] = d.Schedule.Attributes()["interval"]
        } else {
            scheduleAttrValues["interval"] = types.Int64Null()
        }
    }
    
    if resp.Data.TransformationSchedule.TimeOfDay != "" {
        scheduleAttrValues["time_of_day"] = types.StringValue(resp.Data.TransformationSchedule.TimeOfDay)
    } else {
        if !d.Schedule.Attributes()["time_of_day"].IsUnknown() {
            scheduleAttrValues["time_of_day"] = d.Schedule.Attributes()["time_of_day"]
        } else {
            scheduleAttrValues["time_of_day"] = types.StringNull()
        }
    }
    
    if resp.Data.TransformationSchedule.ScheduleType != "" {
        scheduleAttrValues["schedule_type"] = types.StringValue(resp.Data.TransformationSchedule.ScheduleType)
    } else {
        if !d.Schedule.Attributes()["schedule_type"].IsUnknown() {
            scheduleAttrValues["schedule_type"] = d.Schedule.Attributes()["schedule_type"]
        } else {
            scheduleAttrValues["schedule_type"] = types.StringNull()
        }
    }

    if resp.Data.TransformationSchedule.Cron != nil {
        vars := []attr.Value{}
        for _, el := range resp.Data.TransformationSchedule.Cron {
            vars = append(vars, types.StringValue(el))
        }
        if len(vars) > 0 {
            scheduleAttrValues["cron"] = types.SetValueMust(types.StringType, vars)
        } else {
            scheduleAttrValues["cron"] = types.SetNull(types.StringType)
        }
    } else {
        if !d.Schedule.Attributes()["cron"].IsUnknown() {
            scheduleAttrValues["cron"] = d.Schedule.Attributes()["cron"]
        } else {
            scheduleAttrValues["cron"] = types.SetNull(types.StringType)
        }
    }

    if resp.Data.TransformationSchedule.ConnectionIds != nil && resp.Data.ProjectType == "DBT_CORE" {
        vars := []attr.Value{}
        for _, el := range resp.Data.TransformationSchedule.ConnectionIds {
            vars = append(vars, types.StringValue(el))
        }
        if len(vars) > 0 {
            scheduleAttrValues["connection_ids"] = types.SetValueMust(types.StringType, vars)
        } else {
            scheduleAttrValues["connection_ids"] = types.SetNull(types.StringType)
        }
    } else {
        if !d.Schedule.Attributes()["connection_ids"].IsUnknown() {
            scheduleAttrValues["connection_ids"] = d.Schedule.Attributes()["connection_ids"]
        } else {
            scheduleAttrValues["connection_ids"] = types.SetNull(types.StringType)
        }
    }

    if resp.Data.TransformationSchedule.DaysOfWeek != nil {
        vars := []attr.Value{}
        for _, el := range resp.Data.TransformationSchedule.DaysOfWeek {
            vars = append(vars, types.StringValue(el))
        }
        if len(vars) > 0 {
            scheduleAttrValues["days_of_week"] = types.SetValueMust(types.StringType, vars)
        } else {
            scheduleAttrValues["days_of_week"] = types.SetNull(types.StringType)
        }
    } else {
        if !d.Schedule.Attributes()["days_of_week"].IsUnknown() {
            scheduleAttrValues["days_of_week"] = d.Schedule.Attributes()["days_of_week"]
        } else {
            scheduleAttrValues["days_of_week"] = types.SetNull(types.StringType)
        }
    }
    
    d.Schedule = types.ObjectValueMust(scheduleAttrs, scheduleAttrValues)

    configAttrValues := map[string]attr.Value{}
    configAttrValues["upgrade_available"] = types.BoolValue(resp.Data.TransformationConfig.UpgradeAvailable)
    if resp.Data.TransformationConfig.ProjectId != "" {
        configAttrValues["project_id"] = types.StringValue(resp.Data.TransformationConfig.ProjectId)
    } else {
        configAttrValues["project_id"] = types.StringNull()
    }
    
    if resp.Data.TransformationConfig.PackageName != "" {
        configAttrValues["package_name"] = types.StringValue(resp.Data.TransformationConfig.PackageName)
    } else {
        configAttrValues["package_name"] = types.StringNull()
    }

    if resp.Data.TransformationConfig.Name != "" {
        configAttrValues["name"] = types.StringValue(resp.Data.TransformationConfig.Name)
    } else {
        configAttrValues["name"] = types.StringNull()
    }

    if resp.Data.TransformationConfig.ConnectionIds != nil {
        vars := []attr.Value{}
        for _, el := range resp.Data.TransformationConfig.ConnectionIds {
            vars = append(vars, types.StringValue(el))
        }
        if len(vars) > 0 {
            configAttrValues["connection_ids"] = types.SetValueMust(types.StringType, vars)
        } else {
            configAttrValues["connection_ids"] = types.SetNull(types.StringType)
        }
    } else {
        configAttrValues["connection_ids"] = types.SetNull(types.StringType)
    }

    if resp.Data.TransformationConfig.ExcludedModels != nil {
        vars := []attr.Value{}
        for _, el := range resp.Data.TransformationConfig.ExcludedModels {
            vars = append(vars, types.StringValue(el))
        }
        if len(vars) > 0 {
            configAttrValues["excluded_models"] = types.SetValueMust(types.StringType, vars)
        } else {
            configAttrValues["excluded_models"] = types.SetNull(types.StringType)
        }
    } else {
        configAttrValues["excluded_models"] = types.SetNull(types.StringType)
    }

    if resp.Data.TransformationConfig.Steps != nil {
        subItems := []attr.Value{}
        for _, sub := range resp.Data.TransformationConfig.Steps {
            subItem := map[string]attr.Value{}
            subItem["name"] = types.StringValue(sub.Name)
            subItem["command"] = types.StringValue(sub.Command)

            subObjectValue, _ := types.ObjectValue(stepAttrTypes, subItem)
            subItems = append(subItems, subObjectValue)
        }
        configAttrValues["steps"], _ = types.ListValue(stepSetAttrType, subItems)        
    } else {
        configAttrValues["steps"] = types.ListNull(stepSetAttrType)
    }

    d.Config = types.ObjectValueMust(configAttrs, configAttrValues)
}
