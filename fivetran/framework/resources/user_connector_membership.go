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
            svc := r.GetClient().NewUserConnectionMembershipCreate()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectionId(connectorId)
            svc.Role(connectorElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if userConnectorResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create User Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, userConnectorResponse.Code, userConnectorResponse.Message),
                )

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, savedConnectors, data.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
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
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipDelete().UserId(plan.UserId.ValueString()).ConnectionId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
            }
			deletedConnectors = append(deletedConnectors, stateKey)
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipUpdate().UserId(plan.UserId.ValueString()).ConnectionId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnectors, plan.UserId.ValueString(), stateConnectorsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)
                return
            }
			modifiedConnectors = append(modifiedConnectors, stateKey)
        }
    }

	createdConnectors := make([]string, 0)
    for planKey, planValue := range planConnectorsMap {
        _, exists := stateConnectorsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipCreate().UserId(plan.UserId.ValueString()).ConnectionId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connector Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnectors, plan.UserId.ValueString(), stateConnectorsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnectors, plan.UserId.ValueString(), stateConnectorsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, createdConnectors, plan.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
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
            svc := r.GetClient().NewUserConnectionMembershipDelete()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectionId(connectorId)

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete User Connector Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnectors, data.UserId.ValueString(), stateConnectorsMap);
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
            }
			deletedConnectors = append(deletedConnectors, connectorId)
        }
    }
}


func (r *userConnectorMembership) RevertDeleted(ctx context.Context, toRevert []string, userId string, stateConnectorsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipCreate()
		svc.UserId(userId)
		svc.ConnectionId(connectorId)
		svc.Role(stateConnectorsMap[connectorId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Delete action reverted for connectors: %v", reverted),
	fmt.Sprintf("Delete for revert action failed for connectors: %v", failed)
}

func (r *userConnectorMembership) RevertModified(ctx context.Context, toRevert []string, userId string, stateConnectorsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipUpdate()
		svc.UserId(userId)
		svc.ConnectionId(connectorId)
		svc.Role(stateConnectorsMap[connectorId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Update action reverted for connectors: %v", reverted),
	fmt.Sprintf("Update for revert action failed for connectors: %v", failed)
}

func (r *userConnectorMembership) RevertCreated(ctx context.Context, toRevert []string, userId string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectorId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipDelete()
		svc.UserId(userId)
		svc.ConnectionId(connectorId)
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Create action reverted for connectors: %v", reverted),
	fmt.Sprintf("Create for revert action failed for connectors: %v", failed)
}
