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

func TestSchemaEqualRowFilter(t *testing.T) {
	base := `{"salesforce": {"tables": {"Account": {"enabled": true, "row_filter": {"name": "f", "description": "d",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}}}}}`
	cases := []struct {
		name     string
		base     string
		other    string
		expected bool
	}{
		{"same value", base, `{"salesforce": {"tables": {"Account": {"enabled": true, "row_filter": {"description": "d", "name": "f",
			"column_clauses": [{"values": ["012A"], "operator": "IN", "column": "RecordTypeId"}]}}}}}`, true},
		{"changed clause", base, `{"salesforce": {"tables": {"Account": {"enabled": true, "row_filter": {"name": "f", "description": "d",
			"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012B"]}]}}}}}`, false},
		{"missing filter", base, `{"salesforce": {"tables": {"Account": {"enabled": true}}}}`, false},
		{"null and omitted filter", `{"salesforce": {"tables": {"Account": {"enabled": true, "row_filter": null}}}}`,
			`{"salesforce": {"tables": {"Account": {"enabled": true}}}}`, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := schemaEqual(c.base, c.other)
			if err != nil {
				t.Fatal(err)
			}
			if actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}
