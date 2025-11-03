package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	sdk "github.com/fivetran/go-fivetran/groups"
	fivetranUsers "github.com/fivetran/go-fivetran/users"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func GroupUsers() datasource.DataSource {
	return &groupUsers{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &groupUsers{}

type groupUsers struct {
	core.ProviderDatasource
}

func (d *groupUsers) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group_users"
}

func (d *groupUsers) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.GroupUsersDatasource()
}

func (d *groupUsers) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.GroupUsers
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.GroupListUsersResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.GroupListUsersResponse
		svc := d.GetClient().NewGroupListUsers().GroupID(data.Id.ValueString())
		
		svc.Limit(limit)
		if respNextCursor != "" {
			svc.Cursor(respNextCursor)
		}
		tmpResp, err =  svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Read error.",
				fmt.Sprintf("%v; code: %v", err, tmpResp.Code),
			)
			listResponse = sdk.GroupListUsersResponse{}
		}

		listResponse.Data.Items = append(listResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

    accountInfoSvc := d.GetClient().AccountInfo()
    accountInfoResp, err := accountInfoSvc.Do(ctx)
    if err == nil && accountInfoResp.Data.UserId != "" {
        tfUserId := accountInfoResp.Data.UserId

        var filteredItems []fivetranUsers.UserDetailsData
        for _, item := range listResponse.Data.Items {
            if item.ID != tfUserId {
                filteredItems = append(filteredItems, item)
            }
        }
        listResponse.Data.Items = filteredItems
    }

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}