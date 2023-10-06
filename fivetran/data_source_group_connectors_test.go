package fivetran_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataSourceGroupConnectorsMappingMock(t *testing.T) {
	t.Skip("Test is created for local debugging")
	resource.Test(
		t,
		resource.TestCase{
			Providers: testProviders,
			Steps: []resource.TestStep{
				{
					Config: `
					data "fivetran_group_connectors" "test_group_connectors" {
						provider = fivetran-provider
						id = "group"
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.fivetran_group_connectors.test_group_connectors", "connectors.#", "9"),
					),
				},
			},
		},
	)
}
