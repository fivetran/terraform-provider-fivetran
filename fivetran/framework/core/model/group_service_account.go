package model

import (
	"github.com/fivetran/go-fivetran/groups"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupServiceAccount struct {
	ID             types.String `tfsdk:"id"`
	ServiceAccount types.String `tfsdk:"service_account"`
}

func (d *GroupServiceAccount) ReadFromResponse(resp groups.GroupServiceAccountResponse) {
	d.ServiceAccount = types.StringValue(resp.Data.ServiceAccount)
}
