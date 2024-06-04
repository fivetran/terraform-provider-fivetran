package model

import (
    "context"

    "github.com/fivetran/go-fivetran/proxy"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type Proxy struct {
    Id               types.String `tfsdk:"id"`
    AccountId        types.String `tfsdk:"account_id"`
    RegistredAt      types.String `tfsdk:"registred_at"`
    GroupRegion      types.String `tfsdk:"group_region"`
    AuthToken        types.String `tfsdk:"token"`
    Salt             types.String `tfsdk:"salt"`
    CreatedBy        types.String `tfsdk:"created_by"`
    DisplayName      types.String `tfsdk:"display_name"`
    ProxyServerUri   types.String `tfsdk:"proxy_server_uri"`
}

func (d *Proxy) ReadFromResponse(ctx context.Context, resp proxy.ProxyDetailsResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.AccountId = types.StringValue(resp.Data.AccountId)
    d.RegistredAt = types.StringValue(resp.Data.RegistredAt)
    d.GroupRegion = types.StringValue(resp.Data.Region)
    d.AuthToken = types.StringValue(resp.Data.Token)
    d.Salt = types.StringValue(resp.Data.Salt)
    d.CreatedBy = types.StringValue(resp.Data.CreatedBy)
    d.DisplayName = types.StringValue(resp.Data.DisplayName)
}

func (d *Proxy) ReadFromCreateResponse(ctx context.Context, resp proxy.ProxyCreateResponse) {
    d.Id = types.StringValue(resp.Data.AgentId)
    d.AuthToken = types.StringValue(resp.Data.AuthToken)
    d.ProxyServerUri = types.StringValue(resp.Data.ProxyServerUri)
}