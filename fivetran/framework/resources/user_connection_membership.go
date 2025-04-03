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

func UserConnectionMembership() resource.Resource {
    return &userConnectionMembership{}
}

type userConnectionMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &userConnectionMembership{}
var _ resource.ResourceWithImportState = &userConnectionMembership{}

func (r *userConnectionMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user_connection_membership"
}

func (r *userConnectionMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.UserConnectionMembershipResource()
}

func (r *userConnectionMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userConnectionMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserConnectionMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    
	savedConnections := make([]string, 0, len(data.Connection.Elements()))
	for _, connection := range data.Connection.Elements() {
        if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
			connectionId := connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewUserConnectionMembershipCreate()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectionId(connectionId)
            svc.Role(connectionElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if userConnectionResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create User Connection Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, userConnectionResponse.Code, userConnectionResponse.Message),
                )

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, savedConnections, data.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
                return
            }

			savedConnections = append(savedConnections, connectionId)
        }
    }

    userConnectionResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Connection Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectionResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userConnectionResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userConnectionMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserConnectionMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    userConnectionResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read User Connection Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectionResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userConnectionResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userConnectionMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.UserConnectionMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planConnectionsMap := make(map[string]string)
    for _, connection := range plan.Connection.Elements() {
        if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
            planConnectionsMap[connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()] = connectionElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    stateConnectionsMap := make(map[string]string)
    for _, connection := range state.Connection.Elements() {
        if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
            stateConnectionsMap[connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()] = connectionElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
	deletedConnections := make([]string, 0)
	modifiedConnections := make([]string, 0)
    for stateKey, stateValue := range stateConnectionsMap {
        role, found := planConnectionsMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipDelete().UserId(plan.UserId.ValueString()).ConnectionId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connection Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.UserId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
            }
			deletedConnections = append(deletedConnections, stateKey)
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipModify().UserId(plan.UserId.ValueString()).ConnectionId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connection Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.UserId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnections, plan.UserId.ValueString(), stateConnectionsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)
                return
            }
			modifiedConnections = append(modifiedConnections, stateKey)
        }
    }

	createdConnections := make([]string, 0)
    for planKey, planValue := range planConnectionsMap {
        _, exists := stateConnectionsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewUserConnectionMembershipCreate().UserId(plan.UserId.ValueString()).ConnectionId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Connection Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.UserId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnections, plan.UserId.ValueString(), stateConnectionsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, createdConnections, plan.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
                return
            }
			createdConnections = append(createdConnections, planKey)
        }
    }

    userConnectionResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Connection Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userConnectionResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, userConnectionResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userConnectionMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data, state model.UserConnectionMemberships

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    stateConnectionsMap := make(map[string]string)
    for _, connection := range state.Connection.Elements() {
        if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
            stateConnectionsMap[connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()] = connectionElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

	deletedConnections := make([]string, 0)
    for _, connection := range data.Connection.Elements() {
        if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
			connectionId := connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewUserConnectionMembershipDelete()
            svc.UserId(data.UserId.ValueString())
            svc.ConnectionId(connectionId)

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete User Connection Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, data.UserId.ValueString(), stateConnectionsMap);
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
            }
			deletedConnections = append(deletedConnections, connectionId)
        }
    }
}


func (r *userConnectionMembership) RevertDeleted(ctx context.Context, toRevert []string, userId string, stateConnectionsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipCreate()
		svc.UserId(userId)
		svc.ConnectionId(connectionId)
		svc.Role(stateConnectionsMap[connectionId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Delete action reverted for connections: %v", reverted),
	fmt.Sprintf("Delete for revert action failed for connections: %v", failed)
}

func (r *userConnectionMembership) RevertModified(ctx context.Context, toRevert []string, userId string, stateConnectionsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipModify()
		svc.UserId(userId)
		svc.ConnectionId(connectionId)
		svc.Role(stateConnectionsMap[connectionId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Modify action reverted for connections: %v", reverted),
	fmt.Sprintf("Modify for revert action failed for connections: %v", failed)
}

func (r *userConnectionMembership) RevertCreated(ctx context.Context, toRevert []string, userId string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewUserConnectionMembershipDelete()
		svc.UserId(userId)
		svc.ConnectionId(connectionId)
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, userId)
		} else {
			reverted = append(reverted, userId)
		} 
	}
	return fmt.Sprintf("Create action reverted for connections: %v", reverted),
	fmt.Sprintf("Create for revert action failed for connections: %v", failed)
}
