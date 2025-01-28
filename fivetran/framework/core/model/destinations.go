package model

import (
    "context"

    "github.com/fivetran/go-fivetran/destinations"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Destinations struct {
    Id              types.String `tfsdk:"id"` 
    Destinations    types.Set    `tfsdk:"destinations"`
}

func (d *Destinations) ReadFromResponse(ctx context.Context, resp destinations.DestinationsListResponse) {
    elementAttrType := map[string]attr.Type{
        "id":                           types.StringType,
        "group_id":                     types.StringType,
        "service":                      types.StringType,
        "region":                       types.StringType,
        "time_zone_offset":             types.StringType,
        "setup_status":                 types.StringType,
        "networking_method":            types.StringType,
        "hybrid_deployment_agent_id":   types.StringType,
        "private_link_id":              types.StringType,
        "daylight_saving_time_enabled": types.BoolType,
    }

    if resp.Data.Items == nil {
        d.Destinations = types.SetNull(types.ObjectType{AttrTypes: elementAttrType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.ID)
        item["group_id"] = types.StringValue(v.GroupID)
        item["service"] = types.StringValue(v.Service)
        item["region"] = types.StringValue(v.Region)
        item["time_zone_offset"] = types.StringValue(v.TimeZoneOffset)
        item["setup_status"] = types.StringValue(v.SetupStatus)
        item["private_link_id"] = types.StringValue(v.PrivateLinkId)
        item["hybrid_deployment_agent_id"] = types.StringValue(v.HybridDeploymentAgentId)
        item["networking_method"] = types.StringValue(v.NetworkingMethod)
        item["daylight_saving_time_enabled"] = types.BoolValue(v.DaylightSavingTimeEnabled)

        objectValue, _ := types.ObjectValue(elementAttrType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Destinations, _ = types.SetValue(types.ObjectType{AttrTypes: elementAttrType}, items)
}