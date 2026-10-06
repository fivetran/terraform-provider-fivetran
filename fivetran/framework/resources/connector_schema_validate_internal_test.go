package resources

import (
	"reflect"
	"testing"
)

func primaryKeyUpstream() map[string]map[string]map[string]bool {
	// "name" has no upstream value
	return map[string]map[string]map[string]bool{"public": {"users": {"id": true, "email": false}}}
}

func TestChangedPrimaryKeys(t *testing.T) {
	cases := []struct {
		name       string
		configured map[string]map[string]map[string]bool
		expected   []string
	}{
		{
			name:       "unchanged values don't warn",
			configured: map[string]map[string]map[string]bool{"public": {"users": {"id": true, "email": false}}},
			expected:   []string{},
		},
		{
			name:       "changed values warn, sorted",
			configured: map[string]map[string]map[string]bool{"public": {"users": {"id": false, "email": true}}},
			expected: []string{
				"public.users.email (current: false, configured: true)",
				"public.users.id (current: true, configured: false)",
			},
		},
		{
			name: "columns unknown upstream are skipped",
			configured: map[string]map[string]map[string]bool{
				"public": {"users": {"name": true, "missing": true}, "missing_table": {"id": true}},
				"missing_schema": {"users": {"id": true}},
			},
			expected: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := changedPrimaryKeys(c.configured, primaryKeyUpstream())
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func TestCanChangePrimaryKey(t *testing.T) {
	for _, service := range []string{"s3", "gcs", "sftp", "google_sheets", "azure_blob_storage"} {
		if !canChangePrimaryKey(service) {
			t.Errorf("expected %v to support is_primary_key", service)
		}
	}
	for _, service := range []string{"postgres", "CSV", "Google Sheets", ""} {
		if canChangePrimaryKey(service) {
			t.Errorf("expected %v not to support is_primary_key", service)
		}
	}
}
