package resources

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxPackageFileSize = 200 * 1024 * 1024 // 200MB

func ConnectorSdkPackage() resource.Resource {
	return &connectorSdkPackage{}
}

type connectorSdkPackage struct {
	core.ProviderResource
}

// Model — all fields in a single struct, no external model file.
type connectorSdkPackageModel struct {
	ID             types.String `tfsdk:"id"`
	FilePath       types.String `tfsdk:"file_path"`
	FileSha256Hash types.String `tfsdk:"file_sha256_hash"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

var _ resource.ResourceWithConfigure = &connectorSdkPackage{}
var _ resource.ResourceWithImportState = &connectorSdkPackage{}
var _ resource.ResourceWithModifyPlan = &connectorSdkPackage{}

func (r *connectorSdkPackage) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_sdk_package"
}

func (r *connectorSdkPackage) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Package ID (two-word format, e.g. 'happy_harmony').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_path": schema.StringAttribute{
				Required:    true,
				Description: "Path to the .zip file to upload. File is read during plan (for change detection) and during apply (for upload).",
			},
			"file_sha256_hash": schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 hash of the uploaded file as computed and stored by the API. Used for upstream drift detection.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the package was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the package was last updated.",
			},
		},
	}
}

func (r *connectorSdkPackage) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan — plan-time change detection by reading the file and computing SHA-256.
func (r *connectorSdkPackage) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan connectorSdkPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filePath := plan.FilePath.ValueString()
	if filePath == "" {
		return
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot read file",
			fmt.Sprintf("Cannot read file %q: %s", filePath, err),
		)
		return
	}

	if len(fileBytes) > maxPackageFileSize {
		resp.Diagnostics.AddError(
			"File too large",
			fmt.Sprintf("File %q is %d bytes, exceeds 200MB limit.", filePath, len(fileBytes)),
		)
		return
	}

	digest := sha256.Sum256(fileBytes)
	localHash := hex.EncodeToString(digest[:])

	// Compare against current state hash
	var stateHash string
	if !req.State.Raw.IsNull() {
		var state connectorSdkPackageModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if !resp.Diagnostics.HasError() {
			stateHash = state.FileSha256Hash.ValueString()
		}
	}

	if localHash != stateHash {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("file_sha256_hash"), types.StringValue(localHash))...)
	}
}

func (r *connectorSdkPackage) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectorSdkPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read file and compute local hash
	fileBytes, err := os.ReadFile(data.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Cannot read file", fmt.Sprintf("Cannot read file %q: %s", data.FilePath.ValueString(), err))
		return
	}

	if len(fileBytes) > maxPackageFileSize {
		resp.Diagnostics.AddError("File too large", fmt.Sprintf("File %q exceeds 200MB limit.", data.FilePath.ValueString()))
		return
	}

	digest := sha256.Sum256(fileBytes)
	localHash := hex.EncodeToString(digest[:])

	// Upload via API
	file, err := os.Open(data.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Cannot open file", fmt.Sprintf("Cannot open file %q: %s", data.FilePath.ValueString(), err))
		return
	}
	defer file.Close()

	createResp, err := r.GetClient().NewConnectorSdkPackageCreate().
		FileContent(file).
		Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResp.Code, createResp.Message),
		)
		return
	}

	// Upload corruption guard
	if createResp.Data.FileSha256Hash != "" && createResp.Data.FileSha256Hash != localHash {
		resp.Diagnostics.AddError(
			"Upload Corruption Detected",
			fmt.Sprintf("Local hash %q does not match API-returned hash %q. The upload may be corrupted.", localHash, createResp.Data.FileSha256Hash),
		)
		return
	}

	data.ID = types.StringValue(createResp.Data.ID)
	data.FileSha256Hash = types.StringValue(localHash)
	data.CreatedAt = types.StringValue(createResp.Data.CreatedAt.String())
	data.UpdatedAt = types.StringValue(createResp.Data.UpdatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSdkPackage) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectorSdkPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.GetClient().NewConnectorSdkPackageDetails().
		PackageID(data.ID.ValueString()).
		Do(ctx)

	if err != nil {
		if readResp.Code == "NotFound" || strings.Contains(err.Error(), "status code: 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, readResp.Code, readResp.Message),
		)
		return
	}

	// Update computed fields — do NOT update file_path (user-supplied)
	if readResp.Data.FileSha256Hash != "" {
		data.FileSha256Hash = types.StringValue(readResp.Data.FileSha256Hash)
	}
	data.CreatedAt = types.StringValue(readResp.Data.CreatedAt.String())
	data.UpdatedAt = types.StringValue(readResp.Data.UpdatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSdkPackage) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var plan, state connectorSdkPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read file and compute local hash
	fileBytes, err := os.ReadFile(plan.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Cannot read file", fmt.Sprintf("Cannot read file %q: %s", plan.FilePath.ValueString(), err))
		return
	}

	if len(fileBytes) > maxPackageFileSize {
		resp.Diagnostics.AddError("File too large", fmt.Sprintf("File %q exceeds 200MB limit.", plan.FilePath.ValueString()))
		return
	}

	digest := sha256.Sum256(fileBytes)
	localHash := hex.EncodeToString(digest[:])

	// If local hash matches state hash — file hasn't changed, only update file_path in state
	if localHash == state.FileSha256Hash.ValueString() {
		state.FilePath = plan.FilePath
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	// File changed — upload new version
	file, err := os.Open(plan.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Cannot open file", fmt.Sprintf("Cannot open file %q: %s", plan.FilePath.ValueString(), err))
		return
	}
	defer file.Close()

	updateResp, err := r.GetClient().NewConnectorSdkPackageUpdate().
		PackageID(state.ID.ValueString()).
		FileContent(file).
		Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, updateResp.Code, updateResp.Message),
		)
		return
	}

	// Upload corruption guard
	if updateResp.Data.FileSha256Hash != "" && updateResp.Data.FileSha256Hash != localHash {
		resp.Diagnostics.AddError(
			"Upload Corruption Detected",
			fmt.Sprintf("Local hash %q does not match API-returned hash %q. The upload may be corrupted.", localHash, updateResp.Data.FileSha256Hash),
		)
		return
	}

	state.ID = types.StringValue(updateResp.Data.ID)
	state.FilePath = plan.FilePath
	state.FileSha256Hash = types.StringValue(localHash)
	state.CreatedAt = types.StringValue(updateResp.Data.CreatedAt.String())
	state.UpdatedAt = types.StringValue(updateResp.Data.UpdatedAt.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectorSdkPackage) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError("Unconfigured Fivetran Client", "Please report this issue to the provider developers.")
		return
	}

	var data connectorSdkPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.GetClient().NewConnectorSdkPackageDelete().
		PackageID(data.ID.ValueString()).
		Do(ctx)

	if err != nil {
		// Treat 404 as already deleted
		if strings.Contains(err.Error(), "status code: 404") {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResp.Code, deleteResp.Message),
		)
	}
}
