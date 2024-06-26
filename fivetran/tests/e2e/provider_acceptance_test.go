package e2e_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestNewDatasourceE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "fivetran_group" "testgroup" {
					provider = fivetran-provider
					id = "%v"
				}
				`, PredefinedGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}
