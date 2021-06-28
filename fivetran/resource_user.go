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
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"given_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"family_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"verified": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"invited": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"picture": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			// "role": &schema.Schema{ // commented until https://fivetran.height.app/T-109040 is fixed.
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			"logged_in_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated": &schema.Schema{ // internal
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserInvite()

	email := d.Get("email").(string)
	svc.Email(email)

	givenName := d.Get("given_name").(string)
	svc.GivenName(givenName)

	familyName := d.Get("family_name").(string)
	svc.FamilyName(familyName)

	picture := d.Get("picture").(string)
	if picture != "" {
		svc.Picture(picture)
	}

	phone := d.Get("phone").(string)
	if phone != "" {
		svc.Phone(phone)
	}

	role := "ReadOnly" // hardcoded until https://fivetran.height.app/T-109040 is fixed.
	svc.Role(role)

	resp, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "create error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	d.SetId(resp.Data.ID)

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	id := d.Get("id").(string)
	svc.UserID(id)

	resp, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "read error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	kvmap := make(map[string]interface{})
	kvmap["id"] = resp.Data.ID
	kvmap["email"] = resp.Data.Email
	kvmap["given_name"] = resp.Data.GivenName
	kvmap["family_name"] = resp.Data.FamilyName
	kvmap["verified"] = resp.Data.Verified
	kvmap["invited"] = resp.Data.Invited
	kvmap["picture"] = resp.Data.Picture
	kvmap["phone"] = resp.Data.Phone
	kvmap["logged_in_at"] = resp.Data.LoggedInAt.String()
	kvmap["created_at"] = resp.Data.CreatedAt.String()

	if err := set(d, kvmap); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "set error",
			Detail:   fmt.Sprint(err),
		})
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var change bool
	client := m.(*fivetran.Client)
	svc := client.NewUserModify()

	id := d.Get("id").(string)

	svc.UserID(id)

	if d.HasChange("email") {
		diags = append(diags, diag.Diagnostic{ // should verify this or this is REST API role? IMO it's REST API...
			Severity: diag.Error,
			Summary:  "update error",
			Detail:   "field email can't be updated",
		})
		return diags
	}

	if d.HasChange("given_name") {
		givenName := d.Get("given_name").(string)
		svc.GivenName(givenName)
		change = true
	}

	if d.HasChange("family_name") {
		familyName := d.Get("family_name").(string)
		svc.FamilyName(familyName)
		change = true
	}

	if d.HasChange("picture") {
		picture := d.Get("picture").(string)
		svc.Picture(picture)
		change = true
	}

	if d.HasChange("phone") {
		phone, ok := d.GetOk("phone")
		if ok {
			svc.Phone(phone.(string))
			change = true
		}
	}

	if change {
		resp, err := svc.Do(ctx)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "update error",
				Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
			})
			return diags
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDelete()

	id := d.Get("id").(string)
	svc.UserID(id)

	user, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "delete error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, user.Code, user.Message),
		})
		return diags
	}

	d.SetId("")

	return diags
}
