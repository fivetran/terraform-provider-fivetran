package model

import (
	"github.com/fivetran/go-fivetran/users"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type User struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	GivenName  types.String `tfsdk:"given_name"`
	FamilyName types.String `tfsdk:"family_name"`
	Verified   types.Bool   `tfsdk:"verified"`
	Invited    types.Bool   `tfsdk:"invited"`
	Picture    types.String `tfsdk:"picture"`
	Phone      types.String `tfsdk:"phone"`
	Role       types.String `tfsdk:"role"`
	LoggedInAt types.String `tfsdk:"logged_in_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (d *User) ReadFromResponse(resp users.UserDetailsResponse) {
	d.ID = types.StringValue(resp.Data.ID)
	d.Email = types.StringValue(resp.Data.Email)
	d.FamilyName = types.StringValue(resp.Data.FamilyName)
	d.GivenName = types.StringValue(resp.Data.GivenName)

	d.Role = types.StringValue(resp.Data.Role)
	d.Verified = types.BoolValue(*resp.Data.Verified)
	d.Invited = types.BoolValue(*resp.Data.Invited)
	d.LoggedInAt = types.StringValue(resp.Data.LoggedInAt.String())
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt.String())

	if resp.Data.Phone == "" {
		d.Phone = types.StringNull()
	} else {
		d.Phone = types.StringValue(resp.Data.Phone)
	}

	if resp.Data.Picture == "" {
		d.Picture = types.StringNull()
	} else {
		d.Picture = types.StringValue(resp.Data.Picture)
	}
}
