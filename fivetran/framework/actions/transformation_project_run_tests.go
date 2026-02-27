package actions

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TransformationProjectRunTests() action.Action {
	return &transformationProjectRunTests{}
}

type transformationProjectRunTests struct {
	core.ProviderAction
}

type transformationProjectRunTestsConfig struct {
	ProjectId   types.String `tfsdk:"project_id"`
	FailOnTestsFailure types.Bool `tfsdk:"fail_on_tests_failure"`
}

var _ action.ActionWithConfigure = &transformationProjectRunTests{}

func (a *transformationProjectRunTests) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transformation_project_run_tests"
}

func (a *transformationProjectRunTests) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		Description: "Triggers setup tests for a Fivetran transformation project.",
		Attributes: map[string]actionschema.Attribute{
			"project_id": actionschema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the transformation project to test.",
			},
			"fail_on_tests_failure": actionschema.BoolAttribute{
				Optional:    true,
				Description: "If true, the action will produce an error diagnostic when tests result in ERROR status, preventing further plan execution. Defaults to true.",
			},
		},
	}
}

func (a *transformationProjectRunTests) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	if a.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var config transformationProjectRunTestsConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := config.ProjectId.ValueString()
	failOnError := config.FailOnTestsFailure.IsNull() || config.FailOnTestsFailure.IsUnknown() || config.FailOnTestsFailure.ValueBool()

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Running setup tests for transformation project %q...", projectId),
	})

	svc := &transformations.TransformationProjectTestsService{HttpService: a.GetClient().NewHttpService()}
	svc.ExternalLoggingId(projectId)

	testResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Run Transformation Project Tests",
			fmt.Sprintf("%v; code: %v; message: %v", err, testResponse.Code, testResponse.Message),
		)
		return
	}

	status := testResponse.Data.Status

	// Report individual setup test results as progress
	for _, st := range testResponse.Data.SetupTests {
		msg := fmt.Sprintf("Test %q: %s", st.Title, st.Status)
		if st.Message != "" {
			msg += " - " + st.Message
		}
		resp.SendProgress(action.InvokeProgressEvent{Message: msg})
	}

	if status == "ERROR" {
		// Build detail from errors and failed setup tests
		detail := fmt.Sprintf("Transformation project %q tests completed with status: %s.", projectId, status)
		if len(testResponse.Data.Errors) > 0 {
			detail += "\nErrors:"
			for _, e := range testResponse.Data.Errors {
				detail += "\n  - " + e
			}
		}
		for _, st := range testResponse.Data.SetupTests {
			if st.Status == "FAILED" {
				detail += fmt.Sprintf("\n  - Test %q FAILED: %s", st.Title, st.Message)
			}
		}

		if failOnError {
			resp.Diagnostics.AddError("Transformation Project Tests Failed", detail)
		} else {
			resp.Diagnostics.AddWarning("Transformation Project Tests Failed", detail)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Transformation project %q tests completed with status: %s", projectId, status),
	})
}
