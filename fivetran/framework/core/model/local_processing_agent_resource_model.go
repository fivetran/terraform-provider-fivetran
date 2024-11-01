package model

import (
	localprocessingagent "github.com/fivetran/go-fivetran/hybrid_deployment_agent"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type LocalProcessingAgentResourceModel struct {
    Id                  	types.String `tfsdk:"id"`
    DisplayName         	types.String `tfsdk:"display_name"`
    GroupId             	types.String `tfsdk:"group_id"`
    RegisteredAt        	types.String `tfsdk:"registered_at"`
    ConfigJson          	types.String `tfsdk:"config_json"`
    AuthJson            	types.String `tfsdk:"auth_json"`
    DockerComposeYaml   	types.String `tfsdk:"docker_compose_yaml"`
    AuthenticationCounter   types.Int64  `tfsdk:"authentication_counter"`
    Usage               	types.Set    `tfsdk:"usage"`
}

var _ localProcessingAgentModel = &LocalProcessingAgentResourceModel{}

func (d *LocalProcessingAgentResourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetGroupId(value string) {
	d.GroupId = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetDisplayName(value string) {
	d.DisplayName = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetRegisteredAt(value string) {
	d.RegisteredAt = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetConfigJson(value string) {
	d.ConfigJson = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetAuthJson(value string) {
	d.AuthJson = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetDockerComposeYaml(value string) {
	d.DockerComposeYaml = types.StringValue(value)
}
func (d *LocalProcessingAgentResourceModel) SetUsage(value []localprocessingagent.HybridDeploymentAgentUsageDetails) {
    if value == nil {
        d.Usage = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    for _, v := range value {
        item := map[string]attr.Value{}
        item["connection_id"] = types.StringValue(v.ConnectionId)
        item["schema"] = types.StringValue(v.Schema)
        item["service"] = types.StringValue(v.Service)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
	}

    d.Usage, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}
func (d *LocalProcessingAgentResourceModel) ReadFromCreateResponse(resp localprocessingagent.HybridDeploymentAgentCreateResponse) {
	var model localProcessingAgentModel = d
	readLocalProcessingAgentFromCreateResponse(model, resp)
	d.AuthenticationCounter = types.Int64Value(d.AuthenticationCounter.ValueInt64() + 1)
}

func (d *LocalProcessingAgentResourceModel) ReadFromResponse(resp localprocessingagent.HybridDeploymentAgentDetailsResponse) {
	var model localProcessingAgentModel = d
	readLocalProcessingAgentFromResponse(model, resp)
}