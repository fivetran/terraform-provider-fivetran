package mock

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	sdkPackageMockGetHandler    *mock.Handler
	sdkPackageMockPostHandler   *mock.Handler
	sdkPackageMockPatchHandler  *mock.Handler
	sdkPackageMockDeleteHandler *mock.Handler
	sdkPackageMockData          map[string]interface{}
)

func createTempZipFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "code.zip")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	return filePath
}

func computeSha256(content string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:])
}

func setupMockClientSdkPackage(t *testing.T, fileHash string) {
	mockClient.Reset()

	sdkPackageMockGetHandler = mockClient.When(http.MethodGet, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageMockData), nil
		},
	)

	sdkPackageMockPostHandler = mockClient.When(http.MethodPost, "/v1/connector-sdk/packages").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			sdkPackageMockData = createMapFromJsonString(t, `{
				"id": "happy_harmony",
				"connection_id": null,
				"created_by": "user_1",
				"last_updated_by": "user_1",
				"created_at": "2024-01-01T00:00:00.000000Z",
				"updated_at": "2024-01-01T00:00:00.000000Z",
				"file_sha256_hash": "`+fileHash+`"
			}`)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", sdkPackageMockData), nil
		},
	)

	sdkPackageMockDeleteHandler = mockClient.When(http.MethodDelete, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			sdkPackageMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
		},
	)
}

func TestResourceConnectorSdkPackageCreateAndRead(t *testing.T) {
	fileContent := "fake-zip-content-v1"
	fileHash := computeSha256(fileContent)
	filePath := createTempZipFile(t, fileContent)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientSdkPackage(t, fileHash)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, sdkPackageMockDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: `
					resource "fivetran_connector_sdk_package" "test" {
						provider  = fivetran-provider
						file_path = "` + filePath + `"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, sdkPackageMockPostHandler.Interactions, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test", "id", "happy_harmony"),
						resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test", "file_path", filePath),
						resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test", "file_sha256_hash", fileHash),
					),
				},
			},
		},
	)
}

func TestResourceConnectorSdkPackageUpdate(t *testing.T) {
	fileContentV1 := "fake-zip-content-v1"
	fileHashV1 := computeSha256(fileContentV1)
	filePathV1 := createTempZipFile(t, fileContentV1)

	// Create a second file with different content for the update step
	fileContentV2 := "fake-zip-content-v2-updated"
	fileHashV2 := computeSha256(fileContentV2)
	filePathV2 := createTempZipFile(t, fileContentV2)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()

				sdkPackageMockGetHandler = mockClient.When(http.MethodGet, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageMockData), nil
					},
				)

				sdkPackageMockPostHandler = mockClient.When(http.MethodPost, "/v1/connector-sdk/packages").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						sdkPackageMockData = createMapFromJsonString(t, `{
							"id": "happy_harmony",
							"connection_id": null,
							"created_by": "user_1",
							"last_updated_by": "user_1",
							"created_at": "2024-01-01T00:00:00.000000Z",
							"updated_at": "2024-01-01T00:00:00.000000Z",
							"file_sha256_hash": "`+fileHashV1+`"
						}`)
						return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", sdkPackageMockData), nil
					},
				)

				sdkPackageMockPatchHandler = mockClient.When(http.MethodPatch, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						sdkPackageMockData = createMapFromJsonString(t, `{
							"id": "happy_harmony",
							"connection_id": null,
							"created_by": "user_1",
							"last_updated_by": "user_1",
							"created_at": "2024-01-01T00:00:00.000000Z",
							"updated_at": "2024-01-02T00:00:00.000000Z",
							"file_sha256_hash": "`+fileHashV2+`"
						}`)
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageMockData), nil
					},
				)

				sdkPackageMockDeleteHandler = mockClient.When(http.MethodDelete, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						sdkPackageMockData = nil
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", nil), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, sdkPackageMockDeleteHandler.Interactions, 1)
				return nil
			},
			Steps: []resource.TestStep{
				{
					// Step 1: Create
					Config: `
					resource "fivetran_connector_sdk_package" "test" {
						provider  = fivetran-provider
						file_path = "` + filePathV1 + `"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test", "file_sha256_hash", fileHashV1),
					),
				},
				{
					// Step 2: Update with different file
					Config: `
					resource "fivetran_connector_sdk_package" "test" {
						provider  = fivetran-provider
						file_path = "` + filePathV2 + `"
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, sdkPackageMockPatchHandler.Interactions, 1)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test", "file_sha256_hash", fileHashV2),
					),
				},
			},
		},
	)
}

func TestResourceConnectorSdkPackageDeleteNotFound(t *testing.T) {
	fileContent := "fake-zip-content"
	fileHash := computeSha256(fileContent)
	filePath := createTempZipFile(t, fileContent)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()

				mockClient.When(http.MethodGet, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageMockData), nil
					},
				)

				mockClient.When(http.MethodPost, "/v1/connector-sdk/packages").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						sdkPackageMockData = createMapFromJsonString(t, `{
							"id": "happy_harmony",
							"connection_id": null,
							"created_by": "user_1",
							"last_updated_by": "user_1",
							"created_at": "2024-01-01T00:00:00.000000Z",
							"updated_at": "2024-01-01T00:00:00.000000Z",
							"file_sha256_hash": "`+fileHash+`"
						}`)
						return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", sdkPackageMockData), nil
					},
				)

				// DELETE returns 404 — should be treated as success
				mockClient.When(http.MethodDelete, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						response := mock.NewResponse(req, http.StatusNotFound, `{
							"code": "NotFound",
							"message": "Package not found"
						}`)
						return response, nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
					resource "fivetran_connector_sdk_package" "test" {
						provider  = fivetran-provider
						file_path = "` + filePath + `"
					}
					`,
				},
			},
		},
	)
}
