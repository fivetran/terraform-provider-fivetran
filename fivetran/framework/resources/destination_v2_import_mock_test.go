package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A distinct service keeps the cached metadata separate from the other destination mocks.
const destinationV2ImportMockMetadata = `
{
	"id": "redshift_import_mock",
	"name": "Redshift",
	"type": "Warehouse",
	"config": {
		"type": "object",
		"properties": {
			"host":                {"type": "string", "nullable": false},
			"port":                {"type": "integer", "nullable": false},
			"user":                {"type": "string", "nullable": false},
			"database":            {"type": "string", "nullable": false},
			"connection_type":     {"type": "string", "nullable": false},
			"password":            {"type": "string", "format": "password", "nullable": true},
			"auth_type":           {"type": "string", "nullable": true},
			"s3_bucket_auth_type": {"type": "string", "nullable": true}
		}
	}
}`

func destinationV2ImportMockConfig(extra string) string {
	return `
		resource "fivetran_destination_v2" "dest" {
			provider         = fivetran-provider
			group_id         = "group_id"
			service          = "redshift_import_mock"
			time_zone_offset = "0"

			config = {
				host            = "cluster.redshift.amazonaws.com"
				port            = 5439
				user            = "fivetran_user"
				password        = "secret"
				database        = "db"
				connection_type = "Directly"
				` + extra + `
			}
		}`
}

// RD-1321136: import tracks the defaults Fivetran added (auth_type, s3_bucket_auth_type). Applying
// a configuration that doesn't declare them must not send null for them, which would clear them
// on the destination. Once the resource has been updated, removing a declared field clears it again.
func TestResourceDestinationV2ImportUndeclaredNullableFieldsMock(t *testing.T) {
	var destinationData map[string]interface{}
	var patchHandler, deleteHandler *mock.Handler
	var patchConfigs []map[string]interface{}

	setupMocks := func() {
		tfmock.MockClient().Reset()
		destinationData = nil
		patchConfigs = nil

		tfmock.MockClient().When(http.MethodGet, "/v1/metadata/destination-types/redshift_import_mock").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, destinationV2ImportMockMetadata)), nil
			},
		)

		tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				destinationData = tfmock.CreateMapFromJsonString(t, `
				{
					"id": "destination_id",
					"group_id": "group_id",
					"service": "redshift_import_mock",
					"time_zone_offset": "0",
					"setup_status": "connected",
					"daylight_saving_time_enabled": true,
					"networking_method": "Directly",
					"setup_tests": [],
					"config": {
						"host": "cluster.redshift.amazonaws.com",
						"port": 5439,
						"user": "fivetran_user",
						"database": "db",
						"connection_type": "Directly",
						"password": "******",
						"auth_type": "PASSWORD",
						"s3_bucket_auth_type": "IAM_ROLE"
					}
				}`)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", destinationData), nil
			},
		)

		tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				patchConfig, _ := body["config"].(map[string]interface{})
				patchConfigs = append(patchConfigs, patchConfig)
				config := destinationData["config"].(map[string]interface{})
				for k, v := range patchConfig {
					if v == nil {
						delete(config, k)
					} else if k != "password" {
						config[k] = v
					}
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				destinationData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been deleted", nil), nil
			},
		)
	}

	const name = "fivetran_destination_v2.dest"
	lastPatch := func() map[string]interface{} {
		if len(patchConfigs) == 0 {
			t.Fatalf("no PATCH was sent")
		}
		return patchConfigs[len(patchConfigs)-1]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 setupMocks,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{Config: destinationV2ImportMockConfig("")},
			{
				// Forget the destination without deleting it, then import it.
				Config: `
					removed {
						from = fivetran_destination_v2.dest
						lifecycle {
							destroy = false
						}
					}`,
			},
			{
				Config:             destinationV2ImportMockConfig(""),
				ResourceName:       name,
				ImportState:        true,
				ImportStateId:      "destination_id",
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					tfmock.AssertEqual(t, states[0].Attributes["config.auth_type"], "PASSWORD")
					tfmock.AssertEqual(t, states[0].Attributes["config.s3_bucket_auth_type"], "IAM_ROLE")
					return nil
				},
			},
			{
				// The undeclared imported fields stop being tracked but are not cleared upstream.
				Config: destinationV2ImportMockConfig(""),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, patchHandler.Interactions, 1)
						patch := lastPatch()
						tfmock.AssertKeyDoesNotExist(t, patch, "auth_type")
						tfmock.AssertKeyDoesNotExist(t, patch, "s3_bucket_auth_type")
						config := destinationData["config"].(map[string]interface{})
						tfmock.AssertEqual(t, config["auth_type"], "PASSWORD")
						tfmock.AssertEqual(t, config["s3_bucket_auth_type"], "IAM_ROLE")
						return nil
					},
					resource.TestCheckNoResourceAttr(name, "config.auth_type"),
					resource.TestCheckNoResourceAttr(name, "config.s3_bucket_auth_type"),
				),
			},
			{
				// Declare one of them: it is managed again.
				Config: destinationV2ImportMockConfig(`s3_bucket_auth_type = "IAM_ROLE"`),
				Check:  resource.TestCheckResourceAttr(name, "config.s3_bucket_auth_type", "IAM_ROLE"),
			},
			{
				// Removing a declared field after an update clears it, as before.
				Config: destinationV2ImportMockConfig(""),
				Check: func(s *terraform.State) error {
					patch := lastPatch()
					value, sent := patch["s3_bucket_auth_type"]
					if !sent || value != nil {
						t.Errorf("PATCH config = %v, want s3_bucket_auth_type sent as null", patch)
					}
					config := destinationData["config"].(map[string]interface{})
					tfmock.AssertKeyDoesNotExist(t, config, "s3_bucket_auth_type")
					tfmock.AssertEqual(t, config["auth_type"], "PASSWORD")
					return nil
				},
			},
		},
	})
}

// Right after import, an explicit null in config still clears an imported field.
func TestResourceDestinationV2ImportExplicitNullMock(t *testing.T) {
	var destinationData map[string]interface{}
	var patchConfigs []map[string]interface{}

	setupMocks := func() {
		tfmock.MockClient().Reset()
		patchConfigs = nil
		destinationData = tfmock.CreateMapFromJsonString(t, `
		{
			"id": "destination_id",
			"group_id": "group_id",
			"service": "redshift_import_mock",
			"time_zone_offset": "0",
			"setup_status": "connected",
			"daylight_saving_time_enabled": true,
			"networking_method": "Directly",
			"config": {
				"host": "cluster.redshift.amazonaws.com",
				"port": 5439,
				"user": "fivetran_user",
				"database": "db",
				"connection_type": "Directly",
				"password": "******",
				"auth_type": "PASSWORD",
				"s3_bucket_auth_type": "IAM_ROLE"
			}
		}`)

		tfmock.MockClient().When(http.MethodGet, "/v1/metadata/destination-types/redshift_import_mock").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, destinationV2ImportMockMetadata)), nil
			},
		)
		tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)
		tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				patchConfig, _ := body["config"].(map[string]interface{})
				patchConfigs = append(patchConfigs, patchConfig)
				config := destinationData["config"].(map[string]interface{})
				for k, v := range patchConfig {
					if v == nil {
						delete(config, k)
					} else if k != "password" {
						config[k] = v
					}
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)
		tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been deleted", nil), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 setupMocks,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             destinationV2ImportMockConfig(""),
				ResourceName:       "fivetran_destination_v2.dest",
				ImportState:        true,
				ImportStateId:      "destination_id",
				ImportStatePersist: true,
			},
			{
				Config: destinationV2ImportMockConfig(`s3_bucket_auth_type = null`),
				Check: func(s *terraform.State) error {
					if len(patchConfigs) != 1 {
						t.Fatalf("got %d PATCH requests, want 1", len(patchConfigs))
					}
					patch := patchConfigs[0]
					value, sent := patch["s3_bucket_auth_type"]
					if !sent || value != nil {
						t.Errorf("PATCH config = %v, want s3_bucket_auth_type sent as null", patch)
					}
					tfmock.AssertKeyDoesNotExist(t, patch, "auth_type")
					config := destinationData["config"].(map[string]interface{})
					tfmock.AssertKeyDoesNotExist(t, config, "s3_bucket_auth_type")
					tfmock.AssertEqual(t, config["auth_type"], "PASSWORD")
					return nil
				},
			},
		},
	})
}
