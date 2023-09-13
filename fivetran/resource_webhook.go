package fivetran

import (
	"context"
	"fmt"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The webhook ID",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The webhook type (group, account)",
			},
			"group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The group ID",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Your webhooks URL endpoint for your application",
			},
			"events": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The array of event types",
			},
			"active": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Boolean, if set to true, webhooks are immediately sent in response to events",
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
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
			"run_tests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the setup tests should be run",
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	
	if d.Get("type").(string) == "account" {
		svcAcc := client.NewWebhookAccountCreate()
		svcAcc.Url(d.Get("url").(string))
		svcAcc.Secret(d.Get("secret").(string))

		if v, ok := d.GetOk("active"); ok {
			svcAcc.Active(v.(bool))
		}

		if v, ok := d.GetOk("events"); ok {
			svcAcc.Events(xInterfaceStrXStr(v.(*schema.Set).List()))
		}

		resp, err := svcAcc.Do(ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v", err, resp.Code))
		}

		d.SetId(resp.Data.Id)
	} else if d.Get("type").(string) == "group" && d.Get("group_id").(string) != "" {
		svcGroup := client.NewWebhookGroupCreate().GroupId(d.Get("group_id").(string))
		svcGroup.Url(d.Get("url").(string))
		svcGroup.Secret(d.Get("secret").(string))

		if v, ok := d.GetOk("active"); ok {
			svcGroup.Active(v.(bool))
		}

		if v, ok := d.GetOk("events"); ok {
			svcGroup.Events(xInterfaceStrXStr(v.(*schema.Set).List()))
		}

		resp, err := svcGroup.Do(ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v", err, resp.Code))
		}

		d.SetId(resp.Data.Id)
	} else {
		return newDiagAppend(diags, diag.Error, "Incorrect webhook type", "Available values for type field is account or group. If you specify type = group, you need to set group_id")
	}

	resourceWebhookRead(ctx, d, m)

	return diags
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewWebhookDetails()

	svc.WebhookId(d.Get("id").(string)).Do(ctx)

	resp, err := svc.Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v", err, resp.Code))
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

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	svc := client.NewWebhookModify()

	svc.WebhookId(d.Get("id").(string))

	hasChanges := false

	if d.HasChange("url") {
		svc.Url(d.Get("url").(string))
		hasChanges = true
	}

	if d.HasChange("secret") {
		svc.Secret(d.Get("secret").(string))
		hasChanges = true
	}

	if v, ok := d.GetOk("active"); ok {
		svc.Active(v.(bool))
		hasChanges = true
	}

	if d.HasChange("events") {
		vars := make([]string, 0)
		for _, varValue := range d.Get("events").(*schema.Set).List() {
			vars = append(vars, varValue.(string))
		}
		svc.Events(vars)
		hasChanges = true
	}

	if v, ok := d.GetOk("run_tests"); ok {
		svc.RunTests(v.(bool))
	}

	if hasChanges {
		resp, err := svc.Do(ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v", err, resp.Code, resp.Message))
		}		
	}
	
	if v, ok := d.GetOk("run_tests"); ok && v.(bool) && d.HasChange("run_tests") {
		testsSvc := m.(*fivetran.Client).NewWebhookTestsRunner().WebhookId(d.Get("id").(string))
		for _, varValue := range d.Get("events").(*schema.Set).List() {
			testsSvc.Event(varValue.(string))
			resp, err := testsSvc.Do(ctx)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code))
			}
		}
	}

	return resourceWebhookRead(ctx, d, m)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewWebhookDelete()

	resp, err := svc.WebhookId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
