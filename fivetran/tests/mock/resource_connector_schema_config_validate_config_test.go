package mock

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Confirms the "only one of schemas/schema/schemas_json" check now fails at
// terraform plan (via ValidateConfig), not just at apply. No mock HTTP handler
// is registered for this test — if the check ran at apply time instead, this
// test would fail with a connection error rather than the expected diagnostic.
func TestResourceConnectorSchemaConfigValidateConfigConflictingFieldsMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_schema_config" "test_schema" {
				provider     = fivetran-provider
				connector_id = "connector_id"
				schemas_json = jsonencode({
					"schema_1" = {
						enabled = true
					}
				})
				schemas = {
					"schema_1" = {
						enabled = true
					}
				}
			}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("You can use solely one field to define schema settings."),
	}

	resource.Test(
		t,
		resource.TestCase{
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps:                    []resource.TestStep{step1},
		},
	)
}
