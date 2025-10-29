package resources

import (
    "context"
    "fmt"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type userType struct {
    role string
    id   string
}

func GroupUser() resource.Resource {
    return &groupUser{}
}

type groupUser struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &groupUser{}
var _ resource.ResourceWithImportState = &groupUser{}

func (r *groupUser) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_group_users"
}

func (r *groupUser) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.GroupUsersResource()
}

func (r *groupUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *groupUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.GroupUser

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    planUserMap := make(map[string]userType)
    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            planUserMap[userElement.Attributes()["email"].(basetypes.StringValue).ValueString()] = 
                    userType{
                        role: userElement.Attributes()["role"].(basetypes.StringValue).ValueString(),
                        id:   userElement.Attributes()["id"].(basetypes.StringValue).ValueString(),
                    }
        }
    }

    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewGroupAddUser()
            svc.GroupID(data.GroupId.ValueString())
            svc.Email(userElement.Attributes()["email"].(basetypes.StringValue).ValueString())
            svc.Role(userElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            groupUserResponse, err := svc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, groupUserResponse.Code, groupUserResponse.Message),
                )

                return
            }
        }
    }

    groupUserListResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.GroupId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Group User Resource.",
            fmt.Sprintf("%v; code: %v", err, groupUserListResponse.Code),
        )

        return
    }
    var groupUserResponseFinal groups.GroupListUsersResponse
    for _, localUser := range groupUserListResponse.Data.Items {
        _, exists := planUserMap[localUser.Email]
        if exists {
            groupUserResponseFinal.Data.Items = append(groupUserResponseFinal.Data.Items, localUser) 
        }
    }

    data.ReadFromResponse(ctx, groupUserResponseFinal)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *groupUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.GroupUser
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    var isImporting = data.GroupId.IsNull() || data.GroupId.IsUnknown()
    if isImporting {
        data.GroupId = data.Id
    }

    stateUserMap := make(map[string]userType)
    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            stateUserMap[userElement.Attributes()["email"].(basetypes.StringValue).ValueString()] = 
                    userType{
                        role: userElement.Attributes()["role"].(basetypes.StringValue).ValueString(),
                        id:   userElement.Attributes()["id"].(basetypes.StringValue).ValueString(),
                    }
        }
    }

    groupUserListResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.GroupId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Group User Resource.",
            fmt.Sprintf("%v; code: %v", err, groupUserListResponse.Code),
        )

        return
    }
    var groupUserResponseFinal groups.GroupListUsersResponse
    if isImporting {
        groupUserResponseFinal = groupUserListResponse // do not clean up on import
    } else {
        for _, localUser := range groupUserListResponse.Data.Items {
            _, exists := stateUserMap[localUser.Email]
            if exists {
                groupUserResponseFinal.Data.Items = append(groupUserResponseFinal.Data.Items, localUser)
            }
        }
    }

    data.ReadFromResponse(ctx, groupUserResponseFinal)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *groupUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.GroupUser

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planUserMap := make(map[string]userType)
    for _, user := range plan.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            planUserMap[userElement.Attributes()["email"].(basetypes.StringValue).ValueString()] = 
                    userType{
                        role: userElement.Attributes()["role"].(basetypes.StringValue).ValueString(),
                        id:   userElement.Attributes()["id"].(basetypes.StringValue).ValueString(),
                    }
        }
    }

    stateUserMap := make(map[string]userType)
    for _, user := range state.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            stateUserMap[userElement.Attributes()["email"].(basetypes.StringValue).ValueString()] = 
                    userType{
                        role: userElement.Attributes()["role"].(basetypes.StringValue).ValueString(),
                        id:   userElement.Attributes()["id"].(basetypes.StringValue).ValueString(),
                    }
        }
    }

    /* sync */
    for stateKey, stateValue := range stateUserMap {
        planUser, found := planUserMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewGroupRemoveUser().GroupID(plan.GroupId.ValueString()).UserID(stateValue.id).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        } else if planUser.role != stateValue.role {
            if deleteResponse, err := r.GetClient().NewGroupRemoveUser().GroupID(plan.GroupId.ValueString()).UserID(stateValue.id).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )
                return
            }

            if updateResponse, err := r.GetClient().NewGroupAddUser().GroupID(plan.GroupId.ValueString()).Email(stateKey).Role(planUser.role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    for planKey, planValue := range planUserMap {
        _, exists := stateUserMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewGroupAddUser().GroupID(plan.GroupId.ValueString()).Email(planKey).Role(planValue.role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    groupUserResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.GroupId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Group User Resource.",
            fmt.Sprintf("%v; code: %v", err, groupUserResponse.Code),
        )

        return
    }

    var groupUserResponseFinal groups.GroupListUsersResponse
    for _, localUser := range groupUserResponse.Data.Items {
        _, exists := planUserMap[localUser.Email]
        if exists {
            groupUserResponseFinal.Data.Items = append(groupUserResponseFinal.Data.Items, localUser) 
        }
    }

    plan.ReadFromResponse(ctx, groupUserResponseFinal)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.GroupUser

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewGroupRemoveUser()
            svc.GroupID(data.GroupId.ValueString())
            svc.UserID(userElement.Attributes()["id"].(basetypes.StringValue).ValueString())

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Group User Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

                return
            }
        }
    }
}