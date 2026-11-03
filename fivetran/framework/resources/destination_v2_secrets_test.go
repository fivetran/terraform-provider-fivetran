package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func withConfigSecrets(t *testing.T, data model.DestinationV2ResourceModel, secrets map[string]interface{}) model.DestinationV2ResourceModel {
	t.Helper()
	value, diags := core.MapToDynamic(context.Background(), secrets)
	if diags.HasError() {
		t.Fatalf("config_secrets dynamic diagnostics: %v", diags)
	}
	data.ConfigSecrets = value
	return data
}

func TestDestinationV2ValidateConfigAcceptsSecretFieldInConfigSecrets(t *testing.T) {
	t.Parallel()

	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
	data := withConfigSecrets(t,
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"}),
		map[string]interface{}{"password": "secret"})

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, data), &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestDestinationV2ValidateConfigRejectsInvalidConfigSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   map[string]interface{}
		secrets  map[string]interface{}
		wantPath path.Path
	}{
		{
			name:     "non-secret field",
			config:   map[string]interface{}{"port": 443},
			secrets:  map[string]interface{}{"host": "h"},
			wantPath: path.Root("config_secrets").AtName("host"),
		},
		{
			name:     "field set in both config and config_secrets",
			config:   map[string]interface{}{"password": "a"},
			secrets:  map[string]interface{}{"password": "b"},
			wantPath: path.Root("config_secrets").AtName("password"),
		},
		{
			name:     "unknown field",
			config:   map[string]interface{}{"host": "h"},
			secrets:  map[string]interface{}{"pasword": "typo"},
			wantPath: path.Root("config_secrets").AtName("pasword"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
			data := withConfigSecrets(t, destinationV2Model(t, "snowflake", "group_id", tt.config), tt.secrets)

			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, data), &resp)

			assertErrorCount(t, resp.Diagnostics, 1)
			assertDiagnosticPath(t, resp.Diagnostics, tt.wantPath)
		})
	}
}

func TestDestinationV2ModifyPlanCountsConfigSecretsAsRequiredFields(t *testing.T) {
	t.Parallel()

	meta := snowflakeDestinationMetadata()
	meta.Config.Required = []string{"host", "password"}
	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", meta), nil)

	// Terraform plans write-only attributes as null; their values are only in the config.
	plan := destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"})
	req := destinationV2ModifyPlanCreateRequest(t, plan)
	req.Config = tfsdk.Config{
		Raw:    destinationV2TerraformValue(t, withConfigSecrets(t, plan, map[string]interface{}{"password": "secret"})),
		Schema: fivetranSchema.DestinationV2ResourceSchema(),
	}

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestDestinationV2ModifyPlanSkipsRequiredFieldsWhileConfigSecretsUnknown(t *testing.T) {
	t.Parallel()

	meta := snowflakeDestinationMetadata()
	meta.Config.Required = []string{"host", "password"}
	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", meta), nil)

	plan := destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"})
	config := plan
	config.ConfigSecrets = types.DynamicUnknown()
	req := destinationV2ModifyPlanCreateRequest(t, plan)
	req.Config = tfsdk.Config{Raw: destinationV2TerraformValue(t, config), Schema: fivetranSchema.DestinationV2ResourceSchema()}

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestMergeConfigSecretsDoesNotModifyConfig(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{"host": "h"}
	merged := mergeConfigSecrets(config, map[string]interface{}{"password": "secret"})

	if want := map[string]interface{}{"host": "h", "password": "secret"}; !reflect.DeepEqual(merged, want) {
		t.Fatalf("merged config: got %v, want %v", merged, want)
	}
	if want := map[string]interface{}{"host": "h"}; !reflect.DeepEqual(config, want) {
		t.Fatalf("config was modified: got %v, want %v", config, want)
	}
}

func TestApplyConfigSecretsToPatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		patch map[string]interface{}
		send  bool
		want  map[string]interface{}
	}{
		{
			name:  "version changed sends secrets",
			patch: map[string]interface{}{"host": "h"},
			send:  true,
			want:  map[string]interface{}{"host": "h", "password": "secret"},
		},
		{
			name:  "version unchanged does not send secrets",
			patch: map[string]interface{}{"host": "h"},
			send:  false,
			want:  map[string]interface{}{"host": "h"},
		},
		{
			name:  "field moved from config is not cleared",
			patch: map[string]interface{}{"password": nil},
			send:  false,
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			applyConfigSecretsToPatch(tt.patch, map[string]interface{}{"password": "secret"}, tt.send)

			if !reflect.DeepEqual(tt.patch, tt.want) {
				t.Fatalf("patch: got %v, want %v", tt.patch, tt.want)
			}
		})
	}
}
