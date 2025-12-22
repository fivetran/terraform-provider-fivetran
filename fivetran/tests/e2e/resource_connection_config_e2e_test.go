package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceConnectionConfigE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_connection_config"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_test"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				update_method = "XMIN"
        				user = "initial_user"
        				password = "initial_password"
        				host = "initial.example.com"
        				port = "5432"
        				database = "initial_db"
      				}
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.test_connector.id

					config = jsonencode({
						update_method = "XMIN"
						user = "updated_user"
						host = "updated.example.com"
						port = 5432
						database = "updated_db"
					})

					auth = jsonencode({
						password = "updated_password"
					})
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "auth"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_connection_config"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_test"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				update_method = "XMIN"
        				user = "initial_user"
        				password = "initial_password"
        				host = "initial.example.com"
        				port = "5432"
        				database = "initial_db"
      				}
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.test_connector.id

					config = jsonencode({
						user = "updated_user_v2"
						host = "updated2.example.com"
						port = 5433
						database = "updated_db_v2"
					})

					auth = jsonencode({
						password = "updated_password_v2"
					})
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceUpdate(t, "fivetran_connection_config.test_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "auth"),
				),
			},
			{
				ResourceName:      "fivetran_connection_config.test_config",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestResourceConnectionConfigOnlyConfigE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_config_only"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_config_only"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				update_method = "XMIN"
        				user = "initial_user"
        				password = "initial_password"
        				host = "initial.example.com"
        				port = "5432"
        				database = "initial_db"
      				}
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.test_connector.id

					config = jsonencode({
						user = "config_only_user"
						host = "config.example.com"
						port = 5432
						database = "config_only_db"
					})
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "config"),
				),
			},
		},
	})
}

func TestResourceConnectionConfigOnlyAuthE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_auth_only"
			    }

			    resource "fivetran_connector" "test_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_auth_only"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				update_method = "XMIN"
        				user = "initial_user"
        				password = "initial_password"
        				host = "initial.example.com"
        				port = "5432"
        				database = "initial_db"
      				}
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.test_connector.id

					auth = jsonencode({
						password = "auth_only_password"
					})
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "auth"),
				),
			},
		},
	})
}

func TestResourceConnectionConfigMultipleConnectorTypesE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_multi_connector"
			    }

			    resource "fivetran_connector" "mysql_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "mysql"

					destination_schema {
						prefix = "mysql_test"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				user = "mysql_user"
        				password = "mysql_password"
        				host = "mysql.example.com"
        				port = "3306"
        				database = "mysql_db"
      				}
				}

				resource "fivetran_connection_config" "mysql_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.mysql_connector.id

					config = jsonencode({
						update_method = "BINLOG"
						user = "mysql_updated_user"
						host = "mysql-updated.example.com"
						port = 3306
						database = "mysql_updated_db"
					})

					auth = jsonencode({
						password = "mysql_updated_password"
					})
				}

			    resource "fivetran_connector" "postgres_connector" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_multi_test"
					}

					trust_certificates = false
					trust_fingerprints = false
					run_setup_tests = false

      				config {
        				update_method = "XMIN"
        				user = "postgres_user"
        				password = "postgres_password"
        				host = "postgres.example.com"
        				port = "5432"
        				database = "postgres_db"
      				}
				}

				resource "fivetran_connection_config" "postgres_config" {
					provider = fivetran-provider
					connection_id = fivetran_connector.postgres_connector.id

					config = jsonencode({
						update_method = "XMIN"
						user = "postgres_updated_user"
						host = "postgres-updated.example.com"
						port = 5432
						database = "postgres_updated_db"
					})

					auth = jsonencode({
						password = "postgres_updated_password"
					})
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.mysql_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.mysql_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.mysql_config", "config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.mysql_config", "auth"),

					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.postgres_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.postgres_config", "connection_id"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.postgres_config", "config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.postgres_config", "auth"),
				),
			},
		},
	})
}

func testFivetranConnectionConfigResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		connectionId := rs.Primary.Attributes["connection_id"]
		if connectionId == "" {
			return fmt.Errorf("connection_id is not set")
		}

		_, err := client.NewConnectionDetails().ConnectionID(connectionId).Do(context.Background())
		if err != nil {
			return fmt.Errorf("connection %s not found: %w", connectionId, err)
		}

		return nil
	}
}

func testFivetranConnectionConfigResourceUpdate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		connectionId := rs.Primary.Attributes["connection_id"]
		if connectionId == "" {
			return fmt.Errorf("connection_id is not set")
		}

		_, err := client.NewConnectionDetails().ConnectionID(connectionId).Do(context.Background())
		if err != nil {
			return fmt.Errorf("connection %s not found: %w", connectionId, err)
		}

		return nil
	}
}

func testFivetranConnectionConfigResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_connector" && rs.Type != "fivetran_connection_config" {
			continue
		}

		var connectionId string
		if rs.Type == "fivetran_connector" {
			connectionId = rs.Primary.ID
		} else if rs.Type == "fivetran_connection_config" {
			connectionId = rs.Primary.Attributes["connection_id"]
		}

		if connectionId == "" {
			continue
		}

		response, err := client.NewConnectionDetails().ConnectionID(connectionId).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			return err
		}

		if err == nil && !strings.HasPrefix(response.Code, "NotFound_") {
			return fmt.Errorf(`
			Connection %s still exists after deletion.
			Expected response.Code: 'NotFound_Connection'.
			Actual response.Code was: '%s'.
			response.Message: '%s'`, connectionId, response.Code, response.Message)
		}
	}

	return nil
}
