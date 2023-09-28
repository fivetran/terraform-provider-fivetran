package fivetran

import (
    "context"
    "fmt"

    "github.com/fivetran/go-fivetran"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMetadataTables() *schema.Resource {
    return &schema.Resource{
        ReadContext: dataSourceMetadataTablesRead,
        Schema: map[string]*schema.Schema{
            "id": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "The unique identifier for the connector within the Fivetran system. Data-source will represent a set of tables of connector.",
            },
            "metadata_tables": {
                Type:     schema.TypeSet,
                Optional: true,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The unique table identifier",
                        },
                        "parent_id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The unique identifier of the schema associated with the table",
                        },
                        "name_in_source": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The table name in the source",
                        },
                        "name_in_destination": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The table name in the destination",
                        },
                    },
                },
            },
        },
    }
}

func dataSourceMetadataTablesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    id := d.Get("id").(string)

    resp, err := dataSourceMetadataTablesGet(client, ctx, id)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    if err := d.Set("metadata_tables", dataSourceMetadataTablesFlat(&resp)); err != nil {
        return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
    }

    // Enforces ID, there can't be two account-wide datasources
    d.SetId(id)

    return diags
}

// dataSourceMetadataTablesFlat receives a *fivetran.MetadataTablesListResponse and returns a []interface{}
// containing the data type accepted by the "metadata_tables" set.
func dataSourceMetadataTablesFlat(resp *fivetran.MetadataTablesListResponse) []interface{} {
    if resp.Data.Items == nil {
        return make([]interface{}, 0)
    }

    metadata_tables := make([]interface{}, len(resp.Data.Items))
    for i, v := range resp.Data.Items {
        metadata_table := make(map[string]interface{})
        metadata_table["id"] = v.Id
        metadata_table["parent_id"] = v.ParentId
        metadata_table["name_in_source"] = v.NameInSource
        metadata_table["name_in_destination"] = v.NameInDestination

        metadata_tables[i] = metadata_table
    }

    return metadata_tables
}

// dataSourceMetadataTablesGet gets the list of a metadata_tables. It handles limits and cursors.
func dataSourceMetadataTablesGet(client *fivetran.Client, ctx context.Context, id string) (fivetran.MetadataTablesListResponse, error) {
    var resp fivetran.MetadataTablesListResponse
    var respNextCursor string

    for {
        var err error
        var respInner fivetran.MetadataTablesListResponse
        svc := client.NewMetadataTablesList().ConnectorId(id)
        if respNextCursor == "" {
            respInner, err = svc.Limit(limit).Do(ctx)
        }
        if respNextCursor != "" {
            respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        if err != nil {
            return fivetran.MetadataTablesListResponse{}, err
        }

        resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

        if respInner.Data.NextCursor == "" {
            break
        }

        respNextCursor = respInner.Data.NextCursor
    }

    return resp, nil
}
