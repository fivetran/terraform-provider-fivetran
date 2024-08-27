package model

import (
	"github.com/fivetran/go-fivetran/proxy"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProxyAgentResourceModel struct {
	Id             types.String `tfsdk:"id"`
	RegisteredAt   types.String `tfsdk:"registred_at"`
	GroupRegion    types.String `tfsdk:"group_region"`
	AuthToken      types.String `tfsdk:"token"`
	Salt           types.String `tfsdk:"salt"`
	CreatedBy      types.String `tfsdk:"created_by"`
	DisplayName    types.String `tfsdk:"display_name"`
	ProxyServerUri types.String `tfsdk:"proxy_server_uri"`
}

var _ proxyAgentModel = &ProxyAgentResourceModel{}

func (d *ProxyAgentResourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *ProxyAgentResourceModel) SetRegisteredAt(value string) {
	d.RegisteredAt = types.StringValue(value)
}
func (d *ProxyAgentResourceModel) SetGroupRegion(value string) {
	d.GroupRegion = types.StringValue(value)
}
func (d *ProxyAgentResourceModel) SetCreatedBy(value string) {
	d.CreatedBy = types.StringValue(value)
}
func (d *ProxyAgentResourceModel) SetDisplayName(value string) {
	d.DisplayName = types.StringValue(value)
}

func (d *ProxyAgentResourceModel) ReadFromResponse(resp proxy.ProxyDetailsResponse) {
	var model proxyAgentModel = d
	readProxyAgentFromResponse(model, resp)
	if(d.AuthToken.IsUnknown()){
		d.AuthToken = types.StringNull()
	}
	if(d.Salt.IsUnknown()){
		d.Salt = types.StringNull()
	}
}

func (d *ProxyAgentResourceModel) ReadFromCreateResponse(resp proxy.ProxyCreateResponse) {
	d.Id = types.StringValue(resp.Data.AgentId)
	d.AuthToken = types.StringValue(resp.Data.AuthToken)
	d.ProxyServerUri = types.StringValue(resp.Data.ProxyServerUri)
}
