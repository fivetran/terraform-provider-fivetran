package resources_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	sdkPackageGetHandler    *mock.Handler
	sdkPackagePostHandler   *mock.Handler
	sdkPackagePatchHandler  *mock.Handler
	sdkPackageDeleteHandler *mock.Handler
	sdkPackageData          map[string]interface{}
)

func createSdkPackageTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "code.zip")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	return path
}

func sdkPackageSha256(content string) string {
	digest := sha256.Sum256([]byte(content))
	return hex.EncodeToString(digest[:])
}

func sdkPackageResponseData(hash string, updatedAt string) map[string]interface{} {
	return map[string]interface{}{
		"id":               "happy_harmony",
		"connection_id":    nil,
		"created_by":       "user_1",
		"last_updated_by":  "user_1",
		"created_at":       "2024-01-01T00:00:00.000000Z",
		"updated_at":       updatedAt,
		"file_sha256_hash": hash,
	}
}

func onSdkPackagePost(t *testing.T, req *http.Request, hash string) (*http.Response, error) {
	tfmock.AssertEmpty(t, sdkPackageData)
	sdkPackageData = sdkPackageResponseData(hash, "2024-01-01T00:00:00.000000Z")
	return tfmock.FivetranSuccessResponse(t, req, http.StatusCreated, "Success", sdkPackageData), nil
}

func onSdkPackagePatch(t *testing.T, req *http.Request, hash string) (*http.Response, error) {
	tfmock.AssertNotEmpty(t, sdkPackageData)
	sdkPackageData = sdkPackageResponseData(hash, "2024-01-02T00:00:00.000000Z")
	return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageData), nil
}

func setupMockClientSdkPackageResource(t *testing.T, createHash, updateHash string) {
	tfmock.MockClient().Reset()
	sdkPackageData = nil

	sdkPackagePostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/connector-sdk/packages").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onSdkPackagePost(t, req, createHash)
		},
	)

	sdkPackageGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, sdkPackageData)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageData), nil
		},
	)

	sdkPackagePatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onSdkPackagePatch(t, req, updateHash)
		},
	)

	sdkPackageDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			sdkPackageData = nil
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Package deleted", nil), nil
		},
	)
}

func TestResourceConnectorSdkPackageMock(t *testing.T) {
	createContent := "fake-zip-content-v1"
	createHash := sdkPackageSha256(createContent)
	createPath := createSdkPackageTempFile(t, createContent)

	updateContent := "fake-zip-content-v2"
	updateHash := sdkPackageSha256(updateContent)
	updatePath := createSdkPackageTempFile(t, updateContent)

	step1 := resource.TestStep{
		Config: `
			resource "fivetran_connector_sdk_package" "test_package" {
				provider  = fivetran-provider
				file_path = "` + createPath + `"
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, sdkPackagePostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, sdkPackageData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test_package", "id", "happy_harmony"),
			resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test_package", "file_path", createPath),
			resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test_package", "file_sha256_hash", createHash),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_connector_sdk_package" "test_package" {
				provider  = fivetran-provider
				file_path = "` + updatePath + `"
			}`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, sdkPackagePatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test_package", "file_sha256_hash", updateHash),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientSdkPackageResource(t, createHash, updateHash)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, sdkPackageDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, sdkPackageData)
				return nil
			},
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}

func TestResourceConnectorSdkPackageDeleteNotFoundMock(t *testing.T) {
	createContent := "fake-zip-content"
	createHash := sdkPackageSha256(createContent)
	createPath := createSdkPackageTempFile(t, createContent)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				tfmock.MockClient().Reset()
				sdkPackageData = nil

				tfmock.MockClient().When(http.MethodPost, "/v1/connector-sdk/packages").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return onSdkPackagePost(t, req, createHash)
					},
				)

				tfmock.MockClient().When(http.MethodGet, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", sdkPackageData), nil
					},
				)

				// DELETE returns 404 with NotFound_ConnectorSdkPackage code — should be treated as success
				tfmock.MockClient().When(http.MethodDelete, "/v1/connector-sdk/packages/happy_harmony").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return mock.NewResponse(req, http.StatusNotFound, `{
							"code": "NotFound_ConnectorSdkPackage",
							"message": "Package not found"
						}`), nil
					},
				)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: `
						resource "fivetran_connector_sdk_package" "test_package" {
							provider  = fivetran-provider
							file_path = "` + createPath + `"
						}`,
				},
			},
		},
	)
}
