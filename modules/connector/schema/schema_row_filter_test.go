package schema

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestTableRowFilter(t *testing.T) {
	filter := map[string]interface{}{"name": "", "description": "", "column_clauses": []interface{}{map[string]interface{}{"column": "Id", "column_type": "STRING", "operator": "IN", "values": []interface{}{"a"}}}}
	for _, tc := range []struct {
		name             string
		remote, local    interface{}
		managed, changed bool
	}{
		{"create", nil, filter, true, true},
		{"unchanged blank metadata", filter, filter, true, false},
		{"delete", filter, nil, true, true},
		{"already deleted", nil, nil, true, false},
		{"unmanaged", filter, nil, false, false},
		{"update", filter, map[string]interface{}{"name": "new"}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(map[string]interface{}{"enabled": true, "row_filter": tc.remote})
			if err != nil {
				t.Fatal(err)
			}
			var response connections.ConnectionSchemaConfigTableResponse
			if err = json.Unmarshal(raw, &response); err != nil {
				t.Fatal(err)
			}
			remote := &_table{}
			remote.readFromResponse("Account", &response)
			source := map[string]interface{}{"name": "Account", "enabled": true}
			if tc.managed {
				source["row_filter"] = tc.local
			}
			local := &_table{}
			local.readFromResourceData(source, ALLOW_ALL)
			// Readback retains only explicitly managed filters, including explicit null.
			var diags diag.Diagnostics
			state, _ := remote.toStateObject(ALLOW_ALL, local, &diags, "salesforce", false)
			got, present := state["row_filter"]
			if present != tc.managed || (present && !reflect.DeepEqual(got, tc.remote)) {
				t.Fatalf("state row_filter = %#v, present %t", got, present)
			}
			if err = remote.override(local, ALLOW_ALL); err != nil {
				t.Fatal(err)
			}
			if remote.updated != tc.changed {
				t.Fatalf("updated=%t, want %t", remote.updated, tc.changed)
			}
			request := remote.prepareRequest().Request()
			if tc.changed {
				want, _ := json.Marshal(tc.local)
				if string(request.RowFilter) != string(want) {
					t.Fatalf("request = %s, want %s", request.RowFilter, want)
				}
			} else if request.RowFilter != nil {
				t.Fatalf("unchanged filter sent: %s", request.RowFilter)
			}
			create := local.prepareCreateRequest().Request()
			if tc.managed {
				want, _ := json.Marshal(tc.local)
				if string(create.RowFilter) != string(want) {
					t.Fatalf("create = %s, want %s", create.RowFilter, want)
				}
			} else if create.RowFilter != nil {
				t.Fatal("unmanaged create sends filter")
			}
		})
	}
}

func TestTableRowFilterDefaultOperator(t *testing.T) {
	localFilter := map[string]interface{}{"name": "test", "description": "test", "column_clauses": []interface{}{map[string]interface{}{"column": "Id"}, map[string]interface{}{"column": "Other"}}}
	remoteFilter := map[string]interface{}{"name": "test", "description": "test", "column_clauses": []interface{}{map[string]interface{}{"column": "Id"}, map[string]interface{}{"column": "Other"}}, "operator": "AND"}
	local := &_table{rowFilter: localFilter, rowFilterSet: true}
	remote := &_table{rowFilter: remoteFilter}
	if err := remote.override(local, ALLOW_ALL); err != nil {
		t.Fatal(err)
	}
	if remote.updated {
		t.Fatal("default AND must not cause an update")
	}
	var diags diag.Diagnostics
	state, _ := remote.toStateObject(ALLOW_ALL, local, &diags, "salesforce", false)
	if !reflect.DeepEqual(state["row_filter"], localFilter) {
		t.Fatalf("default AND changes state: %#v", state)
	}
	if remoteFilter["operator"] != "AND" {
		t.Fatal("comparison mutated remote filter")
	}
}
