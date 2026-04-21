package e2e_test

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var sdkPackageResourceConfig = `
	resource "fivetran_connector_sdk_package" "test_pkg" {
		provider  = fivetran-provider
		file_path = "%v"
	}`

// TestResourceConnectorSdkPackageE2E covers create + read + destroy
func TestResourceConnectorSdkPackageE2E(t *testing.T) {
	zipPath := writeTestSdkPackageZip(t, map[string][]byte{
		"connector.py": []byte("# test connector\n"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorSdkPackageResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(sdkPackageResourceConfig, zipPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorSdkPackageResourceCreate(t, "fivetran_connector_sdk_package.test_pkg"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "id"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "file_sha256_hash"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "created_at"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "updated_at"),
					resource.TestCheckResourceAttr("fivetran_connector_sdk_package.test_pkg", "file_path", zipPath),
				),
			},
		},
	})
}

func TestResourceConnectorSdkPackageUpdateE2E(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "code.zip")
	writeZipAtPath(t, zipPath, map[string][]byte{
		"connector.py": []byte("# v1\n"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorSdkPackageResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(sdkPackageResourceConfig, zipPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorSdkPackageResourceCreate(t, "fivetran_connector_sdk_package.test_pkg"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "file_sha256_hash"),
				),
			},
			{
				// Rewrite the file in place with different contents. file_path
				// does not change, but the hash should.
				PreConfig: func() {
					writeZipAtPath(t, zipPath, map[string][]byte{
						"connector.py":         []byte("# v2\n"),
						"connector_helpers.py": []byte("# helpers\n"),
					})
				},
				Config: fmt.Sprintf(sdkPackageResourceConfig, zipPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorSdkPackageResourceCreate(t, "fivetran_connector_sdk_package.test_pkg"),
					resource.TestCheckResourceAttrSet("fivetran_connector_sdk_package.test_pkg", "file_sha256_hash"),
				),
			},
		},
	})
}

// TestResourceConnectorSdkPackageImportE2E verifies import populates state
// from the API
func TestResourceConnectorSdkPackageImportE2E(t *testing.T) {
	zipPath := writeTestSdkPackageZip(t, map[string][]byte{
		"connector.py": []byte("# import test\n"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testFivetranConnectorSdkPackageResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(sdkPackageResourceConfig, zipPath),
				Check: resource.ComposeAggregateTestCheckFunc(
					testFivetranConnectorSdkPackageResourceCreate(t, "fivetran_connector_sdk_package.test_pkg"),
				),
			},
			{
				Config:       fmt.Sprintf(sdkPackageResourceConfig, zipPath),
				ResourceName: "fivetran_connector_sdk_package.test_pkg",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := GetResource(t, s, "fivetran_connector_sdk_package.test_pkg")
					return rs.Primary.ID, nil
				},
				ImportStateCheck: ComposeImportStateCheck(
					CheckImportResourceAttrSet("fivetran_connector_sdk_package", "id"),
					CheckImportResourceAttrSet("fivetran_connector_sdk_package", "file_sha256_hash"),
					CheckImportResourceAttrSet("fivetran_connector_sdk_package", "created_at"),
					CheckImportResourceAttrSet("fivetran_connector_sdk_package", "updated_at"),
				),
			},
		},
	})
}

func TestResourceConnectorSdkPackageFileNotFoundE2E(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.zip")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(sdkPackageResourceConfig, missing),
				ExpectError: regexp.MustCompile("(?s)(no such file|cannot find|file_path)"),
			},
		},
	})
}

// testFivetranConnectorSdkPackageResourceCreate fetches the package from the
// API using the id stored in state and asserts the API-reported hash matches
// the state hash (upload-corruption guard round-trip).
func testFivetranConnectorSdkPackageResourceCreate(t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := GetResource(t, s, resourceName)

		response, err := client.NewConnectorSdkPackageDetails().PackageID(rs.Primary.ID).Do(context.Background())
		if err != nil {
			return err
		}

		stateHash := rs.Primary.Attributes["file_sha256_hash"]
		if response.Data.FileSha256Hash != stateHash {
			return fmt.Errorf(
				"API hash %q does not match state hash %q for package %s",
				response.Data.FileSha256Hash, stateHash, rs.Primary.ID,
			)
		}
		return nil
	}
}

func testFivetranConnectorSdkPackageResourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fivetran_connector_sdk_package" {
			continue
		}

		response, err := client.NewConnectorSdkPackageDetails().PackageID(rs.Primary.ID).Do(context.Background())
		if err == nil {
			return errors.New("Connector SDK package " + rs.Primary.ID + " still exists after destroy")
		}
		if err.Error() != "status code: 404; expected: 200" {
			return err
		}
		if !strings.HasPrefix(response.Code, "NotFound") {
			return errors.New("Unexpected response code for deleted package " + rs.Primary.ID + ": " + response.Code)
		}
	}

	return nil
}

// writeTestSdkPackageZip creates a minimal .zip in t.TempDir() populated with
// the given file contents, returning the absolute path. The zip is cleaned up
// automatically when the test ends.
func writeTestSdkPackageZip(t *testing.T, files map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sdk_package.zip")
	writeZipAtPath(t, path, files)
	return path
}

func writeZipAtPath(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %s", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, body := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip create entry %s: %s", name, err)
		}
		if _, err := entry.Write(body); err != nil {
			t.Fatalf("zip write entry %s: %s", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %s", err)
	}
}
