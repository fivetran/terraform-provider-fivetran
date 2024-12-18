package model

import (
    "github.com/fivetran/go-fivetran/hybrid_deployment_agent"
)

type localProcessingAgentModel interface {
    SetId(string)
    SetDisplayName(string)
    SetGroupId(string)
    SetRegisteredAt(string)
    SetConfigJson(string)
    SetAuthJson(string)
    SetDockerComposeYaml(string)
    SetUsage([]hybriddeploymentagent.HybridDeploymentAgentUsageDetails)
}

func readLocalProcessingAgentFromResponse(d localProcessingAgentModel, resp hybriddeploymentagent.HybridDeploymentAgentDetailsResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
    d.SetUsage(resp.Data.Usage)
}

func readLocalProcessingAgentFromCreateResponse(d localProcessingAgentModel, resp hybriddeploymentagent.HybridDeploymentAgentCreateResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
    d.SetConfigJson(resp.Data.Files.ConfigJson)
    d.SetAuthJson(resp.Data.Files.AuthJson)
    d.SetDockerComposeYaml(resp.Data.Files.DockerComposeYaml)
    d.SetUsage(nil)
}