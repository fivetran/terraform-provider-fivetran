package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConnectionV2ImportedMarkerRoundTrip(t *testing.T) {
	ctx := context.Background()
	private := fakePrivateState{}

	if imported, _ := isConnectionV2Imported(ctx, private); imported {
		t.Fatal("imported = true before import")
	}
	if diags := setConnectionV2Imported(ctx, private, true); diags.HasError() {
		t.Fatal(diags)
	}
	if imported, _ := isConnectionV2Imported(ctx, private); !imported {
		t.Fatal("imported = false after import")
	}
	if diags := setConnectionV2Imported(ctx, private, false); diags.HasError() {
		t.Fatal(diags)
	}
	if imported, _ := isConnectionV2Imported(ctx, private); imported {
		t.Fatal("imported = true after update")
	}
	if imported, diags := isConnectionV2Imported(ctx, nil); imported || diags.HasError() {
		t.Fatalf("nil private state: imported = %v, diags = %v", imported, diags)
	}
}

func TestFillImportedImmutableFields(t *testing.T) {
	t.Parallel()

	slot := &metadata.Property{Properties: map[string]*metadata.Property{
		"schema_prefix": {Type: "string", Immutable: true},
		"schema":        {Type: "string", Immutable: true},
		"table":         {Type: "string", Immutable: true},
		"host":          {Type: "string"},
	}}

	tests := []struct {
		name         string
		plan         map[string]interface{}
		state        map[string]interface{}
		isConfig     bool
		stateName    types.String
		want         map[string]interface{}
		wantWarnings int
	}{
		{
			name:      "schema_prefix matches name",
			plan:      map[string]interface{}{"schema_prefix": "pg", "host": "h"},
			state:     map[string]interface{}{"host": "h"},
			isConfig:  true,
			stateName: types.StringValue("pg"),
			want:      map[string]interface{}{"schema_prefix": "pg", "host": "h"},
		},
		{
			name:      "schema matches name",
			plan:      map[string]interface{}{"schema": "sheet"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("sheet"),
			want:      map[string]interface{}{"schema": "sheet"},
		},
		{
			// A different prefix is a real change: it stays missing and forces a replacement.
			name:      "schema_prefix differs from name",
			plan:      map[string]interface{}{"schema_prefix": "other"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("pg"),
			want:      map[string]interface{}{},
		},
		{
			name:      "schema_prefix with null name",
			plan:      map[string]interface{}{"schema_prefix": "pg"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringNull(),
			want:      map[string]interface{}{},
		},
		{
			name:      "single-table schema matches the schema part of name",
			plan:      map[string]interface{}{"schema": "sales", "table": "orders"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("sales.orders"),
			want:      map[string]interface{}{"schema": "sales", "table": "orders"},
		},
		{
			name:      "single-table table differs from name",
			plan:      map[string]interface{}{"schema": "sales", "table": "other"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("sales.orders"),
			want:      map[string]interface{}{"schema": "sales"},
		},
		{
			name:      "schema_prefix compares the whole name",
			plan:      map[string]interface{}{"schema_prefix": "sales"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("sales.orders"),
			want:      map[string]interface{}{},
		},
		{
			// A name without a table part can't hold the table, so it can't be checked.
			name:         "other immutable field missing",
			plan:         map[string]interface{}{"table": "t"},
			state:        map[string]interface{}{},
			isConfig:     true,
			stateName:    types.StringValue("pg"),
			want:         map[string]interface{}{"table": "t"},
			wantWarnings: 1,
		},
		{
			// auth has no schema name fields, so schema is an ordinary missing field there.
			name:         "auth immutable field missing",
			plan:         map[string]interface{}{"schema": "pg"},
			state:        nil,
			isConfig:     false,
			stateName:    types.StringValue("pg"),
			want:         map[string]interface{}{"schema": "pg"},
			wantWarnings: 1,
		},
		{
			name:      "immutable field in state is left for the usual comparison",
			plan:      map[string]interface{}{"table": "new"},
			state:     map[string]interface{}{"table": "old"},
			isConfig:  true,
			stateName: types.StringValue("pg"),
			want:      map[string]interface{}{"table": "old"},
		},
		{
			name:      "mutable field missing",
			plan:      map[string]interface{}{"host": "h"},
			state:     map[string]interface{}{},
			isConfig:  true,
			stateName: types.StringValue("pg"),
			want:      map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			got := fillImportedImmutableFields(tt.plan, tt.state, slot, path.Root("config"), tt.isConfig, tt.stateName, "", &diags)

			if len(got) != len(tt.want) || (len(got) > 0 && !reflect.DeepEqual(got, tt.want)) {
				t.Fatalf("filled state = %v, want %v", got, tt.want)
			}
			assertErrorCount(t, diags, 0)
			assertWarningCount(t, diags, tt.wantWarnings)
		})
	}
}

func TestFillImportedImmutableFieldsDoesNotModifyState(t *testing.T) {
	t.Parallel()

	slot := &metadata.Property{Properties: map[string]*metadata.Property{
		"table": {Type: "string", Immutable: true},
	}}
	state := map[string]interface{}{"host": "h"}
	var diags diag.Diagnostics
	fillImportedImmutableFields(map[string]interface{}{"table": "t"}, state, slot, path.Root("config"), true, types.StringNull(), "", &diags)

	if _, ok := state["table"]; ok {
		t.Fatalf("state = %v, want the input map unchanged", state)
	}
}

func TestFillImportedImmutableFieldsNested(t *testing.T) {
	t.Parallel()

	slot := &metadata.Property{Properties: map[string]*metadata.Property{
		"settings": {Type: "object", Immutable: true, Properties: map[string]*metadata.Property{
			"region": {Type: "string"},
			"tier":   {Type: "string"},
		}},
		"client_access": {Type: "object", Properties: map[string]*metadata.Property{
			"client_id":     {Type: "string", Immutable: true},
			"client_secret": {Type: "string"},
		}},
		"accounts": {Type: "array", Items: &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{
			"id":   {Type: "string", Immutable: true},
			"name": {Type: "string"},
		}}},
	}}

	tests := []struct {
		name         string
		plan         map[string]interface{}
		state        map[string]interface{}
		want         map[string]interface{}
		wantWarnings int
	}{
		{
			name:         "missing parent object",
			plan:         map[string]interface{}{"client_access": map[string]interface{}{"client_id": "c", "client_secret": "s"}},
			state:        nil,
			want:         map[string]interface{}{"client_access": map[string]interface{}{"client_id": "c"}},
			wantWarnings: 1,
		},
		{
			name:         "parent object without the immutable field",
			plan:         map[string]interface{}{"client_access": map[string]interface{}{"client_id": "c", "client_secret": "s"}},
			state:        map[string]interface{}{"client_access": map[string]interface{}{"client_secret": "old"}},
			want:         map[string]interface{}{"client_access": map[string]interface{}{"client_id": "c", "client_secret": "old"}},
			wantWarnings: 1,
		},
		{
			name:  "nested immutable field in state is left for the usual comparison",
			plan:  map[string]interface{}{"client_access": map[string]interface{}{"client_id": "new"}},
			state: map[string]interface{}{"client_access": map[string]interface{}{"client_id": "old"}},
			want:  map[string]interface{}{"client_access": map[string]interface{}{"client_id": "old"}},
		},
		{
			name:         "immutable object returned without a child",
			plan:         map[string]interface{}{"settings": map[string]interface{}{"region": "eu", "tier": "gold"}},
			state:        map[string]interface{}{"settings": map[string]interface{}{"region": "eu"}},
			want:         map[string]interface{}{"settings": map[string]interface{}{"region": "eu", "tier": "gold"}},
			wantWarnings: 1,
		},
		{
			// A child that differs is a real change and is left for the comparison to replace.
			name:  "immutable object child differs",
			plan:  map[string]interface{}{"settings": map[string]interface{}{"region": "us"}},
			state: map[string]interface{}{"settings": map[string]interface{}{"region": "eu"}},
			want:  map[string]interface{}{"settings": map[string]interface{}{"region": "eu"}},
		},
		{
			name: "list items",
			plan: map[string]interface{}{"accounts": []interface{}{
				map[string]interface{}{"id": "a", "name": "A"},
				map[string]interface{}{"id": "b", "name": "B"},
			}},
			state: map[string]interface{}{"accounts": []interface{}{
				map[string]interface{}{"id": "a", "name": "A"},
			}},
			want: map[string]interface{}{"accounts": []interface{}{
				map[string]interface{}{"id": "a", "name": "A"},
				map[string]interface{}{"id": "b"},
			}},
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			got := fillImportedImmutableFields(tt.plan, tt.state, slot, path.Root("auth"), false, types.StringNull(), "", &diags)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("filled state = %v, want %v", got, tt.want)
			}
			assertErrorCount(t, diags, 0)
			assertWarningCount(t, diags, tt.wantWarnings)
		})
	}
}

// After the fill, the replacement check finds no change for an immutable object the API returned
// without a child, and still replaces when a child really differs.
func TestFillImportedImmutableFieldsAvoidsReplacement(t *testing.T) {
	t.Parallel()

	slot := &metadata.Property{Properties: map[string]*metadata.Property{
		"settings": {Type: "object", Immutable: true, Properties: map[string]*metadata.Property{
			"region": {Type: "string"},
			"tier":   {Type: "string"},
		}},
	}}
	state := map[string]interface{}{"settings": map[string]interface{}{"region": "eu"}}

	for _, tt := range []struct {
		name        string
		plan        map[string]interface{}
		wantReplace bool
	}{
		{"missing child", map[string]interface{}{"settings": map[string]interface{}{"region": "eu", "tier": "gold"}}, false},
		{"changed child", map[string]interface{}{"settings": map[string]interface{}{"region": "us", "tier": "gold"}}, true},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			filled := fillImportedImmutableFields(tt.plan, state, slot, path.Root("config"), true, types.StringValue("pg"), "", &diags)
			resp := &resource.ModifyPlanResponse{}
			got := appendImmutableRequiresReplace(tt.plan, filled, slot, path.Root("config"), resp, "", "", connectionWording)
			if got != tt.wantReplace {
				t.Fatalf("replacement = %v, want %v", got, tt.wantReplace)
			}
		})
	}
}
