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
	SetHybridDeploymentAgentId(string)
	SetNetworkingMethod(string)
    SetPrivateLinkId(string)
	SetConfig(map[string]interface{}, bool)
}

func readFromResponse(d destinationModel, resp destinations.DestinationDetailsBase, config map[string]interface{}, isImporting bool) {
	d.SetId(resp.ID)
	d.SetGroupId(resp.GroupID)
	d.SetService(resp.Service)
	d.SetRegion(resp.Region)
	d.SetSetupStatus(resp.SetupStatus)
	d.SetTimeZonOffset(resp.TimeZoneOffset)
	d.SetDaylightSavingTimeEnabled(resp.DaylightSavingTimeEnabled)
	d.SetHybridDeploymentAgentId(resp.HybridDeploymentAgentId)
	d.SetNetworkingMethod(resp.NetworkingMethod)
	d.SetPrivateLinkId(resp.PrivateLinkId)
	d.SetConfig(config, isImporting)
}
