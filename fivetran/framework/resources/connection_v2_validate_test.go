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
		Service:                 types.StringValue(service),
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
