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

func ConnectorCertificate() resource.Resource {
    return &connectorCertificate{}
}

type connectorCertificate struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectorCertificate{}
var _ resource.ResourceWithImportState = &connectorCertificate{}

func (r *connectorCertificate) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_connector_certificates"
}

func (r *connectorCertificate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.CertificateConnectorResource()
}

func (r *connectorCertificate) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectorCertificate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.CertificateConnector

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

    planMap := make(map[string]string)
    for _, item := range data.Certificate.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewCertificateConnectorCertificateApprove()
            svc.ConnectorID(data.ConnectorId.ValueString())
            svc.Hash(element.Attributes()["hash"].(basetypes.StringValue).ValueString())
            svc.EncodedCert(element.Attributes()["encoded_cert"].(basetypes.StringValue).ValueString())
            response, err := svc.Do(ctx)
            if err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Create Connector Certificate Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
                )

                return
            }

            planMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["encoded_cert"].(basetypes.StringValue).ValueString()
        }
    }

    listResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.ConnectorId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Connector Certificate Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

    data.ReadFromResponse(ctx, listResponse, planMap)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorCertificate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.CertificateConnector
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    stateMap := make(map[string]string)
    for _, item := range data.Certificate.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            stateMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["encoded_cert"].(basetypes.StringValue).ValueString()
        }
    }

    listResponse, err := data.ReadFromSource(ctx, r.GetClient(), data.ConnectorId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Connector Certificate Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }


    data.ReadFromResponse(ctx, listResponse, stateMap)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorCertificate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.CertificateConnector

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    planMap := make(map[string]string)
    for _, item := range plan.Certificate.Elements() {
        if element, ok := item.(basetypes.ObjectValue); ok {
            planMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["encoded_cert"].(basetypes.StringValue).ValueString()
        }
    }

    stateMap := make(map[string]string)
    for _, connector := range state.Certificate.Elements() {
        if element, ok := connector.(basetypes.ObjectValue); ok {
            stateMap[element.Attributes()["hash"].(basetypes.StringValue).ValueString()] = element.Attributes()["encoded_cert"].(basetypes.StringValue).ValueString()
        }
    }

    /* sync */
    for stateKey, _ := range stateMap {
        _, found := planMap[stateKey]

        if !found {
            if updateResponse, err := r.GetClient().NewConnectorCertificateRevoke().ConnectorID(plan.ConnectorId.ValueString()).Hash(stateKey).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Connector Certificate Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    for planKey, planValue := range planMap {
        _, exists := stateMap[planKey]

        if !exists {
            if updateResponse, err := r.GetClient().NewCertificateConnectorCertificateApprove().ConnectorID(plan.ConnectorId.ValueString()).Hash(planKey).EncodedCert(planValue).Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Update Connector Certificate Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
                )
                return
            }
        }
    }

    listResponse, err := plan.ReadFromSource(ctx, r.GetClient(), plan.ConnectorId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Connector Certificate Resource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

    plan.ReadFromResponse(ctx, listResponse, planMap)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorCertificate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.CertificateConnector

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    for _, items := range data.Certificate.Elements() {
        if element, ok := items.(basetypes.ObjectValue); ok {
            svc := r.GetClient().NewConnectorCertificateRevoke()
            svc.ConnectorID(data.ConnectorId.ValueString())
            svc.Hash(element.Attributes()["hash"].(basetypes.StringValue).ValueString())
            if deleteResponse, err := svc.Do(ctx); err != nil {
                resp.Diagnostics.AddError(
                    "Unable to Delete Connector Certificate Resource.",
                    fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
                )

                return
            }
        }
    }
}