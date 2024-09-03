package model

import (
	"github.com/fivetran/go-fivetran/proxy"
)

type proxyAgentModel interface {
	SetId(string)
	SetRegisteredAt(string)
	SetGroupRegion(string)
	SetAuthToken(string)
	SetSalt(string)
	SetCreatedBy(string)
	SetDisplayName(string)
}

func readProxyAgentFromResponse(d proxyAgentModel, resp proxy.ProxyDetailsResponse) {
	d.SetId(resp.Data.Id)
	d.SetRegisteredAt(resp.Data.RegisteredAt)
	d.SetGroupRegion(resp.Data.Region)
	d.SetAuthToken(resp.Data.Token)
	d.SetSalt(resp.Data.Salt)
	d.SetCreatedBy(resp.Data.CreatedBy)
	d.SetDisplayName(resp.Data.DisplayName)
}
