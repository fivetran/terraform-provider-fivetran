package resources

import (
	"context"
	"sync"
	"testing"

	fivetran "github.com/fivetran/go-fivetran"
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

func snowflakeDestinationMetadata() *metadata.DestinationMetadata {
	return &metadata.DestinationMetadata{
		ID:                         "snowflake",
		SupportedRegions:           []string{"GCP_US_EAST4", "AWS_US_EAST_1"},
		SupportsHybridDeployment:   false,
		SupportedNetworkingMethods: []string{"Directly", "SshTunnel", "PrivateLink"},
		SupportsPrivateLink:        true,
		PrivateLinkTypes:           []string{"CUSTOMER_MANAGED"},
		Config: metadata.Property{
			Type: "object",
			Properties: map[string]*metadata.Property{
				"host":       {Type: "string"},
				"port":       {Type: "integer"},
				"database":   {Type: "string", Immutable: true},
				"password":   {Type: "string", Format: "password"},
				"auth":       {Type: "string", Enum: []string{"PASSWORD", "KEY_PAIR"}},
				"public_key": {Type: "string", Readonly: true},
				"nested": {
					Type: "object",
					Properties: map[string]*metadata.Property{
						"inner_readonly": {Type: "string", Readonly: true},
						"inner":          {Type: "string"},
					},
				},
			},
		},
	}
}

func destinationV2CacheWith(service string, meta *metadata.DestinationMetadata) *sync.Map {
	cache := &sync.Map{}
	cache.Store(core.DestinationMetadataCacheKey(service), meta)
	return cache
}

func TestDestinationV2ValidateConfigAcceptsKnownFields(t *testing.T) {
	t.Parallel()

	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
	data := destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{
		"host":     "account.snowflakecomputing.com",
		"port":     443,
		"password": "secret",
		"auth":     "PASSWORD",
		"nested":   map[string]interface{}{"inner": "value"},
	})
	data.Region = types.StringValue("GCP_US_EAST4")
	data.NetworkingMethod = types.StringValue("Directly")

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, data), &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestDestinationV2ValidateConfigRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   map[string]interface{}
		wantPath path.Path
	}{
		{
			name:     "unknown field",
			config:   map[string]interface{}{"hots": "typo"},
			wantPath: path.Root("config").AtName("hots"),
		},
		{
			name:     "wrong type",
			config:   map[string]interface{}{"port": "443"},
			wantPath: path.Root("config").AtName("port"),
		},
		{
			name:     "invalid enum",
			config:   map[string]interface{}{"auth": "OAUTH"},
			wantPath: path.Root("config").AtName("auth"),
		},
		{
			name:     "readonly field",
			config:   map[string]interface{}{"public_key": "key"},
			wantPath: path.Root("config").AtName("public_key"),
		},
		{
			name:     "nested readonly field",
			config:   map[string]interface{}{"nested": map[string]interface{}{"inner_readonly": "value"}},
			wantPath: path.Root("config").AtName("nested").AtName("inner_readonly"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, destinationV2Model(t, "snowflake", "group_id", tt.config)), &resp)

			assertErrorCount(t, resp.Diagnostics, 1)
			assertDiagnosticPath(t, resp.Diagnostics, tt.wantPath)
		})
	}
}

func TestDestinationV2ValidateConfigWarnsForUnsupportedRootFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mutate   func(*model.DestinationV2ResourceModel)
		wantPath path.Path
	}{
		{
			name:     "region",
			mutate:   func(d *model.DestinationV2ResourceModel) { d.Region = types.StringValue("AZURE_EASTUS2") },
			wantPath: path.Root("region"),
		},
		{
			name:     "networking method",
			mutate:   func(d *model.DestinationV2ResourceModel) { d.NetworkingMethod = types.StringValue("ProxyAgent") },
			wantPath: path.Root("networking_method"),
		},
		{
			name:     "hybrid deployment",
			mutate:   func(d *model.DestinationV2ResourceModel) { d.HybridDeploymentAgentId = types.StringValue("agent_id") },
			wantPath: path.Root("hybrid_deployment_agent_id"),
		},
		{
			name:     "self-service private link",
			mutate:   func(d *model.DestinationV2ResourceModel) { d.PrivateLinkId = types.StringValue("pl_id") },
			wantPath: path.Root("private_link_id"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
			data := destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"})
			tt.mutate(&data)

			var resp resource.ValidateConfigResponse
			r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, data), &resp)

			assertErrorCount(t, resp.Diagnostics, 0)
			assertWarningCount(t, resp.Diagnostics, 1)
			assertDiagnosticPath(t, resp.Diagnostics, tt.wantPath)
		})
	}
}

func TestDestinationV2ValidateConfigSkipsUnknownValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
	data := destinationV2Model(t, "snowflake", "group_id", nil)
	configValue, diags := types.ObjectValue(
		map[string]attr.Type{"port": types.StringType},
		map[string]attr.Value{"port": types.StringUnknown()},
	)
	if diags.HasError() {
		t.Fatalf("object diagnostics: %v", diags)
	}
	data.Config = types.DynamicValue(configValue)

	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, destinationV2ValidateConfigRequest(t, data), &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
}

func TestDestinationV2ValidateConfigEarlyReturns(t *testing.T) {
	t.Parallel()

	t.Run("skip plan time validation", func(t *testing.T) {
		t.Parallel()
		r := configuredDestinationV2(t, true, nil, nil)
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"hots": "typo"})), &resp)
		assertNoDiagnostics(t, resp.Diagnostics)
	})

	t.Run("unconfigured client and empty cache", func(t *testing.T) {
		t.Parallel()
		r := configuredDestinationV2(t, false, &sync.Map{}, nil)
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"hots": "typo"})), &resp)
		assertNoDiagnostics(t, resp.Diagnostics)
	})

	t.Run("unknown service", func(t *testing.T) {
		t.Parallel()
		r := configuredDestinationV2(t, false, &sync.Map{}, nil)
		data := destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"hots": "typo"})
		data.Service = types.StringUnknown()
		var resp resource.ValidateConfigResponse
		r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, data), &resp)
		assertNoDiagnostics(t, resp.Diagnostics)
	})
}

func TestDestinationV2ValidateConfigReportsMetadataFetchFailure(t *testing.T) {
	t.Parallel()

	client := fivetran.New("key", "secret")
	client.SetHttpClient(errorHTTPClient{})

	r := configuredDestinationV2(t, false, &sync.Map{}, client)
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(context.Background(), destinationV2ValidateConfigRequest(t, destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"})), &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
	assertDiagnosticContains(t, resp.Diagnostics, "skip_plan_time_validation")
}

// Destination metadata lives under its own key prefix, so a connector entry
// with the same service id must not be picked up (and vice versa).
func TestDestinationV2MetadataCacheIsolatedFromConnectorMetadata(t *testing.T) {
	t.Parallel()

	cache := &sync.Map{}
	cache.Store("snowflake", &metadata.ConnectorMetadata{})
	cache.Store(core.DestinationMetadataCacheKey("snowflake"), snowflakeDestinationMetadata())

	destinationMeta, ok, err := core.LoadCachedDestinationMetadata(cache, "snowflake")
	if err != nil || !ok || destinationMeta.ID != "snowflake" {
		t.Fatalf("destination metadata = %v, %v, %v", destinationMeta, ok, err)
	}
	connectorMeta, ok, err := core.LoadCachedConnectorMetadata(cache, "snowflake")
	if err != nil || !ok || connectorMeta == nil {
		t.Fatalf("connector metadata = %v, %v, %v", connectorMeta, ok, err)
	}
}

func TestDestinationV2ModifyPlanRequiresReplaceForImmutableConfigField(t *testing.T) {
	t.Parallel()

	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
	req := destinationV2ModifyPlanRequest(t,
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "new", "host": "h"}),
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "old", "host": "h"}),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 0)
	assertWarningCount(t, resp.Diagnostics, 1)
	assertDiagnosticContains(t, resp.Diagnostics, "recreating this destination")
	assertRequiresReplacePaths(t, resp.RequiresReplace, []path.Path{path.Root("config").AtName("database")})
}

func TestDestinationV2ModifyPlanIgnoresMutableAndUnchangedFields(t *testing.T) {
	t.Parallel()

	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", snowflakeDestinationMetadata()), nil)
	req := destinationV2ModifyPlanRequest(t,
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "same", "host": "new"}),
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "same", "host": "old"}),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	assertRequiresReplacePaths(t, resp.RequiresReplace, nil)
}

func TestDestinationV2ModifyPlanValidatesRequiredFieldsOnCreate(t *testing.T) {
	t.Parallel()

	meta := snowflakeDestinationMetadata()
	meta.Config.Required = []string{"host", "database"}
	r := configuredDestinationV2(t, false, destinationV2CacheWith("snowflake", meta), nil)
	req := destinationV2ModifyPlanCreateRequest(t, destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"host": "h"}))

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
	assertDiagnosticPath(t, resp.Diagnostics, path.Root("config").AtName("database"))
}

func TestDestinationV2ModifyPlanSkipPlanTimeValidation(t *testing.T) {
	t.Parallel()

	r := configuredDestinationV2(t, true, nil, nil)
	req := destinationV2ModifyPlanRequest(t,
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "new"}),
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "old"}),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertNoDiagnostics(t, resp.Diagnostics)
	assertRequiresReplacePaths(t, resp.RequiresReplace, nil)
}

func TestDestinationV2ModifyPlanReportsMetadataFetchFailure(t *testing.T) {
	t.Parallel()

	client := fivetran.New("key", "secret")
	client.SetHttpClient(modifyPlanErrorHTTPClient{})

	r := configuredDestinationV2(t, false, &sync.Map{}, client)
	req := destinationV2ModifyPlanRequest(t,
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "new"}),
		destinationV2Model(t, "snowflake", "group_id", map[string]interface{}{"database": "old"}),
	)

	var resp resource.ModifyPlanResponse
	r.ModifyPlan(context.Background(), req, &resp)

	assertErrorCount(t, resp.Diagnostics, 1)
}

func TestDestinationV2SchemaShape(t *testing.T) {
	t.Parallel()

	r := DestinationV2()
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", schemaResp.Diagnostics)
	}
	if schemaResp.Schema.Version != 0 {
		t.Fatalf("unexpected schema version: got %d, want 0", schemaResp.Schema.Version)
	}
	if len(schemaResp.Schema.Blocks) != 0 {
		t.Fatalf("fivetran_destination_v2 must not declare blocks, got %v", schemaResp.Schema.Blocks)
	}
	attrs := schemaResp.Schema.Attributes
	for _, name := range []string{"local_processing_agent_id", "timeouts"} {
		if _, ok := attrs[name]; ok {
			t.Fatalf("fivetran_destination_v2 must not expose %s", name)
		}
	}
	if !attrs["config"].IsRequired() || attrs["config"].IsSensitive() {
		t.Fatal("config must be required and not sensitive; password fields are protected via metadata projection")
	}
	if secrets := attrs["config_secrets"]; !secrets.IsOptional() || !secrets.IsWriteOnly() || !secrets.IsSensitive() {
		t.Fatal("config_secrets must be optional, write-only and sensitive")
	}
	if version := attrs["config_secrets_version"]; !version.IsOptional() || version.IsComputed() || version.IsWriteOnly() {
		t.Fatal("config_secrets_version must be optional, not computed and stored in state")
	}
	for _, name := range []string{"group_id", "service", "time_zone_offset"} {
		if !attrs[name].IsRequired() {
			t.Fatalf("%s must be required", name)
		}
	}
	for _, name := range []string{"region", "daylight_saving_time_enabled", "networking_method", "private_link_id", "proxy_agent_id", "hybrid_deployment_agent_id", "run_setup_tests", "trust_certificates", "trust_fingerprints"} {
		if !attrs[name].IsOptional() || !attrs[name].IsComputed() {
			t.Fatalf("%s must be optional and computed", name)
		}
	}
}

func configuredDestinationV2(t *testing.T, skip bool, cache *sync.Map, client *fivetran.Client) *destinationV2 {
	t.Helper()

	r := &destinationV2{}
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

func destinationV2Model(t *testing.T, service, groupID string, config map[string]interface{}) model.DestinationV2ResourceModel {
	t.Helper()

	configValue := types.DynamicNull()
	if config != nil {
		var diags diag.Diagnostics
		configValue, diags = core.MapToDynamic(context.Background(), config)
		if diags.HasError() {
			t.Fatalf("config dynamic diagnostics: %v", diags)
		}
	}

	return model.DestinationV2ResourceModel{
		Id:                        types.StringValue("destination_id"),
		GroupId:                   types.StringValue(groupID),
		Service:                   types.StringValue(service),
		Config:                    configValue,
		ConfigSecrets:             types.DynamicNull(),
		ConfigSecretsVersion:      types.Int64Null(),
		Region:                    types.StringNull(),
		TimeZoneOffset:            types.StringValue("0"),
		DaylightSavingTimeEnabled: types.BoolNull(),
		SetupStatus:               types.StringNull(),
		NetworkingMethod:          types.StringNull(),
		PrivateLinkId:             types.StringNull(),
		ProxyAgentId:              types.StringNull(),
		HybridDeploymentAgentId:   types.StringNull(),
		RunSetupTests:             types.BoolNull(),
		TrustCertificates:         types.BoolNull(),
		TrustFingerprints:         types.BoolNull(),
	}
}

func destinationV2TerraformValue(t *testing.T, data model.DestinationV2ResourceModel) tftypes.Value {
	t.Helper()
	ctx := context.Background()

	var object types.Object
	diags := tfsdk.ValueFrom(ctx, data, types.ObjectType{AttrTypes: model.DestinationV2ResourceModelAttrTypes()}, &object)
	if diags.HasError() {
		t.Fatalf("ValueFrom diagnostics: %v", diags)
	}
	raw, err := object.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("converting model to Terraform value: %v", err)
	}
	return raw
}

func destinationV2ValidateConfigRequest(t *testing.T, data model.DestinationV2ResourceModel) resource.ValidateConfigRequest {
	t.Helper()
	return resource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    destinationV2TerraformValue(t, data),
			Schema: fivetranSchema.DestinationV2ResourceSchema(),
		},
	}
}

func destinationV2ModifyPlanRequest(t *testing.T, plan, state model.DestinationV2ResourceModel) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.DestinationV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Raw: destinationV2TerraformValue(t, plan), Schema: schema},
		State: tfsdk.State{Raw: destinationV2TerraformValue(t, state), Schema: schema},
	}
}

func destinationV2ModifyPlanCreateRequest(t *testing.T, plan model.DestinationV2ResourceModel) resource.ModifyPlanRequest {
	t.Helper()
	schema := fivetranSchema.DestinationV2ResourceSchema()
	return resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Raw: destinationV2TerraformValue(t, plan), Schema: schema},
		State: tfsdk.State{Raw: tftypes.NewValue(schema.Type().TerraformType(context.Background()), nil), Schema: schema},
	}
}

func assertDiagnosticPath(t *testing.T, diags diag.Diagnostics, want path.Path) {
	t.Helper()
	for _, d := range diags {
		if withPath, ok := d.(diag.DiagnosticWithPath); ok && withPath.Path().Equal(want) {
			return
		}
	}
	t.Fatalf("no diagnostic at path %s: %v", want, diags)
}
