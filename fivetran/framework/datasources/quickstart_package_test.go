package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	quickstartPackageMappingResponse = `
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
  }
    `
)

var (
	quickstartPackageDataSourceMockGetHandler *mock.Handler

	quickstartPackageDataSourceMockData map[string]interface{}
)

func setupMockClientQuickstartPackageDataSourceConfigMapping(t *testing.T) {
	tfmock.MockClient().Reset()

	quickstartPackageDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/transformations/package-metadata/package_definition_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			quickstartPackageDataSourceMockData = tfmock.CreateMapFromJsonString(t, quickstartPackageMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", quickstartPackageDataSourceMockData), nil
		},
	)
}

func TestDataSourceQuickstartPackageConfigMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
        data "fivetran_quickstart_package" "test" {
            provider = fivetran-provider
            id = "package_definition_id"
        }`,

        Check: resource.ComposeAggregateTestCheckFunc(
            func(s *terraform.State) error {
                tfmock.AssertEqual(t, quickstartPackageDataSourceMockGetHandler.Interactions, 1)
                tfmock.AssertNotEmpty(t, quickstartPackageDataSourceMockData)
                return nil
            },
            resource.TestCheckResourceAttr("data.fivetran_quickstart_package.test", "id", "package_definition_id"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_package.test", "name", "package_definition_name"),
            resource.TestCheckResourceAttr("data.fivetran_quickstart_package.test", "version", "version"),
        ),
    }

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientQuickstartPackageDataSourceConfigMapping(t)
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
