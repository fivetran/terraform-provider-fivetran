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

// Import stores immutable secrets masked, as the API returns them. Configuring the real value,
// in config or in config_secrets, must then update the connection in place instead of replacing it.
func TestConnectionV2ImportedMaskedImmutableSecretMock(t *testing.T) {
	for _, tc := range []struct {
		name          string
		moveToSecrets bool
	}{
		{name: "config"},
		{name: "config_secrets", moveToSecrets: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A distinct service avoids sharing cached metadata with other mock scenarios.
			service := "postgres_imported_" + tc.name
			m := &configSecretsMock{
				service: service,
				metadataResponse: strings.Replace(postgresSecretsMetadataResponse,
					`"format": "password"`, `"format": "password", "immutable": true`, 1),
			}
			const name = "fivetran_connection_v2.test"
			config := func(move bool) string {
				password := `password = "real-password"`
				secrets := ""
				if move {
					password = ""
					secrets = `
  config_secrets = { password = ephemeral.testsecret_value.pw.secret }
  config_secrets_version = 1`
				}
				return fmt.Sprintf(`
ephemeral "testsecret_value" "pw" {
  value = "real-password"
}

resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  service = %q
  group_id = "group_id"
  destination_schema_names = "FIVETRAN_NAMING"
  config = {
    host = "db.example.com"
    port = 5432
    user = "fivetran"
    %s
  }%s
}`, service, password, secrets)
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
					{Config: config(false)},
					// Forget the connection without deleting it, then import it.
					{
						Config: `
removed {
  from = fivetran_connection_v2.test
  lifecycle {
    destroy = false
  }
}`,
						Check: func(s *terraform.State) error {
							tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
							return nil
						},
					},
					{
						Config:             config(false),
						ResourceName:       name,
						ImportState:        true,
						ImportStateId:      "connection_id",
						ImportStatePersist: true,
						ImportStateCheck: func(states []*terraform.InstanceState) error {
							if got := states[0].Attributes["config.password"]; got != "******" {
								return fmt.Errorf("imported config.password = %q, want the masked value", got)
							}
							return nil
						},
					},
					{
						Config: config(tc.moveToSecrets),
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
						}},
						Check: resource.ComposeAggregateTestCheckFunc(
							func(s *terraform.State) error {
								tfmock.AssertEqual(t, len(m.postBodies), 1)
								tfmock.AssertEqual(t, m.deleteHandler.Interactions, 0)
								// Immutable fields are never sent in a PATCH.
								for _, body := range m.patchBodies {
									if config, ok := body["config"].(map[string]interface{}); ok {
										if _, sent := config["password"]; sent {
											return fmt.Errorf("immutable password was included in PATCH")
										}
									}
								}
								return nil
							},
							checkSecretNotInState("******"),
						),
					},
					{Config: config(tc.moveToSecrets), PlanOnly: true},
				},
			})
		})
	}
}
