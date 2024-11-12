package model

import (
    "github.com/fivetran/go-fivetran/hybrid_deployment_agent"
)

type hybridDeploymentAgentModel interface {
    SetId(string)
    SetDisplayName(string)
    SetGroupId(string)
    SetRegisteredAt(string)
    SetConfigJson(string)
    SetAuthJson(string)
    SetDockerComposeYaml(string)
    SetToken(string)
}

func readHybridDeploymentAgentFromResponse(d hybridDeploymentAgentModel, resp hybriddeploymentagent.HybridDeploymentAgentDetailsResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
}

func readHybridDeploymentAgentFromCreateResponse(d hybridDeploymentAgentModel, resp hybriddeploymentagent.HybridDeploymentAgentCreateResponse) {
    d.SetId(resp.Data.Id)
    d.SetDisplayName(resp.Data.DisplayName)
    d.SetGroupId(resp.Data.GroupId)
    d.SetRegisteredAt(resp.Data.RegisteredAt)
    d.SetToken(resp.Data.Token)
    d.SetConfigJson(resp.Data.Files.ConfigJson)
    d.SetAuthJson(resp.Data.Files.AuthJson)
    d.SetDockerComposeYaml(resp.Data.Files.DockerComposeYaml)
}