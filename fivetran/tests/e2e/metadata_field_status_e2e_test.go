package e2e_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
)

func TestConnectorMetadataFieldStatusE2E(t *testing.T) {
	response, err := client.NewMetadataDetails().Service("google_ads").Do(context.Background())
	if err != nil {
		t.Fatalf("fetch google_ads metadata: %v", err)
	}
	if response.Data.ID != "google_ads" {
		t.Fatalf("unexpected metadata service: got %q, want %q", response.Data.ID, "google_ads")
	}

	counts := map[string]int{}
	var missing []string
	var unknown []string

	collectMetadataFieldStatuses("config", response.Data.Config.Properties, counts, &missing, &unknown)
	collectMetadataFieldStatuses("auth", response.Data.Auth.Properties, counts, &missing, &unknown)

	if len(missing) > 0 {
		t.Fatalf("metadata fields missing fieldStatus: %s", strings.Join(missing, ", "))
	}
	if len(unknown) > 0 {
		t.Fatalf("metadata fields have unknown fieldStatus: %s", strings.Join(unknown, ", "))
	}
	if totalMetadataFieldStatuses(counts) == 0 {
		t.Fatal("expected at least one metadata field with fieldStatus")
	}
	if counts[framework.FieldStatusGeneralAvailability] == 0 {
		t.Fatalf("expected at least one general_availability fieldStatus, got counts: %+v", counts)
	}
}

func collectMetadataFieldStatuses(path string, properties map[string]*metadata.Property, counts map[string]int, missing, unknown *[]string) {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	slices.Sort(names)

	for _, name := range names {
		prop := properties[name]
		fieldPath := fmt.Sprintf("%s.%s", path, name)
		if prop == nil {
			*missing = append(*missing, fieldPath)
			continue
		}

		status := framework.MetadataFieldStatus(prop)
		if status == "" {
			*missing = append(*missing, fieldPath)
		} else if !isKnownMetadataFieldStatus(status) {
			*unknown = append(*unknown, fmt.Sprintf("%s=%s", fieldPath, status))
		} else {
			counts[status]++
		}

		if len(prop.Properties) > 0 {
			collectMetadataFieldStatuses(fieldPath, prop.Properties, counts, missing, unknown)
		}
		if prop.Items != nil && len(prop.Items.Properties) > 0 {
			collectMetadataFieldStatuses(fieldPath+"[]", prop.Items.Properties, counts, missing, unknown)
		}
	}
}

func isKnownMetadataFieldStatus(status string) bool {
	switch status {
	case framework.FieldStatusDevelopment,
		framework.FieldStatusPrivatePreview,
		framework.FieldStatusGeneralAvailability,
		framework.FieldStatusSunset:
		return true
	default:
		return false
	}
}

func totalMetadataFieldStatuses(counts map[string]int) int {
	var total int
	for _, count := range counts {
		total += count
	}
	return total
}
