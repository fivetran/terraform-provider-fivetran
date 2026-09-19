package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// rejectHostWithPrivateLink adds a plan-time error when both `private_link_id` and an
// explicit `config.host` are configured together. Fivetran derives `host` from the private
// link referenced by `private_link_id` server-side, so a user-supplied value can silently
// disagree with what the API resolves - which only surfaces later as "Provider produced
// inconsistent result after apply". Terraform's plan-consistency rules never allow a provider
// to override an explicitly configured, known value to unknown, so the only safe fix is to
// reject the conflicting configuration up front, at plan time, rather than let it fail apply.
func rejectHostWithPrivateLink(configAttr types.Object, privateLinkId types.String, configRoot path.Path, diags *diag.Diagnostics) {
	if privateLinkId.IsNull() || privateLinkId.IsUnknown() || privateLinkId.ValueString() == "" {
		return
	}
	if configAttr.IsNull() || configAttr.IsUnknown() {
		return
	}

	hostVal, ok := configAttr.Attributes()["host"].(types.String)
	if !ok || hostVal.IsNull() || hostVal.IsUnknown() {
		return
	}

	diags.AddAttributeError(
		configRoot.AtName("host"),
		"`host` Cannot Be Set Together With `private_link_id`",
		"Fivetran automatically derives `host` from the private link referenced by `private_link_id`. "+
			"Remove `host` from `config` and let Fivetran compute it, or remove `private_link_id` if you want to manage `host` directly.",
	)
}
