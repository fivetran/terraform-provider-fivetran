package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceConnectionFingerprintsMock(t *testing.T) {

	var connectionFingerprintsGetHandler *mock.Handler
	var connectionFingerprintsData map[string]interface{}

	setupConnectionFingerprintsDatasourceMock := func() {
		tfmock.MockClient().Reset()

		connectionFingerprintsGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connection_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				cursor := req.URL.Query().Get("cursor")
				if cursor == "" {
					connectionFingerprintsData = tfmock.CreateMapFromJsonString(t, `
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
					connectorFingerprintsData = tfmock.CreateMapFromJsonString(t, `
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
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", connectionFingerprintsData), nil
			},
		)
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 setupConnectionFingerprintsDatasourceMock,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
					data "fivetran_connection_fingerprints" "test" {
						provider = fivetran-provider
						id = "connection_id"
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, connectorFingerprintsGetHandler.Interactions, 2)
							return nil
						},
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "id", "connection_id"),
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "fingerprints.#", "4"),
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "fingerprints.0.hash", "hash0"),
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "fingerprints.0.public_key", "public_key0"),
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "fingerprints.0.validated_by", "validated_by0"),
						resource.TestCheckResourceAttr("data.fivetran_connection_fingerprints.test", "fingerprints.0.validated_date", "validated_date0"),
					),
				},
			},
		},
	)
}
