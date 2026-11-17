package resources_test

import (
	"fmt"
	"net/http"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// One released version for each fivetran_connector state schema version it writes.
var connectorReleasedSchemaVersions = []string{
	"0.7.3", // schema version 1: config, auth and destination_schema stored as sets
	"1.0.0", // schema version 2: as version 1, servers stored as a set
	"1.1.5", // schema version 2: last release before the plugin framework
	"1.1.6", // schema version 3: config, auth and destination_schema stored as blocks
	"1.5.0", // schema version 4
}

var connectorUpgradeProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fivetran": tfmock.ProtoV6ProviderFactories["fivetran-provider"],
}

// connectorUpgradeMock stores the connector the released provider creates and serves it on both
// the /connectors paths released providers use and the /connections paths the current one uses.
type connectorUpgradeMock struct {
	config      map[string]interface{}
	postBodies  []map[string]interface{}
	patchBodies []map[string]interface{}
	deletes     int
}

func (m *connectorUpgradeMock) setup(t *testing.T) {
	tfmock.MockClient().Reset()
	m.config = nil
	m.postBodies = nil
	m.patchBodies = nil
	m.deletes = 0

	response := func() map[string]interface{} {
		return map[string]interface{}{
			"id":                     "connector_id",
			"group_id":               "group_id",
			"service":                "google_sheets",
			"service_version":        1,
			"schema":                 "gsheets.sheet",
			"connected_by":           "user_id",
			"created_at":             "2024-01-01T00:00:00.000000Z",
			"succeeded_at":           nil,
			"failed_at":              nil,
			"paused":                 true,
			"pause_after_trial":      false,
			"sync_frequency":         360,
			"schedule_type":          "auto",
			"networking_method":      "Directly",
			"data_delay_sensitivity": "NORMAL",
			"status": map[string]interface{}{
				"setup_state":        "connected",
				"sync_state":         "paused",
				"update_state":       "on_schedule",
				"is_historical_sync": false,
				"tasks":              []interface{}{},
				"warnings":           []interface{}{},
			},
			"config": m.config,
		}
	}

	for _, base := range []string{"/v1/connectors", "/v1/connections"} {
		tfmock.MockClient().When(http.MethodPost, base).ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				m.postBodies = append(m.postBodies, body)
				m.config, _ = body["config"].(map[string]interface{})
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", response()), nil
			},
		)
		tfmock.MockClient().When(http.MethodGet, base+"/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", response()), nil
			},
		)
		tfmock.MockClient().When(http.MethodPatch, base+"/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				m.patchBodies = append(m.patchBodies, body)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", response()), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, base+"/connector_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				m.deletes++
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
	}
}

func connectorUpgradeTfConfig(apiURL string, current bool) string {
	header := fmt.Sprintf(`
		provider "fivetran" {
			api_url    = "%s/v1"
			api_key    = "test_key"
			api_secret = "test_secret"
		}`, apiURL)
	if current {
		// The released provider gets its required_providers block from plugin-testing.
		header = `
		terraform {
			required_providers {
				fivetran = { source = "fivetran/fivetran" }
			}
		}` + header
	}
	return header + `
		resource "fivetran_connector" "test" {
			group_id           = "group_id"
			service            = "google_sheets"
			run_setup_tests    = false
			trust_certificates = true
			trust_fingerprints = true

			destination_schema {
				name  = "gsheets"
				table = "sheet"
			}

			config {
				sheet_id    = "sheet_id"
				named_range = "range"
			}

			auth {
				refresh_token = "refresh_token"
				client_access {
					client_id     = "client_id"
					client_secret = "client_secret"
				}
			}
		}`
}

// State written by every released state schema version must upgrade to the current schema
// without an error and without planned changes.
func TestConnectorProviderUpgradeMock(t *testing.T) {
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "fivetran")

	for _, releasedVersion := range connectorReleasedSchemaVersions {
		t.Run(releasedVersion, func(t *testing.T) {
			m := &connectorUpgradeMock{}
			const name = "fivetran_connector.test"
			server := tfmock.NewMockServer(t)

			currentConfig := connectorUpgradeTfConfig(server.URL, true)
			currentStep := func(step resource.TestStep) resource.TestStep {
				step.ProtoV6ProviderFactories = connectorUpgradeProviderFactories
				return step
			}

			resource.Test(t, resource.TestCase{
				PreCheck: func() { m.setup(t) },
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"fivetran": {Source: "fivetran/fivetran", VersionConstraint: releasedVersion},
						},
						Config: connectorUpgradeTfConfig(server.URL, false),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "id", "connector_id"),
						),
					},
					currentStep(resource.TestStep{
						Config: currentConfig,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "id", "connector_id"),
							resource.TestCheckResourceAttr(name, "destination_schema.name", "gsheets"),
							resource.TestCheckResourceAttr(name, "destination_schema.table", "sheet"),
							resource.TestCheckResourceAttr(name, "config.sheet_id", "sheet_id"),
							resource.TestCheckResourceAttr(name, "config.named_range", "range"),
							resource.TestCheckResourceAttr(name, "auth.refresh_token", "refresh_token"),
							resource.TestCheckResourceAttr(name, "auth.client_access.client_id", "client_id"),
							resource.TestCheckResourceAttr(name, "trust_certificates", "true"),
							resource.TestCheckResourceAttr(name, "trust_fingerprints", "true"),
						),
					}),
					currentStep(resource.TestStep{Config: currentConfig, PlanOnly: true}),
				},
			})

			tfmock.AssertEqual(t, len(m.postBodies), 1)
			tfmock.AssertEqual(t, len(m.patchBodies), 0)
		})
	}
}
