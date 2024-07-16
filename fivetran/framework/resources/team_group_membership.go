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

func TeamGroupMembership() resource.Resource {
    return &teamGroupMembership{}
}

type teamGroupMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &teamGroupMembership{}
var _ resource.ResourceWithImportState = &teamGroupMembership{}

func (r *teamGroupMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_team_group_membership"
}

func (r *teamGroupMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.TeamGroupMembershipResource()
}

func (r *teamGroupMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamGroupMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamGroupMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	savedGroups := make([]string, 0, len(data.Group.Elements()))
    for _, group := range data.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
			groupId := groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewTeamGroupMembershipCreate()
            svc.TeamId(data.TeamId.ValueString())
            svc.GroupId(groupId)
            svc.Role(groupElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if teamGroupResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Team Group Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, teamGroupResponse.Code, teamGroupResponse.Message),
                )
				
				r.RevertCreated(ctx, savedGroups, data.TeamId.ValueString())
                return
            }
			savedGroups = append(savedGroups, groupId)
        }
    }

    teamGroupResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamGroupResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamGroupMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamGroupMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    teamGroupResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamGroupResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamGroupMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.TeamGroupMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planGroupsMap := make(map[string]string)
    for _, group := range plan.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
            planGroupsMap[groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()] = groupElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    stateGroupsMap := make(map[string]string)
    for _, group := range state.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
            stateGroupsMap[groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()] = groupElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
	deletedGroups := make([]string, 0)
	modifiedGroups := make([]string, 0)
    for stateKey, stateValue := range stateGroupsMap {
        role, found := planGroupsMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewTeamGroupMembershipDelete().TeamId(plan.TeamId.ValueString()).GroupId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedGroups, plan.TeamId.ValueString(), stateGroupsMap)
                return
            }
			deletedGroups = append(deletedGroups, stateKey)
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewTeamGroupMembershipModify().TeamId(plan.TeamId.ValueString()).GroupId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedGroups, plan.TeamId.ValueString(), stateGroupsMap)
				r.RevertModified(ctx, modifiedGroups, plan.TeamId.ValueString(), stateGroupsMap)
                return
            }
			modifiedGroups = append(modifiedGroups, stateKey)
        }
    }

	createdGroups := make([]string, 0)
    for planKey, planValue := range planGroupsMap {
        _, exists := stateGroupsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewTeamGroupMembershipCreate().TeamId(plan.TeamId.ValueString()).GroupId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				r.RevertDeleted(ctx, deletedGroups, plan.TeamId.ValueString(), stateGroupsMap)
				r.RevertModified(ctx, modifiedGroups, plan.TeamId.ValueString(), stateGroupsMap)
				r.RevertCreated(ctx, createdGroups, plan.TeamId.ValueString())
                return
            }
			createdGroups = append(createdGroups, planKey)
        }
    }

    teamGroupResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamGroupResponse.Code),
        )
        return
    }

    plan.ReadFromResponse(ctx, teamGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamGroupMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data, state model.TeamGroupMemberships

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    stateGroupsMap := make(map[string]string)
    for _, group := range state.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
            stateGroupsMap[groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()] = groupElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }
	
	deletedGroups := make([]string, 0)
    for _, group := range data.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
			groupId := groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewTeamGroupMembershipDelete()
            svc.TeamId(data.TeamId.ValueString())
            svc.GroupId(groupId)

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Team Group Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

				r.RevertDeleted(ctx, deletedGroups, data.TeamId.ValueString(), stateGroupsMap)
                return
            }
			deletedGroups = append(deletedGroups, groupId)
        }
    }
}

func (r *teamGroupMembership) RevertDeleted(ctx context.Context, toRevert []string, teamId string, stateGroupsMap map[string]string) {
	for _, groupId := range toRevert {
		svc := r.GetClient().NewTeamGroupMembershipCreate()
		svc.TeamId(teamId)
		svc.GroupId(groupId)
		svc.Role(stateGroupsMap[groupId])
		svc.Do(ctx)
	}
}

func (r *teamGroupMembership) RevertModified(ctx context.Context, toRevert []string, teamId string, stateGroupsMap map[string]string) {
	for _, groupId := range toRevert {
		svc := r.GetClient().NewTeamGroupMembershipModify()
		svc.TeamId(teamId)
		svc.GroupId(groupId)
		svc.Role(stateGroupsMap[groupId])
		svc.Do(ctx)
	}
}

func (r *teamGroupMembership) RevertCreated(ctx context.Context, toRevert []string, teamId string) {
	for _, groupId := range toRevert {
		svc := r.GetClient().NewTeamGroupMembershipDelete()
		svc.TeamId(teamId)
		svc.GroupId(groupId)
		svc.Do(ctx)
	}
}
