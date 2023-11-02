package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/dbt"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbtModels() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDbtModelsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier for the dbt project within the Fivetran system.",
			},
			"models": dbtModelsSchema(),
		},
	}
}

func dbtModelsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Computed:    true,
		Description: "The collection of dbt Models.",
		Set: func(v interface{}) int {
			return helpers.StringInt32Hash(v.(map[string]interface{})["id"].(string))
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the dbt Model within the Fivetran system.",
				},
				"model_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The dbt Model name.",
				},
				"scheduled": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Boolean specifying whether the model is selected for execution.",
				},
			},
		},
	}
}

func dataSourceDbtModelsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := getAllDbtModelsForProject(client, ctx, d.Get("project_id").(string))
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("models", flattenDbtModels(resp)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID
	d.SetId("0")

	return diags
}

// dataSourceGroupsGetGroups gets the groups list. It handles limits and cursors.
func getAllDbtModelsForProject(client *fivetran.Client, ctx context.Context, projectId string) (dbt.DbtModelsListResponse, error) {
	var resp dbt.DbtModelsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner dbt.DbtModelsListResponse
		svc := client.NewDbtModelsList().ProjectId(projectId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return dbt.DbtModelsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}

func flattenDbtModels(response dbt.DbtModelsListResponse) []interface{} {
	result := make([]interface{}, 0)

	for _, model := range response.Data.Items {
		modelMap := make(map[string]interface{})
		modelMap["id"] = model.ID
		modelMap["model_name"] = model.ModelName
		modelMap["scheduled"] = model.Scheduled
		result = append(result, modelMap)
	}

	return result
}
