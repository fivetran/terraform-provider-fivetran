package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceDestinationFingerprintsMock(t *testing.T) {

	var getHandler *mock.Handler
	var data map[string]interface{}

	setupConnectorFingerprintsDatasourceMock := func() {
		mockClient.Reset()
		getHandler = mockClient.When(http.MethodGet, "/v1/destinations/destination_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				cursor := req.URL.Query().Get("cursor")
				if cursor == "" {
					data = createMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash0",
								"public_key": "public_key0",
								"validated_by": "validated_by0",
								"validated_date": "validated_date0"
							},
							{
								"hash": "hash1",
								"public_key": "public_key1",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
							}
						],
						"next_cursor": "next_cursor"	
					}
					`)
				} else if cursor == "next_cursor" {
					data = createMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash2",
								"public_key": "public_key2",
								"validated_by": "validated_by2",
								"validated_date": "validated_date2"
							},
							{
								"hash": "hash3",
								"public_key": "public_key3",
								"validated_by": "validated_by3",
								"validated_date": "validated_date3"
							}
						]
					}
					`)
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 setupConnectorFingerprintsDatasourceMock,
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
					data "fivetran_destination_fingerprints" "test" {
						provider = fivetran-provider
						id = "destination_id"
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, getHandler.Interactions, 4)
							return nil
						},
						resource.TestCheckResourceAttr("data.fivetran_destination_fingerprints.test", "fingerprints.#", "4"),
						resource.TestCheckResourceAttr("data.fivetran_destination_fingerprints.test", "fingerprints.0.hash", "hash0"),
						resource.TestCheckResourceAttr("data.fivetran_destination_fingerprints.test", "fingerprints.0.public_key", "public_key0"),
						resource.TestCheckResourceAttr("data.fivetran_destination_fingerprints.test", "fingerprints.0.validated_by", "validated_by0"),
						resource.TestCheckResourceAttr("data.fivetran_destination_fingerprints.test", "fingerprints.0.validated_date", "validated_date0"),
					),
				},
			},
		},
	)
}
