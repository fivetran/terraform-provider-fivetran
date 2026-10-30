package resources_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// config_secrets is write-only: the value comes from an ephemeral resource, is sent to the API
// merged with config, and must never reach the plan or the state.

const postgresSecretsMetadataResponse = `{
	"id": "postgres",
	"name": "Postgres",
	"config": {
		"type": "object",
		"required": ["host", "user", "password"],
		"properties": {
			"host": {"type": "string"},
			"port": {"type": "integer"},
			"user": {"type": "string"},
			"password": {"type": "string", "format": "password", "nullable": true}
		}
	},
	"auth": {"type": "object", "properties": {}}
}`

// configSecretsVersionChecks skips the tests on Terraform versions without write-only attributes.
var configSecretsVersionChecks = []tfversion.TerraformVersionCheck{
	tfversion.SkipBelow(version.Must(version.NewVersion("1.11.0"))),
}

// postgresSecretsTfConfig renders a connection with config and, when secret is not empty, a
// config_secrets.password taken from an ephemeral resource. version < 0 leaves out
// config_secrets_version.
func postgresSecretsTfConfig(configPassword, secret string, version int) string {
	passwordLine := ""
	if configPassword != "" {
		passwordLine = fmt.Sprintf("password = %q", configPassword)
	}
	secretsBlock := ""
	if secret != "" {
		secretsBlock = `
			config_secrets = {
				password = ephemeral.testsecret_value.pw.secret
			}`
	}
	versionLine := ""
	if version >= 0 {
		versionLine = fmt.Sprintf("config_secrets_version = %d", version)
	}

	return fmt.Sprintf(`
		ephemeral "testsecret_value" "pw" {
			value = %q
		}

		resource "fivetran_connection_v2" "test" {
			provider = fivetran-provider

			group_id                 = "group_id"
			service                  = "postgres"
			destination_schema_names = "FIVETRAN_NAMING"

			config = {
				host = "db.example.com"
				port = 5432
				user = "fivetran"
				%s
			}
			%s
			%s
		}`, secret, passwordLine, secretsBlock, versionLine)
}

type configSecretsMock struct {
	service          string
	metadataResponse string
	storedConfig     map[string]interface{}
	postBodies       []map[string]interface{}
	patchBodies      []map[string]interface{}
	deleteHandler    *mock.Handler
}

func (m *configSecretsMock) setup(t *testing.T) {
	tfmock.MockClient().Reset()
	m.storedConfig = nil
	m.postBodies = nil
	m.patchBodies = nil
	service := m.service
	if service == "" {
		service = "postgres"
	}
	metadataResponse := m.metadataResponse
	if metadataResponse == "" {
		metadataResponse = postgresSecretsMetadataResponse
	}

	connectionResponse := func() map[string]interface{} {
		// Like the API, return password fields masked.
		config := make(map[string]interface{}, len(m.storedConfig))
		for k, v := range m.storedConfig {
			config[k] = v
		}
		if _, ok := config["password"]; ok {
			config["password"] = "******"
		}
		return map[string]interface{}{
			"id":                       "connection_id",
			"group_id":                 "group_id",
			"service":                  service,
			"service_version":          1,
			"schema":                   "postgres",
			"paused":                   true,
			"destination_schema_names": "FIVETRAN_NAMING",
			"status": map[string]interface{}{
				"setup_state":        "connected",
				"sync_state":         "paused",
				"update_state":       "on_schedule",
				"is_historical_sync": false,
				"tasks":              []interface{}{},
				"warnings":           []interface{}{},
			},
			"setup_tests": []interface{}{},
			"config":      config,
		}
	}

	tfmock.MockClient().When(http.MethodGet, "/v1/metadata/connector-types/"+service).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
				tfmock.CreateMapFromJsonString(t, metadataResponse)), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionResponse()), nil
		},
	)

	tfmock.MockClient().When(http.MethodPost, "/v1/connections").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			m.postBodies = append(m.postBodies, body)
			m.storedConfig = body["config"].(map[string]interface{})
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectionResponse()), nil
		},
	)

	tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			m.patchBodies = append(m.patchBodies, body)
			if config, ok := body["config"].(map[string]interface{}); ok {
				for k, v := range config {
					m.storedConfig[k] = v
				}
			}
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionResponse()), nil
		},
	)

	m.deleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

// checkSecretNotInState fails if any attribute of any resource in state contains secret.
func checkSecretNotInState(secret string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			for key, value := range rs.Primary.Attributes {
				if strings.Contains(value, secret) {
					return fmt.Errorf("secret found in state of %s at %s", name, key)
				}
			}
		}
		return nil
	}
}

func TestConnectionV2ConfigSecretsMock(t *testing.T) {
	m := &configSecretsMock{}
	const resourceName = "fivetran_connection_v2.test"

	// Create: password, required by metadata, is set only in config_secrets. It is sent merged
	// with config and is neither in config nor in config_secrets in state.
	create := resource.TestStep{
		Config: postgresSecretsTfConfig("", "secret-v1", 1),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, len(m.postBodies), 1)
				config := m.postBodies[0]["config"].(map[string]interface{})
				tfmock.AssertKeyExistsAndHasValue(t, config, "password", "secret-v1")
				tfmock.AssertKeyExistsAndHasValue(t, config, "host", "db.example.com")
				return nil
			},
			resource.TestCheckResourceAttr(resourceName, "config.host", "db.example.com"),
			resource.TestCheckNoResourceAttr(resourceName, "config.password"),
			resource.TestCheckNoResourceAttr(resourceName, "config_secrets"),
			resource.TestCheckResourceAttr(resourceName, "config_secrets_version", "1"),
			checkSecretNotInState("secret-v1"),
		),
	}

	// A new secret value without a version change is not detected, by design.
	rotateWithoutVersion := resource.TestStep{
		Config:   postgresSecretsTfConfig("", "secret-v2", 1),
		PlanOnly: true,
	}

	// A version change sends the current secrets, and only them.
	rotate := resource.TestStep{
		Config: postgresSecretsTfConfig("", "secret-v2", 2),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, len(m.patchBodies), 1)
				tfmock.AssertEqual(t, m.patchBodies[0]["config"], map[string]interface{}{"password": "secret-v2"})
				tfmock.AssertEqual(t, m.storedConfig["password"], "secret-v2")
				return nil
			},
			resource.TestCheckResourceAttr(resourceName, "config_secrets_version", "2"),
			checkSecretNotInState("secret-v2"),
		),
	}

	// A field set in both config and config_secrets is rejected at plan time.
	overlap := resource.TestStep{
		Config:      postgresSecretsTfConfig("plain", "secret-v1", 1),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Conflicting Dynamic Field`),
	}
	overlapWithValidationDisabled := overlap
	overlapWithValidationDisabled.Config = `provider "fivetran-provider" {
  skip_plan_time_validation = true
}
` + overlap.Config

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			create,
			{Config: create.Config, PlanOnly: true},
			// Not last: the post-test destroy uses the last step's configuration.
			overlap,
			overlapWithValidationDisabled,
			rotateWithoutVersion,
			rotate,
			{Config: rotate.Config, PlanOnly: true},
		},
	})
}

func TestConnectionV2ConfigSecretsMoveFromConfigMock(t *testing.T) {
	m := &configSecretsMock{}
	const resourceName = "fivetran_connection_v2.test"

	// The password starts in config, so it is stored in state.
	create := resource.TestStep{
		Config: postgresSecretsTfConfig("plain", "", -1),
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "config.password", "plain"),
		),
	}

	// Moving it to config_secrets without setting a version sends the secret instead of clearing
	// the field that is no longer in config, and removes it from state.
	move := resource.TestStep{
		Config: postgresSecretsTfConfig("", "secret-v1", -1),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, len(m.patchBodies), 1)
				tfmock.AssertEqual(t, m.patchBodies[0]["config"], map[string]interface{}{"password": "secret-v1"})
				return nil
			},
			resource.TestCheckNoResourceAttr(resourceName, "config.password"),
			checkSecretNotInState("plain"),
			checkSecretNotInState("secret-v1"),
		),
	}

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			create,
			move,
			{Config: move.Config, PlanOnly: true},
		},
	})
}
