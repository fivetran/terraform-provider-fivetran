package resources_test

import (
	"fmt"
	"net/http"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// One released version for each fivetran_destination state schema version it writes.
var destinationReleasedSchemaVersions = []string{
	"1.1.5", // schema version 0: config stored as a set, booleans as strings
	"1.2.9", // schema version 1: config stored as a block, booleans as bools
}

type destinationUpgradeMock struct {
	config      map[string]interface{}
	postBodies  []map[string]interface{}
	patchBodies []map[string]interface{}
}

func (m *destinationUpgradeMock) setup(t *testing.T) {
	tfmock.MockClient().Reset()
	m.config = nil
	m.postBodies = nil
	m.patchBodies = nil

	response := func() map[string]interface{} {
		return map[string]interface{}{
			"id":                           "destination_id",
			"group_id":                     "destination_id",
			"service":                      "big_query",
			"region":                       "GCP_US_EAST4",
			"time_zone_offset":             "0",
			"setup_status":                 "connected",
			"daylight_saving_time_enabled": true,
			"networking_method":            "Directly",
			"config":                       m.config,
		}
	}

	tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			m.postBodies = append(m.postBodies, body)
			m.config, _ = body["config"].(map[string]interface{})
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", response()), nil
		},
	)
	tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", response()), nil
		},
	)
	tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			m.patchBodies = append(m.patchBodies, tfmock.RequestBodyToJson(t, req))
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", response()), nil
		},
	)
	tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

func destinationUpgradeTfConfig(apiURL, releasedVersion string, current bool) string {
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
	// 1.1.5 has no daylight_saving_time_enabled. The current provider reads it from the API.
	daylightSaving := "daylight_saving_time_enabled = true"
	if releasedVersion == "1.1.5" {
		daylightSaving = ""
	}
	return header + fmt.Sprintf(`
		resource "fivetran_destination" "test" {
			group_id                     = "destination_id"
			service                      = "big_query"
			time_zone_offset             = "0"
			region                       = "GCP_US_EAST4"
			%s
			run_setup_tests              = false
			trust_certificates           = true
			trust_fingerprints           = true

			config {
				project_id        = "project"
				data_set_location = "US"
			}
		}`, daylightSaving)
}

// State written by every released state schema version must upgrade to the current schema
// without an error.
func TestDestinationProviderUpgradeMock(t *testing.T) {
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "fivetran")

	for _, releasedVersion := range destinationReleasedSchemaVersions {
		t.Run(releasedVersion, func(t *testing.T) {
			m := &destinationUpgradeMock{}
			const name = "fivetran_destination.test"
			server := tfmock.NewMockServer(t)

			currentConfig := destinationUpgradeTfConfig(server.URL, releasedVersion, true)
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
						Config: destinationUpgradeTfConfig(server.URL, releasedVersion, false),
						Check:  resource.TestCheckResourceAttr(name, "id", "destination_id"),
					},
					currentStep(resource.TestStep{
						Config: currentConfig,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "id", "destination_id"),
							resource.TestCheckResourceAttr(name, "config.project_id", "project"),
							resource.TestCheckResourceAttr(name, "run_setup_tests", "false"),
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
