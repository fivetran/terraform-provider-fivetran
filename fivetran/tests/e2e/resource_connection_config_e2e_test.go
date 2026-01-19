package e2e_test

import (
	"context"
	"fmt"
	"regexp"
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

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_test"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
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

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_test"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
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
				ImportStateVerifyIgnore: []string{"config", "auth"},
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

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_config_only"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
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

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_auth_only"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

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

func TestResourceConnectionConfigWithNewFieldsE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_new_fields"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_new_fields"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						user = "test_user"
						host = "test.example.com"
						port = 5432
						database = "test_db"
					})

					auth = jsonencode({
						password = "test_password"
					})

					run_setup_tests = false
					trust_certificates = true
					trust_fingerprints = true
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
					resource.TestCheckResourceAttrSet("fivetran_connection_config.test_config", "connection_id"),
					resource.TestCheckResourceAttr("fivetran_connection_config.test_config", "run_setup_tests", "false"),
					resource.TestCheckResourceAttr("fivetran_connection_config.test_config", "trust_certificates", "true"),
					resource.TestCheckResourceAttr("fivetran_connection_config.test_config", "trust_fingerprints", "true"),
				),
			},
		},
	})
}

func TestResourceConnectionConfigSemanticJSONE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_semantic_json"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_semantic"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						host = "test.example.com"
						port = 5432
						database = "test_db"
						update_method = "QUERY_BASED"
					})

					auth = jsonencode({
						password = "test_password"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_semantic_json"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_semantic"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						database = "test_db"
						port = 5432
						host = "test.example.com"
					})

					auth = jsonencode({
						password = "test_password"
					})

					run_setup_tests = false
				}
		  `,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestResourceConnectionConfigEmptyValidationE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_empty_validation"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_empty_test"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id
				}
		  `,
				ExpectError: regexp.MustCompile("at least one of 'config' or 'auth' must be specified"),
			},
		},
	})
}

func TestResourceConnectionConfigCredentialRotationE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_cred_rotation"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_rotation"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						user = "initial_user"
						host = "test.example.com"
						port = 5432
						database = "test_db"
					})

					auth = jsonencode({
						password = "initial_password"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_cred_rotation"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_rotation"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						user = "initial_user"
						host = "test.example.com"
						port = 5432
						database = "test_db"
					})

					auth = jsonencode({
						password = "rotated_password"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceUpdate(t, "fivetran_connection_config.test_config"),
				),
			},
		},
	})
}

func TestResourceConnectionConfigConfigUpdateOnlyE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectionConfigResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_config_update"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_config_update"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						user = "test_user"
						host = "old-host.example.com"
						port = 5432
						database = "test_db"
					})

					auth = jsonencode({
						password = "test_password"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceCreate(t, "fivetran_connection_config.test_config"),
				),
			},
			{
				Config: `
				resource "fivetran_group" "test_group" {
					provider = fivetran-provider
					name = "test_group_config_update"
			    }

			    resource "fivetran_connection" "test_connection" {
					provider = fivetran-provider
					group_id = fivetran_group.test_group.id
					service = "postgres"

					destination_schema {
						prefix = "postgres_config_update"
					}

					config = jsonencode({
						update_method = "QUERY_BASED"
					})

					run_setup_tests = false
				}

				resource "fivetran_connection_config" "test_config" {
					provider = fivetran-provider
					connection_id = fivetran_connection.test_connection.id

					config = jsonencode({
						update_method = "QUERY_BASED"
						user = "test_user"
						host = "new-host.example.com"
						port = 5433
						database = "new_test_db"
					})

					auth = jsonencode({
						password = "test_password"
					})

					run_setup_tests = false
				}
		  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectionConfigResourceUpdate(t, "fivetran_connection_config.test_config"),
				),
			},
		},
	})
}

func testFivetranConnectionConfigResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_connection" && rs.Type != "fivetran_connection_config" {
			continue
		}

		var connectionId string
		if rs.Type == "fivetran_connection" {
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
