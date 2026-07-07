package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ConnectionV2PauseStateResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this pause-state resource. This is the connection ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_id": resourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the connection whose pause state is managed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"paused": resourceSchema.BoolAttribute{
				Required:    true,
				Description: "Whether the connection should be paused.",
			},
		},
		Version: 0,
	}
}
