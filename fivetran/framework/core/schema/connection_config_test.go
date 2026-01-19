package schema

import (
	"testing"
)

func TestConnectionConfigResourceSchema(t *testing.T) {
	schema := ConnectionConfigResourceSchema()

	requiredFields := []string{"connection_id"}
	for _, field := range requiredFields {
		if _, ok := schema.Attributes[field]; !ok {
			t.Errorf("Required field %s not found in schema", field)
		}
	}

	newFields := []string{"run_setup_tests", "trust_certificates", "trust_fingerprints"}
	for _, field := range newFields {
		if _, ok := schema.Attributes[field]; !ok {
			t.Errorf("New field %s not found in schema", field)
		}
	}

	allExpectedFields := []string{"id", "connection_id", "config", "auth", "run_setup_tests", "trust_certificates", "trust_fingerprints"}
	if len(schema.Attributes) != len(allExpectedFields) {
		t.Errorf("Expected %d attributes, got %d", len(allExpectedFields), len(schema.Attributes))
	}

	for _, field := range allExpectedFields {
		if _, ok := schema.Attributes[field]; !ok {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	t.Logf("✅ Schema has all %d expected fields", len(allExpectedFields))
}
