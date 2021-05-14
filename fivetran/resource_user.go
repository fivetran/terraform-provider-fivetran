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
			// "role": &schema.Schema{ // commented until the fix https://fivetran.height.app/T-95317 / https://fivetran.height.app/T-39355 is implemented
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
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*fivetran.Client)
	s := c.NewUserInviteService()

	email := d.Get("email").(string)
	givenName := d.Get("given_name").(string)
	familyName := d.Get("family_name").(string)
	picture := d.Get("picture").(string)
	phone := d.Get("phone").(string)
	role := "ReadOnly" // hardcoded until the fix https://fivetran.height.app/T-95317 / https://fivetran.height.app/T-39355 is implemented

	s.Email(email)
	s.GivenName(givenName)
	s.FamilyName(familyName)
	s.Role(role)

	if picture != "" {
		s.Picture(picture)
	}

	if phone != "" {
		s.Phone(phone)
	}

	user, err := s.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "create error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, user.Code, user.Message),
		})
		return diags
	}

	d.SetId(user.Data.ID)

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*fivetran.Client)
	s := c.NewUserDetailsService()

	id := d.Get("id").(string)

	s.UserId(id)

	user, err := s.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "read error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, user.Code, user.Message),
		})
		return diags
	}

	kvmap := make(map[string]interface{})
	kvmap["id"] = user.Data.ID
	kvmap["email"] = user.Data.Email
	kvmap["given_name"] = user.Data.GivenName
	kvmap["family_name"] = user.Data.FamilyName
	kvmap["verified"] = user.Data.Verified
	kvmap["invited"] = user.Data.Invited
	kvmap["picture"] = user.Data.Picture
	kvmap["phone"] = user.Data.Phone
	kvmap["logged_in_at"] = user.Data.LoggedInAt.String()
	kvmap["created_at"] = user.Data.CreatedAt.String()

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
	c := m.(*fivetran.Client)
	s := c.NewUserModifyService()

	id := d.Get("id").(string)

	s.UserId(id)

	if d.HasChange("email") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "update error",
			Detail:   "field email can't be updated",
		})
		return diags
	}

	if d.HasChange("given_name") {
		givenName := d.Get("given_name").(string)
		s.GivenName(givenName)
		change = true
	}

	if d.HasChange("family_name") {
		familyName := d.Get("family_name").(string)
		s.FamilyName(familyName)
		change = true
	}

	if d.HasChange("picture") {
		picture := d.Get("picture").(string)
		s.Picture(picture)
		change = true
	}

	if d.HasChange("phone") {
		phone, ok := d.GetOk("phone")
		if ok {
			s.Phone(phone.(string))
			change = true
		}
	}

	if change {
		user, err := s.Do(ctx)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "update error",
				Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, user.Code, user.Message),
			})
			return diags
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*fivetran.Client)
	s := c.NewUserDeleteService()

	id := d.Get("id").(string)

	s.UserId(id)

	user, err := s.Do(ctx)
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
