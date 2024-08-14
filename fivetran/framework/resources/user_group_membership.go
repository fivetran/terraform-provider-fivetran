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

func UserGroupMembership() resource.Resource {
    return &userGroupMembership{}
}

type userGroupMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &userGroupMembership{}
var _ resource.ResourceWithImportState = &userGroupMembership{}

func (r *userGroupMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user_group_membership"
}

func (r *userGroupMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.UserGroupMembershipResource()
}

func (r *userGroupMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userGroupMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserGroupMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	savedGroups := make([]string, 0)
    for _, group := range data.Group.Elements() {
        if groupElement, ok := group.(basetypes.ObjectValue); ok {
			groupId := groupElement.Attributes()["group_id"].(basetypes.StringValue).ValueString()
            svc := r.GetClient().NewUserGroupMembershipCreate()
            svc.UserId(data.UserId.ValueString())
            svc.GroupId(groupId)
            svc.Role(groupElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if userGroupResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create User Group Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, userGroupResponse.Code, userGroupResponse.Message),
                )

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, savedGroups, data.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
                return
            }

			savedGroups = append(savedGroups, groupId)
        }
    }

    userGroupResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userGroupResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userGroupMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.UserGroupMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    userGroupResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read User Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userGroupResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, userGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userGroupMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.UserGroupMemberships

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
            if updateResponse, err := r.GetClient().NewUserGroupMembershipDelete().UserId(plan.UserId.ValueString()).GroupId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedGroups, plan.UserId.ValueString(), stateGroupsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
			}
			deletedGroups = append(deletedGroups, stateKey)
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewUserGroupMembershipModify().UserId(plan.UserId.ValueString()).GroupId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )	

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedGroups, plan.UserId.ValueString(), stateGroupsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedGroups, plan.UserId.ValueString(), stateGroupsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)

                return
            }
			modifiedGroups = append(modifiedGroups, stateKey)
        }
    }

	createdGroups := make([]string, 0)
    for planKey, planValue := range planGroupsMap {
        _, exists := stateGroupsMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewUserGroupMembershipCreate().UserId(plan.UserId.ValueString()).GroupId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update User Group Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedGroups, plan.UserId.ValueString(), stateGroupsMap)	
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)

				modRevertMsg, modRevertErr := r.RevertModified(ctx, modifiedGroups, plan.UserId.ValueString(), stateGroupsMap)		
				resp.Diagnostics.AddWarning("Action reverted", modRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", modRevertErr)

				creRvertMsg, creRevertErr :=  r.RevertCreated(ctx, createdGroups, plan.UserId.ValueString())		
				resp.Diagnostics.AddWarning("Action reverted", creRvertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", creRevertErr)
                return
            }
			createdGroups = append(createdGroups, planKey)
        }
    }

    userGroupResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create User Group Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, userGroupResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, userGroupResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userGroupMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data, state model.UserGroupMemberships

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
            svc := r.GetClient().NewUserGroupMembershipDelete()
            svc.UserId(data.UserId.ValueString())
            svc.GroupId(groupId)

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete User Group Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

				delRevertMsg, delRevertErr :=  r.RevertDeleted(ctx, deletedGroups, data.UserId.ValueString(), stateGroupsMap);
				resp.Diagnostics.AddWarning("Action reverted", delRevertMsg)
				resp.Diagnostics.AddWarning("Action reverted failed", delRevertErr)
                return
            }
			deletedGroups = append(deletedGroups, groupId)
        }
    }
}

func (r *userGroupMembership) RevertDeleted(ctx context.Context, toRevert []string, userId string, stateGroupsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, groupId := range toRevert {
		svc := r.GetClient().NewUserGroupMembershipCreate()
		svc.UserId(userId)
		svc.GroupId(groupId)
		svc.Role(stateGroupsMap[groupId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, groupId)
		} else {
			reverted = append(reverted, groupId)
		} 
	}
	return fmt.Sprintf("Delete action reverted for groups: %v", reverted),
	fmt.Sprintf("Delete for revert action failed for groups: %v", failed)
}

func (r *userGroupMembership) RevertModified(ctx context.Context, toRevert []string, userId string, stateGroupsMap map[string]string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, groupId := range toRevert {
		svc := r.GetClient().NewUserGroupMembershipModify()
		svc.UserId(userId)
		svc.GroupId(groupId)
		svc.Role(stateGroupsMap[groupId])
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, groupId)
		} else {
			reverted = append(reverted, groupId)
		} 
	}
	return fmt.Sprintf("Modify action reverted for groups: %v", reverted),
	fmt.Sprintf("Modify for revert action failed for groups: %v", failed)
}

func (r *userGroupMembership) RevertCreated(ctx context.Context, toRevert []string, userId string) (string, string) {
	reverted := []string{}
	failed := []string{}
	for _, groupId := range toRevert {
		svc := r.GetClient().NewUserGroupMembershipDelete()
		svc.UserId(userId)
		svc.GroupId(groupId)
		if _, err := svc.Do(ctx); err != nil {
			failed = append(failed, groupId)
		} else {
			reverted = append(reverted, groupId)
		} 
	}
	return fmt.Sprintf("Create action reverted for groups: %v", reverted),
	fmt.Sprintf("Create for revert action failed for groups: %v", failed)
}

