package model

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	reconcileTestPolarisTypes = map[string]attr.Type{
		"client_id":       types.StringType,
		"polaris_catalog": types.StringType,
	}

	reconcileTestConfigTypes = map[string]attr.Type{
		"bucket":                        types.StringType,
		"password":                      types.StringType,
		"prefix":                        types.StringType,
		"servers":                       types.SetType{ElemType: types.StringType},
		"endpoint_redshift":             types.StringType,
		"polaris_catalog_configuration": types.ObjectType{AttrTypes: reconcileTestPolarisTypes},
	}

	reconcileTestFields = map[string]common.ConfigField{
		"bucket":   {},
		"password": {Sensitive: true},
		"prefix":   {},
		"servers":  {FieldValueType: common.StringList},
		// ApiField is the key of the field in the request
		"endpoint_redshift": {ApiField: "endpoint"},
		"polaris_catalog_configuration": {ItemFields: map[string]common.ConfigField{
			"client_id":       {},
			"polaris_catalog": {},
		}},
	}
)

func reconcileTestObject(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(reconcileTestConfigTypes, attrs)
	if diags.HasError() {
		t.Fatalf("cannot build object: %v", diags)
	}
	return obj
}

// reconcileTestObjectWithNulls builds the config object. Attributes that are not in attrs are null.
func reconcileTestObjectWithNulls(t *testing.T, attrs map[string]attr.Value) types.Object {
	t.Helper()
	all := map[string]attr.Value{
		"bucket":                        types.StringNull(),
		"password":                      types.StringNull(),
		"prefix":                        types.StringNull(),
		"servers":                       types.SetNull(types.StringType),
		"endpoint_redshift":             types.StringNull(),
		"polaris_catalog_configuration": types.ObjectNull(reconcileTestPolarisTypes),
	}
	for name, value := range attrs {
		all[name] = value
	}
	return reconcileTestObject(t, all)
}

func reconcileTestPolaris(t *testing.T, clientId, catalog attr.Value) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(reconcileTestPolarisTypes, map[string]attr.Value{
		"client_id":       clientId,
		"polaris_catalog": catalog,
	})
	if diags.HasError() {
		t.Fatalf("cannot build nested object: %v", diags)
	}
	return obj
}

func reconcileTestRun(t *testing.T, planned, applied types.Object, sent map[string]interface{}) (types.Object, []string) {
	t.Helper()
	result, unconfirmed, diags := reconcileConfigAfterWriteWithFields(context.Background(), planned, applied, sent, reconcileTestFields)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	return result, unconfirmed
}

func TestReconcileConfigAfterWriteKeepsPlannedValuesResolvedToNull(t *testing.T) {
	planned := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket":            types.StringValue("my-bucket"),
		"endpoint_redshift": types.StringValue("endpoint-value"),
		"polaris_catalog_configuration": reconcileTestPolaris(t,
			types.StringValue("client-id"), types.StringValue("catalog")),
	})

	// What the response-derived state looks like: unreturned fields are null
	applied := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket": types.StringValue("my-bucket"),
	})

	sent := map[string]interface{}{
		"bucket":                        "my-bucket",
		"endpoint":                      "endpoint-value",
		"polaris_catalog_configuration": map[string]interface{}{},
	}

	result, unconfirmed := reconcileTestRun(t, planned, applied, sent)

	wantUnconfirmed := []string{
		"config.endpoint_redshift",
		"config.polaris_catalog_configuration",
	}
	if !reflect.DeepEqual(unconfirmed, wantUnconfirmed) {
		t.Errorf("unconfirmed paths: got %v, want %v", unconfirmed, wantUnconfirmed)
	}
	if !result.Equal(planned) {
		t.Errorf("state must keep every planned value resolved to null:\n got %v\nwant %v", result, planned)
	}
}

// Sensitive values, empty strings and empty collections are kept by the default read logic even when
// the response omits them. They must stay as they are and must not be reported.
func TestReconcileConfigAfterWriteDoesNotChangeValuesKeptByRead(t *testing.T) {
	planned := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"password": types.StringValue("secret-value"),
		"prefix":   types.StringValue(""),
		"servers":  types.SetValueMust(types.StringType, []attr.Value{}),
	})
	applied := planned

	sent := map[string]interface{}{
		"password": "secret-value",
		"prefix":   "",
		"servers":  []interface{}{},
	}

	result, unconfirmed := reconcileTestRun(t, planned, applied, sent)

	if len(unconfirmed) != 0 {
		t.Errorf("values kept by read must not be reported, got %v", unconfirmed)
	}
	if !result.Equal(applied) {
		t.Errorf("values kept by read must not change:\n got %v\nwant %v", result, applied)
	}
}

func TestReconcileConfigAfterWriteRecursesIntoReturnedNestedObject(t *testing.T) {
	planned := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket": types.StringValue("my-bucket"),
		"polaris_catalog_configuration": reconcileTestPolaris(t,
			types.StringValue("configured-client-id"), types.StringValue("catalog")),
	})

	// The nested object is returned, but without polaris_catalog. The value from the response is kept
	// for client_id, and the planned value is kept for polaris_catalog.
	applied := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket": types.StringValue("my-bucket"),
		"polaris_catalog_configuration": reconcileTestPolaris(t,
			types.StringValue("api-client-id"), types.StringNull()),
	})

	sent := map[string]interface{}{"polaris_catalog_configuration": map[string]interface{}{}}

	result, unconfirmed := reconcileTestRun(t, planned, applied, sent)

	if want := []string{"config.polaris_catalog_configuration.polaris_catalog"}; !reflect.DeepEqual(unconfirmed, want) {
		t.Errorf("unconfirmed paths: got %v, want %v", unconfirmed, want)
	}

	want := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket": types.StringValue("my-bucket"),
		"polaris_catalog_configuration": reconcileTestPolaris(t,
			types.StringValue("api-client-id"), types.StringValue("catalog")),
	})
	if !result.Equal(want) {
		t.Errorf("nested reconcile:\n got %v\nwant %v", result, want)
	}
}

func TestReconcileConfigAfterWriteKeepsResponseValues(t *testing.T) {
	planned := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"endpoint_redshift": types.StringValue("configured-endpoint"),
	})
	applied := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"endpoint_redshift": types.StringValue("api-endpoint"),
	})

	result, unconfirmed := reconcileTestRun(t, planned, applied, map[string]interface{}{"endpoint": "configured-endpoint"})

	if len(unconfirmed) != 0 {
		t.Errorf("returned value must not be reported, got %v", unconfirmed)
	}
	if !result.Equal(applied) {
		t.Errorf("returned value must come from the response:\n got %v\nwant %v", result, applied)
	}
}

// A field that was not sent (for example, an update that changes only time_zone_offset) keeps its
// planned value, but is not reported.
func TestReconcileConfigAfterWriteReportsOnlySentFields(t *testing.T) {
	planned := reconcileTestObjectWithNulls(t, map[string]attr.Value{
		"bucket":            types.StringValue("my-bucket"),
		"endpoint_redshift": types.StringValue("endpoint-value"),
	})
	applied := reconcileTestObjectWithNulls(t, nil)

	result, unconfirmed := reconcileTestRun(t, planned, applied, map[string]interface{}{"endpoint": "endpoint-value"})

	if want := []string{"config.endpoint_redshift"}; !reflect.DeepEqual(unconfirmed, want) {
		t.Errorf("unconfirmed paths: got %v, want %v", unconfirmed, want)
	}
	if !result.Equal(planned) {
		t.Errorf("state must keep every planned value resolved to null:\n got %v\nwant %v", result, planned)
	}

	_, unconfirmed = reconcileTestRun(t, planned, applied, nil)
	if len(unconfirmed) != 0 {
		t.Errorf("nothing sent, nothing must be reported, got %v", unconfirmed)
	}
}
