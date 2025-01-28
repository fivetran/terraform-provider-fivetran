package datasources_test

import (
    "net/http"
    "testing"

    "github.com/fivetran/go-fivetran/tests/mock"
    tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
    QuickstartPackagesDataSourceMockGetHandler *mock.Handler
    QuickstartPackagesDataSourceMockData       map[string]interface{}
)

const (
    QuickstartPackagesMappingResponse = `
    {
    "items": [
      {
        "id": "package_definition_id",
        "name": "package_definition_name",
        "version": "version",
        "connector_types": [
          "string"
        ],
        "output_model_names": [
          "string"
        ]
      },
      {
        "id": "package_definition_id_2",
        "name": "package_definition_name_2",
        "version": "version_2",
        "connector_types": [
          "string_2"
        ],
        "output_model_names": [
          "string_2"
        ]
      }
    ],
    "next_cursor": null
  }`
)

func setupMockClientQuickstartPackagesDataSourceConfigMapping(t *testing.T) {
    tfmock.MockClient().Reset()

    QuickstartPackagesDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformations/package-metadata").ThenCall(
        func(req *http.Request) (*http.Response, error) {
            QuickstartPackagesDataSourceMockData = tfmock.CreateMapFromJsonString(t, QuickstartPackagesMappingResponse)
            return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", QuickstartPackagesDataSourceMockData), nil
        },
    )
}

func TestDataSourceQuickstartPackagesMappingMock(t *testing.T) {
    step1 := resource.TestStep{
        Config: `
        data "fivetran_quickstart_packages" "test_quickstart_package" {
            provider = fivetran-provider
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, QuickstartPackagesDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, QuickstartPackagesDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.0.id", "package_definition_id"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.0.name", "package_definition_name"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.0.version", "version"),

            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.1.id", "package_definition_id_2"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.1.name", "package_definition_name_2"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_packages.test_quickstart_package", "packages.1.version", "version_2"),
        ),
    }

    resource.Test(
        t,
        resource.TestCase{
            PreCheck: func() {
                setupMockClientQuickstartPackagesDataSourceConfigMapping(t)
            },
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
