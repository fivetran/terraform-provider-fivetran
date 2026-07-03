package resources

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConnectionV2ModifyPlanRequiresReplaceForImmutableDynamicFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		planConfig   map[string]interface{}
		stateConfig  map[string]interface{}
		planAuth     map[string]interface{}
		stateAuth    map[string]interface{}
		meta         *metadata.ConnectorMetadata
		wantPaths    []path.Path
		wantWarnings int
	}{
		{
			name:        "config immutable field changed",
			planConfig:  map[string]interface{}{"schema": "new"},
			stateConfig: map[string]interface{}{"schema": "old"},
			meta: &metadata.ConnectorMetadata{Config: metadata.Property{Properties: map[string]*metadata.Property{
				"schema": {Type: "string", Immutable: true},
			}}},
			wantPaths:    []path.Path{path.Root("config").AtName("schema")},
			wantWarnings: 1,
		},
		{
			name:        "config mutable field changed",
			planConfig:  map[string]interface{}{"schema": "new"},
			stateConfig: map[string]interface{}{"schema": "old"},
			meta: &metadata.ConnectorMetadata{Config: metadata.Property{Properties: map[string]*metadata.Property{
				"schema": {Type: "string"},
			}}},
		},
		{
			name:      "auth immutable field changed",
			planAuth:  map[string]interface{}{"refresh_token": "new"},
			stateAuth: map[string]interface{}{"refresh_token": "old"},
			meta: &metadata.ConnectorMetadata{Auth: metadata.Property{Properties: map[string]*metadata.Property{
				"refresh_token": {Type: "string", Immutable: true},
			}}},
			wantPaths:    []path.Path{path.Root("auth").AtName("refresh_token")},
			wantWarnings: 1,
		},
		{
			name:        "nested immutable field changed",
			planConfig:  map[string]interface{}{"client_access": map[string]interface{}{"client_id": "new"}},
			stateConfig: map[string]interface{}{"client_access": map[string]interface{}{"client_id": "old"}},
			meta: &metadata.ConnectorMetadata{Config: metadata.Property{Properties: map[string]*metadata.Property{
				"client_access": {
					Type: "object",
					Properties: map[string]*metadata.Property{
						"client_id": {Type: "string", Immutable: true},
					},
				},
			}}},
			wantPaths:    []path.Path{path.Root("config").AtName("client_access").AtName("client_id")},
			wantWarnings: 1,
		},
		{
			name:        "multiple immutable fields changed",
			planConfig:  map[string]interface{}{"schema": "new", "table": "new"},
			stateConfig: map[string]interface{}{"schema": "old", "table": "old"},
			meta: &metadata.ConnectorMetadata{Config: metadata.Property{Properties: map[string]*metadata.Property{
				"schema": {Type: "string", Immutable: true},
				"table":  {Type: "string", Immutable: true},
			}}},
			wantPaths: []path.Path{
				path.Root("config").AtName("schema"),
				path.Root("config").AtName("table"),
			},
			wantWarnings: 2,
		},
		{
			name:        "unchanged immutable field",
			planConfig:  map[string]interface{}{"schema": "same"},
			stateConfig: map[string]interface{}{"schema": "same"},
			meta: &metadata.ConnectorMetadata{Config: metadata.Property{Properties: map[string]*metadata.Property{
				"schema": {Type: "string", Immutable: true},
			}}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := &sync.Map{}
			cache.Store("google_ads", tt.meta)

			r := configuredConnectionV2ForValidation(t, false, cache)
			req := connectionV2ModifyPlanRequest(t,
				connectionV2ModifyPlanModel(t, "google_ads", "group_id", tt.planConfig, tt.planAuth),
				connectionV2ModifyPlanModel(t, "google_ads", "group_id", tt.stateConfig, tt.stateAuth),
			)

			var resp resource.ModifyPlanResponse
			r.ModifyPlan(context.Background(), req, &resp)

			assertErrorCount(t, resp.Diagnostics, 0)
			assertWarningCount(t, resp.Diagnostics, tt.wantWarnings)
			assertRequiresReplacePaths(t, resp.RequiresReplace, tt.wantPaths)
		})
	}
}

func TestConnectionV2ModifyPlanWarnsForRootReplacementFields(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("postgres", &metadata.ConnectorMetadata{})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "postgres", "new_group", nil, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "old_group", nil, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 2)
	if len(resp.RequiresReplace) != 0 {
		t.Fatalf("resource-level ModifyPlan should not duplicate schema-level root RequiresReplace paths, got %v", resp.RequiresReplace)
	}
	assertDiagnosticContains(t, resp.Diagnostics, "service")
	assertDiagnosticContains(t, resp.Diagnostics, "group_id")
}

func TestConnectionV2ModifyPlanValidatesRequiredFieldsOnCreate(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Required: []string{"schema", "table"},
			Properties: map[string]*metadata.Property{
				"schema": {Type: "string"},
				"table":  {Type: "string"},
			},
		},
		Auth: metadata.Property{
			Required: []string{"refresh_token"},
			Properties: map[string]*metadata.Property{
				"refresh_token": {Type: "string"},
			},
		},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanCreateRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id",
			map[string]interface{}{"schema": "app"},
			map[string]interface{}{},
		),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 2)
	assertWarningCount(t, resp.Diagnostics, 0)
}

func TestConnectionV2ModifyPlanAcceptsRequiredFieldsOnCreate(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Required: []string{"schema"},
			Properties: map[string]*metadata.Property{
				"schema": {Type: "string"},
			},
		},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanCreateRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "app"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	if len(resp.RequiresReplace) != 0 {
		t.Fatalf("create should not request replacement paths, got %v", resp.RequiresReplace)
	}
}

func TestConnectionV2ModifyPlanDoesNotValidateRequiredFieldsOnUpdate(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Required: []string{"schema"},
			Properties: map[string]*metadata.Property{
				"schema": {Type: "string"},
			},
		},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{}, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "app"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	if len(resp.RequiresReplace) != 0 {
		t.Fatalf("omitting a required field on update should not request replacement, got %v", resp.RequiresReplace)
	}
}

func TestConnectionV2ModifyPlanValidatesRequiredFieldsOnRootReplacement(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("postgres", &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Required: []string{"schema"},
			Properties: map[string]*metadata.Property{
				"schema": {Type: "string"},
			},
		},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "postgres", "group_id", nil, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "existing"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertDiagnosticContains(t, resp.Diagnostics, "schema")
	assertDiagnosticContains(t, resp.Diagnostics, "service")
	if len(resp.RequiresReplace) != 0 {
		t.Fatalf("resource-level ModifyPlan should not duplicate schema-level root RequiresReplace paths, got %v", resp.RequiresReplace)
	}
}

func TestConnectionV2ModifyPlanValidatesRequiredFieldsOnImmutableReplacement(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{
			Required: []string{"schema"},
			Properties: map[string]*metadata.Property{
				"account_name": {Type: "string", Immutable: true},
				"schema":       {Type: "string"},
			},
		},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"account_name": "new"}, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"account_name": "old", "schema": "existing"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertDiagnosticContains(t, resp.Diagnostics, "schema")
	assertRequiresReplacePaths(t, resp.RequiresReplace, []path.Path{path.Root("config").AtName("account_name")})
}

func TestConnectionV2ModifyPlanReportsMetadataFetchFailure(t *testing.T) {
	t.Parallel()

	client := fivetran.New("key", "secret")
	client.SetHttpClient(modifyPlanErrorHTTPClient{})

	r := configuredConnectionV2ForValidationWithClient(t, false, &sync.Map{}, client)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "new"}, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "old"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
}

func TestConnectionV2ModifyPlanSkipPlanTimeValidationBypassesMetadataChecks(t *testing.T) {
	t.Parallel()

	r := configuredConnectionV2ForValidation(t, true, nil)
	req := connectionV2ModifyPlanRequest(t,
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "new"}, nil),
		connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "old"}, nil),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	if len(resp.RequiresReplace) != 0 {
		t.Fatalf("skip_plan_time_validation should bypass metadata-driven replacement checks, got %v", resp.RequiresReplace)
	}
}

func TestConnectionV2ModifyPlanSkipsDeleteAndUnknownService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  resource.ModifyPlanRequest
	}{
		{
			name: "delete",
			req: connectionV2ModifyPlanDeleteRequest(t,
				connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "app"}, nil),
			),
		},
		{
			name: "unknown service",
			req: connectionV2ModifyPlanRequest(t,
				connectionV2ModifyPlanModelWithService(t, types.StringUnknown(), "group_id", map[string]interface{}{"schema": "new"}, nil),
				connectionV2ModifyPlanModel(t, "google_ads", "group_id", map[string]interface{}{"schema": "old"}, nil),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := configuredConnectionV2ForValidation(t, false, &sync.Map{})
			var resp resource.ModifyPlanResponse
			r.ModifyPlan(context.Background(), tt.req, &resp)

			assertNoDiagnostics(t, resp.Diagnostics)
			if len(resp.RequiresReplace) != 0 {
				t.Fatalf("unexpected replacement paths: %v", resp.RequiresReplace)
			}
		})
	}
}

type modifyPlanErrorHTTPClient struct{}

func (modifyPlanErrorHTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("metadata unavailable")
}

func connectionV2ModifyPlanModel(t *testing.T, service, groupID string, config, auth map[string]interface{}) model.ConnectionV2ResourceModel {
	t.Helper()
	return connectionV2ModifyPlanModelWithService(t, types.StringValue(service), groupID, config, auth)
}

func connectionV2ModifyPlanModelWithService(t *testing.T, service types.String, groupID string, config, auth map[string]interface{}) model.ConnectionV2ResourceModel {
	t.Helper()
	ctx := context.Background()

	configValue := types.DynamicNull()
	if config != nil {
		var diags diag.Diagnostics
		configValue, diags = core.MapToDynamic(ctx, config)
		if diags.HasError() {
			t.Fatalf("config dynamic diagnostics: %v", diags)
		}
	}

	authValue := types.DynamicNull()
	if auth != nil {
		var diags diag.Diagnostics
		authValue, diags = core.MapToDynamic(ctx, auth)
		if diags.HasError() {
			t.Fatalf("auth dynamic diagnostics: %v", diags)
		}
	}

	return model.ConnectionV2ResourceModel{
		Id:                      types.StringValue("connection_id"),
		Name:                    types.StringNull(),
		ConnectedBy:             types.StringNull(),
		CreatedAt:               types.StringNull(),
		GroupId:                 types.StringValue(groupID),
		Service:                 service,
		Config:                  configValue,
		Auth:                    authValue,
		SucceededAt:             types.StringNull(),
		FailedAt:                types.StringNull(),
		ServiceVersion:          types.StringNull(),
		SyncFrequency:           types.Int64Null(),
		ScheduleType:            types.StringNull(),
		PauseAfterTrial:         types.BoolNull(),
		DailySyncTime:           types.StringNull(),
		ProxyAgentId:            types.StringNull(),
		NetworkingMethod:        types.StringNull(),
		HybridDeploymentAgentId: types.StringNull(),
		PrivateLinkId:           types.StringNull(),
		DataDelaySensitivity:    types.StringNull(),
		DataDelayThreshold:      types.Int64Null(),
		RunSetupTests:           types.BoolNull(),
		TrustCertificates:       types.BoolNull(),
		TrustFingerprints:       types.BoolNull(),
		Status:                  types.ObjectNull(model.ConnectionV2StatusAttrTypes()),
	}
}

func connectionV2ModifyPlanRequest(t *testing.T, plan, state model.ConnectionV2ResourceModel) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.ConnectionV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan: tfsdk.Plan{
			Raw:    connectionV2TerraformValue(t, plan),
			Schema: schema,
		},
		State: tfsdk.State{
			Raw:    connectionV2TerraformValue(t, state),
			Schema: schema,
		},
	}
}

func connectionV2ModifyPlanCreateRequest(t *testing.T, plan model.ConnectionV2ResourceModel) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.ConnectionV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan: tfsdk.Plan{
			Raw:    connectionV2TerraformValue(t, plan),
			Schema: schema,
		},
		State: tfsdk.State{
			Raw:    connectionV2NullTerraformValue(t),
			Schema: schema,
		},
	}
}

func connectionV2ModifyPlanDeleteRequest(t *testing.T, state model.ConnectionV2ResourceModel) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.ConnectionV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan: tfsdk.Plan{
			Raw:    connectionV2NullTerraformValue(t),
			Schema: schema,
		},
		State: tfsdk.State{
			Raw:    connectionV2TerraformValue(t, state),
			Schema: schema,
		},
	}
}

func connectionV2TerraformValue(t *testing.T, data model.ConnectionV2ResourceModel) tftypes.Value {
	t.Helper()
	ctx := context.Background()

	var object types.Object
	diags := tfsdk.ValueFrom(ctx, data, types.ObjectType{AttrTypes: model.ConnectionV2ResourceModelAttrTypes()}, &object)
	if diags.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", diags)
	}

	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting model to Terraform value: %v", err)
	}
	return raw
}

func connectionV2NullTerraformValue(t *testing.T) tftypes.Value {
	t.Helper()
	ctx := context.Background()
	schema := fivetranSchema.ConnectionV2ResourceSchema()
	return tftypes.NewValue(schema.Type().TerraformType(ctx), nil)
}

func assertRequiresReplacePaths(t *testing.T, got path.Paths, want []path.Path) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("requires replace paths = %v, want %v", got, want)
	}
	for _, p := range want {
		if !got.Contains(p) {
			t.Fatalf("requires replace paths = %v, missing %s", got, p.String())
		}
	}
}

func assertDiagnosticContains(t *testing.T, diags diag.Diagnostics, text string) {
	t.Helper()
	for _, d := range diags {
		if strings.Contains(d.Detail(), text) || strings.Contains(d.Summary(), text) {
			return
		}
	}
	t.Fatalf("diagnostics %v did not contain %q", diags, text)
}
