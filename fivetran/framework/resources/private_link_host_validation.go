package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// rejectHostWithPrivateLink adds a plan-time error when both `private_link_id` and an
// explicit `config.host` are configured together for services where Fivetran derives the host.
// Some services (aurora, magento_mysql_rds, maria_rds, mysql_rds) with AWS_IAM authentication
// require the user to explicitly set `config.host` to the PrivateLink DNS address, so we only
// reject for other service/auth combinations where Fivetran actually derives the host.
func rejectHostWithPrivateLink(configAttr types.Object, privateLinkId types.String, service types.String, auth types.Object, configRoot path.Path, diags *diag.Diagnostics) {
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

	// Allow host when service is one that requires explicit host for AWS_IAM + PrivateLink
	if shouldAllowHostWithPrivateLink(service, auth) {
		return
	}

	diags.AddAttributeError(
		configRoot.AtName("host"),
		"`host` Cannot Be Set Together With `private_link_id`",
		"Fivetran automatically derives `host` from the private link referenced by `private_link_id`. "+
			"Remove `host` from `config` and let Fivetran compute it, or remove `private_link_id` if you want to manage `host` directly.",
	)
}

// shouldAllowHostWithPrivateLink returns true for services that require explicit host configuration
// when using AWS_IAM authentication with PrivateLink.
func shouldAllowHostWithPrivateLink(service types.String, auth types.Object) bool {
	if service.IsNull() || service.IsUnknown() || auth.IsNull() || auth.IsUnknown() {
		return false
	}

	serviceName := service.ValueString()
	// Services that require explicit host with AWS_IAM + PrivateLink
	iamRequiredService := serviceName == "aurora" || serviceName == "magento_mysql_rds" ||
		serviceName == "maria_rds" || serviceName == "mysql_rds"

	if !iamRequiredService {
		return false
	}

	// Check if auth type is AWS_IAM
	authAttrs := auth.Attributes()
	if authType, ok := authAttrs["auth_type"].(types.String); ok && !authType.IsNull() {
		return authType.ValueString() == "AWS_IAM"
	}

	return false
}
