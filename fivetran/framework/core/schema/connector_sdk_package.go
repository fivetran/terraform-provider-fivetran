package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ConnectorSdkPackageResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "Package ID (two-word format, e.g. 'happy_harmony').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_path": resourceSchema.StringAttribute{
				Required:    true,
				Description: "Path to the .zip file to upload. File is read during plan (for change detection) and during apply (for upload).",
			},
			"file_sha256_hash": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 hash of the uploaded file as computed and stored by the API. Used for upstream drift detection.",
			},
			"created_at": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the package was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the package was last updated.",
			},
		},
	}
}
