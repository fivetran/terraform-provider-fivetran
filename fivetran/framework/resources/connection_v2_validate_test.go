package resources

import (
	"context"
	"errors"
	"net/http"
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
)

func TestValidateDynamicObjectAcceptsKnownFields(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(map[string]interface{}{
		"schema":  "app",
		"enabled": true,
		"count":   int64(3),
		"mode":    "AllAccounts",
		"nested": map[string]interface{}{
			"table": "events",
		},
		"reports": []interface{}{
			map[string]interface{}{"report_type": "campaign"},
		},
	}, &metadata.Property{
		Properties: map[string]*metadata.Property{
			"schema":  {Type: "string"},
			"enabled": {Type: "boolean"},
			"count":   {Type: "integer"},
			"mode":    {Type: "string", Enum: []string{"AllAccounts", "SpecificAccounts"}},
			"nested": {
				Type: "object",
				Properties: map[string]*metadata.Property{
					"table": {Type: "string"},
				},
			},
			"reports": {
				Type: "array",
				Items: &metadata.Property{
					Type: "object",
					Properties: map[string]*metadata.Property{
						"report_type": {Type: "string"},
					},
				},
			},
		},
	}, path.Root("config"), &diags)

	if diags.HasError() {
		t.Fatalf("unexpected validation errors: %v", diags)
	}
}

func TestValidateDynamicObjectRejectsUnknownField(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(
		map[string]interface{}{"dog": "good_boy"},
		&metadata.Property{Properties: map[string]*metadata.Property{"schema": {Type: "string"}}},
		path.Root("config"),
		&diags,
	)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", diags.ErrorsCount(), diags)
	}
}

func TestValidateDynamicObjectRejectsWrongType(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(
		map[string]interface{}{"schema": true},
		&metadata.Property{Properties: map[string]*metadata.Property{"schema": {Type: "string"}}},
		path.Root("config"),
		&diags,
	)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", diags.ErrorsCount(), diags)
	}
}

func TestValidateDynamicObjectRejectsInvalidEnum(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(
		map[string]interface{}{"sync_mode": "Nope"},
		&metadata.Property{Properties: map[string]*metadata.Property{
			"sync_mode": {Type: "string", Enum: []string{"AllAccounts", "SpecificAccounts"}},
		}},
		path.Root("config"),
		&diags,
	)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", diags.ErrorsCount(), diags)
	}
}

func TestValidateDynamicValueTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prop       *metadata.Property
		value      interface{}
		wantErrors int
	}{
		{name: "string accepts string", prop: &metadata.Property{Type: "string"}, value: "app"},
		{name: "string rejects bool", prop: &metadata.Property{Type: "string"}, value: true, wantErrors: 1},
		{name: "string rejects int", prop: &metadata.Property{Type: "string"}, value: int64(1), wantErrors: 1},
		{name: "string rejects object", prop: &metadata.Property{Type: "string"}, value: map[string]interface{}{"schema": "app"}, wantErrors: 1},
		{name: "integer accepts int64", prop: &metadata.Property{Type: "integer"}, value: int64(3)},
		{name: "integer accepts exact float64", prop: &metadata.Property{Type: "integer"}, value: float64(3)},
		{name: "integer rejects non-integer float64", prop: &metadata.Property{Type: "integer"}, value: float64(3.5), wantErrors: 1},
		{name: "integer rejects string", prop: &metadata.Property{Type: "integer"}, value: "3", wantErrors: 1},
		{name: "number accepts int64", prop: &metadata.Property{Type: "number"}, value: int64(3)},
		{name: "number accepts float64", prop: &metadata.Property{Type: "number"}, value: float64(3.5)},
		{name: "number rejects string", prop: &metadata.Property{Type: "number"}, value: "3.5", wantErrors: 1},
		{name: "number rejects bool", prop: &metadata.Property{Type: "number"}, value: true, wantErrors: 1},
		{name: "boolean accepts bool", prop: &metadata.Property{Type: "boolean"}, value: true},
		{name: "boolean rejects string", prop: &metadata.Property{Type: "boolean"}, value: "true", wantErrors: 1},
		{name: "null rejects non-nullable field", prop: &metadata.Property{Type: "string"}, value: nil, wantErrors: 1},
		{name: "null accepts nullable field", prop: &metadata.Property{Type: "string", Nullable: true}, value: nil},
		{name: "empty metadata type is tolerated", prop: &metadata.Property{Type: ""}, value: map[string]interface{}{"schema": "app"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			validateDynamicValue(tt.value, tt.prop, path.Root("config").AtName("field"), &diags)
			assertErrorCount(t, diags, tt.wantErrors)
			assertWarningCount(t, diags, 0)
		})
	}
}

func TestValidateDynamicValueCollections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prop       *metadata.Property
		value      interface{}
		wantErrors int
	}{
		{
			name:  "array accepts valid item types",
			prop:  &metadata.Property{Type: "array", Items: &metadata.Property{Type: "string"}},
			value: []interface{}{"one", "two"},
		},
		{
			name:       "array rejects non-array value",
			prop:       &metadata.Property{Type: "array", Items: &metadata.Property{Type: "string"}},
			value:      "not-array",
			wantErrors: 1,
		},
		{
			name:       "array rejects invalid item type",
			prop:       &metadata.Property{Type: "array", Items: &metadata.Property{Type: "string"}},
			value:      []interface{}{"one", true},
			wantErrors: 1,
		},
		{
			name:  "array with nil items skips item validation",
			prop:  &metadata.Property{Type: "array"},
			value: []interface{}{"one", true, map[string]interface{}{"nested": "value"}},
		},
		{
			name: "object accepts nested map",
			prop: &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{
				"table": {Type: "string"},
			}},
			value: map[string]interface{}{"table": "events"},
		},
		{
			name:       "object rejects non-object value",
			prop:       &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{"table": {Type: "string"}}},
			value:      "not-object",
			wantErrors: 1,
		},
		{
			name:       "nested object rejects unknown field",
			prop:       &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{"table": {Type: "string"}}},
			value:      map[string]interface{}{"dog": "good_boy"},
			wantErrors: 1,
		},
		{
			name:       "nested object rejects wrong field type",
			prop:       &metadata.Property{Type: "object", Properties: map[string]*metadata.Property{"table": {Type: "string"}}},
			value:      map[string]interface{}{"table": true},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			validateDynamicValue(tt.value, tt.prop, path.Root("config").AtName("field"), &diags)
			assertErrorCount(t, diags, tt.wantErrors)
			assertWarningCount(t, diags, 0)
		})
	}
}

func TestValidateDynamicObjectRejectsUnknownMetadataType(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(
		map[string]interface{}{"schema": "app"},
		&metadata.Property{Properties: map[string]*metadata.Property{
			"schema": {Type: "wat"},
		}},
		path.Root("config"),
		&diags,
	)

	if diags.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", diags.ErrorsCount(), diags)
	}
}

func TestValidateDynamicObjectFieldStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fieldStatus  string
		wantWarnings int
	}{
		{name: "missing status"},
		{name: "general availability", fieldStatus: core.FieldStatusGeneralAvailability},
		{name: "private preview", fieldStatus: core.FieldStatusPrivatePreview, wantWarnings: 1},
		{name: "development", fieldStatus: core.FieldStatusDevelopment, wantWarnings: 1},
		{name: "sunset", fieldStatus: core.FieldStatusSunset, wantWarnings: 1},
		{name: "unknown status", fieldStatus: "some_future_status", wantWarnings: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var diags diag.Diagnostics
			validateDynamicObject(
				map[string]interface{}{"field": "value"},
				&metadata.Property{Properties: map[string]*metadata.Property{
					"field": {Type: "string", FieldStatus: tt.fieldStatus},
				}},
				path.Root("config"),
				&diags,
			)

			assertErrorCount(t, diags, 0)
			assertWarningCount(t, diags, tt.wantWarnings)
		})
	}
}

func TestValidateDynamicObjectWarnsForNonGAFieldStatus(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	validateDynamicObject(
		map[string]interface{}{"preview_field": "value"},
		&metadata.Property{Properties: map[string]*metadata.Property{
			"preview_field": {Type: "string", FieldStatus: "private_preview"},
		}},
		path.Root("config"),
		&diags,
	)

	if diags.HasError() {
		t.Fatalf("unexpected validation errors: %v", diags)
	}
	if diags.WarningsCount() != 1 {
		t.Fatalf("warnings = %d, want 1: %v", diags.WarningsCount(), diags)
	}
}

func TestConnectionV2ValidateConfigEarlyReturns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service types.String
		config  map[string]interface{}
		auth    map[string]interface{}
	}{
		{
			name:    "null config and auth skips metadata fetch",
			service: types.StringValue("google_ads"),
		},
		{
			name:    "empty config and auth skips metadata fetch",
			service: types.StringValue("google_ads"),
			config:  map[string]interface{}{},
			auth:    map[string]interface{}{},
		},
		{
			name:    "null service skips metadata fetch",
			service: types.StringNull(),
			config:  map[string]interface{}{"schema": "app"},
		},
		{
			name:    "unknown service skips metadata fetch",
			service: types.StringUnknown(),
			config:  map[string]interface{}{"schema": "app"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := configuredConnectionV2ForValidation(t, false, &sync.Map{})
			req := connectionV2ValidateConfigRequestWithService(t, tt.service, tt.config, tt.auth)

			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), req, &resp)

			assertNoDiagnostics(t, resp.Diagnostics)
		})
	}
}

func TestConnectionV2ValidateConfigUsesMetadataCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{Properties: map[string]*metadata.Property{
			"schema":    {Type: "string"},
			"sync_mode": {Type: "string", Enum: []string{"AllAccounts", "SpecificAccounts"}},
		}},
		Auth: metadata.Property{Properties: map[string]*metadata.Property{
			"refresh_token": {Type: "string"},
		}},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{"schema": "app", "sync_mode": "AllAccounts"},
		map[string]interface{}{"refresh_token": "secret"},
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected validation errors: %v", resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigReportsMetadataFetchFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	client := fivetran.New("key", "secret")
	client.SetHttpClient(errorHTTPClient{})

	r := configuredConnectionV2ForValidationWithClient(t, false, &sync.Map{}, client)
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{"schema": "app"},
		nil,
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigSkipsMetadataFetchForEmptyDynamicObjects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := configuredConnectionV2ForValidation(t, false, &sync.Map{})
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{},
		map[string]interface{}{},
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected validation errors: %v", resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigReportsUnconfiguredClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := configuredConnectionV2ForValidation(t, false, &sync.Map{})
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{"schema": "app"},
		nil,
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigReportsUnexpectedMetadataCacheType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cache := &sync.Map{}
	cache.Store("google_ads", "not metadata")

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{"schema": "app"},
		nil,
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.ErrorsCount() != 1 {
		t.Fatalf("errors = %d, want 1: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigReportsMetadataValidationErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cache := &sync.Map{}
	cache.Store("google_ads", &metadata.ConnectorMetadata{
		Config: metadata.Property{Properties: map[string]*metadata.Property{
			"sync_mode": {Type: "string", Enum: []string{"AllAccounts", "SpecificAccounts"}},
		}},
	})

	r := configuredConnectionV2ForValidation(t, false, cache)
	req := connectionV2ValidateConfigRequest(t, "google_ads",
		map[string]interface{}{"dog": "good_boy", "sync_mode": "Nope"},
		nil,
	)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if resp.Diagnostics.ErrorsCount() != 2 {
		t.Fatalf("errors = %d, want 2: %v", resp.Diagnostics.ErrorsCount(), resp.Diagnostics)
	}
}

func TestConnectionV2ValidateConfigSkipReturnsBeforeConfigDecode(t *testing.T) {
	t.Parallel()

	r := configuredConnectionV2ForValidation(t, true, nil)
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("skip_plan_time_validation should return before config decode, got: %v", resp.Diagnostics)
	}
}

func configuredConnectionV2ForValidation(t *testing.T, skip bool, cache *sync.Map) *connectionV2 {
	t.Helper()

	return configuredConnectionV2ForValidationWithClient(t, skip, cache, nil)
}

func configuredConnectionV2ForValidationWithClient(t *testing.T, skip bool, cache *sync.Map, client *fivetran.Client) *connectionV2 {
	t.Helper()

	r := &connectionV2{}
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{
		ProviderData: &core.ProviderResourceData{
			Client:                 client,
			MetadataCache:          cache,
			SkipPlanTimeValidation: skip,
		},
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", resp.Diagnostics)
	}
	return r
}

type errorHTTPClient struct{}

func (errorHTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("metadata unavailable")
}

func connectionV2ValidateConfigRequest(t *testing.T, service string, config, auth map[string]interface{}) resource.ValidateConfigRequest {
	t.Helper()
	return connectionV2ValidateConfigRequestWithService(t, types.StringValue(service), config, auth)
}

func connectionV2ValidateConfigRequestWithService(t *testing.T, service types.String, config, auth map[string]interface{}) resource.ValidateConfigRequest {
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

	data := model.ConnectionV2ResourceModel{
		Id:                      types.StringNull(),
		Name:                    types.StringNull(),
		ConnectedBy:             types.StringNull(),
		CreatedAt:               types.StringNull(),
		GroupId:                 types.StringValue("group_id"),
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

	var object types.Object
	diags := tfsdk.ValueFrom(ctx, data, types.ObjectType{AttrTypes: model.ConnectionV2ResourceModelAttrTypes()}, &object)
	if diags.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", diags)
	}

	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting config to Terraform value: %v", err)
	}

	return resource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    raw,
			Schema: fivetranSchema.ConnectionV2ResourceSchema(),
		},
	}
}

func assertNoDiagnostics(t *testing.T, diags diag.Diagnostics) {
	t.Helper()
	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
}

func assertErrorCount(t *testing.T, diags diag.Diagnostics, want int) {
	t.Helper()
	if diags.ErrorsCount() != want {
		t.Fatalf("errors = %d, want %d: %v", diags.ErrorsCount(), want, diags)
	}
}

func assertWarningCount(t *testing.T, diags diag.Diagnostics, want int) {
	t.Helper()
	if diags.WarningsCount() != want {
		t.Fatalf("warnings = %d, want %d: %v", diags.WarningsCount(), want, diags)
	}
}
