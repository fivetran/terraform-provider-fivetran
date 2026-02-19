package mock

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/actions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var (
	schemaReloadHandler *mock.Handler
)

func TestActionConnectionSchemaReloadBasic(t *testing.T) {
	mockClient.Reset()

	schemaReloadHandler = mockClient.When(http.MethodPost, "/v1/connections/test_connection_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			assertKeyExistsAndHasValue(t, body, "exclude_mode", "PRESERVE")
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
				"schema_change_handling": "ALLOW_ALL",
				"schemas": map[string]any{
					"schema_1": map[string]any{
						"name_in_destination": "schema_1",
						"enabled":             true,
					},
				},
			}), nil
		},
	)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "test_connection_id"
  }
}`,
				PostApplyFunc: func() {
					assertEqual(t, schemaReloadHandler.Interactions, 1)
				},
			},
		},
	})
}

func TestActionConnectionSchemaReloadWithExcludeMode(t *testing.T) {
	mockClient.Reset()

	schemaReloadHandler = mockClient.When(http.MethodPost, "/v1/connections/test_connection_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			assertKeyExistsAndHasValue(t, body, "exclude_mode", "EXCLUDE")
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
				"schema_change_handling": "ALLOW_ALL",
				"schemas":               map[string]any{},
			}), nil
		},
	)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "test_connection_id"
    exclude_mode  = "EXCLUDE"
  }
}`,
				PostApplyFunc: func() {
					assertEqual(t, schemaReloadHandler.Interactions, 1)
				},
			},
		},
	})
}

func TestActionConnectionSchemaReloadWithTimeout(t *testing.T) {
	mockClient.Reset()

	schemaReloadHandler = mockClient.When(http.MethodPost, "/v1/connections/test_connection_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
				"schema_change_handling": "ALLOW_ALL",
				"schemas":               map[string]any{},
			}), nil
		},
	)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "test_connection_id"

    timeouts = {
      invoke = "30m"
    }
  }
}`,
				PostApplyFunc: func() {
					assertEqual(t, schemaReloadHandler.Interactions, 1)
				},
			},
		},
	})
}

func TestActionConnectionSchemaReloadNotFoundError(t *testing.T) {
	mockClient.Reset()

	mockClient.When(http.MethodPost, "/v1/connections/bad_connection_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranResponse(t, req, "NotFound", http.StatusNotFound, "Connection not found", nil), nil
		},
	)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "bad_connection_id"
  }
}`,
				ExpectError: regexp.MustCompile("Unable to Reload Connection Schema"),
			},
		},
	})
}

func TestActionConnectionSchemaReloadTimeoutExceedsMax(t *testing.T) {
	mockClient.Reset()

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "test_connection_id"

    timeouts = {
      invoke = "2h"
    }
  }
}`,
				ExpectError: regexp.MustCompile("Invalid Timeout"),
			},
		},
	})
}

// timeoutError implements net.Error with Timeout() == true to simulate a Go-level network timeout.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "connection timed out" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestActionConnectionSchemaReloadTimeoutThenPollSuccess(t *testing.T) {
	// Use a short poll interval so the test doesn't wait 30s per cycle.
	actions.PollIntervalOverride = 100 * time.Millisecond
	t.Cleanup(func() { actions.PollIntervalOverride = 0 })

	mockClient.Reset()

	// The reload endpoint returns a Go-level timeout error (net.Error with Timeout()==true).
	var reloadHandler *mock.Handler
	reloadHandler = mockClient.When(http.MethodPost, "/v1/connections/test_connection_id/schemas/reload").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return nil, &timeoutError{}
		},
	)

	// The details endpoint returns success on first poll — schema is available.
	var detailsHandler *mock.Handler
	detailsHandler = mockClient.When(http.MethodGet, "/v1/connections/test_connection_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]any{
				"schema_change_handling": "ALLOW_ALL",
				"schemas": map[string]any{
					"schema_1": map[string]any{
						"name_in_destination": "schema_1",
						"enabled":             true,
					},
				},
			}), nil
		},
	)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.fivetran_connection_schema_reload.test]
    }
  }
}

action "fivetran_connection_schema_reload" "test" {
  provider = fivetran-provider
  config {
    connection_id = "test_connection_id"

    timeouts = {
      invoke = "5m"
    }
  }
}`,
				PostApplyFunc: func() {
					// Reload was attempted once and timed out
					assertEqual(t, reloadHandler.Interactions, 1)
					// Polling fell through to the details endpoint and succeeded
					assertEqual(t, detailsHandler.Interactions, 1)
				},
			},
		},
	})
}
