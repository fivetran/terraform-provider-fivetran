package model

import (
	"os"
	"reflect"
	"testing"

	fivetranCommon "github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMain(m *testing.M) {
	fivetranCommon.LoadConfigFieldsMap()
	os.Exit(m.Run())
}

func TestConnectionResourceModelGetDestinatonSchemaForConfigIncludesTableGroupName(t *testing.T) {
	t.Parallel()

	destinationSchema, diags := types.ObjectValue(destinationSchemaAttrTypes(), map[string]attr.Value{
		"name":             types.StringValue("sftp_test_schema"),
		"table":            types.StringNull(),
		"prefix":           types.StringNull(),
		"table_group_name": types.StringValue("sftp_test_table_group"),
	})
	if diags.HasError() {
		t.Fatalf("destination schema diagnostics: %v", diags)
	}

	model := ConnectionResourceModel{
		Service:           types.StringValue("sftp"),
		DestinationSchema: destinationSchema,
	}

	got, err := model.GetDestinatonSchemaForConfig()
	if err != nil {
		t.Fatalf("GetDestinatonSchemaForConfig returned error: %v", err)
	}

	want := map[string]interface{}{
		"schema":           "sftp_test_schema",
		"table_group_name": "sftp_test_table_group",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected config: got %#v, want %#v", got, want)
	}
}

func TestConnectionResourceModelGetDestinatonSchemaForConfigIncludesTable(t *testing.T) {
	t.Parallel()

	destinationSchema, diags := types.ObjectValue(destinationSchemaAttrTypes(), map[string]attr.Value{
		"name":             types.StringValue("schema"),
		"table":            types.StringValue("table"),
		"prefix":           types.StringNull(),
		"table_group_name": types.StringNull(),
	})
	if diags.HasError() {
		t.Fatalf("destination schema diagnostics: %v", diags)
	}

	model := ConnectionResourceModel{
		Service:           types.StringValue("google_sheets"),
		DestinationSchema: destinationSchema,
	}

	got, err := model.GetDestinatonSchemaForConfig()
	if err != nil {
		t.Fatalf("GetDestinatonSchemaForConfig returned error: %v", err)
	}

	want := map[string]interface{}{
		"schema": "schema",
		"table":  "table",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected config: got %#v, want %#v", got, want)
	}
}

func destinationSchemaAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":             types.StringType,
		"table":            types.StringType,
		"prefix":           types.StringType,
		"table_group_name": types.StringType,
	}
}
