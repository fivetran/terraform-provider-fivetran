package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			got := fillImportedImmutableFields(tt.plan, tt.state, slot, path.Root("config"), tt.isConfig, tt.stateName, &diags)

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
	fillImportedImmutableFields(map[string]interface{}{"table": "t"}, state, slot, path.Root("config"), true, types.StringNull(), &diags)

	if _, ok := state["table"]; ok {
		t.Fatalf("state = %v, want the input map unchanged", state)
	}
}
