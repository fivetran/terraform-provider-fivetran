package actions

import (
	"context"
	"net/http"
	"testing"

	fivetranSdk "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestTransformationProjectRunTests_Metadata(t *testing.T) {
	a := &transformationProjectRunTests{}
	req := action.MetadataRequest{ProviderTypeName: "fivetran"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)

	if resp.TypeName != "fivetran_transformation_project_run_tests" {
		t.Errorf("expected type name %q, got %q", "fivetran_transformation_project_run_tests", resp.TypeName)
	}
}

func TestTransformationProjectRunTests_Schema(t *testing.T) {
	a := &transformationProjectRunTests{}
	req := action.SchemaRequest{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema errors: %v", resp.Diagnostics)
	}

	projectIdAttr, ok := resp.Schema.Attributes["project_id"]
	if !ok {
		t.Fatal("expected project_id attribute in schema")
	}
	strAttr, ok := projectIdAttr.(actionschema.StringAttribute)
	if !ok {
		t.Fatal("expected project_id to be StringAttribute")
	}
	if !strAttr.Required {
		t.Error("expected project_id to be required")
	}

	failOnErrorAttr, ok := resp.Schema.Attributes["fail_on_tests_failure"]
	if !ok {
		t.Fatal("expected fail_on_tests_failure attribute in schema")
	}
	boolAttr, ok := failOnErrorAttr.(actionschema.BoolAttribute)
	if !ok {
		t.Fatal("expected fail_on_tests_failure to be BoolAttribute")
	}
	if !boolAttr.Optional {
		t.Error("expected fail_on_tests_failure to be optional")
	}
}

func TestTransformationProjectRunTests_Invoke_StatusReady(t *testing.T) {
	mockClient := mock.NewHttpClient()

	handler := mockClient.When(http.MethodPost, "/v1/transformation-projects/test_project_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, 200, `{
				"code": "Success",
				"message": "Tests completed",
				"data": {
					"id": "test_project_id",
					"status": "READY",
					"setup_tests": [
						{
							"title": "Validate Git Connection",
							"status": "PASSED",
							"message": "Git connection successful"
						},
						{
							"title": "Validate dbt Project",
							"status": "PASSED",
							"message": "dbt project is valid"
						}
					]
				}
			}`), nil
		},
	)

	a := configureAction(t, mockClient)

	var progressMessages []string
	invokeResp := invokeAction(t, a, "test_project_id", nil, &progressMessages)

	if invokeResp.Diagnostics.HasError() {
		t.Fatalf("unexpected invoke errors: %v", invokeResp.Diagnostics)
	}

	if handler.Interactions != 1 {
		t.Errorf("expected 1 API call, got %d", handler.Interactions)
	}

	if len(progressMessages) < 3 {
		t.Errorf("expected at least 3 progress messages (start + 2 tests + completion), got %d", len(progressMessages))
	}
}

func TestTransformationProjectRunTests_Invoke_StatusError_FailOnTestsFailureTrue(t *testing.T) {
	mockClient := mock.NewHttpClient()

	mockClient.When(http.MethodPost, "/v1/transformation-projects/test_project_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, 200, `{
				"code": "Success",
				"data": {
					"id": "test_project_id",
					"status": "ERROR",
					"setup_tests": [
						{
							"title": "Validate Git Connection",
							"status": "PASSED",
							"message": "Git connection successful"
						},
						{
							"title": "Validate dbt Project",
							"status": "FAILED",
							"message": "dbt project compilation failed"
						}
					],
					"errors": ["Git repository not accessible"]
				}
			}`), nil
		},
	)

	a := configureAction(t, mockClient)

	// fail_on_tests_failure = true (explicit)
	failOnError := true
	invokeResp := invokeAction(t, a, "test_project_id", &failOnError, nil)

	if !invokeResp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics for ERROR status with fail_on_tests_failure=true")
	}

	// Verify the error contains meaningful information
	found := false
	for _, d := range invokeResp.Diagnostics.Errors() {
		if d.Summary() == "Transformation Project Tests Failed" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Transformation Project Tests Failed' error diagnostic")
	}
}

func TestTransformationProjectRunTests_Invoke_StatusError_FailOnTestsFailureDefault(t *testing.T) {
	mockClient := mock.NewHttpClient()

	mockClient.When(http.MethodPost, "/v1/transformation-projects/test_project_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, 200, `{
				"code": "Success",
				"data": {
					"id": "test_project_id",
					"status": "ERROR",
					"setup_tests": [
						{
							"title": "Validate dbt Project",
							"status": "FAILED",
							"message": "compilation error"
						}
					],
					"errors": ["compilation error"]
				}
			}`), nil
		},
	)

	a := configureAction(t, mockClient)

	// fail_on_tests_failure not set (nil) -> defaults to true
	invokeResp := invokeAction(t, a, "test_project_id", nil, nil)

	if !invokeResp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics when fail_on_tests_failure defaults to true")
	}
}

func TestTransformationProjectRunTests_Invoke_StatusError_FailOnTestsFailureFalse(t *testing.T) {
	mockClient := mock.NewHttpClient()

	mockClient.When(http.MethodPost, "/v1/transformation-projects/test_project_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, 200, `{
				"code": "Success",
				"data": {
					"id": "test_project_id",
					"status": "ERROR",
					"setup_tests": [
						{
							"title": "Validate dbt Project",
							"status": "FAILED",
							"message": "compilation error"
						}
					],
					"errors": ["compilation error"]
				}
			}`), nil
		},
	)

	a := configureAction(t, mockClient)

	// fail_on_tests_failure = false -> should produce warning, not error
	failOnError := false
	invokeResp := invokeAction(t, a, "test_project_id", &failOnError, nil)

	if invokeResp.Diagnostics.HasError() {
		t.Fatalf("expected no error diagnostics with fail_on_tests_failure=false, got: %v", invokeResp.Diagnostics)
	}

	if len(invokeResp.Diagnostics.Warnings()) == 0 {
		t.Error("expected warning diagnostics for ERROR status with fail_on_tests_failure=false")
	}

	found := false
	for _, d := range invokeResp.Diagnostics.Warnings() {
		if d.Summary() == "Transformation Project Tests Failed" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Transformation Project Tests Failed' warning diagnostic")
	}
}

func TestTransformationProjectRunTests_Invoke_ApiError(t *testing.T) {
	mockClient := mock.NewHttpClient()

	mockClient.When(http.MethodPost, "/v1/transformation-projects/bad_id/test").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return mock.NewResponse(req, 404, `{
				"code": "NotFound",
				"message": "Transformation project not found"
			}`), nil
		},
	)

	a := configureAction(t, mockClient)
	invokeResp := invokeAction(t, a, "bad_id", nil, nil)

	if !invokeResp.Diagnostics.HasError() {
		t.Fatal("expected error diagnostics for API error")
	}
}

// -- helpers --

func configureAction(t *testing.T, mockClient *mock.HttpClient) *transformationProjectRunTests {
	t.Helper()

	client := fivetranSdk.New("test_key", "test_secret")
	client.BaseURL("https://api.fivetran.com/v1")
	client.SetHttpClient(mockClient)

	a := &transformationProjectRunTests{}
	configureReq := action.ConfigureRequest{ProviderData: client}
	configureResp := &action.ConfigureResponse{}
	a.Configure(context.Background(), configureReq, configureResp)

	if configureResp.Diagnostics.HasError() {
		t.Fatalf("unexpected configure errors: %v", configureResp.Diagnostics)
	}
	return a
}

func invokeAction(t *testing.T, a *transformationProjectRunTests, projectId string, failOnError *bool, progressMessages *[]string) *action.InvokeResponse {
	t.Helper()

	invokeReq := action.InvokeRequest{
		Config: buildTestConfig(t, projectId, failOnError),
	}

	if progressMessages == nil {
		msgs := []string{}
		progressMessages = &msgs
	}

	invokeResp := &action.InvokeResponse{
		SendProgress: func(event action.InvokeProgressEvent) {
			*progressMessages = append(*progressMessages, event.Message)
		},
	}

	a.Invoke(context.Background(), invokeReq, invokeResp)
	return invokeResp
}

func buildTestConfig(t *testing.T, projectId string, failOnError *bool) tfsdk.Config {
	t.Helper()

	schemaReq := action.SchemaRequest{}
	schemaResp := &action.SchemaResponse{}
	a := &transformationProjectRunTests{}
	a.Schema(context.Background(), schemaReq, schemaResp)

	configType := schemaResp.Schema.Type().TerraformType(context.Background())

	var failOnErrorValue tftypes.Value
	if failOnError == nil {
		failOnErrorValue = tftypes.NewValue(tftypes.Bool, nil) // null -> use default
	} else {
		failOnErrorValue = tftypes.NewValue(tftypes.Bool, *failOnError)
	}

	rawValue := tftypes.NewValue(configType, map[string]tftypes.Value{
		"project_id":    tftypes.NewValue(tftypes.String, projectId),
		"fail_on_tests_failure": failOnErrorValue,
	})

	return tfsdk.Config{
		Raw:    rawValue,
		Schema: schemaResp.Schema,
	}
}
