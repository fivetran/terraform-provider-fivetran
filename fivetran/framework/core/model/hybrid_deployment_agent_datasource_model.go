package model

import (
	"github.com/fivetran/go-fivetran/hybrid_deployment_agent"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

type HybridDeploymentAgentDatasourceModel struct {
    Id                  types.String `tfsdk:"id"`
    DisplayName         types.String `tfsdk:"display_name"`
    GroupId             types.String `tfsdk:"group_id"`
    RegisteredAt        types.String `tfsdk:"registered_at"`
    Usage               types.Set    `tfsdk:"usage"`
}

var (
    elementType = map[string]attr.Type{
        "connection_id":    types.StringType,
        "schema":           types.StringType,
        "service":          types.StringType,
    }
)

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
func (d *HybridDeploymentAgentDatasourceModel) SetUsage(value []hybriddeploymentagent.HybridDeploymentAgentUsageDetails) {
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

func (d *HybridDeploymentAgentDatasourceModel) SetConfigJson(value string) {}
func (d *HybridDeploymentAgentDatasourceModel) SetAuthJson(value string) {}
func (d *HybridDeploymentAgentDatasourceModel) SetDockerComposeYaml(value string) {}

func (d *HybridDeploymentAgentDatasourceModel) ReadFromResponse(resp hybriddeploymentagent.HybridDeploymentAgentDetailsResponse) {
	var model hybridDeploymentAgentModel = d
	readHybridDeploymentAgentFromResponse(model, resp)
}