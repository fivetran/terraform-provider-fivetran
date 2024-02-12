package resources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TeamConnectorMembership() resource.Resource {
    return &teamConnectorMembership{}
}

type teamConnectorMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &teamConnectorMembership{}
var _ resource.ResourceWithImportState = &teamConnectorMembership{}

func (r *teamConnectorMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_team_connector_membership"
}

func (r *teamConnectorMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.TeamConnectorMembershipResource()
}

func (r *teamConnectorMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamConnectorMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamConnectorMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    for _, connector := range data.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewTeamConnectorMembershipCreate()
            svc.TeamId(data.TeamId.ValueString())
            svc.ConnectorId(connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString())
            svc.Role(connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if teamConnectorResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Team Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, teamConnectorResponse.Code, teamConnectorResponse.Message),
                )

                return
            }
        }
    }

    teamConnectorResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamConnectorResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamConnectorMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamConnectorMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    teamConnectorResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamConnectorResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamConnectorMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.TeamConnectorMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planConnectorsMap := make(map[string]string)
    for _, connector := range plan.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
            planConnectorsMap[connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString()] = connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    stateConnectorsMap := make(map[string]string)
    for _, connector := range state.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
            stateConnectorsMap[connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString()] = connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
    for stateKey, stateValue := range stateConnectorsMap {
        role, found := planConnectorsMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewTeamConnectorMembershipDelete().TeamId(plan.TeamId.ValueString()).ConnectorId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewTeamConnectorMembershipModify().TeamId(plan.TeamId.ValueString()).ConnectorId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    for planKey, planValue := range planConnectorsMap {
        _, exists := stateConnectorsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewTeamConnectorMembershipCreate().TeamId(plan.TeamId.ValueString()).ConnectorId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    teamConnectorResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamConnectorResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, teamConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamConnectorMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamConnectorMemberships

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    for _, connector := range data.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewTeamConnectorMembershipDelete()
            svc.TeamId(data.TeamId.ValueString())
            svc.ConnectorId(connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString())

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Team Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

                return
            }
        }
    }
}