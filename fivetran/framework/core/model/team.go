package model

import (
    "context"

    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type Team struct {
    Id              types.String `tfsdk:"id"`
    Name            types.String `tfsdk:"name"`
    Description     types.String `tfsdk:"description"`
    Role            types.String `tfsdk:"role"`
}

func (d *Team) ReadFromResponse(ctx context.Context, resp teams.TeamsDetailsResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}

func (d *Team) ReadFromCreateResponse(ctx context.Context, resp teams.TeamsCreateResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}

func (d *Team) ReadFromUpdateResponse(ctx context.Context, resp teams.TeamsUpdateResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}