package schema

import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func PrivateLinkResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:    true,
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
                Description: "The cloud provider name.",
            },
            "state": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The state of the private link.",
            },
            "state_summary": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The state of the private link.",
            },
            "created_at": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The date and time the membership was created.",
            },
            "created_by": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the User within the Fivetran system.",
            },
        },
        Blocks: map[string]resourceSchema.Block{
            "config": resourceSchema.SingleNestedBlock{
                Attributes: map[string]resourceSchema.Attribute{
                    "connection_service_name": resourceSchema.StringAttribute{ 
                        Optional:    true,
                        Description: "The name of your connection service.",
                    },
                    "account_url": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The URL of your account.",
                    },
                    "vpce_id": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your Virtual Private Cloud Endpoint.",
                    },
                    "aws_account_id": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your AWS account.",
                    },
                    "cluster_identifier": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The cluster identifier.",
                    },
                    "connection_service_id": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your connection service.",
                    },
                    "workspace_url": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The URL of your workspace.",
                    },
                    "pls_id": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your Azure Private Link service.",
                    },
                    "sub_resource_name": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The name of subresource.",
                    },
                    "private_dns_regions": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "Private DNS Regions.",
                    },
                    "private_connection_service_id": resourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your connection service.",
                    },
                },
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
        },
        Blocks: map[string]datasourceSchema.Block{
            "config": datasourceSchema.SingleNestedBlock{
                Attributes: map[string]datasourceSchema.Attribute{
                    "connection_service_name": datasourceSchema.StringAttribute{ 
                        Optional:    true,
                        Description: "The name of your connection service.",
                    },
                    "account_url": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The URL of your account.",
                    },
                    "vpce_id": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your Virtual Private Cloud Endpoint.",
                    },
                    "aws_account_id": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your AWS account.",
                    },
                    "cluster_identifier": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The cluster identifier.",
                    },
                    "connection_service_id": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your connection service.",
                    },
                    "workspace_url": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The URL of your workspace.",
                    },
                    "pls_id": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your Azure Private Link service.",
                    },
                    "sub_resource_name": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The name of subresource.",
                    },
                    "private_dns_regions": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "Private DNS Regions.",
                    },
                    "private_connection_service_id": datasourceSchema.StringAttribute{
                        Optional:    true,
                        Description: "The ID of your connection service.",
                    },
                },
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
                    },
                },
            },
        },
    }
}
