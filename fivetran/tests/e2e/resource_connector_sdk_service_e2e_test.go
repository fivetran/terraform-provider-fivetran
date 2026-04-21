package e2e_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// sdkConnectorConfigTemplate wires a package, group, and connector with
// service = "connector_sdk" together. %v placeholders: group name, zip path,
// schema name, secret value.
var sdkConnectorConfigTemplate = `
	resource "fivetran_group" "test_group" {
		provider = fivetran-provider
		name     = "%v"
	}

	resource "fivetran_connector_sdk_package" "test_pkg" {
		provider  = fivetran-provider
		file_path = "%v"
	}

	resource "fivetran_connector" "test_conn" {
		provider = fivetran-provider
		group_id = fivetran_group.test_group.id
		service  = "connector_sdk"

		destination_schema {
			name = "%v"
		}

		config {
			package_id     = fivetran_connector_sdk_package.test_pkg.id
			python_version = "3.13"
			secrets_list {
				key   = "api_key"
				value = "%v"
			}
		}

		run_setup_tests    = false
		trust_certificates = false
		trust_fingerprints = false
	}`

// sdkConnectorConfigTwoSecretsTemplate adds a db_password secret alongside
// api_key. %v placeholders: group name, zip path, schema name, api_key value,
// db_password value.
var sdkConnectorConfigTwoSecretsTemplate = `
	resource "fivetran_group" "test_group" {
		provider = fivetran-provider
		name     = "%v"
	}

	resource "fivetran_connector_sdk_package" "test_pkg" {
		provider  = fivetran-provider
		file_path = "%v"
	}

	resource "fivetran_connector" "test_conn" {
		provider = fivetran-provider
		group_id = fivetran_group.test_group.id
		service  = "connector_sdk"

		destination_schema {
			name = "%v"
		}

		config {
			package_id     = fivetran_connector_sdk_package.test_pkg.id
			python_version = "3.13"
			secrets_list {
				key   = "api_key"
				value = "%v"
			}
			secrets_list {
				key   = "db_password"
				value = "%v"
			}
		}

		run_setup_tests    = false
		trust_certificates = false
		trust_fingerprints = false
	}`

// sdkConnectorConfigSingleSecretTemplate lets us parameterize the key name,
// so we can move from {api_key} to {db_password} and assert that the original
// key is removed. %v placeholders: group name, zip path, schema name, key, value.
var sdkConnectorConfigSingleSecretTemplate = `
	resource "fivetran_group" "test_group" {
		provider = fivetran-provider
		name     = "%v"
	}

	resource "fivetran_connector_sdk_package" "test_pkg" {
		provider  = fivetran-provider
		file_path = "%v"
	}

	resource "fivetran_connector" "test_conn" {
		provider = fivetran-provider
		group_id = fivetran_group.test_group.id
		service  = "connector_sdk"

		destination_schema {
			name = "%v"
		}

		config {
			package_id     = fivetran_connector_sdk_package.test_pkg.id
			python_version = "3.13"
			secrets_list {
				key   = "%v"
				value = "%v"
			}
		}

		run_setup_tests    = false
		trust_certificates = false
		trust_fingerprints = false
	}`

// TestResourceConnectorSdkServiceE2E validates the cross-resource wiring:
// package is created first, connector references it via package_id, both
// destroy cleanly in reverse dependency order.
func TestResourceConnectorSdkServiceE2E(t *testing.T) {
	zipPath := writeTestSdkPackageZip(t, map[string][]byte{
		"connector.py": []byte("# sdk service test\n"),
	})
	groupName := "sdk_group_" + strconv.Itoa(seededRand.Int())
	schemaName := "sdk_schema_" + strconv.Itoa(seededRand.Int())

	config := fmt.Sprintf(sdkConnectorConfigTemplate, groupName, zipPath, schemaName, "initial_secret")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyConnectorAndPackage,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					testFivetranConnectorSdkPackageResourceCreate(t, "fivetran_connector_sdk_package.test_pkg"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "service", "connector_sdk"),
					resource.TestCheckResourceAttrPair(
						"fivetran_connector.test_conn", "config.package_id",
						"fivetran_connector_sdk_package.test_pkg", "id",
					),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.python_version", "3.13"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.0.key", "api_key"),
				),
			},
		},
	})
}

// TestResourceConnectorSdkServiceSecretsListValueChangeE2E verifies the
// provider-specific behavior of secrets_list: the API returns masked values
// ("***"), so change detection relies on local state vs config. Changing
// only a secret's value (same key) must be detected at plan time.
func TestResourceConnectorSdkServiceSecretsListValueChangeE2E(t *testing.T) {
	zipPath := writeTestSdkPackageZip(t, map[string][]byte{
		"connector.py": []byte("# sdk secrets test\n"),
	})
	groupName := "sdk_group_" + strconv.Itoa(seededRand.Int())
	schemaName := "sdk_schema_" + strconv.Itoa(seededRand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyConnectorAndPackage,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(sdkConnectorConfigTemplate, groupName, zipPath, schemaName, "v1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.0.value", "v1"),
				),
			},
			{
				Config: fmt.Sprintf(sdkConnectorConfigTemplate, groupName, zipPath, schemaName, "v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.0.value", "v2"),
				),
			},
		},
	})
}

// TestResourceConnectorSdkServiceSecretsListKeyChangeE2E walks the secrets_list
// through: start with one key, add a second, then drop the first — asserting
// that set cardinality and key names are tracked correctly.
func TestResourceConnectorSdkServiceSecretsListKeyChangeE2E(t *testing.T) {
	zipPath := writeTestSdkPackageZip(t, map[string][]byte{
		"connector.py": []byte("# sdk key change test\n"),
	})
	groupName := "sdk_group_" + strconv.Itoa(seededRand.Int())
	schemaName := "sdk_schema_" + strconv.Itoa(seededRand.Int())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             checkDestroyConnectorAndPackage,
		Steps: []resource.TestStep{
			{
				// Start with only api_key.
				Config: fmt.Sprintf(sdkConnectorConfigTemplate, groupName, zipPath, schemaName, "initial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.0.key", "api_key"),
				),
			},
			{
				// Add a second secret (db_password).
				Config: fmt.Sprintf(sdkConnectorConfigTwoSecretsTemplate, groupName, zipPath, schemaName, "initial", "dbpass"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.#", "2"),
				),
			},
			{
				// Remove api_key, keep only db_password.
				Config: fmt.Sprintf(sdkConnectorConfigSingleSecretTemplate, groupName, zipPath, schemaName, "db_password", "dbpass"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorResourceCreate(t, "fivetran_connector.test_conn"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.#", "1"),
					resource.TestCheckResourceAttr("fivetran_connector.test_conn", "config.secrets_list.0.key", "db_password"),
				),
			},
		},
	})
}

// checkDestroyConnectorAndPackage runs both per-resource destroy checks so
// this file doesn't conflict with the single-resource CheckDestroy used in
// other test files.
func checkDestroyConnectorAndPackage(s *terraform.State) error {
	if err := testFivetranConnectorResourceDestroy(s); err != nil {
		return err
	}
	if err := testFivetranConnectorSdkPackageResourceDestroy(s); err != nil {
		return err
	}
	return checkGroupsCleaned(s)
}

// checkGroupsCleaned verifies the test group was deleted along with its
// child resources.
func checkGroupsCleaned(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_group" {
			continue
		}
		_, err := client.NewGroupDetails().GroupID(rs.Primary.ID).Do(context.Background())
		if err == nil {
			return fmt.Errorf("group %s still exists after destroy", rs.Primary.ID)
		}
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
	}
	return nil
}
