package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Computed: true},
			// The REST API doesn't provide a method to change the the user's email address.
			// That's why "ForceNew" is true.
			"email":       {Type: schema.TypeString, Required: true, ForceNew: true},
			"given_name":  {Type: schema.TypeString, Required: true},
			"family_name": {Type: schema.TypeString, Required: true},
			"verified":    {Type: schema.TypeBool, Computed: true},
			"invited":     {Type: schema.TypeBool, Computed: true},
			"picture":     {Type: schema.TypeString, Optional: true},
			"phone":       {Type: schema.TypeString, Optional: true},
			// "role":         {Type: schema.TypeString, Required: true}, // commented until T-109040 is fixed.
			"logged_in_at": {Type: schema.TypeString, Computed: true},
			"created_at":   {Type: schema.TypeString, Computed: true},
			"last_updated": {Type: schema.TypeString, Computed: true}, // internal
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserInvite()

	svc.Email(d.Get("email").(string))
	svc.GivenName(d.Get("given_name").(string))
	svc.FamilyName(d.Get("family_name").(string))
	if v, ok := d.GetOk("picture"); ok {
		svc.Picture(v.(string))
	}
	if v, ok := d.GetOk("phone"); ok {
		svc.Phone(v.(string))
	}
	// The REST API doesn't returns `role` when creating/inviting a new user. Because of that, `role`
	// is being enforced. This should change when T-109040 is fixed.
	svc.Role("ReadOnly")

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	svc.UserID(d.Get("id").(string)).Do(ctx)

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["email"] = resp.Data.Email
	msi["given_name"] = resp.Data.GivenName
	msi["family_name"] = resp.Data.FamilyName
	msi["verified"] = resp.Data.Verified
	msi["invited"] = resp.Data.Invited
	msi["picture"] = resp.Data.Picture
	msi["phone"] = resp.Data.Phone
	msi["logged_in_at"] = resp.Data.LoggedInAt.String()
	msi["created_at"] = resp.Data.CreatedAt.String()
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserModify()

	svc.UserID(d.Get("id").(string))

	if d.HasChange("given_name") {
		svc.GivenName(d.Get("given_name").(string))
	}
	if d.HasChange("family_name") {
		svc.FamilyName(d.Get("family_name").(string))
	}
	if d.HasChange("picture") {
		svc.Picture(d.Get("picture").(string))
	}
	if d.HasChange("phone") {
		svc.Phone(d.Get("phone").(string))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		// resourceUserRead here makes sure the state is updated after a NewUserModify error.
		diags = resourceUserRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDelete()

	resp, err := svc.UserID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
