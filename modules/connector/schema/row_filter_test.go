package schema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func decodeTestRowFilter(t *testing.T, value string) interface{} {
	t.Helper()
	var result interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

const testRowFilter = `{
	"name": "account_record_type_filter",
	"description": "Filter Account rows by RecordTypeId",
	"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]
}`

func TestRowFilterEqual(t *testing.T) {
	for _, tc := range []struct {
		name     string
		upstream string
		local    string
		equal    bool
	}{
		{"same", testRowFilter, testRowFilter, true},
		{"both null", `null`, `null`, true},
		{"upstream missing", `null`, testRowFilter, false},
		{"delete existing", testRowFilter, `null`, false},
		{
			"computed column_type and default operator upstream",
			`{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId", "operator": "AND",
				"column_clauses": [{"column": "RecordTypeId", "column_type": "STRING", "operator": "IN", "values": ["012A"]}]}`,
			testRowFilter,
			true,
		},
		{
			"default operator configured, omitted upstream",
			testRowFilter,
			`{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId", "operator": "AND",
				"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`,
			true,
		},
		{
			"non-default operator upstream only",
			`{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId", "operator": "OR",
				"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`,
			testRowFilter,
			false,
		},
		{
			"different values",
			testRowFilter,
			`{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
				"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012B"]}]}`,
			false,
		},
		{
			"additional clause",
			testRowFilter,
			`{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
				"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]},
					{"column": "Name", "operator": "IS_NOT_NULL"}]}`,
			false,
		},
		{
			"numeric value configured",
			`{"name": "f", "description": "d", "column_clauses": [{"column": "Amount", "operator": "GREATER", "values": ["10"]}]}`,
			`{"name": "f", "description": "d", "column_clauses": [{"column": "Amount", "operator": "GREATER", "values": [10]}]}`,
			true,
		},
		{
			"empty values configured, omitted upstream",
			`{"name": "f", "description": "d", "column_clauses": [{"column": "Name", "operator": "IS_NULL"}]}`,
			`{"name": "f", "description": "d", "column_clauses": [{"column": "Name", "operator": "IS_NULL", "values": []}]}`,
			true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := rowFilterEqual(decodeTestRowFilter(t, tc.upstream), decodeTestRowFilter(t, tc.local))
			if got != tc.equal {
				t.Fatalf("rowFilterEqual() = %v, want %v", got, tc.equal)
			}
		})
	}
}

func TestValidateRowFilter(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		error string
	}{
		{"valid", testRowFilter, ""},
		{"null", `null`, ""},
		{"not an object", `"filter"`, "must be an object or null"},
		{"missing name", `{"description": "d", "column_clauses": [{"column": "c", "operator": "IS_NULL"}]}`, "`row_filter.name` is required"},
		{"missing description", `{"name": "n", "column_clauses": [{"column": "c", "operator": "IS_NULL"}]}`, "`row_filter.description` is required"},
		{"invalid operator", `{"name": "n", "description": "d", "operator": "XOR", "column_clauses": [{"column": "c", "operator": "IS_NULL"}]}`, "must be `AND` or `OR`"},
		{"no clauses", `{"name": "n", "description": "d", "column_clauses": []}`, "at least one clause"},
		{"missing clause column", `{"name": "n", "description": "d", "column_clauses": [{"operator": "IS_NULL"}]}`, "column_clauses[0].column` is required"},
		{"missing clause operator", `{"name": "n", "description": "d", "column_clauses": [{"column": "c"}]}`, "column_clauses[0].operator` is required"},
		{"values not a list", `{"name": "n", "description": "d", "column_clauses": [{"column": "c", "operator": "EQUALS", "values": "v"}]}`, "values` must be a list"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateRowFilter(decodeTestRowFilter(t, tc.value))
			if tc.error == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.error) {
				t.Fatalf("expected error containing %q, got %v", tc.error, err)
			}
		})
	}
}

func rowFilterTestTable(t *testing.T, rowFilter string) *_table {
	t.Helper()
	enabled := true
	allowed := true
	response := &connections.ConnectionSchemaConfigTableResponse{Enabled: &enabled}
	response.EnabledPatchSettings.Allowed = &allowed
	if rowFilter != "" {
		response.RowFilter = json.RawMessage(rowFilter)
	}
	table := &_table{}
	table.readFromResponse("Account", response)
	return table
}

func localRowFilterTestTable(t *testing.T, source map[string]interface{}) *_table {
	t.Helper()
	source[NAME] = "Account"
	source[ENABLED] = true
	table := &_table{}
	table.readFromResourceData(source, ALLOW_ALL)
	return table
}

func rowFilterRequest(t *testing.T, table *_table) map[string]interface{} {
	t.Helper()
	encoded, err := json.Marshal(table.prepareRequest().Request())
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestTableRowFilterRequest(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		upstream := rowFilterTestTable(t, "")
		if err := upstream.override(localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: decodeTestRowFilter(t, testRowFilter)}), ALLOW_ALL); err != nil {
			t.Fatal(err)
		}
		if !upstream.updated {
			t.Fatal("expected table to be updated")
		}
		request := rowFilterRequest(t, upstream)
		if filter, ok := request[ROW_FILTER].(map[string]interface{}); !ok || filter["name"] != "account_record_type_filter" {
			t.Fatalf("expected row_filter in request, got %v", request)
		}
	})
	t.Run("unchanged", func(t *testing.T) {
		upstream := rowFilterTestTable(t, testRowFilter)
		if err := upstream.override(localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: decodeTestRowFilter(t, testRowFilter)}), ALLOW_ALL); err != nil {
			t.Fatal(err)
		}
		if upstream.updated {
			t.Fatal("expected no update for an unchanged row filter")
		}
	})
	t.Run("omitted preserves existing filter", func(t *testing.T) {
		upstream := rowFilterTestTable(t, testRowFilter)
		if err := upstream.override(localRowFilterTestTable(t, map[string]interface{}{}), ALLOW_ALL); err != nil {
			t.Fatal(err)
		}
		if upstream.updated {
			t.Fatal("expected no update when row_filter isn't configured")
		}
		if _, ok := rowFilterRequest(t, upstream)[ROW_FILTER]; ok {
			t.Fatal("expected no row_filter in request")
		}
	})
	t.Run("explicit null deletes filter", func(t *testing.T) {
		upstream := rowFilterTestTable(t, testRowFilter)
		if err := upstream.override(localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: nil}), ALLOW_ALL); err != nil {
			t.Fatal(err)
		}
		request := rowFilterRequest(t, upstream)
		if value, ok := request[ROW_FILTER]; !ok || value != nil {
			t.Fatalf("expected row_filter: null in request, got %v", request)
		}
	})
	t.Run("removed filter is deleted", func(t *testing.T) {
		prior := SchemaConfig{schemas: map[string]*_schema{"salesforce": {tables: map[string]*_table{
			"Account": localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: decodeTestRowFilter(t, testRowFilter)}),
		}}}}
		local := SchemaConfig{schemas: map[string]*_schema{"salesforce": {tables: map[string]*_table{
			"Account": localRowFilterTestTable(t, map[string]interface{}{}),
		}}}}
		local.MarkRemovedRowFilters(prior)

		upstream := rowFilterTestTable(t, testRowFilter)
		if err := upstream.override(local.schemas["salesforce"].tables["Account"], ALLOW_ALL); err != nil {
			t.Fatal(err)
		}
		request := rowFilterRequest(t, upstream)
		if value, ok := request[ROW_FILTER]; !ok || value != nil {
			t.Fatalf("expected row_filter: null in request, got %v", request)
		}

		// the deletion isn't saved into state, the filter isn't configured anymore
		var diags diag.Diagnostics
		state, _ := rowFilterTestTable(t, "").toStateObject(ALLOW_ALL, local.schemas["salesforce"].tables["Account"], &diags, "salesforce", false)
		if _, ok := state[ROW_FILTER]; ok {
			t.Fatalf("expected no row_filter in state, got %v", state)
		}
	})
}

func TestTableRowFilterState(t *testing.T) {
	upstreamFilter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "column_type": "STRING", "operator": "IN", "values": ["012A"]}]}`
	var diags diag.Diagnostics

	t.Run("configured filter is saved in configured shape", func(t *testing.T) {
		local := localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: decodeTestRowFilter(t, testRowFilter)})
		state, _ := rowFilterTestTable(t, upstreamFilter).toStateObject(ALLOW_ALL, local, &diags, "salesforce", false)
		if !rowFilterEqual(state[ROW_FILTER], local.rowFilter) || !strings.Contains(string(encodeRowFilter(state[ROW_FILTER])), "RecordTypeId") {
			t.Fatalf("unexpected row_filter in state: %v", state[ROW_FILTER])
		}
		if strings.Contains(string(encodeRowFilter(state[ROW_FILTER])), "column_type") {
			t.Fatalf("expected computed column_type to be dropped from state: %v", state[ROW_FILTER])
		}
	})
	t.Run("unconfigured filter isn't saved", func(t *testing.T) {
		state, _ := rowFilterTestTable(t, upstreamFilter).toStateObject(ALLOW_ALL, localRowFilterTestTable(t, map[string]interface{}{}), &diags, "salesforce", false)
		if _, ok := state[ROW_FILTER]; ok {
			t.Fatalf("expected no row_filter in state, got %v", state)
		}
	})
	t.Run("configured filter missing upstream", func(t *testing.T) {
		local := localRowFilterTestTable(t, map[string]interface{}{ROW_FILTER: decodeTestRowFilter(t, testRowFilter)})
		state, _ := rowFilterTestTable(t, "").toStateObject(ALLOW_ALL, local, &diags, "salesforce", false)
		if value, ok := state[ROW_FILTER]; !ok || value != nil {
			t.Fatalf("expected row_filter: null in state, got %v", state)
		}
	})
	t.Run("imported filter is saved as is", func(t *testing.T) {
		state, _ := rowFilterTestTable(t, upstreamFilter).toStateObject(ALLOW_ALL, nil, &diags, "salesforce", true)
		if !strings.Contains(string(encodeRowFilter(state[ROW_FILTER])), "column_type") {
			t.Fatalf("expected the whole upstream filter in state, got %v", state[ROW_FILTER])
		}
	})
}
