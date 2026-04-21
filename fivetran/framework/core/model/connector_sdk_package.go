package model

import (
	connectorsdk "github.com/fivetran/go-fivetran/connector_sdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectorSdkPackage struct {
	ID             types.String `tfsdk:"id"`
	FilePath       types.String `tfsdk:"file_path"`
	FileSha256Hash types.String `tfsdk:"file_sha256_hash"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (d *ConnectorSdkPackage) ReadFromResponse(resp connectorsdk.ConnectorSdkPackageResponse) {
	d.ID = types.StringValue(resp.Data.ID)
	if resp.Data.FileSha256Hash != "" {
		d.FileSha256Hash = types.StringValue(resp.Data.FileSha256Hash)
	}
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt.String())
	d.UpdatedAt = types.StringValue(resp.Data.UpdatedAt.String())
}
