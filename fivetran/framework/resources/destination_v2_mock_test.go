package resources_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const destinationV2MockMetadata = `
{
	"id": "snowflake",
	"name": "Snowflake",
	"type": "Warehouse",
	"service_status": "general_availability",
	"supported_regions": ["GCP_US_EAST4", "AWS_US_EAST_1"],
	"supports_hybrid_deployment": false,
	"supported_networking_methods": ["Directly", "SshTunnel", "PrivateLink"],
	"supports_private_link": true,
	"private_link_types": ["CUSTOMER_MANAGED"],
	"config": {
		"type": "object",
		"properties": {
			"host":       {"type": "string", "readonly": false, "nullable": false, "immutable": false},
			"port":       {"type": "integer", "readonly": false, "nullable": false, "immutable": false},
			"user":       {"type": "string", "readonly": false, "nullable": false, "immutable": false},
			"database":   {"type": "string", "readonly": false, "nullable": false, "immutable": true},
			"password":   {"type": "string", "format": "password", "readonly": false, "nullable": true, "immutable": false},
			"public_key": {"type": "string", "readonly": true, "nullable": false, "immutable": false},
			"role":       {"type": "string", "readonly": false, "nullable": true, "immutable": false}
		}
	}
}`

func destinationV2MockConfig(host, database, extra string) string {
	return `
		resource "fivetran_destination_v2" "dest" {
			provider         = fivetran-provider
			group_id         = "group_id"
			service          = "snowflake"
			time_zone_offset = "0"
			region           = "GCP_US_EAST4"
			` + extra + `

			config = {
				host     = "` + host + `"
				port     = 443
				user     = "fivetran_user"
				database = "` + database + `"
				password = "secret"
			}
		}`
}

// assertDestinationV2UserAgent checks that a request carries the destination v2 user-agent suffix.
func assertDestinationV2UserAgent(t *testing.T, req *http.Request) {
	t.Helper()
	userAgent := req.Header.Get("User-Agent")
	if !strings.HasSuffix(userAgent, " fivetran_destination_v2") {
		t.Errorf("%s %s: User-Agent %q does not end with the fivetran_destination_v2 suffix", req.Method, req.URL.Path, userAgent)
	}
}

func TestResourceDestinationV2Mock(t *testing.T) {
	var destinationData map[string]interface{}
	var metadataHandler, postHandler, patchHandler, setupTestsHandler, deleteHandler *mock.Handler

	setupMocks := func() {
		tfmock.MockClient().Reset()
		destinationData = nil

		metadataHandler = tfmock.MockClient().When(http.MethodGet, "/v1/metadata/destination-types/snowflake").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				assertDestinationV2UserAgent(t, req)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, destinationV2MockMetadata)), nil
			},
		)

		postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				assertDestinationV2UserAgent(t, req)
				body := tfmock.RequestBodyToJson(t, req)
				tfmock.AssertKeyExistsAndHasValue(t, body, "region", "GCP_US_EAST4")
				tfmock.AssertKeyExistsAndHasValue(t, body, "run_setup_tests", true)
				config := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{})
				tfmock.AssertKeyExistsAndHasValue(t, config, "host", "account.snowflakecomputing.com")
				tfmock.AssertKeyExistsAndHasValue(t, config, "password", "secret")
				tfmock.AssertKeyDoesNotExist(t, config, "public_key")

				// The API masks the password, adds a readonly field and an upstream-only default.
				// port comes back as a string although the metadata declares it as an integer,
				// the way the destinations API serializes it.
				destinationData = tfmock.CreateMapFromJsonString(t, `
				{
					"id": "destination_id",
					"group_id": "group_id",
					"service": "snowflake",
					"region": "GCP_US_EAST4",
					"time_zone_offset": "0",
					"setup_status": "connected",
					"daylight_saving_time_enabled": true,
					"networking_method": "Directly",
					"setup_tests": [{"title": "Connection test", "status": "PASSED", "message": ""}],
					"config": {
						"host": "account.snowflakecomputing.com",
						"port": "443",
						"user": "fivetran_user",
						"database": "db",
						"password": "******",
						"public_key": "generated-public-key",
						"role": "FIVETRAN_ROLE"
					}
				}`)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", destinationData), nil
			},
		)

		tfmock.MockClient().When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				assertDestinationV2UserAgent(t, req)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		patchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				assertDestinationV2UserAgent(t, req)
				body := tfmock.RequestBodyToJson(t, req)
				config := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{})
				// Only the changed field is sent: no unchanged, password or readonly fields.
				tfmock.AssertEqual(t, len(config), 1)
				tfmock.AssertKeyExistsAndHasValue(t, config, "host", "new.snowflakecomputing.com")
				tfmock.AssertKeyDoesNotExist(t, body, "region")

				destinationData["config"].(map[string]interface{})["host"] = "new.snowflakecomputing.com"
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		setupTestsHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations/destination_id/test").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				tfmock.AssertKeyExistsAndHasValue(t, body, "trust_certificates", true)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				assertDestinationV2UserAgent(t, req)
				destinationData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been deleted", nil), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 setupMocks,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, deleteHandler.Interactions, 1)
			tfmock.AssertEmpty(t, destinationData)
			return nil
		},
		Steps: []resource.TestStep{
			{
				// Create: masked password, readonly and upstream-only fields don't leak into state,
				// and the post-apply refresh/plan is empty (the framework fails the step otherwise).
				Config: destinationV2MockConfig("account.snowflakecomputing.com", "db", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, postHandler.Interactions, 1)
						tfmock.AssertNotEqual(t, metadataHandler.Interactions, 0)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "id", "destination_id"),
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "setup_status", "connected"),
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config.password", "secret"),
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config.host", "account.snowflakecomputing.com"),
					// Stored as the integer the metadata declares, not the string the API returned.
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config.port", "443"),
					resource.TestCheckNoResourceAttr("fivetran_destination_v2.dest", "config.public_key"),
					resource.TestCheckNoResourceAttr("fivetran_destination_v2.dest", "config.role"),
				),
			},
			{
				// Update a mutable field: PATCH carries only that field.
				Config: destinationV2MockConfig("new.snowflakecomputing.com", "db", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, patchHandler.Interactions, 1)
						return nil
					},
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config.host", "new.snowflakecomputing.com"),
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config.password", "secret"),
				),
			},
			{
				// Only a setup-test flag changes: the setup-tests endpoint is used instead of an empty PATCH.
				Config: destinationV2MockConfig("new.snowflakecomputing.com", "db", `trust_certificates = true`),
				Check: func(s *terraform.State) error {
					tfmock.AssertEqual(t, setupTestsHandler.Interactions, 1)
					tfmock.AssertEqual(t, patchHandler.Interactions, 1)
					return nil
				},
			},
			{
				// Changing an immutable config field plans a replacement.
				Config:             destinationV2MockConfig("new.snowflakecomputing.com", "other_db", `trust_certificates = true`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fivetran_destination_v2.dest", plancheck.ResourceActionReplace),
					},
				},
			},
			{
				// Unknown config fields are rejected at plan time.
				Config: `
					resource "fivetran_destination_v2" "dest" {
						provider         = fivetran-provider
						group_id         = "group_id"
						service          = "snowflake"
						time_zone_offset = "0"
						config = {
							hots = "typo"
						}
					}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`does not define field "hots"`),
			},
			{
				// Import tracks API-visible fields only: no masked password, no readonly fields.
				Config:        destinationV2MockConfig("new.snowflakecomputing.com", "db", `trust_certificates = true`),
				ResourceName:  "fivetran_destination_v2.dest",
				ImportState:   true,
				ImportStateId: "destination_id",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					tfmock.AssertEqual(t, len(states), 1)
					attrs := states[0].Attributes
					tfmock.AssertEqual(t, attrs["config.host"], "new.snowflakecomputing.com")
					tfmock.AssertEqual(t, attrs["config.port"], "443")
					tfmock.AssertEqual(t, attrs["config.role"], "FIVETRAN_ROLE")
					tfmock.AssertEmpty(t, attrs["config.password"])
					tfmock.AssertEmpty(t, attrs["config.public_key"])
					return nil
				},
			},
		},
	})
}

func destinationV2SecretsMockConfig(secrets, secretsVersion string) string {
	return `
		resource "fivetran_destination_v2" "dest" {
			provider         = fivetran-provider
			group_id         = "group_id"
			service          = "snowflake"
			time_zone_offset = "0"
			region           = "GCP_US_EAST4"

			config = {
				host     = "account.snowflakecomputing.com"
				port     = 443
				user     = "fivetran_user"
				database = "db"
			}

			config_secrets         = ` + secrets + `
			config_secrets_version = ` + secretsVersion + `
		}`
}

func TestResourceDestinationV2ConfigSecretsMock(t *testing.T) {
	var destinationData map[string]interface{}
	var postHandler, patchHandler *mock.Handler
	var lastPatchConfig map[string]interface{}

	setupMocks := func() {
		tfmock.MockClient().Reset()
		destinationData = nil
		lastPatchConfig = nil

		tfmock.MockClient().When(http.MethodGet, "/v1/metadata/destination-types/snowflake").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, destinationV2MockMetadata)), nil
			},
		)

		postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/destinations").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := tfmock.RequestBodyToJson(t, req)
				config := tfmock.AssertKeyExists(t, body, "config").(map[string]interface{})
				tfmock.AssertKeyExistsAndHasValue(t, config, "host", "account.snowflakecomputing.com")
				tfmock.AssertKeyExistsAndHasValue(t, config, "password", "secret")

				destinationData = tfmock.CreateMapFromJsonString(t, `
				{
					"id": "destination_id",
					"group_id": "group_id",
					"service": "snowflake",
					"region": "GCP_US_EAST4",
					"time_zone_offset": "0",
					"setup_status": "connected",
					"daylight_saving_time_enabled": true,
					"networking_method": "Directly",
					"setup_tests": [],
					"config": {
						"host": "account.snowflakecomputing.com",
						"port": 443,
						"user": "fivetran_user",
						"database": "db",
						"password": "******"
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
				lastPatchConfig = tfmock.AssertKeyExists(t, body, "config").(map[string]interface{})
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", destinationData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				destinationData = nil
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Destination has been deleted", nil), nil
			},
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 setupMocks,
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		// Write-only attributes need Terraform 1.11+.
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
		},
		Steps: []resource.TestStep{
			{
				// Create sends the secret, but it never reaches state.
				Config: destinationV2SecretsMockConfig(`{ password = "secret" }`, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, postHandler.Interactions, 1)
						return nil
					},
					resource.TestCheckNoResourceAttr("fivetran_destination_v2.dest", "config.password"),
					resource.TestCheckNoResourceAttr("fivetran_destination_v2.dest", "config_secrets"),
					resource.TestCheckResourceAttr("fivetran_destination_v2.dest", "config_secrets_version", "1"),
				),
			},
			{
				// A new secret value without a version bump plans no change.
				Config: destinationV2SecretsMockConfig(`{ password = "rotated" }`, "1"),
				Check: func(s *terraform.State) error {
					tfmock.AssertEqual(t, patchHandler.Interactions, 0)
					return nil
				},
			},
			{
				// Bumping the version sends only the secret.
				Config: destinationV2SecretsMockConfig(`{ password = "rotated" }`, "2"),
				Check: func(s *terraform.State) error {
					tfmock.AssertEqual(t, patchHandler.Interactions, 1)
					tfmock.AssertEqual(t, len(lastPatchConfig), 1)
					tfmock.AssertKeyExistsAndHasValue(t, lastPatchConfig, "password", "rotated")
					return nil
				},
			},
			{
				// Only secret fields are accepted in config_secrets.
				Config:      destinationV2SecretsMockConfig(`{ role = "FIVETRAN_ROLE" }`, "2"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`is not a secret field`),
			},
		},
	})
}
