package schema

import (
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func GroupConnectionsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Required:      true,
                Description:   "The ID of this resource.",
            },
            "schema": datasourceSchema.StringAttribute{
                Optional:    true,
                Description: "Optional filter. When defined, the data source will only contain information for the connection with the specified schema name.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "connectors": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the connection within the Fivetran system.",
                        },
                        "group_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the Group within the Fivetran system.",
                        },
                        "schema": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination",
                        },
                        "service": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connector type name within the Fivetran system",
                        },
                        "service_version": datasourceSchema.Int64Attribute{
                            Computed:    true,
                            Description: "The connector type version within the Fivetran system",
                        },
                        "connected_by": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier of the user who has created the connection in your account",
                        },
                        "created_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The timestamp of the time the connection was created in your account",
                        },
                        "succeeded_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The timestamp of the time the connection sync succeeded last time",
                        },
                        "failed_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The timestamp of the time the connection sync failed last time",
                        },
                        "sync_frequency": datasourceSchema.Int64Attribute{
                            Computed:    true,
                            Description: "The connection sync frequency in minutes",
                        },
                        "schedule_type": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connection schedule configuration type. Supported values: auto, manual",
                        },
                        "daily_sync_time": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time",
                        },
                    },
                    Blocks: map[string]datasourceSchema.Block{
                        "status": datasourceSchema.SingleNestedBlock{
                           Attributes: map[string]datasourceSchema.Attribute{
                                "setup_state": datasourceSchema.StringAttribute{
                                    Computed:    true,
                                    Description: "The current setup state of the connection. The available values are: <br /> - incomplete - the setup config is incomplete, the setup tests never succeeded <br /> - connected - the connection is properly set up <br /> - broken - the connection setup config is broken.",
                                },
                                "is_historical_sync": datasourceSchema.BoolAttribute{
                                    Computed:    true,
                                    Description: "The boolean specifying whether the connection should be triggered to re-sync all historical data. If you set this parameter to TRUE, the next scheduled sync will be historical. If the value is FALSE or not specified, the connection will not re-sync historical data. NOTE: When the value is TRUE, only the next scheduled sync will be historical, all subsequent ones will be incremental. This parameter is set to FALSE once the historical sync is completed.",
                                },
                                "sync_state": datasourceSchema.StringAttribute{
                                    Computed:    true,
                                    Description: "The current sync state of the connection. The available values are: <br /> - scheduled - the sync is waiting to be run <br /> - syncing - the sync is currently running <br /> - paused - the sync is currently paused <br /> - rescheduled - the sync is waiting until more API calls are available in the source service.",
                                },
                                "update_state": datasourceSchema.StringAttribute{
                                    Computed:    true,
                                    Description: "The current data update state of the connection. The available values are: <br /> - on_schedule - the sync is running smoothly, no delays <br /> - delayed - the data is delayed for a longer time than expected for the update.",
                                },
                            },
                            Blocks: map[string]datasourceSchema.Block{
                                "tasks": datasourceSchema.SetNestedBlock{
                                    Description: "The collection of tasks for the connection",
                                    NestedObject: datasourceSchema.NestedBlockObject{
                                        Attributes: map[string]datasourceSchema.Attribute{
                                            "code": datasourceSchema.StringAttribute{
                                                Computed:    true,
                                                Description: "Response status code",
                                            },
                                            "message": datasourceSchema.StringAttribute{
                                                Computed:    true,
                                                Description: "Response status text",
                                            },
                                        },
                                    },
                                },
                                "warnings": datasourceSchema.SetNestedBlock{
                                    Description: "The collection of warnings for the connection.",
                                    NestedObject: datasourceSchema.NestedBlockObject{
                                        Attributes: map[string]datasourceSchema.Attribute{
                                            "code": datasourceSchema.StringAttribute{
                                                Computed:    true,
                                                Description: "Response status code",
                                            },
                                            "message": datasourceSchema.StringAttribute{
                                                Computed:    true,
                                                Description: "Response status text",
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}