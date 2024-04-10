package resources_test

import (
	"net/http"
	"testing"
	
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceConnectorFingerprintsMock(t *testing.T) {

	var getHandler *mock.Handler
	var postHandler *mock.Handler
	var deleteHandlers []*mock.Handler

	var data map[string]interface{}

	createDeleteHandler := func(id string) *mock.Handler {
		return tfmock.MockClient().When(http.MethodDelete, "/v1/connectors/connector_id/fingerprints/"+id).ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
	}

	setupConnectorFingerprintsDatasourceMock := func() {
		tfmock.MockClient().Reset()

		getInteraction := 0
		postInteraction := 0

		getHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connectors/connector_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
					data = tfmock.CreateMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash1",
								"public_key": "public_key1",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
							},
							{
								"hash": "hash2",
								"public_key": "public_key2",
								"validated_by": "validated_by2",
								"validated_date": "validated_date2"
							}
						]	
					}
					`)

				if getInteraction >= 3 {
					data = tfmock.CreateMapFromJsonString(t, `
					{
						"items":[
							{
								"hash": "hash1",
								"public_key": "public_key1",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
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

				getInteraction = getInteraction + 1
				return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)

		postHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connectors/connector_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if postInteraction == 0 {
					data = tfmock.CreateMapFromJsonString(t, `
						{
							"hash": "hash1",
							"public_key": "public_key1",
							"validated_by": "validated_by1",
							"validated_date": "validated_date1"
						}
						`)
				}
				if postInteraction == 1 {
					data = tfmock.CreateMapFromJsonString(t, `
						{
							"hash": "hash2",
							"public_key": "public_key2",
							"validated_by": "validated_by2",
							"validated_date": "validated_date2"
						}
						`)
				}
				if postInteraction == 2 {
					data = tfmock.CreateMapFromJsonString(t, `
						{
							"hash": "hash3",
							"public_key": "public_key3",
							"validated_by": "validated_by3",
							"validated_date": "validated_date3"
						}
						`)
				}
				postInteraction = postInteraction + 1
				return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", data), nil
			},
		)

		deleteHandlers = append(deleteHandlers, createDeleteHandler("hash0"))
		deleteHandlers = append(deleteHandlers, createDeleteHandler("hash1"))
		deleteHandlers = append(deleteHandlers, createDeleteHandler("hash2"))
		deleteHandlers = append(deleteHandlers, createDeleteHandler("hash3"))
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck:                 setupConnectorFingerprintsDatasourceMock,
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, deleteHandlers[1].Interactions, 1)
				tfmock.AssertEqual(t, deleteHandlers[2].Interactions, 1)
				tfmock.AssertEqual(t, deleteHandlers[3].Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: `
					resource "fivetran_connector_fingerprints" "test" {
						provider = fivetran-provider
						connector_id = "connector_id"
						fingerprint {
							hash ="hash1"
							public_key = "public_key1"
						}
						fingerprint {
							hash ="hash2"
							public_key = "public_key2"
						}
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, postHandler.Interactions, 2)
							tfmock.AssertEqual(t, getHandler.Interactions, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.#", "2"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.hash", "hash1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.validated_date", "validated_date1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.hash", "hash2"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.public_key", "public_key2"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.validated_by", "validated_by2"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.validated_date", "validated_date2"),
					),
				},
				{
					Config: `
					resource "fivetran_connector_fingerprints" "test" {
						provider = fivetran-provider
						connector_id = "connector_id"
						fingerprint {
							hash ="hash1"
							public_key = "public_key1"
						}
						fingerprint {
							hash ="hash3"
							public_key = "public_key3"
						}
					}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							tfmock.AssertEqual(t, postHandler.Interactions, 3)
							tfmock.AssertEqual(t, deleteHandlers[2].Interactions, 1)
							tfmock.AssertEqual(t, getHandler.Interactions, 4)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.#", "2"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.hash", "hash1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.0.validated_date", "validated_date1"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.hash", "hash3"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.public_key", "public_key3"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.validated_by", "validated_by3"),
						resource.TestCheckResourceAttr("fivetran_connector_fingerprints.test", "fingerprint.1.validated_date", "validated_date3"),
					),
				},
			},
		},
	)
}
