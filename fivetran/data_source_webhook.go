package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhookRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The webhook ID",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The webhook type (group, account)",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The group ID",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Your webhooks URL endpoint for your application",
			},
			"events": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The array of event types",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean, if set to true, webhooks are immediately sent in response to events",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The secret string used for payload signing and masked in the response.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The webhook creation timestamp",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the user who created the webhook.",
			},
		},
	}
}

func dataSourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewWebhookDetails()

	resp, err := svc.WebhookId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.Id
	msi["type"] = resp.Data.Type
	msi["group_id"] = resp.Data.GroupId
	msi["url"] = resp.Data.Url
	msi["events"] = resp.Data.Events
	msi["active"] = resp.Data.Active
	msi["secret"] = resp.Data.Secret
	msi["created_at"] = resp.Data.CreatedAt
	msi["created_by"] = resp.Data.CreatedBy

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.Id)

	return diags
}
