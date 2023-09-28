package fivetran

import (
    "context"
    "fmt"

    "github.com/fivetran/go-fivetran"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMetadataColumns() *schema.Resource {
    return &schema.Resource{
        ReadContext: dataSourceMetadataColumnsRead,
        Schema: map[string]*schema.Schema{
            "id": {
                Type:        schema.TypeString,
                Required:    true,
                Description: "The unique identifier for the connector within the Fivetran system. Data-source will represent a set of columns of connector.",
            },
            "metadata_columns": {
                Type:     schema.TypeSet,
                Optional: true,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The unique column identifier",
                        },
                        "parent_id": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The unique identifier of the table associated with the column",
                        },
                        "name_in_source": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The column name in the source",
                        },
                        "name_in_destination": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The column name in the destination",
                        },
                        "type_in_source": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The column type in the source",
                        },
                        "type_in_destination": {
                            Type:        schema.TypeString,
                            Computed:    true,
                            Description: "The column type in the destination",
                        },
                        "is_primary_key": {
                            Type:        schema.TypeBool,
                            Computed:    true,
                            Description: "The boolean specifying whether the column is a primary key",
                        },
                        "is_foreign_key": {
                            Type:        schema.TypeBool,
                            Computed:    true,
                            Description: "The boolean specifying whether the column is a foreign key",
                        },
                    },
                },
            },
        },
    }
}

func dataSourceMetadataColumnsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    id := d.Get("id").(string)

    resp, err := dataSourceMetadataColumnsGet(client, ctx, id)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    if err := d.Set("metadata_columns", dataSourceMetadataColumnsFlat(&resp)); err != nil {
        return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
    }

    // Enforces ID, there can't be two account-wide datasources
    d.SetId(id)

    return diags
}

// dataSourceMetadataColumnsFlat receives a *fivetran.MetadataColumnsListResponse and returns a []interface{}
// containing the data type accepted by the "metadata_columns" set.
func dataSourceMetadataColumnsFlat(resp *fivetran.MetadataColumnsListResponse) []interface{} {
    if resp.Data.Items == nil {
        return make([]interface{}, 0)
    }

    metadata_columns := make([]interface{}, len(resp.Data.Items))
    for i, v := range resp.Data.Items {
        metadata_column := make(map[string]interface{})
        metadata_column["id"] = v.Id
        metadata_column["parent_id"] = v.ParentId
        metadata_column["name_in_source"] = v.NameInSource
        metadata_column["name_in_destination"] = v.NameInDestination
        metadata_column["type_in_source"] = v.TypeInSource
        metadata_column["type_in_destination"] = v.TypeInDestination
        metadata_column["is_primary_key"] = v.IsPrimaryKey
        metadata_column["is_foreign_key"] = v.IsForeignKey

        metadata_columns[i] = metadata_column
    }

    return metadata_columns
}

// dataSourceMetadataColumnsGet gets the list of a metadata_columns. It handles limits and cursors.
func dataSourceMetadataColumnsGet(client *fivetran.Client, ctx context.Context, id string) (fivetran.MetadataColumnsListResponse, error) {
    var resp fivetran.MetadataColumnsListResponse
    var respNextCursor string

    for {
        var err error
        var respInner fivetran.MetadataColumnsListResponse
        svc := client.NewMetadataColumnsList().ConnectorId(id)
        if respNextCursor == "" {
            respInner, err = svc.Limit(limit).Do(ctx)
        }
        if respNextCursor != "" {
            respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        if err != nil {
            return fivetran.MetadataColumnsListResponse{}, err
        }

        resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

        if respInner.Data.NextCursor == "" {
            break
        }

        respNextCursor = respInner.Data.NextCursor
    }

    return resp, nil
}
