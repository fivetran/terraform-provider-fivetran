package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	if diags := forgetImportedConfigFields(ctx, private); diags.HasError() {
		t.Fatal(diags)
	}
	imported, _ = loadImportedConfigFields(ctx, private)
	if imported != nil {
		t.Fatalf("imported after forget = %v, want nil", imported)
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
