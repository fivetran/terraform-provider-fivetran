package resources

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func externalLoggingV2ApiKeyMetadata() *metadata.LogServiceMetadata {
	return &metadata.LogServiceMetadata{Config: metadata.Property{
		Required: []string{"api_key"},
		Properties: map[string]*metadata.Property{
			"api_key":  {Type: "string", Format: "password", Immutable: true},
			"hostname": {Type: "string"},
		},
	}}
}

func externalLoggingV2ForPlan(t *testing.T, services map[string]*metadata.LogServiceMetadata) *externalLoggingV2 {
	t.Helper()

	cache := &sync.Map{}
	for service, meta := range services {
		cache.Store(core.LogServiceMetadataCacheKey(service), meta)
	}

	r := &externalLoggingV2{}
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &core.ProviderResourceData{MetadataCache: cache},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", resp.Diagnostics)
	}
	return r
}

func externalLoggingV2Model(groupId, service string, config types.Dynamic) model.ExternalLoggingV2ResourceModel {
	return model.ExternalLoggingV2ResourceModel{
		Id:                   types.StringValue(groupId),
		GroupId:              types.StringValue(groupId),
		Service:              types.StringValue(service),
		Enabled:              types.BoolValue(true),
		RunSetupTests:        types.BoolValue(false),
		Config:               config,
		ConfigSecrets:        types.DynamicNull(),
		ConfigSecretsVersion: types.Int64Null(),
	}
}

func externalLoggingV2Value(t *testing.T, groupId, service string, config types.Dynamic) tftypes.Value {
	t.Helper()
	return externalLoggingV2ModelValue(t, externalLoggingV2Model(groupId, service, config))
}

func externalLoggingV2ModelValue(t *testing.T, data model.ExternalLoggingV2ResourceModel) tftypes.Value {
	t.Helper()
	ctx := context.Background()

	schema := fivetranSchema.ExternalLoggingV2ResourceSchema()
	var object types.Object
	if d := tfsdk.ValueFrom(ctx, data, schema.Type(), &object); d.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", d)
	}
	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting model to Terraform value: %v", err)
	}
	return raw
}

func externalLoggingV2ConfigValue(t *testing.T, config map[string]interface{}) types.Dynamic {
	t.Helper()
	value, diags := core.MapToDynamic(context.Background(), config)
	if diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}
	return value
}

func externalLoggingV2CreateRequest(t *testing.T, config types.Dynamic) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.ExternalLoggingV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Raw: externalLoggingV2Value(t, "group_id", "datadog_log", config), Schema: schema},
		State: tfsdk.State{Raw: tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil), Schema: schema},
	}
}

func externalLoggingV2UpdateRequest(t *testing.T, planGroup, planService string, planConfig map[string]interface{}, stateGroup, stateService string, stateConfig map[string]interface{}) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.ExternalLoggingV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Raw: externalLoggingV2Value(t, planGroup, planService, externalLoggingV2ConfigValue(t, planConfig)), Schema: schema},
		State: tfsdk.State{Raw: externalLoggingV2Value(t, stateGroup, stateService, externalLoggingV2ConfigValue(t, stateConfig)), Schema: schema},
	}
}

func TestExternalLoggingV2ModifyPlanOnCreateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     types.Dynamic
		wantErrors int
	}{
		{name: "unknown dynamic is deferred", config: types.DynamicUnknown()},
		{name: "unknown underlying object is deferred", config: types.DynamicValue(types.ObjectUnknown(map[string]attr.Type{}))},
		{name: "known empty object reports missing required field", config: externalLoggingV2ConfigValue(t, map[string]interface{}{}), wantErrors: 1},
		{name: "known config with required field is accepted", config: externalLoggingV2ConfigValue(t, map[string]interface{}{"api_key": "secret"})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})

			var resp resource.ModifyPlanResponse
			r.ModifyPlan(context.Background(), externalLoggingV2CreateRequest(t, tt.config), &resp)

			assertErrorCount(t, resp.Diagnostics, tt.wantErrors)
		})
	}
}

// Without metadata for the service the API validates the config, so no plan-time diagnostics are produced.
func TestExternalLoggingV2ModifyPlanWithoutMetadata(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	// the cached answer of a metadata API that has no description for the service
	cache.Store("no-metadata:"+core.LogServiceMetadataCacheKey("datadog_log"), true)
	r := &externalLoggingV2{}
	var configureResp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &core.ProviderResourceData{MetadataCache: cache},
	}, &configureResp)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), externalLoggingV2CreateRequest(t, externalLoggingV2ConfigValue(t, map[string]interface{}{})), &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestExternalLoggingV2ModifyPlanImmutableFieldRequiresReplace(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
	req := externalLoggingV2UpdateRequest(t,
		"group_id", "datadog_log", map[string]interface{}{"api_key": "new"},
		"group_id", "datadog_log", map[string]interface{}{"api_key": "old"})

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertRequiresReplacePaths(t, resp.RequiresReplace, []path.Path{path.Root("config").AtName("api_key")})
	assertExternalLoggingV2Wording(t, resp.Diagnostics)
}

func TestExternalLoggingV2ModifyPlanMutableFieldUpdatesInPlace(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
	req := externalLoggingV2UpdateRequest(t,
		"group_id", "datadog_log", map[string]interface{}{"api_key": "k", "hostname": "new"},
		"group_id", "datadog_log", map[string]interface{}{"api_key": "k", "hostname": "old"})

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	assertRequiresReplacePaths(t, resp.RequiresReplace, nil)
}

// Changing service or group_id replaces the log service, so the new one is validated like a create.
func TestExternalLoggingV2ModifyPlanValidatesRequiredFieldsOnReplacement(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{
		"datadog_log":   externalLoggingV2ApiKeyMetadata(),
		"new_relic_log": externalLoggingV2ApiKeyMetadata(),
	})

	// service changed: missing required field + replacement warning
	req := externalLoggingV2UpdateRequest(t,
		"group_id", "new_relic_log", map[string]interface{}{"hostname": "host"},
		"group_id", "datadog_log", map[string]interface{}{"api_key": "secret"})
	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)
	assertErrorCount(t, resp.Diagnostics, 1)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertExternalLoggingV2Wording(t, resp.Diagnostics)

	// group_id changed with an empty config: missing required field
	req = externalLoggingV2UpdateRequest(t,
		"other_group", "datadog_log", map[string]interface{}{},
		"group_id", "datadog_log", map[string]interface{}{"api_key": "secret"})
	resp = resource.ModifyPlanResponse{}
	r.ModifyPlan(context.Background(), req, &resp)
	assertErrorCount(t, resp.Diagnostics, 1)

	// unchanged service and group with an empty config: nothing to validate
	req = externalLoggingV2UpdateRequest(t,
		"group_id", "datadog_log", map[string]interface{}{},
		"group_id", "datadog_log", map[string]interface{}{"api_key": "secret"})
	resp = resource.ModifyPlanResponse{}
	r.ModifyPlan(context.Background(), req, &resp)
	assertNoDiagnostics(t, resp.Diagnostics)
}

func assertExternalLoggingV2Wording(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	text := strings.ToLower(diagnosticsText(diags))
	if !strings.Contains(text, "external logging service") {
		t.Errorf("diagnostics should name the external logging service: %s", text)
	}
	for _, word := range []string{"connection", "resync", "account"} {
		if strings.Contains(text, word) {
			t.Errorf("diagnostics should not mention %q: %s", word, text)
		}
	}
}

// Terraform plans write-only attributes as null, so config_secrets is only in the config.
func externalLoggingV2CreateRequestWithSecrets(t *testing.T, config map[string]interface{}, secrets types.Dynamic) resource.ModifyPlanRequest {
	t.Helper()
	req := externalLoggingV2CreateRequest(t, externalLoggingV2ConfigValue(t, config))
	data := externalLoggingV2Model("group_id", "datadog_log", externalLoggingV2ConfigValue(t, config))
	data.ConfigSecrets = secrets
	req.Config = tfsdk.Config{Raw: externalLoggingV2ModelValue(t, data), Schema: fivetranSchema.ExternalLoggingV2ResourceSchema()}
	return req
}

func TestExternalLoggingV2ModifyPlanCountsConfigSecretsAsRequiredFields(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
	req := externalLoggingV2CreateRequestWithSecrets(t,
		map[string]interface{}{"hostname": "host"},
		externalLoggingV2ConfigValue(t, map[string]interface{}{"api_key": "secret"}))

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestExternalLoggingV2ModifyPlanSkipsRequiredFieldsWhileConfigSecretsUnknown(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
	req := externalLoggingV2CreateRequestWithSecrets(t, map[string]interface{}{"hostname": "host"}, types.DynamicUnknown())

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestExternalLoggingV2ValidateConfigSecrets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   map[string]interface{}
		secrets  map[string]interface{}
		wantPath *path.Path
	}{
		{
			name:    "secret field is accepted",
			config:  map[string]interface{}{"hostname": "host"},
			secrets: map[string]interface{}{"api_key": "secret"},
		},
		{
			name:    "secrets alone are validated",
			config:  map[string]interface{}{},
			secrets: map[string]interface{}{"api_key": "secret"},
		},
		{
			name:     "non-secret field",
			config:   map[string]interface{}{"api_key": "secret"},
			secrets:  map[string]interface{}{"hostname": "host"},
			wantPath: pathPtr(path.Root("config_secrets").AtName("hostname")),
		},
		{
			name:     "field set in both config and config_secrets",
			config:   map[string]interface{}{"api_key": "a"},
			secrets:  map[string]interface{}{"api_key": "b"},
			wantPath: pathPtr(path.Root("config_secrets").AtName("api_key")),
		},
		{
			name:     "unknown field",
			config:   map[string]interface{}{},
			secrets:  map[string]interface{}{"api_kye": "typo"},
			wantPath: pathPtr(path.Root("config_secrets").AtName("api_kye")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
			data := externalLoggingV2Model("group_id", "datadog_log", externalLoggingV2ConfigValue(t, tt.config))
			data.ConfigSecrets = externalLoggingV2ConfigValue(t, tt.secrets)
			req := resource.ValidateConfigRequest{Config: tfsdk.Config{
				Raw:    externalLoggingV2ModelValue(t, data),
				Schema: fivetranSchema.ExternalLoggingV2ResourceSchema(),
			}}

			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), req, &resp)

			if tt.wantPath == nil {
				assertNoDiagnostics(t, resp.Diagnostics)
				return
			}
			assertErrorCount(t, resp.Diagnostics, 1)
			assertDiagnosticPath(t, resp.Diagnostics, *tt.wantPath)
		})
	}
}

func pathPtr(p path.Path) *path.Path {
	return &p
}

// Import leaves secrets out of state. Setting an immutable secret afterwards, as when migrating
// from fivetran_external_logging with import blocks, must not replace the log service.
func TestExternalLoggingV2ModifyPlanUntrackedImmutableSecretDoesNotReplace(t *testing.T) {
	t.Parallel()

	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": externalLoggingV2ApiKeyMetadata()})
	req := externalLoggingV2UpdateRequest(t,
		"group_id", "datadog_log", map[string]interface{}{"api_key": "secret", "hostname": "host"},
		"group_id", "datadog_log", map[string]interface{}{"hostname": "host"})

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertRequiresReplacePaths(t, resp.RequiresReplace, nil)
	assertExternalLoggingV2Wording(t, resp.Diagnostics)
}

// An immutable field that isn't a secret is in state after import, so adding it is a real change.
func TestExternalLoggingV2ModifyPlanUntrackedImmutableNonSecretReplaces(t *testing.T) {
	t.Parallel()

	meta := externalLoggingV2ApiKeyMetadata()
	meta.Config.Properties["region"] = &metadata.Property{Type: "string", Immutable: true}
	r := externalLoggingV2ForPlan(t, map[string]*metadata.LogServiceMetadata{"datadog_log": meta})
	req := externalLoggingV2UpdateRequest(t,
		"group_id", "datadog_log", map[string]interface{}{"region": "US1"},
		"group_id", "datadog_log", map[string]interface{}{"hostname": "host"})

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertRequiresReplacePaths(t, resp.RequiresReplace, []path.Path{path.Root("config").AtName("region")})
}
