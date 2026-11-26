package resources_test

import (
	"fmt"
	"strings"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// RD-1321136: config is optional for fivetran_connection_v2. When it is omitted while a secret
// moves to config_secrets, the plan inherits the prior config without the secret, so the update
// keeps the undeclared imported fields in state. A later partial config must still not send null
// for them.
func TestConnectionV2ImportSecretsMigrationThenPartialConfigMock(t *testing.T) {
	// A distinct service avoids sharing cached metadata with other mock scenarios.
	const service = "postgres_imported_partial"
	m := &configSecretsMock{
		service: service,
		metadataResponse: strings.Replace(postgresSecretsMetadataResponse,
			`"port": {"type": "integer"},`,
			`"port": {"type": "integer"}, "update_method": {"type": "string", "nullable": true},`, 1),
	}
	const name = "fivetran_connection_v2.test"
	connection := func(body string) string {
		return fmt.Sprintf(`
ephemeral "testsecret_value" "pw" {
  value = "real-password"
}

resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  service = %q
  group_id = "group_id"
  destination_schema_names = "FIVETRAN_NAMING"
%s
}`, service, body)
	}
	partialConfig := func(port string) string {
		return connection(fmt.Sprintf(`
  config = {
    host = "db.example.com"
    port = %s
    user = "fivetran"
  }
  config_secrets = { password = ephemeral.testsecret_value.pw.secret }
  config_secrets_version = 1`, port))
	}
	fullConfig := connection(`
  config = {
    host = "db.example.com"
    port = 5432
    user = "fivetran"
    password = "real-password"
    update_method = "XMIN"
  }`)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks:   configSecretsVersionChecks,
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
		Steps: []resource.TestStep{
			{Config: fullConfig},
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
				Config:             fullConfig,
				ResourceName:       name,
				ImportState:        true,
				ImportStateId:      "connection_id",
				ImportStatePersist: true,
			},
			{
				// config is omitted and the password moves to config_secrets: config inherits every
				// other imported field.
				Config: connection(`
  config_secrets = { password = ephemeral.testsecret_value.pw.secret }
  config_secrets_version = 1`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, len(m.patchBodies), 1)
						return nil
					},
					resource.TestCheckResourceAttr(name, "config.update_method", "XMIN"),
				),
			},
			{
				// A partial config doesn't declare update_method. The only difference from state is that
				// undeclared imported field, so the plan keeps it in state and changes nothing.
				Config: partialConfig("5432"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, len(m.patchBodies), 1)
						return nil
					},
					resource.TestCheckResourceAttr(name, "config.update_method", "XMIN"),
				),
			},
			{
				// port changes too, so the plan must follow the configuration: update_method shows as
				// removed, stops being tracked, and is not cleared.
				Config: partialConfig("5433"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						tfmock.AssertEqual(t, len(m.patchBodies), 2)
						config, _ := m.patchBodies[1]["config"].(map[string]interface{})
						if _, sent := config["update_method"]; sent {
							return fmt.Errorf("PATCH config = %v, want update_method left out", config)
						}
						tfmock.AssertEqual(t, config["port"], float64(5433))
						tfmock.AssertEqual(t, m.storedConfig["update_method"], "XMIN")
						return nil
					},
					resource.TestCheckNoResourceAttr(name, "config.update_method"),
				),
			},
			{
				// update_method is no longer tracked, so the next plan is empty.
				Config:   partialConfig("5433"),
				PlanOnly: true,
			},
		},
	})
}
