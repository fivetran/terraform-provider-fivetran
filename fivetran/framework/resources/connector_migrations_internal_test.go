package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConvertSetToBlockReportsErrorsWithoutPanicking(t *testing.T) {
	attrTypes := map[string]tftypes.Type{"name": tftypes.String}
	elem := func(name string) tftypes.Value {
		return tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, name),
		})
	}
	setType := tftypes.Set{ElementType: tftypes.Object{AttributeTypes: attrTypes}}

	var diags diag.Diagnostics
	result := convertSetToBlock("destination_schema", tftypes.NewValue(setType, []tftypes.Value{elem("a"), elem("b")}), attrTypes, attrTypes, &diags)
	if !diags.HasError() {
		t.Fatal("expected an error for a set with two elements")
	}
	if !result.IsNull() {
		t.Fatalf("expected a null block, got %v", result)
	}

	diags = nil
	convertSetToBlock("destination_schema", tftypes.NewValue(tftypes.String, "not a set"), attrTypes, attrTypes, &diags)
	if !diags.HasError() {
		t.Fatal("expected an error for a value that is not a set")
	}
}

func TestUnmarshalRawStateLenientDropsUnknownAttributes(t *testing.T) {
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id": tftypes.String,
		"config": tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"host": tftypes.String,
		}}},
	}}
	raw := &tfprotov6.RawState{JSON: []byte(`{"id":"x","removed":"y","config":[{"host":"h","removed_field":"z"}]}`)}

	if _, err := raw.Unmarshal(stateType); err == nil {
		t.Fatal("expected the strict unmarshal to reject unknown attributes")
	}

	value, err := unmarshalRawStateLenient(raw, stateType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var state map[string]tftypes.Value
	if err := value.As(&state); err != nil {
		t.Fatal(err)
	}
	var configs []tftypes.Value
	if err := state["config"].As(&configs); err != nil || len(configs) != 1 {
		t.Fatalf("expected one config element, got %v (%v)", configs, err)
	}
	var config map[string]tftypes.Value
	if err := configs[0].As(&config); err != nil {
		t.Fatal(err)
	}
	var host string
	if err := config["host"].As(&host); err != nil || host != "h" {
		t.Fatalf("expected host h, got %q (%v)", host, err)
	}
}

func TestEmptyStringsToNull(t *testing.T) {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"empty": tftypes.String,
		"set":   tftypes.String,
		"list":  tftypes.Set{ElementType: tftypes.String},
	}}
	value := tftypes.NewValue(objType, map[string]tftypes.Value{
		"empty": tftypes.NewValue(tftypes.String, ""),
		"set":   tftypes.NewValue(tftypes.String, "v"),
		"list":  tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "a")}),
	})

	var diags diag.Diagnostics
	result := emptyStringsToNull("config", value, &diags)
	if diags.HasError() {
		t.Fatal(diags)
	}
	var attrs map[string]tftypes.Value
	if err := result.As(&attrs); err != nil {
		t.Fatal(err)
	}
	if !attrs["empty"].IsNull() {
		t.Errorf("expected \"\" to become null, got %v", attrs["empty"])
	}
	if !attrs["set"].Equal(tftypes.NewValue(tftypes.String, "v")) {
		t.Errorf("expected non-empty string to be kept, got %v", attrs["set"])
	}
	if attrs["list"].IsNull() {
		t.Errorf("expected set to be kept")
	}
}
