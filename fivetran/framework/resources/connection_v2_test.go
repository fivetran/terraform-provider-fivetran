package resources_test

import (
	"context"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/resources"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestConnectionV2SchemaShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := resources.ConnectionV2()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}

	attrs := schemaResp.Schema.Attributes
	if schemaResp.Schema.Version != 0 {
		t.Fatalf("unexpected schema version: got %d, want 0", schemaResp.Schema.Version)
	}
	if _, ok := attrs["destination_schema"]; ok {
		t.Fatal("fivetran_connection_v2 must not expose destination_schema as a root attribute")
	}
	if _, ok := attrs["local_processing_agent_id"]; ok {
		t.Fatal("fivetran_connection_v2 must not expose deprecated local_processing_agent_id")
	}
	if _, ok := attrs["paused"]; ok {
		t.Fatal("fivetran_connection_v2 must not expose paused; pause state is managed by fivetran_connection_v2_pause_state")
	}

	assertStringAttribute(t, attrs, "service", true, false, false)
	assertStringAttribute(t, attrs, "group_id", true, false, false)
	assertDynamicAttribute(t, attrs, "config", true, true, false)
	assertDynamicAttribute(t, attrs, "auth", true, false, true)
	assertInt64Attribute(t, attrs, "sync_frequency", false, true, true)
	assertBoolAttribute(t, attrs, "run_setup_tests", false, true, false)
	assertBoolAttribute(t, attrs, "trust_certificates", false, true, false)
	assertBoolAttribute(t, attrs, "trust_fingerprints", false, true, false)

	status, ok := attrs["status"].(resourceSchema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("status has type %T, want SingleNestedAttribute", attrs["status"])
	}
	if !status.Computed {
		t.Fatal("status should be computed")
	}
	if _, ok := status.Attributes["tasks"].(resourceSchema.SetNestedAttribute); !ok {
		t.Fatalf("status.tasks has type %T, want SetNestedAttribute", status.Attributes["tasks"])
	}
	if _, ok := status.Attributes["warnings"].(resourceSchema.SetNestedAttribute); !ok {
		t.Fatalf("status.warnings has type %T, want SetNestedAttribute", status.Attributes["warnings"])
	}
}

func TestConnectionV2NotRegistered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	p := framework.FivetranProvider()
	for _, resourceFactory := range p.Resources(ctx) {
		r := resourceFactory()
		var metadataResp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "fivetran"}, &metadataResp)
		if metadataResp.TypeName == "fivetran_connection_v2" {
			t.Fatal("fivetran_connection_v2 must remain unregistered until the registration ticket")
		}
	}

	var metadataResp provider.MetadataResponse
	p.Metadata(ctx, provider.MetadataRequest{}, &metadataResp)
	if metadataResp.TypeName != "fivetran" {
		t.Fatalf("unexpected provider type name: got %q", metadataResp.TypeName)
	}
}

func assertStringAttribute(t *testing.T, attrs map[string]resourceSchema.Attribute, name string, required, optional, computed bool) {
	t.Helper()

	attr, ok := attrs[name].(resourceSchema.StringAttribute)
	if !ok {
		t.Fatalf("%s has type %T, want StringAttribute", name, attrs[name])
	}
	assertAttributeMode(t, name, attr.Required, attr.Optional, attr.Computed, required, optional, computed)
}

func assertBoolAttribute(t *testing.T, attrs map[string]resourceSchema.Attribute, name string, required, optional, computed bool) {
	t.Helper()

	attr, ok := attrs[name].(resourceSchema.BoolAttribute)
	if !ok {
		t.Fatalf("%s has type %T, want BoolAttribute", name, attrs[name])
	}
	assertAttributeMode(t, name, attr.Required, attr.Optional, attr.Computed, required, optional, computed)
}

func assertInt64Attribute(t *testing.T, attrs map[string]resourceSchema.Attribute, name string, required, optional, computed bool) {
	t.Helper()

	attr, ok := attrs[name].(resourceSchema.Int64Attribute)
	if !ok {
		t.Fatalf("%s has type %T, want Int64Attribute", name, attrs[name])
	}
	assertAttributeMode(t, name, attr.Required, attr.Optional, attr.Computed, required, optional, computed)
}

func assertDynamicAttribute(t *testing.T, attrs map[string]resourceSchema.Attribute, name string, optional, computed, sensitive bool) {
	t.Helper()

	attr, ok := attrs[name].(resourceSchema.DynamicAttribute)
	if !ok {
		t.Fatalf("%s has type %T, want DynamicAttribute", name, attrs[name])
	}
	assertAttributeMode(t, name, attr.Required, attr.Optional, attr.Computed, false, optional, computed)
	if attr.Sensitive != sensitive {
		t.Fatalf("%s sensitive = %v, want %v", name, attr.Sensitive, sensitive)
	}
}

func assertAttributeMode(t *testing.T, name string, gotRequired, gotOptional, gotComputed, wantRequired, wantOptional, wantComputed bool) {
	t.Helper()

	if gotRequired != wantRequired || gotOptional != wantOptional || gotComputed != wantComputed {
		t.Fatalf(
			"%s mode = required:%v optional:%v computed:%v, want required:%v optional:%v computed:%v",
			name,
			gotRequired,
			gotOptional,
			gotComputed,
			wantRequired,
			wantOptional,
			wantComputed,
		)
	}
}
