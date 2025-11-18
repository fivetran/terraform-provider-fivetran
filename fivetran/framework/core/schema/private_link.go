package schema

import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func PrivateLinkResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The unique identifier for the private link within the Fivetran system.",
            },
            "region": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
            },
            "name": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "The private link name within the account. The name must start with a letter or underscore and can only contain letters, numbers, or underscores. Maximum size of name is 23 characters.",
            },
            "service": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "Service type.",
            },
            "cloud_provider": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The cloud provider name.",
            },
            "state": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The state of the private link.",
            },
            "state_summary": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The state of the private link.",
            },
            "created_at": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The date and time the membership was created.",
            },
            "created_by": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The unique identifier for the User within the Fivetran system.",
            },
            "host": resourceSchema.StringAttribute{
                Computed:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
                Description: "The private link host.",
            },
            "config_map": resourceSchema.MapAttribute{
                ElementType: types.StringType,
                Required:    true,
                PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
                MarkdownDescription: `Configuration.

#### Possible values  
-- ` + "`connection_service_name` (String)" + `: The name of your connection service.
-- ` + "`account_url` (String)" + `: The URL of your account.
-- ` + "`vpce_id` (String)" + `: The ID of your Virtual Private Cloud Endpoint.
-- ` + "`aws_account_id` (String)" + `: The ID of your AWS account.
-- ` + "`cluster_identifier` (String)" + `: The cluster identifier.
-- ` + "`connection_service_id` (String)" + `: The ID of your connection service.
-- ` + "`workspace_url` (String)" + `: The URL of your workspace.
-- ` + "`pls_id` (String)" + `: The ID of your Azure Private Link service.
-- ` + "`sub_resource_name` (String)" + `: The name of subresource.
-- ` + "`private_dns_regions` (String)" + `: Private DNS Regions.
-- ` + "`private_connection_service_id` (String)" + `: The ID of your connection service.`,
            },
        },
    }
}

func PrivateLinkDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the private link within the Fivetran system.",
            },
            "region": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
            },
            "name": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The private link name within the account. The name must start with a letter or underscore and can only contain letters, numbers, or underscores. Maximum size of name is 23 characters.",
            },
            "service": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "Service type.",
            },
            "cloud_provider": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The cloud provider name.",
            },
            "state": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The state of the private link.",
            },
            "state_summary": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The state of the private link.",
            },
            "created_at": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The date and time the membership was created.",
            },
            "created_by": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the User within the Fivetran system.",
            },
            "host": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The private link host.",
            },
            "config_map": resourceSchema.MapAttribute{
                ElementType: types.StringType,
                Computed:    true,
                MarkdownDescription: `Configuration.`,
            },
        },
    }
}

func PrivateLinksDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Blocks: map[string]datasourceSchema.Block{
            "items": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the private link within the Fivetran system.",
                        },
                        "region": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
                        },
                        "name": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The private link name within the account. The name must start with a letter or underscore and can only contain letters, numbers, or underscores. Maximum size of name is 23 characters.",
                        },
                        "service": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "Service type.",
                        },
                        "cloud_provider": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The cloud provider name.",
                        },
                        "state": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The state of the private link.",
                        },
                        "state_summary": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The state of the private link.",
                        },
                        "created_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The date and time the membership was created.",
                        },
                        "created_by": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the User within the Fivetran system.",
                        },
                        "host": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The private link host.",
                        },
                    },
                },
            },
        },
    }
}
