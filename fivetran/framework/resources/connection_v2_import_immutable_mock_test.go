package resources_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// importImmutableMock is a connection whose immutable schema name fields the API leaves out of
// config and returns as the top-level schema.
func importImmutableMock(service, schema string, fields ...string) *configSecretsMock {
	props := ""
	for _, f := range fields {
		props += fmt.Sprintf(` %q: {"type": "string", "immutable": true},`, f)
	}
	return &configSecretsMock{
		service: service,
		metadataResponse: strings.Replace(postgresSecretsMetadataResponse,
			`"port": {"type": "integer"},`, `"port": {"type": "integer"},`+props, 1),
		schema:       schema,
		hiddenConfig: fields,
	}
}

// importImmutableConfig configures the connection with the given schema name fields.
func importImmutableConfig(service string, fields map[string]string) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	extra := ""
	for _, k := range keys {
		extra += fmt.Sprintf("    %s = %q\n", k, fields[k])
	}
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
%s  }
}`, service, extra)
}

// importImmutableSteps creates the connection, forgets it without deleting it, and imports it.
func importImmutableSteps(config string) []resource.TestStep {
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

// importImmutableUpdateSteps expects the first plan after import to update in place without
// sending the hidden fields, and the next plan to be empty.
func importImmutableUpdateSteps(t *testing.T, m *configSecretsMock, config string, want map[string]string) []resource.TestStep {
	const name = "fivetran_connection_v2.test"
	checks := []resource.TestCheckFunc{
		func(s *terraform.State) error {
			tfmock.AssertEqual(t, len(m.postBodies), 1)
			tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
			for _, body := range m.patchBodies {
				if config, ok := body["config"].(map[string]interface{}); ok {
					for _, f := range m.hiddenConfig {
						if _, sent := config[f]; sent {
							return fmt.Errorf("PATCH config = %v, want %s left out", config, f)
						}
					}
				}
			}
			return nil
		},
	}
	for k, v := range want {
		checks = append(checks, resource.TestCheckResourceAttr(name, "config."+k, v))
	}
	return append(importImmutableSteps(config),
		resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
			}},
			Check: resource.ComposeAggregateTestCheckFunc(checks...),
		},
		resource.TestStep{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectEmptyPlan(),
			}},
		},
	)
}

// The imported state has no schema_prefix. When the configured prefix matches the
// imported schema, the first plan updates in place instead of replacing the connection, and the
// update stores schema_prefix, so the next plan is empty.
func TestConnectionV2ImportImmutableFieldMissingFromStateMock(t *testing.T) {
	const service = "postgres_import_immutable"
	m := importImmutableMock(service, "pg", "schema_prefix")
	fields := map[string]string{"schema_prefix": "pg"}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps:                    importImmutableUpdateSteps(t, m, importImmutableConfig(service, fields), fields),
	})
}

// A single-table connector's top-level schema is "schema.table": the imported state has neither
// schema nor table, and both are compared with their part of the name.
func TestConnectionV2ImportImmutableSingleTableSchemaMock(t *testing.T) {
	const service = "postgres_import_immutable_single_table"
	m := importImmutableMock(service, "sales.orders", "schema", "table")
	fields := map[string]string{"schema": "sales", "table": "orders"}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps:                    importImmutableUpdateSteps(t, m, importImmutableConfig(service, fields), fields),
	})
}

// A configured prefix that differs from the imported schema is a real change and still replaces
// the connection.
func TestConnectionV2ImportImmutableFieldMismatchReplacesMock(t *testing.T) {
	const service = "postgres_import_immutable_mismatch"
	const name = "fivetran_connection_v2.test"
	m := importImmutableMock(service, "pg", "schema_prefix")

	steps := append(importImmutableSteps(importImmutableConfig(service, map[string]string{"schema_prefix": "pg"})),
		resource.TestStep{
			Config: importImmutableConfig(service, map[string]string{"schema_prefix": "other"}),
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
