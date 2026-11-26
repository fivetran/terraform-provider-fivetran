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

// stringCoercionConfig renders a connection whose config is the given HCL expression.
func stringCoercionConfig(service, config string) string {
	return fmt.Sprintf(`
variable "ssh" {
  type    = bool
  default = true
}

resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  service = %q
  group_id = "group_id"
  destination_schema_names = "FIVETRAN_NAMING"
  config = %s
}`, service, config)
}

// A string in the configuration for an integer field is accepted, sent as a number, and kept in
// state, so the next plan is empty. A conditional whose branches have different keys gives
// map(string), which turns port = 22 into "22".
func TestConnectionV2StringForIntegerFieldMock(t *testing.T) {
	for _, tc := range []struct {
		name, service, config string
		wantPort             float64
	}{
		{
			name:     "string literal",
			service:  "postgres_string_port",
			config:   `{ host = "h", user = "u", password = "p", port = "5432" }`,
			wantPort: 5432,
		},
		{
			name:     "conditional with different keys",
			service:  "postgres_conditional_port",
			config:   `var.ssh ? { host = "h", user = "u", password = "p", port = 22 } : { host = "h", user = "u", password = "p" }`,
			wantPort: 22,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &configSecretsMock{service: tc.service, metadataResponse: postgresSecretsMetadataResponse}
			config := stringCoercionConfig(tc.service, tc.config)
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { m.setup(t) },
				ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: func(s *terraform.State) error {
							tfmock.AssertEqual(t, len(m.postBodies), 1)
							posted, _ := m.postBodies[0]["config"].(map[string]interface{})
							tfmock.AssertEqual(t, posted["port"], tc.wantPort)
							return nil
						},
					},
					{Config: config, PlanOnly: true},
				},
			})
		})
	}
}

// Writing an immutable integer field as a string with the same value doesn't replace the
// connection, and nothing is sent.
func TestConnectionV2StringForImmutableIntegerFieldMock(t *testing.T) {
	const service = "postgres_immutable_string_port"
	const name = "fivetran_connection_v2.test"
	m := &configSecretsMock{
		service:          service,
		metadataResponse: strings.Replace(postgresSecretsMetadataResponse, `"port": {"type": "integer"}`, `"port": {"type": "integer", "immutable": true}`, 1),
	}
	asNumber := stringCoercionConfig(service, `{ host = "h", user = "u", password = "p", port = 5432 }`)
	asString := stringCoercionConfig(service, `{ host = "h", user = "u", password = "p", port = "5432" }`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { m.setup(t) },
		ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: asNumber},
			{
				Config: asString,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
				}},
				Check: func(s *terraform.State) error {
					tfmock.AssertEqual(t, len(m.postBodies), 1)
					tfmock.AssertEqual(t, len(m.patchBodies), 0)
					tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
					return nil
				},
			},
			{Config: asString, PlanOnly: true},
		},
	})
}
