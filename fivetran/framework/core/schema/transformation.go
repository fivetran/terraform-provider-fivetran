package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TransformationResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: transformationSchema().GetResourceSchema(),
		Blocks:     transformationResourceBlocks(),
	}
}

func TransformationDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: transformationSchema().GetDatasourceSchema(),
		Blocks:     transformationDatasourceBlocks(),
	}
}

func TransformationListDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Blocks: map[string]datasourceSchema.Block{
			"transformations": datasourceSchema.ListNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: transformationSchema().GetDatasourceSchema(),
					Blocks:     transformationDatasourceBlocks(),
				},
			},
		},
	}
}

func transformationSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				ValueType:   core.String,
				IsId:        true,
				Description: "The unique identifier for the dbt Transformation within the Fivetran system.",
			},
			"paused": {
				ValueType:   core.Boolean,
				Description: "The field indicating whether the transformation will be set into the paused state. By default, the value is false.",
			},
			"type": {
				ValueType:   core.String,
				Description: "Transformation type.",
			},
			"created_at": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The timestamp of when the transformation was created in your account.",
			},
			"created_by_id": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The unique identifier for the User within the Fivetran system who created the transformation.",
			},
			"status": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Status of transformation Project (NOT_READY, READY, ERROR).",
			},
			"output_model_names": {
				ValueType:   core.StringsSet,
				Readonly:    true,
				Description: "Identifiers of related models.",
			},
		},
	}
}

func transformationScheduleSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"schedule_type": {
				ValueType:   core.String,
				Description: "The type of the schedule to run the dbt Transformation on. The following values are supported: INTEGRATED, TIME_OF_DAY, INTERVAL. For INTEGRATED schedule type, interval and time_of_day values are ignored and only the days_of_week parameter values are taken into account (but may be empty or null). For TIME_OF_DAY schedule type, the interval parameter value is ignored and the time_of_day values is taken into account along with days_of_week value. For INTERVAL schedule type, time_of_day value is ignored and the interval parameter value is taken into account along with days_of_week value.",
			},
			"days_of_week": {
				ValueType:   core.StringsSet,
				Description: "The set of the days of the week the transformation should be launched on. The following values are supported: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.",
			},
			"interval": {
				ValueType:   core.Integer,
				Description: "The time interval in minutes between subsequent transformation runs.",
			},
			"time_of_day": {
				ValueType:   core.String,
				Description: `The time of the day the transformation should be launched at. Supported values are: "00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00", "22:00", "23:00"`,
			},
			"connection_ids": {
				ValueType:   core.StringsSet,
				Description: "Identifiers of related connectors.",
			},
			"smart_syncing": {
				ValueType:   core.Boolean,
				Description: "The boolean flag that enables the Smart Syncing schedule",
			},
			"cron": {
				ValueType:   core.StringsSet,
				Description: "Cron schedule: list of CRON strings.",
			},
		},
	}
}

func transformationConfigDatasourceSchema() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"project_id": datasourceSchema.StringAttribute{
			Computed: true,
			Description: "The unique identifier for the dbt Core project within the Fivetran system",
		},
		"name": datasourceSchema.StringAttribute{
			Computed: true,
			Description: "The transformation name",
		},
		"package_name": datasourceSchema.StringAttribute{
			Computed: true,
			Description: `The Quickstart transformation package name`,
		},
		"connection_ids": datasourceSchema.SetAttribute{
			ElementType: basetypes.StringType{},
			Computed: true,
			Description: "The list of the connection identifiers to be used for the integrated schedule. Also used to identify package_name automatically if package_name was not specified",
		},
		"excluded_models": datasourceSchema.SetAttribute{
			ElementType: basetypes.StringType{},
			Computed: true,
			Description: "The list of excluded output model names",
		},
		"upgrade_available": datasourceSchema.BoolAttribute{
			Computed: true,
			Description: "The boolean flag indicating that a newer version is available for the transformation package",
		},
		"steps": datasourceSchema.ListNestedAttribute{
			Computed: true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: map[string]datasourceSchema.Attribute{
					"name": datasourceSchema.StringAttribute{
						Computed: true,
						Description: "The step name",
					},
					"command": datasourceSchema.StringAttribute{
						Computed: true,
						Description: "The dbt command in the transformation step",
					},
				},
			},
		},
	}
}

func transformationConfigResourceSchema() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"project_id": resourceSchema.StringAttribute{
			Optional:    true,
			Description: "The unique identifier for the dbt Core project within the Fivetran system",
		},
		"name": resourceSchema.StringAttribute{
			Optional:    true,
			Description: "The transformation name",
		},
		"package_name": resourceSchema.StringAttribute{
			Optional:    true,
			Description: `The Quickstart transformation package name`,
		},
		"connection_ids": resourceSchema.SetAttribute{
			Optional:    true,
			ElementType: basetypes.StringType{},
			Description: "The list of the connection identifiers to be used for the integrated schedule. Also used to identify package_name automatically if package_name was not specified",
		},
		"excluded_models": resourceSchema.SetAttribute{
			Optional:    true,
			ElementType: basetypes.StringType{},
			Description: "The list of excluded output model names",
		},
		"upgrade_available": resourceSchema.BoolAttribute{
			Computed: 	true,
			Optional:   true,
			Description: "The boolean flag indicating that a newer version is available for the transformation package",
		},
		"steps": resourceSchema.ListNestedAttribute{
			Optional:    true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: map[string]resourceSchema.Attribute{
					"name": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "The step name",
					},
					"command": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "The dbt command in the transformation step",
					},
				},
			},
		},
	}
}

func transformationResourceBlocks() map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"schedule": resourceSchema.SingleNestedBlock{
			Attributes: transformationScheduleSchema().GetResourceSchema(),
		},
		"transformation_config": resourceSchema.SingleNestedBlock{
			Attributes: transformationConfigResourceSchema(),
		},
	}
}

func transformationDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"schedule": datasourceSchema.SingleNestedBlock{
			Attributes: transformationScheduleSchema().GetDatasourceSchema(),
		},
		"transformation_config": datasourceSchema.SingleNestedBlock{
			Attributes: transformationConfigDatasourceSchema(),
		},
	}
}
