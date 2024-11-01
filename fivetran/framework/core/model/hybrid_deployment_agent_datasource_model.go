package model

import (
	"github.com/fivetran/go-fivetran/hybrid_deployment_agent"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HybridDeploymentAgentDatasourceModel struct {
    Id                  types.String `tfsdk:"id"`
    DisplayName         types.String `tfsdk:"display_name"`
    GroupId             types.String `tfsdk:"group_id"`
    RegisteredAt        types.String `tfsdk:"registered_at"`
}

var _ hybridDeploymentAgentModel = &HybridDeploymentAgentDatasourceModel{}

func (d *HybridDeploymentAgentDatasourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *HybridDeploymentAgentDatasourceModel) SetGroupId(value string) {
	d.GroupId = types.StringValue(value)
}
func (d *HybridDeploymentAgentDatasourceModel) SetDisplayName(value string) {
	d.DisplayName = types.StringValue(value)
}
func (d *HybridDeploymentAgentDatasourceModel) SetRegisteredAt(value string) {
	d.RegisteredAt = types.StringValue(value)
}
func (d *HybridDeploymentAgentDatasourceModel) SetConfigJson(value string) {}
func (d *HybridDeploymentAgentDatasourceModel) SetAuthJson(value string) {}
func (d *HybridDeploymentAgentDatasourceModel) SetDockerComposeYaml(value string) {}
func (d *HybridDeploymentAgentDatasourceModel) SetToken(value string) {}

func (d *HybridDeploymentAgentDatasourceModel) ReadFromResponse(resp hybriddeploymentagent.HybridDeploymentAgentDetailsResponse) {
	var model hybridDeploymentAgentModel = d
	readHybridDeploymentAgentFromResponse(model, resp)
}