package model

import (
	"github.com/fivetran/go-fivetran/hybrid_deployment_agent"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HybridDeploymentAgentResourceModel struct {
    Id                  	types.String `tfsdk:"id"`
    DisplayName         	types.String `tfsdk:"display_name"`
    GroupId             	types.String `tfsdk:"group_id"`
    RegisteredAt        	types.String `tfsdk:"registered_at"`
    ConfigJson          	types.String `tfsdk:"config_json"`
    AuthJson            	types.String `tfsdk:"auth_json"`
    DockerComposeYaml   	types.String `tfsdk:"docker_compose_yaml"`
    AuthenticationCounter   types.Int64  `tfsdk:"authentication_counter"`
}

var _ hybridDeploymentAgentModel = &HybridDeploymentAgentResourceModel{}

func (d *HybridDeploymentAgentResourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetGroupId(value string) {
	d.GroupId = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetDisplayName(value string) {
	d.DisplayName = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetRegisteredAt(value string) {
	d.RegisteredAt = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetConfigJson(value string) {
	d.ConfigJson = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetAuthJson(value string) {
	d.AuthJson = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) SetDockerComposeYaml(value string) {
	d.DockerComposeYaml = types.StringValue(value)
}
func (d *HybridDeploymentAgentResourceModel) ReadFromCreateResponse(resp hybriddeploymentagent.HybridDeploymentAgentCreateResponse) {
	var model hybridDeploymentAgentModel = d
	readHybridDeploymentAgentFromCreateResponse(model, resp)
	d.AuthenticationCounter = types.Int64Value(d.AuthenticationCounter.ValueInt64() + 1)
}

func (d *HybridDeploymentAgentResourceModel) ReadFromResponse(resp hybriddeploymentagent.HybridDeploymentAgentDetailsResponse) {
	var model hybridDeploymentAgentModel = d
	readHybridDeploymentAgentFromResponse(model, resp)
}