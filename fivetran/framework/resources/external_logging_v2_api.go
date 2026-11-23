package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/common"
	externallogging "github.com/fivetran/go-fivetran/external_logging"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// The group-level endpoints (/external-logging/{groupId}) and the account-level ones
// (/external-logging/account) return the same response types, so both levels share one code path in
// the resource, and only these helpers pick the endpoint.

// externalLoggingV2Target addresses one log service: the one of the destination (group) groupId, or
// the account-level one when groupId is empty. The account-level endpoints have no ID in the path.
type externalLoggingV2Target struct {
	groupId string
}

// externalLoggingV2TargetOf addresses the log service a resource manages. group_id alone decides
// the level, as everywhere in the resource; for a destination log service it equals id.
func externalLoggingV2TargetOf(data model.ExternalLoggingV2ResourceModel) externalLoggingV2Target {
	if data.IsAccountLevel() {
		return externalLoggingV2Target{}
	}
	return externalLoggingV2Target{groupId: data.GroupId.ValueString()}
}

func (t externalLoggingV2Target) accountLevel() bool {
	return t.groupId == ""
}

func (r *externalLoggingV2) readExternalLogging(ctx context.Context, target externalLoggingV2Target) (externallogging.ExternalLoggingCustomResponse, error) {
	if target.accountLevel() {
		return r.GetClient().NewExternalLoggingAccountDetails().
			DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	}
	return r.GetClient().NewExternalLoggingDetails().
		ExternalLoggingId(target.groupId).
		DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
}

// createExternalLogging creates a log service without running the setup tests: the group-level
// create endpoint doesn't take run_setup_tests, so both levels run them through the dedicated
// endpoint once the service is saved in state.
func (r *externalLoggingV2) createExternalLogging(ctx context.Context, target externalLoggingV2Target, service string, enabled bool, config map[string]interface{}) (externallogging.ExternalLoggingCustomResponse, error) {
	if target.accountLevel() {
		return r.GetClient().NewExternalLoggingAccountCreate().
			Service(service).
			Enabled(enabled).
			ConfigCustom(&config).
			DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	}
	return r.GetClient().NewExternalLoggingCreate().
		GroupId(target.groupId).
		Service(service).
		Enabled(enabled).
		ConfigCustom(&config).
		DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
}

// updateExternalLogging sends a PATCH. enabled and config are left out of the request when nil.
func (r *externalLoggingV2) updateExternalLogging(ctx context.Context, target externalLoggingV2Target, enabled *bool, runSetupTests bool, config map[string]interface{}) (externallogging.ExternalLoggingCustomResponse, error) {
	if target.accountLevel() {
		updateSvc := r.GetClient().NewExternalLoggingAccountUpdate().RunSetupTests(runSetupTests)
		if enabled != nil {
			updateSvc.Enabled(*enabled)
		}
		if config != nil {
			updateSvc.ConfigCustom(&config)
		}
		return updateSvc.DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	}
	updateSvc := r.GetClient().NewExternalLoggingUpdate().
		ExternalLoggingId(target.groupId).
		RunSetupTests(runSetupTests)
	if enabled != nil {
		updateSvc.Enabled(*enabled)
	}
	if config != nil {
		updateSvc.ConfigCustom(&config)
	}
	return updateSvc.DoCustomWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
}

func (r *externalLoggingV2) deleteExternalLogging(ctx context.Context, target externalLoggingV2Target) (common.CommonResponse, error) {
	if target.accountLevel() {
		return r.GetClient().NewExternalLoggingAccountDelete().
			DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	}
	return r.GetClient().NewExternalLoggingDelete().
		ExternalLoggingId(target.groupId).
		DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
}

func (r *externalLoggingV2) runSetupTests(ctx context.Context, target externalLoggingV2Target, diags *diag.Diagnostics) {
	var (
		response externallogging.ExternalLoggingSetupTestsResponse
		err      error
	)
	if target.accountLevel() {
		response, err = r.GetClient().NewExternalLoggingAccountSetupTests().
			DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	} else {
		response, err = r.GetClient().NewExternalLoggingSetupTests().
			ExternalLoggingId(target.groupId).
			DoWithUserAgentSuffix(ctx, externalLoggingV2UserAgentSuffix)
	}
	if err != nil {
		diags.AddError(
			"Unable to Run External Logging Setup Tests.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}
	warnFailedLogSetupTests(response.Data.SetupTests, diags)
}
