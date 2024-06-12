package model

import (
	"github.com/fivetran/go-fivetran/proxy"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProxyAgentDatasourceModel struct {
    Id               types.String `tfsdk:"id"`
    AccountId        types.String `tfsdk:"account_id"`
    RegistredAt      types.String `tfsdk:"registred_at"`
    GroupRegion      types.String `tfsdk:"group_region"`
    AuthToken        types.String `tfsdk:"token"`
    Salt             types.String `tfsdk:"salt"`
    CreatedBy        types.String `tfsdk:"created_by"`
    DisplayName      types.String `tfsdk:"display_name"`
}

var _ proxyAgentModel = &ProxyAgentDatasourceModel{}

func (d *ProxyAgentDatasourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetAccountId(value string) {
	d.AccountId = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetRegistredAt(value string) {
	d.RegistredAt = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetGroupRegion(value string) {
	d.GroupRegion = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetAuthToken(value string) {
	d.AuthToken = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetSalt(value string) {
	d.Salt = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetCreatedBy(value string) {
	d.CreatedBy = types.StringValue(value)
}
func (d *ProxyAgentDatasourceModel) SetDisplayName(value string) {
	d.DisplayName = types.StringValue(value)
}

func (d *ProxyAgentDatasourceModel) ReadFromResponse(resp proxy.ProxyDetailsResponse) {
	var model proxyAgentModel = d
	readProxyAgentFromResponse(model, resp)
}
