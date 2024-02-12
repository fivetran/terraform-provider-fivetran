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

func TeamUserMembership() resource.Resource {
    return &teamUserMembership{}
}

type teamUserMembership struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &teamUserMembership{}
var _ resource.ResourceWithImportState = &teamUserMembership{}

func (r *teamUserMembership) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_team_user_membership"
}

func (r *teamUserMembership) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.TeamUserMembershipResource()
}

func (r *teamUserMembership) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamUserMembership) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamUserMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewTeamUserMembershipCreate()
            svc.TeamId(data.TeamId.ValueString())
            svc.UserId(userElement.Attributes()["user_id"].(basetypes.StringValue).ValueString())
            svc.Role(userElement.Attributes()["role"].(basetypes.StringValue).ValueString())
            if teamUserResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Team User Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, teamUserResponse.Code, teamUserResponse.Message),
                )

                return
            }
        }
    }

    teamUserResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team User Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamUserResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamUserResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamUserMembership) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamUserMemberships
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    teamUserResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team User Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamUserResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, teamUserResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamUserMembership) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.TeamUserMemberships

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planUsersMap := make(map[string]string)
    for _, user := range plan.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            planUsersMap[userElement.Attributes()["user_id"].(basetypes.StringValue).ValueString()] = userElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    stateUsersMap := make(map[string]string)
    for _, user := range state.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            stateUsersMap[userElement.Attributes()["user_id"].(basetypes.StringValue).ValueString()] = userElement.Attributes()["role"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
    for stateKey, stateValue := range stateUsersMap {
        role, found := planUsersMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewTeamUserMembershipDelete().TeamId(plan.TeamId.ValueString()).UserId(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team User Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        } else if role != stateValue {
            if updateResponse, err := r.GetClient().NewTeamUserMembershipModify().TeamId(plan.TeamId.ValueString()).UserId(stateKey).Role(role).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team User Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    for planKey, planValue := range planUsersMap {
        _, exists := stateUsersMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewTeamUserMembershipCreate().TeamId(plan.TeamId.ValueString()).UserId(planKey).Role(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Team User Membership Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    teamUserResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Team User Memberships Resource.",
            fmt.Sprintf("%v; code: %v", err, teamUserResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, teamUserResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamUserMembership) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.TeamUserMemberships

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    for _, user := range data.User.Elements() {
        if userElement, ok := user.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewTeamUserMembershipDelete()
            svc.TeamId(data.TeamId.ValueString())
            svc.UserId(userElement.Attributes()["user_id"].(basetypes.StringValue).ValueString())

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Team User Memberships Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

                return
            }
        }
    }
}