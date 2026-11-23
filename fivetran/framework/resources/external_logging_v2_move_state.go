package resources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const accountExternalLoggingV2TypeName = "fivetran_account_external_logging_v2"

var _ resource.ResourceWithMoveState = &externalLoggingV2{}

// MoveState lets a `moved` block (Terraform 1.8 or later) move a fivetran_account_external_logging_v2
// resource to fivetran_external_logging_v2 without recreating the log service. Both manage the same
// API object with the same ID, and the account-level form of this resource has no group_id.
func (r *externalLoggingV2) MoveState(ctx context.Context) []resource.StateMover {
	sourceSchema := fivetranSchema.AccountExternalLoggingV2ResourceSchema()
	return []resource.StateMover{
		{
			SourceSchema: &sourceSchema,
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				// Other source types are left to other movers; Terraform reports that none applied.
				if req.SourceTypeName != accountExternalLoggingV2TypeName || req.SourceState == nil {
					return
				}

				var source model.AccountExternalLoggingV2ResourceModel
				resp.Diagnostics.Append(req.SourceState.Get(ctx, &source)...)
				if resp.Diagnostics.HasError() {
					return
				}

				target := model.ExternalLoggingV2ResourceModel{
					Id:                   source.Id,
					GroupId:              types.StringNull(),
					Service:              source.Service,
					Enabled:              source.Enabled,
					RunSetupTests:        source.RunSetupTests,
					Config:               source.Config,
					ConfigSecrets:        types.DynamicNull(),
					ConfigSecretsVersion: types.Int64Null(),
				}
				resp.Diagnostics.Append(resp.TargetState.Set(ctx, &target)...)
				// Keeps the list of imported config fields, so fields left out of config are still
				// not cleared on the first apply after the move.
				resp.TargetPrivate = req.SourcePrivate
			},
		},
	}
}
