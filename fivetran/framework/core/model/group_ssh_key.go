package model

import (
	"github.com/fivetran/go-fivetran/groups"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupSshKey struct {
	ID        types.String `tfsdk:"id"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (d *GroupSshKey) ReadFromResponse(resp groups.GroupSshKeyResponse) {
	d.PublicKey = types.StringValue(resp.Data.PublicKey)
}
