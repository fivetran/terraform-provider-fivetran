package resources_test

import (
	"fmt"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Released versions that created fivetran_connection_v2 state with a different schema: each one
// adds attributes the previous one did not have.
var connectionV2ReleasedSchemaVersions = []string{
	"1.9.39", // first release, without destination_schema_names
	"1.9.40", // connect_card_config, destination_schema_names, destination_configuration
	"1.9.47", // schedule
	"1.9.50", // data_checks
	"1.9.54", // latest release
}

// The in-process provider takes the registry address of the released one, so Terraform treats
// the switch as a provider upgrade of the same resource, as it is for users.
const connectionV2UpgradeProviderNamespace = "fivetran"

var connectionV2UpgradeProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fivetran":   tfmock.ProtoV6ProviderFactories["fivetran-provider"],
	"testsecret": tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret["testsecret"],
}

// connectionV2UpgradeTfConfig renders the connection for both the released and the current
// provider. secret != "" moves the password to config_secrets, which needs the current provider.
func connectionV2UpgradeTfConfig(apiURL, releasedVersion string, current bool, secret string) string {
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
				fivetran   = { source = "fivetran/fivetran" }
				testsecret = { source = "fivetran/testsecret" }
			}
		}` + header
	}

	password := `password = "legacy-password"`
	secrets := ""
	ephemeralSecret := ""
	if secret != "" {
		password = ""
		secrets = `
			config_secrets = {
				password = ephemeral.testsecret_value.pw.secret
			}
			config_secrets_version = 1`
		ephemeralSecret = fmt.Sprintf(`
		ephemeral "testsecret_value" "pw" {
			value = %q
		}`, secret)
	}

	// 1.9.39 has no destination_schema_names. The current provider reads it from the API.
	destinationSchemaNames := `destination_schema_names = "FIVETRAN_NAMING"`
	if !current && releasedVersion == "1.9.39" {
		destinationSchemaNames = ""
	}

	return header + ephemeralSecret + fmt.Sprintf(`
		resource "fivetran_connection_v2" "test" {
			group_id                 = "group_id"
			service                  = "postgres"
			%s

			config = {
				host = "db.example.com"
				port = 5432
				user = "fivetran"
				%s
			}
			%s
		}`, destinationSchemaNames, password, secrets)
}

// State written by a released provider must upgrade to the current schema without changes, and
// the upgraded resource must then be able to move its password to config_secrets.
func TestConnectionV2ProviderUpgradeMock(t *testing.T) {
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", connectionV2UpgradeProviderNamespace)

	for _, releasedVersion := range connectionV2ReleasedSchemaVersions {
		t.Run(releasedVersion, func(t *testing.T) {
			m := &configSecretsMock{}
			const name = "fivetran_connection_v2.test"
			server := tfmock.NewMockServer(t)

			released := connectionV2UpgradeTfConfig(server.URL, releasedVersion, false, "")
			current := connectionV2UpgradeTfConfig(server.URL, releasedVersion, true, "")
			moved := connectionV2UpgradeTfConfig(server.URL, releasedVersion, true, "new-password")

			currentStep := func(step resource.TestStep) resource.TestStep {
				step.ProtoV6ProviderFactories = connectionV2UpgradeProviderFactories
				return step
			}

			resource.Test(t, resource.TestCase{
				TerraformVersionChecks: configSecretsVersionChecks,
				PreCheck:               func() { m.setup(t) },
				CheckDestroy: func(s *terraform.State) error {
					tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
					return nil
				},
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"fivetran": {Source: "fivetran/fivetran", VersionConstraint: releasedVersion},
						},
						Config: released,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "config.password", "legacy-password"),
							func(s *terraform.State) error {
								tfmock.AssertEqual(t, len(m.postBodies), 1)
								return nil
							},
						),
					},
					// The upgraded state plans no changes and sends nothing.
					currentStep(resource.TestStep{
						Config: current,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectEmptyPlan(),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							func(s *terraform.State) error {
								tfmock.AssertEqual(t, len(m.postBodies), 1)
								tfmock.AssertEqual(t, len(m.patchBodies), 0)
								return nil
							},
							resource.TestCheckResourceAttr(name, "id", "connection_id"),
							resource.TestCheckResourceAttr(name, "config.host", "db.example.com"),
							resource.TestCheckResourceAttr(name, "config.password", "legacy-password"),
							resource.TestCheckNoResourceAttr(name, "config_secrets"),
							resource.TestCheckNoResourceAttr(name, "config_secrets_version"),
						),
					}),
					currentStep(resource.TestStep{Config: current, PlanOnly: true}),
					// Moving the password to config_secrets updates the connection in place.
					currentStep(resource.TestStep{
						Config: moved,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							func(s *terraform.State) error {
								tfmock.AssertEqual(t, len(m.postBodies), 1)
								tfmock.AssertEqual(t, len(m.patchBodies), 1)
								if len(m.patchBodies) == 1 {
									tfmock.AssertEqual(t, m.patchBodies[0]["config"], map[string]interface{}{"password": "new-password"})
								}
								return nil
							},
							resource.TestCheckResourceAttr(name, "config.host", "db.example.com"),
							resource.TestCheckNoResourceAttr(name, "config.password"),
							resource.TestCheckResourceAttr(name, "config_secrets_version", "1"),
							checkSecretNotInState("legacy-password"),
							checkSecretNotInState("new-password"),
						),
					}),
					currentStep(resource.TestStep{Config: moved, PlanOnly: true}),
				},
			})
		})
	}
}
