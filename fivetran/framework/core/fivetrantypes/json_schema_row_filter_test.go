package fivetrantypes

import (
	"fmt"
	"strings"
	"testing"
)

func TestRowFilterSemanticEquality(t *testing.T) {
	filter := `{"name":"filter","description":"selection","operator":"OR","column_clauses":[{"column":"Id","column_type":"STRING","operator":"IN","values":["a"]},{"column":"Other","column_type":"STRING","operator":"IN","values":["b"]}]}`
	schema := func(field string) string {
		return fmt.Sprintf(`{"salesforce":{"enabled":true,"tables":{"Account":{"enabled":true%s}}}}`, field)
	}
	original := schema(`,"row_filter":` + filter)
	for _, tc := range []struct {
		name, other string
		equal       bool
	}{
		{"unchanged", original, true},
		{"missing", schema(""), false},
		{"null", schema(`,"row_filter":null`), false},
		{"name", strings.Replace(original, `"filter"`, `"new"`, 1), false},
		{"description", strings.Replace(original, "selection", "new", 1), false},
		{"column", strings.Replace(original, `"Id"`, `"NewId"`, 1), false},
		{"type", strings.Replace(original, `"STRING"`, `"LONG"`, 1), false},
		{"clause operator", strings.Replace(original, `"IN"`, `"EQUALS"`, 1), false},
		{"values", strings.Replace(original, `"a"`, `"c"`, 1), false},
		{"filter operator", strings.Replace(original, `"OR"`, `"AND"`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			equal, err := schemaEqual(original, tc.other)
			if err != nil {
				t.Fatal(err)
			}
			if equal != tc.equal {
				t.Fatalf("equal = %t, want %t", equal, tc.equal)
			}
		})
	}
	implicit := strings.Replace(original, `"operator":"OR",`, "", 1)
	explicit := strings.Replace(original, `"OR"`, `"AND"`, 1)
	if equal, err := schemaEqual(implicit, explicit); err != nil || !equal {
		t.Fatalf("default AND: equal=%t err=%v", equal, err)
	}
	if equal, err := schemaEqual(schema(""), schema(`,"row_filter":null`)); err != nil || equal {
		t.Fatalf("omission vs deletion: equal=%t err=%v", equal, err)
	}
}

func TestRowFilterInvalidOperatorIsNotEquivalent(t *testing.T) {
	for _, count := range []int{0, 1, 11} {
		clauses := strings.Repeat(`{"column":"Id","column_type":"STRING","operator":"IN","values":["a"]},`, count)
		clauses = strings.TrimSuffix(clauses, ",")
		implicit := fmt.Sprintf(`{"s":{"tables":{"t":{"row_filter":{"name":"test","description":"test","column_clauses":[%s]}}}}}`, clauses)
		explicit := strings.Replace(implicit, `"name":"test"`, `"operator":"AND","name":"test"`, 1)
		if equal, err := schemaEqual(implicit, explicit); err != nil || equal {
			t.Errorf("%d clauses: equal=%t err=%v", count, equal, err)
		}
	}
}
