package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMetadataSchemas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMetadataSchemasRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier for the connector within the Fivetran system. Data-source will represent a set of schemas of connector.",
			},
			"metadata_schemas": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set: func(v interface{}) int {
					return helpers.StringInt32Hash(v.(map[string]interface{})["id"].(string))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique schema identifier",
						},
						"name_in_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The schema name in the source",
						},
						"name_in_destination": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The schema name in the destination",
						},
					},
				},
			},
		},
	}
}

func dataSourceMetadataSchemasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	id := d.Get("id").(string)

	resp, err := dataSourceMetadataSchemasGet(client, ctx, id)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	if err := d.Set("metadata_schemas", dataSourceMetadataSchemasFlat(&resp)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId(id)

	return diags
}

// dataSourceMetadataSchemasFlat receives a *fivetran.MetadataSchemasListResponse and returns a []interface{}
// containing the data type accepted by the "metadata_schemas" set.
func dataSourceMetadataSchemasFlat(resp *fivetran.MetadataSchemasListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	metadata_schemas := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		metadata_schema := make(map[string]interface{})
		metadata_schema["id"] = v.Id
		metadata_schema["name_in_source"] = v.NameInSource
		metadata_schema["name_in_destination"] = v.NameInDestination

		metadata_schemas[i] = metadata_schema
	}

	return metadata_schemas
}

// dataSourceMetadataSchemasGet gets the list of a metadata_schemas. It handles limits and cursors.
func dataSourceMetadataSchemasGet(client *fivetran.Client, ctx context.Context, id string) (fivetran.MetadataSchemasListResponse, error) {
	var resp fivetran.MetadataSchemasListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.MetadataSchemasListResponse
		svc := client.NewMetadataSchemasList().ConnectorId(id)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.MetadataSchemasListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
