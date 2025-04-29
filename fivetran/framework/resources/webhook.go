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

func Webhook() resource.Resource {
    return &webhook{}
}

type webhook struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &webhook{}
var _ resource.ResourceWithImportState = &webhook{}

func (r *webhook) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhook) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.WebhookResource()
}

func (r *webhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


func (r *webhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Webhook

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    if resp.Diagnostics.HasError() {
        return
    }

    if data.Type.ValueString() == "account" {
        r.createAccount(ctx, data, resp)
    } else if data.Type.ValueString() == "group" && !data.GroupId.IsUnknown() && !data.GroupId.IsNull() {
        r.createGroup(ctx, data, resp)
    } else {
        resp.Diagnostics.AddError(
            "Incorrect webhook type",
            "Available values for type field is account or group. If you specify type = group, you need to set group_id",
        )
    }
}

func (r *webhook) createAccount(ctx context.Context, data model.Webhook, resp *resource.CreateResponse) {
    svc := r.GetClient().NewWebhookAccountCreate()
    svc.Url(data.Url.ValueString())
    svc.Active(core.GetBoolOrDefault(data.Active, false))
    svc.Secret(data.Secret.ValueString())

    elements := make([]string, 0, len(data.Events.Elements()))
    
    diag := data.Events.ElementsAs(ctx, &elements, false)
    resp.Diagnostics.Append(diag...)
    if resp.Diagnostics.HasError() {
        return
    }

    svc.Events(elements)

    webhookResponse, err := svc.Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Webhook Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, webhookResponse.Code, webhookResponse.Message),
        )

        return
    }

    data.ReadFromResponse(ctx, webhookResponse)

    runTests := core.GetBoolOrDefault(data.RunTests, false)

    if runTests {
        testsSvc := r.GetClient().NewWebhookTest().WebhookId(data.Id.ValueString())
        for _, varValue := range data.Events.Elements() {
            testsSvc.Event(varValue.String())
            response, err := testsSvc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Start Webhook Tests.",
                    fmt.Sprintf("%v; code: %v", err, response.Code),
                )
            }
        }

        // nothing to read
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhook) createGroup(ctx context.Context, data model.Webhook, resp *resource.CreateResponse) {
    svc := r.GetClient().NewWebhookGroupCreate()
    svc.GroupId(data.GroupId.ValueString())
    svc.Url(data.Url.ValueString())
    svc.Active(core.GetBoolOrDefault(data.Active, false))
    svc.Secret(data.Secret.ValueString())

    elements := make([]string, 0, len(data.Events.Elements()))
    
    diag := data.Events.ElementsAs(ctx, &elements, false)
    resp.Diagnostics.Append(diag...)
    if resp.Diagnostics.HasError() {
        return
    }
    
    svc.Events(elements)

    webhookResponse, err := svc.Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Webhook Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, webhookResponse.Code, webhookResponse.Message),
        )

        return
    }

    data.ReadFromResponse(ctx, webhookResponse)

    runTests := core.GetBoolOrDefault(data.RunTests, false)

    if runTests {
        testsSvc := r.GetClient().NewWebhookTest().WebhookId(data.Id.ValueString())
        for _, varValue := range data.Events.Elements() {
            testsSvc.Event(varValue.String())
            response, err := testsSvc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Start Webhook Tests.",
                    fmt.Sprintf("%v; code: %v", err, response.Code),
                )
            }
        }

        // nothing to read
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Webhook

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    webhookResponse, err := r.GetClient().NewWebhookDetails().WebhookId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Webhook Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, webhookResponse.Code, webhookResponse.Message),
        )
        return
    }

    data.ReadFromResponse(ctx, webhookResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.Webhook
    hasChanges := false

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    svc := r.GetClient().NewWebhookUpdate().WebhookId(state.Id.ValueString())
    
    active := core.GetBoolOrDefault(plan.Active, false)
    activeState := core.GetBoolOrDefault(state.Active, false)
    runTests := core.GetBoolOrDefault(plan.RunTests, false)
    runTestsState := core.GetBoolOrDefault(state.RunTests, false)

    if !plan.Url.Equal(state.Url) {
        svc.Url(plan.Url.ValueString())
        hasChanges = true
    }

    if !plan.Secret.Equal(state.Secret) {
        svc.Secret(plan.Secret.ValueString())
        state.Secret = plan.Secret
        hasChanges = true
    }

    if active != activeState {
        svc.Active(active)
        hasChanges = true
    }

    if !plan.Events.Equal(state.Events) {
        elements := make([]string, 0, len(plan.Events.Elements()))
    
        diag := plan.Events.ElementsAs(ctx, &elements, false)
        resp.Diagnostics.Append(diag...)
        if resp.Diagnostics.HasError() {
            return
        }
    
        svc.Events(elements)
        hasChanges = true
    }

    if hasChanges {
        webhookResponse, err := svc.Do(ctx)        
        if err != nil {
            resp.Diagnostics.AddError(
                "Unable to Update Webhook Resource.",
                fmt.Sprintf("%v; code: %v; message: %v", err, webhookResponse.Code, webhookResponse.Message),
            )
            return
        }

        state.ReadFromResponse(ctx, webhookResponse)
    }

    if runTests && runTests != runTestsState {
        testsSvc := r.GetClient().NewWebhookTest().WebhookId(state.Id.ValueString())
        for _, varValue := range state.Events.Elements() {
            testsSvc.Event(varValue.String())
            response, err := testsSvc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Start Webhook Tests.",
                    fmt.Sprintf("%v; code: %v", err, response.Code),
                )
            }
        }

        // nothing to read
        state.RunTests = plan.RunTests
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Webhook

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewWebhookDelete().WebhookId(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Webhook Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}
