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

func UserConnectorMembership() resource.Resource {
    return &userConnectorMembership{}
}

type userConnectorMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &userConnectorMembership{}
var _ resource.ResourceWithImportState = &userConnectorMembership{}

func (r *userConnectorMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user_connector_membership"
}

func (r *userConnectorMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.UserConnectorMembershipResource()
}

func (r *userConnectorMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userConnectorMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserConnectorMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    
	savedConnectors := make([]string, 0, len(data.Connector.Elements()))
	for _, connector := range data.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
			connectorId := connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewUserConnectorMembershipCreate()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectorId(connectorId)
            svc.Role(connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if userConnectorResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create User Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, userConnectorResponse.Code, userConnectorResponse.Message),
                )

				r.RevertCreated(ctx, savedConnectors, data.UserId.ValueString())
                return
            }

			savedConnectors = append(savedConnectors, connectorId)
        }
    }

    userConnectorResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectorResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userConnectorMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserConnectorMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    userConnectorResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read User Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectorResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userConnectorMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.UserConnectorMemberships

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
	deletedConnectors := make([]string, 0)
	modifiedConnectors := make([]string, 0)
    for stateKey, stateValue := range stateConnectorsMap {
        role, found := planConnectorsMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewUserConnectorMembershipDelete().UserId(plan.UserId.ValueString()).ConnectorId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)
                return
            }
			deletedConnectors = append(deletedConnectors, stateKey)
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewUserConnectorMembershipModify().UserId(plan.UserId.ValueString()).ConnectorId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)
				r.RevertModified(ctx, modifiedConnectors, plan.UserId.ValueString(), stateConnectorsMap)
                return
            }
			modifiedConnectors = append(modifiedConnectors, stateKey)
        }
    }

	createdConnectors := make([]string, 0)
    for planKey, planValue := range planConnectorsMap {
        _, exists := stateConnectorsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewUserConnectorMembershipCreate().UserId(plan.UserId.ValueString()).ConnectorId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)
				r.RevertModified(ctx, modifiedConnectors, plan.UserId.ValueString(), stateConnectorsMap)
				r.RevertCreated(ctx, createdConnectors, plan.UserId.ValueString())
                return
            }
			createdConnectors = append(createdConnectors, planKey)
        }
    }

    userConnectorResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Connector Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectorResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, userConnectorResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userConnectorMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data, state model.UserConnectorMemberships

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    stateConnectorsMap := make(map[string]string)
    for _, connector := range state.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
            stateConnectorsMap[connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString()] = connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

	deletedConnectors := make([]string, 0)
    for _, connector := range data.Connector.Elements() {
        if connectorElement, ok := connector.(basetypes.ObjectValue); ok {
			connectorId := connectorElement.Attributes()["connector_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewUserConnectorMembershipDelete()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectorId(connectorId)

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete User Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

				r.RevertDeleted(ctx, deletedConnectors, data.UserId.ValueString(), stateConnectorsMap)
                return
            }
			deletedConnectors = append(deletedConnectors, connectorId)
        }
    }
}


func (r *userConnectorMembership) RevertDeleted(ctx context.Context, toRevert []string, userId string, stateConnectorsMap map[string]string) {
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectorMembershipCreate()
		svc.UserId(userId)
		svc.ConnectorId(connectorId)
		svc.Role(stateConnectorsMap[connectorId])
		svc.Do(ctx)
	}
}

func (r *userConnectorMembership) RevertModified(ctx context.Context, toRevert []string, userId string, stateConnectorsMap map[string]string) {
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectorMembershipModify()
		svc.UserId(userId)
		svc.ConnectorId(connectorId)
		svc.Role(stateConnectorsMap[connectorId])
		svc.Do(ctx)
	}
}

func (r *userConnectorMembership) RevertCreated(ctx context.Context, toRevert []string, userId string) {
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectorMembershipDelete()
		svc.UserId(userId)
		svc.ConnectorId(connectorId)
		svc.Do(ctx)
	}
}
