package model

import (
	"github.com/fivetran/go-fivetran/destinations"
)

type destinationModel interface {
	SetId(string)
	SetGroupId(string)
	SetService(string)
	SetRegion(string)
	SetTimeZonOffset(string)
	SetSetupStatus(string)
	SetDaylightSavingTimeEnabled(bool)
    SetLocalProcessingAgentId(string)
    SetNetworkingMethod(string)
    SetPrivateLinkId(string)
	SetConfig(map[string]interface{})
}

func readFromResponse(d destinationModel, resp destinations.DestinationDetailsBase, config map[string]interface{}) {
	d.SetId(resp.ID)
	d.SetGroupId(resp.GroupID)
	d.SetService(resp.Service)
	d.SetRegion(resp.Region)
	d.SetSetupStatus(resp.SetupStatus)
	d.SetTimeZonOffset(resp.TimeZoneOffset)
	d.SetDaylightSavingTimeEnabled(resp.DaylightSavingTimeEnabled)
	d.SetLocalProcessingAgentId(resp.HybridDeploymentAgentId)
	d.SetNetworkingMethod(resp.NetworkingMethod)
	d.SetPrivateLinkId(resp.PrivateLinkId)
	d.SetConfig(config)
}
