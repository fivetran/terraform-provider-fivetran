package fivetrantypes

import "testing"

func TestSchemaEqualPrimaryKey(t *testing.T) {
	base := `{"public": {"tables": {"users": {"columns": {"id": {"enabled": true, "is_primary_key": true}}}}}}`
	cases := []struct {
		name     string
		other    string
		expected bool
	}{
		{"same value", `{"public": {"tables": {"users": {"columns": {"id": {"enabled": "true", "is_primary_key": "true"}}}}}}`, true},
		{"changed value", `{"public": {"tables": {"users": {"columns": {"id": {"enabled": true, "is_primary_key": false}}}}}}`, false},
		{"removed value", `{"public": {"tables": {"users": {"columns": {"id": {"enabled": true}}}}}}`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := schemaEqual(base, c.other)
			if err != nil {
				t.Fatal(err)
			}
			if actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}
