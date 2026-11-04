package resources_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Exercise the actual protocol upgrade handler with schema-version-0 JSON from before
// config_secrets existed. An upgrade must add null attributes without losing config.
func TestConnectionV2ConfigSecretsUpgradeOldState(t *testing.T) {
	ctx := context.Background()
	schema := fivetranSchema.ConnectionV2ResourceSchema()
	if schema.Version != 0 {
		t.Fatal("update this test to exercise the schema-version-0 state upgrader")
	}
	oldState := make(map[string]interface{})
	for name := range schema.Attributes {
		if name != "config_secrets" && name != "config_secrets_version" {
			oldState[name] = nil
		}
	}
	oldState["id"] = "connection_id"
	oldState["service"] = "postgres"
	oldState["group_id"] = "group_id"
	oldState["config"] = map[string]interface{}{
		"type":  []interface{}{"object", map[string]string{"host": "string", "password": "string"}},
		"value": map[string]string{"host": "db.example.com", "password": "legacy-password"},
	}
	raw, err := json.Marshal(oldState)
	if err != nil {
		t.Fatal(err)
	}
	server := providerserver.NewProtocol6(framework.FivetranProvider())()
	resp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: "fivetran_connection_v2",
		Version:  0,
		RawState: &tfprotov6.RawState{JSON: raw},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("state upgrade: %s: %s", d.Summary, d.Detail)
		}
	}
	if resp.UpgradedState == nil {
		t.Fatal("missing upgraded state")
	}
	value, err := resp.UpgradedState.Unmarshal(schema.Type().TerraformType(ctx))
	if err != nil {
		t.Fatal(err)
	}
	state := tfsdk.State{Raw: value, Schema: schema}
	var data model.ConnectionV2ResourceModel
	if diags := state.Get(ctx, &data); diags.HasError() {
		t.Fatalf("decode upgraded state: %v", diags)
	}
	if data.Id.ValueString() != "connection_id" || data.Service.ValueString() != "postgres" || data.GroupId.ValueString() != "group_id" {
		t.Fatal("upgrade changed resource identity")
	}
	if !data.ConfigSecrets.IsNull() || !data.ConfigSecretsVersion.IsNull() {
		t.Fatal("new optional attributes must be null in upgraded state")
	}
	config, diags := core.DynamicToMap(ctx, data.Config)
	if diags.HasError() {
		t.Fatalf("decode legacy config: %v", diags)
	}
	if config["host"] != "db.example.com" || config["password"] != "legacy-password" {
		t.Fatal("upgrade changed legacy config values")
	}
}
