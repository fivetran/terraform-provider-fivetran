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

func TestAccountExternalLoggingV2ModifyPlanOnCreateConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyObject, diags := core.MapToDynamic(ctx, map[string]interface{}{})
	if diags.HasError() {
		t.Fatalf("empty config diagnostics: %v", diags)
	}
	apiKeyOnly, diags := core.MapToDynamic(ctx, map[string]interface{}{"api_key": "secret"})
	if diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}

	tests := []struct {
		name       string
		config     types.Dynamic
		wantErrors int
	}{
		{name: "unknown dynamic is deferred", config: types.DynamicUnknown()},
		{name: "unknown underlying object is deferred", config: types.DynamicValue(types.ObjectUnknown(map[string]attr.Type{}))},
		{name: "known empty object reports missing required field", config: emptyObject, wantErrors: 1},
		{name: "known config with required field is accepted", config: apiKeyOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := &sync.Map{}
			cache.Store(logServiceMetadataCachePrefix+"datadog_log", &metadata.ConnectorMetadata{
				Config: metadata.Property{
					Required: []string{"api_key"},
					Properties: map[string]*metadata.Property{
						"api_key": {Type: "string"},
					},
				},
			})

			r := &accountExternalLoggingV2{}
			var configureResp resource.ConfigureResponse
			r.Configure(ctx, resource.ConfigureRequest{
				ProviderData: &core.ProviderResourceData{MetadataCache: cache},
			}, &configureResp)
			if configureResp.Diagnostics.HasError() {
				t.Fatalf("configure diagnostics: %v", configureResp.Diagnostics)
			}

			schema := fivetranSchema.AccountExternalLoggingV2ResourceSchema()
			plan := model.AccountExternalLoggingV2ResourceModel{
				Id:            types.StringUnknown(),
				Service:       types.StringValue("datadog_log"),
				Enabled:       types.BoolValue(true),
				RunSetupTests: types.BoolValue(false),
				Config:        tt.config,
			}
			var object types.Object
			if d := tfsdk.ValueFrom(ctx, plan, schema.Type(), &object); d.HasError() {
				t.Fatalf("ValueFrom diagnostics: %v", d)
			}
			raw, err := object.ToTerraformValue(ctx)
			if err != nil {
				t.Fatalf("converting plan to Terraform value: %v", err)
			}

			req := resource.ModifyPlanRequest{
				Plan:  tfsdk.Plan{Raw: raw, Schema: schema},
				State: tfsdk.State{Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), nil), Schema: schema},
			}
			var resp resource.ModifyPlanResponse
			r.ModifyPlan(ctx, req, &resp)

			assertErrorCount(t, resp.Diagnostics, tt.wantErrors)
		})
	}
}

func accountExternalLoggingV2ForPlan(t *testing.T, services map[string]*metadata.ConnectorMetadata) *accountExternalLoggingV2 {
	t.Helper()

	cache := &sync.Map{}
	for service, meta := range services {
		cache.Store(logServiceMetadataCachePrefix+service, meta)
	}

	r := &accountExternalLoggingV2{}
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &core.ProviderResourceData{MetadataCache: cache},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", resp.Diagnostics)
	}
	return r
}

func accountExternalLoggingV2PlanValue(t *testing.T, service string, config map[string]interface{}) tftypes.Value {
	t.Helper()
	ctx := context.Background()

	configValue, diags := core.MapToDynamic(ctx, config)
	if diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}

	schema := fivetranSchema.AccountExternalLoggingV2ResourceSchema()
	var object types.Object
	if d := tfsdk.ValueFrom(ctx, model.AccountExternalLoggingV2ResourceModel{
		Id:            types.StringValue("account_id"),
		Service:       types.StringValue(service),
		Enabled:       types.BoolValue(true),
		RunSetupTests: types.BoolValue(false),
		Config:        configValue,
	}, schema.Type(), &object); d.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", d)
	}
	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting model to Terraform value: %v", err)
	}
	return raw
}

func accountExternalLoggingV2UpdateRequest(t *testing.T, planService string, planConfig map[string]interface{}, stateService string, stateConfig map[string]interface{}) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.AccountExternalLoggingV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Raw: accountExternalLoggingV2PlanValue(t, planService, planConfig), Schema: schema},
		State: tfsdk.State{Raw: accountExternalLoggingV2PlanValue(t, stateService, stateConfig), Schema: schema},
	}
}

func diagnosticsText(diags diag.Diagnostics) string {
	var text strings.Builder
	for _, d := range diags {
		text.WriteString(d.Summary() + " " + d.Detail() + "\n")
	}
	return text.String()
}

func requiredApiKeyMetadata() *metadata.ConnectorMetadata {
	return &metadata.ConnectorMetadata{Config: metadata.Property{
		Required: []string{"api_key"},
		Properties: map[string]*metadata.Property{
			"api_key":  {Type: "string", Immutable: true},
			"hostname": {Type: "string"},
		},
	}}
}

// Changing service replaces the log service. Terraform deletes the old one before creating the new one,
// so the new service must be validated like a create, also when config is empty.
func TestAccountExternalLoggingV2ModifyPlanValidatesRequiredFieldsWhenServiceChanges(t *testing.T) {
	t.Parallel()

	r := accountExternalLoggingV2ForPlan(t, map[string]*metadata.ConnectorMetadata{
		"new_relic_log": requiredApiKeyMetadata(),
		"datadog_log":   requiredApiKeyMetadata(),
	})

	for name, planConfig := range map[string]map[string]interface{}{
		"empty config":             {},
		"config without the field": {"hostname": "host"},
	} {
		req := accountExternalLoggingV2UpdateRequest(t, "new_relic_log", planConfig, "datadog_log", map[string]interface{}{"api_key": "secret"})
		var resp resource.ModifyPlanResponse
		r.ModifyPlan(context.Background(), req, &resp)

		if resp.Diagnostics.ErrorsCount() != 1 {
			t.Errorf("%s: errors = %d, want 1 missing required field: %v", name, resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
		}
		if resp.Diagnostics.WarningsCount() != 1 {
			t.Errorf("%s: warnings = %d, want 1 replacement warning: %v", name, resp.Diagnostics.WarningsCount(), resp.Diagnostics)
		}
	}

	// the same required field present: only the replacement warning
	req := accountExternalLoggingV2UpdateRequest(t, "new_relic_log", map[string]interface{}{"api_key": "secret"}, "datadog_log", map[string]interface{}{"api_key": "secret"})
	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)
	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 1)
}

func TestAccountExternalLoggingV2ModifyPlanEmptyConfigIsAcceptedWhenServiceIsUnchanged(t *testing.T) {
	t.Parallel()

	r := accountExternalLoggingV2ForPlan(t, map[string]*metadata.ConnectorMetadata{"datadog_log": requiredApiKeyMetadata()})
	req := accountExternalLoggingV2UpdateRequest(t, "datadog_log", map[string]interface{}{}, "datadog_log", map[string]interface{}{"api_key": "secret"})

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

// The diagnostics describe a log service, not a connection: replacing it does not resync any data.
func TestAccountExternalLoggingV2ModifyPlanDiagnosticsDescribeTheLogService(t *testing.T) {
	t.Parallel()

	r := accountExternalLoggingV2ForPlan(t, map[string]*metadata.ConnectorMetadata{"datadog_log": requiredApiKeyMetadata()})
	forbidden := []string{"connection", "resync"}
	assertDescribesLogService := func(t *testing.T, name string, diags diag.Diagnostics) {
		t.Helper()
		text := strings.ToLower(diagnosticsText(diags))
		if !strings.Contains(text, "external logging service") {
			t.Errorf("%s: diagnostics should name the external logging service: %s", name, text)
		}
		for _, word := range forbidden {
			if strings.Contains(text, word) {
				t.Errorf("%s: diagnostics should not mention %q: %s", name, word, text)
			}
		}
	}

	// immutable config field changed
	req := accountExternalLoggingV2UpdateRequest(t, "datadog_log", map[string]interface{}{"api_key": "new"}, "datadog_log", map[string]interface{}{"api_key": "old"})
	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)
	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertRequiresReplacePaths(t, resp.RequiresReplace, []path.Path{path.Root("config").AtName("api_key")})
	assertDescribesLogService(t, "immutable change", resp.Diagnostics)

	// required field missing on create
	createReq := resource.ModifyPlanRequest{
		Plan: tfsdk.Plan{
			Raw:    accountExternalLoggingV2PlanValue(t, "datadog_log", map[string]interface{}{"hostname": "host"}),
			Schema: fivetranSchema.AccountExternalLoggingV2ResourceSchema(),
		},
		State: tfsdk.State{
			Raw:    tftypes.NewValue(fivetranSchema.AccountExternalLoggingV2ResourceSchema().Type().TerraformType(context.Background()), nil),
			Schema: fivetranSchema.AccountExternalLoggingV2ResourceSchema(),
		},
	}
	var createResp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), createReq, &createResp)
	assertErrorCount(t, createResp.Diagnostics, 1)
	assertDescribesLogService(t, "missing required field", createResp.Diagnostics)

	// service changed
	changeReq := accountExternalLoggingV2UpdateRequest(t, "datadog_log", map[string]interface{}{"api_key": "k"}, "cloudwatch", map[string]interface{}{"api_key": "k"})
	var changeResp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), changeReq, &changeResp)
	assertWarningCount(t, changeResp.Diagnostics, 1)
	assertDescribesLogService(t, "service change", changeResp.Diagnostics)
}
