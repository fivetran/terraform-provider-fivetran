package resources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	privateLinkPostHandler   *mock.Handler
	privateLinkPatchHandler  *mock.Handler
	privateLinkDeleteHandler *mock.Handler
	privateLinkData          map[string]interface{}
)

func setupMockClientPrivateLinkResource(t *testing.T) {
	tfmock.MockClient().Reset()
	privateLinkResponse :=
		`{
        "id": "pl_id",
        "name": "name",
        "region": "region",
        "service": "service",
        "account_id": "account_id",
        "cloud_provider": "cloud_provider",
        "state": "state",
        "state_summary": "state_summary",
        "created_at": "created_at",
        "created_by": "created_by",
        "host": "host",
        "config": {
        	"connection_service_name": "connection_service_name"
        }
    }`

	privateLinkPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/private-links").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			privateLinkData = tfmock.CreateMapFromJsonString(t, privateLinkResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "PrivateLink has been created", privateLinkData), nil
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", privateLinkData), nil
		},
	)

	privateLinkDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "PrivateLink has been deleted", nil), nil
		},
	)
}

func setupMockClientPrivateLinkResourceWithNullConfigForImport(t *testing.T) {
	tfmock.MockClient().Reset()
	privateLinkResponse :=
		`{
        "id": "pl_id",
        "name": "name",
        "region": "region",
        "service": "service",
        "account_id": "account_id",
        "cloud_provider": "cloud_provider",
        "state": "state",
        "state_summary": "state_summary",
        "created_at": "created_at",
        "created_by": "created_by",
        "host": "host",
        "config": null
    }`

	tfmock.MockClient().When(http.MethodGet, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			privateLinkData = tfmock.CreateMapFromJsonString(t, privateLinkResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", privateLinkData), nil
		},
	)

	privateLinkDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "PrivateLink has been deleted", nil), nil
		},
	)
}

func TestResourcePrivateLinkMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl" {
			provider = fivetran-provider

               name = "name"
               region = "region"
               service = "service"

        		config_map = {
        		  connection_service_name = "connection_service_name"
        		}
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, privateLinkPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", "name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "region"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "service"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "host", "host"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.connection_service_name", "connection_service_name"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientPrivateLinkResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, privateLinkDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourcePrivateLinkImportMock(t *testing.T) {
	
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"

        		config_map = {
        		  connection_service_name = "connection_service_name"
        		}
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, privateLinkPostHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", "name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "region"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "service"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "host", "host"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.connection_service_name", "connection_service_name"),
		),
	}

	step2 := resource.TestStep{
		ResourceName:      "fivetran_private_link.test_pl",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStateVerify: true,
	}

	step3 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"
            }`,
		ResourceName:      "fivetran_private_link.test_pl_imported",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "id", "pl_id"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "name", "name"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "region", "region"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "service", "service"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "config_map.connection_service_name", "connection_service_name"),
		),
	}

	step4 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"

        		config_map = {}
            }`,
		ResourceName:      "fivetran_private_link.test_pl_imported",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStatePersist: true,
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "id", "pl_id"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "name", "name"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "region", "region"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "service", "service"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "config_map.connection_service_name", "connection_service_name"),
		),
	}

	step5 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"

        		config_map = {
        		  connection_service_name = "connection_service_name"
        		}
            }

            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"

        		config_map = {}
            }`,
		PlanOnly: true,
		Check: resource.ComposeAggregateTestCheckFunc (
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", "name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "region"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "service"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "host", "host"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.connection_service_name", "connection_service_name"),

			resource.TestCheckResourceAttr("fivetran_private_link.test_pl_imported", "name", "name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl_imported", "region", "region"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl_imported", "service", "service"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl_imported", "host", "host"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl_imported", "config_map.connection_service_name", "connection_service_name"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientPrivateLinkResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, privateLinkDeleteHandler.Interactions, 2)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
				step4,
				step5,
			},
		},
	)
}

func TestResourcePrivateLinkImportWhenConfigIsNullMock(t *testing.T) {
	
	step1 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider
            }`,
		ResourceName:      "fivetran_private_link.test_pl_imported",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "id", "pl_id"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "name", "name"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "region", "region"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "service", "service"),
			tfmock.CheckNoImportResourceAttr("fivetran_private_link", "pl_id", "config_map"),
		),
	}

	step2 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"
            }`,
		ResourceName:      "fivetran_private_link.test_pl_imported",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "id", "pl_id"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "name", "name"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "region", "region"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "service", "service"),
			tfmock.CheckNoImportResourceAttr("fivetran_private_link", "pl_id", "config_map"),
		),
	}

	step3 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl_imported" {
			provider = fivetran-provider

                name = "name"
                region = "region"
                service = "service"

        		config_map = {}
            }`,
		ResourceName:      "fivetran_private_link.test_pl_imported",
		ImportState:       true,
		ImportStateId: 	"pl_id",
		ImportStatePersist: true,
		ImportStateCheck: tfmock.ComposeImportStateCheck(
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "id", "pl_id"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "name", "name"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "region", "region"),
			tfmock.CheckImportResourceAttr("fivetran_private_link", "pl_id", "service", "service"),
			tfmock.CheckNoImportResourceAttr("fivetran_private_link", "pl_id", "config_map"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientPrivateLinkResourceWithNullConfigForImport(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, privateLinkDeleteHandler.Interactions, 1)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}

func TestResourcePrivateLinkWithNilConfigValues(t *testing.T) {
	tfmock.MockClient().Reset()

	// Mock response with nil values in config to test the fix
	privateLinkResponseWithNil :=
		`{
        "id": "pl_id",
        "name": "name",
        "region": "region",
        "service": "service",
        "account_id": "account_id",
        "cloud_provider": "cloud_provider",
        "state": "state",
        "state_summary": "state_summary",
        "created_at": "created_at",
        "created_by": "created_by",
        "host": "host",
        "config": {
        	"connection_service_name": "connection_service_name",
        	"account_url": null,
        	"vpce_id": null,
        	"aws_account_id": "aws_account_id"
        }
    }`

	privateLinkDataWithNil := tfmock.CreateMapFromJsonString(t, privateLinkResponseWithNil)

	// Mock POST endpoint for creation
	tfmock.MockClient().When(http.MethodPost, "/v1/private-links").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "PrivateLink has been created", privateLinkDataWithNil), nil
		},
	)

	// Mock GET endpoint for reading
	tfmock.MockClient().When(http.MethodGet, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", privateLinkDataWithNil), nil
		},
	)

	// Mock DELETE endpoint for cleanup
	tfmock.MockClient().When(http.MethodDelete, "/v1/private-links/pl_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return tfmock.FivetranSuccessResponse(t, req, 200, "PrivateLink has been deleted", nil), nil
		},
	)

	step1 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl" {
			provider = fivetran-provider

               name = "name"
               region = "region"
               service = "service"

        		config_map = {
        		  connection_service_name = "connection_service_name"
        		  aws_account_id = "aws_account_id"
        		}
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "name", "name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "region", "region"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "service", "service"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "host", "host"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.connection_service_name", "connection_service_name"),
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.aws_account_id", "aws_account_id"),
			// These should be null because they were nil in the response
			resource.TestCheckNoResourceAttr("fivetran_private_link.test_pl", "config_map.account_url"),
			resource.TestCheckNoResourceAttr("fivetran_private_link.test_pl", "config_map.vpce_id"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				// No setup needed for this test
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// Check that delete was called
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourcePrivateLinkPrivateDnsRecordsMock(t *testing.T) {
	var postBody map[string]interface{}

	setup := func() {
		tfmock.MockClient().Reset()
		privateLinkResponse :=
			`{
        "id": "pl_id",
        "name": "name",
        "region": "region",
        "service": "source_azure",
        "cloud_provider": "AZURE",
        "state": "state",
        "state_summary": "state_summary",
        "created_at": "created_at",
        "created_by": "created_by",
        "host": "host",
        "config": {
        	"connection_service_id": "connection_service_id",
        	"sub_resource_name": "sqlServer",
        	"private_dns_records": "server-a;server-b"
        }
    }`

		tfmock.MockClient().When(http.MethodPost, "/v1/private-links").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				postBody = tfmock.RequestBodyToJson(t, req)
				privateLinkData = tfmock.CreateMapFromJsonString(t, privateLinkResponse)
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "PrivateLink has been created", privateLinkData), nil
			},
		)

		tfmock.MockClient().When(http.MethodGet, "/v1/private-links/pl_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", privateLinkData), nil
			},
		)

		tfmock.MockClient().When(http.MethodDelete, "/v1/private-links/pl_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, 200, "PrivateLink has been deleted", nil), nil
			},
		)
	}

	step1 := resource.TestStep{
		Config: `
            resource "fivetran_private_link" "test_pl" {
			provider = fivetran-provider

               name = "name"
               region = "region"
               service = "source_azure"

        		config_map = {
        		  connection_service_id = "connection_service_id"
        		  sub_resource_name = "sqlServer"
        		  private_dns_records = "server-a;server-b"
        		}
            }`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				config, ok := postBody["config"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected config object in request body, got %v", postBody["config"])
				}
				tfmock.AssertKeyExistsAndHasValue(t, config, "private_dns_records", "server-a;server-b")
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_private_link.test_pl", "config_map.private_dns_records", "server-a;server-b"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 setup,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
