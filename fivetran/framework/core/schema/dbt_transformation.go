package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DbtTransformationResourceSchema(ctx context.Context) resourceSchema.Schema {
	return resourceSchema.Schema{
		DeprecationMessage: "This resource is Deprecated, please follow the 1.5.0 migration guide to update the schema",
		Attributes: dbtTransformationSchema().GetResourceSchema(),
		Blocks:     dbtTransformationResourceBlocks(ctx),
	}
}

func DbtTransformationDatasourceSchema() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		DeprecationMessage: "This datasource is Deprecated, please follow the 1.5.0 migration guide to update the schema",
		Attributes: dbtTransformationSchema().GetDatasourceSchema(),
		Blocks:     dbtTransformationDatasourceBlocks(),
	}
}

func dbtTransformationSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				ValueType:   core.String,
				IsId:        true,
				Description: "The unique identifier for the dbt Transformation within the Fivetran system.",
			},
			"dbt_project_id": {
				ValueType:   core.String,
				ForceNew:    true,
				Required:    true,
				Description: "The unique identifier for the dbt Project within the Fivetran system.",
			},
			"dbt_model_name": {
				ValueType:   core.String,
				ForceNew:    true,
				Required:    true,
				Description: "Target dbt Model name.",
			},
			"run_tests": {
				ValueType:   core.Boolean,
				Description: "The field indicating whether the tests have been configured for dbt Transformation. By default, the value is false.",
			},
			"paused": {
				ValueType:   core.Boolean,
				Description: "The field indicating whether the transformation will be set into the paused state. By default, the value is false.",
			},
			"dbt_model_id": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The unique identifier for the dbt Model within the Fivetran system.",
			},
			"output_model_name": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The dbt Model name.",
			},
			"created_at": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The timestamp of the dbt Transformation creation.",
			},
			"connector_ids": {
				ValueType:   core.StringsSet,
				Readonly:    true,
				Description: "Identifiers of related connectors.",
			},
			"model_ids": {
				ValueType:   core.StringsSet,
				Readonly:    true,
				Description: "Identifiers of related models.",
			},
		},
	}
}

func dbtTransformationScheduleSchema() core.Schema {
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
		},
	}
}

func dbtTransformationResourceBlocks(ctx context.Context) map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"schedule": resourceSchema.SingleNestedBlock{
			Attributes: dbtTransformationScheduleSchema().GetResourceSchema(),
		},
		"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true}),
	}
}

func dbtTransformationDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"schedule": datasourceSchema.SingleNestedBlock{
			Attributes: dbtTransformationScheduleSchema().GetDatasourceSchema(),
		},
	}
}
