package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,
		Schema:      getConnectorSchema(true, 0),
	}
}

func dataSourceConnectorRead(ctx context.Context, resourceData *schema.ResourceData, clientInterface interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := clientInterface.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(resourceData.Get("id").(string)).DoCustom(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	dataBucket := getConnectorRead(nil, resp, 0)

	for k, v := range dataBucket {
		if err := resourceData.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	resourceData.SetId(resp.Data.ID)

	return diags
}
