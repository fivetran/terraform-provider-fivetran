package model

import (
	"github.com/fivetran/go-fivetran/proxy"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProxyAgentResourceModel struct {
	Id             			types.String `tfsdk:"id"`
	RegisteredAt   			types.String `tfsdk:"registred_at"`
	GroupRegion    			types.String `tfsdk:"group_region"`
	AuthToken      			types.String `tfsdk:"token"`
	CreatedBy      			types.String `tfsdk:"created_by"`
	DisplayName    			types.String `tfsdk:"display_name"`
	ClientCert 				types.String `tfsdk:"client_cert"`
	ClientPrivateKey 		types.String `tfsdk:"client_private_key"`
    RegenerationCounter   	types.Int64  `tfsdk:"regeneration_counter"`
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
	d.RegenerationCounter = types.Int64Value(d.RegenerationCounter.ValueInt64() + 1)
	if(d.AuthToken.IsUnknown()){
		d.AuthToken = types.StringNull()
	}
}

func (d *ProxyAgentResourceModel) ReadFromCreateResponse(resp proxy.ProxyCreateResponse) {
	d.Id = types.StringValue(resp.Data.AgentId)
	d.AuthToken = types.StringValue(resp.Data.AuthToken)
	d.ClientCert = types.StringValue(resp.Data.ClientCert)
	d.ClientPrivateKey = types.StringValue(resp.Data.ClientPrivateKey)
}
