package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type fakePrivateState map[string][]byte

func (f fakePrivateState) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return f[key], nil
}

func (f fakePrivateState) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if len(value) == 0 {
		delete(f, key)
		return nil
	}
	f[key] = value
	return nil
}

func TestImportedConfigFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()
	private := fakePrivateState{}

	config, diags := core.MapToDynamic(ctx, map[string]interface{}{"host": "h", "auth_type": "PASSWORD"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if diags := storeImportedConfigFields(ctx, private, config); diags.HasError() {
		t.Fatal(diags)
	}
	imported, diags := loadImportedConfigFields(ctx, private)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if want := map[string]bool{"host": true, "auth_type": true}; !reflect.DeepEqual(imported, want) {
		t.Fatalf("imported = %v, want %v", imported, want)
	}

	if diags := setImportedConfigFields(ctx, private, nil); diags.HasError() {
		t.Fatal(diags)
	}
	imported, _ = loadImportedConfigFields(ctx, private)
	if imported != nil {
		t.Fatalf("imported after clearing = %v, want nil", imported)
	}
}

func TestOmitUndeclaredImportedConfigFields(t *testing.T) {
	imported := map[string]bool{"host": true, "auth_type": true, "s3_bucket_auth_type": true, "role": true}
	state := map[string]interface{}{"host": "h", "auth_type": "PASSWORD", "s3_bucket_auth_type": "IAM_ROLE", "role": "R", "schema": "s"}
	plan := map[string]interface{}{"host": "new", "role": nil}

	fields := undeclaredImportedConfigFields(imported, plan, state)
	// schema was not imported (declared by the user since), so it is not listed.
	if want := []string{"auth_type", "s3_bucket_auth_type"}; !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %v, want %v", fields, want)
	}

	patch := map[string]interface{}{"host": "new", "auth_type": nil, "s3_bucket_auth_type": nil, "role": nil, "schema": nil}
	omitUndeclaredImportedConfigFields(patch, fields)
	// Explicit null (role) and removal of a non-imported field (schema) still clear.
	want := map[string]interface{}{"host": "new", "role": nil, "schema": nil}
	if !reflect.DeepEqual(patch, want) {
		t.Fatalf("patch = %v, want %v", patch, want)
	}
}

func TestConfigWithFieldsEqualsState(t *testing.T) {
	state := map[string]interface{}{"host": "h", "port": float64(1), "always_encrypted": true}
	cases := []struct {
		name   string
		plan   map[string]interface{}
		fields []string
		want   bool
	}{
		{"only undeclared fields differ", map[string]interface{}{"host": "h", "port": float64(1)}, []string{"always_encrypted"}, true},
		{"declared field changes", map[string]interface{}{"host": "h", "port": float64(2)}, []string{"always_encrypted"}, false},
		{"declared field removed", map[string]interface{}{"host": "h"}, []string{"always_encrypted"}, false},
		{"field added", map[string]interface{}{"host": "h", "port": float64(1), "user": "u"}, []string{"always_encrypted"}, false},
		{"unknown value", map[string]interface{}{"host": core.DynamicUnknownValue{}, "port": float64(1)}, []string{"always_encrypted"}, false},
		{"nested unknown value", map[string]interface{}{"host": "h", "port": []interface{}{core.DynamicUnknownValue{}}}, []string{"always_encrypted"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := configWithFieldsEqualsState(tc.plan, state, tc.fields); got != tc.want {
				t.Fatalf("configWithFieldsEqualsState() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPlanUnchangedExceptComputedUnknowns(t *testing.T) {
	statusType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"state": tftypes.String}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":   tftypes.String,
		"paused": tftypes.Bool,
		"status": statusType,
	}}
	obj := func(name, paused, status tftypes.Value) tftypes.Value {
		return tftypes.NewValue(objType, map[string]tftypes.Value{"name": name, "paused": paused, "status": status})
	}
	str := func(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
	boolean := func(v bool) tftypes.Value { return tftypes.NewValue(tftypes.Bool, v) }
	status := func(v tftypes.Value) tftypes.Value {
		return tftypes.NewValue(statusType, map[string]tftypes.Value{"state": v})
	}
	nullStr := tftypes.NewValue(tftypes.String, nil)
	unknownStr := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	unknownBool := tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	nullStatus := tftypes.NewValue(statusType, nil)
	unknownStatus := tftypes.NewValue(statusType, tftypes.UnknownValue)

	state := obj(str("postgres"), boolean(true), status(str("connected")))
	cases := []struct {
		name   string
		plan   tftypes.Value
		config tftypes.Value
		want   bool
	}{
		{"equal", state, obj(nullStr, boolean(true), nullStatus), true},
		{"computed attributes unknown", obj(unknownStr, boolean(true), unknownStatus), obj(nullStr, boolean(true), nullStatus), true},
		{"nested computed attribute unknown", obj(str("postgres"), boolean(true), status(unknownStr)), obj(nullStr, boolean(true), status(nullStr)), true},
		{"configured attribute changes", obj(unknownStr, boolean(false), unknownStatus), obj(nullStr, boolean(false), nullStatus), false},
		{"configured value unknown", obj(str("postgres"), unknownBool, status(str("connected"))), obj(nullStr, unknownBool, nullStatus), false},
		{"computed value set in config is unknown", obj(unknownStr, boolean(true), status(str("connected"))), obj(unknownStr, boolean(true), nullStatus), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := planUnchangedExceptComputedUnknowns(tc.plan, tc.config, state); got != tc.want {
				t.Fatalf("planUnchangedExceptComputedUnknowns() = %v, want %v", got, tc.want)
			}
		})
	}
}
