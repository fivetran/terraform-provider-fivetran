package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WebhookResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The webhook ID",
			},
			"type": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The webhook type (group, account)",
			},
			"group_id": resourceSchema.StringAttribute{
				Optional:	   true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The group ID",
			},
			"url": resourceSchema.StringAttribute{
				Required:      true,
				Description:   "Your webhooks URL endpoint for your application",
			},
			"events": resourceSchema.SetAttribute{
				Required:      true,
				Description:   "The array of event types",
				ElementType:   types.StringType,
			},
			"active": resourceSchema.BoolAttribute{
				Required:      true,
				Description:   "Boolean, if set to true, webhooks are immediately sent in response to events",
			},
			"secret": resourceSchema.StringAttribute{
				Required:      true,
				Sensitive:     true,
				Description:   "The secret string used for payload signing and masked in the response.",
			},
			"created_at": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The webhook creation timestamp",
			},
			"created_by": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The ID of the user who created the webhook.",
			},
			"run_tests": resourceSchema.BoolAttribute{
				Optional:	   true,
				Description:   "Specifies whether the setup tests should be run",
			},
		},
	}
}

func webhookDatasourceAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			Required:      true,
			Description:   "The webhook ID",
		},
		"type": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "The webhook type (group, account)",
		},
		"group_id": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "The group ID",
		},
		"url": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "Your webhooks URL endpoint for your application",
		},
		"events": datasourceSchema.SetAttribute{
			Computed:      true,
			Description:   "The array of event types",
			ElementType:   types.StringType,
		},
		"active": datasourceSchema.BoolAttribute{
			Computed:      true,
			Description:   "Boolean, if set to true, webhooks are immediately sent in response to events",
		},
		"secret": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "The secret string used for payload signing and masked in the response.",
		},
		"created_at": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "The webhook creation timestamp",
		},
		"created_by": datasourceSchema.StringAttribute{
			Computed:      true,
			Description:   "The ID of the user who created the webhook.",
		},
		"run_tests": datasourceSchema.BoolAttribute{
			Computed:      true,
			Description:   "Specifies whether the setup tests should be run",
		},
	}
}

func WebhookDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: webhookDatasourceAttributes(),
	}
}

func WebhooksDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"webhooks": datasourceSchema.SetNestedAttribute{
				Computed:      true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: webhookDatasourceAttributes(),
				},
			},
		},
	}
}