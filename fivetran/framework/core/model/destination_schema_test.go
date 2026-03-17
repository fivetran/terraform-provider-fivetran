package model

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func makeModel(destinationSchema string) *ConnectorV2ResourceModel {
	return &ConnectorV2ResourceModel{
		DestinationSchema: types.StringValue(destinationSchema),
	}
}

func TestParseDestinationSchemaConfigs_Empty(t *testing.T) {
	m := makeModel("")
	_, err := m.ParseDestinationSchemaConfigs()
	if err == nil {
		t.Fatal("expected error for empty destination_schema, got nil")
	}
}

func TestParseDestinationSchemaConfigs_PlainSchema(t *testing.T) {
	m := makeModel("my_schema")
	configs, err := m.ParseDestinationSchemaConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(configs))
	}
	// First candidate: schema
	if configs[0]["schema"] != "my_schema" {
		t.Errorf("expected configs[0][schema] = my_schema, got %v", configs[0]["schema"])
	}
	if _, ok := configs[0]["schema_prefix"]; ok {
		t.Error("configs[0] should not have schema_prefix")
	}
	// Second candidate: schema_prefix
	if configs[1]["schema_prefix"] != "my_schema" {
		t.Errorf("expected configs[1][schema_prefix] = my_schema, got %v", configs[1]["schema_prefix"])
	}
	if _, ok := configs[1]["schema"]; ok {
		t.Error("configs[1] should not have schema")
	}
}

func TestParseDestinationSchemaConfigs_DotSeparated(t *testing.T) {
	m := makeModel("my_schema.my_table")
	configs, err := m.ParseDestinationSchemaConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(configs))
	}
	// First candidate: schema + table
	if configs[0]["schema"] != "my_schema" {
		t.Errorf("expected configs[0][schema] = my_schema, got %v", configs[0]["schema"])
	}
	if configs[0]["table"] != "my_table" {
		t.Errorf("expected configs[0][table] = my_table, got %v", configs[0]["table"])
	}
	// Second candidate: schema + table_group_name
	if configs[1]["schema"] != "my_schema" {
		t.Errorf("expected configs[1][schema] = my_schema, got %v", configs[1]["schema"])
	}
	if configs[1]["table_group_name"] != "my_table" {
		t.Errorf("expected configs[1][table_group_name] = my_table, got %v", configs[1]["table_group_name"])
	}
	if _, ok := configs[1]["table"]; ok {
		t.Error("configs[1] should not have table")
	}
}

func TestParseDestinationSchemaConfigs_DotSeparatedOrder(t *testing.T) {
	// Verify retry order: table is tried before table_group_name
	m := makeModel("schema.table")
	configs, _ := m.ParseDestinationSchemaConfigs()
	if _, ok := configs[0]["table"]; !ok {
		t.Error("first candidate should use 'table', not 'table_group_name'")
	}
	if _, ok := configs[1]["table_group_name"]; !ok {
		t.Error("second candidate should use 'table_group_name'")
	}
}

func TestParseDestinationSchemaConfigs_PlainOrder(t *testing.T) {
	// Verify retry order: schema is tried before schema_prefix
	m := makeModel("schema")
	configs, _ := m.ParseDestinationSchemaConfigs()
	if _, ok := configs[0]["schema"]; !ok {
		t.Error("first candidate should use 'schema'")
	}
	if _, ok := configs[1]["schema_prefix"]; !ok {
		t.Error("second candidate should use 'schema_prefix'")
	}
}

func TestParseDestinationSchemaConfigs_MultipleDots(t *testing.T) {
	// Only the first dot should be used as separator; remaining dots stay in the table part.
	m := makeModel("schema.table.extra")
	configs, err := m.ParseDestinationSchemaConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configs[0]["schema"] != "schema" {
		t.Errorf("expected schema = 'schema', got %v", configs[0]["schema"])
	}
	if configs[0]["table"] != "table.extra" {
		t.Errorf("expected table = 'table.extra', got %v", configs[0]["table"])
	}
	if configs[1]["table_group_name"] != "table.extra" {
		t.Errorf("expected table_group_name = 'table.extra', got %v", configs[1]["table_group_name"])
	}
}

func TestParseDestinationSchemaConfigs_LeadingDot(t *testing.T) {
	// ".table" — empty schema name, table = "table"
	m := makeModel(".table")
	configs, err := m.ParseDestinationSchemaConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configs[0]["schema"] != "" {
		t.Errorf("expected empty schema, got %v", configs[0]["schema"])
	}
	if configs[0]["table"] != "table" {
		t.Errorf("expected table = 'table', got %v", configs[0]["table"])
	}
}

func TestParseDestinationSchemaConfigs_TrailingDot(t *testing.T) {
	// "schema." — schema = "schema", table = ""
	m := makeModel("schema.")
	configs, err := m.ParseDestinationSchemaConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configs[0]["schema"] != "schema" {
		t.Errorf("expected schema = 'schema', got %v", configs[0]["schema"])
	}
	if configs[0]["table"] != "" {
		t.Errorf("expected empty table, got %v", configs[0]["table"])
	}
}

// TestReadFromContainer_PreservesExistingDestinationSchema verifies that readFromContainer
// does NOT overwrite destination_schema when the model already has a value (normal read/refresh).
func TestReadFromContainer_PreservesExistingDestinationSchema(t *testing.T) {
	m := &ConnectorV2ResourceModel{
		DestinationSchema: types.StringValue("my_schema.my_table"),
	}
	c := ConnectorModelContainer{Schema: "my_schema"}
	m.readFromContainer(c)
	if m.DestinationSchema.ValueString() != "my_schema.my_table" {
		t.Errorf("expected destination_schema to be preserved as 'my_schema.my_table', got %v", m.DestinationSchema.ValueString())
	}
}

// TestReadFromContainer_FallsBackToAPISchemaWhenUnknown verifies that readFromContainer
// uses the API schema value when destination_schema is unknown (e.g. during import).
func TestReadFromContainer_FallsBackToAPISchemaWhenUnknown(t *testing.T) {
	m := &ConnectorV2ResourceModel{
		DestinationSchema: types.StringUnknown(),
	}
	c := ConnectorModelContainer{Schema: "my_schema"}
	m.readFromContainer(c)
	if m.DestinationSchema.ValueString() != "my_schema" {
		t.Errorf("expected destination_schema = 'my_schema' from API, got %v", m.DestinationSchema.ValueString())
	}
}

// TestReadFromContainer_FallsBackToAPISchemaWhenNull verifies that readFromContainer
// uses the API schema value when destination_schema is null.
func TestReadFromContainer_FallsBackToAPISchemaWhenNull(t *testing.T) {
	m := &ConnectorV2ResourceModel{
		DestinationSchema: types.StringNull(),
	}
	c := ConnectorModelContainer{Schema: "my_schema"}
	m.readFromContainer(c)
	if m.DestinationSchema.ValueString() != "my_schema" {
		t.Errorf("expected destination_schema = 'my_schema' from API, got %v", m.DestinationSchema.ValueString())
	}
}
