package schema

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// A table with no locally-declared columns must still warn about upstream columns that
// are misaligned with the schema_change_handling policy (drift the API already flags),
// without writing anything into state — the existing "no noise for unconfigured columns"
// behavior must be preserved.
func TestTableToStateObjectWarnsAboutPolicyMisalignedColumnsWithNoLocalColumns(t *testing.T) {
	upstream := _table{
		_element: _element{name: "orders", enabled: true},
		columns: map[string]*_column{
			"internal_id": {
				_element: _element{name: "internal_id", enabled: true},
			},
		},
	}
	local := &_table{
		_element: _element{name: "orders", enabled: true},
		columns:  map[string]*_column{},
	}

	var diags diag.Diagnostics
	result, _ := upstream.toStateObject(BLOCK_ALL, local, &diags, "public", false)

	columns, ok := result[COLUMN].([]interface{})
	if !ok {
		t.Fatalf("expected result[%q] to be a []interface{}, got %T", COLUMN, result[COLUMN])
	}
	if len(columns) != 0 {
		t.Fatalf("expected no columns written to state for a table with no locally-declared columns, got %v", columns)
	}

	if diags.WarningsCount() == 0 {
		t.Fatalf("expected a warning about the policy-misaligned column, got none")
	}
	found := false
	for _, d := range diags.Warnings() {
		if strings.Contains(d.Detail(), "internal_id") && strings.Contains(d.Detail(), "orders") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a warning mentioning column `internal_id` in table `orders`, got: %+v", diags)
	}
}

// Same scenario, but the upstream column's enabled state DOES match the policy default —
// no drift, so no warning should fire.
func TestTableToStateObjectNoWarningWhenColumnMatchesPolicy(t *testing.T) {
	upstream := _table{
		_element: _element{name: "orders", enabled: true},
		columns: map[string]*_column{
			"internal_id": {
				_element: _element{name: "internal_id", enabled: false},
			},
		},
	}
	local := &_table{
		_element: _element{name: "orders", enabled: true},
		columns:  map[string]*_column{},
	}

	var diags diag.Diagnostics
	result, _ := upstream.toStateObject(BLOCK_ALL, local, &diags, "public", false)

	columns, ok := result[COLUMN].([]interface{})
	if !ok {
		t.Fatalf("expected result[%q] to be a []interface{}, got %T", COLUMN, result[COLUMN])
	}
	if len(columns) != 0 {
		t.Fatalf("expected no columns written to state, got %v", columns)
	}

	if diags.WarningsCount() != 0 {
		t.Fatalf("expected no warning when the column already matches the policy default, got: %+v", diags)
	}
}

// A table not declared locally at all (local == nil, e.g. the whole table is absent from
// HCL) goes through the same `else` branch as a declared-but-columnless table, so it must
// warn about policy-misaligned upstream columns too, with the same no-state-noise guarantee.
func TestTableToStateObjectWarnsAboutPolicyMisalignedColumnsWithNoLocalTable(t *testing.T) {
	upstream := _table{
		_element: _element{name: "orders", enabled: true},
		columns: map[string]*_column{
			"internal_id": {
				_element: _element{name: "internal_id", enabled: true},
			},
		},
	}

	var diags diag.Diagnostics
	result, _ := upstream.toStateObject(BLOCK_ALL, nil, &diags, "public", false)

	if _, ok := result[COLUMN]; ok {
		t.Fatalf("expected no %q key written to state for an undeclared table, got %v", COLUMN, result[COLUMN])
	}

	if diags.WarningsCount() == 0 {
		t.Fatalf("expected a warning about the policy-misaligned column, got none")
	}
}
