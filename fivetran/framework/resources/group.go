package resources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

func Group() resource.Resource {
    return &group{}
}

type group struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &group{}
var _ resource.ResourceWithImportState = &group{}

func (r *group) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *group) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.GroupResource()
}

func (r *group) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func (r *group) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Group

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    if resp.Diagnostics.HasError() {
        return
    }

    svc := r.GetClient().NewGroupCreate()
    svc.Name(data.Name.ValueString())

    groupCreateResponse, err := svc.Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Group Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, groupCreateResponse.Code, groupCreateResponse.Message),
        )

        return
    }

    data.ReadFromResponse(ctx, groupCreateResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *group) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Group

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    groupReadResponse, err := r.GetClient().NewGroupDetails().GroupID(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Group Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, groupReadResponse.Code, groupReadResponse.Message),
        )
        return
    }

    data.ReadFromResponse(ctx, groupReadResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *group) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.Group
    hasChanges := false

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    svc := r.GetClient().NewGroupUpdate().GroupID(state.Id.ValueString())
    
    if !plan.Name.Equal(state.Name) {
        svc.Name(plan.Name.ValueString())
        hasChanges = true
    }

    if hasChanges {
        groupUpdateResponse, err := svc.Do(ctx)        
        if err != nil {
            resp.Diagnostics.AddError(
                "Unable to Update Group Resource.",
                fmt.Sprintf("%v; code: %v; message: %v", err, groupUpdateResponse.Code, groupUpdateResponse.Message),
            )
            return
        }

        state.ReadFromResponse(ctx, groupUpdateResponse)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *group) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Group

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewGroupDelete().GroupID(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Group Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}