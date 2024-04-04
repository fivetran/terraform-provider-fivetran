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

func DestinationFingerprint() resource.Resource {
    return &destinationFingerprint{}
}

type destinationFingerprint struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &destinationFingerprint{}
var _ resource.ResourceWithImportState = &destinationFingerprint{}

func (r *destinationFingerprint) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_destination_fingerprints"
}

func (r *destinationFingerprint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.FingerprintDestinationResource()
}

func (r *destinationFingerprint) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *destinationFingerprint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.FingerprintDestination

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    for _, item := range data.Fingerprint.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewCertificateDestinationFingerprintApprove()
            svc.DestinationID(data.DestinationId.ValueString())
            svc.Hash(element.Attributes()["hash"].(basetypes.StringValue).ValueString())
            svc.PublicKey(element.Attributes()["public_key"].(basetypes.StringValue).ValueString())
            response, err := svc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Destination Fingerprint Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
                )

                return
            }
        }
    }

    listResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.DestinationId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Destination Fingerprint Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, listResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destinationFingerprint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.FingerprintDestination
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    listResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.DestinationId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Destination Fingerprint Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, listResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *destinationFingerprint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.FingerprintDestination

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planMap := make(map[string]string)
    for _, item := range plan.Fingerprint.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            planMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["public_key"].(basetypes.StringValue).ValueString()
        }
    }

    stateMap := make(map[string]string)
    for _, item := range state.Fingerprint.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            stateMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["public_key"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
    for stateKey, _ := range stateMap {
        _, found := planMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewDestinationFingerprintRevoke().DestinationID(plan.DestinationId.ValueString()).Hash(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Destination Fingerprint Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    for planKey, planValue := range planMap {
        _, exists := stateMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewCertificateDestinationFingerprintApprove().DestinationID(plan.DestinationId.ValueString()).Hash(planKey).PublicKey(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Destination Fingerprint Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    listResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.DestinationId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Destination Fingerprint Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, listResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *destinationFingerprint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.FingerprintDestination

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    for _, items := range data.Fingerprint.Elements() {
        if element, ok := items.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewDestinationFingerprintRevoke()
            svc.DestinationID(data.DestinationId.ValueString())
            svc.Hash(element.Attributes()["hash"].(basetypes.StringValue).ValueString())

            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Destination Fingerprint Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

                return
            }
        }
    }
}