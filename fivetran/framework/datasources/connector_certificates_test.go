package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataSourceConnectorCertificatesMock(t *testing.T) {

	var getHandler *mock.Handler
	var data map[string]interface{}

	setupConnectorFingerprintsDatasourceMock := func() {
		tfmock.MockClient().Reset()

		getHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connections/connector_id/certificates").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				cursor := req.URL.Query().Get("cursor")
				if cursor == "" {
					data = tfmock.CreateMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash0",
								"public_key": "public_key0",
								"name": "name0",
								"type": "type0",
								"sha1": "sha10",
								"sha256": "sha2560",
								"validated_by": "validated_by0",
								"validated_date": "validated_date0"
							},
							{
								"hash": "hash1",
								"public_key": "public_key1",
								"name": "name1",
								"type": "type1",
								"sha1": "sha11",
								"sha256": "sha2561",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
							}
						],
						"next_cursor": "next_cursor"	
					}
					`)
				} else if cursor == "next_cursor" {
					data = tfmock.CreateMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash2",
								"public_key": "public_key2",
								"name": "name2",
								"type": "type2",
								"sha1": "sha12",
								"sha256": "sha2562",
								"validated_by": "validated_by2",
								"validated_date": "validated_date2"
							},
							{
								"hash": "hash3",
								"public_key": "public_key3",
								"name": "name3",
								"type": "type3",
								"sha1": "sha13",
								"sha256": "sha2563",
								"validated_by": "validated_by3",
								"validated_date": "validated_date3"
							}
						]
					}
					`)
				}
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 setupConnectorFingerprintsDatasourceMock,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
					data "fivetran_connector_certificates" "test" {
						provider = fivetran-provider
						id = "connector_id"
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, getHandler.Interactions, 2)
							return nil
						},
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "id", "connector_id"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "connector_id", "connector_id"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.#", "4"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.hash", "hash0"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.name", "name0"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.type", "type0"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.sha1", "sha10"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.sha256", "sha2560"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.public_key", "public_key0"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.validated_by", "validated_by0"),
						resource.TestCheckResourceAttr("data.fivetran_connector_certificates.test", "certificates.0.validated_date", "validated_date0"),
					),
				},
			},
		},
	)
}
