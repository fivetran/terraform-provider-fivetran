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

func TeamConnectionMembership() resource.Resource {
	return &teamConnectionMembership{}
}

type teamConnectionMembership struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &teamConnectionMembership{}
var _ resource.ResourceWithImportState = &teamConnectionMembership{}

func (r *teamConnectionMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_connection_membership"
}

func (r *teamConnectionMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamConnectionMembershipResource()
}

func (r *teamConnectionMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamConnectionMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamConnectionMemberships

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	savedConnections := make([]string, 0, len(data.Connection.Elements()))
	for _, connection := range data.Connection.Elements() {
		if connectionElement, ok := connection.(basetypes.ObjectValue); ok {
			connectionId := connectionElement.Attributes()["connection_id"].(basetypes.StringValue).ValueString()
			svc := r.GetClient().NewTeamConnectionMembershipCreate()
			svc.TeamId(data.TeamId.ValueString())
			svc.ConnectionId(connectionId)
			svc.Role(connectionElement.Attributes()["role"].(basetypes.StringValue).ValueString())
			if createResponse, err := svc.Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Create Team Connector Memberships Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
				)

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, savedConnections, data.TeamId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
				return
			}
			savedConnections = append(savedConnections, connectionId)
		}
	}

	createResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Team Connection Memberships Resource.",
			fmt.Sprintf("%v; code: %v", err, createResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamConnectionMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamConnectionMemberships
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	readResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Team Connection Memberships Resource.",
			fmt.Sprintf("%v; code: %v", err, readResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, readResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamConnectionMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.TeamConnectionMembership

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
			if updateResponse, err := r.GetClient().NewTeamConnectionMembershipDelete().TeamId(plan.TeamId.ValueString()).ConnectionId(stateKey).Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Team Connection Membership Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
				)

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.TeamId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				return
			}
			deletedConnections = append(deletedConnections, stateKey)
		} else if role != stateValue {
			if updateResponse, err := r.GetClient().NewTeamConnectionMembershipModify().TeamId(plan.TeamId.ValueString()).ConnectionId(stateKey).Role(role).Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Team Connection Membership Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
				)

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.TeamId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnections, plan.TeamId.ValueString(), stateConnectionsMap)		
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
			if updateResponse, err := r.GetClient().NewTeamConnectionMembershipCreate().TeamId(plan.TeamId.ValueString()).ConnectionId(planKey).Role(planValue).Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Update Team Connection Membership Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
				)
				
				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, plan.TeamId.ValueString(), stateConnectionsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedConnections, plan.TeamId.ValueString(), stateConnectionsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, createdConnections, plan.TeamId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
				return
			}
			createdConnections = append(createdConnections, planKey)
		}
	}

	teamConnectionResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.TeamId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Team Connection Memberships Resource.",
			fmt.Sprintf("%v; code: %v", err, teamConnectionResponse.Code),
		)

		return
	}

	plan.ReadFromResponse(ctx, teamConnectionResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamConnectionMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data, state model.TeamConnectionMemberships

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
			svc := r.GetClient().NewTeamConnectionMembershipDelete()
			svc.TeamId(data.TeamId.ValueString())
			svc.ConnectionId(connectionId)

			if deleteResponse, err := svc.Do(ctx); err != nil {
				resp.Diagnostics.AddError(
					"Unable to Delete Team Connection Memberships Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
				)

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedConnections, data.TeamId.ValueString(), stateConnectionsMap);
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
				return
			}
			deletedConnections = append(deletedConnections, connectionId)
		}
	}
}

func (r *teamConnectionMembership) RevertDeleted(ctx context.Context, toRevert []string, teamId string, stateConnectionsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewTeamConnectionMembershipCreate()
		svc.TeamId(teamId)
		svc.ConnectionId(connectionId)
		svc.Role(stateConnectionsMap[connectionId])

		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, connectionId)
		} else {
			reverted = append(reverted, connectionId)
		} 
	}
	return fmt.Sprintf("Delete action reverted for connections: %v", reverted),
	fmt.Sprintf("Delete for revert action failed for connections: %v", failed)
}

func (r *teamConnectionMembership) RevertModified(ctx context.Context, toRevert []string, teamId string, stateConnectionsMap map[string]string)  (string, string)  {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewTeamConnectionMembershipModify()
		svc.TeamId(teamId)
		svc.ConnectionId(connectionId)
		svc.Role(stateConnectionsMap[connectionId])
		
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, connectionId)
		} else {
			reverted = append(reverted, connectionId)
		} 
	}
	return fmt.Sprintf("Modify action reverted for connections: %v", reverted),
	fmt.Sprintf("Modify for revert action failed for connections: %v", failed)
}

func (r *teamConnectionMembership) RevertCreated(ctx context.Context, toRevert []string, teamId string)  (string, string)  {
	reverted := []string{}
	failed := []string{}
	for _, connectionId := range toRevert {
		svc := r.GetClient().NewTeamConnectionMembershipDelete()
		svc.TeamId(teamId)
		svc.ConnectionId(connectorId)
		
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, connectionId)
		} else {
			reverted = append(reverted, connectionId)
		} 
		reverted = append(reverted, connectionId)
	}
	return fmt.Sprintf("Created action reverted for connections: %v", reverted),
	fmt.Sprintf("Created for revert action failed for connections: %v", failed)
}
