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

// importImmutableMock is a connection whose immutable schema_prefix the API leaves out of config
// and returns as the top-level schema.
func importImmutableMock(service string) *configSecretsMock {
	return &configSecretsMock{
		service: service,
		metadataResponse: strings.Replace(postgresSecretsMetadataResponse,
			`"port": {"type": "integer"},`,
			`"port": {"type": "integer"}, "schema_prefix": {"type": "string", "immutable": true},`, 1),
		schema:       "pg",
		hiddenConfig: []string{"schema_prefix"},
	}
}

func importImmutableConfig(service, prefix string) string {
	return fmt.Sprintf(`
resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  service = %q
  group_id = "group_id"
  destination_schema_names = "FIVETRAN_NAMING"
  config = {
    host = "db.example.com"
    port = 5432
    user = "fivetran"
    password = "secret"
    schema_prefix = %q
  }
}`, service, prefix)
}

// importImmutableSteps creates the connection, forgets it without deleting it, and imports it.
func importImmutableSteps(service string) []resource.TestStep {
	config := importImmutableConfig(service, "pg")
	return []resource.TestStep{
		{Config: config},
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
			Config:             config,
			ResourceName:       "fivetran_connection_v2.test",
			ImportState:        true,
			ImportStateId:      "connection_id",
			ImportStatePersist: true,
		},
	}
}

// The imported state has no schema_prefix. When the configured prefix matches the
// imported schema, the first plan updates in place instead of replacing the connection, and the
// update stores schema_prefix, so the next plan is empty.
func TestConnectionV2ImportImmutableFieldMissingFromStateMock(t *testing.T) {
	const service = "postgres_import_immutable"
	const name = "fivetran_connection_v2.test"
	m := importImmutableMock(service)
	config := importImmutableConfig(service, "pg")

	steps := append(importImmutableSteps(service),
		resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
			}},
			Check: resource.ComposeAggregateTestCheckFunc(
				func(s *terraform.State) error {
					tfmock.AssertEqual(t, len(m.postBodies), 1)
					tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
					for _, body := range m.patchBodies {
						if config, ok := body["config"].(map[string]interface{}); ok {
							if _, sent := config["schema_prefix"]; sent {
								return fmt.Errorf("PATCH config = %v, want schema_prefix left out", config)
							}
						}
					}
					return nil
				},
				resource.TestCheckResourceAttr(name, "config.schema_prefix", "pg"),
			),
		},
		resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectEmptyPlan(),
			}},
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// A configured prefix that differs from the imported schema is a real change and still replaces
// the connection.
func TestConnectionV2ImportImmutableFieldMismatchReplacesMock(t *testing.T) {
	const service = "postgres_import_immutable_mismatch"
	const name = "fivetran_connection_v2.test"
	m := importImmutableMock(service)

	steps := append(importImmutableSteps(service),
		resource.TestStep{
			Config: importImmutableConfig(service, "other"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(name, plancheck.ResourceActionReplace),
			}},
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps:                    steps,
	})
}
