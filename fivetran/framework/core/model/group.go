package model

import (
    "context"
    "time"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type Group struct {
    Id           types.String `tfsdk:"id"`
    Name         types.String `tfsdk:"name"`
    CreatedAt    types.String `tfsdk:"created_at"`
    LastUpdated  types.String `tfsdk:"last_updated"`
}

func (d *Group) ReadFromResponse(ctx context.Context, resp groups.GroupDetailsResponse) {
    d.Id = types.StringValue(resp.Data.ID)
    d.Name = types.StringValue(resp.Data.Name)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt.String())
    d.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
}