package schema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestColumnToStateObjectPrimaryKey(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }
	cases := []struct {
		name     string
		upstream *bool
		local    *bool
		expected interface{}
	}{
		{name: "upstream value wins", upstream: boolPtr(false), local: boolPtr(true), expected: "false"},
		{name: "configured value kept when upstream has none", upstream: nil, local: boolPtr(true), expected: "true"},
		{name: "absent when neither has a value", upstream: nil, local: nil, expected: nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			upstream := _column{_element: _element{name: "id", enabled: true, isPrimaryKey: c.upstream}}
			local := &_column{_element: _element{name: "id", enabled: true, isPrimaryKey: c.local}}
			var diags diag.Diagnostics
			result, include := upstream.toStateObject(ALLOW_ALL, local, &diags, "public", "users", false)
			if !include {
				t.Fatalf("expected configured column to be included")
			}
			if actual, ok := result[IS_PRIMARY_KEY]; c.expected == nil && ok || c.expected != nil && actual != c.expected {
				t.Errorf("expected is_primary_key %v, got %v (present: %v)", c.expected, actual, ok)
			}
		})
	}
}
