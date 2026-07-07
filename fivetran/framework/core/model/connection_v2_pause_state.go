package model

import (
	"github.com/fivetran/go-fivetran/connections"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectionV2PauseStateResourceModel struct {
	Id           types.String `tfsdk:"id"`
	ConnectionId types.String `tfsdk:"connection_id"`
	Paused       types.Bool   `tfsdk:"paused"`
}

func (d *ConnectionV2PauseStateResourceModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.ConnectionId = types.StringValue(resp.Data.ID)
	d.Paused = boolPointerValue(resp.Data.Paused)
}
