package model

import (
    "context"

    "github.com/fivetran/go-fivetran/webhooks"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type Webhook struct {
    Id         types.String `tfsdk:"id"`
    Type       types.String `tfsdk:"type"`
    GroupId    types.String `tfsdk:"group_id"`
    Url        types.String `tfsdk:"url"`
    Active     types.Bool   `tfsdk:"active"`
    CreatedBy  types.String `tfsdk:"created_by"`
    CreatedAt  types.String `tfsdk:"created_at"`
    RunTests   types.Bool   `tfsdk:"run_tests"`

    Events     types.Set    `tfsdk:"events"`
    Secret     types.String `tfsdk:"secret"`
}

func (d *Webhook) ReadFromResponse(ctx context.Context, resp webhooks.WebhookResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Type = types.StringValue(resp.Data.Type)
    if resp.Data.GroupId == "" {
        d.GroupId = types.StringNull()
    } else {
        d.GroupId = types.StringValue(resp.Data.GroupId)
    }
    
    d.Url = types.StringValue(resp.Data.Url)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
    d.CreatedBy = types.StringValue(resp.Data.CreatedBy)
    d.Active = types.BoolValue(resp.Data.Active)

    d.Events, _ = types.SetValueFrom(ctx, types.StringType, resp.Data.Events)
}