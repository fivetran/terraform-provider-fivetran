package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceDestinationCertificatesMock(t *testing.T) {

	var getDestinationHandler *mock.Handler
	var getHandler *mock.Handler
	var postHandler *mock.Handler
	var deleteHandlers []*mock.Handler

	var data map[string]interface{}

	createDeleteHandler := func(id string) *mock.Handler {
		return mockClient.When(http.MethodDelete, "/v1/destinations/destination_id/certificates/"+id).ThenCall(
			func(req *http.Request) (*http.Response, error) {
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
			},
		)
	}

	setup := func() {
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

		getHandler = mockClient.When(http.MethodGet, "/v1/destinations/destination_id/certificates").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if getInteraction == 0 {
					data = createMapFromJsonString(t, `
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
								"name": "name1",
								"type": "type1",
								"sha1": "sha11",
								"sha256": "sha2561",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
							},
							{
								"hash": "hash2",
								"public_key": "public_key2",
								"name": "name2",
								"type": "type2",
								"sha1": "sha12",
								"sha256": "sha2562",
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
								"name": "name1",
								"type": "type1",
								"sha1": "sha11",
								"sha256": "sha2561",
								"validated_by": "validated_by1",
								"validated_date": "validated_date1"
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
				getInteraction = getInteraction + 1
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", data), nil
			},
		)

		postHandler = mockClient.When(http.MethodPost, "/v1/destinations/destination_id/certificates").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				if postInteraction == 0 {
					assertKeyExistsAndHasValue(t, body, "hash", "hash2")
					assertKeyExistsAndHasValue(t, body, "encoded_cert", "cert2")
					data = createMapFromJsonString(t, `
						{
							"hash": "hash2",
							"public_key": "public_key2",
							"name": "name2",
							"type": "type2",
							"sha1": "sha12",
							"sha256": "sha2562",
							"validated_by": "validated_by2",
							"validated_date": "validated_date2"
						}
						`)
				}
				if postInteraction == 1 {
					assertKeyExistsAndHasValue(t, body, "hash", "hash3")
					assertKeyExistsAndHasValue(t, body, "encoded_cert", "cert3")
					data = createMapFromJsonString(t, `
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
			PreCheck:  setup,
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
					resource "fivetran_destination_certificates" "test" {
						provider = fivetran-provider
						destination_id = "destination_id"
						certificate {
							hash ="hash1"
							encoded_cert = "cert1"
						}
						certificate {
							hash ="hash2"
							encoded_cert = "cert2"
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
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.#", "2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.hash", "hash1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.encoded_cert", "cert1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.name", "name1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.type", "type1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.sha1", "sha11"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.sha256", "sha2561"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.validated_date", "validated_date1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.hash", "hash2"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.encoded_cert", "cert2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.name", "name2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.type", "type2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.sha1", "sha12"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.sha256", "sha2562"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.public_key", "public_key2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.validated_by", "validated_by2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.validated_date", "validated_date2"),
					),
				},
				{
					Config: `
					resource "fivetran_destination_certificates" "test" {
						provider = fivetran-provider
						destination_id = "destination_id"
						certificate {
							hash ="hash1"
							encoded_cert = "cert1"
						}
						certificate {
							hash ="hash3"
							encoded_cert = "cert3"
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
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.#", "2"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.hash", "hash1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.encoded_cert", "cert1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.name", "name1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.type", "type1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.sha1", "sha11"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.sha256", "sha2561"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.public_key", "public_key1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.validated_by", "validated_by1"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.0.validated_date", "validated_date1"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.hash", "hash3"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.encoded_cert", "cert3"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.name", "name3"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.type", "type3"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.sha1", "sha13"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.sha256", "sha2563"),

						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.public_key", "public_key3"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.validated_by", "validated_by3"),
						resource.TestCheckResourceAttr("fivetran_destination_certificates.test", "certificate.1.validated_date", "validated_date3"),
					),
				},
			},
		},
	)
}
