package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ConnectionConfigResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   "The unique identifier for this configuration (same as connection_id).",
			},
			"connection_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the connection within the Fivetran system.",
			},
			"config": schema.StringAttribute{
				Optional:    true,
				CustomType:  fivetrantypes.JsonConfigType{},
				Description: "Connection config in Json format, following [Fivetran API endpoint contract](https://fivetran.com/docs/rest-api/api-reference/connections/create-connection) for `config` field",
			},
			"auth": schema.StringAttribute{
				Optional:    true,
				CustomType:  fivetrantypes.JsonConfigType{},
				Description: "Connection auth config in Json format, following [Fivetran API endpoint contract](https://fivetran.com/docs/rest-api/api-reference/connections/create-connection) for `auth` field",
			},
		},
	}
}