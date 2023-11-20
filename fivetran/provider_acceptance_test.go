package fivetran_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestNewDatasourceE2E(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV5ProviderFactories: protoV5ProviderFactory,
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
