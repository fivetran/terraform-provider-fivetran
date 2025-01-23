package model

import (
	"context"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Transformations struct {
	Transformations types.List `tfsdk:"transformations"`
}

var (
	elemTypeAttrs = map[string]attr.Type{
        "id":    					types.StringType,
        "status": 					types.StringType,
        "type": 					types.StringType,
        "created_at": 				types.StringType,
        "created_by_id": 			types.StringType,
        "paused": 					types.BoolType,
        "output_model_names": 		types.SetType{ElemType: types.StringType},
        "schedule": 				types.ObjectType{AttrTypes: scheduleAttrs},
        "transformation_config": 	types.ObjectType{AttrTypes: configAttrs},
    }
)
func (d *Transformations) ReadFromResponse(ctx context.Context, resp transformations.TransformationsListResponse) {
	if resp.Data.Items == nil {
		d.Transformations = types.ListNull(types.ObjectType{AttrTypes: elemTypeAttrs})
	} else {
		items := []attr.Value{}
		for _, v := range resp.Data.Items {
			item := map[string]attr.Value{}
			item["id"] = types.StringValue(v.Id)
			item["status"] = types.StringValue(v.Status)
			item["type"] = types.StringValue(v.ProjectType)
			item["created_at"] = types.StringValue(v.CreatedAt)
			item["created_by_id"] = types.StringValue(v.CreatedById)
			item["paused"] = types.BoolValue(v.Paused)
			
			if v.OutputModelNames != nil {
				item["output_model_names"] = types.SetValueMust(types.StringType, stringListToAttrList(v.OutputModelNames))
			} else {
				item["output_model_names"] = types.SetNull(types.StringType)
			}

			scheduleAttrValues := map[string]attr.Value{}
			scheduleAttrValues["smart_syncing"] = types.BoolValue(v.TransformationSchedule.SmartSyncing)

			if v.TransformationSchedule.ScheduleType == "INTERVAL" || v.TransformationSchedule.Interval > 0 {
				scheduleAttrValues["interval"] = types.Int64Value(int64(v.TransformationSchedule.Interval))
			} else {
				scheduleAttrValues["interval"] = types.Int64Null()
			}
	
			if v.TransformationSchedule.TimeOfDay != "" {
				scheduleAttrValues["time_of_day"] = types.StringValue(v.TransformationSchedule.TimeOfDay)
			} else {
				scheduleAttrValues["time_of_day"] = types.StringNull()
			}
	
			if v.TransformationSchedule.ScheduleType != "" {
				scheduleAttrValues["schedule_type"] = types.StringValue(v.TransformationSchedule.ScheduleType)
			} else {
				scheduleAttrValues["schedule_type"] = types.StringNull()
			}

			if v.TransformationSchedule.Cron != nil {
		    	vars := []attr.Value{}
    			for _, el := range v.TransformationSchedule.Cron {
		        	vars = append(vars, types.StringValue(el))
    			}
		    	if len(vars) > 0 {
        			scheduleAttrValues["cron"] = types.SetValueMust(types.StringType, vars)
		    	} else {
        			scheduleAttrValues["cron"] = types.SetNull(types.StringType)
		    	}
			} else {
				scheduleAttrValues["cron"] = types.SetNull(types.StringType)
			}

			if v.TransformationSchedule.ConnectionIds != nil {
    			vars := []attr.Value{}
		    	for _, el := range v.TransformationSchedule.ConnectionIds {
        			vars = append(vars, types.StringValue(el))
		    	}
    			if len(vars) > 0 {
		        	scheduleAttrValues["connection_ids"] = types.SetValueMust(types.StringType, vars)
    			} else {
		        	scheduleAttrValues["connection_ids"] = types.SetNull(types.StringType)
    			}
			} else {
				scheduleAttrValues["connection_ids"] = types.SetNull(types.StringType)
			}

			if v.TransformationSchedule.DaysOfWeek != nil {
    			vars := []attr.Value{}
		    	for _, el := range v.TransformationSchedule.DaysOfWeek {
        			vars = append(vars, types.StringValue(el))
		    	}
    			if len(vars) > 0 {
		        	scheduleAttrValues["days_of_week"] = types.SetValueMust(types.StringType, vars)
    			} else {
        			scheduleAttrValues["days_of_week"] = types.SetNull(types.StringType)
		    	}
			} else {
				scheduleAttrValues["days_of_week"] = types.SetNull(types.StringType)
			}
	
			item["schedule"] = types.ObjectValueMust(scheduleAttrs, scheduleAttrValues)

			configAttrValues := map[string]attr.Value{}
			configAttrValues["upgrade_available"] = types.BoolValue(v.TransformationConfig.UpgradeAvailable)
			if v.TransformationConfig.ProjectId != "" {
				configAttrValues["project_id"] = types.StringValue(v.TransformationConfig.ProjectId)
			} else {
				configAttrValues["project_id"] = types.StringNull()
			}
	
			if v.TransformationConfig.PackageName != "" {
				configAttrValues["package_name"] = types.StringValue(v.TransformationConfig.PackageName)
			} else {
				configAttrValues["package_name"] = types.StringNull()
			}

			if v.TransformationConfig.Name != "" {
				configAttrValues["name"] = types.StringValue(v.TransformationConfig.Name)
			} else {
				configAttrValues["name"] = types.StringNull()
			}

			if v.TransformationConfig.ConnectionIds != nil {
    			vars := []attr.Value{}
		    	for _, el := range v.TransformationConfig.ConnectionIds {
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

			if v.TransformationConfig.ExcludedModels != nil {
    			vars := []attr.Value{}
		    	for _, el := range v.TransformationConfig.ExcludedModels {
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

		    subItems := []attr.Value{}
		    for _, sub := range v.TransformationConfig.Steps {
    			subItem := map[string]attr.Value{}
		        subItem["name"] = types.StringValue(sub.Name)
        		subItem["command"] = types.StringValue(sub.Command)

		        subObjectValue, _ := types.ObjectValue(stepAttrTypes, subItem)
        		subItems = append(subItems, subObjectValue)
		    }
		   	configAttrValues["steps"], _ = types.ListValue(stepSetAttrType, subItems)

			item["transformation_config"] = types.ObjectValueMust(configAttrs, configAttrValues)

			objectValue, _ := types.ObjectValue(elemTypeAttrs, item)
			items = append(items, objectValue)
		}
		d.Transformations, _ = types.ListValue(types.ObjectType{AttrTypes: elemTypeAttrs}, items)
	}
}