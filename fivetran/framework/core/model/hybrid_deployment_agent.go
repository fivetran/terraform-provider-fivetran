package model

import (
    "github.com/fivetran/go-fivetran/local_processing_agent"
)

type hybridDeploymentAgentModel interface {
    SetId(string)
    SetDisplayName(string)
    SetGroupId(string)
    SetRegisteredAt(string)
    SetConfigJson(string)
    SetAuthJson(string)
    SetDockerComposeYaml(string)
    SetUsage([]localprocessingagent.LocalProcessingAgentUsageDetails)
}

func readHybridDeploymentAgentFromResponse(d hybridDeploymentAgentModel, resp localprocessingagent.LocalProcessingAgentDetailsResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
    d.SetUsage(resp.Data.Usage)
}

func readHybridDeploymentAgentFromCreateResponse(d hybridDeploymentAgentModel, resp localprocessingagent.LocalProcessingAgentCreateResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
    d.SetConfigJson(resp.Data.Files.ConfigJson)
    d.SetAuthJson(resp.Data.Files.AuthJson)
    d.SetDockerComposeYaml(resp.Data.Files.DockerComposeYaml)
    d.SetUsage(nil)
}