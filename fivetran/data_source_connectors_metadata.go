package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConnectorsMetadata() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorsMetadataRead,
		Schema: map[string]*schema.Schema{
			"sources": dataSourceConnectorsMetadataSchemaSources(),
		},
	}
}

func dataSourceConnectorsMetadataSchemaSources() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeSet,
		// Uncomment Optional:true, before re-generating docs
		//Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the connector within the Fivetran system",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
				"icon_url": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
				"link_to_docs": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
				"link_to_erd": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "",
				},
			},
		},
	}
}

func dataSourceConnectorsMetadataRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceConnectorsMetadataGetMetadata(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("sources", dataSourceConnectorsMetadataFlattenMetadata(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID
	d.SetId("0")

	return diags
}

// dataSourceConnectorsMetadataFlattenMetadata receives a *fivetran.ConnectorsSourceMetadataResponse and returns a []interface{}
// containing the data type accepted by the "sources" set.
func dataSourceConnectorsMetadataFlattenMetadata(resp *fivetran.ConnectorsSourceMetadataResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	sources := make([]interface{}, len(resp.Data.Items), len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		source := make(map[string]interface{})
		source["id"] = v.ID
		source["name"] = v.Name
		source["type"] = v.Type
		source["description"] = v.Description
		source["icon_url"] = v.IconURL
		source["link_to_docs"] = v.LinkToDocs
		source["link_to_erd"] = v.LinkToErd

		sources[i] = source
	}

	return sources
}

// dataSourceConnectorsMetadataGetMetadata gets the connectors source metadata. It handles limits and cursors.
func dataSourceConnectorsMetadataGetMetadata(client *fivetran.Client, ctx context.Context) (fivetran.ConnectorsSourceMetadataResponse, error) {
	var resp fivetran.ConnectorsSourceMetadataResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.ConnectorsSourceMetadataResponse
		svc := client.NewConnectorsSourceMetadata()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.ConnectorsSourceMetadataResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
