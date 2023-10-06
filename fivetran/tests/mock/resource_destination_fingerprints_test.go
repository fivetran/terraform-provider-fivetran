package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceDestinationFingerprintsMock(t *testing.T) {

	var getDestinationHandler *mock.Handler
	var getHandler *mock.Handler
	var postHandler *mock.Handler
	var deleteHandlers []*mock.Handler

	var data map[string]interface{}

	createDeleteHandler := func(id string) *mock.Handler {
		return mockClient.When(http.MethodDelete, "/v1/destinations/destination_id/fingerprints/"+id).ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
	}

	setupConnectorFingerprintsDatasourceMock := func() {
		mockClient.Reset()

		getInteraction := 0
		postInteraction := 0

		getDestinationHandler = mockClient.When(http.MethodGet, "/v1/destinations/destination_id").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				data = createMapFromJsonString(t, `
					{
						"id": "destination_id"
					}
					`)
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)

		getHandler = mockClient.When(http.MethodGet, "/v1/destinations/destination_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if getInteraction == 0 {
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
						]	
					}
					`)
				}
				if getInteraction >= 1 && getInteraction < 5 {
					data = createMapFromJsonString(t, `
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
				}
				if getInteraction >= 5 {
					data = createMapFromJsonString(t, `
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
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)

		postHandler = mockClient.When(http.MethodPost, "/v1/destinations/destination_id/fingerprints").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if postInteraction == 0 {
					assertKeyExistsAndHasValue(t, body, "hash", "hash2")
					assertKeyExistsAndHasValue(t, body, "public_key", "public_key2")
					data = createMapFromJsonString(t, `
						{
							"hash": "hash2",
							"public_key": "public_key2",
							"validated_by": "validated_by2",
							"validated_date": "validated_date2"
						}
						`)
				}
				if postInteraction == 1 {
					assertKeyExistsAndHasValue(t, body, "hash", "hash3")
					assertKeyExistsAndHasValue(t, body, "public_key", "public_key3")
					data = createMapFromJsonString(t, `
						{
							"hash": "hash3",
							"public_key": "public_key3",
							"validated_by": "validated_by3",
							"validated_date": "validated_date3"
						}
						`)
				}
				postInteraction = postInteraction + 1
				return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", data), nil
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
			PreCheck:  setupConnectorFingerprintsDatasourceMock,
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, deleteHandlers[0].Interactions, 1)
				assertEqual(t, deleteHandlers[1].Interactions, 1)
				assertEqual(t, deleteHandlers[2].Interactions, 1)
				assertEqual(t, deleteHandlers[3].Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: `
					resource "fivetran_destination_fingerprints" "test" {
						provider = fivetran-provider
						destination_id = "destination_id"
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
							assertEqual(t, getDestinationHandler.Interactions, 1)
							assertEqual(t, postHandler.Interactions, 1)
							assertEqual(t, deleteHandlers[0].Interactions, 1)
							assertEqual(t, getHandler.Interactions, 2)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.#", "2"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.hash", "hash1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.validated_date", "validated_date1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.hash", "hash2"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.public_key", "public_key2"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.validated_by", "validated_by2"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.validated_date", "validated_date2"),
					),
				},
				{
					Config: `
					resource "fivetran_destination_fingerprints" "test" {
						provider = fivetran-provider
						destination_id = "destination_id"
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
							assertEqual(t, getDestinationHandler.Interactions, 1)
							assertEqual(t, postHandler.Interactions, 2)
							assertEqual(t, deleteHandlers[2].Interactions, 1)
							assertEqual(t, getHandler.Interactions, 6)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.#", "2"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.hash", "hash1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.0.validated_date", "validated_date1"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.hash", "hash3"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.public_key", "public_key3"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.validated_by", "validated_by3"),
						resource.TestCheckResourceAttr("fivetran_destination_fingerprints.test", "fingerprint.1.validated_date", "validated_date3"),
					),
				},
			},
		},
	)
}
