package resources_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// auth_secrets is write-only: the value comes from an ephemeral resource, is sent to the API
// merged with auth, and must never reach the plan or the state. The API ignores unknown request
// fields, so the tests check where each value lands in the request body.

const authSecretsMetadataResponse = `{
	"id": "%s",
	"name": "Auth Service",
	"config": {
		"type": "object",
		"properties": {
			"host": {"type": "string"},
			"password": {"type": "string", "format": "password", "nullable": true}
		}
	},
	"auth": {
		"type": "object",
		"required": ["client_id", "client_secret"],
		"properties": {
			"client_id": {"type": "string"},
			"client_secret": {"type": "string", "format": "password", "nullable": true},
			"tenant": {"type": "string", "immutable": true}
		}
	}
}`

type authSecretsMock struct {
	service       string
	postBodies    []map[string]interface{}
	patchBodies   []map[string]interface{}
	deleteHandler *mock.Handler
}

func (m *authSecretsMock) setup(t *testing.T) {
	tfmock.MockClient().Reset()
	m.postBodies = nil
	m.patchBodies = nil
	storedConfig := map[string]interface{}{}

	// Like the API, never return auth, and return password fields masked.
	connectionResponse := func() map[string]interface{} {
		config := make(map[string]interface{}, len(storedConfig))
		for k, v := range storedConfig {
			config[k] = v
		}
		if _, ok := config["password"]; ok {
			config["password"] = "******"
		}
		return map[string]interface{}{
			"id":                       "connection_id",
			"group_id":                 "group_id",
			"service":                  m.service,
			"service_version":          1,
			"schema":                   "auth_service",
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

	tfmock.MockClient().When(http.MethodGet, "/v1/metadata/connector-types/"+m.service).ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success",
				tfmock.CreateMapFromJsonString(t, fmt.Sprintf(authSecretsMetadataResponse, m.service))), nil
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
			storedConfig = map[string]interface{}{}
			if config, ok := body["config"].(map[string]interface{}); ok {
				storedConfig = config
			}
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectionResponse()), nil
		},
	)
	tfmock.MockClient().When(http.MethodPatch, "/v1/connections/connection_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			m.patchBodies = append(m.patchBodies, body)
			if config, ok := body["config"].(map[string]interface{}); ok {
				for k, v := range config {
					storedConfig[k] = v
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

// authSecretsTfConfig renders a connection with the given auth and auth_secrets bodies (HCL
// attribute lines). An empty auth leaves the attribute out, and secretsVersion < 0 leaves out
// auth_secrets_version. extra is added to the resource block as is.
func authSecretsTfConfig(service, secret, auth, authSecrets string, secretsVersion int, extra string) string {
	authBlock := ""
	if auth != "" {
		authBlock = fmt.Sprintf("auth = {\n%s\n}", auth)
	}
	secretsBlock := ""
	if authSecrets != "" {
		secretsBlock = fmt.Sprintf("auth_secrets = {\n%s\n}", authSecrets)
	}
	versionLine := ""
	if secretsVersion >= 0 {
		versionLine = fmt.Sprintf("auth_secrets_version = %d", secretsVersion)
	}
	return fmt.Sprintf(`
ephemeral "testsecret_value" "pw" {
  value = %q
}

resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider

  group_id                 = "group_id"
  service                  = %q
  destination_schema_names = "FIVETRAN_NAMING"

  config = {
    host = "db.example.com"
  }
  %s
  %s
  %s
  %s
}`, secret, service, authBlock, secretsBlock, versionLine, extra)
}

func requestSection(t *testing.T, body map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	section, _ := body[name].(map[string]interface{})
	return section
}

func TestConnectionV2AuthSecretsMock(t *testing.T) {
	m := &authSecretsMock{service: "auth_service_basic"}
	const name = "fivetran_connection_v2.test"
	const clientID = `client_id = "client"`
	const secretRef = `client_secret = ephemeral.testsecret_value.pw.secret`

	// Create: client_secret, required by metadata, is set only in auth_secrets. It is sent in the
	// auth object, not in config, and is neither in auth nor in auth_secrets in state.
	create := resource.TestStep{
		Config: authSecretsTfConfig(m.service, "secret-v1", clientID, secretRef, 1, ""),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, len(m.postBodies), 1)
				tfmock.AssertEqual(t, requestSection(t, m.postBodies[0], "auth"),
					map[string]interface{}{"client_id": "client", "client_secret": "secret-v1"})
				tfmock.AssertEqual(t, requestSection(t, m.postBodies[0], "config"),
					map[string]interface{}{"host": "db.example.com"})
				return nil
			},
			resource.TestCheckResourceAttr(name, "auth.client_id", "client"),
			resource.TestCheckNoResourceAttr(name, "auth.client_secret"),
			resource.TestCheckNoResourceAttr(name, "auth_secrets"),
			resource.TestCheckResourceAttr(name, "auth_secrets_version", "1"),
			checkSecretNotInState("secret-v1"),
		),
	}

	// A new secret value without a version change is not detected, by design.
	rotateWithoutVersion := resource.TestStep{
		Config:   authSecretsTfConfig(m.service, "secret-v2", clientID, secretRef, 1, ""),
		PlanOnly: true,
	}

	// A version change sends the current auth secrets, and only them.
	rotate := resource.TestStep{
		Config: authSecretsTfConfig(m.service, "secret-v2", clientID, secretRef, 2, ""),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, len(m.patchBodies), 1)
				tfmock.AssertEqual(t, requestSection(t, m.patchBodies[0], "auth"),
					map[string]interface{}{"client_secret": "secret-v2"})
				if _, ok := m.patchBodies[0]["config"]; ok {
					t.Errorf("auth_secrets rotation sent config: %v", m.patchBodies[0]["config"])
				}
				return nil
			},
			resource.TestCheckResourceAttr(name, "auth_secrets_version", "2"),
			checkSecretNotInState("secret-v2"),
		),
	}

	// A field set in both auth and auth_secrets is rejected at plan time.
	overlap := resource.TestStep{
		Config:      authSecretsTfConfig(m.service, "secret-v1", clientID+"\nclient_secret = \"plain\"", secretRef, 1, ""),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Field "client_secret" is set in both auth and auth_secrets`),
	}
	overlapWithValidationDisabled := overlap
	overlapWithValidationDisabled.Config = `provider "fivetran-provider" {
  skip_plan_time_validation = true
}
` + overlap.Config

	// On create, the required client_secret is missing from both auth and auth_secrets.
	missingRequired := resource.TestStep{
		Config:      authSecretsTfConfig(m.service, "secret-v1", clientID, `tenant = "t"`, 1, ""),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`(?s)Missing Required.*client_secret`),
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
			missingRequired,
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

func TestConnectionV2AuthSecretsWithConfigSecretsMock(t *testing.T) {
	m := &authSecretsMock{service: "auth_service_both"}
	const name = "fivetran_connection_v2.test"
	tfConfig := func(version int) string {
		return authSecretsTfConfig(m.service, "auth-secret", `client_id = "client"`,
			`client_secret = ephemeral.testsecret_value.pw.secret`, version, fmt.Sprintf(`
  config_secrets = {
    password = "config-secret"
  }
  config_secrets_version = 1`))
	}

	// Each secret goes to its own section, and bumping one version sends only that section.
	create := resource.TestStep{
		Config: tfConfig(1),
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, requestSection(t, m.postBodies[0], "config"),
					map[string]interface{}{"host": "db.example.com", "password": "config-secret"})
				tfmock.AssertEqual(t, requestSection(t, m.postBodies[0], "auth"),
					map[string]interface{}{"client_id": "client", "client_secret": "auth-secret"})
				return nil
			},
			checkSecretNotInState("config-secret"),
			checkSecretNotInState("auth-secret"),
		),
	}
	bumpAuth := resource.TestStep{
		Config: tfConfig(2),
		Check: func(s *terraform.State) error {
			tfmock.AssertEqual(t, len(m.patchBodies), 1)
			tfmock.AssertEqual(t, requestSection(t, m.patchBodies[0], "auth"),
				map[string]interface{}{"client_secret": "auth-secret"})
			if _, ok := m.patchBodies[0]["config"]; ok {
				t.Errorf("auth_secrets_version change sent config: %v", m.patchBodies[0]["config"])
			}
			return nil
		},
	}

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		Steps: []resource.TestStep{
			create,
			{Config: create.Config, PlanOnly: true},
			bumpAuth,
			{Config: bumpAuth.Config, PlanOnly: true},
			{Config: bumpAuth.Config, Check: resource.TestCheckNoResourceAttr(name, "auth_secrets")},
		},
	})
}

func TestConnectionV2AuthSecretsMoveFromAuthMock(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version int
		auth    string
		secrets string
		want    map[string]interface{}
	}{
		{
			name: "secret_only_without_version", version: -1,
			auth: `client_id = "client"`, secrets: `client_secret = ephemeral.testsecret_value.pw.secret`,
			want: map[string]interface{}{"client_secret": "secret-v1"},
		},
		{
			name: "secret_only_unchanged_version", version: 1,
			auth: `client_id = "client"`, secrets: `client_secret = ephemeral.testsecret_value.pw.secret`,
			want: map[string]interface{}{"client_secret": "secret-v1"},
		},
		{
			name: "all_fields_auth_omitted", version: 1,
			secrets: "client_id = \"client\"\nclient_secret = ephemeral.testsecret_value.pw.secret",
			want:    map[string]interface{}{"client_id": "client", "client_secret": "secret-v1"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &authSecretsMock{service: "auth_service_move_" + tc.name}
			const name = "fivetran_connection_v2.test"
			create := authSecretsTfConfig(m.service, "unused", "client_id = \"client\"\nclient_secret = \"plain\"", "", tc.version, "")
			move := authSecretsTfConfig(m.service, "secret-v1", tc.auth, tc.secrets, tc.version, "")

			// Moving fields to auth_secrets without a version change sends their values instead
			// of clearing the fields that are no longer in auth, and removes them from state.
			resource.Test(t, resource.TestCase{
				TerraformVersionChecks:   configSecretsVersionChecks,
				PreCheck:                 func() { m.setup(t) },
				ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
				CheckDestroy: func(s *terraform.State) error {
					tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
					return nil
				},
				Steps: []resource.TestStep{
					{Config: create, Check: resource.TestCheckResourceAttr(name, "auth.client_secret", "plain")},
					{
						Config: move,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							func(s *terraform.State) error {
								tfmock.AssertEqual(t, len(m.postBodies), 1)
								tfmock.AssertEqual(t, len(m.patchBodies), 1)
								if len(m.patchBodies) == 1 {
									tfmock.AssertEqual(t, requestSection(t, m.patchBodies[0], "auth"), tc.want)
								}
								return nil
							},
							resource.TestCheckNoResourceAttr(name, "auth.client_secret"),
							resource.TestCheckNoResourceAttr(name, "auth_secrets"),
							checkSecretNotInState("plain"),
							checkSecretNotInState("secret-v1"),
						),
					},
					{Config: move, PlanOnly: true},
				},
			})
		})
	}
}

func TestConnectionV2AuthSecretsImmutableMock(t *testing.T) {
	const name = "fivetran_connection_v2.test"
	const base = "client_id = \"client\"\nclient_secret = \"plain\""

	// Moving an unchanged immutable field only removes it from state: nothing is sent and the
	// connection is not replaced.
	t.Run("unchanged_migration", func(t *testing.T) {
		m := &authSecretsMock{service: "auth_service_immutable_unchanged"}
		create := authSecretsTfConfig(m.service, "unused", base+"\ntenant = \"t1\"", "", 1, "")
		move := authSecretsTfConfig(m.service, "unused", base, `tenant = "t1"`, 1, "")
		resource.Test(t, resource.TestCase{
			TerraformVersionChecks:   configSecretsVersionChecks,
			PreCheck:                 func() { m.setup(t) },
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
			Steps: []resource.TestStep{
				{Config: create},
				{
					Config: move,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
					}},
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, len(m.postBodies), 1)
							tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
							for _, body := range m.patchBodies {
								if _, sent := requestSection(t, body, "auth")["tenant"]; sent {
									t.Errorf("immutable tenant was included in PATCH: %v", body)
								}
							}
							return nil
						},
						resource.TestCheckNoResourceAttr(name, "auth.tenant"),
					),
				},
				{Config: move, PlanOnly: true},
			},
		})
	})

	// Moving an immutable field with a changed value replaces the connection.
	t.Run("changed_migration", func(t *testing.T) {
		m := &authSecretsMock{service: "auth_service_immutable_changed"}
		create := authSecretsTfConfig(m.service, "unused", base+"\ntenant = \"t1\"", "", 1, "")
		move := authSecretsTfConfig(m.service, "unused", base, `tenant = "t2"`, 1, "")
		resource.Test(t, resource.TestCase{
			TerraformVersionChecks:   configSecretsVersionChecks,
			PreCheck:                 func() { m.setup(t) },
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
			Steps: []resource.TestStep{
				{Config: create},
				{
					Config: move,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(name, plancheck.ResourceActionDestroyBeforeCreate),
					}},
					Check: func(s *terraform.State) error {
						tfmock.AssertEqual(t, len(m.postBodies), 2)
						tfmock.AssertEqual(t, requestSection(t, m.postBodies[1], "auth"),
							map[string]interface{}{"client_id": "client", "client_secret": "plain", "tenant": "t2"})
						return nil
					},
				},
			},
		})
	})
}

// The API never returns auth, so imported state has neither auth nor auth_secrets_version. The
// first apply after import sends auth_secrets because the version changes from unset.
func TestConnectionV2AuthSecretsAfterImportMock(t *testing.T) {
	m := &authSecretsMock{service: "auth_service_import"}
	const name = "fivetran_connection_v2.test"
	tfConfig := authSecretsTfConfig(m.service, "secret-v1", `client_id = "client"`,
		`client_secret = ephemeral.testsecret_value.pw.secret`, 1, "")

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		CheckDestroy: func(s *terraform.State) error {
			tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
			return nil
		},
		Steps: []resource.TestStep{
			{Config: tfConfig},
			// Forget the connection without deleting it, then import it.
			{
				Config: `
removed {
  from = fivetran_connection_v2.test
  lifecycle {
    destroy = false
  }
}`,
			},
			{
				Config:             tfConfig,
				ResourceName:       name,
				ImportState:        true,
				ImportStateId:      "connection_id",
				ImportStatePersist: true,
			},
			{
				Config: tfConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, len(m.postBodies), 1)
						tfmock.AssertEqual(t, len(m.patchBodies), 1)
						if len(m.patchBodies) == 1 {
							tfmock.AssertEqual(t, requestSection(t, m.patchBodies[0], "auth"),
								map[string]interface{}{"client_id": "client", "client_secret": "secret-v1"})
						}
						return nil
					},
					resource.TestCheckResourceAttr(name, "auth.client_id", "client"),
					checkSecretNotInState("secret-v1"),
				),
			},
			{Config: tfConfig, PlanOnly: true},
		},
	})
}

// Without auth_secrets_version, an imported connection keeps its credentials: auth_secrets are
// write-only and the version stays unset, so the first plan after import is empty and nothing is
// sent.
func TestConnectionV2AuthSecretsAfterImportWithoutVersionMock(t *testing.T) {
	m := &authSecretsMock{service: "auth_service_import_no_version"}
	const name = "fivetran_connection_v2.test"
	created := authSecretsTfConfig(m.service, "secret-v1", `client_id = "client"`,
		`client_secret = ephemeral.testsecret_value.pw.secret`, 1, "")
	imported := authSecretsTfConfig(m.service, "secret-v1", "", `client_id = "client"
    client_secret = ephemeral.testsecret_value.pw.secret`, -1, "")

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		Steps: []resource.TestStep{
			{Config: created},
			// Forget the connection without deleting it, then import it.
			{
				Config: `
removed {
  from = fivetran_connection_v2.test
  lifecycle {
    destroy = false
  }
}`,
			},
			{
				Config:             imported,
				ResourceName:       name,
				ImportState:        true,
				ImportStateId:      "connection_id",
				ImportStatePersist: true,
			},
			{
				Config: imported,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				}},
				Check: func(s *terraform.State) error {
					tfmock.AssertEqual(t, len(m.patchBodies), 0)
					return nil
				},
			},
		},
	})
}
