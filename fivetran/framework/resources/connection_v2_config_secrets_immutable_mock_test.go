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

func TestConnectionV2ConfigSecretsImmutableMigrationMock(t *testing.T) {
	for _, tc := range []struct {
		name         string
		metadataFlag string
		movePassword bool
	}{
		{name: "immutable_only", metadataFlag: "immutable"},
		{name: "immutable_and_mutable", metadataFlag: "immutable", movePassword: true},
		{name: "readonly_and_mutable", metadataFlag: "readonly", movePassword: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A distinct service avoids sharing cached metadata with other mock scenarios.
			service := "postgres_" + tc.name
			m := &configSecretsMock{
				service: service,
				metadataResponse: strings.Replace(postgresSecretsMetadataResponse,
					`"host": {"type": "string"}`,
					fmt.Sprintf(`"host": {"type": "string", "%s": true}`, tc.metadataFlag), 1),
			}
			const name = "fivetran_connection_v2.test"
			config := func(move bool, version int) string {
				host := `host = "db.example.com"`
				password := `password = "plain-password"`
				secrets := ""
				if move {
					host = ""
					secrets = `host = "db.example.com"`
					if tc.movePassword {
						password = ""
						secrets += fmt.Sprintf("\npassword = %q", fmt.Sprintf("rotated-password-%d", version))
					}
				}
				return fmt.Sprintf(`
resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  service = %q
  group_id = "group_id"
  destination_schema_names = "FIVETRAN_NAMING"
  config = {
    %s
    %s
    port = 5432
    user = "fivetran"
  }
  config_secrets = { %s }
  config_secrets_version = %d
}`, service, host, password, secrets, version)
			}
			check := func(version int) resource.TestCheckFunc {
				return func(s *terraform.State) error {
					if len(m.postBodies) != 1 || m.deleteHandler.Interactions != 0 {
						return fmt.Errorf("migration replaced the connection")
					}
					wantPatches := 0
					if tc.movePassword {
						wantPatches = version
					}
					if len(m.patchBodies) != wantPatches {
						return fmt.Errorf("got %d PATCH requests, want %d", len(m.patchBodies), wantPatches)
					}
					for i, body := range m.patchBodies {
						patch := body["config"].(map[string]interface{})
						if _, sent := patch["host"]; sent {
							return fmt.Errorf("%s field was included in PATCH", tc.metadataFlag)
						}
						tfmock.AssertEqual(t, patch, map[string]interface{}{"password": fmt.Sprintf("rotated-password-%d", i+1)})
					}
					tfmock.AssertEqual(t, m.storedConfig["host"], "db.example.com")
					return nil
				}
			}
			steps := []resource.TestStep{
				{Config: config(false, 1), Check: resource.TestCheckResourceAttr(name, "config.host", "db.example.com")},
			}
			if tc.metadataFlag == "immutable" {
				steps = append(steps, resource.TestStep{
					Config:   strings.Replace(config(true, 1), `host = "db.example.com"`, `host = "other.example.com"`, 1),
					PlanOnly: true, ExpectNonEmptyPlan: true,
					ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(name, plancheck.ResourceActionReplace),
					}},
				})
			}
			steps = append(steps,
				resource.TestStep{
					Config: config(true, 1),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
					}},
					Check: resource.ComposeAggregateTestCheckFunc(check(1), resource.TestCheckNoResourceAttr(name, "config.host")),
				},
				resource.TestStep{Config: config(true, 1), PlanOnly: true},
				resource.TestStep{Config: config(true, 2), Check: resource.ComposeAggregateTestCheckFunc(check(2), resource.TestCheckNoResourceAttr(name, "config.host"))},
				resource.TestStep{Config: config(true, 2), PlanOnly: true},
			)
			resource.Test(t, resource.TestCase{
				TerraformVersionChecks:   configSecretsVersionChecks,
				PreCheck:                 func() { m.setup(t) },
				ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactoriesWithEphemeralSecret,
				CheckDestroy: func(s *terraform.State) error {
					tfmock.AssertEqual(t, m.deleteHandler.Interactions, 1)
					return nil
				},
				Steps: steps,
			})
		})
	}
}
