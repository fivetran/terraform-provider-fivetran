package resources_test

import (
	"fmt"
	"testing"

	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestConnectionV2ConfigSecretsMigrationMock(t *testing.T) {
	for _, tc := range []struct {
		name           string
		secret         string
		version        int
		moveAll        bool
		omitConfig     bool
		bumpVersion    bool
		nullConfig     bool
		skipValidation bool
	}{
		{name: "same_password_without_version", secret: "legacy-password", version: -1},
		{name: "new_password_unchanged_version", secret: "new-password", version: 1},
		{name: "all_fields_explicit_empty_config", secret: "new-password", version: 1, moveAll: true},
		{name: "all_fields_omitted_config", secret: "new-password", version: 1, moveAll: true, omitConfig: true},
		{name: "all_fields_omitted_config_version_bump", secret: "new-password", version: 1, moveAll: true, omitConfig: true, bumpVersion: true},
		{name: "all_fields_omitted_config_without_version", secret: "new-password", version: -1, moveAll: true, omitConfig: true},
		{name: "all_fields_null_config", secret: "new-password", version: 1, moveAll: true, nullConfig: true},
		{name: "all_fields_omitted_config_validation_disabled", secret: "new-password", version: 1, moveAll: true, omitConfig: true, skipValidation: true},
		{name: "password_only_omitted_config_preserves_other_fields", secret: "new-password", version: 1, omitConfig: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &configSecretsMock{}
			const name = "fivetran_connection_v2.test"
			create := postgresSecretsTfConfig("legacy-password", "", tc.version)
			move := postgresSecretsTfConfig("", tc.secret, tc.version)
			wantPatch := map[string]interface{}{"password": tc.secret}
			if tc.moveAll || tc.omitConfig {
				moveVersion := tc.version
				if tc.bumpVersion {
					moveVersion++
				}
				config := "config = {}"
				if tc.omitConfig {
					config = ""
				}
				if tc.nullConfig {
					config = "config = null"
				}
				version := ""
				if moveVersion >= 0 {
					version = fmt.Sprintf("config_secrets_version = %d", moveVersion)
				}
				otherSecrets := ""
				if tc.moveAll {
					otherSecrets = `host = "db.example.com"
    port = 5432
    user = "fivetran"`
					wantPatch = map[string]interface{}{"host": "db.example.com", "port": float64(5432), "user": "fivetran", "password": tc.secret}
				}
				move = fmt.Sprintf(`
ephemeral "testsecret_value" "pw" {
  value = %q
}
resource "fivetran_connection_v2" "test" {
  provider = fivetran-provider
  group_id = "group_id"
  service = "postgres"
  destination_schema_names = "FIVETRAN_NAMING"
  %s
  config_secrets = {
    %s
    password = ephemeral.testsecret_value.pw.secret
  }
  %s
}`, tc.secret, config, otherSecrets, version)
			}
			if tc.skipValidation {
				const providerConfig = `provider "fivetran-provider" {
  skip_plan_time_validation = true
}
`
				create = providerConfig + create
				move = providerConfig + move
			}
			otherFieldsCheck := resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(name, "config.host", "db.example.com"),
				resource.TestCheckResourceAttr(name, "config.port", "5432"),
				resource.TestCheckResourceAttr(name, "config.user", "fivetran"),
			)
			if tc.moveAll {
				otherFieldsCheck = resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(name, "config.host"),
					resource.TestCheckNoResourceAttr(name, "config.port"),
					resource.TestCheckNoResourceAttr(name, "config.user"),
				)
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
					{Config: create, Check: resource.TestCheckResourceAttr(name, "config.password", "legacy-password")},
					{Config: create, PlanOnly: true},
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
									tfmock.AssertEqual(t, m.patchBodies[0]["config"], wantPatch)
								}
								tfmock.AssertEqual(t, m.storedConfig["password"], tc.secret)
								return nil
							},
							resource.TestCheckNoResourceAttr(name, "config.password"),
							resource.TestCheckNoResourceAttr(name, "config_secrets"),
							otherFieldsCheck,
							checkSecretNotInState("legacy-password"),
							checkSecretNotInState(tc.secret),
						),
					},
					{Config: move, PlanOnly: true},
				},
			})
		})
	}
}
