package resources

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxPackageFileSize = 200 * 1024 * 1024 // 200MB

func ConnectorSdkPackage() resource.Resource {
	return &connectorSdkPackage{}
}

type connectorSdkPackage struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectorSdkPackage{}
var _ resource.ResourceWithImportState = &connectorSdkPackage{}
var _ resource.ResourceWithModifyPlan = &connectorSdkPackage{}

func (r *connectorSdkPackage) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_sdk_package"
}

func (r *connectorSdkPackage) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ConnectorSdkPackageResource()
}

func (r *connectorSdkPackage) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ModifyPlan computes the local SHA-256 hash at plan time so file changes are detected in the plan diff.
func (r *connectorSdkPackage) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan model.ConnectorSdkPackage
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

	var stateHash string
	if !req.State.Raw.IsNull() {
		var state model.ConnectorSdkPackage
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		stateHash = state.FileSha256Hash.ValueString()
	}

	if localHash != stateHash {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("file_sha256_hash"), types.StringValue(localHash))...)
	}
}

func (r *connectorSdkPackage) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectorSdkPackage
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileBytes, localHash := readFileAndHash(data.FilePath.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.GetClient().NewConnectorSdkPackageCreate().
		FileContent(bytes.NewReader(fileBytes)).
		Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResp.Code, createResp.Message),
		)
		return
	}

	if createResp.Data.FileSha256Hash != "" && createResp.Data.FileSha256Hash != localHash {
		resp.Diagnostics.AddError(
			"Upload Corruption Detected",
			fmt.Sprintf("Local hash %q does not match API-returned hash %q. The upload may be corrupted.", localHash, createResp.Data.FileSha256Hash),
		)
		return
	}

	data.ReadFromResponse(createResp)
	data.FileSha256Hash = types.StringValue(localHash)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSdkPackage) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectorSdkPackage
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.GetClient().NewConnectorSdkPackageDetails().
		PackageID(data.ID.ValueString()).
		Do(ctx)

	if err != nil {
		if strings.HasPrefix(readResp.Code, "NotFound") {
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
	data.ReadFromResponse(readResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSdkPackage) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state model.ConnectorSdkPackage
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileBytes, localHash := readFileAndHash(plan.FilePath.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip upload if file hasn't changed — only sync file_path into state
	if localHash == state.FileSha256Hash.ValueString() {
		state.FilePath = plan.FilePath
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	updateResp, err := r.GetClient().NewConnectorSdkPackageUpdate().
		PackageID(state.ID.ValueString()).
		FileContent(bytes.NewReader(fileBytes)).
		Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, updateResp.Code, updateResp.Message),
		)
		return
	}

	if updateResp.Data.FileSha256Hash != "" && updateResp.Data.FileSha256Hash != localHash {
		resp.Diagnostics.AddError(
			"Upload Corruption Detected",
			fmt.Sprintf("Local hash %q does not match API-returned hash %q. The upload may be corrupted.", localHash, updateResp.Data.FileSha256Hash),
		)
		return
	}

	state.FilePath = plan.FilePath
	state.ReadFromResponse(updateResp)
	state.FileSha256Hash = types.StringValue(localHash)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectorSdkPackage) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectorSdkPackage
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.GetClient().NewConnectorSdkPackageDelete().
		PackageID(data.ID.ValueString()).
		Do(ctx)

	if err != nil {
		// Treat 404 as already deleted — idempotent destroy
		if strings.HasPrefix(deleteResp.Code, "NotFound") {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Delete Connector SDK Package",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResp.Code, deleteResp.Message),
		)
	}
}

// readFileAndHash reads the file at filePath and returns its bytes and hex-encoded SHA-256 hash.
// On failure, appends an error to the provided diagnostics and returns nil, "".
func readFileAndHash(filePath string, diags *diag.Diagnostics) ([]byte, string) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		diags.AddError("Cannot read file", fmt.Sprintf("Cannot read file %q: %s", filePath, err))
		return nil, ""
	}
	if len(fileBytes) > maxPackageFileSize {
		diags.AddError("File too large", fmt.Sprintf("File %q exceeds 200MB limit.", filePath))
		return nil, ""
	}
	digest := sha256.Sum256(fileBytes)
	return fileBytes, hex.EncodeToString(digest[:])
}
