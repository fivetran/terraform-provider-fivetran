package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWebhooks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhooksRead,
		Schema: map[string]*schema.Schema{
			"webhooks": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
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
				},
			},
		},
	}
}

func dataSourceWebhooksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceWebhooksGetWebhooks(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	if err := d.Set("webhooks", dataSourceWebhooksFlattenWebhooks(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceWebhooksFlattenWebhooks receives a *fivetran.WebhookListResponse and returns a []interface{}
// containing the data type accepted by the "webhooks" set.
func dataSourceWebhooksFlattenWebhooks(resp *fivetran.WebhookListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	webhooks := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		webhook := make(map[string]interface{})
		webhook["id"] = v.Id
		webhook["type"] = v.Type
		webhook["group_id"] = v.GroupId
		webhook["url"] = v.Url
		webhook["events"] = v.Events
		webhook["active"] = v.Active
		webhook["secret"] = v.Secret
		webhook["created_at"] = v.CreatedAt
		webhook["created_by"] = v.CreatedBy

		webhooks[i] = webhook
	}

	return webhooks
}

// dataSourceWebhooksGetWebhooks gets the webhooks list of a group. It handles limits and cursors.
func dataSourceWebhooksGetWebhooks(client *fivetran.Client, ctx context.Context) (fivetran.WebhookListResponse, error) {
	var resp fivetran.WebhookListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.WebhookListResponse
		svc := client.NewWebhookList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.WebhookListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
